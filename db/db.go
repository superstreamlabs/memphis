// Copyright 2021-2022 The Memphis Authors
// Licensed under the Apache License, Version 2.0 (the “License”);
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an “AS IS” BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package db

import (
	"memphis-broker/conf"
	"memphis-broker/server"

	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var configuration = conf.GetConfig()
var logger = log.Default()

const (
	dbOperationTimeout = 20
)

func InitializeDbConnection(s *server.Server) error {
	ctx, cancel := context.WithTimeout(context.TODO(), dbOperationTimeout*time.Second)

	var clientOptions *options.ClientOptions
	if configuration.DOCKER_ENV != "" {
		clientOptions = options.Client().ApplyURI(configuration.MONGO_URL).SetConnectTimeout(dbOperationTimeout * time.Second)
	} else {
		auth := options.Credential{
			AuthSource: configuration.DB_NAME,
			Username:   configuration.MONGO_USER,
			Password:   configuration.MONGO_PASS,
		}
		clientOptions = options.Client().ApplyURI(configuration.MONGO_URL).SetAuth(auth).SetConnectTimeout(dbOperationTimeout * time.Second)
	}

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return err
		// panic("Failed to create Mongodb client: " + err.Error())
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		return err
		// panic("Failed to create Mongo DB client: " + err.Error())
	}

	s.DbClient = client
	s.DbCtx = ctx
	s.DbCancel = cancel
	s.Noticef("[INFO] Established connection with the DB")
	return nil
}

func GetCollection(collectionName string, s *server.Server) *mongo.Collection {
	var collection *mongo.Collection = s.DbClient.Database(configuration.DB_NAME).Collection(collectionName)
	return collection
}

func Close(s *server.Server) {
	defer s.DbCancel()
	defer func() {
		if err := s.DbClient.Disconnect(s.DbCtx); err != nil {
			s.Errorf("Failed to close Mongodb client: " + err.Error())
			panic("Failed to close Mongodb client: " + err.Error())
		}
	}()
}
