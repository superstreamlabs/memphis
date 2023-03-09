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
package db

import (
	"memphis/conf"
	"memphis/models"

	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var configuration = conf.GetConfig()
var usersCollection *mongo.Collection
var imagesCollection *mongo.Collection
var stationsCollection *mongo.Collection
var connectionsCollection *mongo.Collection
var producersCollection *mongo.Collection
var consumersCollection *mongo.Collection
var systemKeysCollection *mongo.Collection
var auditLogsCollection *mongo.Collection
var tagsCollection *mongo.Collection
var schemasCollection *mongo.Collection
var schemaVersionCollection *mongo.Collection
var sandboxUsersCollection *mongo.Collection
var integrationsCollection *mongo.Collection
var configurationsCollection *mongo.Collection

const (
	dbOperationTimeout = 20
)

type logger interface {
	Noticef(string, ...interface{})
	Errorf(string, ...interface{})
}

type DbInstance struct {
	Client *mongo.Client
	Ctx    context.Context
	Cancel context.CancelFunc
}

func InitializeDbConnection(l logger) (DbInstance, error) {
	ctx, cancel := context.WithTimeout(context.TODO(), dbOperationTimeout*time.Second)

	var clientOptions *options.ClientOptions
	if configuration.DOCKER_ENV != "" || configuration.LOCAL_CLUSTER_ENV {
		clientOptions = options.Client().ApplyURI(configuration.MONGO_URL).SetConnectTimeout(dbOperationTimeout * time.Second)
	} else {
		auth := options.Credential{
			Username: configuration.MONGO_USER,
			Password: configuration.MONGO_PASS,
		}
		if !configuration.EXTERNAL_MONGO {
			auth.AuthSource = configuration.DB_NAME
		}

		clientOptions = options.Client().ApplyURI(configuration.MONGO_URL).SetAuth(auth).SetConnectTimeout(dbOperationTimeout * time.Second)
	}

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		cancel()
		return DbInstance{}, err
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		cancel()
		return DbInstance{}, err
	}
	usersCollection = GetCollection("users", client)
	imagesCollection = GetCollection("images", client)
	stationsCollection = GetCollection("stations", client)
	connectionsCollection = GetCollection("connections", client)
	producersCollection = GetCollection("producers", client)
	consumersCollection = GetCollection("consumers", client)
	systemKeysCollection = GetCollection("system_keys", client)
	auditLogsCollection = GetCollection("audit_logs", client)
	tagsCollection = GetCollection("tags", client)
	schemasCollection = GetCollection("schemas", client)
	schemaVersionCollection = GetCollection("schema_versions", client)
	sandboxUsersCollection = GetCollection("sandbox_users", client)
	integrationsCollection = GetCollection("integrations", client)
	configurationsCollection = GetCollection("configurations", client)

	l.Noticef("Established connection with the DB")
	return DbInstance{Client: client, Ctx: ctx, Cancel: cancel}, nil
}

func GetCollection(collectionName string, dbClient *mongo.Client) *mongo.Collection {
	dbName := configuration.DB_NAME
	if configuration.EXTERNAL_MONGO {
		dbName = "memphis-db"
	}
	var collection *mongo.Collection = dbClient.Database(dbName).Collection(collectionName)
	return collection
}

func Close(dbi DbInstance, l logger) {
	defer dbi.Cancel()
	defer func() {
		if err := dbi.Client.Disconnect(dbi.Ctx); err != nil {
			l.Errorf("Failed to close Mongodb client: " + err.Error())
		}
	}()
}

// System Keys Functions
func GetSystemKey(key string) (bool, models.SystemKey, error) {
	filter := bson.M{"key": key}
	var systemKey models.SystemKey
	err := systemKeysCollection.FindOne(context.TODO(), filter).Decode(&systemKey)
	if err == mongo.ErrNoDocuments {
		return false, models.SystemKey{}, nil
	}
	if err != nil {
		return true, models.SystemKey{}, err
	}
	return true, systemKey, nil
}

func InsertSystemKey(key string, value string) error {
	systemKey := models.SystemKey{
		ID:    primitive.NewObjectID(),
		Key:   key,
		Value: value,
	}
	_, err := systemKeysCollection.InsertOne(context.TODO(), systemKey)
	return err
}

func EditSystemKey(key string, value string) error {
	_, err := systemKeysCollection.UpdateOne(context.TODO(),
		bson.M{"key": "analytics"},
		bson.M{"$set": bson.M{"value": value}},
	)
	if err != nil {
		return err
	}
	return nil
}

// Configuration Functions
func GetConfiguration(key string, isString bool) (bool, models.ConfigurationsStringValue, models.ConfigurationsIntValue, error) {
	var configurationsStringValue models.ConfigurationsStringValue
	var configurationsIntValue models.ConfigurationsIntValue
	filter := bson.M{"key": key}
	if isString {
		err := configurationsCollection.FindOne(context.TODO(), filter).Decode(&configurationsStringValue)
		if err == mongo.ErrNoDocuments {
			return false, models.ConfigurationsStringValue{}, models.ConfigurationsIntValue{}, nil
		}
		if err != nil {
			return true, models.ConfigurationsStringValue{}, models.ConfigurationsIntValue{}, err
		}
		return true, configurationsStringValue, models.ConfigurationsIntValue{}, err
	} else {
		err := configurationsCollection.FindOne(context.TODO(), filter).Decode(&configurationsIntValue)
		return true, models.ConfigurationsStringValue{}, configurationsIntValue, err
	}
}

func InsertConfiguration(key string, stringValue string, intValue int, isString bool) error {
	if isString {
		config := models.ConfigurationsStringValue{
			ID:    primitive.NewObjectID(),
			Key:   key,
			Value: stringValue,
		}
		_, err := configurationsCollection.InsertOne(context.TODO(), config)
		if err != nil {
			return err
		}
	} else {
		config := models.ConfigurationsIntValue{
			ID:    primitive.NewObjectID(),
			Key:   key,
			Value: intValue,
		}
		_, err := configurationsCollection.InsertOne(context.TODO(), config)
		if err != nil {
			return err
		}
	}
	return nil
}

func UpdateConfiguration(key string, stringValue string, intValue int, isString bool) error {
	filter := bson.M{"key": key}
	opts := options.Update().SetUpsert(true)
	var update primitive.M
	if isString {
		update = bson.M{
			"$set": bson.M{
				"value": stringValue,
			},
		}
	} else {
		update = bson.M{
			"$set": bson.M{
				"value": intValue,
			},
		}
	}
	_, err := configurationsCollection.UpdateOne(context.TODO(), filter, update, opts)
	if err != nil {
		return err
	}
	return nil
}

// Connection Functions
func InsertConnection(connection models.Connection) error {
	_, err := connectionsCollection.InsertOne(context.TODO(), connection)
	if err != nil {
		return err
	}
	return err
}

func UpdateConnection(connectionId primitive.ObjectID, isActive bool) error {
	_, err := connectionsCollection.UpdateOne(context.TODO(),
		bson.M{"_id": connectionId},
		bson.M{"$set": bson.M{"is_active": isActive}},
	)
	if err != nil {
		return err
	}
	return nil
}

func UpdateConncetionsOfDeletedUser(username string) error {
	_, err := connectionsCollection.UpdateMany(context.TODO(),
		bson.M{"created_by_user": username},
		bson.M{"$set": bson.M{"created_by_user": username + "(deleted)"}},
	)
	if err != nil {
		return err
	}
	return nil
}

func GetConnectionByID(connectionId primitive.ObjectID) (bool, models.Connection, error) {
	filter := bson.M{"_id": connectionId}
	var connection models.Connection
	err := connectionsCollection.FindOne(context.TODO(), filter).Decode(&connection)
	if err == mongo.ErrNoDocuments {
		return false, connection, nil
	} else if err != nil {
		return true, connection, err
	}
	return true, connection, nil
}

func KillRelevantConnections(ids []primitive.ObjectID) error {
	_, err := connectionsCollection.UpdateMany(context.TODO(),
		bson.M{"_id": bson.M{"$in": ids}},
		bson.M{"$set": bson.M{"is_active": false}},
	)
	if err != nil {
		return err
	}

	return nil
}

func GetActiveConnections() ([]models.Connection, error) {
	var connections []models.Connection
	cursor, err := connectionsCollection.Find(context.TODO(), bson.M{"is_active": true})
	if err != nil {
		return connections, err
	}
	if err = cursor.All(context.TODO(), &connections); err != nil {
		return connections, err
	}

	return connections, nil
}

// Audit Logs Functions
func InsertAuditLogs(auditLogs []interface{}) error {
	_, err := auditLogsCollection.InsertMany(context.TODO(), auditLogs)
	if err != nil {
		return err
	}
	return nil
}

func GetAuditLogsByStation(name string) ([]models.AuditLog, error) {
	var auditLogs []models.AuditLog

	cursor, err := auditLogsCollection.Find(context.TODO(), bson.M{"station_name": name, "creation_date": bson.M{
		"$gte": (time.Now().AddDate(0, 0, -5)),
	}})
	if err != nil {
		return []models.AuditLog{}, err
	}

	if err = cursor.All(context.TODO(), &auditLogs); err != nil {
		return []models.AuditLog{}, err
	}

	if len(auditLogs) == 0 {
		auditLogs = []models.AuditLog{}
	}

	return auditLogs, nil
}

func RemoveAllAuditLogsByStation(name string) error {
	_, err := auditLogsCollection.DeleteMany(context.TODO(), bson.M{"station_name": name})
	if err != nil {
		return err
	}
	return nil
}

// Station Functions
func GetActiveStations() ([]models.Station, error) {
	var stations []models.Station
	cursor, err := stationsCollection.Find(context.TODO(), bson.M{
		"$or": []interface{}{
			bson.M{"is_deleted": false},
			bson.M{"is_deleted": bson.M{"$exists": false}},
		},
	})
	if err != nil {
		return []models.Station{}, err
	}

	if err = cursor.All(context.TODO(), &stations); err != nil {
		return []models.Station{}, err
	}
	return stations, nil
}

func GetStationByName(name string) (bool, models.Station, error) {
	var station models.Station
	err := stationsCollection.FindOne(context.TODO(), bson.M{
		"name": name,
		"$or": []interface{}{
			bson.M{"is_deleted": false},
			bson.M{"is_deleted": bson.M{"$exists": false}},
		},
	}).Decode(&station)
	if err == mongo.ErrNoDocuments {
		return false, models.Station{}, nil
	} else if err != nil {
		return true, models.Station{}, err
	}
	return true, station, nil
}

func UpdateNewStation(stationName string, username string, retentionType string, retentionValue int, storageType string, replicas int, dedupEnabled bool, dedupWindowMillis int, schemaDetails models.SchemaDetails, idempotencyWindow int64, isNative bool, dlsConfiguration models.DlsConfiguration, tieredStorageEnabled bool) (models.Station, int64, error) {
	var update bson.M
	var emptySchemaDetailsResponse struct{}
	newStation := models.Station{
		ID:                   primitive.NewObjectID(),
		Name:                 stationName,
		CreatedByUser:        username,
		CreationDate:         time.Now(),
		IsDeleted:            false,
		RetentionType:        retentionType,
		RetentionValue:       retentionValue,
		StorageType:          storageType,
		Replicas:             replicas,
		DedupEnabled:         dedupEnabled,      // TODO deprecated
		DedupWindowInMs:      dedupWindowMillis, // TODO deprecated
		LastUpdate:           time.Now(),
		Schema:               schemaDetails,
		Functions:            []models.Function{},
		IdempotencyWindow:    idempotencyWindow,
		IsNative:             isNative,
		DlsConfiguration:     dlsConfiguration,
		TieredStorageEnabled: tieredStorageEnabled,
	}
	if schemaDetails.SchemaName != "" {
		update = bson.M{
			"$setOnInsert": bson.M{
				"_id":                      newStation.ID,
				"retention_type":           newStation.RetentionType,
				"retention_value":          newStation.RetentionValue,
				"storage_type":             newStation.StorageType,
				"replicas":                 newStation.Replicas,
				"dedup_enabled":            newStation.DedupEnabled,    // TODO deprecated
				"dedup_window_in_ms":       newStation.DedupWindowInMs, // TODO deprecated
				"created_by_user":          newStation.CreatedByUser,
				"creation_date":            newStation.CreationDate,
				"last_update":              newStation.LastUpdate,
				"functions":                newStation.Functions,
				"schema":                   newStation.Schema,
				"idempotency_window_in_ms": newStation.IdempotencyWindow,
				"is_native":                newStation.IsNative,
				"dls_configuration":        newStation.DlsConfiguration,
				"tiered_storage_enabled":   newStation.TieredStorageEnabled,
			},
		}
	} else {
		update = bson.M{
			"$setOnInsert": bson.M{
				"_id":                      newStation.ID,
				"retention_type":           newStation.RetentionType,
				"retention_value":          newStation.RetentionValue,
				"storage_type":             newStation.StorageType,
				"replicas":                 newStation.Replicas,
				"dedup_enabled":            newStation.DedupEnabled,    // TODO deprecated
				"dedup_window_in_ms":       newStation.DedupWindowInMs, // TODO deprecated
				"created_by_user":          newStation.CreatedByUser,
				"creation_date":            newStation.CreationDate,
				"last_update":              newStation.LastUpdate,
				"functions":                newStation.Functions,
				"schema":                   emptySchemaDetailsResponse,
				"idempotency_window_in_ms": newStation.IdempotencyWindow,
				"dls_configuration":        newStation.DlsConfiguration,
				"is_native":                newStation.IsNative,
				"tiered_storage_enabled":   newStation.TieredStorageEnabled,
			},
		}
	}
	filter := bson.M{"name": newStation.Name, "is_deleted": false}
	opts := options.Update().SetUpsert(true)
	updateResults, err := stationsCollection.UpdateOne(context.TODO(), filter, update, opts)
	if err != nil {
		return models.Station{}, 0, err
	}
	return newStation, updateResults.MatchedCount, nil
}

func GetAllStationsDetails() ([]models.ExtendedStation, error) {
	var stations []models.ExtendedStation
	cursor, err := stationsCollection.Aggregate(context.TODO(), mongo.Pipeline{
		bson.D{{"$match", bson.D{{"$or", []interface{}{
			bson.D{{"is_deleted", false}},
			bson.D{{"is_deleted", bson.D{{"$exists", false}}}},
		}}}}},
		bson.D{{"$lookup", bson.D{{"from", "producers"}, {"localField", "_id"}, {"foreignField", "station_id"}, {"as", "producers"}}}},
		bson.D{{"$lookup", bson.D{{"from", "consumers"}, {"localField", "_id"}, {"foreignField", "station_id"}, {"as", "consumers"}}}},
		bson.D{{"$project", bson.D{{"_id", 1}, {"name", 1}, {"retention_type", 1}, {"retention_value", 1}, {"storage_type", 1}, {"replicas", 1}, {"idempotency_window_in_ms", 1}, {"created_by_user", 1}, {"creation_date", 1}, {"last_update", 1}, {"functions", 1}, {"dls_configuration", 1}, {"is_native", 1}, {"producers", 1}, {"consumers", 1}, {"tiered_storage_enabled", 1}}}},
	})
	if err == mongo.ErrNoDocuments {
		return []models.ExtendedStation{}, nil
	}
	if err != nil {
		return []models.ExtendedStation{}, err
	}

	if err = cursor.All(context.TODO(), &stations); err != nil {
		return []models.ExtendedStation{}, err
	}
	return stations, nil
}

func DeleteStations(stationNames []string) error {
	_, err := stationsCollection.UpdateMany(context.TODO(),
		bson.M{
			"name": bson.M{"$in": stationNames},
			"$or": []interface{}{
				bson.M{"is_deleted": false},
				bson.M{"is_deleted": bson.M{"$exists": false}},
			},
		},
		bson.M{"$set": bson.M{"is_deleted": true}},
	)
	if err != nil {
		return err
	}
	return nil
}

func DeleteStation(name string) error {
	_, err := stationsCollection.UpdateOne(context.TODO(),
		bson.M{
			"name": name,
			"$or": []interface{}{
				bson.M{"is_deleted": false},
				bson.M{"is_deleted": bson.M{"$exists": false}},
			},
		},
		bson.M{"$set": bson.M{"is_deleted": true}},
	)
	if err != nil {
		return err
	}
	return nil
}

func AttachSchemaToStation(stationName string, schemaDetails models.SchemaDetails) error {
	_, err := stationsCollection.UpdateOne(context.TODO(), bson.M{"name": stationName, "is_deleted": false}, bson.M{"$set": bson.M{"schema": schemaDetails}})
	if err != nil {
		return err
	}
	return nil
}

func DetachSchemaFromStation(stationName string) error {
	_, err := stationsCollection.UpdateOne(context.TODO(),
		bson.M{
			"name": stationName,
			"$or": []interface{}{
				bson.M{"is_deleted": false},
				bson.M{"is_deleted": bson.M{"$exists": false}},
			},
		},
		bson.M{"$set": bson.M{"schema": bson.M{}}},
	)
	if err != nil {
		return err
	}
	return nil
}

func UpdateStationDlsConfig(stationName string, dlsConfiguration models.DlsConfiguration) error {
	filter := bson.M{
		"name": stationName,
		"$or": []interface{}{
			bson.M{"is_deleted": false},
			bson.M{"is_deleted": bson.M{"$exists": false}},
		}}

	update := bson.M{
		"$set": bson.M{
			"dls_configuration": dlsConfiguration,
		},
	}
	opts := options.Update().SetUpsert(true)

	_, err := stationsCollection.UpdateOne(context.TODO(), filter, update, opts)
	if err != nil {
		return err
	}
	return nil
}

func UpdateIsNativeOldStations() error {
	_, err := stationsCollection.UpdateMany(context.TODO(),
		bson.M{"is_native": bson.M{"$exists": false}},
		bson.M{"$set": bson.M{"is_native": true}},
	)
	if err != nil {
		return err
	}
	return nil
}

func UpdateStationsOfDeletedUser(username string) error {
	_, err := stationsCollection.UpdateMany(context.TODO(),
		bson.M{"created_by_user": username},
		bson.M{"$set": bson.M{"created_by_user": username + "(deleted)"}},
	)
	if err != nil {
		return err
	}
	return nil
}

func GetStationNamesUsingSchema(schemaName string) ([]string, error) {
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

func GetCountStationsUsingSchema(schemaName string) (int, error) {
	filter := bson.M{"schema.name": schemaName, "is_deleted": false}
	countStations, err := stationsCollection.CountDocuments(context.TODO(), filter)
	if err != nil {
		return 0, err
	}

	return int(countStations), nil
}

func RemoveSchemaFromAllUsingStations(schemaName string) error {
	_, err := stationsCollection.UpdateMany(context.TODO(),
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

// Producer Functions
func GetProducersByConnectionIDWithStationDetails(connectionId primitive.ObjectID) ([]models.ExtendedProducer, error) {
	var producers []models.ExtendedProducer
	cursor, err := producersCollection.Aggregate(context.TODO(), mongo.Pipeline{
		bson.D{{"$match", bson.D{{"connection_id", connectionId}, {"is_active", true}}}},
		bson.D{{"$lookup", bson.D{{"from", "stations"}, {"localField", "station_id"}, {"foreignField", "_id"}, {"as", "station"}}}},
		bson.D{{"$unwind", bson.D{{"path", "$station"}, {"preserveNullAndEmptyArrays", true}}}},
		bson.D{{"$project", bson.D{{"_id", 1}, {"name", 1}, {"type", 1}, {"connection_id", 1}, {"created_by_user", 1}, {"creation_date", 1}, {"is_active", 1}, {"is_deleted", 1}, {"station_name", "$station.name"}}}},
		bson.D{{"$project", bson.D{{"station", 0}, {"connection", 0}}}},
	})
	if err != nil {
		return []models.ExtendedProducer{}, err
	}
	if err = cursor.All(context.TODO(), &producers); err != nil {
		return []models.ExtendedProducer{}, err
	}
	return producers, nil
}

func UpdateProducersConnection(connectionId primitive.ObjectID, isActive bool) error {
	_, err := producersCollection.UpdateMany(context.TODO(),
		bson.M{"connection_id": connectionId},
		bson.M{"$set": bson.M{"is_active": isActive}},
	)
	if err != nil {
		return err
	}
	return nil
}

func GetProducerByConnectionID(name string, connectionId primitive.ObjectID) (bool, models.Producer, error) {
	filter := bson.M{"name": name, "connection_id": connectionId}
	var producer models.Producer
	err := producersCollection.FindOne(context.TODO(), filter).Decode(&producer)
	if err == mongo.ErrNoDocuments {
		return false, models.Producer{}, err
	}
	if err != nil {
		return true, models.Producer{}, err
	}
	return true, producer, nil
}

func GetProducerByStationIDAndUsername(username string, stationId primitive.ObjectID, connectionId primitive.ObjectID) (bool, models.Producer, error) {
	filter := bson.M{"name": username, "station_id": stationId, "connection_id": connectionId}
	var producer models.Producer
	err := producersCollection.FindOne(context.TODO(), filter).Decode(&producer)
	if err == mongo.ErrNoDocuments {
		return false, models.Producer{}, err
	}
	if err != nil {
		return true, models.Producer{}, err
	}
	return true, producer, nil
}

func GetActiveProducerByStationID(producerName string, stationId primitive.ObjectID) (bool, models.Producer, error) {
	filter := bson.M{"name": producerName, "station_id": stationId, "is_active": true}
	var producer models.Producer
	err := producersCollection.FindOne(context.TODO(), filter).Decode(&producer)
	if err == mongo.ErrNoDocuments {
		return false, producer, nil
	} else if err != nil {
		return true, producer, err
	}
	return true, producer, nil
}

func UpdateNewProducer(name string, stationId primitive.ObjectID, producerType string, connectionIdObj primitive.ObjectID, createdByUser string) (models.Producer, int64, error) {
	newProducer := models.Producer{
		ID:            primitive.NewObjectID(),
		Name:          name,
		StationId:     stationId,
		Type:          producerType,
		ConnectionId:  connectionIdObj,
		CreatedByUser: createdByUser,
		IsActive:      true,
		CreationDate:  time.Now(),
		IsDeleted:     false,
	}

	filter := bson.M{"name": newProducer.Name, "station_id": stationId, "is_active": true, "is_deleted": false}
	update := bson.M{
		"$setOnInsert": bson.M{
			"_id":             newProducer.ID,
			"type":            newProducer.Type,
			"connection_id":   newProducer.ConnectionId,
			"created_by_user": newProducer.CreatedByUser,
			"creation_date":   newProducer.CreationDate,
		},
	}
	opts := options.Update().SetUpsert(true)
	updateResults, err := producersCollection.UpdateOne(context.TODO(), filter, update, opts)
	if err != nil {
		return newProducer, 0, err
	}
	return newProducer, updateResults.MatchedCount, nil
}

func GetAllProducers() ([]models.ExtendedProducer, error) {
	var producers []models.ExtendedProducer
	cursor, err := producersCollection.Aggregate(context.TODO(), mongo.Pipeline{
		bson.D{{"$match", bson.D{}}},
		bson.D{{"$lookup", bson.D{{"from", "stations"}, {"localField", "station_id"}, {"foreignField", "_id"}, {"as", "station"}}}},
		bson.D{{"$unwind", bson.D{{"path", "$station"}, {"preserveNullAndEmptyArrays", true}}}},
		bson.D{{"$lookup", bson.D{{"from", "connections"}, {"localField", "connection_id"}, {"foreignField", "_id"}, {"as", "connection"}}}},
		bson.D{{"$unwind", bson.D{{"path", "$connection"}, {"preserveNullAndEmptyArrays", true}}}},
		bson.D{{"$project", bson.D{{"_id", 1}, {"name", 1}, {"type", 1}, {"connection_id", 1}, {"created_by_user", 1}, {"creation_date", 1}, {"is_active", 1}, {"is_deleted", 1}, {"station_name", "$station.name"}, {"client_address", "$connection.client_address"}}}},
		bson.D{{"$project", bson.D{{"station", 0}, {"connection", 0}}}},
	})
	if err != nil {
		return []models.ExtendedProducer{}, err
	}

	if err = cursor.All(context.TODO(), &producers); err != nil {
		return []models.ExtendedProducer{}, err
	}
	return producers, nil
}

func GetProducersByStationID(stationId primitive.ObjectID) ([]models.ExtendedProducer, error) {
	var producers []models.ExtendedProducer

	cursor, err := producersCollection.Aggregate(context.TODO(), mongo.Pipeline{
		bson.D{{"$match", bson.D{{"station_id", stationId}}}},
		bson.D{{"$sort", bson.D{{"creation_date", -1}}}},
		bson.D{{"$lookup", bson.D{{"from", "stations"}, {"localField", "station_id"}, {"foreignField", "_id"}, {"as", "station"}}}},
		bson.D{{"$unwind", bson.D{{"path", "$station"}, {"preserveNullAndEmptyArrays", true}}}},
		bson.D{{"$lookup", bson.D{{"from", "connections"}, {"localField", "connection_id"}, {"foreignField", "_id"}, {"as", "connection"}}}},
		bson.D{{"$unwind", bson.D{{"path", "$connection"}, {"preserveNullAndEmptyArrays", true}}}},
		bson.D{{"$project", bson.D{{"_id", 1}, {"name", 1}, {"type", 1}, {"connection_id", 1}, {"created_by_user", 1}, {"creation_date", 1}, {"is_active", 1}, {"is_deleted", 1}, {"station_name", "$station.name"}, {"client_address", "$connection.client_address"}}}},
		bson.D{{"$project", bson.D{{"station", 0}, {"connection", 0}}}},
	})

	if err != nil {
		return []models.ExtendedProducer{}, err
	}

	if err = cursor.All(context.TODO(), &producers); err != nil {
		return []models.ExtendedProducer{}, err
	}
	return producers, nil
}

func DeleteProducer(name string, stationId primitive.ObjectID) (bool, models.Producer, error) {
	var producer models.Producer
	err := producersCollection.FindOneAndUpdate(context.TODO(),
		bson.M{"name": name, "station_id": stationId, "is_active": true},
		bson.M{"$set": bson.M{"is_active": false, "is_deleted": true}},
	).Decode(&producer)
	if err == mongo.ErrNoDocuments {
		return false, models.Producer{}, nil
	}
	if err != nil {
		return true, models.Producer{}, err
	}
	return true, producer, nil
}

func DeleteProducersByStationID(stationId primitive.ObjectID) error {
	_, err := producersCollection.UpdateMany(context.TODO(),
		bson.M{"station_id": stationId},
		bson.M{"$set": bson.M{"is_active": false, "is_deleted": true}},
	)
	if err != nil {
		return err
	}
	return nil
}

func CountActiveProudcersByStationID(stationId primitive.ObjectID) (int64, error) {
	activeCount, err := producersCollection.CountDocuments(context.TODO(), bson.M{"station_id": stationId, "is_active": true})
	if err != nil {
		return 0, err
	}
	return activeCount, nil
}

func CountAllActiveProudcers() (int64, error) {
	producersCount, err := producersCollection.CountDocuments(context.TODO(), bson.M{"is_active": true})
	if err != nil {
		return 0, err
	}
	return producersCount, nil
}

func UpdateProducersOfDeletedUser(username string) error {
	_, err := producersCollection.UpdateMany(context.TODO(),
		bson.M{"created_by_user": username},
		bson.M{"$set": bson.M{"created_by_user": username + "(deleted)", "is_active": false}},
	)
	if err != nil {
		return err
	}
	return nil
}

func KillProducersByConnections(connectionIds []primitive.ObjectID) error {
	_, err := producersCollection.UpdateMany(context.TODO(),
		bson.M{"connection_id": bson.M{"$in": connectionIds}},
		bson.M{"$set": bson.M{"is_active": false}},
	)
	if err != nil {
		return err
	}

	return nil
}

// Consumer Functions
func GetActiveConsumerByCG(consumersGroup string, stationId primitive.ObjectID) (bool, models.Consumer, error) {
	filter := bson.M{"consumers_group": consumersGroup, "station_id": stationId, "is_deleted": false}
	var consumer models.Consumer
	err := consumersCollection.FindOne(context.TODO(), filter).Decode(&consumer)
	if err == mongo.ErrNoDocuments {
		return false, models.Consumer{}, nil
	}
	if err != nil {
		return true, models.Consumer{}, err
	}
	return true, consumer, nil
}

func UpdateNewConsumer(name string, stationId primitive.ObjectID, consumerType string, connectionIdObj primitive.ObjectID, createdByUser string, cgName string, maxAckTime int, maxMsgDeliveries int, startConsumeFromSequence uint64, lastMessages int64) (models.Consumer, int64, error) {
	newConsumer := models.Consumer{
		ID:                       primitive.NewObjectID(),
		Name:                     name,
		StationId:                stationId,
		Type:                     consumerType,
		ConnectionId:             connectionIdObj,
		CreatedByUser:            createdByUser,
		ConsumersGroup:           cgName,
		IsActive:                 true,
		CreationDate:             time.Now(),
		IsDeleted:                false,
		MaxAckTimeMs:             int64(maxAckTime),
		MaxMsgDeliveries:         maxMsgDeliveries,
		StartConsumeFromSequence: startConsumeFromSequence,
		LastMessages:             lastMessages,
	}
	filter := bson.M{"name": newConsumer.Name, "station_id": stationId, "is_active": true, "is_deleted": false}
	update := bson.M{
		"$setOnInsert": bson.M{
			"_id":                         newConsumer.ID,
			"type":                        newConsumer.Type,
			"connection_id":               newConsumer.ConnectionId,
			"created_by_user":             newConsumer.CreatedByUser,
			"consumers_group":             newConsumer.ConsumersGroup,
			"creation_date":               newConsumer.CreationDate,
			"max_ack_time_ms":             newConsumer.MaxAckTimeMs,
			"max_msg_deliveries":          newConsumer.MaxMsgDeliveries,
			"start_consume_from_sequence": newConsumer.StartConsumeFromSequence,
			"last_messages":               newConsumer.LastMessages,
		},
	}
	opts := options.Update().SetUpsert(true)
	updateResults, err := consumersCollection.UpdateOne(context.TODO(), filter, update, opts)
	if err != nil {
		return newConsumer, 0, err
	}
	return newConsumer, updateResults.MatchedCount, nil
}

func GetAllConsumers() ([]models.ExtendedConsumer, error) {
	var consumers []models.ExtendedConsumer
	cursor, err := consumersCollection.Aggregate(context.TODO(), mongo.Pipeline{
		bson.D{{"$match", bson.D{}}},
		bson.D{{"$lookup", bson.D{{"from", "stations"}, {"localField", "station_id"}, {"foreignField", "_id"}, {"as", "station"}}}},
		bson.D{{"$unwind", bson.D{{"path", "$station"}, {"preserveNullAndEmptyArrays", true}}}},
		bson.D{{"$lookup", bson.D{{"from", "connections"}, {"localField", "connection_id"}, {"foreignField", "_id"}, {"as", "connection"}}}},
		bson.D{{"$unwind", bson.D{{"path", "$connection"}, {"preserveNullAndEmptyArrays", true}}}},
		bson.D{{"$project", bson.D{{"_id", 1}, {"name", 1}, {"type", 1}, {"connection_id", 1}, {"created_by_user", 1}, {"consumers_group", 1}, {"creation_date", 1}, {"is_active", 1}, {"is_deleted", 1}, {"max_ack_time_ms", 1}, {"max_msg_deliveries", 1}, {"station_name", "$station.name"}, {"client_address", "$connection.client_address"}}}},
	})
	if err != nil {
		return []models.ExtendedConsumer{}, err
	}
	if err = cursor.All(context.TODO(), &consumers); err != nil {
		return []models.ExtendedConsumer{}, err
	}
	return consumers, nil
}

func GetAllConsumersByStation(stationId primitive.ObjectID) ([]models.ExtendedConsumer, error) {
	var consumers []models.ExtendedConsumer
	cursor, err := consumersCollection.Aggregate(context.TODO(), mongo.Pipeline{
		bson.D{{"$match", bson.D{{"station_id", stationId}}}},
		bson.D{{"$lookup", bson.D{{"from", "stations"}, {"localField", "station_id"}, {"foreignField", "_id"}, {"as", "station"}}}},
		bson.D{{"$unwind", bson.D{{"path", "$station"}, {"preserveNullAndEmptyArrays", true}}}},
		bson.D{{"$lookup", bson.D{{"from", "connections"}, {"localField", "connection_id"}, {"foreignField", "_id"}, {"as", "connection"}}}},
		bson.D{{"$unwind", bson.D{{"path", "$connection"}, {"preserveNullAndEmptyArrays", true}}}},
		bson.D{{"$project", bson.D{{"_id", 1}, {"name", 1}, {"type", 1}, {"connection_id", 1}, {"created_by_user", 1}, {"consumers_group", 1}, {"creation_date", 1}, {"is_active", 1}, {"is_deleted", 1}, {"max_ack_time_ms", 1}, {"max_msg_deliveries", 1}, {"station_name", "$station.name"}, {"client_address", "$connection.client_address"}}}},
		bson.D{{"$project", bson.D{{"station", 0}, {"connection", 0}}}},
	})
	if err != nil {
		return []models.ExtendedConsumer{}, err
	}

	if err = cursor.All(context.TODO(), &consumers); err != nil {
		return []models.ExtendedConsumer{}, err
	}
	return consumers, nil
}

func DeleteConsumer(name string, stationId primitive.ObjectID) (bool, models.Consumer, error) {
	var consumer models.Consumer
	err := consumersCollection.FindOneAndUpdate(context.TODO(),
		bson.M{"name": name, "station_id": stationId, "is_active": true},
		bson.M{"$set": bson.M{"is_active": false, "is_deleted": true}},
	).Decode(&consumer)
	if err == mongo.ErrNoDocuments {
		return false, models.Consumer{}, nil
	}
	if err != nil {
		return true, models.Consumer{}, err
	}
	_, err = consumersCollection.UpdateMany(context.TODO(),
		bson.M{"name": name, "station_id": stationId},
		bson.M{"$set": bson.M{"is_active": false, "is_deleted": true}},
	)
	if err == mongo.ErrNoDocuments {
		return false, models.Consumer{}, err
	}
	if err != nil {
		return true, models.Consumer{}, err
	}
	return true, consumer, nil
}

func DeleteConsumersByStationID(stationId primitive.ObjectID) error {
	_, err := consumersCollection.UpdateMany(context.TODO(),
		bson.M{"station_id": stationId},
		bson.M{"$set": bson.M{"is_active": false, "is_deleted": true}},
	)
	if err != nil {
		return err
	}
	return nil
}

func CountActiveConsumersInCG(consumersGroup string, stationId primitive.ObjectID) (int64, error) {
	count, err := consumersCollection.CountDocuments(context.TODO(), bson.M{"station_id": stationId, "consumers_group": consumersGroup, "is_deleted": false})
	if err != nil {
		return 0, err
	}
	return count, nil
}

func CountActiveConsumersByStationID(stationId primitive.ObjectID) (int64, error) {
	activeCount, err := consumersCollection.CountDocuments(context.TODO(), bson.M{"station_id": stationId, "is_active": true})
	if err != nil {
		return 0, err
	}
	return activeCount, nil
}

func CountAllActiveConsumers() (int64, error) {
	consumersCount, err := consumersCollection.CountDocuments(context.TODO(), bson.M{"is_active": true})
	if err != nil {
		return 0, err
	}
	return consumersCount, nil
}

func GetConsumerGroupMembers(cgName string, stationId primitive.ObjectID) ([]models.CgMember, error) {
	var consumers []models.CgMember

	cursor, err := consumersCollection.Aggregate(context.TODO(), mongo.Pipeline{
		bson.D{{"$match", bson.D{{"consumers_group", cgName}, {"station_id", stationId}}}},
		bson.D{{"$sort", bson.D{{"creation_date", -1}}}},
		bson.D{{"$lookup", bson.D{{"from", "connections"}, {"localField", "connection_id"}, {"foreignField", "_id"}, {"as", "connection"}}}},
		bson.D{{"$unwind", bson.D{{"path", "$connection"}, {"preserveNullAndEmptyArrays", true}}}},
		bson.D{{"$project", bson.D{{"name", 1}, {"created_by_user", 1}, {"is_active", 1}, {"is_deleted", 1}, {"max_ack_time_ms", 1}, {"max_msg_deliveries", 1}, {"client_address", "$connection.client_address"}}}},
		bson.D{{"$project", bson.D{{"station", 0}, {"connection", 0}}}},
	})
	if err != nil {
		return []models.CgMember{}, err
	}

	if err = cursor.All(context.TODO(), &consumers); err != nil {
		return []models.CgMember{}, err
	}
	return consumers, nil
}

func GetConsumersByConnectionIDWithStationDetails(connectionId primitive.ObjectID) ([]models.ExtendedConsumer, error) {
	var consumers []models.ExtendedConsumer
	cursor, err := consumersCollection.Aggregate(context.TODO(), mongo.Pipeline{
		bson.D{{"$match", bson.D{{"connection_id", connectionId}, {"is_active", true}}}},
		bson.D{{"$lookup", bson.D{{"from", "stations"}, {"localField", "station_id"}, {"foreignField", "_id"}, {"as", "station"}}}},
		bson.D{{"$unwind", bson.D{{"path", "$station"}, {"preserveNullAndEmptyArrays", true}}}},
		bson.D{{"$project", bson.D{{"_id", 1}, {"name", 1}, {"type", 1}, {"connection_id", 1}, {"created_by_user", 1}, {"creation_date", 1}, {"is_active", 1}, {"is_deleted", 1}, {"station_name", "$station.name"}}}},
		bson.D{{"$project", bson.D{{"station", 0}, {"connection", 0}}}}})
	if err != nil {
		return []models.ExtendedConsumer{}, err
	}
	if err = cursor.All(context.TODO(), &consumers); err != nil {
		return []models.ExtendedConsumer{}, err
	}
	return consumers, nil
}

func GetActiveConsumerByStationID(consumerName string, stationId primitive.ObjectID) (bool, models.Consumer, error) {
	filter := bson.M{"name": consumerName, "station_id": stationId, "is_active": true}
	var consumer models.Consumer
	err := consumersCollection.FindOne(context.TODO(), filter).Decode(&consumer)
	if err == mongo.ErrNoDocuments {
		return false, consumer, nil
	} else if err != nil {
		return true, consumer, err
	}
	return true, consumer, nil
}

func UpdateConsumersConnection(connectionId primitive.ObjectID, isActive bool) error {
	_, err := consumersCollection.UpdateMany(context.TODO(),
		bson.M{"connection_id": connectionId},
		bson.M{"$set": bson.M{"is_active": isActive}},
	)
	if err != nil {
		return err
	}
	return nil
}

func UpdateConsumersOfDeletedUser(username string) error {
	_, err := consumersCollection.UpdateMany(context.TODO(),
		bson.M{"created_by_user": username},
		bson.M{"$set": bson.M{"created_by_user": username + "(deleted)", "is_active": false}},
	)
	if err != nil {
		return err
	}
	return nil
}

func KillConsumersByConnections(connectionIds []primitive.ObjectID) error {
	_, err := consumersCollection.UpdateMany(context.TODO(),
		bson.M{"connection_id": bson.M{"$in": connectionIds}},
		bson.M{"$set": bson.M{"is_active": false}},
	)
	if err != nil {
		return err
	}

	return nil
}

// Schema Functions
func GetSchemaByName(name string) (bool, models.Schema, error) {
	var schema models.Schema
	err := schemasCollection.FindOne(context.TODO(), bson.M{"name": name}).Decode(&schema)
	if err == mongo.ErrNoDocuments {
		return false, models.Schema{}, nil
	}
	if err != nil {
		return true, models.Schema{}, err
	}
	return true, schema, nil
}

func GetSchemaVersionsBySchemaID(id primitive.ObjectID) ([]models.SchemaVersion, error) {
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

func GetActiveVersionBySchemaID(id primitive.ObjectID) (models.SchemaVersion, error) {
	var schemaVersion models.SchemaVersion
	err := schemaVersionCollection.FindOne(context.TODO(), bson.M{"schema_id": id, "active": true}).Decode(&schemaVersion)
	if err != nil {
		return models.SchemaVersion{}, err
	}
	return schemaVersion, nil
}

func UpdateSchemasOfDeletedUser(username string) error {
	_, err := schemasCollection.UpdateMany(context.TODO(),
		bson.M{"created_by_user": username},
		bson.M{"$set": bson.M{"created_by_user": username + "(deleted)"}},
	)
	if err != nil {
		return err
	}
	return nil
}

func GetSchemaVersionByID(version int, schemaId primitive.ObjectID) (bool, models.SchemaVersion, error) {
	var schemaVersion models.SchemaVersion
	filter := bson.M{"schema_id": schemaId, "version_number": version}
	err := schemaVersionCollection.FindOne(context.TODO(), filter).Decode(&schemaVersion)
	if err == mongo.ErrNoDocuments {
		return false, schemaVersion, nil
	} else if err != nil {
		return true, schemaVersion, err
	}
	return true, schemaVersion, nil
}

func UpdateSchemaActiveVersion(schemaId primitive.ObjectID, versionNumber int) error {
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

func GetShcemaVersionsCount(schemaId primitive.ObjectID) (int, error) {
	countVersions, err := schemaVersionCollection.CountDocuments(context.TODO(), bson.M{"schema_id": schemaId})
	if err != nil {
		return 0, err
	}

	return int(countVersions), err
}

func GetAllSchemasDetails() ([]models.ExtendedSchema, error) {
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
	return schemas, nil
}

func FindAndDeleteSchema(schemaIds []primitive.ObjectID) error {
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

func UpdateNewSchema(schemaName string, schemaType string) (models.Schema, int64, error) {
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
	opts := options.Update().SetUpsert(true)
	updateResults, err := schemasCollection.UpdateOne(context.TODO(), filter, update, opts)
	if err != nil {
		return models.Schema{}, 0, err
	}
	return newSchema, updateResults.MatchedCount, nil
}

func UpdateNewSchemaVersion(schemaVersionNumber int, username string, schemaContent string, schemaId primitive.ObjectID, messageStructName string, descriptor string, active bool) (models.SchemaVersion, int64, error) {
	newSchemaVersion := models.SchemaVersion{
		ID:                primitive.NewObjectID(),
		VersionNumber:     schemaVersionNumber,
		Active:            active,
		CreatedByUser:     username,
		CreationDate:      time.Now(),
		SchemaContent:     schemaContent,
		SchemaId:          schemaId,
		MessageStructName: messageStructName,
		Descriptor:        descriptor,
	}
	filter := bson.M{"schema_id": schemaId, "version_number": schemaVersionNumber}
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
		return models.SchemaVersion{}, 0, err
	}
	return newSchemaVersion, updateResults.MatchedCount, nil
}

// Integration Functions
func GetIntegration(name string) (bool, models.Integration, error) {
	filter := bson.M{"name": name}
	var integration models.Integration
	err := integrationsCollection.FindOne(context.TODO(),
		filter).Decode(&integration)
	if err == mongo.ErrNoDocuments {
		return false, models.Integration{}, nil
	}
	if err != nil {
		return true, models.Integration{}, err
	}
	return true, integration, err
}

func GetAllIntegrations() (bool, []models.Integration, error) {
	var integrations []models.Integration
	cursor, err := integrationsCollection.Find(context.TODO(), bson.M{})
	if err == mongo.ErrNoDocuments {
		return false, []models.Integration{}, nil
	}
	if err != nil {
		return true, []models.Integration{}, err
	}
	if err = cursor.All(context.TODO(), &integrations); err != nil {
		return true, []models.Integration{}, err
	}
	return true, integrations, nil
}

func DeleteIntegration(name string) error {
	filter := bson.M{"name": name}
	_, err := integrationsCollection.DeleteOne(context.TODO(), filter)
	if err != nil {
		return err
	}
	return nil
}

func InsertNewIntegration(name string, keys map[string]string, properties map[string]bool) (models.Integration, error) {
	integration := models.Integration{
		ID:         primitive.NewObjectID(),
		Name:       name,
		Keys:       keys,
		Properties: properties,
	}
	_, err := integrationsCollection.InsertOne(context.TODO(), integration)
	if err != nil {
		return models.Integration{}, err
	}
	return integration, nil
}

func UpdateIntegration(name string, keys map[string]string, properties map[string]bool) (models.Integration, error) {
	var integration models.Integration
	filter := bson.M{"name": name}
	err := integrationsCollection.FindOneAndUpdate(context.TODO(),
		filter,
		bson.M{"$set": bson.M{"keys": keys, "properties": properties}}).Decode(&integration)
	if err == mongo.ErrNoDocuments {
		integration = models.Integration{
			ID:         primitive.NewObjectID(),
			Name:       name,
			Keys:       keys,
			Properties: properties,
		}
		_, err = integrationsCollection.InsertOne(context.TODO(), integration)
		if err != nil {
			return models.Integration{}, err
		}
	} else if err != nil {
		return models.Integration{}, err
	}
	return integration, nil
}

// User Functions
func CreateUser(username string, userType string, hashedPassword string, fullName string, subscription bool, avatarId int) (models.User, error) {
	var id primitive.ObjectID
	if userType == "root" {
		id, _ = primitive.ObjectIDFromHex("6314c8f7ef142f3f04fccdc3") // default root user id
	} else {
		id = primitive.NewObjectID()
	}
	newUser := models.User{
		ID:              id,
		Username:        username,
		Password:        hashedPassword,
		FullName:        fullName,
		Subscribtion:    subscription,
		HubUsername:     "",
		HubPassword:     "",
		UserType:        userType,
		CreationDate:    time.Now(),
		AlreadyLoggedIn: false,
		AvatarId:        avatarId,
	}
	_, err := usersCollection.InsertOne(context.TODO(), newUser)
	if err != nil {
		if userType == "root" {
			if mongo.IsDuplicateKeyError(err) {
				return newUser, nil
			}
		}
		return models.User{}, err
	}
	return newUser, nil
}

func ChangeUserPassword(username string, hashedPassword string) error {
	_, err := usersCollection.UpdateOne(context.TODO(),
		bson.M{"username": username},
		bson.M{"$set": bson.M{"password": hashedPassword}},
	)
	if err != nil {
		return err
	}
	return nil
}

func GetRootUser() (bool, models.User, error) {
	filter := bson.M{"user_type": "root"}
	var user models.User
	err := usersCollection.FindOne(context.TODO(), filter).Decode(&user)
	if err == mongo.ErrNoDocuments {
		return false, models.User{}, nil
	} else if err != nil {
		return true, models.User{}, err
	}
	return true, user, nil
}

func GetUserByUsername(username string) (bool, models.User, error) {
	filter := bson.M{"username": username}
	var user models.User
	err := usersCollection.FindOne(context.TODO(), filter).Decode(&user)
	if err == mongo.ErrNoDocuments {
		return false, models.User{}, nil
	} else if err != nil {
		return true, models.User{}, err
	}
	return true, user, nil
}

func GetAllUsers() ([]models.FilteredGenericUser, error) {
	var users []models.FilteredGenericUser

	cursor, err := usersCollection.Find(context.TODO(), bson.M{})
	if err != nil {
		return []models.FilteredGenericUser{}, err
	}

	if err = cursor.All(context.TODO(), &users); err != nil {
		return []models.FilteredGenericUser{}, err
	}
	return users, nil
}

func GetAllApplicationUsers() ([]models.FilteredApplicationUser, error) {
	var users []models.FilteredApplicationUser

	cursor, err := usersCollection.Find(context.TODO(), bson.M{
		"$or": []interface{}{
			bson.M{"user_type": "application"},
			bson.M{"user_type": "root"},
		},
	})
	if err != nil {
		return []models.FilteredApplicationUser{}, err
	}

	if err = cursor.All(context.TODO(), &users); err != nil {
		return []models.FilteredApplicationUser{}, err
	}
	return users, nil
}

func UpdateUserAlreadyLoggedIn(userId primitive.ObjectID) {
	usersCollection.UpdateOne(context.TODO(),
		bson.M{"_id": userId},
		bson.M{"$set": bson.M{"already_logged_in": true}},
	)
}

func UpdateSkipGetStarted(username string) error {
	_, err := usersCollection.UpdateOne(context.TODO(),
		bson.M{"username": username},
		bson.M{"$set": bson.M{"skip_get_started": true}},
	)
	if err != nil {
		return err
	}
	return nil
}

func DeleteUser(username string) error {
	_, err := usersCollection.DeleteOne(context.TODO(), bson.M{"username": username})
	if err != nil {
		return err
	}
	return nil
}

func EditHubCreds(username string, hubUsername string, password string) error {
	_, err := usersCollection.UpdateOne(context.TODO(),
		bson.M{"username": username},
		bson.M{"$set": bson.M{"hub_username": hubUsername, "hub_password": password}},
	)
	if err != nil {
		return err
	}
	return nil
}

func EditAvatar(username string, avatarId int) error {
	_, err := usersCollection.UpdateOne(context.TODO(),
		bson.M{"username": username},
		bson.M{"$set": bson.M{"avatar_id": avatarId}},
	)
	if err != nil {
		return err
	}
	return nil
}

func GetAllActiveUsers() ([]models.FilteredUser, error) { // This function executed on stations collection
	var userList []models.FilteredUser

	cursorUsers, err := stationsCollection.Aggregate(context.TODO(), mongo.Pipeline{
		bson.D{{"$match", bson.D{{"$or", []interface{}{bson.D{{"is_deleted", false}}, bson.D{{"is_deleted", bson.D{{"$exists", false}}}}}}}}},
		bson.D{{"$lookup", bson.D{{"from", "users"}, {"localField", "created_by_user"}, {"foreignField", "username"}, {"as", "usersList"}}}},
		bson.D{{"$unwind", bson.D{{"path", "$usersList"}, {"preserveNullAndEmptyArrays", true}}}},
		bson.D{{"$group", bson.D{{"_id", "$usersList.username"}, {"items", bson.D{{"$addToSet", bson.D{{"name", "$usersList.username"}}}}}}}},
	})
	if err != nil {
		return []models.FilteredUser{}, err
	}

	if err = cursorUsers.All(context.TODO(), &userList); err != nil {
		return []models.FilteredUser{}, err
	}
	return userList, nil
}

// Tags Functions
func UpdateNewTag(name string, color string, stationArr []primitive.ObjectID, schemaArr []primitive.ObjectID, userArr []primitive.ObjectID) (models.Tag, error) {
	newTag := models.Tag{
		ID:   primitive.NewObjectID(),
		Name: name, Color: color,
		Stations: stationArr,
		Schemas:  schemaArr,
		Users:    userArr,
	}

	filter := bson.M{"name": newTag.Name}
	update := bson.M{
		"$setOnInsert": bson.M{
			"_id":      newTag.ID,
			"name":     newTag.Name,
			"color":    newTag.Color,
			"stations": newTag.Stations,
			"schemas":  newTag.Schemas,
			"users":    newTag.Users,
		},
	}
	opts := options.Update().SetUpsert(true)
	_, err := tagsCollection.UpdateOne(context.TODO(), filter, update, opts)
	if err != nil {
		return models.Tag{}, err
	}
	return newTag, nil
}

func AddTagToEntity(tagName string, entity string, entity_id primitive.ObjectID) error {
	var entityDBList string
	switch entity {
	case "station":
		entityDBList = "stations"
	case "schema":
		entityDBList = "schemas"
	case "user":
		entityDBList = "users"
	}
	filter := bson.M{"name": tagName}
	update := bson.M{
		"$addToSet": bson.M{entityDBList: entity_id},
	}
	opts := options.Update().SetUpsert(true)
	_, err := tagsCollection.UpdateOne(context.TODO(), filter, update, opts)
	if err != nil {
		return err
	}
	return nil
}

func RemoveAllTagsFromEntity(entity string, entity_id primitive.ObjectID) error {
	_, err := tagsCollection.UpdateMany(context.TODO(), bson.M{}, bson.M{"$pull": bson.M{entity: entity_id}})
	if err != nil {
		return err
	}
	return nil
}

func RemoveTagFromEntity(tagName string, entity string, entity_id primitive.ObjectID) error {
	var entityDBList string
	switch entity {
	case "station":
		entityDBList = "stations"
	case "schema":
		entityDBList = "schemas"
	case "user":
		entityDBList = "users"
	}
	_, err := tagsCollection.UpdateOne(context.TODO(), bson.M{"name": tagName},
		bson.M{"$pull": bson.M{entityDBList: entity_id}})
	if err != nil {
		return err
	}
	return nil
}

func GetTagsByEntityID(entity string, id primitive.ObjectID) ([]models.Tag, error) {
	var entityDBList string
	switch entity {
	case "station":
		entityDBList = "stations"
	case "schema":
		entityDBList = "schemas"
	case "user":
		entityDBList = "users"
	}
	var tags []models.Tag
	cursor, err := tagsCollection.Find(context.TODO(), bson.M{entityDBList: id})
	if err != nil {
		return []models.Tag{}, err
	}
	if err = cursor.All(context.TODO(), &tags); err != nil {
		return []models.Tag{}, err
	}
	return tags, nil
}

func GetTagsByEntityType(entity string) ([]models.Tag, error) {
	var tags []models.Tag
	var entityDBList string
	switch entity {
	case "station":
		entityDBList = "stations"
	case "schema":
		entityDBList = "schemas"
	case "user":
		entityDBList = "users"
	default:
		entityDBList = ""
	}
	var cursor *mongo.Cursor
	if entityDBList == "" { // Get All
		cur, err := tagsCollection.Find(context.TODO(), bson.M{})
		if err != nil {
			return []models.Tag{}, err
		}
		cursor = cur
	} else {
		cur, err := tagsCollection.Find(context.TODO(), bson.M{entityDBList: bson.M{"$not": bson.M{"$size": 0}}})
		if err != nil {
			return []models.Tag{}, err
		}
		cursor = cur
	}

	if err := cursor.All(context.TODO(), &tags); err != nil {
		return []models.Tag{}, err
	}
	return tags, nil
}

func GetAllUsedTags() ([]models.Tag, error) {
	var tags []models.Tag
	filter := bson.M{"$or": []interface{}{bson.M{"schemas": bson.M{"$exists": true, "$not": bson.M{"$size": 0}}}, bson.M{"stations": bson.M{"$exists": true, "$not": bson.M{"$size": 0}}}, bson.M{"users": bson.M{"$exists": true, "$not": bson.M{"$size": 0}}}}}
	cursor, err := tagsCollection.Find(context.TODO(), filter)
	if err != nil {
		return []models.Tag{}, err
	}
	if err = cursor.All(context.TODO(), &tags); err != nil {
		return []models.Tag{}, err
	}
	return tags, nil
}

func GetTagByName(name string) (bool, models.Tag, error) {
	filter := bson.M{
		"name": name,
	}
	var tag models.Tag
	err := tagsCollection.FindOne(context.TODO(), filter).Decode(&tag)
	if err == mongo.ErrNoDocuments {
		return false, tag, nil
	} else if err != nil {
		return true, tag, err
	}
	return true, tag, nil
}

// Sandbox Functions
func InsertNewSanboxUser(username string, email string, firstName string, lastName string, profilePic string) (models.SandboxUser, error) {
	user := models.SandboxUser{
		ID:              primitive.NewObjectID(),
		Username:        username,
		Email:           email,
		Password:        "",
		FirstName:       firstName,
		LastName:        lastName,
		HubUsername:     "",
		HubPassword:     "",
		UserType:        "",
		CreationDate:    time.Now(),
		AlreadyLoggedIn: false,
		AvatarId:        1,
		ProfilePic:      profilePic,
	}

	_, err := sandboxUsersCollection.InsertOne(context.TODO(), user)
	if err != nil {
		return models.SandboxUser{}, err
	}
	return user, nil
}

func UpdateSandboxUserAlreadyLoggedIn(userId primitive.ObjectID) {
	sandboxUsersCollection.UpdateOne(context.TODO(),
		bson.M{"_id": userId},
		bson.M{"$set": bson.M{"already_logged_in": true}},
	)
}

func GetSandboxUser(username string) (bool, models.SandboxUser, error) {
	filter := bson.M{"username": username}
	var user models.SandboxUser
	err := sandboxUsersCollection.FindOne(context.TODO(), filter).Decode(&user)
	if err == mongo.ErrNoDocuments {
		return false, user, nil
	}
	if err != nil {
		return true, models.SandboxUser{}, err
	}
	return true, user, nil
}

func UpdateSkipGetStartedSandbox(username string) error {
	_, err := sandboxUsersCollection.UpdateOne(context.TODO(),
		bson.M{"username": username},
		bson.M{"$set": bson.M{"skip_get_started": true}},
	)
	if err != nil {
		return err
	}
	return nil
}

// Image Functions
func InsertImage(name string, base64Encoding string) error {
	newImage := models.Image{
		ID:    primitive.NewObjectID(),
		Name:  name,
		Image: base64Encoding,
	}
	_, err := imagesCollection.InsertOne(context.TODO(), newImage)
	if err != nil {
		return err
	}
	return nil
}

func DeleteImage(name string) error {
	_, err := imagesCollection.DeleteOne(context.TODO(), bson.M{"name": name})
	if err != nil {
		return err
	}
	return nil
}

func GetImage(name string) (bool, models.Image, error) {
	var image models.Image
	err := imagesCollection.FindOne(context.TODO(), bson.M{"name": name}).Decode(&image)
	if err == mongo.ErrNoDocuments {
		return false, models.Image{}, nil
	} else if err != nil {
		return true, models.Image{}, err
	}
	return true, image, nil
}
