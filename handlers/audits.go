// Copyright 2021-2022 The Memphis Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package handlers

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"

	"memphis-control-plane/logger"
	"memphis-control-plane/models"
	"memphis-control-plane/utils"
)

type AuditlogsHandler struct{}


func CreateAuditLogs(auditLogs []interface{}){
	_, err := auditLogsCollection.InsertMany(context.TODO(), auditLogs)
	if err != nil {
		logger.Error("CreateAuditLogs error: " + err.Error())
		return
	}
}

func (ah AuditlogsHandler) GetAllAuditLogsByStation(c *gin.Context) {
	var body models.GetAllAuditLogsByStationSchema
	ok := utils.Validate(c, &body, false, nil)
	if !ok{
		return
	}

	exist, station, err := IsStationExist(body.StationName)
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	if !exist {
		c.AbortWithStatusJSON(configuration.SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": "Station does not exist"})
		return
	}

	var auditLogs []models.AuditLog
	cursor, err := auditLogsCollection.Find(context.TODO(), bson.M{"station_name": station.Name, "creation_date": bson.M{
		"$gte": (time.Now().AddDate(0, 0, -30)),
	},})
	if err != nil {
		logger.Warn("GetAllAuditLogsByStation error: " + err.Error())
	}

	if err = cursor.All(context.TODO(), &auditLogs); err != nil {
		logger.Warn("GetAllAuditLogsByStation error: " + err.Error())
	}

	if len(auditLogs) == 0 {
		c.IndentedJSON(200, []string{})
	} else {
		c.IndentedJSON(200, auditLogs)
	}
}

func RemoveAllAuditLogsByStation(stationName string) {
	_, err := auditLogsCollection.DeleteMany(context.TODO(), bson.M{"station_name": stationName})
	if err != nil {
		logger.Warn("RemoveAllAuditLogsByStation error: " + err.Error())
		return
	}
}
