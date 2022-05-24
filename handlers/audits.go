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

type AuditsHandler struct{}


func CreateAudits(audits []interface{}){
	_, err := auditsCollection.InsertMany(context.TODO(), audits)
	if err != nil {
		logger.Error("CreateAudits error: " + err.Error())
		return
	}
}

func (ah AuditsHandler) GetAllAuditsByStation(c *gin.Context) {
	var body models.GetAllAuditsByStationSchema
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

	var audits []models.Audit
	cursor, err := auditsCollection.Find(context.TODO(), bson.M{"station_name": station.Name, "creation_date": bson.M{
		"$gte": (time.Now().AddDate(0, 0, -30)),
	},})
	if err != nil {
		logger.Warn("GetAllAuditsByStation error: " + err.Error())
	}

	if err = cursor.All(context.TODO(), &audits); err != nil {
		logger.Warn("GetAllAuditsByStation error: " + err.Error())
	}

	if len(audits) == 0 {
		c.IndentedJSON(200, []string{})
	} else {
		c.IndentedJSON(200, audits)
	}
}

func RemoveAllAuditsByStation(stationName string) {
	_, err := auditsCollection.DeleteMany(context.TODO(), bson.M{"station_name": stationName})
	if err != nil {
		logger.Warn("RemoveAllAuditsByStation error: " + err.Error())
		return
	}
}
