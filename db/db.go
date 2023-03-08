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
	"errors"
	"fmt"
	"memphis/conf"
	"strings"

	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"database/sql"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/stdlib"
)

var configuration = conf.GetConfig()

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
	Client *sql.DB
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
		if err := db.Client.Close(); err != nil {
			l.Errorf("Failed to close PostgresSql client: " + err.Error())
		}
	}()
}

func AddInexToTable(indexName, tableName, field string, dbPostgreSQL DbPostgreSQLInstance) error {
	ctx, cancelfunc := context.WithTimeout(context.Background(), dbOperationTimeout*time.Second)
	defer cancelfunc()
	addIndexQuery := fmt.Sprintf("CREATE INDEX %s ON %s (%s)", indexName, tableName, field)
	db := dbPostgreSQL.Client
	_, err := db.ExecContext(ctx, addIndexQuery)
	if err != nil {
		return err
	}
	return nil
}

func createTablesInDb(dbPostgreSQL DbPostgreSQLInstance) error {
	auditLogsTable := `CREATE TABLE IF NOT EXISTS audit_logs(
		id INTEGER NOT NULL,
		station_name VARCHAR NOT NULL,
		message VARCHAR NOT NULL,
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
		id INTEGER NOT NULL,
		username VARCHAR NOT NULL UNIQUE,
		password VARCHAR NOT NULL,
		type enum NOT NULL,
		already_logged_in BOOL NOT NULL,
		created_at TIMESTAMP NOT NULL,
		avatr_id SERIAL NOT NULL,
		full_name VARCHAR,
		subscription BOOL NOT NULL,
		skip_get_started BOOL NOT NULL,
		PRIMARY KEY (id)
		);`

	configurationsTable := `CREATE TABLE IF NOT EXISTS configurations(
		id INTEGER NOT NULL,
		key VARCHAR NOT NULL UNIQUE,
		value VARCHAR NOT NULL,
		PRIMARY KEY (id)
		);`

	connectionsTable := `CREATE TABLE IF NOT EXISTS connections(
		id INTEGER NOT NULL,
		created_by INTEGER NOT NULL,
		is_active BOOL NOT NULL,
		created_at TIMESTAMP NOT NULL,
		client_address VARCHAR NOT NULL,
		PRIMARY KEY (id),
		CONSTRAINT fk_created_by
			FOREIGN KEY(created_by)
			REFERENCES users(id)
		);`

	integrationsTable := `CREATE TABLE IF NOT EXISTS integrations(
		id INTEGER NOT NULL,
		name VARCHAR NOT NULL,
		keys JSON,
		properties JSON,
		PRIMARY KEY (id)
		);`

	schemasTable := `
	CREATE TYPE enum_type AS ENUM ('json', 'graphql', 'protobuf');
	CREATE TABLE IF NOT EXISTS schemas(
		id INTEGER NOT NULL,
		name VARCHAR NOT NULL,
		type enum_type NOT NULL,
		PRIMARY KEY (id)
		);
		CREATE INDEX name
		ON schemas (name);`

	tagsTable := `CREATE TABLE IF NOT EXISTS tags(
		id INTEGER NOT NULL,
		name VARCHAR NOT NULL,
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
		id INTEGER NOT NULL,
		name VARCHAR NOT NULL,
		station_id INTEGER NOT NULL,
		connection_id INTEGER NOT NULL,
		consumers_group VARCHAR NOT NULL,
		max_ack_time_ms BIGSERIAL NOT NULL,
		created_by INTEGER NOT NULL,
		is_active BOOL NOT NULL,
		is_deleted BOOL NOT NULL,
		created_at TIMESTAMP NOT NULL,
		max_msg_deliveries SERIAL NOT NULL,
		start_consume_from_seq BIGSERIAL NOT NULL,
		last_msgs BIGSERIAL NOT NULL,
		type enum_type_consumer NOT NULL,
		PRIMARY KEY (id),
		CONSTRAINT fk_created_by
			FOREIGN KEY(created_by)
			REFERENCES users(id),
		CONSTRAINT fk_connection_id
			FOREIGN KEY(connection_id)
			REFERENCES connections(id)
		);
		CREATE INDEX station_id
		ON consumers (station_id);`

	stationsTable := `
	CREATE TYPE enum_retention_type AS ENUM ('msg_age_sec', 'size', 'bytes');
	CREATE TYPE enum_storage_type AS ENUM ('disk', 'memory');
	CREATE TABLE IF NOT EXISTS stations(
		id INTEGER NOT NULL,
		name VARCHAR NOT NULL,
		retention_type enum_retention_type NOT NULL,
		storage_type enum_storage_type NOT NULL,
		replicas BIGSERIAL NOT NULL,
		created_by INTEGER NOT NULL,
		created_at TIMESTAMP NOT NULL,
		updated_at TIMESTAMP NOT NULL,
		is_deleted BOOL NOT NULL,
		idempotency_window_ms BIGSERIAL NOT NULL,
		is_native BOOL NOT NULL,
		tiered_storage_enabled BOOL NOT NULL,
		dls_config JSON NOT NULL,
		schema_name VARCHAR,
		schema_version_number SERIAL,
		PRIMARY KEY (id),
		CONSTRAINT fk_created_by
			FOREIGN KEY(created_by)
			REFERENCES users(id)
		);`

	schemaVersionsTable := `CREATE TABLE IF NOT EXISTS schema_versions(
		id INEGER NOT NULL,
		version_number SERIAL NOT NULL,
		active BOOL NOT NULL,
		created_by INTEGER NOT NULL,
		created_at TIMESTAMP NOT NULL,
		schema_content TEXT NOT NULL,
		schema_id INTEGER NOT NULL,
		msg_struct_name VARCHAR,
		descriptor TEXT,
		PRIMARY KEY (id),
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
		id INTEGER NOT NULL,
		name VARCHAR NOT NULL,
		station_id INTEGER NOT NULL,
		connection_id INTEGER NOT NULL,	
		created_by INTEGER NOT NULL,
		is_active BOOL NOT NULL,
		is_deleted BOOL NOT NULL,
		created_at TIMESTAMP NOT NULL,
		type enum_producer_type NOT NULL,
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
		ON producers(station_id);`

	db := dbPostgreSQL.Client
	ctx := dbPostgreSQL.Ctx
	cancelfunc := dbPostgreSQL.Cancel

	_, err := db.ExecContext(ctx, usersTable)
	if err != nil {
		var pgErr *pgconn.PgError
		errPg := errors.As(err, &pgErr)
		if errPg && !strings.Contains(pgErr.Message, "already exists") {
			cancelfunc()
			return err
		}
	}

	_, err = db.ExecContext(ctx, connectionsTable)
	if err != nil {
		var pgErr *pgconn.PgError
		errPg := errors.As(err, &pgErr)
		if errPg && !strings.Contains(pgErr.Message, "already exists") {
			cancelfunc()
			return err
		}
	}
	_, err = db.ExecContext(ctx, auditLogsTable)
	if err != nil {
		var pgErr *pgconn.PgError
		errPg := errors.As(err, &pgErr)
		if errPg && !strings.Contains(pgErr.Message, "already exists") {
			cancelfunc()
			return err
		}
	}

	_, err = db.ExecContext(ctx, configurationsTable)
	if err != nil {
		var pgErr *pgconn.PgError
		errPg := errors.As(err, &pgErr)
		if errPg && !strings.Contains(pgErr.Message, "already exists") {
			cancelfunc()
			return err
		}
	}

	_, err = db.ExecContext(ctx, integrationsTable)
	if err != nil {
		var pgErr *pgconn.PgError
		errPg := errors.As(err, &pgErr)
		if errPg && !strings.Contains(pgErr.Message, "already exists") {
			cancelfunc()
			return err
		}
	}

	_, err = db.ExecContext(ctx, schemasTable)
	if err != nil {
		var pgErr *pgconn.PgError
		errPg := errors.As(err, &pgErr)
		if errPg && !strings.Contains(pgErr.Message, "already exists") {
			cancelfunc()
			return err
		}
	}

	_, err = db.ExecContext(ctx, tagsTable)
	if err != nil {
		var pgErr *pgconn.PgError
		errPg := errors.As(err, &pgErr)
		if errPg && !strings.Contains(pgErr.Message, "already exists") {
			cancelfunc()
			return err
		}
	}

	_, err = db.ExecContext(ctx, consumersTable)
	if err != nil {
		var pgErr *pgconn.PgError
		errPg := errors.As(err, &pgErr)
		if errPg && !strings.Contains(pgErr.Message, "already exists") {
			cancelfunc()
			return err
		}
	}

	_, err = db.ExecContext(ctx, stationsTable)
	if err != nil {
		var pgErr *pgconn.PgError
		errPg := errors.As(err, &pgErr)
		if errPg && !strings.Contains(pgErr.Message, "already exists") {
			cancelfunc()
			return err
		}
	}

	_, err = db.ExecContext(ctx, schemaVersionsTable)
	if err != nil {
		var pgErr *pgconn.PgError
		errPg := errors.As(err, &pgErr)
		if errPg && !strings.Contains(pgErr.Message, "already exists") {
			cancelfunc()
			return err
		}
	}

	_, err = db.ExecContext(ctx, producersTable)
	if err != nil {
		var pgErr *pgconn.PgError
		errPg := errors.As(err, &pgErr)
		if errPg && !strings.Contains(pgErr.Message, "already exists") {
			cancelfunc()
			return err
		}
	}

	return nil
}

func InitalizePostgreSQLDbConnection(l logger) (DbPostgreSQLInstance, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), dbOperationTimeout*time.Second)

	connConfig, err := pgx.ParseConfig(configuration.POSTGRESQL_URL)
	if err != nil {
		cancelfunc()
		return DbPostgreSQLInstance{}, err
	}
	connStr := stdlib.RegisterConnConfig(connConfig)
	dbPostgreSQL, err := sql.Open("pgx", connStr)
	if err != nil {
		cancelfunc()
		return DbPostgreSQLInstance{}, err
	}

	dbPostgreSQL.SetMaxOpenConns(3)

	err = dbPostgreSQL.Ping()
	if err != nil {
		cancelfunc()
		return DbPostgreSQLInstance{}, err
	}
	l.Noticef("Established connection with the PostgreSQL DB")
	dbPostgre := DbPostgreSQLInstance{Ctx: ctx, Cancel: cancelfunc, Client: dbPostgreSQL}
	err = createTablesInDb(dbPostgre)
	if err != nil {
		cancelfunc()
		return DbPostgreSQLInstance{}, err
	}

	// err = AddInexToTable("id_index", "users", "id", dbPostgre)
	// if err != nil {
	// 	return DbPostgreSQLInstance{}, err
	// }

	return DbPostgreSQLInstance{Client: dbPostgreSQL, Ctx: ctx, Cancel: cancelfunc}, nil

}
