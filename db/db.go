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
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"memphis/conf"
	"strings"

	"memphis/models"

	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	// "github.com/jackc/pgx/pgtype"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
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

var postgresConnection DbPostgreSQLInstance

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

type DbPostgreSQLInstance struct {
	Client *pgxpool.Pool
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

func ClosePostgresSql(db DbPostgreSQLInstance, l logger) {
	defer db.Cancel()
	defer func() {
		db.Client.Close()
	}()
}

func JoinTable(dbPostgreSQL *pgxpool.Pool) error {
	ctx, cancelfunc := context.WithTimeout(context.Background(), dbOperationTimeout*time.Second)
	defer cancelfunc()
	query := `SELECT users
    FROM users AS q
    JOIN tags AS a ON q.id = a.users
    WHERE q.id = $1`

	conn, err := dbPostgreSQL.Acquire(ctx)
	if err != nil {
		return err
	}

	defer conn.Release()
	stmt, err := conn.Conn().Prepare(ctx, "join", query)
	if err != nil {
		return err
	}
	_, err = conn.Conn().Exec(ctx, stmt.Name, 1)
	if err != nil {
		return err
	}

	return nil
}

func InsertToTable(dbPostgreSQL *pgxpool.Pool) error {
	ctx, cancelfunc := context.WithTimeout(context.Background(), dbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := dbPostgreSQL.Acquire(ctx)
	if err != nil {
		cancelfunc()
		return err
	}
	defer conn.Release()

	query := `INSERT INTO users (username, password, type, already_logged_in, created_at, avatar_id, full_name, subscription, skip_get_started)
    VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9);`
	_, err = conn.Conn().Prepare(ctx, "insert into", query)
	if err != nil {
		return err
	}

	// createdAt := time.Now()
	_, err = conn.Conn().Exec(ctx, "insert into", "username", "t1212", "root", true, "2005-05-13 07:15:31.123456789", 1, "ttd", true, true)
	if err != nil {
		return err
	}

	return nil
}

func SelectFromTable(dbPostgreSQL *pgxpool.Pool) error {
	ctx, cancelfunc := context.WithTimeout(context.Background(), dbOperationTimeout*time.Second)
	defer cancelfunc()

	conn, err := dbPostgreSQL.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()
	query := `SELECT username FROM users WHERE username = $1`
	stmt, err := conn.Conn().Prepare(ctx, "select_from", query)
	if err != nil {
		return err
	}
	var username string
	rows := conn.Conn().QueryRow(ctx, stmt.Name, "test")

	err = rows.Scan(&username)
	if err != nil {
		if err == pgx.ErrNoRows {
			return err
		}
		return err
	}

	return nil

}

func updateFieldInTable(dbPostgreSQL *pgxpool.Pool) error {
	ctx, cancelfunc := context.WithTimeout(context.Background(), dbOperationTimeout*time.Second)
	defer cancelfunc()

	conn, err := dbPostgreSQL.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()
	query := `UPDATE users
	SET username = $2
	WHERE id = $1
	RETURNING id, username;`
	stmt, err := conn.Conn().Prepare(ctx, "update", query)
	if err != nil {
		return err
	}
	var username string
	var id int
	rows := conn.Conn().QueryRow(ctx, stmt.Name, 7, "test")
	err = rows.Scan(&id, &username)
	if err != nil {
		return err
	}

	return nil
}

func dropRowInTable(dbPostgreSQL *pgxpool.Pool) error {
	ctx, cancelfunc := context.WithTimeout(context.Background(), dbOperationTimeout*time.Second)
	defer cancelfunc()

	conn, err := dbPostgreSQL.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()
	query := `DELETE FROM users WHERE id = $1;`
	stmt, err := conn.Conn().Prepare(ctx, "drop", query)
	if err != nil {
		return err
	}
	_, err = conn.Conn().Exec(ctx, stmt.Name, 7)
	if err != nil {
		return err
	}
	return nil
}

func AddIndexToTable(indexName, tableName, field string, dbPostgreSQL DbPostgreSQLInstance) error {
	ctx, cancelfunc := context.WithTimeout(context.Background(), dbOperationTimeout*time.Second)
	defer cancelfunc()
	addIndexQuery := "CREATE INDEX" + pgx.Identifier{indexName}.Sanitize() + "ON" + pgx.Identifier{tableName}.Sanitize() + "(" + pgx.Identifier{field}.Sanitize() + ")"
	db := dbPostgreSQL.Client
	_, err := db.Exec(ctx, addIndexQuery)
	if err != nil {
		return err
	}
	return nil
}

func createTables(dbPostgreSQL DbPostgreSQLInstance) error {
	cancelfunc := dbPostgreSQL.Cancel
	defer cancelfunc()
	auditLogsTable := `CREATE TABLE IF NOT EXISTS audit_logs(
		id SERIAL NOT NULL,
		station_name VARCHAR NOT NULL,
		message TEXT NOT NULL,
		created_by INTEGER NOT NULL,
		created_at TIMESTAMP NOT NULL,
		PRIMARY KEY (id),
		CONSTRAINT fk_created_by
			FOREIGN KEY(created_by)
			REFERENCES users(id)
		);
	CREATE INDEX station_name
	ON audit_logs (station_name);`

	usersTable := `
	CREATE TYPE enum AS ENUM ('root', 'management', 'application');
	CREATE TABLE IF NOT EXISTS users(
		id SERIAL NOT NULL,
		username VARCHAR NOT NULL UNIQUE,
		password TEXT NOT NULL,
		type enum NOT NULL DEFAULT 'root',
		already_logged_in BOOL NOT NULL DEFAULT false,
		created_at TIMESTAMP NOT NULL,
		avatar_id SERIAL NOT NULL,
		full_name VARCHAR,
		subscription BOOL NOT NULL DEFAULT false,
		skip_get_started BOOL NOT NULL DEFAULT false,
		PRIMARY KEY (id)
		);`

	configurationsTable := `CREATE TABLE IF NOT EXISTS configurations(
		id SERIAL NOT NULL,
		key VARCHAR NOT NULL UNIQUE,
		value TEXT NOT NULL,
		PRIMARY KEY (id)
		);`

	connectionsTable := `CREATE TABLE IF NOT EXISTS connections(
		id VARCHAR NOT NULL,
		created_by INTEGER NOT NULL,
		is_active BOOL NOT NULL DEFAULT false,
		created_at TIMESTAMP NOT NULL,
		client_address VARCHAR NOT NULL,
		PRIMARY KEY (id),
		CONSTRAINT fk_created_by
			FOREIGN KEY(created_by)
			REFERENCES users(id)
		);`

	integrationsTable := `CREATE TABLE IF NOT EXISTS integrations(
		id SERIAL NOT NULL,
		name VARCHAR NOT NULL UNIQUE,
		keys JSON NOT NULL DEFAULT '{}',
		properties JSON NOT NULL DEFAULT '{}',
		PRIMARY KEY (id)
		);`

	schemasTable := `
	CREATE TYPE enum_type AS ENUM ('json', 'graphql', 'protobuf');
	CREATE TABLE IF NOT EXISTS schemas(
		id SERIAL NOT NULL,
		name VARCHAR NOT NULL,
		type enum_type NOT NULL DEFAULT 'protobuf',
		PRIMARY KEY (id),
		UNIQUE(name)
		);
		CREATE INDEX name
		ON schemas (name);`

	tagsTable := `CREATE TABLE IF NOT EXISTS tags(
		id SERIAL NOT NULL,
		name VARCHAR NOT NULL UNIQUE,
		color VARCHAR NOT NULL,
		users INTEGER[] ,
		stations INTEGER[],
		schemas INTEGER[],
		PRIMARY KEY (id)
		);
		CREATE INDEX name_tag
		ON tags (name);`

	consumersTable := `
	CREATE TYPE enum_type_consumer AS ENUM ('application', 'connector');
	CREATE TABLE IF NOT EXISTS consumers(
		id SERIAL NOT NULL,
		name VARCHAR NOT NULL,
		station_id INTEGER NOT NULL,
		type enum_type_consumer NOT NULL DEFAULT 'application',
		connection_id VARCHAR NOT NULL,
		consumers_group VARCHAR NOT NULL,
		max_ack_time_ms SERIAL NOT NULL,
		created_by INTEGER NOT NULL,
		is_active BOOL NOT NULL DEFAULT true,
		created_at TIMESTAMP NOT NULL,
		is_deleted BOOL NOT NULL DEFAULT false,
		max_msg_deliveries SERIAL NOT NULL,
		start_consume_from_seq SERIAL NOT NULL,
		last_msgs SERIAL NOT NULL,
		PRIMARY KEY (id),
		CONSTRAINT fk_created_by
			FOREIGN KEY(created_by)
			REFERENCES users(id),
		CONSTRAINT fk_connection_id
			FOREIGN KEY(connection_id)
			REFERENCES connections(id),
		CONSTRAINT fk_station_id
			FOREIGN KEY(station_id)
			REFERENCES stations(id)
		);
		CREATE INDEX station_id
		ON consumers (station_id);
		CREATE UNIQUE INDEX unique_consumer_table ON consumers(name, station_id, is_active) WHERE is_active = true`

	stationsTable := `
	CREATE TYPE enum_retention_type AS ENUM ('message_age_sec', 'messages', 'bytes');
	CREATE TYPE enum_storage_type AS ENUM ('file', 'memory');
	CREATE TABLE IF NOT EXISTS stations(
		id SERIAL NOT NULL,
		name VARCHAR NOT NULL,
		retention_type enum_retention_type NOT NULL DEFAULT 'message_age_sec',
		retention_value SERIAL NOT NULL,
		storage_type enum_storage_type NOT NULL DEFAULT 'file',
		replicas SERIAL NOT NULL,
		created_by INTEGER NOT NULL,
		created_at TIMESTAMP NOT NULL,
		updated_at TIMESTAMP NOT NULL,
		is_deleted BOOL NOT NULL,
		schema_name VARCHAR,
		schema_version_number SERIAL,
		idempotency_window_ms SERIAL NOT NULL,
		is_native BOOL NOT NULL ,
		dls_configuration_poison BOOL NOT NULL DEFAULT true,
		dls_configuration_schemaverse BOOL NOT NULL DEFAULT true,
		tiered_storage_enabled BOOL NOT NULL,
		PRIMARY KEY (id),
		CONSTRAINT fk_created_by
			FOREIGN KEY(created_by)
			REFERENCES users(id)
		);
		CREATE UNIQUE INDEX unique_station_name_deleted ON stations(name, is_deleted) WHERE is_deleted = false;`

	schemaVersionsTable := `CREATE TABLE IF NOT EXISTS schema_versions(
		id SERIAL NOT NULL,
		version_number SERIAL NOT NULL,
		active BOOL NOT NULL DEFAULT false,
		created_by INTEGER NOT NULL,
		created_at TIMESTAMP NOT NULL,
		schema_content TEXT NOT NULL,
		schema_id INTEGER NOT NULL,
		msg_struct_name VARCHAR,
		descriptor TEXT,
		PRIMARY KEY (id),
		UNIQUE(version_number, schema_id),
		CONSTRAINT fk_created_by
			FOREIGN KEY(created_by)
			REFERENCES users(id),
		CONSTRAINT fk_schema_id
			FOREIGN KEY(schema_id)
			REFERENCES schemas(id)
		);`

	producersTable := `
	CREATE TYPE enum_producer_type AS ENUM ('application', 'connector');
	CREATE TABLE IF NOT EXISTS producers(
		id SERIAL NOT NULL,
		name VARCHAR NOT NULL,
		station_id INTEGER NOT NULL,
		type enum_producer_type NOT NULL DEFAULT 'application',
		connection_id VARCHAR NOT NULL,	
		created_by INTEGER NOT NULL,
		is_active BOOL NOT NULL DEFAULT true,
		created_at TIMESTAMP NOT NULL,
		is_deleted BOOL NOT NULL DEFAULT false,
		PRIMARY KEY (id),
		CONSTRAINT fk_created_by
			FOREIGN KEY(created_by)
			REFERENCES users(id),
		CONSTRAINT fk_station_id
			FOREIGN KEY(station_id)
			REFERENCES stations(id),
		CONSTRAINT fk_connection_id
			FOREIGN KEY(connection_id)
			REFERENCES connections(id)
		);
		CREATE INDEX producer_station_id
		ON producers(station_id);
		CREATE UNIQUE INDEX unique_producer_table ON producers(name, station_id, is_active) WHERE is_active = true;`

	db := dbPostgreSQL.Client
	ctx := dbPostgreSQL.Ctx

	_, err := db.Exec(ctx, usersTable)
	if err != nil {
		var pgErr *pgconn.PgError
		errPg := errors.As(err, &pgErr)
		if errPg && !strings.Contains(pgErr.Message, "already exists") {
			return err
		}
	}

	_, err = db.Exec(ctx, connectionsTable)
	if err != nil {
		var pgErr *pgconn.PgError
		errPg := errors.As(err, &pgErr)
		if errPg && !strings.Contains(pgErr.Message, "already exists") {
			return err
		}
	}

	_, err = db.Exec(ctx, auditLogsTable)
	if err != nil {
		var pgErr *pgconn.PgError
		errPg := errors.As(err, &pgErr)
		if errPg && !strings.Contains(pgErr.Message, "already exists") {
			return err
		}
	}

	_, err = db.Exec(ctx, configurationsTable)
	if err != nil {
		var pgErr *pgconn.PgError
		errPg := errors.As(err, &pgErr)
		if errPg && !strings.Contains(pgErr.Message, "already exists") {
			return err
		}
	}

	_, err = db.Exec(ctx, integrationsTable)
	if err != nil {
		var pgErr *pgconn.PgError
		errPg := errors.As(err, &pgErr)
		if errPg && !strings.Contains(pgErr.Message, "already exists") {
			return err
		}
	}

	_, err = db.Exec(ctx, schemasTable)
	if err != nil {
		var pgErr *pgconn.PgError
		errPg := errors.As(err, &pgErr)
		if errPg && !strings.Contains(pgErr.Message, "already exists") {
			return err
		}
	}

	_, err = db.Exec(ctx, tagsTable)
	if err != nil {
		var pgErr *pgconn.PgError
		errPg := errors.As(err, &pgErr)
		if errPg && !strings.Contains(pgErr.Message, "already exists") {
			return err
		}
	}

	_, err = db.Exec(ctx, stationsTable)
	if err != nil {
		var pgErr *pgconn.PgError
		errPg := errors.As(err, &pgErr)
		if errPg && !strings.Contains(pgErr.Message, "already exists") {
			return err
		}
	}
	_, err = db.Exec(ctx, consumersTable)
	if err != nil {
		var pgErr *pgconn.PgError
		errPg := errors.As(err, &pgErr)
		if errPg && !strings.Contains(pgErr.Message, "already exists") {
			return err
		}
	}

	_, err = db.Exec(ctx, schemaVersionsTable)
	if err != nil {
		var pgErr *pgconn.PgError
		errPg := errors.As(err, &pgErr)
		if errPg && !strings.Contains(pgErr.Message, "already exists") {
			return err
		}
	}
	_, err = db.Exec(ctx, producersTable)
	if err != nil {
		var pgErr *pgconn.PgError
		errPg := errors.As(err, &pgErr)
		if errPg && !strings.Contains(pgErr.Message, "already exists") {
			return err
		}
	}

	return nil
}

func InitalizePostgreSQLDbConnection(l logger) (DbPostgreSQLInstance, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), dbOperationTimeout*time.Second)

	defer cancelfunc()
	postgreSqlUser := configuration.POSTGRESQL_USER
	postgreSqlPassword := configuration.POSTGRESQL_PASS
	postgreSqlDbName := configuration.POSTGRESQL_DBNAME
	postgreSqlHost := configuration.POSTGRESQL_HOST
	postgreSqlPort := configuration.POSTGRESQL_PORT
	var postgreSqlUrl string
	if configuration.POSTGRESQL_TLS_ENABLED {
		postgreSqlUrl = "postgres://" + postgreSqlUser + "@" + postgreSqlHost + ":" + postgreSqlPort + "/" + postgreSqlDbName + "?sslmode=verify-full"
	} else {
		postgreSqlUrl = "postgres://" + postgreSqlUser + ":" + postgreSqlPassword + "@" + postgreSqlHost + ":" + postgreSqlPort + "/" + postgreSqlDbName + "?sslmode=disable"
	}

	config, err := pgxpool.ParseConfig(postgreSqlUrl)
	if err != nil {
		return DbPostgreSQLInstance{}, err
	}
	config.MaxConns = 5

	if configuration.POSTGRESQL_TLS_ENABLED {
		cert, err := tls.LoadX509KeyPair(configuration.POSTGRESQL_TLS_CRT, configuration.POSTGRESQL_TLS_KEY)
		if err != nil {
			return DbPostgreSQLInstance{}, err
		}

		CACert, err := os.ReadFile(configuration.POSTGRESQL_TLS_CA)
		if err != nil {
			return DbPostgreSQLInstance{}, err
		}

		CACertPool := x509.NewCertPool()
		CACertPool.AppendCertsFromPEM(CACert)
		config.ConnConfig.TLSConfig = &tls.Config{Certificates: []tls.Certificate{cert}, RootCAs: CACertPool, InsecureSkipVerify: true}
	}

	dbPostgreSQL, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return DbPostgreSQLInstance{}, err
	}

	err = dbPostgreSQL.Ping(ctx)
	if err != nil {
		return DbPostgreSQLInstance{}, err
	}
	l.Noticef("Established connection with the meta-data storage")
	dbPostgre := DbPostgreSQLInstance{Ctx: ctx, Cancel: cancelfunc, Client: dbPostgreSQL}
	err = createTables(dbPostgre)
	if err != nil {
		return DbPostgreSQLInstance{}, err
	}
	postgresConnection = DbPostgreSQLInstance{Client: dbPostgreSQL, Ctx: ctx, Cancel: cancelfunc}

	_ = InsertToTable(postgresConnection.Client)
	return postgresConnection, nil
}

// System Keys Functions
func GetSystemKey(key string) (bool, models.SystemKey, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), dbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := postgresConnection.Client.Acquire(ctx)
	if err != nil {
		return true, models.SystemKey{}, err
	}
	defer conn.Release()
	query := `SELECT * FROM configurations WHERE key = $1 LIMIT 1`
	stmt, err := conn.Conn().Prepare(ctx, "get_system_key", query)
	if err != nil {
		return true, models.SystemKey{}, err
	}
	rows, err := conn.Conn().Query(ctx, stmt.Name, key)
	if err != nil {
		return true, models.SystemKey{}, err
	}
	defer rows.Close()
	systemKeys, err := pgx.CollectRows(rows, pgx.RowToStructByPos[models.SystemKey])
	if len(systemKeys) == 0 {
		return false, models.SystemKey{}, nil
	}
	if err != nil {
		return true, models.SystemKey{}, err
	}
	return true, systemKeys[0], nil
}

func InsertSystemKey(key string, value string) error {
	// systemKey := models.SystemKey{
	// 	ID:    primitive.NewObjectID(),
	// 	Key:   key,
	// 	Value: value,
	// }
	// _, err := systemKeysCollection.InsertOne(context.TODO(), systemKey)
	// return err
	return nil
}

func InsertSystemKeyPg(key string, stringValue string, intValue int, isString bool) error {
	err := InsertConfiguration(key, stringValue, intValue, isString)
	if err != nil {
		return err
	}
	return nil
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
	// var configurationsIntValue models.ConfigurationsIntValue
	// filter := bson.M{"key": key}
	// if isString {
	// 	err := configurationsCollection.FindOne(context.TODO(), filter).Decode(&configurationsStringValue)
	// 	if err == mongo.ErrNoDocuments {
	// 		return false, models.ConfigurationsStringValue{}, models.ConfigurationsIntValue{}, nil
	// 	}
	// 	if err != nil {
	// 		return true, models.ConfigurationsStringValue{}, models.ConfigurationsIntValue{}, err
	// 	}
	// 	return true, configurationsStringValue, models.ConfigurationsIntValue{}, nil
	// } else {
	// 	err := configurationsCollection.FindOne(context.TODO(), filter).Decode(&configurationsIntValue)
	// 	if err == mongo.ErrNoDocuments {
	// 		return false, models.ConfigurationsStringValue{}, models.ConfigurationsIntValue{}, nil
	// 	}
	// 	if err != nil {
	// 		return true, models.ConfigurationsStringValue{}, models.ConfigurationsIntValue{}, err
	// 	}
	// 	return true, models.ConfigurationsStringValue{}, configurationsIntValue, nil
	// }
	ctx, cancelfunc := context.WithTimeout(context.Background(), dbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := postgresConnection.Client.Acquire(ctx)
	if err != nil {
		return true, models.ConfigurationsStringValue{}, models.ConfigurationsIntValue{}, err
	}
	defer conn.Release()
	query := `SELECT value FROM configurations WHERE key = $1`
	stmt, err := conn.Conn().Prepare(ctx, "get_configuration", query)
	if err != nil {
		return true, models.ConfigurationsStringValue{}, models.ConfigurationsIntValue{}, err
	}
	rows, err := conn.Conn().Query(ctx, stmt.Name, key)
	if err != nil {
		return true, models.ConfigurationsStringValue{}, models.ConfigurationsIntValue{}, err
	}
	defer rows.Close()
	configurations, err := pgx.CollectRows(rows, pgx.RowToStructByPos[models.ConfigurationsStringValue])
	if len(configurations) == 0 {
		return false, models.ConfigurationsStringValue{}, models.ConfigurationsIntValue{}, nil
	}
	if err != nil {
		return true, models.ConfigurationsStringValue{}, models.ConfigurationsIntValue{}, err
	}
	return true, configurations[0], models.ConfigurationsIntValue{}, nil
}

func InsertConfiguration(key string, stringValue string, intValue int, isString bool) error {
	ctx, cancelfunc := context.WithTimeout(context.Background(), dbOperationTimeout*time.Second)
	defer cancelfunc()

	conn, err := postgresConnection.Client.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	if isString {
		query := `INSERT INTO configurations( 
			key, 
			value) 
		VALUES($1, $2) 
		RETURNING id`

		stmt, err := conn.Conn().Prepare(ctx, "insert_new_configuration", query)
		if err != nil {
			return err
		}

		newConfiguration := models.ConfigurationsStringValuePg{}
		rows, err := conn.Conn().Query(ctx, stmt.Name,
			key, stringValue)
		if err != nil {
			return err
		}
		for rows.Next() {
			err := rows.Scan(&newConfiguration.ID)
			if err != nil {
				return err
			}
		}

		if err := rows.Err(); err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) {
				fmt.Println(pgErr.Detail)
				if pgErr.Detail != "" {
					if strings.Contains(pgErr.Detail, "already exists") {
						return errors.New("configuration " + newConfiguration.Key + " already exists")
					} else {
						return errors.New(pgErr.Detail)
					}
				} else {
					return errors.New(pgErr.Message)
				}
			} else {
				return err
			}
		}
	} else {
		query := `INSERT INTO configurations( 
			key, 
			value) 
		VALUES($1, $2) 
		RETURNING id`

		stmt, err := conn.Conn().Prepare(ctx, "insert_new_configuration", query)
		if err != nil {
			return err
		}

		newConfiguration := models.ConfigurationsIntValuePg{}
		rows, err := conn.Conn().Query(ctx, stmt.Name,
			key, stringValue)
		if err != nil {
			return err
		}
		for rows.Next() {
			err := rows.Scan(&newConfiguration.ID)
			if err != nil {
				return err
			}
		}

		if err := rows.Err(); err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) {
				fmt.Println(pgErr.Detail)
				if pgErr.Detail != "" {
					return errors.New(pgErr.Detail)
				} else {
					return errors.New(pgErr.Message)
				}
			} else {
				return err
			}
		}
	}
	return nil
}

func UpsertConfiguration(key string, stringValue string, intValue int, isString bool) error {
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
func InsertConnection(connection models.ConnectionPg) error {
	ctx, cancelfunc := context.WithTimeout(context.Background(), dbOperationTimeout*time.Second)
	defer cancelfunc()

	conn, err := postgresConnection.Client.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	query := `INSERT INTO connections ( 
		id,
		created_by, 
		is_active, 
		created_at,
		client_address) 
    VALUES($1, $2, $3, $4, $5) RETURNING id`

	stmt, err := conn.Conn().Prepare(ctx, "insert_connection", query)
	if err != nil {
		return err
	}

	newConnection := models.ConnectionPg{}
	createdAt := time.Now()

	rows, err := conn.Conn().Query(ctx, stmt.Name, connection.ID,
		connection.CreatedBy, connection.IsActive, createdAt, connection.ClientAddress)
	if err != nil {
		return err
	}
	for rows.Next() {
		err := rows.Scan(&newConnection.ID)
		if err != nil {
			return err
		}
	}

	if err := rows.Err(); err != nil {
		return err
	}
	if err := rows.Err(); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			fmt.Println(pgErr.Detail)
			if pgErr.Detail != "" {
				if strings.Contains(pgErr.Detail, "already exists") {
					return errors.New("connection " + newConnection.ID + " already exists")
				} else {
					return errors.New(pgErr.Detail)
				}
			} else {
				return errors.New(pgErr.Message)
			}
		} else {
			return err
		}
	}
	return nil
}

func UpdateConnection(connectionId string, isActive bool) error {
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

func GetConnectionByID(connectionId string) (bool, models.Connection, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), dbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := postgresConnection.Client.Acquire(ctx)
	if err != nil {
		return true, models.Connection{}, err
	}
	defer conn.Release()
	query := `SELECT * FROM connections AS c WHERE id = $1`
	stmt, err := conn.Conn().Prepare(ctx, "get_connection_by_id", query)
	if err != nil {
		return true, models.Connection{}, err
	}
	rows, err := conn.Conn().Query(ctx, stmt.Name, connectionId)
	if err != nil {
		return true, models.Connection{}, err
	}
	defer rows.Close()
	connections, err := pgx.CollectRows(rows, pgx.RowToStructByPos[models.Connection])
	if len(connections) == 0 {
		return false, models.Connection{}, nil
	}
	if err != nil {
		return true, models.Connection{}, err
	}
	return true, connections[0], nil
}

func KillRelevantConnections(ids []string) error {
	// _, err := connectionsCollection.UpdateMany(context.TODO(),
	// 	bson.M{"_id": bson.M{"$in": ids}},
	// 	bson.M{"$set": bson.M{"is_active": false}},
	// )
	// if err != nil {
	// 	return err
	// }

	return nil
}

func GetActiveConnections() ([]models.Connection, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), dbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := postgresConnection.Client.Acquire(ctx)
	if err != nil {
		return []models.Connection{}, err
	}
	defer conn.Release()
	query := `SELECT connections FROM connections WHERE is_active = true`
	stmt, err := conn.Conn().Prepare(ctx, "get_active_connection", query)
	if err != nil {
		return []models.Connection{}, err
	}
	rows, err := conn.Conn().Query(ctx, stmt.Name)
	if err != nil {
		return []models.Connection{}, err
	}
	defer rows.Close()
	connections, err := pgx.CollectRows(rows, pgx.RowToStructByPos[models.Connection])
	if len(connections) == 0 {
		return []models.Connection{}, nil
	}
	if err != nil {
		return []models.Connection{}, err
	}
	return connections, nil
}

func InsertAuditLogs(auditLogs []interface{}) error {
	ctx, cancelfunc := context.WithTimeout(context.Background(), dbOperationTimeout*time.Second)
	defer cancelfunc()

	conn, err := postgresConnection.Client.Acquire(ctx)
	if err != nil {
		return err
	}

	defer conn.Release()
	var auditLog []models.AuditLogPg

	b, err := json.Marshal(auditLogs)
	err = json.Unmarshal(b, &auditLog)
	if err != nil {
		return err
	}

	stationName := auditLog[0].StationName
	message := auditLog[0].Message
	createdBy := auditLog[0].CreatedBy
	createdAt := auditLog[0].CreatedAt

	query := `INSERT INTO audit_logs ( 
		station_name, 
		message, 
		created_by,
		created_at
		) 
    VALUES($1, $2, $3, $4) RETURNING id`

	stmt, err := conn.Conn().Prepare(ctx, "insert_audit_logs", query)
	if err != nil {
		return err
	}

	newAuditLog := models.AuditLogPg{}
	rows, err := conn.Conn().Query(ctx, stmt.Name,
		stationName, message, createdBy, createdAt)
	if err != nil {
		return err
	}
	for rows.Next() {
		err := rows.Scan(&newAuditLog.ID)
		if err != nil {
			return err
		}
	}

	if err := rows.Err(); err != nil {
		return err
	}

	return nil
}

func GetAuditLogsByStation(name string) ([]models.AuditLog, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), dbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := postgresConnection.Client.Acquire(ctx)
	if err != nil {
		return []models.AuditLog{}, err
	}
	defer conn.Release()
	query := `SELECT * FROM audit_logs AS a WHERE c.station_name = $1`
	stmt, err := conn.Conn().Prepare(ctx, "get_audit_logs_by_station", query)
	if err != nil {
		return []models.AuditLog{}, err
	}
	rows, err := conn.Conn().Query(ctx, stmt.Name, name)
	if err != nil {
		return []models.AuditLog{}, err
	}
	defer rows.Close()
	auditLogs, err := pgx.CollectRows(rows, pgx.RowToStructByPos[models.AuditLog])
	if len(auditLogs) == 0 {
		return []models.AuditLog{}, nil
	}
	if err != nil {
		return []models.AuditLog{}, err
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
	ctx, cancelfunc := context.WithTimeout(context.Background(), dbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := postgresConnection.Client.Acquire(ctx)
	if err != nil {
		return []models.Station{}, err
	}
	defer conn.Release()
	query := `SELECT stations FROM stations AS s WHERE s.is_deleted = false OR s.is_deleted IS NULL`
	stmt, err := conn.Conn().Prepare(ctx, "get_active_stations", query)
	if err != nil {
		return []models.Station{}, err
	}
	rows, err := conn.Conn().Query(ctx, stmt.Name)
	if err != nil {
		return []models.Station{}, err
	}
	defer rows.Close()
	stations, err := pgx.CollectRows(rows, pgx.RowToStructByPos[models.Station])
	if err == pgx.ErrNoRows {
		return []models.Station{}, nil
	}
	if err != nil {
		return []models.Station{}, err
	}
	return stations, nil
}

func GetStationByName(name string) (bool, models.Station, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), dbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := postgresConnection.Client.Acquire(ctx)
	if err != nil {
		return true, models.Station{}, err
	}
	defer conn.Release()
	query := `SELECT * FROM stations WHERE name = $1 AND (is_deleted = false OR is_deleted IS NULL)`
	stmt, err := conn.Conn().Prepare(ctx, "get_station_by_name", query)
	if err != nil {
		return true, models.Station{}, err
	}
	rows, err := conn.Conn().Query(ctx, stmt.Name, name)
	if err != nil {
		return true, models.Station{}, err
	}
	defer rows.Close()
	stations, err := pgx.CollectRows(rows, pgx.RowToStructByPos[models.Station])
	if len(stations) == 0 {
		return false, models.Station{}, nil
	}
	if err != nil {
		return true, models.Station{}, err
	}
	return true, stations[0], nil
}

// TODO: username should be int
func UpsertNewStation(
	stationName string,
	userId int,
	retentionType string,
	retentionValue int,
	storageType string,
	replicas int,
	schemaDetails models.SchemaDetails,
	idempotencyWindow int64,
	isNative bool,
	dlsConfiguration models.DlsConfiguration,
	tieredStorageEnabled bool) (models.Station, int64, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), dbOperationTimeout*time.Second)
	defer cancelfunc()

	conn, err := postgresConnection.Client.Acquire(ctx)
	if err != nil {
		return models.Station{}, 0, err
	}
	defer conn.Release()

	query := `INSERT INTO stations ( 
		name, 
		retention_type, 
		retention_value,
		storage_type, 
		replicas, 
		created_by, 
		created_at, 
		updated_at, 
		is_deleted, 
		schema_name,
		schema_version_number,
		idempotency_window_ms, 
		is_native, 
		dls_configuration_poison, 
		dls_configuration_schemaverse,
		tiered_storage_enabled
		) 
    VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16) RETURNING id`

	stmt, err := conn.Conn().Prepare(ctx, "upsert_new_station", query)
	if err != nil {
		return models.Station{}, 0, err
	}

	// newStation := models.StationPg{}
	createAt := time.Now()
	updatedAt := time.Now()
	newStation := models.Station{
		Name:                        stationName,
		CreatedBy:                   userId,
		CreatedAt:                   createAt,
		IsDeleted:                   false,
		RetentionType:               retentionType,
		RetentionValue:              retentionValue,
		StorageType:                 storageType,
		Replicas:                    replicas,
		UpdatedAt:                   updatedAt,
		SchemaName:                  schemaDetails.SchemaName,
		SchemaVersionNumber:         schemaDetails.VersionNumber,
		IdempotencyWindow:           idempotencyWindow,
		IsNative:                    isNative,
		DlsConfigurationPoison:      dlsConfiguration.Poison,
		DlsConfigurationSchemaverse: dlsConfiguration.Schemaverse,
		TieredStorageEnabled:        tieredStorageEnabled,
	}

	//TODO: change the 1 to username
	rows, err := conn.Conn().Query(ctx, stmt.Name,
		stationName, retentionType, retentionValue, storageType, replicas, 1, createAt, updatedAt,
		false, schemaDetails.SchemaName, schemaDetails.VersionNumber, idempotencyWindow, isNative, dlsConfiguration.Poison, dlsConfiguration.Schemaverse, tieredStorageEnabled)
	if err != nil {
		return models.Station{}, 0, err
	}
	for rows.Next() {
		err := rows.Scan(&newStation.ID)
		if err != nil {
			return models.Station{}, 0, err
		}
	}

	if err := rows.Err(); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			fmt.Println(pgErr.Detail)
			if pgErr.Detail != "" {
				if strings.Contains(pgErr.Detail, "already exists") {
					return models.Station{}, 0, errors.New("Station" + newStation.Name + " already exists")
				} else {
					return models.Station{}, 0, errors.New(pgErr.Detail)
				}
			} else {
				return models.Station{}, 0, errors.New(pgErr.Message)
			}
		} else {
			return models.Station{}, 0, err
		}
	}

	rowsAffected := rows.CommandTag().RowsAffected()
	return newStation, rowsAffected, nil
}

// TODO: Aggregate
func GetAllStationsDetails() ([]models.ExtendedStation, error) {
	var stations []models.ExtendedStation
	// cursor, err := stationsCollection.Aggregate(context.TODO(), mongo.Pipeline{
	// 	bson.D{{"$match", bson.D{{"$or", []interface{}{
	// 		bson.D{{"is_deleted", false}},
	// 		bson.D{{"is_deleted", bson.D{{"$exists", false}}}},
	// 	}}}}},
	// 	bson.D{{"$lookup", bson.D{{"from", "producers"}, {"localField", "_id"}, {"foreignField", "station_id"}, {"as", "producers"}}}},
	// 	bson.D{{"$lookup", bson.D{{"from", "consumers"}, {"localField", "_id"}, {"foreignField", "station_id"}, {"as", "consumers"}}}},
	// 	bson.D{{"$project", bson.D{{"_id", 1}, {"name", 1}, {"retention_type", 1}, {"retention_value", 1}, {"storage_type", 1}, {"replicas", 1}, {"idempotency_window_in_ms", 1}, {"created_by_user", 1}, {"creation_date", 1}, {"last_update", 1}, {"dls_configuration", 1}, {"is_native", 1}, {"producers", 1}, {"consumers", 1}, {"tiered_storage_enabled", 1}}}},
	// })
	// if err == mongo.ErrNoDocuments {
	// 	return []models.ExtendedStation{}, nil
	// }
	// if err != nil {
	// 	return []models.ExtendedStation{}, err
	// }

	// if err = cursor.All(context.TODO(), &stations); err != nil {
	// 	return []models.ExtendedStation{}, err
	// }
	// return stations, nil

	ctx, cancelfunc := context.WithTimeout(context.Background(), dbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := postgresConnection.Client.Acquire(ctx)
	if err != nil {
		return []models.ExtendedStation{}, err
	}
	defer conn.Release()
	query := `
	SELECT stations
	FROM stations AS s
	LEFT JOIN producers AS p
	ON s._id = p.station_id
	LEFT JOIN consumers AS c
	ON s._id = c.station_id
	WHERE s.is_deleted = false OR s.is_deleted IS NULL
	GROUP BY
		s._id, s.name, s.retention_type, s.retention_value, s.storage_type, s.replicas,
		s.idempotency_window_in_ms, s.created_by_user, s.creation_date, s.last_update,
		s.dls_configuration, s.is_native, s.tiered_storage_enabled`

	stmt, err := conn.Conn().Prepare(ctx, "get_all_stations_details", query)
	if err != nil {
		return []models.ExtendedStation{}, err
	}

	rows, err := conn.Conn().Query(ctx, stmt.Name)
	if err != nil {
		return []models.ExtendedStation{}, err
	}
	defer rows.Close()
	for rows.Next() {
		var station models.ExtendedStation
		var consumers []models.Consumer
		var producers []models.Producer
		if err := rows.Scan(
			&station,
			&producers,
			&consumers,
		); err != nil {
			return []models.ExtendedStation{}, err
		}
		station.Consumers = consumers
		station.Producers = producers
		stations = append(stations, station)
	}
	if err := rows.Err(); err != nil {
		return []models.ExtendedStation{}, err
	}

	return stations, nil

}

func DeleteStationsByNames(stationNames []string) error {
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

func UpsertStationDlsConfig(stationName string, dlsConfiguration models.DlsConfiguration) error {
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

// TODO: Aggregate
func GetStationNamesUsingSchema(schemaName string) ([]string, error) {
	// var stations []models.Station
	// cursor, err := stationsCollection.Aggregate(context.TODO(), mongo.Pipeline{
	// 	bson.D{{"$unwind", bson.D{{"path", "$schema"}, {"preserveNullAndEmptyArrays", true}}}},
	// 	bson.D{{"$match", bson.D{{"schema.name", schemaName}, {"is_deleted", false}}}},
	// 	bson.D{{"$project", bson.D{{"name", 1}}}},
	// })
	// if err != nil {
	// 	return []string{}, err
	// }

	// if err = cursor.All(context.TODO(), &stations); err != nil {
	// 	return []string{}, err
	// }
	// if len(stations) == 0 {
	// 	return []string{}, nil
	// }

	// for _, station := range stations {
	// 	stationNames = append(stationNames, station.Name)
	// }

	// return stationNames, nil

	var stationNames []string
	ctx, cancelfunc := context.WithTimeout(context.Background(), dbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := postgresConnection.Client.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Release()
	query := `
		SELECT name FROM stations
		WHERE schema_name = $1 AND is_deleted = false
	`
	stmt, err := conn.Conn().Prepare(ctx, "get_station_names_using_schema", query)
	if err != nil {
		return nil, err
	}
	rows, err := conn.Conn().Query(ctx, stmt.Name, schemaName)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var stationName string
		err := rows.Scan(&stationName)
		if err != nil {
			return nil, err
		}
		stationNames = append(stationNames, stationName)
	}

	if rows.Err() != nil {
		return nil, err
	}

	return stationNames, nil

}

func GetCountStationsUsingSchema(schemaName string) (int, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), dbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := postgresConnection.Client.Acquire(ctx)
	if err != nil {
		return 0, err
	}
	defer conn.Release()
	query := `SELECT COUNT(*) FROM stations WHERE schema_name = $1 AND is_deleted = false`
	stmt, err := conn.Conn().Prepare(ctx, "get_count_stations_using_schema", query)
	if err != nil {
		return 0, err
	}
	var count int
	err = conn.Conn().QueryRow(ctx, stmt.Name, schemaName).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
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

// TODO: Aggregate
func GetProducersByConnectionIDWithStationDetails(connectionId string) ([]models.ExtendedProducer, error) {
	var producers []models.ExtendedProducer
	// cursor, err := producersCollection.Aggregate(context.TODO(), mongo.Pipeline{
	// 	bson.D{{"$match", bson.D{{"connection_id", connectionId}, {"is_active", true}}}},
	// 	bson.D{{"$lookup", bson.D{{"from", "stations"}, {"localField", "station_id"}, {"foreignField", "_id"}, {"as", "station"}}}},
	// 	bson.D{{"$unwind", bson.D{{"path", "$station"}, {"preserveNullAndEmptyArrays", true}}}},
	// 	bson.D{{"$project", bson.D{{"_id", 1}, {"name", 1}, {"type", 1}, {"connection_id", 1}, {"created_by_user", 1}, {"creation_date", 1}, {"is_active", 1}, {"is_deleted", 1}, {"station_name", "$station.name"}}}},
	// 	bson.D{{"$project", bson.D{{"station", 0}, {"connection", 0}}}},
	// })
	// if err != nil {
	// 	return []models.ExtendedProducer{}, err
	// }
	// if err = cursor.All(context.TODO(), &producers); err != nil {
	// 	return []models.ExtendedProducer{}, err
	// }
	// return producers, nil

	ctx, cancelfunc := context.WithTimeout(context.Background(), dbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := postgresConnection.Client.Acquire(ctx)
	if err != nil {
		return []models.ExtendedProducer{}, err
	}
	defer conn.Release()
	query := `
	SELECT producers, s.name AS station_name
	FROM producers p
	LEFT JOIN stations s ON p.station_id = s._id
	WHERE p.connection_id = $1 AND p.is_active = true`
	stmt, err := conn.Conn().Prepare(ctx, "get_producers_by_connection_id_with_station_details", query)
	if err != nil {
		return []models.ExtendedProducer{}, err
	}
	rows, err := conn.Conn().Query(ctx, stmt.Name, connectionId)
	if err != nil {
		return []models.ExtendedProducer{}, err
	}
	defer rows.Close()
	for rows.Next() {
		var producer models.ExtendedProducer
		err := rows.Scan(&producer)
		if err != nil {
			return []models.ExtendedProducer{}, err
		}
		producers = append(producers, producer)
	}
	if err := rows.Err(); err != nil {
		return []models.ExtendedProducer{}, err
	}

	return producers, nil

}

func UpdateProducersConnection(connectionId string, isActive bool) error {
	_, err := producersCollection.UpdateMany(context.TODO(),
		bson.M{"connection_id": connectionId},
		bson.M{"$set": bson.M{"is_active": isActive}},
	)
	if err != nil {
		return err
	}
	return nil
}

func GetProducerByNameAndConnectionID(name string, connectionId string) (bool, models.Producer, error) {
	var producer models.Producer
	// filter := bson.M{"name": name, "connection_id": connectionId}
	// err := producersCollection.FindOne(context.TODO(), filter).Decode(&producer)
	// if err == mongo.ErrNoDocuments {
	// 	return false, models.Producer{}, err
	// }
	// if err != nil {
	// 	return true, models.Producer{}, err
	// }
	// return true, producer, nil

	ctx, cancelfunc := context.WithTimeout(context.Background(), dbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := postgresConnection.Client.Acquire(ctx)
	if err != nil {
		return true, models.Producer{}, err
	}
	defer conn.Release()
	query := `SELECT * FROM producers WHERE name = $1 AND connection_id = $2`
	stmt, err := conn.Conn().Prepare(ctx, "get_producer_by_name_and_connection_id", query)
	if err != nil {
		return true, models.Producer{}, err
	}
	err = conn.QueryRow(ctx, stmt.Name, name, connectionId).Scan(&producer)
	if err == pgx.ErrNoRows {
		return false, models.Producer{}, nil
	}
	if err != nil {
		return true, models.Producer{}, err
	}

	return true, producer, nil
}

func GetProducerByStationIDAndUsername(username string, stationId int, connectionId string) (bool, models.Producer, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), dbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := postgresConnection.Client.Acquire(ctx)
	if err != nil {
		return true, models.Producer{}, err
	}
	defer conn.Release()
	query := `SELECT * FROM producers WHERE name = $1 AND station_id = $2 AND connection_id = $3 LIMIT 1`
	stmt, err := conn.Conn().Prepare(ctx, "get_producer_by_station_id_and_username", query)
	if err != nil {
		return true, models.Producer{}, err
	}
	rows, err := conn.Conn().Query(ctx, stmt.Name, username, stationId, connectionId)
	if err != nil {
		return true, models.Producer{}, err
	}
	defer rows.Close()
	producers, err := pgx.CollectRows(rows, pgx.RowToStructByPos[models.Producer])
	if len(producers) == 0 {
		return false, models.Producer{}, nil
	}
	if err != nil {
		return true, models.Producer{}, err
	}
	return true, producers[0], nil
}

func GetActiveProducerByStationID(producerName string, stationId int) (bool, models.Producer, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), dbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := postgresConnection.Client.Acquire(ctx)
	if err != nil {
		return true, models.Producer{}, err
	}
	defer conn.Release()

	query := `SELECT * FROM producers WHERE name = $1 AND station_id = $2 AND is_active = true`
	stmt, err := conn.Conn().Prepare(ctx, "get_active_producer_by_station_id", query)
	if err != nil {
		return true, models.Producer{}, err
	}
	rows, err := conn.Conn().Query(ctx, stmt.Name, producerName, stationId)
	if err != nil {
		return true, models.Producer{}, err
	}
	defer rows.Close()
	producers, err := pgx.CollectRows(rows, pgx.RowToStructByPos[models.Producer])
	if len(producers) == 0 {
		return false, models.Producer{}, nil
	}
	if err != nil {
		return true, models.Producer{}, err
	}
	return true, producers[0], nil
}

func UpsertNewProducer(name string, stationId int, producerType string, connectionIdObj string, createdByUser int) (models.Producer, int64, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), dbOperationTimeout*time.Second)
	defer cancelfunc()

	conn, err := postgresConnection.Client.Acquire(ctx)
	if err != nil {
		return models.Producer{}, 0, err
	}
	defer conn.Release()

	query := `INSERT INTO producers ( 
		name, 
		station_id, 
		connection_id,
		created_by, 
		is_active, 
		is_deleted, 
		created_at, 
		type) 
    VALUES($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id`

	stmt, err := conn.Conn().Prepare(ctx, "upsert_new_producer", query)
	if err != nil {
		return models.Producer{}, 0, err
	}

	newProducer := models.Producer{}
	createAt := time.Now()
	isActive := true
	isDeleted := false

	rows, err := conn.Conn().Query(ctx, stmt.Name, name, stationId, connectionIdObj, createdByUser, isActive, isDeleted, createAt, producerType)
	if err != nil {
		return models.Producer{}, 0, err
	}
	for rows.Next() {
		err := rows.Scan(&newProducer.ID)
		if err != nil {
			return models.Producer{}, 0, err
		}
	}

	if err := rows.Err(); err != nil {
		return models.Producer{}, 0, err
	}

	// if err := rows.Err(); err != nil {
	// 	var pgErr *pgconn.PgError
	// 	if errors.As(err, &pgErr) {
	// 		fmt.Println(pgErr.Detail)
	// 		if pgErr.Detail != "" {
	// 			return models.Schema{}, 0, errors.New(pgErr.Detail)
	// 		} else {
	// 			return models.Schema{}, 0, errors.New(pgErr.Message)
	// 		}
	// 		// => syntax error at end of input
	// 		// fmt.Println(pgErr.Code)    // => 42601
	// 	}else{
	// 		return models.Schema{}, 0, err
	// 	}
	// }

	rowsAffected := rows.CommandTag().RowsAffected()
	newProducer = models.Producer{
		ID:           newProducer.ID,
		Name:         name,
		StationId:    stationId,
		Type:         producerType,
		ConnectionId: connectionIdObj,
		CreatedBy:    createdByUser,
		IsActive:     isActive,
		CreatedAt:    time.Now(),
		IsDeleted:    isDeleted,
	}
	return newProducer, rowsAffected, nil
}

// TODO: Aggregate
func GetAllProducers() ([]models.ExtendedProducer, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), dbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := postgresConnection.Client.Acquire(ctx)
	if err != nil {
		return []models.ExtendedProducer{}, err
	}
	defer conn.Release()
	query := `
		SELECT p._id, p.name, p.type, p.connection_id, p.created_by_user, p.creation_date, p.is_active, p.is_deleted, s.name AS station_name, c.client_address AS client_address
		FROM producers p
		LEFT JOIN stations ON p.station_id = s._id
		LEFT JOIN connections ON p.connection_id = c._id
	`
	stmt, err := conn.Conn().Prepare(ctx, "get_all_producers", query)
	if err != nil {
		return []models.ExtendedProducer{}, err
	}
	rows, err := conn.Conn().Query(ctx, stmt.Name)
	if err != nil {
		return []models.ExtendedProducer{}, err
	}
	defer rows.Close()
	producers, err := pgx.CollectRows(rows, pgx.RowToStructByPos[models.ExtendedProducer])
	if len(producers) == 0 {
		return []models.ExtendedProducer{}, nil
	}
	if err != nil {
		return []models.ExtendedProducer{}, err
	}
	return producers, nil
}

// TODO: Aggregate
func GetProducersByStationID(stationId int) ([]models.ExtendedProducer, error) {
	var producers []models.ExtendedProducer

	// cursor, err := producersCollection.Aggregate(context.TODO(), mongo.Pipeline{
	// 	bson.D{{"$match", bson.D{{"station_id", stationId}}}},
	// 	bson.D{{"$sort", bson.D{{"creation_date", -1}}}},
	// 	bson.D{{"$lookup", bson.D{{"from", "stations"}, {"localField", "station_id"}, {"foreignField", "_id"}, {"as", "station"}}}},
	// 	bson.D{{"$unwind", bson.D{{"path", "$station"}, {"preserveNullAndEmptyArrays", true}}}},
	// 	bson.D{{"$lookup", bson.D{{"from", "connections"}, {"localField", "connection_id"}, {"foreignField", "_id"}, {"as", "connection"}}}},
	// 	bson.D{{"$unwind", bson.D{{"path", "$connection"}, {"preserveNullAndEmptyArrays", true}}}},
	// 	bson.D{{"$project", bson.D{{"_id", 1}, {"name", 1}, {"type", 1}, {"connection_id", 1}, {"created_by_user", 1}, {"creation_date", 1}, {"is_active", 1}, {"is_deleted", 1}, {"station_name", "$station.name"}, {"client_address", "$connection.client_address"}}}},
	// 	bson.D{{"$project", bson.D{{"station", 0}, {"connection", 0}}}},
	// })

	// if err != nil {
	// 	return []models.ExtendedProducer{}, err
	// }

	// if err = cursor.All(context.TODO(), &producers); err != nil {
	// 	return []models.ExtendedProducer{}, err
	// }
	// return producers, nil

	ctx, cancelfunc := context.WithTimeout(context.Background(), dbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := postgresConnection.Client.Acquire(ctx)
	if err != nil {
		return []models.ExtendedProducer{}, err
	}
	defer conn.Release()
	query := `
		SELECT
			p._id,
			p.name,
			p.type,
			p.connection_id,
			p.created_by_user,
			p.creation_date,
			p.is_active,
			p.is_deleted,
			s.name AS station_name,
			c.client_address AS client_address
		FROM
			producers p
			LEFT JOIN stations s ON p.station_id = s._id
			LEFT JOIN connections c ON p.connection_id = c._id
		WHERE
			p.station_id = $1
		ORDER BY
			p.creation_date DESC
	`
	stmt, err := conn.Conn().Prepare(ctx, "get_producers_by_station_id", query)
	if err != nil {
		return []models.ExtendedProducer{}, err
	}
	rows, err := conn.Conn().Query(context.Background(), stmt.Name, stationId)
	if err != nil {
		return []models.ExtendedProducer{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var producer models.ExtendedProducer

		err = rows.Scan(&producer)
		if err != nil {
			return []models.ExtendedProducer{}, err
		}

		producers = append(producers, producer)
	}

	if err = rows.Err(); err != nil {
		return []models.ExtendedProducer{}, err
	}

	return producers, nil

}

func DeleteProducerByNameAndStationID(name string, stationId int) (bool, models.Producer, error) {
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

func DeleteProducersByStationID(stationId int) error {
	_, err := producersCollection.UpdateMany(context.TODO(),
		bson.M{"station_id": stationId},
		bson.M{"$set": bson.M{"is_active": false, "is_deleted": true}},
	)
	if err != nil {
		return err
	}
	return nil
}

func CountActiveProudcersByStationID(stationId int) (int64, error) {
	var activeCount int64
	ctx, cancelfunc := context.WithTimeout(context.Background(), dbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := postgresConnection.Client.Acquire(ctx)
	if err != nil {
		return 0, err
	}
	defer conn.Release()
	query := `SELECT COUNT(*) FROM producers WHERE station_id = $1 AND is_active = true`
	stmt, err := conn.Conn().Prepare(ctx, "count_active_producers_by_station_id", query)
	if err != nil {
		return 0, err
	}
	err = conn.Conn().QueryRow(ctx, stmt.Name, stationId).Scan(&activeCount)
	if err != nil {
		return 0, err
	}

	return activeCount, nil
}

func CountAllActiveProudcers() (int64, error) {
	var producersCount int64
	ctx, cancelfunc := context.WithTimeout(context.Background(), dbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := postgresConnection.Client.Acquire(ctx)
	if err != nil {
		return 0, err
	}
	defer conn.Release()
	query := `SELECT COUNT(*) FROM producers WHERE is_active = true`
	stmt, err := conn.Conn().Prepare(ctx, "count_all_active_producers", query)
	if err != nil {
		return 0, err
	}
	err = conn.Conn().QueryRow(ctx, stmt.Name).Scan(&producersCount)
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

func KillProducersByConnections(connectionIds []string) error {
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
func GetActiveConsumerByCG(consumersGroup string, stationId int) (bool, models.Consumer, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), dbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := postgresConnection.Client.Acquire(ctx)
	if err != nil {
		return true, models.Consumer{}, err
	}
	defer conn.Release()

	query := `SELECT * FROM consumers WHERE consumers_group = $1 AND station_id = $2 AND is_deleted = false LIMIT 1`
	stmt, err := conn.Conn().Prepare(ctx, "get_active_consumer_by_cg", query)
	if err != nil {
		return true, models.Consumer{}, err
	}
	rows, err := conn.Conn().Query(ctx, stmt.Name, consumersGroup, stationId)
	if err != nil {
		return true, models.Consumer{}, err
	}
	defer rows.Close()
	consumers, err := pgx.CollectRows(rows, pgx.RowToStructByPos[models.Consumer])
	if len(consumers) == 0 {
		return false, models.Consumer{}, nil
	}
	if err != nil {
		return true, models.Consumer{}, err
	}
	return true, consumers[0], nil
}

func UpsertNewConsumer(name string,
	stationId int,
	consumerType string,
	connectionIdObj string,
	createdByUser int,
	cgName string,
	maxAckTime int,
	maxMsgDeliveries int,
	startConsumeFromSequence uint64,
	lastMessages int64) (models.Consumer, int64, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), dbOperationTimeout*time.Second)
	defer cancelfunc()

	conn, err := postgresConnection.Client.Acquire(ctx)
	if err != nil {
		return models.Consumer{}, 0, err
	}
	defer conn.Release()

	query := `INSERT INTO consumers ( 
		name, 
		station_id,
		connection_id,
		consumers_group,
		max_ack_time_ms,
		created_by,
		is_active, 
		is_deleted, 
		created_at,
		max_msg_deliveries,
		start_consume_from_seq,
		last_msgs,
		type) 
    VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13) 
	RETURNING id`

	stmt, err := conn.Conn().Prepare(ctx, "upsert_new_consumer", query)
	if err != nil {
		return models.Consumer{}, 0, err
	}

	newConsumer := models.Consumer{}
	createdAt := time.Now()
	isActive := true
	isDeleted := false

	rows, err := conn.Conn().Query(ctx, stmt.Name,
		name, stationId, connectionIdObj, cgName, maxAckTime, createdByUser, isActive, isDeleted, createdAt, maxMsgDeliveries, startConsumeFromSequence, lastMessages, consumerType)
	if err != nil {
		return models.Consumer{}, 0, err
	}
	for rows.Next() {
		err := rows.Scan(&newConsumer.ID)
		if err != nil {
			return models.Consumer{}, 0, err
		}
	}

	if err := rows.Err(); err != nil {
		return models.Consumer{}, 0, err
	}

	// if err := rows.Err(); err != nil {
	// 	var pgErr *pgconn.PgError
	// 	if errors.As(err, &pgErr) {
	// 		fmt.Println(pgErr.Detail)
	// 		if pgErr.Detail != "" {
	// 			return models.Schema{}, 0, errors.New(pgErr.Detail)
	// 		} else {
	// 			return models.Schema{}, 0, errors.New(pgErr.Message)
	// 		}
	// 		// => syntax error at end of input
	// 		// fmt.Println(pgErr.Code)    // => 42601
	// 	}else{
	// 		return models.Schema{}, 0, err
	// 	}
	// }

	rowsAffected := rows.CommandTag().RowsAffected()
	newConsumer = models.Consumer{
		ID:                  newConsumer.ID,
		Name:                name,
		StationId:           stationId,
		Type:                consumerType,
		ConnectionId:        connectionIdObj,
		CreatedBy:           createdByUser,
		ConsumersGroup:      cgName,
		IsActive:            isActive,
		CreatedAt:           time.Now(),
		IsDeleted:           isDeleted,
		MaxAckTimeMs:        int64(maxAckTime),
		MaxMsgDeliveries:    maxMsgDeliveries,
		StartConsumeFromSeq: startConsumeFromSequence,
		LastMessages:        lastMessages,
	}
	return newConsumer, rowsAffected, nil
}

// TODO: Aggregate
func GetAllConsumers() ([]models.ExtendedConsumer, error) {
	var consumers []models.ExtendedConsumer
	// cursor, err := consumersCollection.Aggregate(context.TODO(), mongo.Pipeline{
	// 	bson.D{{"$match", bson.D{}}},
	// 	bson.D{{"$lookup", bson.D{{"from", "stations"}, {"localField", "station_id"}, {"foreignField", "_id"}, {"as", "station"}}}},
	// 	bson.D{{"$unwind", bson.D{{"path", "$station"}, {"preserveNullAndEmptyArrays", true}}}},
	// 	bson.D{{"$lookup", bson.D{{"from", "connections"}, {"localField", "connection_id"}, {"foreignField", "_id"}, {"as", "connection"}}}},
	// 	bson.D{{"$unwind", bson.D{{"path", "$connection"}, {"preserveNullAndEmptyArrays", true}}}},
	// 	bson.D{{"$project", bson.D{{"_id", 1}, {"name", 1}, {"type", 1}, {"connection_id", 1}, {"created_by_user", 1}, {"consumers_group", 1}, {"creation_date", 1}, {"is_active", 1}, {"is_deleted", 1}, {"max_ack_time_ms", 1}, {"max_msg_deliveries", 1}, {"station_name", "$station.name"}, {"client_address", "$connection.client_address"}}}},
	// })
	// if err != nil {
	// 	return []models.ExtendedConsumer{}, err
	// }
	// if err = cursor.All(context.TODO(), &consumers); err != nil {
	// 	return []models.ExtendedConsumer{}, err
	// }
	// return consumers, nil

	ctx, cancelfunc := context.WithTimeout(context.Background(), dbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := postgresConnection.Client.Acquire(ctx)
	if err != nil {
		return []models.ExtendedConsumer{}, err
	}
	query := `
		SELECT c._id, c.name, c.type, c.connection_id, c.created_by_user, c.consumers_group, c.creation_date,
			   c.is_active, c.is_deleted, c.max_ack_time_ms, c.max_msg_deliveries, s.name as station_name,
			   con.client_address as client_address
		FROM consumers c
		LEFT JOIN stations s ON c.station_id = s._id
		LEFT JOIN connections con ON c.connection_id = con._id
	`
	stmt, err := conn.Conn().Prepare(ctx, "get_all_consumers", query)
	if err != nil {
		return []models.ExtendedConsumer{}, err
	}
	rows, err := conn.Conn().Query(context.Background(), stmt.Name)
	if err != nil {
		return []models.ExtendedConsumer{}, err
	}
	defer rows.Close()
	for rows.Next() {
		var consumer models.ExtendedConsumer
		err := rows.Scan(&consumer)
		if err != nil {
			return []models.ExtendedConsumer{}, err
		}
		consumers = append(consumers, consumer)
	}
	if err = rows.Err(); err != nil {
		return []models.ExtendedConsumer{}, err
	}

	return consumers, nil
}

// TODO: Aggregate
func GetAllConsumersByStation(stationId int) ([]models.ExtendedConsumer, error) {
	var consumers []models.ExtendedConsumer
	// cursor, err := consumersCollection.Aggregate(context.TODO(), mongo.Pipeline{
	// 	bson.D{{"$match", bson.D{{"station_id", stationId}}}},
	// 	bson.D{{"$lookup", bson.D{{"from", "stations"}, {"localField", "station_id"}, {"foreignField", "_id"}, {"as", "station"}}}},
	// 	bson.D{{"$unwind", bson.D{{"path", "$station"}, {"preserveNullAndEmptyArrays", true}}}},
	// 	bson.D{{"$lookup", bson.D{{"from", "connections"}, {"localField", "connection_id"}, {"foreignField", "_id"}, {"as", "connection"}}}},
	// 	bson.D{{"$unwind", bson.D{{"path", "$connection"}, {"preserveNullAndEmptyArrays", true}}}},
	// 	bson.D{{"$project", bson.D{{"_id", 1}, {"name", 1}, {"type", 1}, {"connection_id", 1}, {"created_by_user", 1}, {"consumers_group", 1}, {"creation_date", 1}, {"is_active", 1}, {"is_deleted", 1}, {"max_ack_time_ms", 1}, {"max_msg_deliveries", 1}, {"station_name", "$station.name"}, {"client_address", "$connection.client_address"}}}},
	// 	bson.D{{"$project", bson.D{{"station", 0}, {"connection", 0}}}},
	// })
	// if err != nil {
	// 	return []models.ExtendedConsumer{}, err
	// }

	// if err = cursor.All(context.TODO(), &consumers); err != nil {
	// 	return []models.ExtendedConsumer{}, err
	// }
	// return consumers, nil

	ctx, cancelfunc := context.WithTimeout(context.Background(), dbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := postgresConnection.Client.Acquire(ctx)
	if err != nil {
		return []models.ExtendedConsumer{}, err
	}
	query := `
		SELECT
			c._id,
			c.name,
			c.type,
			c.connection_id,
			c.created_by_user,
			c.consumers_group,
			c.creation_date,
			c.is_active,
			c.is_deleted,
			c.max_ack_time_ms,
			c.max_msg_deliveries,
			s.name as station_name,
			con.client_address as client_address
		FROM
			consumers c
			LEFT JOIN stations s ON s._id = c.station_id
			LEFT JOIN connections con ON con._id = c.connection_id
		WHERE
			c.station_id = $1
	`

	stmt, err := conn.Conn().Prepare(ctx, "get_all_consumers_by_station", query)
	if err != nil {
		return []models.ExtendedConsumer{}, err
	}
	rows, err := conn.Conn().Query(context.Background(), stmt.Name, stationId)
	if err != nil {
		return []models.ExtendedConsumer{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var consumer models.ExtendedConsumer
		err := rows.Scan(&consumer)
		if err != nil {
			return []models.ExtendedConsumer{}, err
		}
		consumers = append(consumers, consumer)
	}
	if err = rows.Err(); err != nil {
		return []models.ExtendedConsumer{}, err
	}
	return consumers, nil

}

func DeleteConsumer(name string, stationId int) (bool, models.Consumer, error) {
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

func DeleteConsumersByStationID(stationId int) error {
	_, err := consumersCollection.UpdateMany(context.TODO(),
		bson.M{"station_id": stationId},
		bson.M{"$set": bson.M{"is_active": false, "is_deleted": true}},
	)
	if err != nil {
		return err
	}
	return nil
}

func CountActiveConsumersInCG(consumersGroup string, stationId int) (int64, error) {
	var count int64
	ctx, cancelfunc := context.WithTimeout(context.Background(), dbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := postgresConnection.Client.Acquire(ctx)
	if err != nil {
		return 0, err
	}
	defer conn.Release()
	query := `SELECT COUNT(*) FROM consumers WHERE station_id = $1 AND consumers_group = $2 AND is_deleted = false`
	stmt, err := conn.Conn().Prepare(ctx, "count_active_consumers_in_cg", query)
	if err != nil {
		return 0, err
	}
	err = conn.Conn().QueryRow(ctx, stmt.Name, stationId, consumersGroup).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil

}

func CountActiveConsumersByStationID(stationId int) (int64, error) {
	var activeCount int64
	ctx, cancelfunc := context.WithTimeout(context.Background(), dbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := postgresConnection.Client.Acquire(ctx)
	if err != nil {
		return 0, err
	}
	defer conn.Release()
	query := `SELECT COUNT(*) FROM consumers WHERE station_id = $1 AND is_active = true`
	stmt, err := conn.Conn().Prepare(ctx, "count_active_consumers_by_station_id", query)
	if err != nil {
		return 0, err
	}
	err = conn.Conn().QueryRow(ctx, stmt.Name, stationId).Scan(&activeCount)
	if err != nil {
		return 0, err
	}

	return activeCount, nil
}

func CountAllActiveConsumers() (int64, error) {
	var consumersCount int64
	ctx, cancelfunc := context.WithTimeout(context.Background(), dbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := postgresConnection.Client.Acquire(ctx)
	if err != nil {
		return 0, err
	}
	defer conn.Release()
	query := `SELECT COUNT(*) FROM consumers WHERE is_active = true`
	stmt, err := conn.Conn().Prepare(ctx, "count_all_active_producers", query)
	if err != nil {
		return 0, err
	}
	err = conn.Conn().QueryRow(ctx, stmt.Name).Scan(&consumersCount)
	if err != nil {
		return 0, err
	}

	return consumersCount, nil
}

// TODO: Aggregate
func GetConsumerGroupMembers(cgName string, stationId int) ([]models.CgMember, error) {
	var consumers []models.CgMember

	// cursor, err := consumersCollection.Aggregate(context.TODO(), mongo.Pipeline{
	// 	bson.D{{"$match", bson.D{{"consumers_group", cgName}, {"station_id", stationId}}}},
	// 	bson.D{{"$sort", bson.D{{"creation_date", -1}}}},
	// 	bson.D{{"$lookup", bson.D{{"from", "connections"}, {"localField", "connection_id"}, {"foreignField", "_id"}, {"as", "connection"}}}},
	// 	bson.D{{"$unwind", bson.D{{"path", "$connection"}, {"preserveNullAndEmptyArrays", true}}}},
	// 	bson.D{{"$project", bson.D{{"name", 1}, {"created_by_user", 1}, {"is_active", 1}, {"is_deleted", 1}, {"max_ack_time_ms", 1}, {"max_msg_deliveries", 1}, {"client_address", "$connection.client_address"}}}},
	// 	bson.D{{"$project", bson.D{{"station", 0}, {"connection", 0}}}},
	// })
	// if err != nil {
	// 	return []models.CgMember{}, err
	// }

	// if err = cursor.All(context.TODO(), &consumers); err != nil {
	// 	return []models.CgMember{}, err
	// }
	// return consumers, nil

	ctx, cancelfunc := context.WithTimeout(context.Background(), dbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := postgresConnection.Client.Acquire(ctx)
	if err != nil {
		return []models.CgMember{}, err
	}
	query := `
		SELECT
			cg_member.name,
			cg_member.created_by_user,
			cg_member.is_active,
			cg_member.is_deleted,
			cg_member.max_ack_time_ms,
			cg_member.max_msg_deliveries,
			connection.client_address
		FROM
			cg_member
			INNER JOIN connections ON cg_member.connection_id = connections._id
		WHERE
			cg_member.consumers_group = $1
			AND cg_member.station_id = $2
		ORDER BY
			cg_member.creation_date DESC
	`
	stmt, err := conn.Conn().Prepare(ctx, "get_consumer_group_members", query)
	if err != nil {
		return []models.CgMember{}, err
	}
	rows, err := conn.Conn().Query(ctx, stmt.Name, cgName, stationId)
	if err != nil {
		return []models.CgMember{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var consumer models.CgMember
		err = rows.Scan(&consumer)
		if err != nil {
			return []models.CgMember{}, fmt.Errorf("error scanning row: %w", err)
		}
		consumers = append(consumers, consumer)
	}

	if err = rows.Err(); err != nil {
		return []models.CgMember{}, fmt.Errorf("error iterating over rows: %w", err)
	}

	return consumers, nil

}

// TODO: Aggregate
func GetConsumersByConnectionIDWithStationDetails(connectionId string) ([]models.ExtendedConsumer, error) {
	var consumers []models.ExtendedConsumer
	// cursor, err := consumersCollection.Aggregate(context.TODO(), mongo.Pipeline{
	// 	bson.D{{"$match", bson.D{{"connection_id", connectionId}, {"is_active", true}}}},
	// 	bson.D{{"$lookup", bson.D{{"from", "stations"}, {"localField", "station_id"}, {"foreignField", "_id"}, {"as", "station"}}}},
	// 	bson.D{{"$unwind", bson.D{{"path", "$station"}, {"preserveNullAndEmptyArrays", true}}}},
	// 	bson.D{{"$project", bson.D{{"_id", 1}, {"name", 1}, {"type", 1}, {"connection_id", 1}, {"created_by_user", 1}, {"creation_date", 1}, {"is_active", 1}, {"is_deleted", 1}, {"station_name", "$station.name"}}}},
	// 	bson.D{{"$project", bson.D{{"station", 0}, {"connection", 0}}}}})
	// if err != nil {
	// 	return []models.ExtendedConsumer{}, err
	// }
	// if err = cursor.All(context.TODO(), &consumers); err != nil {
	// 	return []models.ExtendedConsumer{}, err
	// }
	// return consumers, nil

	ctx, cancelfunc := context.WithTimeout(context.Background(), dbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := postgresConnection.Client.Acquire(ctx)
	if err != nil {
		return []models.ExtendedConsumer{}, err
	}
	query := `
		SELECT
			consumers._id,
			consumers.name,
			consumers.type,
			consumers.connection_id,
			consumers.created_by_user,
			consumers.creation_date,
			consumers.is_active,
			consumers.is_deleted,
			stations.name AS station_name
		FROM
			consumers
			LEFT JOIN stations ON consumers.station_id = stations._id
		WHERE
			consumers.connection_id = $1
			AND consumers.is_active = true
	`
	stmt, err := conn.Conn().Prepare(ctx, "get_consumers_by_connection_id_with_station_details", query)
	if err != nil {
		return []models.ExtendedConsumer{}, err
	}
	rows, err := conn.Conn().Query(ctx, stmt.Name, connectionId)
	if err != nil {
		return []models.ExtendedConsumer{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var consumer models.ExtendedConsumer
		err = rows.Scan(&consumer)
		if err != nil {
			return []models.ExtendedConsumer{}, fmt.Errorf("error scanning row: %w", err)
		}
		consumers = append(consumers, consumer)
	}
	if err = rows.Err(); err != nil {
		return []models.ExtendedConsumer{}, fmt.Errorf("error iterating over rows: %w", err)
	}

	return consumers, nil
}

func GetActiveConsumerByStationID(consumerName string, stationId int) (bool, models.Consumer, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), dbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := postgresConnection.Client.Acquire(ctx)
	if err != nil {
		return true, models.Consumer{}, nil
	}
	defer conn.Release()
	query := `SELECT * FROM consumers WHERE name = $1 AND station_id = $2 AND is_active = true LIMIT 1`
	stmt, err := conn.Conn().Prepare(ctx, "get_active_consumer_by_station_id", query)
	if err != nil {
		return true, models.Consumer{}, err
	}
	rows, err := conn.Conn().Query(ctx, stmt.Name, consumerName, stationId)
	if err != nil {
		return true, models.Consumer{}, err
	}
	defer rows.Close()
	consumers, err := pgx.CollectRows(rows, pgx.RowToStructByPos[models.Consumer])
	if len(consumers) == 0 {
		return false, models.Consumer{}, nil
	}
	if err != nil {
		return true, models.Consumer{}, err
	}
	return true, consumers[0], nil
}

func UpdateConsumersConnection(connectionId string, isActive bool) error {
	// _, err := consumersCollection.UpdateMany(context.TODO(),
	// 	bson.M{"connection_id": connectionId},
	// 	bson.M{"$set": bson.M{"is_active": isActive}},
	// )
	// if err != nil {
	// 	return err
	// }
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

func KillConsumersByConnections(connectionIds []string) error {
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
	ctx, cancelfunc := context.WithTimeout(context.Background(), dbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := postgresConnection.Client.Acquire(ctx)
	if err != nil {
		return true, models.Schema{}, err
	}
	defer conn.Release()
	query := `SELECT * FROM schemas WHERE name = $1`
	stmt, err := conn.Conn().Prepare(ctx, "get_schema_by_name", query)
	if err != nil {
		return true, models.Schema{}, err
	}
	rows, err := conn.Conn().Query(ctx, stmt.Name, name)
	if err != nil {
		return true, models.Schema{}, err
	}
	defer rows.Close()
	schemas, err := pgx.CollectRows(rows, pgx.RowToStructByPos[models.Schema])
	if len(schemas) == 0 {
		return false, models.Schema{}, nil
	}
	if err != nil {
		return true, models.Schema{}, err
	}
	return true, schemas[0], nil
}

func GetSchemaVersionsBySchemaID(id int) ([]models.SchemaVersion, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), dbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := postgresConnection.Client.Acquire(ctx)
	if err != nil {
		return []models.SchemaVersion{}, err
	}
	defer conn.Release()
	query := `SELECT * FROM schema_versions WHERE schema_id=$1 ORDER BY creation_date DESC`
	stmt, err := conn.Conn().Prepare(ctx, "get_schema_versions_by_schema_id", query)
	if err != nil {
		cancelfunc()
		return []models.SchemaVersion{}, err
	}
	rows, err := conn.Conn().Query(ctx, stmt.Name, id)
	if err != nil {
		return []models.SchemaVersion{}, err
	}
	defer rows.Close()
	schemaVersions, err := pgx.CollectRows(rows, pgx.RowToStructByPos[models.SchemaVersion])
	if len(schemaVersions) == 0 {
		return []models.SchemaVersion{}, nil
	}
	if err != nil {
		return []models.SchemaVersion{}, err
	}
	return schemaVersions, nil
}

func GetActiveVersionBySchemaID(id int) (models.SchemaVersion, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), dbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := postgresConnection.Client.Acquire(ctx)
	if err != nil {
		return models.SchemaVersion{}, err
	}
	defer conn.Release()
	query := `SELECT * FROM schema_versions WHERE schema_id=$1 AND active=true`
	stmt, err := conn.Conn().Prepare(ctx, "get_active_version_by_schema_id", query)
	if err != nil {
		return models.SchemaVersion{}, err
	}
	rows, err := conn.Conn().Query(ctx, stmt.Name, id)
	if err != nil {
		return models.SchemaVersion{}, err
	}
	defer rows.Close()
	schemas, err := pgx.CollectRows(rows, pgx.RowToStructByPos[models.SchemaVersion])
	if len(schemas) == 0 {
		return models.SchemaVersion{}, nil
	}
	if err != nil {
		return models.SchemaVersion{}, err
	}
	return schemas[0], nil
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

func GetSchemaVersionByNumberAndID(version int, schemaId int) (bool, models.SchemaVersion, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), dbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := postgresConnection.Client.Acquire(ctx)
	if err != nil {
		return true, models.SchemaVersion{}, err
	}
	defer conn.Release()
	query := `SELECT * FROM schema_versions WHERE schema_id=$1 AND version_number=$2`
	stmt, err := conn.Conn().Prepare(ctx, "get_active_version_by_number_and_id", query)
	if err != nil {
		return true, models.SchemaVersion{}, err
	}
	rows, err := conn.Conn().Query(ctx, stmt.Name, version, schemaId)
	if err != nil {
		return true, models.SchemaVersion{}, err
	}
	defer rows.Close()
	schemas, err := pgx.CollectRows(rows, pgx.RowToStructByPos[models.SchemaVersion])
	if len(schemas) == 0 {
		return false, models.SchemaVersion{}, nil
	}
	if err != nil {
		return true, models.SchemaVersion{}, err
	}
	return true, schemas[0], nil
}

func UpdateSchemaActiveVersion(schemaId int, versionNumber int) error {
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

func GetShcemaVersionsCount(schemaId int) (int, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), dbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := postgresConnection.Client.Acquire(ctx)
	if err != nil {
		return 0, err
	}
	defer conn.Release()
	query := `SELECT COUNT(*) FROM schema_versions WHERE schema_id=$1`
	stmt, err := conn.Conn().Prepare(ctx, "get_schema_versions_count", query)
	if err != nil {
		return 0, err
	}
	var count int
	err = conn.Conn().QueryRow(ctx, stmt.Name, schemaId).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

// TODO: Aggregate
func GetAllSchemasDetails() ([]models.ExtendedSchema, error) {
	var schemas []models.ExtendedSchema
	// cursor, err := schemasCollection.Aggregate(context.TODO(), mongo.Pipeline{
	// 	bson.D{{"$lookup", bson.D{{"from", "schema_versions"}, {"localField", "_id"}, {"foreignField", "schema_id"}, {"as", "extendedSchema"}}}},
	// 	bson.D{{"$unwind", bson.D{{"path", "$extendedSchema"}, {"preserveNullAndEmptyArrays", true}}}},
	// 	bson.D{{"$match", bson.D{{"extendedSchema.version_number", 1}}}},
	// 	bson.D{{"$lookup", bson.D{{"from", "schema_versions"}, {"localField", "_id"}, {"foreignField", "schema_id"}, {"as", "activeVersion"}}}},
	// 	bson.D{{"$unwind", bson.D{{"path", "$activeVersion"}, {"preserveNullAndEmptyArrays", true}}}},
	// 	bson.D{{"$match", bson.D{{"activeVersion.active", true}}}},
	// 	bson.D{{"$project", bson.D{{"_id", 1}, {"name", 1}, {"type", 1}, {"created_by_user", "$extendedSchema.created_by_user"}, {"creation_date", "$extendedSchema.creation_date"}, {"version_number", "$activeVersion.version_number"}}}},
	// 	bson.D{{"$sort", bson.D{{"creation_date", -1}}}},
	// })
	// if err != nil {
	// 	return []models.ExtendedSchema{}, err
	// }
	// if err = cursor.All(context.TODO(), &schemas); err != nil {
	// 	return []models.ExtendedSchema{}, err
	// }
	// return schemas, nil

	ctx, cancelfunc := context.WithTimeout(context.Background(), dbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := postgresConnection.Client.Acquire(ctx)
	if err != nil {
		return []models.ExtendedSchema{}, err
	}
	query := `SELECT s._id, s.name, s.type, sv.created_by_user, sv.creation_date, sv.version_number
	          FROM schemas AS s
	          LEFT JOIN schema_versions sv ON s._id = sv.schema_id AND sv.version_number = 1
	          LEFT JOIN schema_versions active_sv ON s._id = active_sv.schema_id AND active_sv.active = true
	          WHERE active_sv._id IS NOT NULL
	          ORDER BY sv.creation_date DESC`

	stmt, err := conn.Conn().Prepare(ctx, "get_all_schemas_details", query)
	if err != nil {
		return []models.ExtendedSchema{}, err
	}
	rows, err := conn.Conn().Query(ctx, stmt.Name)
	if err != nil {
		return []models.ExtendedSchema{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var schema models.ExtendedSchema
		if err := rows.Scan(&schema); err != nil {
			return []models.ExtendedSchema{}, err
		}
		schemas = append(schemas, schema)
	}

	if err := rows.Err(); err != nil {
		return []models.ExtendedSchema{}, err
	}

	return schemas, nil
}

func FindAndDeleteSchema(schemaIds []int) error {
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

func UpsertNewSchema(schemaName string, schemaType string) (models.Schema, int64, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), dbOperationTimeout*time.Second)
	defer cancelfunc()

	conn, err := postgresConnection.Client.Acquire(ctx)
	if err != nil {
		return models.Schema{}, 0, err
	}
	defer conn.Release()

	query := `INSERT INTO schemas ( 
		name, 
		type) 
    VALUES($1, $2) RETURNING id`

	stmt, err := conn.Conn().Prepare(ctx, "upsert_new_schema", query)
	if err != nil {
		return models.Schema{}, 0, err
	}

	newSchema := models.Schema{}
	rows, err := conn.Conn().Query(ctx, stmt.Name, schemaName, schemaType)
	if err != nil {
		return models.Schema{}, 0, err
	}
	for rows.Next() {
		err := rows.Scan(&newSchema.ID)
		if err != nil {
			return models.Schema{}, 0, err
		}
	}

	if err := rows.Err(); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			fmt.Println(pgErr.Detail)
			if pgErr.Detail != "" {
				if strings.Contains(pgErr.Detail, "already exists") {
					return models.Schema{}, 0, errors.New("Schema" + newSchema.Name + " already exists")
				} else {
					return models.Schema{}, 0, errors.New(pgErr.Detail)
				}
			} else {
				return models.Schema{}, 0, errors.New(pgErr.Message)
			}
		} else {
			return models.Schema{}, 0, err
		}
	}

	rowsAffected := rows.CommandTag().RowsAffected()
	newSchema = models.Schema{
		ID:   newSchema.ID,
		Name: schemaName,
		Type: schemaType,
	}
	return newSchema, rowsAffected, nil
}

func UpsertNewSchemaVersion(schemaVersionNumber int, username int, schemaContent string, schemaId int, messageStructName string, descriptor string, active bool) (models.SchemaVersion, int64, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), dbOperationTimeout*time.Second)
	defer cancelfunc()

	conn, err := postgresConnection.Client.Acquire(ctx)
	if err != nil {
		return models.SchemaVersion{}, 0, err
	}
	defer conn.Release()

	query := `INSERT INTO schema_versions ( 
		version_number,
		active,
		created_by,
		created_at,
		schema_content,
		schema_id,
		msg_struct_name,
		descriptor)
    VALUES($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id`

	stmt, err := conn.Conn().Prepare(ctx, "upsert_new_schema_version", query)
	if err != nil {
		return models.SchemaVersion{}, 0, err
	}

	newSchemaVersion := models.SchemaVersion{}
	createdAt := time.Now()

	rows, err := conn.Conn().Query(ctx, stmt.Name, schemaVersionNumber, active, username, createdAt, schemaContent, schemaId, messageStructName, descriptor)
	if err != nil {
		return models.SchemaVersion{}, 0, err
	}
	for rows.Next() {
		err := rows.Scan(&newSchemaVersion.ID)
		if err != nil {
			return models.SchemaVersion{}, 0, err
		}
	}

	if err := rows.Err(); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			fmt.Println(pgErr.Detail)
			if pgErr.Detail != "" {
				if strings.Contains(pgErr.Detail, "already exists") {
					return models.SchemaVersion{}, 0, errors.New("version already exists")
				} else {
					return models.SchemaVersion{}, 0, errors.New(pgErr.Detail)
				}
			} else {
				return models.SchemaVersion{}, 0, errors.New(pgErr.Message)
			}
		} else {
			return models.SchemaVersion{}, 0, err
		}

	}

	rowsAffected := rows.CommandTag().RowsAffected()
	newSchemaVersion = models.SchemaVersion{
		ID:                newSchemaVersion.ID,
		VersionNumber:     schemaVersionNumber,
		Active:            active,
		CreatedBy:         username,
		CreatedAt:         time.Now(),
		SchemaContent:     schemaContent,
		SchemaId:          schemaId,
		MessageStructName: messageStructName,
		Descriptor:        descriptor,
	}
	return newSchemaVersion, rowsAffected, nil
}

// Integration Functions
func GetIntegration(name string) (bool, models.Integration, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), dbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := postgresConnection.Client.Acquire(ctx)
	if err != nil {
		return true, models.Integration{}, err
	}
	defer conn.Release()
	defer conn.Release()
	query := `SELECT * FROM integrations WHERE name=$1`
	stmt, err := conn.Conn().Prepare(ctx, "get_integration", query)
	if err != nil {
		return true, models.Integration{}, err
	}
	rows, err := conn.Conn().Query(ctx, stmt.Name, name)
	if err != nil {
		return true, models.Integration{}, err
	}
	defer rows.Close()
	integrations, err := pgx.CollectRows(rows, pgx.RowToStructByPos[models.Integration])
	if len(integrations) == 0 {
		return false, models.Integration{}, nil
	}
	if err != nil {
		return true, models.Integration{}, err
	}
	return true, integrations[0], nil
}

func GetAllIntegrations() (bool, []models.Integration, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), dbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := postgresConnection.Client.Acquire(ctx)
	if err != nil {
		return true, []models.Integration{}, err
	}
	defer conn.Release()
	query := `SELECT * FROM integrations`
	stmt, err := conn.Conn().Prepare(ctx, "get_all_integrations", query)
	if err != nil {
		return true, []models.Integration{}, err
	}
	rows, err := conn.Conn().Query(ctx, stmt.Name)
	if err != nil {
		return true, []models.Integration{}, err
	}
	defer rows.Close()
	integrations, err := pgx.CollectRows(rows, pgx.RowToStructByPos[models.Integration])
	if err == pgx.ErrNoRows {
		return false, []models.Integration{}, nil
	}
	if err != nil {
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
	ctx, cancelfunc := context.WithTimeout(context.Background(), dbOperationTimeout*time.Second)
	defer cancelfunc()

	conn, err := postgresConnection.Client.Acquire(ctx)
	if err != nil {
		return models.Integration{}, err
	}
	defer conn.Release()

	query := `INSERT INTO integrations ( 
		name, 
		keys,
		properties) 
    VALUES($1, $2, $3) RETURNING id`

	stmt, err := conn.Conn().Prepare(ctx, "insert_new_integration", query)
	if err != nil {
		return models.Integration{}, err
	}

	newIntegration := models.Integration{}
	rows, err := conn.Conn().Query(ctx, stmt.Name, name, keys, properties)
	if err != nil {
		return models.Integration{}, err
	}
	for rows.Next() {
		err := rows.Scan(&newIntegration.ID)
		if err != nil {
			return models.Integration{}, err
		}
	}

	if err := rows.Err(); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			fmt.Println(pgErr.Detail)
			if pgErr.Detail != "" {
				if strings.Contains(pgErr.Detail, "already exists") {
					return models.Integration{}, errors.New("Integration" + newIntegration.Name + " already exists")
				} else {
					return models.Integration{}, errors.New(pgErr.Detail)
				}
			} else {
				return models.Integration{}, errors.New(pgErr.Message)
			}
		} else {
			return models.Integration{}, err
		}
	}
	newIntegration = models.Integration{
		ID:         newIntegration.ID,
		Name:       name,
		Keys:       keys,
		Properties: properties,
	}
	return newIntegration, nil
}

func UpdateIntegration(name string, keys map[string]string, properties map[string]bool) (models.Integration, error) {
	var integration models.Integration
	return integration, nil
}

// User Functions
func CreateUser(username string, userType string, hashedPassword string, fullName string, subscription bool, avatarId int) (models.User, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), dbOperationTimeout*time.Second)
	defer cancelfunc()

	conn, err := postgresConnection.Client.Acquire(ctx)
	if err != nil {
		return models.User{}, err
	}
	defer conn.Release()

	query := `INSERT INTO users ( 
		username,
		password,
		type,
		already_logged_in,
		created_at,
		avatar_id,
		full_name, 
		subscription,
		skip_get_started) 
    VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING id`

	stmt, err := conn.Conn().Prepare(ctx, "create_new_user", query)
	if err != nil {
		return models.User{}, err
	}
	createdAt := time.Now()
	skipGetStarted := false
	alreadyLoggedIn := false

	newUser := models.User{}
	rows, err := conn.Conn().Query(ctx, stmt.Name, username, hashedPassword, userType, alreadyLoggedIn, createdAt, avatarId, fullName, subscription, skipGetStarted)
	if err != nil {
		return models.User{}, err
	}
	for rows.Next() {
		err := rows.Scan(&newUser.ID)
		if err != nil {
			return models.User{}, err
		}
	}

	if err := rows.Err(); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			fmt.Println(pgErr.Detail)
			if pgErr.Detail != "" {
				if strings.Contains(pgErr.Detail, "already exists") {
					return models.User{}, errors.New("User " + newUser.Username + " already exists")
				} else {
					return models.User{}, errors.New(pgErr.Detail)
				}
			} else {
				return models.User{}, errors.New(pgErr.Message)
			}
		} else {
			return models.User{}, err
		}
	}

	newUser = models.User{
		ID:              newUser.ID,
		Username:        username,
		Password:        hashedPassword,
		FullName:        fullName,
		Subscribtion:    subscription,
		UserType:        userType,
		CreatedAt:       createdAt,
		AlreadyLoggedIn: alreadyLoggedIn,
		AvatarId:        avatarId,
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
	ctx, cancelfunc := context.WithTimeout(context.Background(), dbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := postgresConnection.Client.Acquire(ctx)
	if err != nil {
		return true, models.User{}, err
	}
	defer conn.Release()
	query := `SELECT * FROM users WHERE user_type = 'root'`
	stmt, err := conn.Conn().Prepare(ctx, "get_root_user", query)
	if err != nil {
		return true, models.User{}, err
	}
	rows, err := conn.Conn().Query(ctx, stmt.Name)
	if err != nil {
		return true, models.User{}, err
	}
	defer rows.Close()
	users, err := pgx.CollectRows(rows, pgx.RowToStructByPos[models.User])
	if len(users) == 0 {
		return false, models.User{}, nil
	}
	if err != nil {
		return true, models.User{}, err
	}
	return true, users[0], nil
}

func GetUserByUsername(username string) (bool, models.User, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), dbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := postgresConnection.Client.Acquire(ctx)
	if err != nil {
		return true, models.User{}, err
	}
	defer conn.Release()
	// TODO: replace to this line query := `SELECT * FROM users WHERE username = $1`
	query := `SELECT * FROM users WHERE username = $1`
	stmt, err := conn.Conn().Prepare(ctx, "get_user_by_username", query)
	if err != nil {
		return true, models.User{}, err
	}
	rows, err := conn.Conn().Query(ctx, stmt.Name, username)
	if err != nil {
		return true, models.User{}, err
	}
	defer rows.Close()
	users, err := pgx.CollectRows(rows, pgx.RowToStructByPos[models.User])
	if len(users) == 0 {
		return false, models.User{}, nil
	}
	if err != nil {
		return true, models.User{}, err
	}
	return true, users[0], nil
}

func GetUserByUserId(userId int) (bool, models.User, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), dbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := postgresConnection.Client.Acquire(ctx)
	if err != nil {
		return true, models.User{}, err
	}
	defer conn.Release()
	query := `SELECT * FROM users WHERE id = $1`
	stmt, err := conn.Conn().Prepare(ctx, "get_user_by_username", query)
	if err != nil {
		return true, models.User{}, err
	}
	rows, err := conn.Conn().Query(ctx, stmt.Name, userId)
	if err != nil {
		return true, models.User{}, err
	}
	defer rows.Close()
	users, err := pgx.CollectRows(rows, pgx.RowToStructByPos[models.User])
	if len(users) == 0 {
		return false, models.User{}, nil
	}
	if err != nil {
		return true, models.User{}, err
	}
	return true, users[0], nil
}

func GetAllUsers() ([]models.FilteredGenericUser, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), dbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := postgresConnection.Client.Acquire(ctx)
	if err != nil {
		return []models.FilteredGenericUser{}, err
	}
	defer conn.Release()
	query := `SELECT * FROM users`
	stmt, err := conn.Conn().Prepare(ctx, "get_all_users", query)
	if err != nil {
		return []models.FilteredGenericUser{}, err
	}
	rows, err := conn.Conn().Query(ctx, stmt.Name)
	if err != nil {
		return []models.FilteredGenericUser{}, err
	}
	defer rows.Close()
	users, err := pgx.CollectRows(rows, pgx.RowToStructByPos[models.FilteredGenericUser])
	if len(users) == 0 {
		return []models.FilteredGenericUser{}, nil
	}
	if err != nil {
		return []models.FilteredGenericUser{}, err
	}
	return users, nil
}

func GetAllApplicationUsers() ([]models.FilteredApplicationUser, error) {

	ctx, cancelfunc := context.WithTimeout(context.Background(), dbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := postgresConnection.Client.Acquire(ctx)
	if err != nil {
		return []models.FilteredApplicationUser{}, err
	}
	defer conn.Release()
	query := `SELECT * FROM users WHERE user_type='application'`
	stmt, err := conn.Conn().Prepare(ctx, "get_all_application_users", query)
	if err != nil {
		return []models.FilteredApplicationUser{}, err
	}
	rows, err := conn.Conn().Query(ctx, stmt.Name)
	if err != nil {
		return []models.FilteredApplicationUser{}, err
	}
	defer rows.Close()
	users, err := pgx.CollectRows(rows, pgx.RowToStructByPos[models.FilteredApplicationUser])
	if len(users) == 0 {
		return []models.FilteredApplicationUser{}, nil
	}
	if err != nil {
		return []models.FilteredApplicationUser{}, err
	}
	return users, nil
}

func UpdateUserAlreadyLoggedIn(userId int) {
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

// TODO: Aggregate
func GetAllActiveUsers() ([]models.FilteredUser, error) { // This function executed on stations collection
	ctx, cancelfunc := context.WithTimeout(context.Background(), dbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := postgresConnection.Client.Acquire(ctx)
	if err != nil {
		return []models.FilteredUser{}, err
	}
	query := `
		SELECT u.username, ARRAY_AGG(DISTINCT s.name) AS items
		FROM stations AS s
		LEFT JOIN users AS u ON s.created_by_user = u.username
		WHERE s.is_deleted = false OR s.is_deleted IS NULL
		GROUP BY u.username;
	`
	stmt, err := conn.Conn().Prepare(ctx, "get_all_active_users", query)
	if err != nil {
		return []models.FilteredUser{}, err
	}
	rows, err := conn.Conn().Query(ctx, stmt.Name)
	if err != nil {
		return []models.FilteredUser{}, err
	}
	defer rows.Close()
	userList, err := pgx.CollectRows(rows, pgx.RowToStructByPos[models.FilteredUser])
	if err == pgx.ErrNoRows {
		return []models.FilteredUser{}, nil
	}
	if err != nil {
		return []models.FilteredUser{}, err
	}
	return userList, nil
}

// Tags Functions
func UpsertNewTag(name string, color string, stationArr []int, schemaArr []int, userArr []int) (models.Tag, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), dbOperationTimeout*time.Second)
	defer cancelfunc()

	conn, err := postgresConnection.Client.Acquire(ctx)
	if err != nil {
		return models.Tag{}, err
	}
	defer conn.Release()

	query := `INSERT INTO tags ( 
		name,
		color,
		users,
		stations,
		schemas) 
    VALUES($1, $2, $3, $4, $5) RETURNING id`

	stmt, err := conn.Conn().Prepare(ctx, "upsert_new_tag", query)
	if err != nil {
		return models.Tag{}, err
	}

	newTag := models.Tag{}
	rows, err := conn.Conn().Query(ctx, stmt.Name, name, color, userArr, stationArr, schemaArr)
	if err != nil {
		return models.Tag{}, err
	}
	for rows.Next() {
		err := rows.Scan(&newTag.ID)
		if err != nil {
			return models.Tag{}, err
		}
	}

	if err := rows.Err(); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			fmt.Println(pgErr.Detail)
			if pgErr.Detail != "" {
				if strings.Contains(pgErr.Detail, "already exists") {
					return models.Tag{}, errors.New("Tag" + newTag.Name + " already exists")
				} else {
					return models.Tag{}, errors.New(pgErr.Detail)
				}
			} else {
				return models.Tag{}, errors.New(pgErr.Message)
			}
		} else {
			return models.Tag{}, err
		}
	}

	newTag = models.Tag{
		ID:   newTag.ID,
		Name: name, Color: color,
		Stations: stationArr,
		Schemas:  schemaArr,
		Users:    userArr,
	}
	return newTag, nil

}

func UpsertEntityToTag(tagName string, entity string, entity_id int) error {
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

func RemoveAllTagsFromEntity(entity string, entity_id int) error {
	_, err := tagsCollection.UpdateMany(context.TODO(), bson.M{}, bson.M{"$pull": bson.M{entity: entity_id}})
	if err != nil {
		return err
	}
	return nil
}

func RemoveTagFromEntity(tagName string, entity string, entity_id int) error {
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

func GetTagsByEntityID(entity string, id int) ([]models.Tag, error) {
	var entityDBList string
	switch entity {
	case "station":
		entityDBList = "stations"
	case "schema":
		entityDBList = "schemas"
	case "user":
		entityDBList = "users"
	}
	ctx, cancelfunc := context.WithTimeout(context.Background(), dbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := postgresConnection.Client.Acquire(ctx)
	if err != nil {
		return []models.Tag{}, err
	}
	defer conn.Release()
	query := fmt.Sprintf(`SELECT * FROM tags WHERE $1 = ANY(%s)`, entityDBList)
	stmt, err := conn.Conn().Prepare(ctx, "get_tags_by_entity_id", query)
	if err != nil {
		return []models.Tag{}, err
	}
	rows, err := conn.Conn().Query(ctx, stmt.Name, id)

	if err != nil {
		return nil, err
	}
	defer rows.Close()
	tags, err := pgx.CollectRows(rows, pgx.RowToStructByPos[models.Tag])
	if len(tags) == 0 {
		return []models.Tag{}, err
	}
	if err != nil {
		return []models.Tag{}, err
	}
	return tags, nil
}

func GetTagsByEntityType(entity string) ([]models.Tag, error) {
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

	ctx, cancelfunc := context.WithTimeout(context.Background(), dbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := postgresConnection.Client.Acquire(ctx)
	if err != nil {
		return []models.Tag{}, err
	}
	defer conn.Release()
	var query string
	if entityDBList == "" { // Get All
		query = `SELECT * FROM tags`
	} else {
		query = fmt.Sprintf(`SELECT * FROM tags WHERE %s IS NOT NULL AND array_length(%s, 1) > 0`, entityDBList, entityDBList)
	}
	stmt, err := conn.Conn().Prepare(ctx, "get_tags_by_entity_type", query)
	if err != nil {
		return []models.Tag{}, err
	}
	rows, err := conn.Conn().Query(ctx, stmt.Name)

	if err != nil {
		return nil, err
	}
	defer rows.Close()
	tags, err := pgx.CollectRows(rows, pgx.RowToStructByPos[models.Tag])
	if len(tags) == 0 {
		return []models.Tag{}, nil
	}
	if err != nil {
		return []models.Tag{}, err
	}

	return tags, nil
}

func GetAllUsedTags() ([]models.Tag, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), dbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := postgresConnection.Client.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Release()
	query := `SELECT * FROM tags WHERE ARRAY_LENGTH(schemas, 1) > 0 OR ARRAY_LENGTH(stations, 1) > 0 OR ARRAY_LENGTH(users, 1) > 0`
	stmt, err := conn.Conn().Prepare(ctx, "get_all_used_tags", query)
	if err != nil {
		return nil, err
	}
	rows, err := conn.Conn().Query(ctx, stmt.Name)

	if err != nil {
		return nil, err
	}
	defer rows.Close()
	tags, err := pgx.CollectRows(rows, pgx.RowToStructByPos[models.Tag])
	if len(tags) == 0 {
		return []models.Tag{}, nil
	}
	if err != nil {
		return []models.Tag{}, err
	}
	return tags, nil
}

func GetTagByName(name string) (bool, models.Tag, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), dbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := postgresConnection.Client.Acquire(ctx)
	if err != nil {
		return true, models.Tag{}, err
	}
	defer conn.Release()
	query := `SELECT * FROM tags WHERE name=$1`
	stmt, err := conn.Conn().Prepare(ctx, "get_tag_by_name", query)
	if err != nil {
		return true, models.Tag{}, err
	}
	rows, err := conn.Conn().Query(ctx, stmt.Name, name)
	if err != nil {
		return true, models.Tag{}, err
	}
	defer rows.Close()
	tags, err := pgx.CollectRows(rows, pgx.RowToStructByPos[models.Tag])
	if len(tags) == 0 {
		return false, models.Tag{}, nil
	}
	if err != nil {
		return true, models.Tag{}, err
	}
	return true, tags[0], nil
}

// Sandbox Functions
func InsertNewSanboxUser(username string, email string, firstName string, lastName string, profilePic string) (models.SandboxUser, error) {
	user := models.SandboxUser{}
	return user, nil
}

func UpdateSandboxUserAlreadyLoggedIn(userId int) {
	sandboxUsersCollection.UpdateOne(context.TODO(),
		bson.M{"_id": userId},
		bson.M{"$set": bson.M{"already_logged_in": true}},
	)
}

func GetSandboxUser(username string) (bool, models.SandboxUser, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), dbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := postgresConnection.Client.Acquire(ctx)
	if err != nil {
		return true, models.SandboxUser{}, err
	}
	defer conn.Release()
	query := `SELECT * FROM sandbox_users WHERE username = $1`
	stmt, err := conn.Conn().Prepare(ctx, "get_sandbox_user", query)
	if err != nil {
		return true, models.SandboxUser{}, err
	}
	rows, err := conn.Conn().Query(ctx, stmt.Name, username)
	if err != nil {
		return true, models.SandboxUser{}, err
	}
	defer rows.Close()
	users, err := pgx.CollectRows(rows, pgx.RowToStructByPos[models.SandboxUser])
	if len(users) == 0 {
		return false, models.SandboxUser{}, nil
	}
	if err != nil {
		return true, models.SandboxUser{}, err
	}
	return true, users[0], nil
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
func InsertImage(name string, base64Encoding string, intValue int, isString bool) error {
	err := InsertConfiguration(name, base64Encoding, intValue, isString)
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
	ctx, cancelfunc := context.WithTimeout(context.Background(), dbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := postgresConnection.Client.Acquire(ctx)
	if err != nil {
		return true, models.Image{}, err
	}
	defer conn.Release()
	query := `SELECT value FROM configurations WHERE key = $1`
	stmt, err := conn.Conn().Prepare(ctx, "get_image", query)
	if err != nil {
		return true, models.Image{}, err
	}
	rows, err := conn.Conn().Query(ctx, stmt.Name, name)
	if err != nil {
		return true, models.Image{}, err
	}
	defer rows.Close()
	images, err := pgx.CollectRows(rows, pgx.RowToStructByPos[models.Image])
	if len(images) == 0 {
		return false, models.Image{}, nil
	}
	if err != nil {
		return true, models.Image{}, err
	}
	return true, images[0], nil
}
