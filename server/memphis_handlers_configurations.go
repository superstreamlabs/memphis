// Credit for The NATS.IO Authors
// Copyright 2021-2022 The Memphis Authors
// Licensed under the Apache License, Version 2.0 (the “License”);
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an “AS IS” BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.package server
package server

import (
	"context"
	"encoding/json"
	"memphis-broker/models"
	"memphis-broker/utils"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ConfigurationsHandler struct{}

var userMgmtHandler UserMgmtHandler

func (s *Server) initializeConfigurations() {
	var pmRetention models.ConfigurationsIntValue
	err := configurationsCollection.FindOne(context.TODO(), bson.M{"key": "pm_retention"}).Decode(&pmRetention)
	if err != nil {
		if err != mongo.ErrNoDocuments {
			s.Errorf("initializeConfigurations error: " + err.Error())
		}
		POISON_MSGS_RETENTION_IN_HOURS = configuration.POISON_MSGS_RETENTION_IN_HOURS
	} else {
		POISON_MSGS_RETENTION_IN_HOURS = pmRetention.Value
	}
	var logsRetention models.ConfigurationsIntValue
	err = configurationsCollection.FindOne(context.TODO(), bson.M{"key": "logs_retention"}).Decode(&logsRetention)
	if err != nil {
		if err != mongo.ErrNoDocuments {
			s.Errorf("initializeConfigurations error: " + err.Error())
		}
		LOGS_RETENTION_IN_DAYS, err = strconv.Atoi(configuration.LOGS_RETENTION_IN_DAYS)
		if err != nil {
			s.Errorf("initializeConfigurations error: " + err.Error())
			LOGS_RETENTION_IN_DAYS = 30 //default
		}
	} else {
		LOGS_RETENTION_IN_DAYS = logsRetention.Value
	}
}

func (ch ConfigurationsHandler) EditClusterConfig(c *gin.Context) {
	var body models.EditClusterConfigSchema
	ok := utils.Validate(c, &body, false, nil)
	if !ok {
		return
	}
	if POISON_MSGS_RETENTION_IN_HOURS != body.PMRetention {
		err := changePMRetention(body.PMRetention)
		if err != nil {
			serv.Errorf("EditConfigurations error: " + err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}
	}
	if LOGS_RETENTION_IN_DAYS != body.LogsRetention {
		err := changeLogsRetention(body.LogsRetention)
		if err != nil {
			serv.Errorf("EditConfigurations error: " + err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}
	}

	c.IndentedJSON(200, gin.H{"pm_retention": POISON_MSGS_RETENTION_IN_HOURS, "logs_retention": LOGS_RETENTION_IN_DAYS})
}

func changePMRetention(pmRetention int) error {
	POISON_MSGS_RETENTION_IN_HOURS = pmRetention
	msg, err := json.Marshal(models.ConfigurationsUpdate{Type: "pm_retention", Update: POISON_MSGS_RETENTION_IN_HOURS})
	if err != nil {
		return err
	}
	err = serv.sendInternalAccountMsgWithReply(serv.GlobalAccount(), CONFIGURATIONS_UPDATES_SUBJ, _EMPTY_, nil, msg, true)
	if err != nil {
		return err
	}
	filter := bson.M{"key": "pm_retention"}
	update := bson.M{
		"$set": bson.M{
			"value": POISON_MSGS_RETENTION_IN_HOURS,
		},
	}
	opts := options.Update().SetUpsert(true)
	_, err = configurationsCollection.UpdateOne(context.TODO(), filter, update, opts)
	if err != nil {
		return err
	}
	return nil
}

func changeLogsRetention(logsRetention int) error {
	LOGS_RETENTION_IN_DAYS = logsRetention
	filter := bson.M{"key": "logs_retention"}
	update := bson.M{
		"$set": bson.M{
			"value": LOGS_RETENTION_IN_DAYS,
		},
	}
	opts := options.Update().SetUpsert(true)
	_, err := configurationsCollection.UpdateOne(context.TODO(), filter, update, opts)
	if err != nil {
		return err
	}
	retentionDur := time.Duration(LOGS_RETENTION_IN_DAYS) * time.Hour * 24
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

func (ch ConfigurationsHandler) GetClusterConfig(c *gin.Context) {
	c.IndentedJSON(200, gin.H{"pm_retention": POISON_MSGS_RETENTION_IN_HOURS, "logs_retention": LOGS_RETENTION_IN_DAYS})
}
