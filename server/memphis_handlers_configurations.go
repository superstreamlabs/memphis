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
	"memphis/analytics"
	"memphis/db"
	"memphis/models"
	"memphis/utils"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type ConfigurationsHandler struct{ S *Server }

func (ch ConfigurationsHandler) GetClusterConfig(c *gin.Context) {
	shouldSendAnalytics, _ := shouldSendAnalytics()
	if shouldSendAnalytics {
		user, _ := getUserDetailsFromMiddleware(c)
		analytics.SendEvent(user.Username, "user-enter-cluster-config-page")
	}
	c.IndentedJSON(200, gin.H{
		"dls_retention":           ch.S.opts.DlsRetentionHours,
		"logs_retention":          ch.S.opts.LogsRetentionDays,
		"broker_host":             ch.S.opts.BrokerHost,
		"ui_host":                 ch.S.opts.UiHost,
		"rest_gw_host":            ch.S.opts.RestGwHost,
		"tiered_storage_time_sec": ch.S.opts.TieredStorageUploadIntervalSec,
		"max_msg_size_mb":         ch.S.opts.MaxPayload / 1024 / 1024,
	})
}

func (ch ConfigurationsHandler) EditClusterConfig(c *gin.Context) {
	// if err := DenyForSandboxEnv(c); err != nil {
	// 	return
	// }

	var body models.EditClusterConfigSchema
	ok := utils.Validate(c, &body, false, nil)
	if !ok {
		return
	}
	if ch.S.opts.DlsRetentionHours != body.DlsRetention {
		err := changeDlsRetention(body.DlsRetention)
		if err != nil {
			serv.Errorf("EditConfigurations: " + err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}
	}
	if ch.S.opts.LogsRetentionDays != body.LogsRetention {
		err := changeLogsRetention(body.LogsRetention)
		if err != nil {
			serv.Errorf("EditConfigurations: " + err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}
	}
	if ch.S.opts.TieredStorageUploadIntervalSec != body.TSTimeSec {
		if body.TSTimeSec > 3600 || body.TSTimeSec < 5 {
			serv.Errorf("EditConfigurations: Tiered storage time can't be less than 5 seconds or more than 60 minutes")
			c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": "Tiered storage time can't be less than 5 seconds or more than 60 minutes"})
		} else {
			err := changeTSTime(body.TSTimeSec)
			if err != nil {
				serv.Errorf("EditConfigurations: " + err.Error())
				c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
				return
			}
		}
	}

	brokerHost := strings.ToLower(body.BrokerHost)
	if ch.S.opts.BrokerHost != brokerHost {
		err := EditClusterCompHost("broker_host", brokerHost)
		if err != nil {
			serv.Errorf("EditConfigurations: " + err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}
	}

	uiHost := strings.ToLower(body.UiHost)
	if ch.S.opts.UiHost != uiHost {
		err := EditClusterCompHost("ui_host", uiHost)
		if err != nil {
			serv.Errorf("EditConfigurations: " + err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}
	}

	restGWHost := strings.ToLower(body.RestGWHost)
	if ch.S.opts.RestGwHost != restGWHost {
		err := EditClusterCompHost("rest_gw_host", restGWHost)
		if err != nil {
			serv.Errorf("EditConfigurations: " + err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}
	}

	if ch.S.opts.MaxPayload != int32(body.MaxMsgSizeMb) {
		err := changeMaxMsgSize(body.MaxMsgSizeMb)
		if err != nil {
			serv.Errorf("EditConfigurations: " + err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}
	}

	// send signal to reload config
	err := serv.sendInternalAccountMsgWithReply(serv.GlobalAccount(), CONFIGURATIONS_RELOAD_SIGNAL_SUBJ, _EMPTY_, nil, _EMPTY_, true)
	if err != nil {
		serv.Errorf("EditConfigurations: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	shouldSendAnalytics, _ := shouldSendAnalytics()
	if shouldSendAnalytics {
		user, _ := getUserDetailsFromMiddleware(c)
		analytics.SendEvent(user.Username, "user-update-cluster-config")
	}

	c.IndentedJSON(200, gin.H{
		"dls_retention":           ch.S.opts.DlsRetentionHours,
		"logs_retention":          ch.S.opts.LogsRetentionDays,
		"broker_host":             ch.S.opts.BrokerHost,
		"ui_host":                 ch.S.opts.UiHost,
		"rest_gw_host":            ch.S.opts.RestGwHost,
		"tiered_storage_time_sec": ch.S.opts.TieredStorageUploadIntervalSec,
		"max_msg_size_mb":         ch.S.opts.MaxPayload / 1024 / 1024,
	})
}

func changeDlsRetention(dlsRetention int) error {
	err := db.UpsertConfiguration("dls_retention", strconv.Itoa(dlsRetention), strings.ToLower(db.GlobalTenantName))
	if err != nil {
		return err
	}
	return nil
}

func changeLogsRetention(logsRetention int) error {
	err := db.UpsertConfiguration("logs_retention", strconv.Itoa(logsRetention), strings.ToLower(db.GlobalTenantName))
	if err != nil {
		return err
	}

	retentionDur := time.Duration(logsRetention) * time.Hour * 24
	err = serv.memphisUpdateStream(&StreamConfig{
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
	err := db.UpsertConfiguration("tiered_storage_time_sec", strconv.Itoa(tsTime), strings.ToLower(db.GlobalTenantName))
	if err != nil {
		return err
	}

	return nil
}

func EditClusterCompHost(key string, host string) error {
	key = strings.ToLower(key)
	host = strings.ToLower(host)
	err := db.UpsertConfiguration(key, host, strings.ToLower(db.GlobalTenantName))
	if err != nil {
		return err
	}

	return nil
}

func changeMaxMsgSize(newSize int) error {
	err := db.UpsertConfiguration("max_msg_size_mb", strconv.Itoa(newSize), strings.ToLower(db.GlobalTenantName))
	if err != nil {
		return err
	}

	return nil
}
