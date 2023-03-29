// Copyright 2022-2023 The Memphis.dev Authors
// Licensed under the Memphis Business Source License 1.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// Changed License: [Apache License, Version 2.0 (https://www.apache.org/licenses/LICENSE-2.0), as published by the Apache Foundation.
//
// https://github.com/memphisdev/memphis/blob/master/LICENSE
//
// Additional Use Grant: You may make use of the Licensed Work (i) only as part of your own product or service, provided it is not a message broker or a message queue product or service; and (ii) provided that you do not use, provide, distribute, or make available the Licensed Work as a Service.
// A "Service" is a commercial offering, product, hosted, or managed service, that allows third parties (other than your own employees and contractors acting on your behalf) to access and/or use the Licensed Work or a substantial set of the features or functionality of the Licensed Work to third parties as a software-as-a-service, platform-as-a-service, infrastructure-as-a-service or other similar services that compete with Licensor products or services.
package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"memphis/db"
	"memphis/models"
	"sync"

	"strconv"
	"strings"
	"time"
)

const CONN_STATUS_SUBJ = "$memphis_connection_status"
const INTEGRATIONS_UPDATES_SUBJ = "$memphis_integration_updates"
const CONFIGURATIONS_RELOAD_SIGNAL_SUBJ = "$memphis_config_reload_signal"
const NOTIFICATION_EVENTS_SUBJ = "$memphis_notifications"
const PM_RESEND_ACK_SUBJ = "$memphis_pm_acks"
const TIERED_STORAGE_CONSUMER = "$memphis_tiered_storage_consumer"
const SCHEMAVERSE_DLS_SUBJ = "$memphis_schemaverse_dls"

var LastReadThroughput models.Throughput
var LastWriteThroughput models.Throughput
var tieredStorageMsgsMap *concurrentMap[[]StoredMsg]
var tieredStorageMapLock sync.Mutex

func (s *Server) ListenForZombieConnCheckRequests() error {
	_, err := s.subscribeOnGlobalAcc(CONN_STATUS_SUBJ, CONN_STATUS_SUBJ+"_sid", func(_ *client, subject, reply string, msg []byte) {
		go func(msg []byte) {
			connInfo := &ConnzOptions{Limit: s.GlobalAccount().MaxActiveConnections()}
			conns, _ := s.Connz(connInfo)
			connectionIds := make(map[string]string)
			for _, conn := range conns.Conns {
				connId := strings.Split(conn.Name, "::")[0]
				if connId != "" {
					connectionIds[connId] = ""
				}
			}

			if len(connectionIds) > 0 { // in case there are connections
				bytes, err := json.Marshal(connectionIds)
				if err != nil {
					s.Errorf("ListenForZombieConnCheckRequests: " + err.Error())
				} else {
					s.sendInternalAccountMsgWithReply(s.GlobalAccount(), reply, _EMPTY_, nil, bytes, true)
				}
			}
		}(copyBytes(msg))
	})
	if err != nil {
		return err
	}
	return nil
}

func (s *Server) ListenForIntegrationsUpdateEvents() error {
	_, err := s.subscribeOnGlobalAcc(INTEGRATIONS_UPDATES_SUBJ, INTEGRATIONS_UPDATES_SUBJ+"_sid", func(_ *client, subject, reply string, msg []byte) {
		go func(msg []byte) {
			var integrationUpdate models.CreateIntegrationSchema
			err := json.Unmarshal(msg, &integrationUpdate)
			if err != nil {
				s.Errorf("ListenForIntegrationsUpdateEvents: " + err.Error())
				return
			}
			switch strings.ToLower(integrationUpdate.Name) {
			case "slack":
				if s.opts.UiHost == "" {
					EditClusterCompHost("ui_host", integrationUpdate.UIUrl)
				}
				CacheDetails("slack", integrationUpdate.Keys, integrationUpdate.Properties)
			case "s3":
				CacheDetails("s3", integrationUpdate.Keys, integrationUpdate.Properties)
			default:
				s.Warnf("ListenForIntegrationsUpdateEvents: %s %s", strings.ToLower(integrationUpdate.Name), "unknown integration")
				return
			}
		}(copyBytes(msg))
	})
	if err != nil {
		return err
	}
	return nil
}

func (s *Server) ListenForConfigReloadEvents() error {
	_, err := s.subscribeOnGlobalAcc(CONFIGURATIONS_RELOAD_SIGNAL_SUBJ, CONFIGURATIONS_RELOAD_SIGNAL_SUBJ+"_sid", func(_ *client, subject, reply string, msg []byte) {
		go func(msg []byte) {
			// reload config
			err := s.Reload()
			if err != nil {
				s.Errorf("Failed reloading: " + err.Error())
			}
		}(copyBytes(msg))
	})
	if err != nil {
		return err
	}
	return nil
}

func (s *Server) ListenForNotificationEvents() error {
	err := s.queueSubscribe(NOTIFICATION_EVENTS_SUBJ, NOTIFICATION_EVENTS_SUBJ+"_group", func(_ *client, subject, reply string, msg []byte) {
		go func(msg []byte) {
			var notification models.Notification
			err := json.Unmarshal(msg, &notification)
			if err != nil {
				s.Errorf("ListenForNotificationEvents: " + err.Error())
				return
			}
			notificationMsg := notification.Msg
			if notification.Code != "" {
				notificationMsg = notificationMsg + "\n```" + notification.Code + "```"
			}
			err = SendNotification(notification.Title, notificationMsg, notification.Type)
			if err != nil {
				return
			}
		}(copyBytes(msg))
	})
	if err != nil {
		return err
	}
	return nil
}

func (s *Server) ListenForPoisonMsgAcks() error {
	err := s.queueSubscribe(PM_RESEND_ACK_SUBJ, PM_RESEND_ACK_SUBJ+"_group", func(_ *client, subject, reply string, msg []byte) {
		go func(msg []byte) {
			var msgToAck models.PmAckMsg
			err := json.Unmarshal(msg, &msgToAck)
			if err != nil {
				s.Errorf("ListenForPoisonMsgAcks: " + err.Error())
				return
			}
			err = db.RemoveCgFromDlsMsg(msgToAck.ID, msgToAck.CgName)
			if err != nil {
				return
			}

		}(copyBytes(msg))
	})
	if err != nil {
		return err
	}
	return nil
}

func getThroughputSubject(serverName string) string {
	return throughputStreamNameV1 + tsep + serverName
}

func (s *Server) InitializeThroughputSampling() error {
	v, err := serv.Varz(nil)
	if err != nil {
		return err
	}

	LastReadThroughput = models.Throughput{
		Bytes:       v.OutBytes,
		BytesPerSec: 0,
	}
	LastWriteThroughput = models.Throughput{
		Bytes:       v.InBytes,
		BytesPerSec: 0,
	}

	go s.CalculateSelfThroughput()

	return nil
}

func (s *Server) CalculateSelfThroughput() error {
	for range time.Tick(time.Second * 1) {
		v, err := serv.Varz(nil)
		if err != nil {
			return err
		}

		currentWrite := v.InBytes - LastWriteThroughput.Bytes
		LastWriteThroughput = models.Throughput{
			Bytes:       v.InBytes,
			BytesPerSec: currentWrite,
		}
		currentRead := v.OutBytes - LastReadThroughput.Bytes
		LastReadThroughput = models.Throughput{
			Bytes:       v.OutBytes,
			BytesPerSec: currentRead,
		}
		serverName := s.opts.ServerName
		subj := getThroughputSubject(serverName)
		tpMsg := models.BrokerThroughput{
			Name:  serverName,
			Read:  currentRead,
			Write: currentWrite,
		}
		s.sendInternalAccountMsg(s.GlobalAccount(), subj, tpMsg)
	}

	return nil
}

func (s *Server) StartBackgroundTasks() error {
	s.ListenForPoisonMessages()
	err := s.ListenForZombieConnCheckRequests()
	if err != nil {
		return errors.New("Failed subscribing for zombie conns check requests: " + err.Error())
	}

	err = s.ListenForIntegrationsUpdateEvents()
	if err != nil {
		return errors.New("Failed subscribing for integrations updates: " + err.Error())
	}

	err = s.ListenForNotificationEvents()
	if err != nil {
		return errors.New("Failed subscribing for schema validation updates: " + err.Error())
	}

	err = s.ListenForPoisonMsgAcks()
	if err != nil {
		return errors.New("Failed subscribing for poison message acks: " + err.Error())
	}

	err = s.ListenForConfigReloadEvents()
	if err != nil {
		return errors.New("Failed subscribing for configurations update: " + err.Error())
	}

	// creating consumer + start listening
	err = s.ListenForTieredStorageMessages()
	if err != nil {
		return errors.New("Failed to subscribe for tiered storage messages" + err.Error())
	}

	err = s.ListenForSchemaverseDlsEvents()
	if err != nil {
		return errors.New("Failed to subscribe for schemaverse dls" + err.Error())
	}
	go s.RemoveOldDlsMsgs()

	// send JS API request to get more messages
	go s.sendPeriodicJsApiFetchTieredStorageMsgs()
	go s.uploadMsgsToTier2Storage()

	err = s.InitializeThroughputSampling()
	if err != nil {
		return err
	}
	return nil
}

func (s *Server) uploadMsgsToTier2Storage() {
	currentTimeFrame := s.opts.TieredStorageUploadIntervalSec
	ticker := time.NewTicker(time.Duration(currentTimeFrame) * time.Second)
	for range ticker.C {
		if s.opts.TieredStorageUploadIntervalSec != currentTimeFrame {
			currentTimeFrame = s.opts.TieredStorageUploadIntervalSec
			ticker.Reset(time.Duration(currentTimeFrame) * time.Second)
			// update consumer when TIERED_STORAGE_TIME_FRAME_SEC configuration was changed
			durableName := TIERED_STORAGE_CONSUMER
			tieredStorageTimeFrame := time.Duration(currentTimeFrame) * time.Second
			filterSubject := tieredStorageStream + ".>"
			cc := ConsumerConfig{
				DeliverPolicy: DeliverAll,
				AckPolicy:     AckExplicit,
				Durable:       durableName,
				FilterSubject: filterSubject,
				AckWait:       time.Duration(2) * tieredStorageTimeFrame,
				MaxAckPending: -1,
				MaxDeliver:    1,
			}
			err := serv.memphisAddConsumer(tieredStorageStream, &cc)
			if err != nil {
				serv.Errorf("Failed add tiered storage consumer: " + err.Error())
				return
			}
			TIERED_STORAGE_CONSUMER_CREATED = true

		}
		tieredStorageMapLock.Lock()
		if len(tieredStorageMsgsMap.m) > 0 {
			err := flushMapToTire2Storage()
			if err != nil {
				serv.Errorf("Failed upload messages to tiered 2 storage: " + err.Error())
				tieredStorageMapLock.Unlock()
				continue
			}
		}

		for i, msgs := range tieredStorageMsgsMap.m {
			for _, msg := range msgs {
				reply := msg.ReplySubject
				s.sendInternalAccountMsg(s.GlobalAccount(), reply, []byte(_EMPTY_))
			}
			tieredStorageMsgsMap.Delete(i)
		}
		tieredStorageMapLock.Unlock()
	}
}

func (s *Server) sendPeriodicJsApiFetchTieredStorageMsgs() {
	ticker := time.NewTicker(2 * time.Second)
	for range ticker.C {
		if TIERED_STORAGE_CONSUMER_CREATED && TIERED_STORAGE_STREAM_CREATED {
			durableName := TIERED_STORAGE_CONSUMER
			subject := fmt.Sprintf(JSApiRequestNextT, tieredStorageStream, durableName)
			reply := durableName + "_reply"
			amount := 1000
			req := []byte(strconv.FormatUint(uint64(amount), 10))
			serv.sendInternalAccountMsgWithReply(serv.GlobalAccount(), subject, reply, nil, req, true)
		}
	}
}

func (s *Server) ListenForTieredStorageMessages() error {
	tieredStorageMsgsMap = NewConcurrentMap[[]StoredMsg]()

	subject := TIERED_STORAGE_CONSUMER + "_reply"
	err := serv.queueSubscribe(subject, subject+"_sid", func(_ *client, subject, reply string, msg []byte) {
		go func(subject, reply string, msg []byte) {
			//Ignore 409 Exceeded MaxWaiting cases
			if reply != "" {
				rawMsg := strings.Split(string(msg), CR_LF+CR_LF)
				var tieredStorageMsg TieredStorageMsg
				if len(rawMsg) == 2 {
					err := json.Unmarshal([]byte(rawMsg[1]), &tieredStorageMsg)
					if err != nil {
						serv.Errorf("ListenForTieredStorageMessages: Failed unmarshalling tiered storage message: " + err.Error())
						return
					}
				} else {
					serv.Errorf("ListenForTieredStorageMessages: Invalid tiered storage message structure: message must contains msg-id header")
					return
				}
				payload := tieredStorageMsg.Buf
				replySubj := reply
				rawTs := tokenAt(reply, 8)
				seq, _, _ := ackReplyInfo(reply)
				intTs, err := strconv.Atoi(rawTs)
				if err != nil {
					serv.Errorf("ListenForTieredStorageMessages: Failed convert rawTs from string to int")
					return
				}

				dataFirstIdx := 0
				dataFirstIdx = getHdrLastIdxFromRaw(payload) + 1
				if dataFirstIdx > len(payload)-len(CR_LF) {
					s.Errorf("memphis error parsing")
					return
				}
				dataLen := len(payload) - dataFirstIdx
				header := payload[:dataFirstIdx]
				data := payload[dataFirstIdx : dataFirstIdx+dataLen]
				message := StoredMsg{
					Subject:      tieredStorageMsg.StationName,
					Sequence:     uint64(seq),
					Data:         data,
					Header:       header,
					Time:         time.Unix(0, int64(intTs)),
					ReplySubject: replySubj,
				}

				s.storeInTieredStorageMap(message)
			}
		}(subject, reply, copyBytes(msg))
	})
	if err != nil {
		serv.Errorf("ListenForTieredStorageMessages: Failed queueSubscribe tiered storage: " + err.Error())
		return err
	}

	return nil
}

func (s *Server) ListenForSchemaverseDlsEvents() error {
	err := s.queueSubscribe(SCHEMAVERSE_DLS_SUBJ, SCHEMAVERSE_DLS_SUBJ+"_group", func(_ *client, subject, reply string, msg []byte) {
		go func(msg []byte) {
			var message models.SchemaVerseDlsMessageSdk
			err := json.Unmarshal(msg, &message)
			if err != nil {
				serv.Errorf("ListenForSchemaverseDlsEvents: " + err.Error())
				return
			}

			exist, station, err := db.GetStationByName(message.StationName)
			if err != nil {
				serv.Errorf("ListenForSchemaverseDlsEvents: " + err.Error())
				return
			}
			if !exist {
				serv.Warnf("ListenForSchemaverseDlsEvents: station " + message.StationName + "couldn't been found")
				return

			}

			exist, p, err := db.GetProducerByNameAndConnectionID(message.Producer.Name, message.Producer.ConnectionId)
			if err != nil {
				serv.Errorf("ListenForSchemaverseDlsEvents: " + err.Error())
				return
			}

			if !exist {
				serv.Warnf("ListenForSchemaverseDlsEvents: producer " + p.Name + " couldn't been found")
				return
			}

			poisnedCgs := []string{}
			_, err = db.InsertPoisonedCgMessages(station.ID, 0, p.ID, poisnedCgs, models.MessagePayload(message.Message), "schema", message.ValidationError)
			if err != nil {
				serv.Errorf("ListenForSchemaverseDlsEvents: " + err.Error())
				return
			}
		}(copyBytes(msg))
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *Server) RemoveOldDlsMsgs() error {
	ticker := time.NewTicker(2 * time.Minute)
	for range ticker.C {
		configurationTime := time.Now().Add(time.Hour * time.Duration(-s.opts.DlsRetentionHours))
		err := db.DeleteOldDlsMessageByRetention(configurationTime)
		if err != nil {
			serv.Errorf("RemoveOldDlsMsgs: " + err.Error())
			return err
		}
	}
	return nil
}
