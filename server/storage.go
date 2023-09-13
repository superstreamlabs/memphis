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
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/memphisdev/memphis/models"
)

func flushMapToTier2Storage() error {
	for t, tenant := range tieredStorageMsgsMap.m {
		if IsStorageLimitExceeded(t) {
			serv.Warnf("[tenant:%s]flushMapToTier2Storage: %s", t, ErrUpgradePlan.Error())
			continue
		}
		if ValidataAccessToFeature(t, "feature-storage-tiering") {
			for k, f := range StorageFunctionsMap {
				switch k {
				case "s3":
					if tenantIntegrations, ok := IntegrationsConcurrentCache.Load(t); !ok {
						continue
					} else {
						if _, ok = tenantIntegrations["s3"].(models.Integration); ok {
							err := f.(func(string, map[string][]StoredMsg) error)(t, tenant)
							if err != nil {
								return err
							}
						}
					}
				default:
					return errors.New("failed uploading to tiered storage : unsupported integration")
				}
			}
		} else {
			serv.Warnf("[tenant: %v]flushMapToTier2Storage: Has no access to feature-storage-tiering in its Plan", t)
		}
	}
	return nil

}

func (s *Server) sendToTier2Storage(storageType interface{}, buf []byte, seq uint64, tierStorageType string) error {
	storedType := reflect.TypeOf(storageType).Elem().Name()
	var streamName, tenantName string
	switch storedType {
	case "fileStore":
		fileStore := storageType.(*fileStore)
		streamName = fileStore.cfg.StreamConfig.Name
		tenantName = fileStore.account.Name
	case "memStore":
		memStore := storageType.(*memStore)
		streamName = memStore.cfg.Name
		tenantName = memStore.account.Name
	}

	if IsStorageLimitExceeded(tenantName) {
		serv.Warnf("[tenant:%s]sendToTier2Storage: %s", tenantName, ErrUpgradePlan.Error())
		return nil
	}

	for k := range StorageFunctionsMap {
		if ValidataAccessToFeature(tenantName, "feature-storage-tiering") {
			switch k {
			case "s3":
				if tenantIntegrations, ok := IntegrationsConcurrentCache.Load(tenantName); !ok {
					continue
				} else {
					if _, ok = tenantIntegrations["s3"].(models.Integration); ok {
						msgId := map[string]string{}
						seqNumber := strconv.Itoa(int(seq))
						msgId["msg-id"] = streamName + seqNumber
						if tenantName == "" {
							tenantName = serv.MemphisGlobalAccountString()
						}
						subject := fmt.Sprintf("%s.%s.%s", tieredStorageStream, streamName, tenantName)
						// TODO: if the stream is not exists save the messages in buffer
						if TIERED_STORAGE_STREAM_CREATED {
							tierStorageMsg := TieredStorageMsg{
								Buf:         buf,
								StationName: streamName,
								TenantName:  tenantName,
							}

							msg, err := json.Marshal(tierStorageMsg)
							if err != nil {
								return err
							}
							s.sendInternalAccountMsgWithHeadersWithEcho(s.MemphisGlobalAccount(), subject, msg, msgId)
						}
					}
				}
			default:
				return errors.New("failed send to tiered storage : unsupported integration")
			}
		}
	}
	return nil
}

func (s *Server) storeInTieredStorageMap(msg StoredMsg) {
	tieredStorageMapLock.Lock()
	stationName := msg.Subject
	tenantName := msg.TenantName
	if strings.Contains(msg.Subject, "#") {
		stationName = strings.Replace(msg.Subject, "#", ".", -1)
	}
	_, ok := tieredStorageMsgsMap.Load(tenantName)
	if !ok {
		tieredStorageMsgsMap.Add(tenantName, map[string][]StoredMsg{})
	}

	tieredStorageMsgsMap.m[tenantName][stationName] = append(tieredStorageMsgsMap.m[tenantName][stationName], msg)
	tieredStorageMapLock.Unlock()
}

func (s *Server) handleNewTieredStorageMsg(msg []byte, reply string) {
	rawMsg := strings.Split(string(msg), CR_LF+CR_LF)
	var tieredStorageMsg TieredStorageMsg
	if len(rawMsg) == 2 {
		err := json.Unmarshal([]byte(rawMsg[1]), &tieredStorageMsg)
		if err != nil {
			s.Errorf("ListenForTieredStorageMessages: Failed unmarshalling tiered storage message: " + err.Error())
			return
		}
	} else {
		s.Errorf("ListenForTieredStorageMessages: Invalid tiered storage message structure: message must contains msg-id header")
		return
	}
	payload := tieredStorageMsg.Buf
	rawTs := tokenAt(reply, 8)
	seq, _, _ := ackReplyInfo(reply)
	intTs, err := strconv.Atoi(rawTs)
	if err != nil {
		s.Errorf("ListenForTieredStorageMessages: Failed convert rawTs from string to int")
		return
	}

	dataFirstIdx := 0
	dataFirstIdx = getHdrLastIdxFromRaw(payload) + 1
	if dataFirstIdx > len(payload)-len(CR_LF) {
		s.Errorf("ListenForTieredStorageMessages: memphis error while parsing headers")
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
		ReplySubject: reply,
		TenantName:   tieredStorageMsg.TenantName,
	}

	s.storeInTieredStorageMap(message)
}
