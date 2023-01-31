// Copyright 2022-2023 The Memphis.dev Authors
// Licensed under the Memphis Business Source License 1.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// Changed License: [Apache License, Version 2.0 (https://www.apache.org/licenses/LICENSE-2.0), as published by the Apache Foundation.
//
// https://github.com/memphisdev/memphis-broker/blob/master/LICENSE
//
// Additional Use Grant: You may make use of the Licensed Work (i) only as part of your own product or service, provided it is not a message broker or a message queue product or service; and (ii) provided that you do not use, provide, distribute, or make available the Licensed Work as a Service.
// A "Service" is a commercial offering, product, hosted, or managed service, that allows third parties (other than your own employees and contractors acting on your behalf) to access and/or use the Licensed Work or a substantial set of the features or functionality of the Licensed Work to third parties as a software-as-a-service, platform-as-a-service, infrastructure-as-a-service or other similar services that compete with Licensor products or services.
package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"memphis-broker/models"

	"strconv"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var UI_url string

const CONN_STATUS_SUBJ = "$memphis_connection_status"
const INTEGRATIONS_UPDATES_SUBJ = "$memphis_integration_updates"
const CONFIGURATIONS_UPDATES_SUBJ = "$memphis_configurations_updates"
const NOTIFICATION_EVENTS_SUBJ = "$memphis_notifications"
const PM_RESEND_ACK_SUBJ = "$memphis_pm_acks"
const STORAGE_UPDATES_SUBJ = "$memphis_tiered_storage"

var LastReadThroughput models.Throughput
var LastWriteThroughput models.Throughput

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
	_, err := s.subscribeOnGlobalAcc(INTEGRATIONS_UPDATES_SUBJ, INTEGRATIONS_UPDATES_SUBJ+"_sid"+s.Name(), func(_ *client, subject, reply string, msg []byte) {
		go func(msg []byte) {
			var integrationUpdate models.CreateIntegrationSchema
			err := json.Unmarshal(msg, &integrationUpdate)
			if err != nil {
				s.Errorf("ListenForIntegrationsUpdateEvents: " + err.Error())
				return
			}
			switch strings.ToLower(integrationUpdate.Name) {
			case "slack":
				systemKeysCollection.UpdateOne(context.TODO(), bson.M{"key": "ui_url"},
					bson.M{"$set": bson.M{"value": integrationUpdate.UIUrl}})
				UI_url = integrationUpdate.UIUrl
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

func (s *Server) ListenForConfogurationsUpdateEvents() error {
	_, err := s.subscribeOnGlobalAcc(CONFIGURATIONS_UPDATES_SUBJ, CONFIGURATIONS_UPDATES_SUBJ+"_sid"+s.Name(), func(_ *client, subject, reply string, msg []byte) {
		go func(msg []byte) {
			var configurationsUpdate models.ConfigurationsUpdate
			err := json.Unmarshal(msg, &configurationsUpdate)
			if err != nil {
				s.Errorf("ListenForConfogurationsUpdateEvents: " + err.Error())
				return
			}
			switch strings.ToLower(configurationsUpdate.Type) {
			case "pm_retention":
				POISON_MSGS_RETENTION_IN_HOURS = int(configurationsUpdate.Update.(float64))
			default:
				return
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

func ackPoisonMsgV0(msgId string, cgName string) error {
	splitId := strings.Split(msgId, dlsMsgSep)
	stationName := splitId[0]
	sn, err := StationNameFromStr(stationName)
	if err != nil {
		return err
	}
	streamName := fmt.Sprintf(dlsStreamName, sn.Intern())
	uid := serv.memphis.nuid.Next()
	durableName := "$memphis_fetch_dls_consumer_" + uid
	amount := uint64(1)
	internalCgName := replaceDelimiters(cgName)
	filter := GetDlsSubject("poison", sn.Intern(), msgId, internalCgName)
	timeout := 30 * time.Second
	msgs, err := serv.memphisGetMessagesByFilter(streamName, filter, 0, amount, timeout)

	if len(msgs) != 1 {
		return errors.New("message was not found")
	}

	msg := msgs[0]
	var dlsMsg models.DlsMessage
	err = json.Unmarshal(msg.Data, &dlsMsg)
	if err != nil {
		return err
	}

	err = serv.memphisRemoveConsumer(streamName, durableName)
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
			//This check for backward compatability
			if msgToAck.CgName != "" {
				err = ackPoisonMsgV0(msgToAck.ID, msgToAck.CgName)
				if err != nil {
					s.Errorf("ListenForPoisonMsgAcks: " + err.Error())
					return
				}
			} else {
				splitId := strings.Split(msgToAck.ID, dlsMsgSep)
				stationName := splitId[0]
				sn, err := StationNameFromStr(stationName)
				if err != nil {
					s.Errorf("ListenForPoisonMsgAcks: " + err.Error())
					return
				}
				streamName := fmt.Sprintf(dlsStreamName, sn.Intern())
				seq, err := strconv.ParseInt(msgToAck.Sequence, 10, 64)
				if err != nil {
					s.Errorf("ListenForPoisonMsgAcks: " + err.Error())
					return
				}
				_, err = s.memphisDeleteMsgFromStream(streamName, uint64(seq))
				if err != nil {
					s.Errorf("ListenForPoisonMsgAcks: " + err.Error())
					return
				}
			}

		}(copyBytes(msg))
	})
	if err != nil {
		return err
	}
	return nil
}

func getThroughputSubject(serverName string) string {
	key := serverName
	if key == _EMPTY_ {
		key = "broker"
	}
	return throughputStreamName + tsep + key
}

func (s *Server) InitializeThroughputSampling() error {
	v, err := serv.Varz(nil)
	if err != nil {
		return err
	}

	LastReadThroughput = models.Throughput{
		Bytes:       v.InBytes,
		BytesPerSec: 0,
	}
	LastWriteThroughput = models.Throughput{
		Bytes:       v.OutBytes,
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

		currentWrite := v.OutBytes - LastWriteThroughput.Bytes
		LastWriteThroughput = models.Throughput{
			Bytes:       v.OutBytes,
			BytesPerSec: currentWrite,
		}
		currentRead := v.InBytes - LastReadThroughput.Bytes
		LastReadThroughput = models.Throughput{
			Bytes:       v.InBytes,
			BytesPerSec: currentRead,
		}
		subj := getThroughputSubject(configuration.SERVER_NAME)
		tpMsg := models.BrokerThroughput{
			Name:  configuration.SERVER_NAME,
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

	err = s.ListenForConfogurationsUpdateEvents()
	if err != nil {
		return errors.New("Failed subscribing for confogurations update: " + err.Error())
	}
	err = s.ListenForTierStorageMessages()
	if err != nil {
		return errors.New("Failed subscribing for tiered storage update: " + err.Error())
	}

	filter := bson.M{"key": "ui_url"}
	var systemKey models.SystemKey
	err = systemKeysCollection.FindOne(context.TODO(), filter).Decode(&systemKey)
	if err == mongo.ErrNoDocuments {
		UI_url = ""
		uiUrlKey := models.SystemKey{
			ID:    primitive.NewObjectID(),
			Key:   "ui_url",
			Value: "",
		}

		_, err = systemKeysCollection.InsertOne(context.TODO(), uiUrlKey)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	} else {
		UI_url = systemKey.Value
	}

	err = s.InitializeThroughputSampling()
	if err != nil {
		return err
	}
	return nil
}

func (s *Server) ConsumeStorageMsgs(durableName string, errs chan error) {
	timeout := time.Duration(configuration.MAX_ACK_TIME_SECONDS_STORAGE) * time.Second
	ping := 5 * time.Second

	for {
		var msgs []StoredMsg

		select {
		case <-time.After(ping):
			streamInfo, err := serv.memphisStreamInfo(STORAGE_UPDATES_SUBJ)
			if err != nil {
				serv.Errorf("ConsumeStorageMsgs: " + err.Error())
				errs <- err
				return
			}

			if streamInfo.State.Msgs == 0 {
				ping = ping + (20 * time.Second)
				continue
			}
			responseChan := make(chan StoredMsg)
			subject := fmt.Sprintf(JSApiRequestNextT, STORAGE_UPDATES_SUBJ, durableName)
			reply := durableName + "_reply"
			amount := 1000
			req := []byte(strconv.FormatUint(uint64(amount), 10))
			sub, err := serv.subscribeOnGlobalAcc(reply, reply+"_sid", func(_ *client, subject, reply string, msg []byte) {
				go func(respCh chan StoredMsg, subject, reply string, msg []byte) {
					replySubjs := reply
					rawTs := tokenAt(reply, 8)
					seq, _, _ := ackReplyInfo(reply)
					intTs, err := strconv.Atoi(rawTs)
					if err != nil {
						serv.Errorf("ConsumeStorageMsgs: " + err.Error())
						errs <- err
						return
					}

					dataFirstIdx := 0
					dataLen := len(msg)
					dataFirstIdx = getHdrLastIdxFromRaw(msg) + 1
					if dataFirstIdx > len(msg)-len(CR_LF) {
						s.Errorf("ConsumeStorageMsgs: memphis error parsing in station get messages")
					}

					dataLen = len(msg) - dataFirstIdx
					dataLen -= len(CR_LF)
					header := msg[:dataFirstIdx]
					data := msg[dataFirstIdx : dataFirstIdx+dataLen]
					respCh <- StoredMsg{
						Subject:       subject,
						Sequence:      uint64(seq),
						Data:          data,
						Header:        header,
						Time:          time.Unix(0, int64(intTs)),
						ReplySubjects: replySubjs,
					}
				}(responseChan, subject, reply, copyBytes(msg))
			})
			if err != nil {
				serv.Errorf("ConsumeStorageMsgs: " + err.Error())
				errs <- err
				return
			}

			serv.sendInternalAccountMsgWithReply(serv.GlobalAccount(), subject, reply, nil, req, true)
			timer := time.NewTimer(time.Duration(timeout))

			go func() {
				for i := 0; i < amount; i++ {
					select {
					case <-timer.C:
						timer.Stop()
						serv.unsubscribeOnGlobalAcc(sub)
						err := s.uploadToS3Storage(msgs)
						if err != nil {
							serv.Errorf("ConsumeStorageMsgs: " + err.Error())
							errs <- err
							break
						}
						// ack
						for _, msg := range msgs {
							reply := msg.ReplySubjects
							serv.sendInternalAccountMsg(serv.GlobalAccount(), reply, []byte(_EMPTY_))
						}
					case msg := <-responseChan:
						msgs = append(msgs, msg)
					}
				}
			}()

		}
		time.Sleep(5 * time.Second)
	}
}

func (s *Server) ListenForTierStorageMessages() error {

	durableName := "storage_consumer"
	cc := ConsumerConfig{
		DeliverPolicy: DeliverAll,
		AckPolicy:     AckExplicit,
		Durable:       durableName,
		FilterSubject: "$memphis_tiered_storage.>",
		AckWait:       time.Duration(configuration.MAX_ACK_TIME_SECONDS_STORAGE) * time.Second,
	}
	err := serv.memphisAddConsumer(STORAGE_UPDATES_SUBJ, &cc)
	if err != nil {
		return err
	}

	chErrs := make(chan error, 1)
	go s.ConsumeStorageMsgs(durableName, chErrs)
	err = <-chErrs
	if err != nil {
		s.Errorf("ListenForTierStorageMessages" + err.Error())
		return err
	}

	return nil
}
