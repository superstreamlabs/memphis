// Credit for The NATS.IO Authors
// Copyright 2021-2022 The Memphis Authors
// Licensed under the MIT License (the "License");
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

// This license limiting reselling the software itself "AS IS".

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.
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
