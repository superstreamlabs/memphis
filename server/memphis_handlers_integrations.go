// Copyright 2022-2023 The Memphis.dev Authors
// Licensed under the Memphis Business Source License 1.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// Changed License: [Apache License, Version 2.0 (https://www.apache.org/licenses/LICENSE-2.0), as published by the Apache Foundation.
//
// https://github.com/memphisdev/memphis-broker/blob/master/LICENSE
//
// Additional Use Grant: You may make use of the Licensed Work (i) only as part of your own product or service, provided it is not a message broker or a message queue product or service; and (ii) provided that you do not use, provide, distribute, or make available the Licensed Work as a Service.
// A "Service" is a commercial offering, product, hosted, or managed service, that allows third parties (other than your own employees and contractors acting on your behalf) to access and/or use the Licensed Work or a substantial set of the features or functionality of the Licensed Work to third parties as a software-as-a-service, platform-as-a-service, infrastructure-as-a-service or other similar services that compete with Licensor products or services.
package server

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"memphis-broker/analytics"
	"memphis-broker/models"
	"memphis-broker/utils"

	"github.com/gin-gonic/gin"
	"github.com/slack-go/slack"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

const sendNotificationType = "send_notification"

type IntegrationsHandler struct{ S *Server }

type MyProvider struct {
	AccessKeyID     string `json:"access_key_id"`
	SecretAccessKey string `json:"secret_acess_key_id"`
}

func (it IntegrationsHandler) CreateS3Integrtation(keys map[string]string) (map[string]string, error) {
	accessKey := keys["access_key"]
	secretKey := keys["secret_key"]
	region := keys["region"]
	bucketName := keys["bucket_name"]

	provider := &credentials.StaticProvider{Value: credentials.Value{
		AccessKeyID:     accessKey,
		SecretAccessKey: secretKey,
	}}

	_, err := provider.Retrieve()
	if err != nil {
		err = errors.New("Retrive failure " + err.Error())
		return map[string]string{}, err
	}

	credentials := credentials.NewCredentials(provider)

	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(region),
		Credentials: credentials},
	)
	if err != nil {
		err = errors.New("NewSession failure " + err.Error())
		return map[string]string{}, err
	}

	svc := s3.New(sess)
	_, err = svc.HeadBucket(&s3.HeadBucketInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		err = errors.New("create a S3 client with additional configuration failure " + err.Error())
		return map[string]string{}, err
	}

	acl, err := svc.GetBucketAcl(&s3.GetBucketAclInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		err = errors.New("GetBucketAcl error" + err.Error())
		return map[string]string{}, err
	}

	permission := *acl.Grants[0].Permission
	permissionValue := permission

	if permissionValue != "FULL_CONTROL" {
		err = errors.New("you should full control permission: read, write and delete " + err.Error())
		return map[string]string{}, err
	}

	uploader := s3manager.NewUploader(sess)

	if configuration.SERVER_NAME == "" {
		configuration.SERVER_NAME = "memphis"
	}

	reader := strings.NewReader(string("test") + " " + configuration.SERVER_NAME)
	// Upload the object to S3.
	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(configuration.SERVER_NAME),
		Body:   reader,
	})
	if err != nil {
		err = errors.New("failed to upload the obeject to S3 " + err.Error())
		return map[string]string{}, err
	}

	serv.Noticef("Object " + *aws.String(configuration.SERVER_NAME) + " successfully uploaded to S3")

	//delete the object
	_, err = svc.DeleteObject(&s3.DeleteObjectInput{Bucket: aws.String(bucketName), Key: aws.String(configuration.SERVER_NAME)})
	if err != nil {
		err = errors.New("Unable to delete object from bucket " + bucketName + err.Error())
		return map[string]string{}, err
	}
	err = svc.WaitUntilObjectNotExists(&s3.HeadObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(configuration.SERVER_NAME),
	})
	if err != nil {
		err = errors.New("Error occurred while waiting for object to be deleted from bucket " + bucketName + err.Error())
		return map[string]string{}, err
	}
	serv.Noticef("Object " + *aws.String(configuration.SERVER_NAME) + " successfully deleted")

	return keys, nil
}

func (it IntegrationsHandler) CreateIntegration(c *gin.Context) {
	if err := DenyForSandboxEnv(c); err != nil {
		return
	}

	var body models.CreateIntegrationSchema
	ok := utils.Validate(c, &body, false, nil)
	if !ok {
		return
	}
	var integration models.Integration
	integrationType := strings.ToLower(body.Name)
	switch integrationType {
	case "slack":
		var authToken, channelID, uiUrl string
		var pmAlert, svfAlert, disconnectAlert bool
		authToken, ok := body.Keys["auth_token"]
		if !ok {
			serv.Warnf("CreateIntegration: Must provide auth token for slack integration")
			c.AbortWithStatusJSON(configuration.SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": "Must provide auth token for slack integration"})
		}
		channelID, ok = body.Keys["channel_id"]
		if !ok {
			if !ok {
				serv.Warnf("CreateIntegration: Must provide channel ID for slack integration")
				c.AbortWithStatusJSON(configuration.SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": "Must provide channel ID for slack integration"})
			}
		}
		uiUrl = body.UIUrl
		if uiUrl == "" {
			serv.Warnf("CreateIntegration: Must provide channel ID for slack integration")
			c.AbortWithStatusJSON(500, gin.H{"message": "Must provide UI url for slack integration"})
		}

		pmAlert, ok = body.Properties[PoisonMAlert]
		if !ok {
			pmAlert = false
		}
		svfAlert, ok = body.Properties[SchemaVAlert]
		if !ok {
			svfAlert = false
		}
		disconnectAlert, ok = body.Properties[DisconEAlert]
		if !ok {
			disconnectAlert = false
		}

		keys, properties := CreateIntegrationsKeysAndProperties(integrationType, authToken, channelID, pmAlert, svfAlert, disconnectAlert, "", "", "", "")
		slackIntegration, err := CreateSlackIntegration(keys, properties, body.UIUrl)
		if err != nil {
			if strings.Contains(err.Error(), "Invalid auth token") || strings.Contains(err.Error(), "Invalid channel ID") || strings.Contains(err.Error(), "already exists") {
				serv.Warnf("CreateSlackIntegration: " + err.Error())
				c.AbortWithStatusJSON(configuration.SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": err.Error()})
				return
			} else {
				serv.Errorf("CreateSlackIntegration: " + err.Error())
				c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
				return
			}
		}
		integration = slackIntegration
		if integration.Keys["auth_token"] != "" {
			integration.Keys["auth_token"] = "xoxb-****"
		}
	case "s3":
		keys, err := it.CreateS3Integrtation(body.Keys)
		if err != nil {
			serv.Warnf("CreateIntegration: " + err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": err.Error()})
		}
		keys, properties := CreateIntegrationsKeysAndProperties(integrationType, "", "", false, false, false, keys["access_key"], keys["secret_key"], keys["bucket_name"], keys["region"])
		s3Integration, err := createS3Integration(keys, properties)

		if err != nil {
			serv.Warnf("CreateIntegration: " + err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}
		integration = s3Integration
		if integration.Keys["secret_key"] != "" {
			lastCharsSecretKey := integration.Keys["secret_key"][len(integration.Keys["secret_key"])-4:]
			integration.Keys["secret_key"] = "****" + lastCharsSecretKey
		}
	default:
		serv.Warnf("CreateIntegration: Unsupported integration type")
		c.AbortWithStatusJSON(400, gin.H{"message": "CreateIntegration error: Unsupported integration type"})
	}

	shouldSendAnalytics, _ := shouldSendAnalytics()
	if shouldSendAnalytics {
		user, _ := getUserDetailsFromMiddleware(c)
		analytics.SendEvent(user.Username, "user-create-integration-"+integrationType)
	}
	c.IndentedJSON(200, integration)
}

func (it IntegrationsHandler) UpdateIntegration(c *gin.Context) {
	if err := DenyForSandboxEnv(c); err != nil {
		return
	}

	var body models.CreateIntegrationSchema
	ok := utils.Validate(c, &body, false, nil)
	if !ok {
		return
	}
	var integration models.Integration
	switch strings.ToLower(body.Name) {
	case "slack":
		var authToken, channelID, uiUrl string
		var pmAlert, svfAlert, disconnectAlert bool
		authToken, ok := body.Keys["auth_token"]
		if !ok {
			serv.Warnf("CreateIntegration: Must provide auth token for slack integration")
			c.AbortWithStatusJSON(configuration.SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": "Must provide auth token for slack integration"})
		}
		channelID, ok = body.Keys["channel_id"]
		if !ok {
			if !ok {
				serv.Warnf("CreateIntegration: Must provide channel ID for slack integration")
				c.AbortWithStatusJSON(configuration.SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": "Must provide channel ID for slack integration"})
			}
		}
		uiUrl = body.UIUrl
		if uiUrl == "" {
			serv.Warnf("CreateIntegration: Must provide channel ID for slack integration")
			c.AbortWithStatusJSON(configuration.SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": "Must provide channel ID for slack integration"})
		}
		pmAlert, ok = body.Properties[PoisonMAlert]
		if !ok {
			pmAlert = false
		}
		svfAlert, ok = body.Properties[SchemaVAlert]
		if !ok {
			svfAlert = false
		}
		disconnectAlert, ok = body.Properties[DisconEAlert]
		if !ok {
			disconnectAlert = false
		}

		slackIntegration, err := updateSlackIntegration(authToken, channelID, pmAlert, svfAlert, disconnectAlert, body.UIUrl)
		if err != nil {
			if strings.Contains(err.Error(), "Invalid auth token") || strings.Contains(err.Error(), "Invalid channel ID") {
				serv.Warnf("UpdateSlackIntegration: " + err.Error())
				c.AbortWithStatusJSON(configuration.SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": err.Error()})
				return
			} else {
				serv.Errorf("UpdateSlackIntegration: " + err.Error())
				c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
				return
			}
		}
		integration = slackIntegration
		if integration.Keys["auth_token"] != "" {
			integration.Keys["auth_token"] = "xoxb-****"
		}
	case "s3":
		integrationType := strings.ToLower(body.Name)
		keys, properties := CreateIntegrationsKeysAndProperties(integrationType, "", "", false, false, false, body.Keys["access_key"], body.Keys["secret_key"], body.Keys["bucket_name"], body.Keys["region"])
		s3Integration, err := updateS3Integration(keys, properties)
		if err != nil {
			serv.Errorf("updateS3Integration: " + err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}
		integration = s3Integration
		if integration.Keys["secret_key"] != "" {
			lastCharsSecretKey := integration.Keys["secret_key"][len(integration.Keys["secret_key"])-4:]
			integration.Keys["secret_key"] = "****" + lastCharsSecretKey
		}

	default:
		serv.Warnf("CreateIntegration: Unsupported integration type - " + body.Name)
		c.AbortWithStatusJSON(400, gin.H{"message": "CreateIntegration: Unsupported integration type - " + body.Name})
	}

	c.IndentedJSON(200, integration)
}

func createS3Integration(keys map[string]string, properties map[string]bool) (models.Integration, error) {
	var s3Integration models.Integration
	filter := bson.M{"name": "s3"}
	err := integrationsCollection.FindOne(context.TODO(),
		filter).Decode(&s3Integration)
	if err == mongo.ErrNoDocuments {
		s3Integration = models.Integration{
			ID:         primitive.NewObjectID(),
			Name:       "s3",
			Keys:       keys,
			Properties: properties,
		}
		_, insertErr := integrationsCollection.InsertOne(context.TODO(), s3Integration)
		if insertErr != nil {
			return s3Integration, insertErr
		}

		integrationToUpdate := models.CreateIntegrationSchema{
			Name:       "s3",
			Keys:       keys,
			Properties: properties,
		}
		msg, err := json.Marshal(integrationToUpdate)
		if err != nil {
			return s3Integration, err
		}
		err = serv.sendInternalAccountMsgWithReply(serv.GlobalAccount(), INTEGRATIONS_UPDATES_SUBJ, _EMPTY_, nil, msg, true)
		if err != nil {
			return s3Integration, err
		}
		return s3Integration, nil

	} else if err != nil {
		return s3Integration, err
	}
	return s3Integration, errors.New("S3 integration already exists")

}

func CreateSlackIntegration(keys map[string]string, properties map[string]bool, uiUrl string) (models.Integration, error) {
	var slackIntegration models.Integration
	filter := bson.M{"name": "slack"}
	err := integrationsCollection.FindOne(context.TODO(),
		filter).Decode(&slackIntegration)
	if err == mongo.ErrNoDocuments {
		err := testSlackIntegration(keys["auth_token"], keys["channel_id"], "Slack integration with Memphis was added successfully")
		if err != nil {
			return slackIntegration, err
		}
		slackIntegration = models.Integration{
			ID:         primitive.NewObjectID(),
			Name:       "slack",
			Keys:       keys,
			Properties: properties,
		}
		_, insertErr := integrationsCollection.InsertOne(context.TODO(), slackIntegration)
		if insertErr != nil {
			return slackIntegration, insertErr
		}

		integrationToUpdate := models.CreateIntegrationSchema{
			Name:       "slack",
			Keys:       keys,
			Properties: properties,
			UIUrl:      uiUrl,
		}
		msg, err := json.Marshal(integrationToUpdate)
		if err != nil {
			return slackIntegration, err
		}
		err = serv.sendInternalAccountMsgWithReply(serv.GlobalAccount(), INTEGRATIONS_UPDATES_SUBJ, _EMPTY_, nil, msg, true)
		if err != nil {
			return slackIntegration, err
		}
		update := models.ConfigurationsUpdate{
			Type:   sendNotificationType,
			Update: properties[SchemaVAlert],
		}
		serv.SendUpdateToClients(update)

		return slackIntegration, nil
	} else if err != nil {
		return slackIntegration, err
	}
	return slackIntegration, errors.New("Slack integration already exists")
}

func updateS3Integration(keys map[string]string, properties map[string]bool) (models.Integration, error) {
	var s3Integration models.Integration
	filter := bson.M{"name": "s3"}
	err := integrationsCollection.FindOneAndUpdate(context.TODO(),
		filter,
		bson.M{"$set": bson.M{"keys": keys, "properties": properties}}).Decode(&s3Integration)
	if err == mongo.ErrNoDocuments {
		s3Integration = models.Integration{
			ID:         primitive.NewObjectID(),
			Name:       "s3",
			Keys:       keys,
			Properties: properties,
		}
		integrationsCollection.InsertOne(context.TODO(), s3Integration)
	} else if err != nil {
		return s3Integration, err
	}

	integrationToUpdate := models.CreateIntegrationSchema{
		Name:       "s3",
		Keys:       keys,
		Properties: properties,
	}

	msg, err := json.Marshal(integrationToUpdate)
	if err != nil {
		return s3Integration, err
	}
	err = serv.sendInternalAccountMsgWithReply(serv.GlobalAccount(), INTEGRATIONS_UPDATES_SUBJ, _EMPTY_, nil, msg, true)
	if err != nil {
		return s3Integration, err
	}

	s3Integration.Keys = keys
	s3Integration.Properties = properties
	return s3Integration, nil
}

func updateSlackIntegration(authToken string, channelID string, pmAlert bool, svfAlert bool, disconnectAlert bool, uiUrl string) (models.Integration, error) {
	var slackIntegration models.Integration
	if authToken == "" {
		var integrationFromDb models.Integration
		filter := bson.M{"name": "slack"}
		err := integrationsCollection.FindOne(context.TODO(), filter).Decode(&integrationFromDb)
		if err != nil {
			return slackIntegration, err
		}
		authToken = integrationFromDb.Keys["auth_token"]
	}

	err := testSlackIntegration(authToken, channelID, "Slack integration with Memphis was updated successfully")
	if err != nil {
		return slackIntegration, err
	}
	keys, properties := CreateIntegrationsKeysAndProperties("slack", authToken, channelID, pmAlert, svfAlert, disconnectAlert, "", "", "", "")
	filter := bson.M{"name": "slack"}
	err = integrationsCollection.FindOneAndUpdate(context.TODO(),
		filter,
		bson.M{"$set": bson.M{"keys": keys, "properties": properties}}).Decode(&slackIntegration)
	if err == mongo.ErrNoDocuments {
		slackIntegration = models.Integration{
			ID:         primitive.NewObjectID(),
			Name:       "slack",
			Keys:       keys,
			Properties: properties,
		}
		integrationsCollection.InsertOne(context.TODO(), slackIntegration)
	} else if err != nil {
		return slackIntegration, err
	}

	integrationToUpdate := models.CreateIntegrationSchema{
		Name:       "slack",
		Keys:       keys,
		Properties: properties,
		UIUrl:      uiUrl,
	}

	msg, err := json.Marshal(integrationToUpdate)
	if err != nil {
		return slackIntegration, err
	}
	err = serv.sendInternalAccountMsgWithReply(serv.GlobalAccount(), INTEGRATIONS_UPDATES_SUBJ, _EMPTY_, nil, msg, true)
	if err != nil {
		return slackIntegration, err
	}

	update := models.ConfigurationsUpdate{
		Type:   sendNotificationType,
		Update: svfAlert,
	}
	serv.SendUpdateToClients(update)

	slackIntegration.Keys = keys
	slackIntegration.Properties = properties
	return slackIntegration, nil
}

func testSlackIntegration(authToken string, channelID string, message string) error {
	slackClientTemp := slack.New(authToken)
	_, err := slackClientTemp.AuthTest()
	if err != nil {
		return errors.New("Invalid auth token")
	}
	attachment := slack.Attachment{
		AuthorName: "Memphis",
		Text:       message,
		Color:      "#6557FF",
	}

	_, _, err = slackClientTemp.PostMessage(
		channelID,
		slack.MsgOptionAttachments(attachment),
	)
	if err != nil {
		return errors.New("Invalid channel ID")
	}
	return nil
}

func CreateIntegrationsKeysAndProperties(integrationType, authToken string, channelID string, pmAlert bool, svfAlert bool, disconnectAlert bool, accessKey, secretKey, bucketName, region string) (map[string]string, map[string]bool) {
	keys := make(map[string]string)
	properties := make(map[string]bool)
	switch integrationType {
	case "slack":
		keys["auth_token"] = authToken
		keys["channel_id"] = channelID
		properties[PoisonMAlert] = pmAlert
		properties[SchemaVAlert] = svfAlert
		properties[DisconEAlert] = disconnectAlert
	case "s3":
		keys["access_key"] = accessKey
		keys["secret_key"] = secretKey
		keys["bucket_name"] = bucketName
		keys["region"] = region
	}

	return keys, properties
}

func (it IntegrationsHandler) GetIntegrationDetails(c *gin.Context) {
	var body models.GetIntegrationDetailsSchema
	ok := utils.Validate(c, &body, false, nil)
	if !ok {
		return
	}
	filter := bson.M{"name": strings.ToLower(body.Name)}
	var integration models.Integration
	err := integrationsCollection.FindOne(context.TODO(),
		filter).Decode(&integration)
	if err == mongo.ErrNoDocuments {
		c.IndentedJSON(200, nil)
		return
	} else if err != nil {
		serv.Errorf("GetIntegrationDetails: Integration " + body.Name + ": " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	if integration.Name == "slack" && integration.Keys["auth_token"] != "" {
		integration.Keys["auth_token"] = "xoxb-****"
	}
	c.IndentedJSON(200, integration)
}

func (it IntegrationsHandler) GetAllIntegrations(c *gin.Context) {
	var integrations []models.Integration
	cursor, err := integrationsCollection.Find(context.TODO(), bson.M{})
	if err == mongo.ErrNoDocuments {
	} else if err != nil {
		serv.Errorf("GetAllIntegrations: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	if err = cursor.All(context.TODO(), &integrations); err != nil {
		serv.Errorf("GetAllIntegrations: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	for i := 0; i < len(integrations); i++ {
		if integrations[i].Name == "slack" && integrations[i].Keys["auth_token"] != "" {
			integrations[i].Keys["auth_token"] = "xoxb-****"
		}
	}

	shouldSendAnalytics, _ := shouldSendAnalytics()
	if shouldSendAnalytics {
		user, _ := getUserDetailsFromMiddleware(c)
		analytics.SendEvent(user.Username, "user-enter-integration-page")
	}

	c.IndentedJSON(200, integrations)
}

func (it IntegrationsHandler) DisconnectIntegration(c *gin.Context) {
	if err := DenyForSandboxEnv(c); err != nil {
		return
	}

	var body models.DisconnectIntegrationSchema
	ok := utils.Validate(c, &body, false, nil)
	if !ok {
		return
	}

	integrationType := strings.ToLower(body.Name)
	filter := bson.M{"name": integrationType}
	_, err := integrationsCollection.DeleteOne(context.TODO(), filter)
	if err != nil {
		serv.Errorf("DisconnectIntegration: Integration " + body.Name + ": " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	integrationUpdate := models.Integration{
		Name:       strings.ToLower(body.Name),
		Keys:       nil,
		Properties: nil,
	}

	msg, err := json.Marshal(integrationUpdate)
	if err != nil {
		serv.Errorf("DisconnectIntegration: Integration " + body.Name + ": " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	err = serv.sendInternalAccountMsgWithReply(serv.GlobalAccount(), INTEGRATIONS_UPDATES_SUBJ, _EMPTY_, nil, msg, true)
	if err != nil {
		serv.Errorf("DisconnectIntegration: Integration " + body.Name + ": " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	switch body.Name {
	case "slack":
		update := models.ConfigurationsUpdate{
			Type:   sendNotificationType,
			Update: false,
		}
		serv.SendUpdateToClients(update)
	}

	shouldSendAnalytics, _ := shouldSendAnalytics()
	if shouldSendAnalytics {
		user, _ := getUserDetailsFromMiddleware(c)
		analytics.SendEvent(user.Username, "user-disconnect-integration-"+integrationType)
	}
	c.IndentedJSON(200, gin.H{})
}

func (it IntegrationsHandler) RequestIntegration(c *gin.Context) {
	var body models.RequestIntegrationSchema
	ok := utils.Validate(c, &body, false, nil)
	if !ok {
		return
	}

	shouldSendAnalytics, _ := shouldSendAnalytics()
	if shouldSendAnalytics {
		user, _ := getUserDetailsFromMiddleware(c)
		param := analytics.EventParam{
			Name:  "request-content",
			Value: body.RequestContent,
		}
		analyticsParams := []analytics.EventParam{param}
		analytics.SendEventWithParams(user.Username, analyticsParams, "user-request-integration")
	}

	c.IndentedJSON(200, gin.H{})
}
