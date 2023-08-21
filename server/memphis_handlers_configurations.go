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
	"strconv"
	"strings"
	"time"

	"github.com/memphisdev/memphis/db"
)

type ConfigurationsHandler struct{ S *Server }

func changeDlsRetention(dlsRetention int, tenantName string) error {
	err := db.UpsertConfiguration("dls_retention", strconv.Itoa(dlsRetention), tenantName)
	if err != nil {
		return err
	}
	return nil
}

func changeGCProducersConsumersRetentionHours(retention int, tenantName string) error {
	err := db.UpsertConfiguration("gc_producer_consumer_retention_hours", strconv.Itoa(retention), tenantName)
	if err != nil {
		return err
	}
	return nil
}

func changeLogsRetention(logsRetention int) error {
	err := db.UpsertConfiguration("logs_retention", strconv.Itoa(logsRetention), serv.MemphisGlobalAccountString())
	if err != nil {
		return err
	}

	retentionDur := time.Duration(logsRetention) * time.Hour * 24
	err = serv.memphisUpdateStream(serv.MemphisGlobalAccountString(), &StreamConfig{
		Name:         syslogsStreamName,
		Subjects:     []string{syslogsStreamName + ".>"},
		Retention:    LimitsPolicy,
		MaxAge:       retentionDur,
		MaxConsumers: -1,
		Discard:      DiscardOld,
		Storage:      FileStorage,
	})
	if err != nil {
		return err
	}
	return nil
}

func changeTSTime(tsTime int) error {
	err := db.UpsertConfiguration("tiered_storage_time_sec", strconv.Itoa(tsTime), serv.MemphisGlobalAccountString())
	if err != nil {
		return err
	}

	return nil
}

func EditClusterCompHost(key string, host string) error {
	key = strings.ToLower(key)
	host = strings.ToLower(host)
	err := db.UpsertConfiguration(key, host, serv.MemphisGlobalAccountString())
	if err != nil {
		return err
	}

	return nil
}

func changeMaxMsgSize(newSize int) error {
	err := db.UpsertConfiguration("max_msg_size_mb", strconv.Itoa(newSize), serv.MemphisGlobalAccountString())
	if err != nil {
		return err
	}

	return nil
}
