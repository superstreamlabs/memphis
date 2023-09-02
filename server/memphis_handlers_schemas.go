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
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/memphisdev/memphis/analytics"
	"github.com/memphisdev/memphis/db"
	"github.com/memphisdev/memphis/memphis_cache"
	"github.com/memphisdev/memphis/models"
	"github.com/memphisdev/memphis/utils"

	"github.com/gin-gonic/gin"
	"github.com/graph-gophers/graphql-go"
	"github.com/hamba/avro/v2"
	"github.com/jhump/protoreflect/desc/protoparse"
	"github.com/santhosh-tekuri/jsonschema/v5"
)

type SchemasHandler struct{ S *Server }

const (
	schemaObjectName                    = "Schema"
	SCHEMA_VALIDATION_ERROR_STATUS_CODE = 555
	schemaUpdatesSubjectTemplate        = "$memphis_schema_updates_%s"
)

var (
	ErrNoSchema = errors.New("no schemas found")
)

func validateProtobufContent(schemaContent string) error {
	parser := protoparse.Parser{
		Accessor: func(filename string) (io.ReadCloser, error) {
			return io.NopCloser(strings.NewReader(schemaContent)), nil
		},
	}
	_, err := parser.ParseFiles("")
	if err != nil {
		return fmt.Errorf("your Proto file is invalid: %v", err.Error())
	}

	return nil
}

func validateJsonSchemaContent(schemaContent string) error {
	_, err := jsonschema.CompileString("test", schemaContent)
	if err != nil {
		return errors.New("your json schema is invalid")
	}

	return nil
}

func validateGraphqlSchemaContent(schemaContent string) error {
	_, err := graphql.ParseSchema(schemaContent, nil)
	if err != nil {
		return err
	}
	return nil
}

func validateAvroSchemaContent(schemaContent string) error {
	_, err := avro.Parse(schemaContent)
	if err != nil {
		return fmt.Errorf("your Avro file is invalid: %v", err.Error())
	}
	return nil
}

func generateProtobufDescriptor(schemaName string, schemaVersionNum int, schemaContent string) ([]byte, error) {
	filename := fmt.Sprintf("%v_%v.proto", schemaName, schemaVersionNum)
	descFilename := fmt.Sprintf("%v_%v_desc", schemaName, schemaVersionNum)
	err := os.WriteFile(filename, []byte(schemaContent), 0644)
	if err != nil {
		return nil, err
	}

	protoCmd := "protoc"
	args := []string{"--descriptor_set_out=" + descFilename, filename}
	cmd := exec.Command(protoCmd, args...)
	err = cmd.Run()
	if err != nil {
		return nil, err
	}

	descContent, err := os.ReadFile(descFilename)
	if err != nil {
		return nil, err
	}

	for _, tmpFile := range []string{filename, descFilename} {
		err = os.Remove(tmpFile)
		if err != nil {
			return nil, err
		}
	}

	return descContent, nil
}

func validateSchemaName(schemaName string) error {
	return validateName(schemaName, schemaObjectName)
}

func validateSchemaType(schemaType string) error {
	invalidTypeErrStr := "unsupported schema type"
	invalidTypeErr := errors.New(invalidTypeErrStr)

	if schemaType == "protobuf" || schemaType == "json" || schemaType == "graphql" || schemaType == "avro" {
		return nil
	} else {
		return invalidTypeErr
	}
}

func validateSchemaContent(schemaContent, schemaType string) error {
	if len(schemaContent) == 0 {
		return errors.New("your schema content is invalid")
	}

	switch schemaType {
	case "protobuf":
		err := validateProtobufContent(schemaContent)
		if err != nil {
			return err
		}
	case "json":
		err := validateJsonSchemaContent(schemaContent)
		if err != nil {
			return err
		}
	case "graphql":
		err := validateGraphqlSchemaContent(schemaContent)
		if err != nil {
			return err
		}
	case "avro":
		err := validateAvroSchemaContent(schemaContent)
		if err != nil {
			return err
		}
	}
	return nil
}

func generateSchemaDescriptor(schemaName string, schemaVersionNum int, schemaContent, schemaType string) (string, error) {
	if len(schemaContent) == 0 {
		return "", errors.New("attempt to generate schema descriptor with empty schema")
	}

	if schemaType != "protobuf" {
		return "", errors.New("descriptor generation with schema type: " + schemaType + ", while protobuf is expected")
	}

	descriptor, err := generateProtobufDescriptor(schemaName, schemaVersionNum, schemaContent)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(descriptor), nil
}

func validateMessageStructName(messageStructName string) error {
	if messageStructName == "" {
		return errors.New("message struct name is required when schema type is Protobuf")
	}
	return nil
}

func generateSchemaUpdateInit(schema models.Schema) (*models.SchemaUpdateInit, error) {
	activeVersion, err := getActiveVersionBySchemaId(schema.ID)
	if err != nil {
		return nil, err
	}

	return &models.SchemaUpdateInit{
		SchemaName: schema.Name,
		ActiveVersion: models.SchemaUpdateVersion{
			VersionNumber:     activeVersion.VersionNumber,
			Descriptor:        activeVersion.Descriptor,
			Content:           activeVersion.SchemaContent,
			MessageStructName: activeVersion.MessageStructName,
		},
		SchemaType: schema.Type,
	}, nil
}

func getSchemaUpdateInitFromStation(sn StationName, tenantName string) (*models.SchemaUpdateInit, error) {
	schema, err := getSchemaByStationName(sn, tenantName)
	if err != nil {
		return nil, err
	}

	return generateSchemaUpdateInit(schema)
}

func (s *Server) updateStationProducersOfSchemaChange(tenantName string, sn StationName, schemaUpdate models.SchemaUpdate) {
	subject := fmt.Sprintf(schemaUpdatesSubjectTemplate, sn.Intern())
	msg, err := json.Marshal(schemaUpdate)
	if err != nil {
		s.Errorf("[tenant: %v]updateStationProducersOfSchemaChange: marshal failed at station %v", tenantName, sn.external)
		return
	}

	account, err := s.lookupAccount(tenantName)
	if err != nil {
		s.Errorf("[tenant: %v]updateStationProducersOfSchemaChange at lookupAccount: %v", tenantName, err.Error())
		return
	}
	s.sendInternalAccountMsg(account, subject, msg)
}

func getSchemaVersionsBySchemaId(id int) ([]models.SchemaVersion, error) {
	schemaVersions, err := db.GetSchemaVersionsBySchemaID(id)
	if err != nil {
		return []models.SchemaVersion{}, err
	}
	return schemaVersions, nil
}

func getActiveVersionBySchemaId(id int) (models.SchemaVersion, error) {
	schemaVersion, err := db.GetActiveVersionBySchemaID(id)
	if err != nil {
		return models.SchemaVersion{}, err
	}
	return schemaVersion, nil
}

func getSchemaByStationName(sn StationName, tenantName string) (models.Schema, error) {
	exist, station, err := db.GetStationByName(sn.Ext(), tenantName)
	if err != nil {
		serv.Errorf("[tenant: %v]getSchemaByStation: At station %v: %v", tenantName, sn.external, err.Error())
		return models.Schema{}, err
	}
	if !exist {
		errMsg := fmt.Sprintf("[tenant: %v]Station %v does not exist", tenantName, station.Name)
		serv.Warnf("getSchemaByStation: " + errMsg)
		return models.Schema{}, errors.New(errMsg)
	}
	if station.SchemaName == "" {
		return models.Schema{}, ErrNoSchema
	}

	exist, schema, err := db.GetSchemaByName(station.SchemaName, station.TenantName)
	if err != nil {
		serv.Errorf("[tenant: %v]getSchemaByStation at GetSchemaByName: Schema %v at station %v: %v", tenantName, station.SchemaName, station.Name, err.Error())
		return models.Schema{}, err
	}
	if !exist {
		serv.Warnf("[tenant: %v]getSchemaByStation: Schema %v does not exist", tenantName, station.SchemaName)
		return models.Schema{}, ErrNoSchema
	}

	return schema, nil
}

func (sh SchemasHandler) GetSchemaByStationName(stationName StationName, tenantName string) (models.Schema, error) {
	return getSchemaByStationName(stationName, tenantName)
}

func (sh SchemasHandler) getSchemaVersionsBySchemaId(schemaId int) ([]models.SchemaVersion, error) {
	return getSchemaVersionsBySchemaId(schemaId)
}

func (sh SchemasHandler) getExtendedSchemaDetailsUpdateAvailable(schemaVersion int, schema models.Schema, tenantName string) (models.ExtendedSchemaDetails, error) {
	var schemaVersions []models.SchemaVersion
	exist, usedSchemaVersion, err := db.GetSchemaVersionByNumberAndID(schemaVersion, schema.ID)
	if err != nil {
		return models.ExtendedSchemaDetails{}, err
	}
	if !exist {
		return models.ExtendedSchemaDetails{}, fmt.Errorf("schema version %v does not exist for schema %v", strconv.Itoa(schemaVersion), schema.Name)
	}

	if !usedSchemaVersion.Active {
		activeSchemaVersion, err := getActiveVersionBySchemaId(schema.ID)
		if err != nil {
			return models.ExtendedSchemaDetails{}, err
		}
		schemaVersions = append(schemaVersions, usedSchemaVersion, activeSchemaVersion)

	} else {
		schemaVersions = append(schemaVersions, usedSchemaVersion)
	}

	var extedndedSchemaDetails models.ExtendedSchemaDetails
	stations, err := db.GetStationNamesUsingSchema(schema.Name, tenantName)
	if err != nil {
		return models.ExtendedSchemaDetails{}, err
	}

	tagsHandler := TagsHandler{S: sh.S}
	tags, err := tagsHandler.GetTagsByEntityWithID("schema", schema.ID)
	if err != nil {
		return models.ExtendedSchemaDetails{}, err
	}

	extedndedSchemaDetails = models.ExtendedSchemaDetails{
		ID:                schema.ID,
		SchemaName:        schema.Name,
		Type:              schema.Type,
		Versions:          schemaVersions,
		UsedStations:      stations,
		Tags:              tags,
		CreatedByUsername: schema.CreatedByUsername,
	}

	return extedndedSchemaDetails, nil
}

func (sh SchemasHandler) getExtendedSchemaDetails(schema models.Schema, tenantName string) (models.ExtendedSchemaDetails, error) {
	schemaVersions, err := sh.getSchemaVersionsBySchemaId(schema.ID)
	if err != nil {
		return models.ExtendedSchemaDetails{}, err
	}

	var extedndedSchemaDetails models.ExtendedSchemaDetails
	stations, err := db.GetStationNamesUsingSchema(schema.Name, tenantName)
	if err != nil {
		return models.ExtendedSchemaDetails{}, err
	}

	tagsHandler := TagsHandler{S: sh.S}
	tags, err := tagsHandler.GetTagsByEntityWithID("schema", schema.ID)
	if err != nil {
		return models.ExtendedSchemaDetails{}, err
	}

	extedndedSchemaDetails = models.ExtendedSchemaDetails{
		ID:                schema.ID,
		SchemaName:        schema.Name,
		Type:              schema.Type,
		Versions:          schemaVersions,
		UsedStations:      stations,
		Tags:              tags,
		CreatedByUsername: schema.CreatedByUsername,
	}

	return extedndedSchemaDetails, nil
}

func (sh SchemasHandler) GetAllSchemasDetails(tenantName string) ([]models.ExtendedSchema, error) {
	schemas, err := db.GetAllSchemasDetails(tenantName)
	if err != nil {
		return []models.ExtendedSchema{}, err
	}
	if len(schemas) == 0 {
		return []models.ExtendedSchema{}, nil
	}

	for i, schema := range schemas {
		stations, err := db.GetCountStationsUsingSchema(schema.Name, tenantName)
		if err != nil {
			return []models.ExtendedSchema{}, err
		}
		if stations > 0 {
			schema.Used = true
		} else {
			schema.Used = false
		}
		tagsHandler := TagsHandler{S: sh.S}
		tags, err := tagsHandler.GetTagsByEntityWithID("schema", schemas[i].ID)
		if err != nil {
			return []models.ExtendedSchema{}, err
		}
		schema.Tags = tags
		schemas[i] = schema
	}
	if err != nil {
		return []models.ExtendedSchema{}, err
	}
	return schemas, nil
}

func (sh SchemasHandler) CreateNewSchema(c *gin.Context) {
	var body models.CreateNewSchema
	ok := utils.Validate(c, &body, false, nil)
	if !ok {
		return
	}
	schemaName := strings.ToLower(body.Name)
	err := validateSchemaName(schemaName)
	if err != nil {
		serv.Warnf("CreateNewSchema at validateSchemaName: %v", err.Error())
		c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": err.Error()})
		return
	}
	user, err := getUserDetailsFromMiddleware(c)
	if err != nil {
		serv.Errorf("CreateNewSchema at getUserDetailsFromMiddleware: Schema %v: %v", schemaName, err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	tenantName := user.TenantName
	exist, _, err := db.GetSchemaByName(schemaName, tenantName)
	if err != nil {
		serv.Errorf("[tenant: %v][user: %v]CreateNewSchema at GetSchemaByName: Schema %v: %v", user.TenantName, user.Username, schemaName, err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server Error"})
		return
	}
	if exist {
		errMsg := fmt.Sprintf("Schema with the name %v already exists", schemaName)
		serv.Warnf("[tenant: %v][user: %v]CreateNewSchema: %v", user.TenantName, user.Username, errMsg)
		c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": errMsg})
		return
	}
	schemaType := strings.ToLower(body.Type)
	err = validateSchemaType(schemaType)
	if err != nil {
		serv.Warnf("[tenant: %v][user: %v]CreateNewSchema at validateSchemaType: Schema %v: %v", user.TenantName, user.Username, schemaName, err.Error())
		c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": err.Error()})
		return
	}
	messageStructName := body.MessageStructName
	if schemaType == "protobuf" {
		err := validateMessageStructName(messageStructName)
		if err != nil {
			serv.Warnf("[tenant: %v][user: %v]CreateNewSchema at validateMessageStructName: Schema %v: %v", user.TenantName, user.Username, schemaName, err.Error())
			c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": err.Error()})
			return
		}
	}

	schemaContent := body.SchemaContent
	err = validateSchemaContent(schemaContent, schemaType)
	if err != nil {
		serv.Warnf("[tenant: %v][user: %v]CreateNewSchema at validateSchemaContent: Schema %v: %v", user.TenantName, user.Username, schemaName, err.Error())
		c.AbortWithStatusJSON(SCHEMA_VALIDATION_ERROR_STATUS_CODE, gin.H{"message": err.Error()})
		return
	}
	schemaVersionNumber := 1
	descriptor := ""
	if schemaType == "protobuf" {
		descriptor, err = generateSchemaDescriptor(schemaName, schemaVersionNumber, schemaContent, schemaType)
		if err != nil {
			serv.Warnf("[tenant: %v][user: %v]CreateNewSchema at generateSchemaDescriptor: Schema %v: %v", user.TenantName, user.Username, schemaName, err.Error())
			c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": err.Error()})
			return
		}
	}

	newSchema, rowsUpdated, err := db.InsertNewSchema(schemaName, schemaType, user.Username, tenantName)
	if err != nil {
		serv.Errorf("[tenant: %v][user: %v]CreateNewSchema at InsertNewSchema: Schema %v: %v", user.TenantName, user.Username, schemaName, err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	if rowsUpdated == 1 {
		_, _, err = db.InsertNewSchemaVersion(schemaVersionNumber, user.ID, user.Username, schemaContent, newSchema.ID, messageStructName, descriptor, true, tenantName)
		if err != nil {
			serv.Errorf("[tenant: %v][user: %v]CreateNewSchema at InsertNewSchemaVersion: %v", user.TenantName, user.Username, err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}
		message := fmt.Sprintf("[tenant: %v][user: %v]Schema %v has been created by %v", user.TenantName, user.Username, schemaName, user.Username)
		serv.Noticef(message)
	} else {
		errMsg := fmt.Sprintf("Schema with the name %v already exists", schemaName)
		serv.Warnf("[tenant: %v][user: %v]CreateNewSchema: %v", user.TenantName, user.Username, errMsg)
		c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": errMsg})
		return
	}

	if len(body.Tags) > 0 {
		err = AddTagsToEntity(body.Tags, "schema", newSchema.ID, tenantName, "")
		if err != nil {
			serv.Errorf("[tenant: %v][user: %v]CreateNewSchema at AddTagsToEntity: Failed creating tag at schema %v: %v", user.TenantName, user.Username, schemaName, err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}
	} else {
		err = CreateDefaultTags("schema", newSchema.ID, tenantName)
		if err != nil {
			serv.Errorf("[tenant: %v][user: %v]createNewSchema at CreateDefaultTags: %v", tenantName, user.Username, err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}
	}

	shouldSendAnalytics, _ := shouldSendAnalytics()
	if shouldSendAnalytics {
		analyticsParams := map[string]interface{}{"schema-name": newSchema.Name}
		analytics.SendEvent(user.TenantName, user.Username, analyticsParams, "user-create-schema")
	}

	c.IndentedJSON(200, newSchema)
}

func (sh SchemasHandler) GetAllSchemas(c *gin.Context) {
	user, err := getUserDetailsFromMiddleware(c)
	if err != nil {
		serv.Errorf("GetAllSchemas: %v", err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	schemas, err := sh.GetAllSchemasDetails(user.TenantName)
	if err != nil {
		serv.Errorf("[tenant: %v][user: %v]GetAllSchemas at db.GetAllSchemasDetails: %v", user.TenantName, user.Username, err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	shouldSendAnalytics, _ := shouldSendAnalytics()
	if shouldSendAnalytics {
		analyticsParams := make(map[string]interface{})
		analytics.SendEvent(user.TenantName, user.Username, analyticsParams, "user-enter-schemas-page")
	}

	c.IndentedJSON(200, schemas)
}

func (sh SchemasHandler) GetSchemaDetails(c *gin.Context) {
	var body models.GetSchemaDetails
	ok := utils.Validate(c, &body, false, nil)
	if !ok {
		return
	}
	schemaName := strings.ToLower(body.SchemaName)
	user, err := getUserDetailsFromMiddleware(c)
	if err != nil {
		serv.Errorf("GetSchemaDetails at getUserDetailsFromMiddleware: Schema %v: %v", schemaName, err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	exist, schema, err := db.GetSchemaByName(schemaName, user.TenantName)
	if err != nil {
		serv.Errorf("[tenant: %v][user: %v]GetSchemaDetails at GetSchemaByName: Schema %v: %v", user.TenantName, user.Username, body.SchemaName, err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	if !exist {
		errMsg := fmt.Sprintf("Schema %v does not exist", body.SchemaName)
		serv.Warnf("[tenant: %v][user: %v]GetSchemaDetails: %v", user.TenantName, user.Username, errMsg)
		c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": errMsg})
		return
	}

	schemaDetails, err := sh.getExtendedSchemaDetails(schema, user.TenantName)
	if err != nil {
		serv.Errorf("[tenant: %v][user: %v]GetSchemaDetails at getExtendedSchemaDetails: Schema %v: %v", user.TenantName, user.Username, schemaName, err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	shouldSendAnalytics, _ := shouldSendAnalytics()
	if shouldSendAnalytics {
		user, _ := getUserDetailsFromMiddleware(c)
		analyticsParams := map[string]interface{}{"schema-name": schemaName}
		analytics.SendEvent(user.TenantName, user.Username, analyticsParams, "user-enter-schema-details")
	}

	c.IndentedJSON(200, schemaDetails)
}

func deleteSchemaFromStations(s *Server, schemaName string, tenantName string) error {
	stationNames, err := db.GetStationNamesUsingSchema(schemaName, tenantName)
	if err != nil {
		return err
	}
	for _, name := range stationNames {
		sn, err := StationNameFromStr(name)
		if err != nil {
			return err
		}
		removeSchemaFromStation(s, sn, false, tenantName)
	}

	err = db.RemoveSchemaFromAllUsingStations(schemaName, tenantName)
	if err != nil {
		s.Errorf("[tenant: %v]deleteSchemaFromStations at RemoveSchemaFromAllUsingStations: Schema %v: %v", tenantName, schemaName, err.Error())
		return err
	}

	return nil
}

func (sh SchemasHandler) RemoveSchema(c *gin.Context) {
	var body models.RemoveSchema
	ok := utils.Validate(c, &body, false, nil)
	if !ok {
		return
	}
	var schemaIds []int
	user, err := getUserDetailsFromMiddleware(c)
	if err != nil {
		serv.Errorf("RemoveSchema: %v", err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	tenantName := user.TenantName
	for _, name := range body.SchemaNames {
		schemaName := strings.ToLower(name)
		exist, schema, err := db.GetSchemaByName(schemaName, tenantName)
		if err != nil {
			serv.Errorf("[tenant: %v][user: %v]RemoveSchema at GetSchemaByName: Schema %v: %v", user.TenantName, user.Username, schemaName, err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}
		if exist {
			DeleteTagsFromSchema(schema.ID)
			err := deleteSchemaFromStations(sh.S, schema.Name, tenantName)
			if err != nil {
				serv.Errorf("[tenant: %v][user: %v]RemoveSchema at deleteSchemaFromStations: Schema %v: %v", user.TenantName, user.Username, schemaName, err.Error())
				c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
				return
			}

			schemaIds = append(schemaIds, schema.ID)
		}
	}

	if len(schemaIds) > 0 {
		err := db.FindAndDeleteSchema(schemaIds)
		if err != nil {
			serv.Errorf("[tenant: %v][user: %v]RemoveSchema at FindAndDeleteSchema: Schema %v", user.TenantName, user.Username, err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}
		for _, name := range body.SchemaNames {
			serv.Noticef("[tenant: %v][user: %v]Schema %v has been deleted", user.TenantName, user.Username, name)
		}
	}

	shouldSendAnalytics, _ := shouldSendAnalytics()
	if shouldSendAnalytics {
		analyticsParams := make(map[string]interface{})
		analytics.SendEvent(user.TenantName, user.Username, analyticsParams, "user-remove-schema")
	}

	c.IndentedJSON(200, gin.H{})
}

func (sh SchemasHandler) CreateNewVersion(c *gin.Context) {
	var body models.CreateNewVersion
	ok := utils.Validate(c, &body, false, nil)
	if !ok {
		return
	}

	user, err := getUserDetailsFromMiddleware(c)
	if err != nil {
		serv.Errorf("CreateNewVersion at getUserDetailsFromMiddleware: Schema %v: %v", body.SchemaName, err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	schemaName := strings.ToLower(body.SchemaName)
	exist, schema, err := db.GetSchemaByName(schemaName, user.TenantName)
	if err != nil {
		serv.Errorf("[tenant: %v][user: %v]CreateNewVersion at GetSchemaByName: Schema %v: %v", user.TenantName, user.Username, body.SchemaName, err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server Error"})
		return
	}
	if !exist {
		errMsg := fmt.Sprintf("Schema %v does not exist", body.SchemaName)
		serv.Warnf("[tenant: %v][user: %v]CreateNewVersion: %v", user.TenantName, user.Username, errMsg)
		c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": errMsg})
		return
	}

	messageStructName := body.MessageStructName
	if schema.Type == "protobuf" {
		err := validateMessageStructName(messageStructName)
		if err != nil {
			serv.Errorf("[tenant: %v][user: %v]CreateNewVersion at validateMessageStructName: Schema %v: %v", user.TenantName, user.Username, body.SchemaName, err.Error())
			c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": err.Error()})
			return
		}
	}
	schemaContent := body.SchemaContent
	err = validateSchemaContent(schemaContent, schema.Type)
	if err != nil {
		serv.Errorf("[tenant: %v][user: %v]CreateNewVersion at validateSchemaContent: Schema %v: %v", user.TenantName, user.Username, body.SchemaName, err.Error())
		c.AbortWithStatusJSON(SCHEMA_VALIDATION_ERROR_STATUS_CODE, gin.H{"message": err.Error()})
		return
	}

	countVersions, err := db.GetShcemaVersionsCount(schema.ID, user.TenantName)
	if err != nil {
		serv.Errorf("[tenant: %v][user: %v]CreateNewVersion at GetShcemaVersionsCount: Schema %v: %v", user.TenantName, user.Username, body.SchemaName, err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	versionNumber := countVersions + 1
	descriptor := ""
	if schema.Type == "protobuf" {
		descriptor, err = generateSchemaDescriptor(schemaName, versionNumber, schemaContent, schema.Type)
		if err != nil {
			serv.Warnf("[tenant: %v][user: %v]CreateNewVersion at generateSchemaDescriptor: Schema %v: %v", user.TenantName, user.Username, body.SchemaName, err.Error())
			c.AbortWithStatusJSON(SCHEMA_VALIDATION_ERROR_STATUS_CODE, gin.H{"message": err.Error()})
			return
		}
	}
	newSchemaVersion, rowsUpdated, err := db.InsertNewSchemaVersion(versionNumber, user.ID, user.Username, schemaContent, schema.ID, messageStructName, descriptor, false, user.TenantName)
	if err != nil {
		serv.Warnf("[tenant: %v][user: %v]CreateNewVersion at InsertNewSchemaVersion: %v", user.TenantName, user.Username, err.Error())
		c.AbortWithStatusJSON(SCHEMA_VALIDATION_ERROR_STATUS_CODE, gin.H{"message": err.Error()})
		return
	}
	if rowsUpdated == 1 {
		serv.Noticef("[tenant: %v][user: %v]Schema Version %v has been created by %v", user.TenantName, user.Username, strconv.Itoa(newSchemaVersion.VersionNumber), user.Username)
	} else {
		serv.Warnf("[tenant: %v][user: %v]CreateNewVersion: Schema %v: Version %v already exists", user.TenantName, user.Username, body.SchemaName, strconv.Itoa(newSchemaVersion.VersionNumber))
		c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": "Version already exists"})
		return
	}
	extedndedSchemaDetails, err := sh.getExtendedSchemaDetails(schema, user.TenantName)
	if err != nil {
		serv.Errorf("[tenant: %v][user: %v]CreateNewVersion at getExtendedSchemaDetails: Schema %v: %v", user.TenantName, user.Username, body.SchemaName, err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	shouldSendAnalytics, _ := shouldSendAnalytics()
	if shouldSendAnalytics {
		analyticsParams := make(map[string]interface{})
		analytics.SendEvent(user.TenantName, user.Username, analyticsParams, "user-create-new-schema-version")
	}

	c.IndentedJSON(200, extedndedSchemaDetails)

}

func (sh SchemasHandler) RollBackVersion(c *gin.Context) {
	var body models.RollBackVersion
	ok := utils.Validate(c, &body, false, nil)
	if !ok {
		return
	}

	var extedndedSchemaDetails models.ExtendedSchemaDetails

	user, err := getUserDetailsFromMiddleware(c)
	if err != nil {
		serv.Errorf("RollBackVersion at getUserDetailsFromMiddleware: Schema %v: %v", body.SchemaName, err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server Error"})
		return
	}
	schemaName := strings.ToLower(body.SchemaName)
	exist, schema, err := db.GetSchemaByName(schemaName, user.TenantName)
	if err != nil {
		serv.Errorf("[tenant: %v][user: %v]RollBackVersion at GetSchemaByName: Schema %v: %v", user.TenantName, user.Username, body.SchemaName, err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server Error"})
		return
	}
	if !exist {
		errMsg := fmt.Sprintf("Schema %v does not exist", body.SchemaName)
		serv.Warnf("[tenant: %v][user: %v]RollBackVersion: %v", user.TenantName, user.Username, errMsg)
		c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": errMsg})
		return
	}

	schemaVersion := body.VersionNumber
	exist, _, err = db.GetSchemaVersionByNumberAndID(schemaVersion, schema.ID)
	if err != nil {
		serv.Errorf("[tenant: %v][user: %v]RollBackVersion at GetSchemaVersionByNumberAndID: Schema %v version %v: %v", user.TenantName, user.Username, body.SchemaName, strconv.Itoa(schemaVersion), err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	if !exist {
		errMsg := fmt.Sprintf("Schema %v version %v does not exist", body.SchemaName, strconv.Itoa(schemaVersion))
		serv.Warnf("[tenant: %v][user: %v]RollBackVersion: %v", user.TenantName, user.Username, errMsg)
		c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": errMsg})
		return
	}

	countVersions, err := db.GetShcemaVersionsCount(schema.ID, user.TenantName)
	if err != nil {
		serv.Errorf("[tenant: %v][user: %v]RollBackVersion at GetShcemaVersionsCount: Schema %v: %v", user.TenantName, user.Username, body.SchemaName, err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	if countVersions > 1 {
		err = db.UpdateSchemaActiveVersion(schema.ID, body.VersionNumber)
		if err != nil {
			serv.Errorf("[tenant: %v][user: %v]RollBackVersion at UpdateSchemaActiveVersion: Schema %v: %v", user.TenantName, user.Username, body.SchemaName, err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": err.Error()})
			return
		}
	}
	extedndedSchemaDetails, err = sh.getExtendedSchemaDetails(schema, user.TenantName)
	if err != nil {
		serv.Errorf("[tenant: %v][user: %v]RollBackVersion at getExtendedSchemaDetails: Schema %v: %v", user.TenantName, user.Username, body.SchemaName, err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	shouldSendAnalytics, _ := shouldSendAnalytics()
	if shouldSendAnalytics {
		analyticsParams := make(map[string]interface{})
		analytics.SendEvent(user.TenantName, user.Username, analyticsParams, "user-rollback-schema-version")
	}

	c.IndentedJSON(200, extedndedSchemaDetails)
}

func (sh SchemasHandler) ValidateSchema(c *gin.Context) {
	var body models.ValidateSchema
	ok := utils.Validate(c, &body, false, nil)
	if !ok {
		return
	}

	schemaType := strings.ToLower(body.SchemaType)
	err := validateSchemaType(schemaType)
	if err != nil {
		serv.Warnf("ValidateSchema at validateSchemaType: Schema type %v: %v", schemaType, err.Error())
		c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": err.Error()})
		return
	}

	schemaContent := body.SchemaContent
	err = validateSchemaContent(schemaContent, schemaType)
	if err != nil {
		serv.Warnf("ValidateSchema at validateSchemaContent: Schema type %v: %v", schemaType, err.Error())
		c.AbortWithStatusJSON(SCHEMA_VALIDATION_ERROR_STATUS_CODE, gin.H{"message": err.Error()})
		return
	}

	shouldSendAnalytics, _ := shouldSendAnalytics()
	if shouldSendAnalytics {
		user, _ := getUserDetailsFromMiddleware(c)
		analyticsParams := make(map[string]interface{})
		analytics.SendEvent(user.TenantName, user.Username, analyticsParams, "user-validate-schema")
	}

	c.IndentedJSON(200, gin.H{
		"is_valid": true,
	})
}

func (s *Server) createSchemaDirect(c *client, reply string, msg []byte) {
	var csr CreateSchemaReq
	var resp SchemaResponse
	var tenantName string
	tenantName, message, err := s.getTenantNameAndMessage(msg)
	if err != nil {
		s.Errorf("createSchemaDirect at getTenantNameAndMessage- failed creating Schema: %v", err.Error())
		respondWithRespErr(s.MemphisGlobalAccountString(), s, reply, err, &resp)
		return
	}
	if err := json.Unmarshal([]byte(message), &csr); err != nil {
		s.Errorf("[tenant: %v]createSchemaDirect at json.Unmarshal - failed creating Schema: %v", tenantName, err.Error())
		respondWithRespErr(s.MemphisGlobalAccountString(), s, reply, err, &resp)
		return
	}

	err = validateSchemaContent(csr.SchemaContent, csr.Type)
	if err != nil {
		s.Warnf("[tenant: %v]createSchemaDirect at validateSchemaContent- Schema is not in the right %v format, error: %v", tenantName, csr.Type, err.Error())
		respondWithRespErr(s.MemphisGlobalAccountString(), s, reply, err, &resp)
		return
	}

	if csr.Type == "protobuf" {
		csr.MessageStructName, err = getProtoMessageStructName(csr.SchemaContent)
		if err != nil {
			s.Errorf("[tenant: %v]createSchemaDirect at getProtoMessageStructName- failed creating Schema: %v : %v", tenantName, csr.Name, err.Error())
			respondWithRespErr(s.MemphisGlobalAccountString(), s, reply, err, &resp)
		}
	}

	if csr.Type == "protobuf" {
		err := validateMessageStructName(csr.MessageStructName)
		if err != nil {
			s.Warnf("[tenant: %v]createSchemaDirect at validateMessageStructName- failed creating Schema: %v : %v", tenantName, csr.Name, err.Error())
			respondWithRespErr(s.MemphisGlobalAccountString(), s, reply, err, &resp)
			return
		}
	}

	exist, existedSchema, err := db.GetSchemaByName(csr.Name, tenantName)
	if err != nil {
		s.Errorf("[tenant: %v]createSchemaDirect at GetSchemaByName- failed creating Schema: %v : %v", tenantName, csr.Name, err.Error())
		respondWithRespErr(s.MemphisGlobalAccountString(), s, reply, err, &resp)
		return
	}

	if exist {
		if existedSchema.Type == csr.Type {
			err = s.updateSchemaVersion(existedSchema.ID, tenantName, csr)
			if err != nil {
				s.Errorf("[tenant: %v]createSchemaDirect at updateSchemaVersion - failed creating Schema: %v : %v", tenantName, csr.Name, err.Error())
				respondWithRespErr(s.MemphisGlobalAccountString(), s, reply, err, &resp)
				return
			}
			respondWithRespErr(s.MemphisGlobalAccountString(), s, reply, err, &resp)
			return
		} else {
			s.Warnf("[tenant: %v]createSchemaDirect: %v Bad Schema Type", tenantName, csr.Name)
			badTypeError := fmt.Sprintf("%v already exist with type - %v", csr.Name, existedSchema.Type)
			respondWithRespErr(s.MemphisGlobalAccountString(), s, reply, errors.New(badTypeError), &resp)
			return
		}
	}

	err = s.createNewSchema(csr, tenantName)
	if err != nil {
		s.Errorf("[tenant: %v]createSchemaDirect - failed creating Schema: %v : %v", tenantName, csr.Name, err.Error())
		respondWithRespErr(s.MemphisGlobalAccountString(), s, reply, err, &resp)
		return
	}

	respondWithRespErr(s.MemphisGlobalAccountString(), s, reply, err, &resp)

}

func (s *Server) updateSchemaVersion(schemaID int, tenantName string, newSchemaReq CreateSchemaReq) error {
	_, user, err := memphis_cache.GetUser(newSchemaReq.CreatedByUsername, tenantName, false)
	if err != nil {
		s.Errorf("[tenant: %v]updateSchemaVersion at memphis_cache.GetUser: Schema %v: %v", tenantName, newSchemaReq.Name, err.Error())
		return err
	}

	countVersions, err := db.GetShcemaVersionsCount(schemaID, user.TenantName)
	if err != nil {
		s.Errorf("[tenant: %v][user: %v]updateSchemaVersion at db.GetShcemaVersionsCount: Schema %v: %v", tenantName, user.Username, newSchemaReq.Name, err.Error())
		return err
	}

	_, currentSchema, err := db.GetSchemaVersionByNumberAndID(countVersions, schemaID)
	if err != nil {
		s.Errorf("[tenant: %v][user: %v]updateSchemaVersion at db.GetSchemaVersionByNumberAndID: Schema %v: %v", tenantName, user.Username, newSchemaReq.Name, err.Error())
		return err
	}

	if currentSchema.SchemaContent == newSchemaReq.SchemaContent {
		alreadyExistInDB := fmt.Sprintf("%v already exist in the db", newSchemaReq.Name)
		s.Errorf(alreadyExistInDB)
		return errors.New(alreadyExistInDB)
	}

	versionNumber := countVersions + 1

	descriptor := ""
	if newSchemaReq.Type == "protobuf" {
		descriptor, err = generateSchemaDescriptor(newSchemaReq.Name, 1, newSchemaReq.SchemaContent, newSchemaReq.Type)
		if err != nil {
			s.Errorf("[tenant: %v][user: %v]CreateNewSchemaDirectn: could not create proto descriptor for %v: %v", tenantName, user.Username, newSchemaReq.Name, err.Error())
			return err
		}
	}

	newSchemaVersion, rowsUpdated, err := db.InsertNewSchemaVersion(versionNumber, user.ID, user.Username, newSchemaReq.SchemaContent, schemaID, newSchemaReq.MessageStructName, descriptor, false, tenantName)
	if err != nil {
		s.Errorf("[tenant: %v][user: %v]updateSchemaVersion: %v", tenantName, user.Username, err.Error())
		return err
	}
	if rowsUpdated == 1 {
		message := fmt.Sprintf("[tenant: %v][user: %v]Schema Version %v has been created by %v", tenantName, user.Username, strconv.Itoa(newSchemaVersion.VersionNumber), user.Username)
		s.Noticef(message)
		return nil
	} else {
		s.Errorf("[tenant: %v][user: %v]updateSchemaVersion: schema update failed", tenantName, user.Username)
		return errors.New("updateSchemaVersion: schema update failed")
	}

}

func (s *Server) createNewSchema(newSchemaReq CreateSchemaReq, tenantName string) error {
	schemaVersionNumber := 1

	_, user, err := memphis_cache.GetUser(newSchemaReq.CreatedByUsername, tenantName, false)
	if err != nil {
		s.Errorf("[tenant: %v]createNewSchema at memphis_cache.GetUser: %v", tenantName, err.Error())
		return err
	}

	descriptor := ""
	if newSchemaReq.Type == "protobuf" {
		descriptor, err = generateSchemaDescriptor(newSchemaReq.Name, 1, newSchemaReq.SchemaContent, newSchemaReq.Type)
		if err != nil {
			s.Errorf("[tenant: %v][user: %v]CreateNewSchema at generateSchemaDescriptor: Schema %v: %v", tenantName, user.Username, newSchemaReq.Name, err.Error())
			return err
		}
	}

	newSchema, rowUpdated, err := db.InsertNewSchema(newSchemaReq.Name, newSchemaReq.Type, newSchemaReq.CreatedByUsername, tenantName)
	if err != nil {
		s.Errorf("[tenant: %v][user: %v]createNewSchema at db.InsertNewSchema: %v", tenantName, user.Username, err.Error())
		return err
	}

	if rowUpdated == 1 {
		_, _, err := db.InsertNewSchemaVersion(schemaVersionNumber, user.ID, user.Username, newSchemaReq.SchemaContent, newSchema.ID, newSchemaReq.MessageStructName, descriptor, true, tenantName)
		if err != nil {
			s.Errorf("[tenant: %v][user: %v]createNewSchema at db.InsertNewSchemaVersion: %v", tenantName, user.Username, err.Error())
			return err
		}
	}

	err = CreateDefaultTags("schema", newSchema.ID, tenantName)
	if err != nil {
		s.Errorf("[tenant: %v][user: %v]createNewSchema at CreateDefaultTags: %v", tenantName, user.Username, err.Error())
		return err
	}

	return nil
}

func getProtoMessageStructName(schema_content string) (string, error) {
	parser := protoparse.Parser{
		Accessor: func(filename string) (io.ReadCloser, error) {
			return io.NopCloser(strings.NewReader(string(schema_content))), nil
		},
	}
	something, err := parser.ParseFiles("")
	if err != nil {
		return "", errors.New("your Proto file is invalid: " + err.Error())
	}
	return something[0].GetMessageTypes()[0].GetName(), nil
}
