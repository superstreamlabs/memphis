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
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"

	"memphis/db"
	"memphis/models"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"
)

type TieredStorageMsg struct {
	Buf         []byte `json:"buf"`
	StationName string `json:"station_name"`
}

func cacheDetailsS3(keys map[string]string, properties map[string]bool, tenantName string) {
	s3Integration := models.Integration{}
	s3Integration.Keys = make(map[string]string)
	s3Integration.Properties = make(map[string]bool)
	if keys == nil {
		deleteIntegrationFromTenant(tenantName, "s3", IntegrationsConcurrentCache)
		return
	}

	s3Integration.Keys["access_key"] = keys["access_key"]
	s3Integration.Keys["secret_key"] = keys["secret_key"]
	s3Integration.Keys["bucket_name"] = keys["bucket_name"]
	s3Integration.Keys["region"] = keys["region"]
	s3Integration.Name = "s3"
	if _, ok := IntegrationsConcurrentCache.Load(tenantName); !ok {
		IntegrationsConcurrentCache.Add(tenantName, map[string]interface{}{"s3": s3Integration})
	} else {
		err := addIntegrationToTenant(tenantName, "s3", IntegrationsConcurrentCache, s3Integration)
		if err != nil {
			serv.Errorf("cacheDetailsSlack: " + err.Error())
			return
		}
	}
}

func (it IntegrationsHandler) handleCreateS3Integration(tenantName string, keys map[string]string) (models.Integration, int, error) {
	statusCode, _, err := it.handleS3Integrtation(tenantName, keys)
	if err != nil {
		return models.Integration{}, statusCode, err
	}

	keys, properties := createIntegrationsKeysAndProperties("s3", "", "", false, false, false, keys["access_key"], keys["secret_key"], keys["bucket_name"], keys["region"])
	s3Integration, err := createS3Integration(tenantName, keys, properties)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			return models.Integration{}, SHOWABLE_ERROR_STATUS_CODE, err
		} else {
			return models.Integration{}, 500, err
		}
	}
	return s3Integration, statusCode, nil
}

func (it IntegrationsHandler) handleUpdateS3Integration(body models.CreateIntegrationSchema) (models.Integration, int, error) {
	statusCode, keys, err := it.handleS3Integrtation(body.TenantName, body.Keys)
	if err != nil {
		return models.Integration{}, statusCode, err
	}
	integrationType := strings.ToLower(body.Name)
	keys, properties := createIntegrationsKeysAndProperties(integrationType, "", "", false, false, false, keys["access_key"], keys["secret_key"], keys["bucket_name"], keys["region"])
	s3Integration, err := updateS3Integration(body.TenantName, keys, properties)
	if err != nil {
		return s3Integration, 500, err
	}
	return s3Integration, statusCode, nil
}

func (it IntegrationsHandler) handleS3Integrtation(tenantName string, keys map[string]string) (int, map[string]string, error) {
	accessKey := keys["access_key"]
	secretKey := keys["secret_key"]
	region := keys["region"]
	bucketName := keys["bucket_name"]

	if keys["secret_key"] == "" {
		exist, integrationFromDb, err := db.GetIntegration("s3", tenantName)
		if err != nil {
			return 500, map[string]string{}, err
		}
		if !exist {
			return SHOWABLE_ERROR_STATUS_CODE, map[string]string{}, errors.New("secret key is invalid")
		}
		if value, ok := integrationFromDb.Keys["secret_key"]; ok {
			decryptedValue, err := DecryptAES(value)
			if err != nil {
				return 500, map[string]string{}, err
			}
			integrationFromDb.Keys["secret_key"] = decryptedValue
		}
		secretKey = integrationFromDb.Keys["secret_key"]
		keys["secret_key"] = secretKey
	}

	provider := credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")
	_, err := provider.Retrieve(context.Background())
	if err != nil {
		if strings.Contains(err.Error(), "static credentials are empty") {
			return SHOWABLE_ERROR_STATUS_CODE, map[string]string{}, errors.New("credentials are empty")
		} else {
			return 500, map[string]string{}, err
		}
	}

	cfg, err := awsconfig.LoadDefaultConfig(context.Background(),
		awsconfig.WithCredentialsProvider(provider),
		awsconfig.WithRegion(region),
	)
	if err != nil {
		return 500, map[string]string{}, err
	}

	svc := s3.NewFromConfig(cfg)
	if err != nil {
		err = errors.New("NewSession failure " + err.Error())
		return 500, map[string]string{}, err
	}

	statusCode, err := testS3Integration(svc, bucketName)
	if err != nil {
		return statusCode, map[string]string{}, err
	}
	return statusCode, keys, nil
}

func createS3Integration(tenantName string, keys map[string]string, properties map[string]bool) (models.Integration, error) {
	exist, s3Integration, err := db.GetIntegration("s3", tenantName)
	if err != nil {
		return models.Integration{}, err
	} else if !exist {
		cloneKeys := copyMaps(keys)
		encryptedValue, err := EncryptAES([]byte(keys["secret_key"]))
		if err != nil {
			return models.Integration{}, err
		}
		cloneKeys["secret_key"] = encryptedValue
		integrationRes, insertErr := db.InsertNewIntegration(tenantName, "s3", cloneKeys, properties)
		if insertErr != nil {
			return models.Integration{}, insertErr
		}
		s3Integration = integrationRes
		integrationToUpdate := models.CreateIntegrationSchema{
			Name:       "s3",
			Keys:       keys,
			Properties: properties,
			TenantName: tenantName,
		}
		msg, err := json.Marshal(integrationToUpdate)
		if err != nil {
			return models.Integration{}, err
		}
		err = serv.sendInternalAccountMsgWithReply(serv.GlobalAccount(), INTEGRATIONS_UPDATES_SUBJ, _EMPTY_, nil, msg, true)
		if err != nil {
			return models.Integration{}, err
		}
		s3Integration.Keys["secret_key"] = hideS3SecretKey(keys["secret_key"])
		return s3Integration, nil
	}
	return models.Integration{}, errors.New("s3 integration already exists")

}

func updateS3Integration(tenantName string, keys map[string]string, properties map[string]bool) (models.Integration, error) {
	cloneKeys := copyMaps(keys)
	encryptedValue, err := EncryptAES([]byte(keys["secret_key"]))
	if err != nil {
		return models.Integration{}, err
	}
	cloneKeys["secret_key"] = encryptedValue
	s3Integration, err := db.UpdateIntegration(tenantName, "s3", cloneKeys, properties)
	if err != nil {
		return models.Integration{}, err
	}

	integrationToUpdate := models.CreateIntegrationSchema{
		Name:       "s3",
		Keys:       keys,
		Properties: properties,
		TenantName: tenantName,
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

func testS3Integration(svc *s3.Client, bucketName string) (int, error) {
	_, err := svc.HeadBucket(context.Background(), &s3.HeadBucketInput{
		Bucket: aws.String(bucketName),
	})
	var statusCode int
	if err != nil {
		if strings.Contains(err.Error(), "Forbidden") {
			err = errors.New("invalid access key or secret key")
			statusCode = SHOWABLE_ERROR_STATUS_CODE
		} else if strings.Contains(err.Error(), "404") {
			err = errors.New("bucket does not exist")
			statusCode = SHOWABLE_ERROR_STATUS_CODE
		} else if strings.Contains(err.Error(), "Moved Permanently") {
			err = errors.New("bucket name does not exist in the selected region, it is probably placed in another region")
			statusCode = SHOWABLE_ERROR_STATUS_CODE
		} else if strings.Contains(err.Error(), "send request failed") {
			err = errors.New("upload failed")
			statusCode = SHOWABLE_ERROR_STATUS_CODE
		} else if strings.Contains(err.Error(), "could not find region configuration") {
			var oe *smithy.OperationError
			if errors.As(err, &oe) {
				err = errors.New(oe.Error() + " : region name is empty")
			}
			statusCode = SHOWABLE_ERROR_STATUS_CODE
		} else if strings.Contains(err.Error(), "validation error(s) found") || strings.Contains(err.Error(), "BadRequest: Bad Request") {
			err = errors.New("invalid bucket name")
			statusCode = SHOWABLE_ERROR_STATUS_CODE
		} else if strings.Contains(err.Error(), "incorrect region") {
			var oe *smithy.OperationError
			if errors.As(err, &oe) {
				err = errors.New(oe.Error() + " : incorrect region")
			}
			statusCode = SHOWABLE_ERROR_STATUS_CODE
		} else {
			statusCode = 500
		}
		return statusCode, err
	}

	acl, err := svc.GetBucketAcl(context.Background(), &s3.GetBucketAclInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		err = errors.New("getBucketAcl error: " + err.Error())
		return 500, err
	}

	permission := acl.Grants[0].Permission

	if permission != types.PermissionFullControl {
		err = errors.New("creds should have full access on this bucket")
		return SHOWABLE_ERROR_STATUS_CODE, err
	}

	uploader := manager.NewUploader(svc)
	reader := strings.NewReader(string("test"))
	// Upload the object to S3.
	_, err = uploader.Upload(context.Background(), &s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String("memphis"),
		Body:   reader,
	})
	if err != nil {
		err = errors.New("could not upload objects - " + err.Error())
		return SHOWABLE_ERROR_STATUS_CODE, err
	}
	_, err = svc.DeleteObject(context.Background(), &s3.DeleteObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String("memphis"),
	})
	if err != nil {
		err = errors.New("could not delete objects - " + err.Error())
		return SHOWABLE_ERROR_STATUS_CODE, err
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

type Msg struct {
	Payload string            `json:"payload"`
	Headers map[string]string `json:"headers"`
}

func (s *Server) uploadToS3Storage() error {
	//TODO: remove GetAllTenants and iterate the concurrent map
	tenants, _ := db.GetAllTenants()
	for _, tenant := range tenants {
		if len(tieredStorageMsgsMap.m) > 0 {
			var credentialsMap models.Integration
			if tenantIntegrations, ok := IntegrationsConcurrentCache.Load(tenant.Name); !ok {
				continue
			} else {
				if credentialsMap, ok = tenantIntegrations["s3"].(models.Integration); !ok {
					continue
				}
			}
			provider := credentials.NewStaticCredentialsProvider(
				credentialsMap.Keys["access_key"],
				credentialsMap.Keys["secret_key"],
				"",
			)
			_, err := provider.Retrieve(context.Background())
			if err != nil {
				err = errors.New("uploadToS3Storage: Invalid credentials")
				return err
			}
			cfg, err := awsconfig.LoadDefaultConfig(context.Background(),
				awsconfig.WithCredentialsProvider(provider),
				awsconfig.WithRegion(credentialsMap.Keys["region"]),
			)
			svc := s3.NewFromConfig(cfg)
			if err != nil {
				err = errors.New("uploadToS3Storage failure " + err.Error())
				return err
			}
			uploader := manager.NewUploader(svc)
			uid := serv.memphis.nuid.Next()
			var objectName string

			for k, msgs := range tieredStorageMsgsMap.m {
				var messages []Msg
				for _, msg := range msgs {
					objectName = k + "/" + uid + "(" + strconv.Itoa(len(msgs)) + ").json"

					var headers string
					hdrs := map[string]string{}
					if len(msg.Header) > 0 {
						headers = strings.ToLower(string(msg.Header))
						headersSplit := strings.Split(headers, CR_LF)
						for _, header := range headersSplit {
							if header != "" && !strings.Contains(header, "nats") {
								keyVal := strings.Split(header, ":")
								key := strings.TrimSpace(keyVal[0])
								value := strings.TrimSpace(keyVal[1])
								hdrs[key] = value
							}
						}
					} else {
						headers = ""
					}

					encodedMsg := hex.EncodeToString(msg.Data)
					message := Msg{Payload: encodedMsg, Headers: hdrs}
					messages = append(messages, message)
				}
				// Upload the object to S3.
				var buf bytes.Buffer
				err := json.NewEncoder(&buf).Encode(messages)
				if err != nil {
					return err
				}
				_, err = uploader.Upload(context.Background(), &s3.PutObjectInput{
					Bucket: aws.String(credentialsMap.Keys["bucket_name"]),
					Key:    aws.String(objectName),
					Body:   &buf,
				})
				if err != nil {
					err = errors.New("uploadToS3Storage: failed to upload object to S3: " + err.Error())
					return err
				}
				serv.Noticef("new file has been uploaded to S3: %s", objectName)
			}
		}
	}

	return nil

}
