package server

import (
	"context"
	"errors"
	"fmt"
	"memphis-broker/models"
	"memphis-broker/utils"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type SchemasHandler struct{ S *Server }
type SchemasVersionsHandler struct{ S *Server }

const (
	schemaObjectName = "Schema"
)

var (
	validKeyRe              = regexp.MustCompile(`\A[-/_=\.a-zA-Z0-9]+\z`)
	ErrObjectConfigRequired = errors.New("nats: object-store config required")
	ErrBadObjectMeta        = errors.New("nats: object-store meta information invalid")
	ErrObjectNotFound       = errors.New("nats: object not found")
	ErrInvalidStoreName     = errors.New("nats: invalid object-store name")
	ErrInvalidObjectName    = errors.New("nats: invalid object name")
	ErrDigestMismatch       = errors.New("nats: received a corrupt object, digests do not match")
	ErrNoObjectsFound       = errors.New("nats: no objects found")
)

func validateSchemaName(schemaName string) error {
	return validateName(schemaName, schemaObjectName)
}

func validateSchemaType(schemaType string) error {
	invalidTypeErrStr := fmt.Sprintf("%v should be protobuf", schemaType)
	invalidTypeErr := errors.New(invalidTypeErrStr)

	if schemaType == "protobuf" {
		return nil
	} else {
		return invalidTypeErr
	}

}

func validateSchemaContent(schemaContent string) error {
	invalidSchemaContentErrStr := fmt.Sprintf("%v should be protobuf", schemaContent)
	invalidSchemaContentErr := errors.New(invalidSchemaContentErrStr)
	if strings.Contains(schemaContent, "syntax") {
		return nil

	} else {
		return invalidSchemaContentErr

	}
}

func (sh SchemasHandler) GetSchemaDetailsByVersionNumber(schemaName string, versionNumber int) (models.SchemaVersion, error) {
	var schema models.Schema

	err := schemasCollection.FindOne(context.TODO(), bson.M{
		"name": schemaName,
	}).Decode(&schema)

	if err != nil {
		return models.SchemaVersion{}, err
	}

	var schemaVersion models.SchemaVersion
	filter := bson.M{
		"version_number": versionNumber,
		"$or": []interface{}{
			bson.M{"_id": bson.M{"$in": schema.Versions}},
		},
	}

	err = schemasVersionCollection.FindOne(context.TODO(), filter).Decode(&schemaVersion)
	if err != nil {
		return schemaVersion, err
	}
	return schemaVersion, nil
}

func (sh SchemasHandler) GetAllSchemasDetails() ([]models.ExtendedSchema, error) {
	var schemas []models.Schema
	cursor, err := schemasCollection.Aggregate(context.TODO(), mongo.Pipeline{
		bson.D{{"$match", bson.D{{"$or", []interface{}{
			bson.D{{"is_deleted", false}},
			bson.D{{"is_deleted", bson.D{{"$exists", false}}}},
		}}}}},
	})

	if err != nil {
		return []models.ExtendedSchema{}, err
	}
	if err = cursor.All(context.TODO(), &schemas); err != nil {
		return []models.ExtendedSchema{}, err
	}

	var schemaVersion models.SchemaVersion
	var extednedSchemas []models.ExtendedSchema
	for _, schema := range schemas {
		for _, version := range schema.Versions {
			filter := bson.M{"_id": version}
			err = schemasVersionCollection.FindOne(context.TODO(), filter).Decode(&schemaVersion)
			if schemaVersion.VersionNumber == 1 {
				extSchema := models.ExtendedSchema{
					ID:            schema.ID,
					Name:          schema.Name,
					Type:          schema.Type,
					CreatedByUser: schemaVersion.CreatedByUser,
					CreationDate:  schemaVersion.CreationDate,
				}
				extednedSchemas = append(extednedSchemas, extSchema)
				if err != nil {
					return []models.ExtendedSchema{}, err
				}
			}
		}

	}

	if len(extednedSchemas) == 0 {
		return []models.ExtendedSchema{}, nil
	} else {
		return extednedSchemas, nil
	}
}

func (sh SchemasHandler) GetAllSchemaVersionsByVersionNumber(versionNumber int) ([]models.SchemaVersion, error) {
	var schemasVersion []models.SchemaVersion

	filter := bson.M{"version_number": versionNumber}
	cursor, err := schemasVersionCollection.Find(context.TODO(), filter)
	if err != nil {
		return []models.SchemaVersion{}, err
	}
	if err = cursor.All(context.TODO(), &schemasVersion); err != nil {
		return schemasVersion, err
	}

	if err != nil {
		return []models.SchemaVersion{}, err
	}

	if len(schemasVersion) == 0 {
		return []models.SchemaVersion{}, nil
	} else {
		return schemasVersion, nil
	}
}

func (sh SchemasHandler) FindAndDeleteSchema(schemaName string) error {
	var schema models.Schema
	filter := bson.M{"name": schemaName}
	err := schemasCollection.FindOne(context.TODO(), filter).Decode(&schema)
	if err != nil {
		return err
	}
	schemaVersions := schema.Versions
	var schemaVersion models.SchemaVersion
	for _, version := range schemaVersions {
		filterVersion := bson.M{"_id": version}
		err = schemasVersionCollection.FindOneAndDelete(context.TODO(), filterVersion).Decode(&schemaVersion)
		if err != nil {
			return err
		}
	}
	_, err = schemasCollection.DeleteOne(context.TODO(), filter)

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
		c.AbortWithStatusJSON(500, gin.H{"message": err.Error()})
		return
	}
	exist, _, err := IsSchemaExist(schemaName)
	if exist {
		serv.Warnf("Schema with that name already exists")
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	user, err := getUserDetailsFromMiddleware(c)
	if err != nil {
		serv.Errorf("CreateStation error: " + err.Error())
		c.AbortWithStatusJSON(401, gin.H{"message": "Unauthorized"})
	}

	schemaType := strings.ToLower(body.Type)
	err = validateSchemaType(schemaType)

	if err != nil {
		serv.Warnf(err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": err.Error()})
		return
	}
	schemaContent := strings.ToLower(body.SchemaContent)
	err = validateSchemaContent(schemaContent)
	if err != nil {
		serv.Warnf(err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": err.Error()})
		return
	}
	filter := bson.M{"name": schemaName}
	var schema models.Schema
	err = schemasCollection.FindOne(context.TODO(), filter).Decode(&schema)
	if err != nil && !strings.Contains(err.Error(), "no documents") {
		serv.Errorf("CreateSchema error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	versionNumber := len(schema.Versions) + 1
	var active bool
	if versionNumber == 1 {
		active = true

	} else {
		active = false
	}
	newSchemaVersion := models.SchemaVersion{
		ID:            primitive.NewObjectID(),
		VersionNumber: versionNumber,
		Active:        active,
		CreatedByUser: user.Username,
		CreationDate:  time.Now(),
		SchemaContent: schemaContent,
	}
	newSchema := models.Schema{
		ID:        primitive.NewObjectID(),
		Name:      schemaName,
		Type:      schemaType,
		Versions:  schema.Versions,
		IsDeleted: false,
	}

	filter = bson.M{"name": newSchema.Name, "is_deleted": false}
	update := bson.M{
		"$setOnInsert": bson.M{
			"_id":  newSchema.ID,
			"name": newSchema.Name,
			"type": newSchema.Type,
		},
		"$push": bson.M{"versions": newSchemaVersion.ID},
	}
	opts := options.Update().SetUpsert(true)
	updateResults, err := schemasCollection.UpdateOne(context.TODO(), filter, update, opts)
	if err != nil {
		serv.Errorf("CreateSchema error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	if updateResults.MatchedCount > 0 {
		serv.Warnf("Scheam with the same name is already exist")
		newSchema.ID = schema.ID
	} else {
		message := "Schema " + schemaName + " has been created"
		serv.Noticef(message)
		var auditLogs []interface{}
		newAuditLog := models.AuditLog{
			ID:            primitive.NewObjectID(),
			StationName:   schemaName,
			Message:       message,
			CreatedByUser: user.Username,
			CreationDate:  time.Now(),
			UserType:      user.UserType,
		}
		auditLogs = append(auditLogs, newAuditLog)
		err = CreateAuditLogs(auditLogs)
		if err != nil {
			serv.Warnf("CreateSchema error: " + err.Error())
		}
	}

	filter = bson.M{"_id": newSchemaVersion.ID, "is_deleted": false}
	update = bson.M{
		"$setOnInsert": bson.M{
			"_id":             newSchemaVersion.ID,
			"version_number":  newSchemaVersion.VersionNumber,
			"active":          newSchemaVersion.Active,
			"created_by_user": newSchemaVersion.CreatedByUser,
			"creation_date":   newSchemaVersion.CreationDate,
			"schema_content":  newSchemaVersion.SchemaContent,
		},
	}
	_, err = schemasVersionCollection.UpdateOne(context.TODO(), filter, update, opts)

	if err != nil {
		serv.Errorf("CreateSchema error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	newSchema.Versions = append(newSchema.Versions, newSchemaVersion.ID)
	c.IndentedJSON(200, newSchema)
}

func (sh SchemasHandler) GetAllSchemas(c *gin.Context) {
	schemas, err := sh.GetAllSchemasDetails()
	if err != nil {
		serv.Errorf("GetAllSchemas error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	c.IndentedJSON(200, schemas)
}

func (sh SchemasHandler) GetSchemaDetails(c *gin.Context) {
	var body models.GetSchemaDetailsByVersionNumber
	ok := utils.Validate(c, &body, false, nil)
	if !ok {
		return
	}
	schemaName := strings.ToLower(body.SchemaName)
	schemaDetails, err := sh.GetSchemaDetailsByVersionNumber(schemaName, body.VersionNumber)

	if err != nil {
		serv.Errorf("GetSchemaDetails error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	extedndedSchemaDetails := models.ExtendedSchemaDetails{
		ID:            schemaDetails.ID,
		SchemaName:    schemaName,
		VersionNumber: schemaDetails.VersionNumber,
		Active:        schemaDetails.Active,
		CreatedByUser: schemaDetails.CreatedByUser,
		CreationDate:  schemaDetails.CreationDate,
		SchemaContent: schemaDetails.SchemaContent,
	}
	c.IndentedJSON(200, extedndedSchemaDetails)
}

func (sh SchemasHandler) RemoveSchema(c *gin.Context) {
	var body models.RemoveSchema
	ok := utils.Validate(c, &body, false, nil)
	if !ok {
		return
	}
	schemaName := strings.ToLower(body.SchemaName)
	err := sh.FindAndDeleteSchema(schemaName)

	if err != nil {
		serv.Errorf("RemoveSchema error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return

	}
	c.IndentedJSON(200, []string{})
}
