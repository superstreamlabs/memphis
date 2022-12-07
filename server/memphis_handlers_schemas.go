package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"memphis-broker/analytics"
	"memphis-broker/models"
	"memphis-broker/utils"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	graphqlParser "github.com/graphql-go/graphql/language/parser"
	"github.com/jhump/protoreflect/desc/protoparse"
	"github.com/santhosh-tekuri/jsonschema/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type SchemasHandler struct{ S *Server }

const (
	schemaObjectName                    = "Schema"
	SCHEMA_VALIDATION_ERROR_STATUS_CODE = 555
	schemaUpdatesSubjectTemplate        = "$memphis_schema_updates_%s"
)

var (
	ErrNoSchema = errors.New("No schemas found")
)

func validateProtobufContent(schemaContent string) error {
	parser := protoparse.Parser{
		Accessor: func(filename string) (io.ReadCloser, error) {
			return ioutil.NopCloser(strings.NewReader(schemaContent)), nil
		},
	}
	_, err := parser.ParseFiles("")
	if err != nil {
		return errors.New("Your Proto file is invalid: " + err.Error())
	}

	return nil
}

func validateJsonSchemaContent(schemaContent string) error {
	_, err := jsonschema.CompileString("test", schemaContent)
	if err != nil {
		return errors.New("Your json schema is invalid")
	}

	return nil
}

func validateGraphqlSchemaContent(schemaContent string) error {
	parseParams := graphqlParser.ParseParams{Source: schemaContent}
	_, err := graphqlParser.Parse(parseParams)
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

	descContent, err := ioutil.ReadFile(descFilename)
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
	invalidTypeErrStr := fmt.Sprintf("unsupported schema type")
	invalidTypeErr := errors.New(invalidTypeErrStr)
	invalidSupportTypeErrStr := fmt.Sprintf("Avro is not supported at this time")
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
		return errors.New("Your schema content is invalid")
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
		break
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
		return "", errors.New("Descriptor generation with schema type: " + schemaType + ", while protobuf is expected")
	}

	descriptor, err := generateProtobufDescriptor(schemaName, schemaVersionNum, schemaContent)
	if err != nil {
		return "", err
	}
	return string(descriptor), nil
}

func validateMessageStructName(messageStructName string) error {
	if messageStructName == "" {
		return errors.New("Message struct name is required when schema type is Protobuf")
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

func getSchemaUpdateInitFromStation(sn StationName) (*models.ProducerSchemaUpdateInit, error) {
	schema, err := getSchemaByStationName(sn)
	if err != nil {
		return nil, err
	}

	return generateSchemaUpdateInit(schema)
}

func (s *Server) updateStationProducersOfSchemaChange(sn StationName, schemaUpdate models.ProducerSchemaUpdate) {
	subject := fmt.Sprintf(schemaUpdatesSubjectTemplate, sn.Intern())
	msg, err := json.Marshal(schemaUpdate)
	if err != nil {
		s.Errorf("schema change update: marshal failed")
		return
	}
	s.sendInternalAccountMsg(s.GlobalAccount(), subject, msg)
}

func getSchemaVersionsBySchemaId(id primitive.ObjectID) ([]models.SchemaVersion, error) {
	var schemaVersions []models.SchemaVersion
	filter := bson.M{"schema_id": id}
	findOptions := options.Find()
	findOptions.SetSort(bson.M{"creation_date": -1})

	cursor, err := schemaVersionCollection.Find(context.TODO(), filter, findOptions)
	if err != nil {
		return []models.SchemaVersion{}, err
	}
	if err = cursor.All(context.TODO(), &schemaVersions); err != nil {
		return []models.SchemaVersion{}, err
	}

	return schemaVersions, nil
}

func getActiveVersionBySchemaId(schemaId primitive.ObjectID) (models.SchemaVersion, error) {
	var schemaVersion models.SchemaVersion
	err := schemaVersionCollection.FindOne(context.TODO(), bson.M{"schema_id": schemaId, "active": true}).Decode(&schemaVersion)
	if err != nil {
		return models.SchemaVersion{}, err
	}
	return schemaVersion, nil
}

func getSchemaByStationName(sn StationName) (models.Schema, error) {
	var schema models.Schema

	exist, station, err := IsStationExist(sn)
	if err != nil {
		serv.Errorf("getSchemaByStation error: " + err.Error())
		return schema, err
	}
	if !exist {
		serv.Warnf("Station " + station.Name + " does not exist")
		return schema, err
	}
	if station.Schema.SchemaName == "" {
		return schema, ErrNoSchema
	}

	err = schemasCollection.FindOne(context.TODO(), bson.M{"name": station.Schema.SchemaName}).Decode(&schema)
	if err != nil {
		serv.Errorf("getSchemaByStation error: " + err.Error())
		return schema, err
	}

	return schema, nil
}

func (sh SchemasHandler) GetSchemaByStationName(stationName StationName) (models.Schema, error) {
	return getSchemaByStationName(stationName)
}

func (sh SchemasHandler) GetSchemaVersion(stationVersion int, schemaId primitive.ObjectID) (models.SchemaVersion, error) {
	var schemaVersion models.SchemaVersion
	err := schemaVersionCollection.FindOne(context.TODO(), bson.M{"schema_id": schemaId, "version_number": stationVersion}).Decode(&schemaVersion)
	if err != nil {
		return models.SchemaVersion{}, err
	}

	return schemaVersion, nil
}

func (sh SchemasHandler) updateActiveVersion(schemaId primitive.ObjectID, versionNumber int) error {
	_, err := schemaVersionCollection.UpdateMany(context.TODO(),
		bson.M{"schema_id": schemaId},
		bson.M{"$set": bson.M{"active": false}},
	)
	if err != nil {
		return err
	}

	_, err = schemaVersionCollection.UpdateOne(context.TODO(), bson.M{"schema_id": schemaId, "version_number": versionNumber}, bson.M{"$set": bson.M{"active": true}})
	if err != nil {
		return err
	}
	return nil
}

func (sh SchemasHandler) getVersionsCount(schemaId primitive.ObjectID) (int, error) {
	countVersions, err := schemaVersionCollection.CountDocuments(context.TODO(), bson.M{"schema_id": schemaId})
	if err != nil {
		return 0, err
	}

	return int(countVersions), err
}

func (sh SchemasHandler) getSchemaVersionsBySchemaId(schemaId primitive.ObjectID) ([]models.SchemaVersion, error) {
	return getSchemaVersionsBySchemaId(schemaId)
}

func (sh SchemasHandler) getUsingStationsByName(schemaName string) ([]string, error) {
	var stations []models.Station
	cursor, err := stationsCollection.Aggregate(context.TODO(), mongo.Pipeline{
		bson.D{{"$unwind", bson.D{{"path", "$schema"}, {"preserveNullAndEmptyArrays", true}}}},
		bson.D{{"$match", bson.D{{"schema.name", schemaName}, {"is_deleted", false}}}},
		bson.D{{"$project", bson.D{{"name", 1}}}},
	})
	if err != nil {
		return []string{}, err
	}

	if err = cursor.All(context.TODO(), &stations); err != nil {
		return []string{}, err
	}
	if len(stations) == 0 {
		return []string{}, nil
	}

	var stationNames []string
	for _, station := range stations {
		stationNames = append(stationNames, station.Name)
	}

	return stationNames, nil
}

func (sh SchemasHandler) getStationsBySchemaCount(schemaName string) (int, error) {
	filter := bson.M{"schema.name": schemaName, "is_deleted": false}
	countStations, err := stationsCollection.CountDocuments(context.TODO(), filter)
	if err != nil {
		return 0, err
	}

	return int(countStations), nil

}

func (sh SchemasHandler) getExtendedSchemaDetailsUpdateAvailable(schemaVersion int, schema models.Schema) (models.ExtendedSchemaDetails, error) {
	var schemaVersions []models.SchemaVersion
	usedSchemaVersion, err := sh.GetSchemaVersion(schemaVersion, schema.ID)
	if err != nil {
		return models.ExtendedSchemaDetails{}, err
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
	stations, err := sh.getUsingStationsByName(schema.Name)
	if err != nil {
		return models.ExtendedSchemaDetails{}, err
	}

	tagsHandler := TagsHandler{S: sh.S}
	tags, err := tagsHandler.GetTagsBySchema(schema.ID)
	if err != nil {
		return models.ExtendedSchemaDetails{}, err
	}

	extedndedSchemaDetails = models.ExtendedSchemaDetails{
		ID:           schema.ID,
		SchemaName:   schema.Name,
		Type:         schema.Type,
		Versions:     schemaVersions,
		UsedStations: stations,
		Tags:         tags,
	}

	return extedndedSchemaDetails, nil
}

func (sh SchemasHandler) getExtendedSchemaDetails(schema models.Schema) (models.ExtendedSchemaDetails, error) {
	schemaVersions, err := sh.getSchemaVersionsBySchemaId(schema.ID)
	if err != nil {
		return models.ExtendedSchemaDetails{}, err
	}

	var extedndedSchemaDetails models.ExtendedSchemaDetails
	stations, err := sh.getUsingStationsByName(schema.Name)
	if err != nil {
		return models.ExtendedSchemaDetails{}, err
	}

	tagsHandler := TagsHandler{S: sh.S}
	tags, err := tagsHandler.GetTagsBySchema(schema.ID)
	if err != nil {
		return models.ExtendedSchemaDetails{}, err
	}

	extedndedSchemaDetails = models.ExtendedSchemaDetails{
		ID:           schema.ID,
		SchemaName:   schema.Name,
		Type:         schema.Type,
		Versions:     schemaVersions,
		UsedStations: stations,
		Tags:         tags,
	}

	return extedndedSchemaDetails, nil
}

func (sh SchemasHandler) getSchemaDetailsBySchemaName(schemaName string) (models.ExtendedSchemaDetails, error) {
	var schema models.Schema
	err := schemasCollection.FindOne(context.TODO(), bson.M{"name": schemaName}).Decode(&schema)
	if err != nil {
		return models.ExtendedSchemaDetails{}, err
	}

	extedndedSchemaDetails, err := sh.getExtendedSchemaDetails(schema)
	if err != nil {
		return models.ExtendedSchemaDetails{}, err
	}

	return extedndedSchemaDetails, nil
}

func (sh SchemasHandler) GetAllSchemasDetails() ([]models.ExtendedSchema, error) {
	var schemas []models.ExtendedSchema
	cursor, err := schemasCollection.Aggregate(context.TODO(), mongo.Pipeline{
		bson.D{{"$lookup", bson.D{{"from", "schema_versions"}, {"localField", "_id"}, {"foreignField", "schema_id"}, {"as", "extendedSchema"}}}},
		bson.D{{"$unwind", bson.D{{"path", "$extendedSchema"}, {"preserveNullAndEmptyArrays", true}}}},
		bson.D{{"$match", bson.D{{"extendedSchema.version_number", 1}}}},
		bson.D{{"$lookup", bson.D{{"from", "schema_versions"}, {"localField", "_id"}, {"foreignField", "schema_id"}, {"as", "activeVersion"}}}},
		bson.D{{"$unwind", bson.D{{"path", "$activeVersion"}, {"preserveNullAndEmptyArrays", true}}}},
		bson.D{{"$match", bson.D{{"activeVersion.active", true}}}},
		bson.D{{"$project", bson.D{{"_id", 1}, {"name", 1}, {"type", 1}, {"created_by_user", "$extendedSchema.created_by_user"}, {"creation_date", "$extendedSchema.creation_date"}, {"version_number", "$activeVersion.version_number"}}}},
		bson.D{{"$sort", bson.D{{"creation_date", -1}}}},
	})
	if err != nil {
		return []models.ExtendedSchema{}, err
	}

	if err = cursor.All(context.TODO(), &schemas); err != nil {
		return []models.ExtendedSchema{}, err
	}
	if len(schemas) == 0 {
		return []models.ExtendedSchema{}, nil
	}

	var extedndedSchemasDetails []models.ExtendedSchema
	for i, schema := range schemas {
		stations, err := sh.getStationsBySchemaCount(schema.Name)
		if err != nil {
			return []models.ExtendedSchema{}, err
		}

		var used bool
		if stations > 0 {
			used = true
		} else {
			used = false
		}

		tagsHandler := TagsHandler{S: sh.S}
		tags, err := tagsHandler.GetTagsBySchema(schemas[i].ID)
		if err != nil {
			return []models.ExtendedSchema{}, err
		}
		schemaUpdated := models.ExtendedSchema{
			ID:                  schema.ID,
			Name:                schema.Name,
			Type:                schema.Type,
			CreatedByUser:       schema.CreatedByUser,
			CreationDate:        schema.CreationDate,
			ActiveVersionNumber: schema.ActiveVersionNumber,
			Used:                used,
			Tags:                tags,
		}

		extedndedSchemasDetails = append(extedndedSchemasDetails, schemaUpdated)
	}
	if err != nil {
		return []models.ExtendedSchema{}, err
	}
	return extedndedSchemasDetails, nil
}

func (sh SchemasHandler) findAndDeleteSchema(schemaIds []primitive.ObjectID) error {
	filter := bson.M{"schema_id": bson.M{"$in": schemaIds}}
	_, err := schemaVersionCollection.DeleteMany(context.TODO(), filter)
	if err != nil {
		return err
	}

	filter = bson.M{"_id": bson.M{"$in": schemaIds}}
	_, err = schemasCollection.DeleteMany(context.TODO(), filter)
	if err != nil {
		return err
	}
	return nil
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
		serv.Warnf(err.Error())
		c.AbortWithStatusJSON(configuration.SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": err.Error()})
		return
	}
	exist, _, err := IsSchemaExist(schemaName)
	if err != nil {
		serv.Errorf("CreateNewSchema error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server Error"})
		return
	}
	if exist {
		serv.Warnf("Schema with that name already exists")
		c.AbortWithStatusJSON(configuration.SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": "Schema with that name already exists"})
		return
	}
	user, err := getUserDetailsFromMiddleware(c)
	if err != nil {
		serv.Errorf("CreateNewSchema error: " + err.Error())
		c.AbortWithStatusJSON(401, gin.H{"message": "Unauthorized"})
		return
	}
	schemaType := strings.ToLower(body.Type)
	err = validateSchemaType(schemaType)
	if err != nil {
		serv.Warnf(err.Error())
		c.AbortWithStatusJSON(configuration.SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": err.Error()})
		return
	}
	messageStructName := body.MessageStructName
	if schemaType == "protobuf" {
		err := validateMessageStructName(messageStructName)
		if err != nil {
			serv.Warnf(err.Error())
			c.AbortWithStatusJSON(configuration.SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": err.Error()})
			return
		}
	}

	schemaContent := body.SchemaContent
	err = validateSchemaContent(schemaContent, schemaType)
	if err != nil {
		serv.Warnf(err.Error())
		c.AbortWithStatusJSON(SCHEMA_VALIDATION_ERROR_STATUS_CODE, gin.H{"message": err.Error()})
		return
	}
	schemaVersionNumber := 1
	descriptor := ""
	if schemaType == "protobuf" {
		descriptor, err = generateSchemaDescriptor(schemaName, schemaVersionNumber, schemaContent, schemaType)
		if err != nil {
			serv.Warnf(err.Error())
			c.AbortWithStatusJSON(configuration.SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": err.Error()})
			return
		}
	}

	newSchema := models.Schema{
		ID:   primitive.NewObjectID(),
		Name: schemaName,
		Type: schemaType,
	}

	filter := bson.M{"name": newSchema.Name}
	update := bson.M{
		"$setOnInsert": bson.M{
			"_id":  newSchema.ID,
			"type": newSchema.Type,
		},
	}

	newSchemaVersion := models.SchemaVersion{
		ID:                primitive.NewObjectID(),
		VersionNumber:     schemaVersionNumber,
		Active:            true,
		CreatedByUser:     user.Username,
		CreationDate:      time.Now(),
		SchemaContent:     schemaContent,
		SchemaId:          newSchema.ID,
		MessageStructName: messageStructName,
		Descriptor:        descriptor,
	}

	opts := options.Update().SetUpsert(true)
	updateResults, err := schemasCollection.UpdateOne(context.TODO(), filter, update, opts)
	if err != nil {
		serv.Errorf("CreateSchema error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	if updateResults.MatchedCount == 0 {
		_, err = schemaVersionCollection.InsertOne(context.TODO(), newSchemaVersion)
		if err != nil {
			serv.Errorf("CreateSchema error: " + err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}
		message := "Schema " + schemaName + " has been created by " + user.Username
		serv.Noticef(message)
	} else {
		serv.Warnf("Schema with that name already exists")
		c.AbortWithStatusJSON(configuration.SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": "Schema with that name already exists"})
		return
	}

	if len(body.Tags) > 0 {
		err = AddTagsToEntity(body.Tags, "schema", newSchema.ID)
		if err != nil {
			serv.Errorf("Failed creating tag: %v", err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}
	}

	shouldSendAnalytics, _ := shouldSendAnalytics()
	if shouldSendAnalytics {
		user, _ := getUserDetailsFromMiddleware(c)
		analytics.SendEvent(user.Username, "user-create-schema")
	}

	c.IndentedJSON(200, newSchema)
}

func (sh SchemasHandler) GetAllSchemas(c *gin.Context) {
	schemas, err := sh.GetAllSchemasDetails()
	if err != nil {
		serv.Errorf("GetAllSchemas error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	shouldSendAnalytics, _ := shouldSendAnalytics()
	if shouldSendAnalytics {
		user, _ := getUserDetailsFromMiddleware(c)
		analytics.SendEvent(user.Username, "user-enter-schema-page")
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
	exist, _, err := IsSchemaExist(schemaName)
	if err != nil {
		serv.Errorf("GetSchemaDetails error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	if !exist {
		serv.Warnf("Schema does not exist")
		c.AbortWithStatusJSON(configuration.SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": "Schema does not exist"})
		return
	}

	schemaDetails, err := sh.getSchemaDetailsBySchemaName(schemaName)
	if err != nil {
		serv.Errorf("GetSchemaDetails error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	shouldSendAnalytics, _ := shouldSendAnalytics()
	if shouldSendAnalytics {
		user, _ := getUserDetailsFromMiddleware(c)
		analytics.SendEvent(user.Username, "user-enter-schema-details")
	}

	c.IndentedJSON(200, schemaDetails)
}

func deleteSchemaFromStation(s *Server, schemaName string) error {
	var stations []models.Station
	cursor, err := stationsCollection.Find(nil, bson.M{"schema.name": schemaName})
	if err != nil {
		return err
	}

	if err = cursor.All(nil, &stations); err != nil {
		return err
	}

	for _, station := range stations {
		sn, err := StationNameFromStr(station.Name)
		if err != nil {
			return err
		}
		exist, station, err := IsStationExist(sn)
		if err != nil {
			s.Errorf("deleteSchemaFromStation error: " + err.Error())
			return err
		}
		if !exist {
			serv.Warnf("Station " + station.Name + " does not exist")
			continue
		}

		removeSchemaFromStation(s, sn, false)
	}

	_, err = stationsCollection.UpdateMany(context.TODO(),
		bson.M{
			"schema.name": schemaName,
		},
		bson.M{"$set": bson.M{"schema": bson.M{}}},
	)
	if err != nil {
		return err
	}

	return nil
}

func (sh SchemasHandler) RemoveSchema(c *gin.Context) {
	if err := DenyForSandboxEnv(c); err != nil {
		return
	}

	var body models.RemoveSchema
	ok := utils.Validate(c, &body, false, nil)
	if !ok {
		return
	}
	var schemaIds []primitive.ObjectID

	for _, name := range body.SchemaNames {
		schemaName := strings.ToLower(name)
		exist, schema, err := IsSchemaExist(schemaName)
		if err != nil {
			serv.Errorf("RemoveSchema error: " + err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}
		if exist {
			DeleteTagsFromSchema(schema.ID)
			err := deleteSchemaFromStation(sh.S, schema.Name)
			if err != nil {
				serv.Errorf("RemoveSchema error: " + err.Error())
				c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
				return
			}

			schemaIds = append(schemaIds, schema.ID)
		}
	}

	if len(schemaIds) > 0 {
		err := sh.findAndDeleteSchema(schemaIds)
		if err != nil {
			serv.Errorf("RemoveSchema error: " + err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}
		for _, name := range body.SchemaNames {
			serv.Noticef("Schema " + name + " has been deleted")
		}
	}

	shouldSendAnalytics, _ := shouldSendAnalytics()
	if shouldSendAnalytics {
		user, _ := getUserDetailsFromMiddleware(c)
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

	schemaName := strings.ToLower(body.SchemaName)
	exist, schema, err := IsSchemaExist(schemaName)
	if err != nil {
		serv.Errorf("CreateNewVersion error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server Error"})
		return
	}
	if !exist {
		serv.Warnf("Schema does not exist")
		c.AbortWithStatusJSON(configuration.SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": "Schema does not exist"})
		return
	}

	user, err := getUserDetailsFromMiddleware(c)
	if err != nil {
		serv.Errorf("CreateNewVersion error: " + err.Error())
		c.AbortWithStatusJSON(401, gin.H{"message": "Unauthorized"})
		return
	}

	messageStructName := body.MessageStructName
	if schema.Type == "protobuf" {
		err := validateMessageStructName(messageStructName)
		if err != nil {
			serv.Warnf(err.Error())
			c.AbortWithStatusJSON(configuration.SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": err.Error()})
			return
		}
	}
	schemaContent := body.SchemaContent
	err = validateSchemaContent(schemaContent, schema.Type)
	if err != nil {
		serv.Warnf(err.Error())
		c.AbortWithStatusJSON(SCHEMA_VALIDATION_ERROR_STATUS_CODE, gin.H{"message": err.Error()})
		return
	}

	countVersions, err := sh.getVersionsCount(schema.ID)
	if err != nil {
		serv.Errorf("CreateNewVersion error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	versionNumber := countVersions + 1
	descriptor := ""
	if schema.Type == "protobuf" {
		descriptor, err = generateSchemaDescriptor(schemaName, versionNumber, schemaContent, schema.Type)
		if err != nil {
			serv.Warnf(err.Error())
			c.AbortWithStatusJSON(SCHEMA_VALIDATION_ERROR_STATUS_CODE, gin.H{"message": err.Error()})
			return
		}
	}
	newSchemaVersion := models.SchemaVersion{
		ID:                primitive.NewObjectID(),
		VersionNumber:     versionNumber,
		Active:            false,
		CreatedByUser:     user.Username,
		CreationDate:      time.Now(),
		SchemaContent:     schemaContent,
		SchemaId:          schema.ID,
		MessageStructName: messageStructName,
		Descriptor:        descriptor,
	}

	filter := bson.M{"schema_id": schema.ID, "version_number": newSchemaVersion.VersionNumber}
	update := bson.M{
		"$setOnInsert": bson.M{
			"_id":                 newSchemaVersion.ID,
			"active":              newSchemaVersion.Active,
			"created_by_user":     newSchemaVersion.CreatedByUser,
			"creation_date":       newSchemaVersion.CreationDate,
			"schema_content":      newSchemaVersion.SchemaContent,
			"message_struct_name": newSchemaVersion.MessageStructName,
			"descriptor":          newSchemaVersion.Descriptor,
		},
	}

	opts := options.Update().SetUpsert(true)
	updateResults, err := schemaVersionCollection.UpdateOne(context.TODO(), filter, update, opts)
	if err != nil {
		serv.Errorf("CreateNewVersion error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	if updateResults.MatchedCount == 0 {
		message := "Schema Version " + strconv.Itoa(newSchemaVersion.VersionNumber) + " has been created by " + user.Username
		serv.Noticef(message)
	} else {
		serv.Warnf("Version already exists")
		c.AbortWithStatusJSON(configuration.SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": "Version already exists"})
		return
	}
	extedndedSchemaDetails, err := sh.getExtendedSchemaDetails(schema)
	if err != nil {
		serv.Errorf("CreateNewVersion error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	shouldSendAnalytics, _ := shouldSendAnalytics()
	if shouldSendAnalytics {
		user, _ := getUserDetailsFromMiddleware(c)
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

	schemaName := strings.ToLower(body.SchemaName)

	exist, schema, err := IsSchemaExist(schemaName)
	if err != nil {
		serv.Errorf("RollBackVersion error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server Error"})
		return
	}
	if !exist {
		serv.Warnf("Schema does not exist")
		c.AbortWithStatusJSON(configuration.SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": "Schema does not exist"})
		return
	}

	schemaVersion := body.VersionNumber
	exist, _, err = isSchemaVersionExists(schemaVersion, schema.ID)

	if err != nil {
		serv.Errorf("RollBackVersion error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	if !exist {
		serv.Warnf("Schema Version does not exist")
		c.AbortWithStatusJSON(configuration.SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": "Schema version does not exist"})
		return
	}

	countVersions, err := sh.getVersionsCount(schema.ID)
	if err != nil {
		serv.Errorf("RollBackVersion error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	if countVersions > 1 {
		err = sh.updateActiveVersion(schema.ID, body.VersionNumber)
		if err != nil {
			serv.Errorf("RollBackVersion error: " + err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": err.Error()})
			return
		}
	}
	extedndedSchemaDetails, err = sh.getExtendedSchemaDetails(schema)
	if err != nil {
		serv.Errorf("RollBackVersion error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	shouldSendAnalytics, _ := shouldSendAnalytics()
	if shouldSendAnalytics {
		user, _ := getUserDetailsFromMiddleware(c)
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
		serv.Warnf(err.Error())
		c.AbortWithStatusJSON(configuration.SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": err.Error()})
		return
	}

	schemaContent := body.SchemaContent
	err = validateSchemaContent(schemaContent, schemaType)
	if err != nil {
		serv.Warnf(err.Error())
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
