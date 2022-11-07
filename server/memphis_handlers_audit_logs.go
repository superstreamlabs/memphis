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
	"memphis-broker/models"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

type AuditLogsHandler struct{}

func CreateAuditLogs(auditLogs []interface{}) error {
	_, err := auditLogsCollection.InsertMany(context.TODO(), auditLogs)
	if err != nil {
		return err
	}
	return nil
}

func (ah AuditLogsHandler) GetAuditLogsByStation(station models.Station) ([]models.AuditLog, error) {
	var auditLogs []models.AuditLog

	cursor, err := auditLogsCollection.Find(context.TODO(), bson.M{"station_name": station.Name, "creation_date": bson.M{
		"$gte": (time.Now().AddDate(0, 0, -5)),
	}})
	if err != nil {
		return auditLogs, err
	}

	if err = cursor.All(context.TODO(), &auditLogs); err != nil {
		return auditLogs, err
	}

	if len(auditLogs) == 0 {
		auditLogs = []models.AuditLog{}
	}

	return auditLogs, nil
}

func RemoveAllAuditLogsByStation(stationName string) error {
	_, err := auditLogsCollection.DeleteMany(context.TODO(), bson.M{"station_name": stationName})
	if err != nil {
		return err
	}
	return nil
}
