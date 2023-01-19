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
package storage

import (
	"context"
	"memphis-broker/integrations/notifications"
	"memphis-broker/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func InitializeS3Connection(c *mongo.Client) error {
	filter := bson.M{"name": "s3"}
	var s3Integration models.Integration
	err := notifications.IntegrationsCollection.FindOne(context.TODO(),
		filter).Decode(&s3Integration)
	if err == mongo.ErrNoDocuments {
		return nil
	} else if err != nil {
		return err
	}
	CacheS3Details(s3Integration.Keys, s3Integration.Properties)
	return nil
}

func clearS3Cache() {
	delete(notifications.NotificationIntegrationsMap, "s3")
}

func CacheS3Details(keys map[string]string, properties map[string]bool) {
	awsIntegration, ok := notifications.NotificationIntegrationsMap["s3"].(models.AwsIntegration)
	if !ok {
		awsIntegration = models.AwsIntegration{}
		awsIntegration.Keys = make(map[string]string)
		awsIntegration.Properties = make(map[string]bool)
	}
	if keys == nil {
		clearS3Cache()
		return
	}

	awsIntegration.Keys["access_key"] = keys["access_key"]
	awsIntegration.Keys["secret_key"] = keys["secret_key"]
	awsIntegration.Keys["bucket_name"] = keys["bucket_name"]
	awsIntegration.Keys["region"] = keys["region"]
	awsIntegration.Name = "s3"
	notifications.NotificationIntegrationsMap["s3"] = awsIntegration
}
