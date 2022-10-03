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
	// exist, _, err := IsSchemaExist(schemaName)
	// if err != nil {
	// 	c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
	// 	return
	// }
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
	if err != nil && !strings.Contains(err.Error(), "no documents")  {
		serv.Errorf("CreateSchema error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	versionNumber := len(schema.Versions) + 1
	newSchemaVersion := models.SchemaVersion{
		ID:            primitive.NewObjectID(),
		VersionNumber: versionNumber,
		Active:        false,
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
	}else{
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
