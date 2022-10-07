package server

import (
	"context"
	"errors"
	"fmt"
	"memphis-broker/models"
	"memphis-broker/utils"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type SchemasHandler struct{ S *Server }

const (
	schemaObjectName = "Schema"
)

func validateSchemaName(schemaName string) error {
	return validateName(schemaName, schemaObjectName)
}

func validateSchemaType(schemaType string) error {
	invalidTypeErrStr := fmt.Sprintf("%v unsupported schema type", schemaType)
	invalidTypeErr := errors.New(invalidTypeErrStr)
	invalidSupportTypeErrStr := fmt.Sprintf("%v Json/Avro types are not supported at this time", schemaType)
	invalidSupportTypeErr := errors.New(invalidSupportTypeErrStr)

	if schemaType == "protobuf" {
		return nil
	} else if schemaType == "avro" || schemaType == "json" {
		return invalidSupportTypeErr
	} else {
		return invalidTypeErr
	}
}

func validateSchemaContent(schemaContent, schemaType string) error {
	switch schemaType {
	case "protobuf":
		if strings.Contains(schemaContent, "syntax") {
			return nil
		}
		break
	case "json":
		break
	case "avro":
		break
	default:
		invalidSchemaContentErrStr := fmt.Sprintf("%v Your Schema is invalid")
		invalidSchemaContentErr := errors.New(invalidSchemaContentErrStr)
		return invalidSchemaContentErr
	}
	return nil
}

func (sh SchemasHandler) GetSchemaDetailsBySchemaName(schemaName string) (models.ExtendedSchemaDetails, error) {
	var schema models.Schema
	err := schemasCollection.FindOne(context.TODO(), bson.M{"name": schemaName}).Decode(&schema)
	if err != nil {
		return models.ExtendedSchemaDetails{}, err
	}
	var schemaVersion []models.SchemaVersion
	filter := bson.M{"_id": bson.M{"$in": schema.Versions}}
	cursor, err := schemaVersionCollection.Find(context.TODO(), filter)
	if err != nil {
		return models.ExtendedSchemaDetails{}, err
	}
	if err = cursor.All(context.TODO(), &schemaVersion); err != nil {
		return models.ExtendedSchemaDetails{}, err
	}

	if len(schemaVersion) == 0 {
		return models.ExtendedSchemaDetails{}, err
	}
	extedndedSchemaDetails := models.ExtendedSchemaDetails{
		ID:         schema.ID,
		SchemaName: schema.Name,
		Type:       schema.Type,
		Versions:   schemaVersion,
	}
	return extedndedSchemaDetails, nil
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
			err = schemaVersionCollection.FindOne(context.TODO(), filter).Decode(&schemaVersion)
			if err != nil {
				return []models.ExtendedSchema{}, err
			}
			if schemaVersion.VersionNumber == 1 {
				extSchema := models.ExtendedSchema{
					ID:            schema.ID,
					Name:          schema.Name,
					Type:          schema.Type,
					CreatedByUser: schemaVersion.CreatedByUser,
					CreationDate:  schemaVersion.CreationDate,
				}
				extednedSchemas = append(extednedSchemas, extSchema)
			}
		}

	}

	if len(extednedSchemas) == 0 {
		return []models.ExtendedSchema{}, nil
	} else {
		return extednedSchemas, nil
	}
}

func (sh SchemasHandler) findAndDeleteSchema(schemaName string) error {
	var schema models.Schema
	filter := bson.M{"name": schemaName}
	err := schemasCollection.FindOne(context.TODO(), filter).Decode(&schema)
	if err != nil {
		return err
	}
	filter = bson.M{"_id": bson.M{"$in": schema.Versions}}
	_, err = schemaVersionCollection.DeleteMany(context.TODO(), filter)

	if err != nil {
		return err
	}
	_, err = schemasCollection.DeleteOne(context.TODO(), schema)

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
		c.AbortWithStatusJSON(configuration.SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": "Schema with that name already exists"})
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
	}
	schemaType := strings.ToLower(body.Type)
	err = validateSchemaType(schemaType)

	if err != nil {
		serv.Warnf(err.Error())
		c.AbortWithStatusJSON(configuration.SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": err.Error()})
		return
	}
	schemaContent := body.SchemaContent
	err = validateSchemaContent(schemaContent, schemaType)
	if err != nil {
		serv.Warnf(err.Error())
		c.AbortWithStatusJSON(configuration.SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": err.Error()})
		return
	}
	newSchemaVersion := models.SchemaVersion{
		ID:            primitive.NewObjectID(),
		VersionNumber: 1,
		Active:        true,
		CreatedByUser: user.Username,
		CreationDate:  time.Now(),
		SchemaContent: schemaContent,
	}
	newSchema := models.Schema{
		ID:       primitive.NewObjectID(),
		Name:     schemaName,
		Type:     schemaType,
		Versions: []primitive.ObjectID{newSchemaVersion.ID},
	}

	_, err = schemasCollection.InsertOne(context.TODO(), newSchema)
	if err != nil {
		serv.Errorf("CreateSchema error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	_, err = schemaVersionCollection.InsertOne(context.TODO(), newSchemaVersion)
	if err != nil {
		serv.Errorf("CreateSchema error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	message := "Schema " + schemaName + " has been created"
	serv.Noticef(message)
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
	var body models.GetSchemaDetails
	ok := utils.Validate(c, &body, false, nil)
	if !ok {
		return
	}
	schemaName := strings.ToLower(body.SchemaName)
	exist, _, err := IsSchemaExist(schemaName)
	if err != nil {
		c.AbortWithStatusJSON(configuration.SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": "Server error"})
		return
	}
	if !exist {
		serv.Warnf("Schema does not exist")
		c.AbortWithStatusJSON(configuration.SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": "Schema does not exist"})
		return
	}

	schemaDetails, err := sh.GetSchemaDetailsBySchemaName(schemaName)

	if err != nil {
		serv.Errorf("GetSchemaDetails error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	c.IndentedJSON(200, schemaDetails)
}

func (sh SchemasHandler) RemoveSchema(c *gin.Context) {
	var body models.RemoveSchema
	ok := utils.Validate(c, &body, false, nil)
	if !ok {
		return
	}
	schemaName := strings.ToLower(body.SchemaName)
	exist, _, err := IsSchemaExist(schemaName)
	if err != nil {
		c.AbortWithStatusJSON(configuration.SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": "Server error"})
		return
	}
	if !exist {
		serv.Warnf("Schema does not exist")
		c.AbortWithStatusJSON(configuration.SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": "Schema does not exist"})
		return
	}
	err = sh.findAndDeleteSchema(schemaName)

	if err != nil {
		serv.Errorf("RemoveSchema error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return

	}
	serv.Noticef("Schema " + schemaName + " has been deleted")
	c.IndentedJSON(200, gin.H{})
}
