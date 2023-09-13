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
	"errors"
	"fmt"
	"regexp"
	"strings"
	"sync"

	"github.com/memphisdev/memphis/conf"
	"github.com/memphisdev/memphis/db"
	"github.com/memphisdev/memphis/models"

	"github.com/gin-gonic/gin"
	"github.com/nats-io/nuid"
)

type Handlers struct {
	Producers      ProducersHandler
	Consumers      ConsumersHandler
	AuditLogs      AuditLogsHandler
	Stations       StationsHandler
	Monitoring     MonitoringHandler
	PoisonMsgs     PoisonMessagesHandler
	Tags           TagsHandler
	Schemas        SchemasHandler
	Integrations   IntegrationsHandler
	Configurations ConfigurationsHandler
	Tenants        TenantHandler
	Billing        BillingHandler
	userMgmt       UserMgmtHandler
	AsyncTasks     AsyncTasksHandler
	Functions      FunctionsHandler
}

var serv *Server
var configuration = conf.GetConfig()

type srvMemphis struct {
	serverID               string
	nuid                   *nuid.NUID
	activateSysLogsPubFunc func()
	fallbackLogQ           *ipQueue[fallbackLog]
	jsApiMu                *BufferedMutex
	ws                     memphisWS
}

type memphisWS struct {
	subscriptions *concurrentMap[memphisWSReqTenantsToFiller]
	quitCh        chan struct{}
}

type BufferedMutex struct {
	mu     sync.Mutex
	sem    chan struct{}
	buffer int
}

func NewBufferedMutex(buffer int) *BufferedMutex {
	if buffer <= 0 {
		buffer = 1 // Ensure at least one lock can be acquired
	}
	return &BufferedMutex{
		sem:    make(chan struct{}, buffer),
		buffer: buffer,
	}
}

func (bm *BufferedMutex) Lock() {
	bm.sem <- struct{}{}
	bm.mu.Lock()
}

func (bm *BufferedMutex) Unlock() {
	bm.mu.Unlock()
	<-bm.sem
}

func (s *Server) InitializeMemphisHandlers() {
	serv = s
	s.memphis.nuid = nuid.New()

	s.initializeSDKHandlers()
	s.initWS()

}

func getUserDetailsFromMiddleware(c *gin.Context) (models.User, error) {
	user, _ := c.Get("user")
	userModel := user.(models.User)
	if len(userModel.Username) == 0 {
		return userModel, errors.New("username is empty")
	}
	return userModel, nil
}

func CreateDefaultStation(tenantName string, s *Server, sn StationName, userId int, username, schemaName string, schemaVersionNumber int) (models.Station, bool, error) {
	stationName := sn.Ext()
	replicas := getDefaultReplicas()
	err := s.CreateStream(tenantName, sn, "message_age_sec", 3600, "file", 120000, replicas, false, 1)
	if err != nil {
		return models.Station{}, false, err
	}

	newStation, rowsUpdated, err := db.InsertNewStation(stationName, userId, username, "message_age_sec", 3600, "file", replicas, schemaName, schemaVersionNumber, 120000, true, models.DlsConfiguration{Poison: true, Schemaverse: true}, false, tenantName, []int{1}, 1)
	if err != nil {
		return models.Station{}, false, err
	}
	if rowsUpdated == 0 {
		return models.Station{}, false, nil
	}

	err = CreateDefaultTags("station", newStation.ID, tenantName)
	if err != nil {
		return models.Station{}, false, err
	}

	return newStation, true, nil
}

func CreateDefaultSchema(username, tenantName string, userId int) (string, error) {
	defaultSchemaName := "demo-schema"
	defualtSchemaType := "json"
	defualtSchemaContent := `{
		"$id": "https://example.com/address.schema.json",
		"description": "An address similar to http://microformats.org/wiki/h-card",
		"type": "object",
		"properties": {
			"post-office-box": {
			"type": "number"
			},
			"extended-address": {
			"type": "string"
			},
			"street-address": {
			"type": "string"
			},
			"locality": {
			"type": "string"
			},
			"region": {
			"type": "string"
			},
			"postal-code": {
			"type": "string"
			},
			"country-name": {
			"type": "string"
			}
		},
		"required": [ "locality" ]
	}`
	newSchema, rowsUpdated, err := db.InsertNewSchema(defaultSchemaName, defualtSchemaType, username, tenantName)
	if err != nil {
		return "", err
	}

	if rowsUpdated == 1 {
		_, _, err = db.InsertNewSchemaVersion(1, userId, username, defualtSchemaContent, newSchema.ID, "", "", true, tenantName)
		if err != nil {
			return "", err
		}
	} else {
		errMsg := fmt.Sprintf("Schema with the name %v already exists ", newSchema.Name)
		return "", fmt.Errorf(errMsg)
	}

	err = CreateDefaultTags("schema", newSchema.ID, tenantName)
	if err != nil {
		return "", err
	}
	return newSchema.Name, nil
}

func CreateDefaultTags(tagType string, id int, tenantName string) error {
	defaultTags := models.CreateTag{Name: "default"}
	color := "0, 165, 255"
	err := AddTagsToEntity([]models.CreateTag{defaultTags}, tagType, id, tenantName, color)
	if err != nil {
		return err
	}
	return nil
}

func validateName(name, objectType string) error {
	emptyErrStr := fmt.Sprintf("%v name can not be empty", objectType)
	tooLongErrStr := fmt.Sprintf("%v should be under 128 characters", objectType)
	invalidCharErrStr := fmt.Sprintf("Only alphanumeric and the '_', '-', '.' characters are allowed in %v", objectType)
	firstLetterErrStr := fmt.Sprintf("%v name can not start or end with non alphanumeric character", objectType)

	emptyErr := errors.New(emptyErrStr)
	tooLongErr := errors.New(tooLongErrStr)
	invalidCharErr := errors.New(invalidCharErrStr)
	firstLetterErr := errors.New(firstLetterErrStr)

	if len(name) == 0 {
		return emptyErr
	}

	if len(name) > 128 {
		return tooLongErr
	}

	re := regexp.MustCompile("^[a-z0-9_.-]*$")

	validName := re.MatchString(name)
	if !validName {
		return invalidCharErr
	}

	if name[0:1] == "." || name[0:1] == "-" || name[0:1] == "_" || name[len(name)-1:] == "." || name[len(name)-1:] == "-" || name[len(name)-1:] == "_" {
		return firstLetterErr
	}

	return nil
}

const (
	delimiterToReplace   = "."
	delimiterReplacement = "#"
)

func replaceDelimiters(name string) string {
	return strings.Replace(name, delimiterToReplace, delimiterReplacement, -1)
}

func revertDelimiters(name string) string {
	return strings.Replace(name, delimiterReplacement, delimiterToReplace, -1)
}
