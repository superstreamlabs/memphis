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
	"encoding/json"
	"errors"
	"fmt"
	"memphis-broker/models"
	"reflect"
	"strconv"
	"strings"
)

func flushMapToTire2Storage() error {
	for k, f := range StorageFunctionsMap {
		switch k {
		case "s3":
			_, ok := IntegrationsCache["s3"].(models.Integration)
			if ok {
				err := f.(func() error)()
				if err != nil {
					return err
				}
			}
		default:
			return errors.New("Failed uploading to tiered storage : unsupported integration")
		}
	}
	return nil

}

func (s *Server) sendToTier2Storage(storageType interface{}, buf []byte, seq uint64, tierStorageType string) error {
	storedType := reflect.TypeOf(storageType).Elem().Name()
	var streamName string
	var tieredStorageEnabled bool
	switch storedType {
	case "fileStore":
		fileStore := storageType.(*fileStore)
		streamName = fileStore.cfg.StreamConfig.Name
		tieredStorageEnabled = fileStore.cfg.StreamConfig.TieredStorageEnabled
	case "memStore":
		memStore := storageType.(*memStore)
		streamName = memStore.cfg.Name
		tieredStorageEnabled = memStore.cfg.TieredStorageEnabled
	}

	fmt.Println("tieredStorageEnabled", tieredStorageEnabled)

	if tieredStorageEnabled {
		msgId := map[string]string{}
		seqNumber := strconv.Itoa(int(seq))
		msgId["msg-id"] = streamName + seqNumber

		_, ok := StorageFunctionsMap[tierStorageType]
		if ok {
			subject := fmt.Sprintf("%s.%s", tieredStorageStream, streamName)
			// TODO: if the stream is not exists save the messages in buffer
			if TIERED_STORAGE_STREAM_CREATED {
				tierStorageMsg := TieredStorageMsg{
					Buf:         buf,
					StationName: streamName,
				}

				msg, err := json.Marshal(tierStorageMsg)
				if err != nil {
					return err
				}
				s.sendInternalAccountMsgWithHeaders(s.GlobalAccount(), subject, msg, msgId)
			}
		}
	}
	return nil
}

func (s *Server) storeInTieredStorageMap(msg StoredMsg) {
	tieredStorageMapLock.Lock()
	stationName := msg.Subject
	if strings.Contains(msg.Subject, "#") {
		stationName = strings.Replace(msg.Subject, "#", ".", -1)
	}
	_, ok := tieredStorageMsgsMap.Load(stationName)
	if !ok {
		tieredStorageMsgsMap.Add(stationName, []StoredMsg{})
	}

	tieredStorageMsgsMap.m[stationName] = append(tieredStorageMsgsMap.m[stationName], msg)
	tieredStorageMapLock.Unlock()
}
