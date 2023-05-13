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
	"memphis/analytics"
	"memphis/db"
	"memphis/models"
	"memphis/utils"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/graph-gophers/graphql-go"
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
		return errors.New("your Proto file is invalid: " + err.Error())
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
	invalidSupportTypeErrStr := "avro is not supported at this time"
	invalidSupportTypeErr := errors.New(invalidSupportTypeErrStr)

	if schemaType == "protobuf" || schemaType == "json" || schemaType == "graphql" {
		return nil
	} else if schemaType == "avro" {
		return invalidSupportTypeErr
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
		break
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

func generateSchemaUpdateInit(schema models.Schema) (*models.ProducerSchemaUpdateInit, error) {
	activeVersion, err := getActiveVersionBySchemaId(schema.ID)
	if err != nil {
		return nil, err
	}

	return &models.ProducerSchemaUpdateInit{
		SchemaName: schema.Name,
		ActiveVersion: models.ProducerSchemaUpdateVersion{
			VersionNumber:     activeVersion.VersionNumber,
			Descriptor:        activeVersion.Descriptor,
			Content:           activeVersion.SchemaContent,
			MessageStructName: activeVersion.MessageStructName,
		},
		SchemaType: schema.Type,
	}, nil
}

func getSchemaUpdateInitFromStation(sn StationName, tenantName string) (*models.ProducerSchemaUpdateInit, error) {
	schema, err := getSchemaByStationName(sn, tenantName)
	if err != nil {
		return nil, err
	}

	return generateSchemaUpdateInit(schema)
}

func (s *Server) updateStationProducersOfSchemaChange(tenantName string, sn StationName, schemaUpdate models.ProducerSchemaUpdate) {
	subject := fmt.Sprintf(schemaUpdatesSubjectTemplate, sn.Intern())
	msg, err := json.Marshal(schemaUpdate)
	if err != nil {
		s.Errorf("updateStationProducersOfSchemaChange: marshal failed at station " + sn.external)
		return
	}

	account, err := s.lookupAccount(tenantName)
	if err != nil {
		s.Errorf("updateStationProducersOfSchemaChange " + err.Error())
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
		serv.Errorf("getSchemaByStation: At station " + sn.external + ": " + err.Error())
		return models.Schema{}, err
	}
	if !exist {
		errMsg := "Station " + station.Name + " does not exist"
		serv.Warnf("getSchemaByStation: " + errMsg)
		return models.Schema{}, errors.New(errMsg)
	}
	if station.SchemaName == "" {
		return models.Schema{}, ErrNoSchema
	}

	exist, schema, err := db.GetSchemaByName(station.SchemaName, station.TenantName)
	if err != nil {
		serv.Errorf("getSchemaByStation: Schema" + station.SchemaName + "at station " + station.Name + err.Error())
		return models.Schema{}, err
	}
	if !exist {
		serv.Warnf("getSchemaByStation: Schema " + station.SchemaName + " does not exist")
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
		return models.ExtendedSchemaDetails{}, errors.New("Schema version " + strconv.Itoa(schemaVersion) + " does not exist for schema " + schema.Name)
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
		serv.Warnf("CreateNewSchema: " + err.Error())
		c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": err.Error()})
		return
	}
	user, err := getUserDetailsFromMiddleware(c)
	if err != nil {
		serv.Errorf("CreateNewSchema: Schema " + schemaName + ": " + err.Error())
		c.AbortWithStatusJSON(401, gin.H{"message": "Unauthorized"})
		return
	}
	tenantName := user.TenantName
	exist, _, err := db.GetSchemaByName(schemaName, tenantName)
	if err != nil {
		serv.Errorf("CreateNewSchema: Schema " + schemaName + ": " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server Error"})
		return
	}
	if exist {
		errMsg := "Schema with the name " + schemaName + " already exists"
		serv.Warnf("CreateNewSchema: " + errMsg)
		c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": errMsg})
		return
	}
	schemaType := strings.ToLower(body.Type)
	err = validateSchemaType(schemaType)
	if err != nil {
		serv.Warnf("CreateNewSchema: Schema " + schemaName + ": " + err.Error())
		c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": err.Error()})
		return
	}
	messageStructName := body.MessageStructName
	if schemaType == "protobuf" {
		err := validateMessageStructName(messageStructName)
		if err != nil {
			serv.Warnf("CreateNewSchema: Schema " + schemaName + ": " + err.Error())
			c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": err.Error()})
			return
		}
	}

	schemaContent := body.SchemaContent
	err = validateSchemaContent(schemaContent, schemaType)
	if err != nil {
		serv.Warnf("CreateNewSchema: Schema " + schemaName + ": " + err.Error())
		c.AbortWithStatusJSON(SCHEMA_VALIDATION_ERROR_STATUS_CODE, gin.H{"message": err.Error()})
		return
	}
	schemaVersionNumber := 1
	descriptor := ""
	if schemaType == "protobuf" {
		descriptor, err = generateSchemaDescriptor(schemaName, schemaVersionNumber, schemaContent, schemaType)
		if err != nil {
			serv.Warnf("CreateNewSchema: Schema " + schemaName + ": " + err.Error())
			c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": err.Error()})
			return
		}
	}

	newSchema, rowsUpdated, err := db.InsertNewSchema(schemaName, schemaType, user.Username, tenantName)
	if err != nil {
		serv.Errorf("CreateNewSchema: Schema " + schemaName + ": " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	if rowsUpdated == 1 {
		_, _, err = db.InsertNewSchemaVersion(schemaVersionNumber, user.ID, user.Username, schemaContent, newSchema.ID, messageStructName, descriptor, true, tenantName)
		if err != nil {
			serv.Errorf("CreateNewSchema: " + err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}
		message := "Schema " + schemaName + " has been created by " + user.Username
		serv.Noticef(message)
	} else {
		errMsg := "Schema with the name " + schemaName + " already exists"
		serv.Warnf("CreateNewSchema: " + errMsg)
		c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": errMsg})
		return
	}

	if len(body.Tags) > 0 {
		err = AddTagsToEntity(body.Tags, "schema", newSchema.ID, tenantName)
		if err != nil {
			serv.Errorf("CreateNewSchema: Failed creating tag at schema " + schemaName + ": " + err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}
	}

	shouldSendAnalytics, _ := shouldSendAnalytics()
	if shouldSendAnalytics {
		param := analytics.EventParam{
			Name:  "schema-name",
			Value: newSchema.Name,
		}
		analyticsParams := []analytics.EventParam{param}
		analytics.SendEventWithParams(user.Username, analyticsParams, "user-create-schema")
	}

	c.IndentedJSON(200, newSchema)
}

func (sh SchemasHandler) GetAllSchemas(c *gin.Context) {
	user, err := getUserDetailsFromMiddleware(c)
	if err != nil {
		serv.Errorf("GetAllSchemas: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	schemas, err := sh.GetAllSchemasDetails(user.TenantName)
	if err != nil {
		serv.Errorf("GetAllSchemas: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	shouldSendAnalytics, _ := shouldSendAnalytics()
	if shouldSendAnalytics {
		analytics.SendEvent(user.Username, "user-enter-schemas-page")
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
		serv.Errorf("GetSchemaDetails: Schema " + schemaName + ": " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	exist, schema, err := db.GetSchemaByName(schemaName, user.TenantName)
	if err != nil {
		serv.Errorf("GetSchemaDetails: Schema " + body.SchemaName + ": " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	if !exist {
		errMsg := "Schema " + body.SchemaName + " does not exist"
		serv.Warnf("GetSchemaDetails: " + errMsg)
		c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": errMsg})
		return
	}

	schemaDetails, err := sh.getExtendedSchemaDetails(schema, user.TenantName)
	if err != nil {
		serv.Errorf("GetSchemaDetails: Schema " + schemaName + ": " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	shouldSendAnalytics, _ := shouldSendAnalytics()
	if shouldSendAnalytics {
		user, _ := getUserDetailsFromMiddleware(c)
		param := analytics.EventParam{
			Name:  "schema-name",
			Value: schemaName,
		}
		analyticsParams := []analytics.EventParam{param}
		analytics.SendEventWithParams(user.Username, analyticsParams, "user-enter-schema-details")
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
		s.Errorf("deleteSchemaFromStations: Schema " + schemaName + ": " + err.Error())
		return err
	}

	return nil
}

func (sh SchemasHandler) RemoveSchema(c *gin.Context) {
	// if err := DenyForSandboxEnv(c); err != nil {
	// 	return
	// }
	var body models.RemoveSchema
	ok := utils.Validate(c, &body, false, nil)
	if !ok {
		return
	}
	var schemaIds []int
	user, err := getUserDetailsFromMiddleware(c)
	if err != nil {
		serv.Errorf("RemoveSchema: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	tenantName := user.TenantName
	for _, name := range body.SchemaNames {
		schemaName := strings.ToLower(name)
		exist, schema, err := db.GetSchemaByName(schemaName, tenantName)
		if err != nil {
			serv.Errorf("RemoveSchema: Schema " + schemaName + ": " + err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}
		if exist {
			DeleteTagsFromSchema(schema.ID)
			err := deleteSchemaFromStations(sh.S, schema.Name, tenantName)
			if err != nil {
				serv.Errorf("RemoveSchema: Schema " + schemaName + ": " + err.Error())
				c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
				return
			}

			schemaIds = append(schemaIds, schema.ID)
		}
	}

	if len(schemaIds) > 0 {
		err := db.FindAndDeleteSchema(schemaIds)
		if err != nil {
			serv.Errorf("RemoveSchema: " + err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}
		for _, name := range body.SchemaNames {
			serv.Noticef("Schema " + name + " has been deleted")
		}
	}

	shouldSendAnalytics, _ := shouldSendAnalytics()
	if shouldSendAnalytics {
		analytics.SendEvent(user.Username, "user-remove-schema")
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
		serv.Errorf("CreateNewVersion: Schema " + body.SchemaName + ": " + err.Error())
		c.AbortWithStatusJSON(401, gin.H{"message": "Unauthorized"})
		return
	}
	schemaName := strings.ToLower(body.SchemaName)
	exist, schema, err := db.GetSchemaByName(schemaName, user.TenantName)
	if err != nil {
		serv.Errorf("CreateNewVersion: Schema" + body.SchemaName + ": " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server Error"})
		return
	}
	if !exist {
		errMsg := "Schema " + body.SchemaName + " does not exist"
		serv.Warnf("CreateNewVersion: " + errMsg)
		c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": errMsg})
		return
	}

	messageStructName := body.MessageStructName
	if schema.Type == "protobuf" {
		err := validateMessageStructName(messageStructName)
		if err != nil {
			serv.Errorf("CreateNewVersion: Schema " + body.SchemaName + ": " + err.Error())
			c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": err.Error()})
			return
		}
	}
	schemaContent := body.SchemaContent
	err = validateSchemaContent(schemaContent, schema.Type)
	if err != nil {
		serv.Warnf("CreateNewVersion: Schema " + body.SchemaName + ": " + err.Error())
		c.AbortWithStatusJSON(SCHEMA_VALIDATION_ERROR_STATUS_CODE, gin.H{"message": err.Error()})
		return
	}

	countVersions, err := db.GetShcemaVersionsCount(schema.ID, user.TenantName)
	if err != nil {
		serv.Errorf("CreateNewVersion: Schema " + body.SchemaName + ": " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	versionNumber := countVersions + 1
	descriptor := ""
	if schema.Type == "protobuf" {
		descriptor, err = generateSchemaDescriptor(schemaName, versionNumber, schemaContent, schema.Type)
		if err != nil {
			serv.Warnf("CreateNewVersion: Schema " + body.SchemaName + ": " + err.Error())
			c.AbortWithStatusJSON(SCHEMA_VALIDATION_ERROR_STATUS_CODE, gin.H{"message": err.Error()})
			return
		}
	}
	newSchemaVersion, rowsUpdated, err := db.InsertNewSchemaVersion(versionNumber, user.ID, user.Username, schemaContent, schema.ID, messageStructName, descriptor, false, user.TenantName)
	if err != nil {
		serv.Warnf("CreateNewVersion: " + err.Error())
		c.AbortWithStatusJSON(SCHEMA_VALIDATION_ERROR_STATUS_CODE, gin.H{"message": err.Error()})
		return
	}
	if rowsUpdated == 1 {
		message := "Schema Version " + strconv.Itoa(newSchemaVersion.VersionNumber) + " has been created by " + user.Username
		serv.Noticef(message)
	} else {
		serv.Warnf("CreateNewVersion: Schema " + body.SchemaName + ": Version " + strconv.Itoa(newSchemaVersion.VersionNumber) + " already exists")
		c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": "Version already exists"})
		return
	}
	extedndedSchemaDetails, err := sh.getExtendedSchemaDetails(schema, user.TenantName)
	if err != nil {
		serv.Errorf("CreateNewVersion: Schema " + body.SchemaName + ": " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	shouldSendAnalytics, _ := shouldSendAnalytics()
	if shouldSendAnalytics {
		analytics.SendEvent(user.Username, "user-create-new-schema-version")
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
		serv.Errorf("RollBackVersion: Schema " + body.SchemaName + ": " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server Error"})
		return
	}
	schemaName := strings.ToLower(body.SchemaName)
	exist, schema, err := db.GetSchemaByName(schemaName, user.TenantName)
	if err != nil {
		serv.Errorf("RollBackVersion: Schema " + body.SchemaName + ": " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server Error"})
		return
	}
	if !exist {
		errMsg := "Schema " + body.SchemaName + " does not exist"
		serv.Warnf("RollBackVersion: " + errMsg)
		c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": errMsg})
		return
	}

	schemaVersion := body.VersionNumber
	exist, _, err = db.GetSchemaVersionByNumberAndID(schemaVersion, schema.ID)
	if err != nil {
		serv.Errorf("RollBackVersion: Schema " + body.SchemaName + " version " + strconv.Itoa(schemaVersion) + ": " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	if !exist {
		errMsg := "Schema " + body.SchemaName + " version " + strconv.Itoa(schemaVersion) + " does not exist"
		serv.Warnf("RollBackVersion: " + errMsg)
		c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": errMsg})
		return
	}

	countVersions, err := db.GetShcemaVersionsCount(schema.ID, user.TenantName)
	if err != nil {
		serv.Errorf("RollBackVersion: Schema " + body.SchemaName + ": " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	if countVersions > 1 {
		err = db.UpdateSchemaActiveVersion(schema.ID, body.VersionNumber)
		if err != nil {
			serv.Errorf("RollBackVersion: Schema " + body.SchemaName + ": " + err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": err.Error()})
			return
		}
	}
	extedndedSchemaDetails, err = sh.getExtendedSchemaDetails(schema, user.TenantName)
	if err != nil {
		serv.Errorf("RollBackVersion: Schema " + body.SchemaName + ": " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	shouldSendAnalytics, _ := shouldSendAnalytics()
	if shouldSendAnalytics {
		analytics.SendEvent(user.Username, "user-rollback-schema-version")
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
		serv.Warnf("ValidateSchema: Schema type " + schemaType + ": " + err.Error())
		c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": err.Error()})
		return
	}

	schemaContent := body.SchemaContent
	err = validateSchemaContent(schemaContent, schemaType)
	if err != nil {
		serv.Warnf("ValidateSchema: Schema type " + schemaType + ": " + err.Error())
		c.AbortWithStatusJSON(SCHEMA_VALIDATION_ERROR_STATUS_CODE, gin.H{"message": err.Error()})
		return
	}

	shouldSendAnalytics, _ := shouldSendAnalytics()
	if shouldSendAnalytics {
		user, _ := getUserDetailsFromMiddleware(c)
		analytics.SendEvent(user.Username, "user-validate-schema")
	}

	c.IndentedJSON(200, gin.H{
		"is_valid": true,
	})
}
