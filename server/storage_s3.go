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
	"memphis-broker/models"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func cacheDetailsS3(keys map[string]string, properties map[string]bool) {
	s3Integration, ok := IntegrationsCache["s3"].(models.Integration)
	if !ok {
		s3Integration = models.Integration{}
		s3Integration.Keys = make(map[string]string)
		s3Integration.Properties = make(map[string]bool)
	}
	if keys == nil {
		clearCache("s3")
		return
	}

	s3Integration.Keys["access_key"] = keys["access_key"]
	s3Integration.Keys["secret_key"] = keys["secret_key"]
	s3Integration.Keys["bucket_name"] = keys["bucket_name"]
	s3Integration.Keys["region"] = keys["region"]
	s3Integration.Name = "s3"
	IntegrationsCache["s3"] = s3Integration
}

func (it IntegrationsHandler) handleCreateS3Integration(keys map[string]string, integrationType string) (models.Integration, int, error) {
	statusCode, _, err := it.handleS3Integrtation(keys)
	if err != nil {
		return models.Integration{}, statusCode, err
	}

	keys, properties := createIntegrationsKeysAndProperties(integrationType, "", "", false, false, false, keys["access_key"], keys["secret_key"], keys["bucket_name"], keys["region"])
	s3Integration, err := createS3Integration(keys, properties)
	if err != nil {
		return models.Integration{}, 500, err
	}
	return s3Integration, statusCode, nil
}

func (it IntegrationsHandler) handleUpdateS3Integration(body models.CreateIntegrationSchema) (models.Integration, int, error) {
	statusCode, keys, err := it.handleS3Integrtation(body.Keys)
	if err != nil {
		return models.Integration{}, statusCode, err
	}
	integrationType := strings.ToLower(body.Name)
	keys, properties := createIntegrationsKeysAndProperties(integrationType, "", "", false, false, false, keys["access_key"], keys["secret_key"], keys["bucket_name"], keys["region"])
	s3Integration, err := updateS3Integration(keys, properties)
	if err != nil {
		return s3Integration, 500, err
	}
	return s3Integration, statusCode, nil
}

func (it IntegrationsHandler) handleS3Integrtation(keys map[string]string) (int, map[string]string, error) {
	accessKey := keys["access_key"]
	secretKey := keys["secret_key"]
	region := keys["region"]
	bucketName := keys["bucket_name"]

	if keys["secret_key"] == "" {
		var integrationFromDb models.Integration
		filter := bson.M{"name": "s3"}
		err := integrationsCollection.FindOne(context.TODO(), filter).Decode(&integrationFromDb)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return configuration.SHOWABLE_ERROR_STATUS_CODE, map[string]string{}, errors.New("secret key is invalid")
			}
			return 500, map[string]string{}, err
		}
		secretKey = integrationFromDb.Keys["secret_key"]
		keys["secret_key"] = secretKey
	}
	provider := &credentials.StaticProvider{Value: credentials.Value{
		AccessKeyID:     accessKey,
		SecretAccessKey: secretKey,
	}}

	_, err := provider.Retrieve()
	if err != nil {
		if strings.Contains(err.Error(), "static credentials are empty") {
			return configuration.SHOWABLE_ERROR_STATUS_CODE, map[string]string{}, errors.New("credentials are empty")
		} else {
			return 500, map[string]string{}, err
		}
	}

	credentials := credentials.NewCredentials(provider)
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(region),
		Credentials: credentials},
	)
	if err != nil {
		err = errors.New("NewSession failure " + err.Error())
		return 500, map[string]string{}, err
	}

	svc := s3.New(sess)
	statusCode, err := testS3Integration(sess, svc, bucketName)
	if err != nil {
		return statusCode, map[string]string{}, err
	}
	return statusCode, keys, nil
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
		s3Integration.Keys["secret_key"] = hideS3SecretKey(keys["secret_key"])
		return s3Integration, nil

	} else if err != nil {
		return s3Integration, err
	}
	return s3Integration, errors.New("S3 integration already exists")

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

	keys["secret_key"] = hideS3SecretKey(keys["secret_key"])
	s3Integration.Keys = keys
	s3Integration.Properties = properties
	return s3Integration, nil
}

func testS3Integration(sess *session.Session, svc *s3.S3, bucketName string) (int, error) {
	_, err := svc.HeadBucket(&s3.HeadBucketInput{
		Bucket: aws.String(bucketName),
	})
	var statusCode int
	if err != nil {
		if strings.Contains(err.Error(), "Forbidden") {
			err = errors.New("Invalid access key or secret key")
			statusCode = configuration.SHOWABLE_ERROR_STATUS_CODE
		} else if strings.Contains(err.Error(), "NotFound: Not Found") {
			err = errors.New("Bucket name is not exists")
			statusCode = configuration.SHOWABLE_ERROR_STATUS_CODE
		} else if strings.Contains(err.Error(), "send request failed") {
			err = errors.New("Invalid region")
			statusCode = configuration.SHOWABLE_ERROR_STATUS_CODE
		} else if strings.Contains(err.Error(), "could not find region configuration") {
			err = errors.New("Invalid region: region is empty")
			statusCode = configuration.SHOWABLE_ERROR_STATUS_CODE
		} else if strings.Contains(err.Error(), "validation error(s) found") || strings.Contains(err.Error(), "BadRequest: Bad Request") {
			err = errors.New("Invalid bucket name")
			statusCode = configuration.SHOWABLE_ERROR_STATUS_CODE
		} else {
			statusCode = 500
		}
		err = errors.New("create a S3 client with additional configuration failure: " + err.Error())
		return statusCode, err
	}

	acl, err := svc.GetBucketAcl(&s3.GetBucketAclInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		err = errors.New("GetBucketAcl error: " + err.Error())
		return 500, err
	}

	permission := *acl.Grants[0].Permission
	permissionValue := permission

	if permissionValue != "FULL_CONTROL" {
		err = errors.New("you should full control permission: read, write and delete " + err.Error())
		return 500, err
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
		return 500, err
	}
	//delete the object
	_, err = svc.DeleteObject(&s3.DeleteObjectInput{Bucket: aws.String(bucketName), Key: aws.String(configuration.SERVER_NAME)})
	if err != nil {
		err = errors.New("Unable to delete object from bucket " + bucketName + err.Error())
		return 500, err
	}
	err = svc.WaitUntilObjectNotExists(&s3.HeadObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(configuration.SERVER_NAME),
	})
	if err != nil {
		err = errors.New("Error occurred while waiting for object to be deleted from bucket " + bucketName + err.Error())
		return 500, err
	}
	return 0, nil
}

func hideS3SecretKey(secretKey string) string {
	if secretKey != "" {
		lastCharsSecretKey := secretKey[len(secretKey)-4:]
		secretKey = "****" + lastCharsSecretKey
		return secretKey
	}
	return secretKey

}
