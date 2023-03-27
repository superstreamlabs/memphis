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

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var configuration = conf.GetConfig()

var MetadataDbClient MetadataStorage

const (
	DbOperationTimeout = 20
)

type logger interface {
	Noticef(string, ...interface{})
	Errorf(string, ...interface{})
}

type MetadataStorage struct {
	Client *pgxpool.Pool
	Ctx    context.Context
	Cancel context.CancelFunc
}

func CloseMetadataDb(db MetadataStorage, l logger) {
	defer db.Cancel()
	defer func() {
		db.Client.Close()
	}()
}

func AddIndexToTable(indexName, tableName, field string, MetadataDbClient MetadataStorage) error {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	addIndexQuery := "CREATE INDEX" + pgx.Identifier{indexName}.Sanitize() + "ON" + pgx.Identifier{tableName}.Sanitize() + "(" + pgx.Identifier{field}.Sanitize() + ")"
	db := MetadataDbClient.Client
	_, err := db.Exec(ctx, addIndexQuery)
	if err != nil {
		return err
	}
	return nil
}

func createTables(MetadataDbClient MetadataStorage) error {
	cancelfunc := MetadataDbClient.Cancel
	defer cancelfunc()
	auditLogsTable := `CREATE TABLE IF NOT EXISTS audit_logs(
		id SERIAL NOT NULL,
		station_name VARCHAR NOT NULL,
		message TEXT NOT NULL,
		created_by INTEGER NOT NULL,
		created_by_username VARCHAR NOT NULL,
		created_at TIMESTAMP NOT NULL,
		PRIMARY KEY (id));
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
		PRIMARY KEY (id));`

	configurationsTable := `CREATE TABLE IF NOT EXISTS configurations(
		id SERIAL NOT NULL,
		key VARCHAR NOT NULL UNIQUE,
		value TEXT NOT NULL,
		PRIMARY KEY (id));`

	connectionsTable := `CREATE TABLE IF NOT EXISTS connections(
		id VARCHAR NOT NULL,
		created_by INTEGER,
		created_by_username VARCHAR NOT NULL,
		is_active BOOL NOT NULL DEFAULT false,
		created_at TIMESTAMP NOT NULL,
		client_address VARCHAR NOT NULL,
		PRIMARY KEY (id));`

	integrationsTable := `CREATE TABLE IF NOT EXISTS integrations(
		id SERIAL NOT NULL,
		name VARCHAR NOT NULL UNIQUE,
		keys JSON NOT NULL DEFAULT '{}',
		properties JSON NOT NULL DEFAULT '{}',
		PRIMARY KEY (id));`

	schemasTable := `
	CREATE TYPE enum_type AS ENUM ('json', 'graphql', 'protobuf');
	CREATE TABLE IF NOT EXISTS schemas(
		id SERIAL NOT NULL,
		name VARCHAR NOT NULL,
		type enum_type NOT NULL DEFAULT 'protobuf',
		created_by_username VARCHAR NOT NULL,
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
		created_by INTEGER,
		created_by_username VARCHAR NOT NULL,
		is_active BOOL NOT NULL DEFAULT true,
		created_at TIMESTAMP NOT NULL,
		is_deleted BOOL NOT NULL DEFAULT false,
		max_msg_deliveries SERIAL NOT NULL,
		start_consume_from_seq SERIAL NOT NULL,
		last_msgs SERIAL NOT NULL,
		PRIMARY KEY (id),
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
		created_by_username VARCHAR NOT NULL,
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
		PRIMARY KEY (id));
		CREATE UNIQUE INDEX unique_station_name_deleted ON stations(name, is_deleted) WHERE is_deleted = false;`

	schemaVersionsTable := `CREATE TABLE IF NOT EXISTS schema_versions(
		id SERIAL NOT NULL,
		version_number SERIAL NOT NULL,
		active BOOL NOT NULL DEFAULT false,
		created_by INTEGER NOT NULL,
		created_by_username VARCHAR NOT NULL,
		created_at TIMESTAMP NOT NULL,
		schema_content TEXT NOT NULL,
		schema_id INTEGER NOT NULL,
		msg_struct_name VARCHAR,
		descriptor bytea,
		PRIMARY KEY (id),
		UNIQUE(version_number, schema_id),
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
		created_by_username VARCHAR NOT NULL,
		is_active BOOL NOT NULL DEFAULT true,
		created_at TIMESTAMP NOT NULL,
		is_deleted BOOL NOT NULL DEFAULT false,
		PRIMARY KEY (id),
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

	dlsMessagesTable := `
	CREATE TABLE IF NOT EXISTS dls_messages(
		id SERIAL NOT NULL,    
		station_id INT NOT NULL,
		message_seq INT NOT NULL,
		producer_id INT NOT NULL, 
		poisoned_cgs VARCHAR[],
		message_details JSON NOT NULL,    
		updated_at TIMESTAMP NOT NULL,
		message_type VARCHAR NOT NULL,
		PRIMARY KEY (id)
	)`

	db := MetadataDbClient.Client
	ctx := MetadataDbClient.Ctx

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

	_, err = db.Exec(ctx, dlsMessagesTable)
	if err != nil {
		var pgErr *pgconn.PgError
		errPg := errors.As(err, &pgErr)
		if errPg && !strings.Contains(pgErr.Message, "already exists") {
			return err
		}
	}

	return nil
}

func InitalizeMetadataDbConnection(l logger) (MetadataStorage, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)

	defer cancelfunc()
	metadataDbUser := configuration.METADATA_DB_USER
	metadataDbPassword := configuration.METADATA_DB_PASS
	metadataDbName := configuration.METADATA_DB_DBNAME
	metadataDbHost := configuration.METADATA_DB_HOST
	metadataDbPort := configuration.METADATA_DB_PORT
	var metadataDbUrl string
	if configuration.METADATA_DB_TLS_ENABLED {
		metadataDbUrl = "postgres://" + metadataDbUser + "@" + metadataDbHost + ":" + metadataDbPort + "/" + metadataDbName + "?sslmode=verify-full"
	} else {
		metadataDbUrl = "postgres://" + metadataDbUser + ":" + metadataDbPassword + "@" + metadataDbHost + ":" + metadataDbPort + "/" + metadataDbName + "?sslmode=disable"
	}

	config, err := pgxpool.ParseConfig(metadataDbUrl)
	if err != nil {
		return MetadataStorage{}, err
	}
	config.MaxConns = 5

	if configuration.METADATA_DB_TLS_ENABLED {
		cert, err := tls.LoadX509KeyPair(configuration.METADATA_DB_TLS_CRT, configuration.METADATA_DB_TLS_KEY)
		if err != nil {
			return MetadataStorage{}, err
		}

		CACert, err := os.ReadFile(configuration.METADATA_DB_TLS_CA)
		if err != nil {
			return MetadataStorage{}, err
		}

		CACertPool := x509.NewCertPool()
		CACertPool.AppendCertsFromPEM(CACert)
		config.ConnConfig.TLSConfig = &tls.Config{Certificates: []tls.Certificate{cert}, RootCAs: CACertPool, InsecureSkipVerify: true}
	}

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return MetadataStorage{}, err
	}

	err = pool.Ping(ctx)
	if err != nil {
		return MetadataStorage{}, err
	}
	l.Noticef("Established connection with the meta-data storage")
	client := MetadataStorage{Ctx: ctx, Cancel: cancelfunc, Client: pool}
	err = createTables(client)
	if err != nil {
		return MetadataStorage{}, err
	}
	MetadataDbClient = MetadataStorage{Client: pool, Ctx: ctx, Cancel: cancelfunc}
	return MetadataDbClient, nil
}

// System Keys Functions
func GetSystemKey(key string) (bool, models.SystemKey, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return false, models.SystemKey{}, err
	}
	defer conn.Release()
	query := `SELECT * FROM configurations WHERE key = $1 LIMIT 1`
	stmt, err := conn.Conn().Prepare(ctx, "get_system_key", query)
	if err != nil {
		return false, models.SystemKey{}, err
	}
	rows, err := conn.Conn().Query(ctx, stmt.Name, key)
	if err != nil {
		return false, models.SystemKey{}, err
	}
	defer rows.Close()
	systemKeys, err := pgx.CollectRows(rows, pgx.RowToStructByPos[models.SystemKey])
	if err != nil {
		return false, models.SystemKey{}, err
	}
	if len(systemKeys) == 0 {
		return false, models.SystemKey{}, nil
	}
	return true, systemKeys[0], nil
}

func InsertSystemKey(key string, value string) error {
	err := InsertConfiguration(key, value)
	if err != nil {
		return err
	}
	return nil
}

func EditConfigurationValue(key string, value string) error {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()
	query := `UPDATE configurations SET value = $2 WHERE key = $1`
	stmt, err := conn.Conn().Prepare(ctx, "edit_configuration_value", query)
	if err != nil {
		return err
	}
	_, err = conn.Conn().Query(ctx, stmt.Name, key, value)
	if err != nil {
		return err
	}
	return nil
}

// Configuration Functions
func GetConfiguration(key string) (bool, models.ConfigurationsValue, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return false, models.ConfigurationsValue{}, err
	}
	defer conn.Release()
	query := `SELECT * FROM configurations WHERE key = $1 LIMIT 1`
	stmt, err := conn.Conn().Prepare(ctx, "get_configuration", query)
	if err != nil {
		return false, models.ConfigurationsValue{}, err
	}
	rows, err := conn.Conn().Query(ctx, stmt.Name, key)
	if err != nil {
		return false, models.ConfigurationsValue{}, err
	}
	defer rows.Close()
	configurations, err := pgx.CollectRows(rows, pgx.RowToStructByPos[models.ConfigurationsValue])
	if err != nil {
		return false, models.ConfigurationsValue{}, err
	}
	if len(configurations) == 0 {
		return false, models.ConfigurationsValue{}, nil
	}
	if configurations[0].Value == "" {
		return false, models.ConfigurationsValue{}, nil
	}
	return true, configurations[0], nil
}

func InsertConfiguration(key string, value string) error {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()

	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	query := `INSERT INTO configurations( 
			key, 
			value) 
		VALUES($1, $2) 
		RETURNING id`

	stmt, err := conn.Conn().Prepare(ctx, "insert_new_configuration", query)
	if err != nil {
		return err
	}

	newConfiguration := models.ConfigurationsValue{}
	rows, err := conn.Conn().Query(ctx, stmt.Name,
		key, value)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&newConfiguration.ID)
		if err != nil {
			return err
		}
	}

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
			if pgErr.Detail != "" {
				if strings.Contains(pgErr.Detail, "already exists") {
					return errors.New("configuration" + key + " already exists")
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

func UpdateConfiguration(key string, value string) error {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()
	query := `UPDATE configurations SET value = $2 WHERE key = $1`
	stmt, err := conn.Conn().Prepare(ctx, "update_configuration", query)
	if err != nil {
		return err
	}
	_, err = conn.Conn().Query(ctx, stmt.Name, key, value)
	if err != nil {
		return err
	}
	return nil
}

// Connection Functions
func InsertConnection(connection models.Connection) error {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()

	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	query := `INSERT INTO connections ( 
		id,
		created_by, 
		created_by_username,
		is_active, 
		created_at,
		client_address) 
    VALUES($1, $2, $3, $4, $5, $6) RETURNING id`

	stmt, err := conn.Conn().Prepare(ctx, "insert_connection", query)
	if err != nil {
		return err
	}

	createdAt := time.Now()

	rows, err := conn.Conn().Query(ctx, stmt.Name, connection.ID,
		connection.CreatedBy, connection.CreatedByUsername, connection.IsActive, createdAt, connection.ClientAddress)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&connection.ID)
		if err != nil {
			return err
		}
	}

	if err := rows.Err(); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Detail != "" {
				if strings.Contains(pgErr.Detail, "already exists") {
					return errors.New("connection " + connection.ID + " already exists")
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
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()
	query := `UPDATE connections SET is_active = $1 WHERE id = $2`
	stmt, err := conn.Conn().Prepare(ctx, "update_connection", query)
	if err != nil {
		return err
	}
	_, err = conn.Conn().Query(ctx, stmt.Name, isActive, connectionId)
	if err != nil {
		return err
	}
	return nil
}

func UpdateConncetionsOfDeletedUser(userId int) error {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()
	query := `UPDATE connections SET created_by = 0, created_by_username = CONCAT(created_by_username, '(deleted)') WHERE created_by = $1 AND created_by_username NOT LIKE '%(deleted)'`
	stmt, err := conn.Conn().Prepare(ctx, "update_connection_of_deleted_user", query)
	if err != nil {
		return err
	}
	_, err = conn.Conn().Query(ctx, stmt.Name, userId)
	if err != nil {
		return err
	}
	return nil
}

func GetConnectionByID(connectionId string) (bool, models.Connection, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return false, models.Connection{}, err
	}
	defer conn.Release()
	query := `SELECT * FROM connections AS c WHERE id = $1 LIMIT 1`
	stmt, err := conn.Conn().Prepare(ctx, "get_connection_by_id", query)
	if err != nil {
		return false, models.Connection{}, err
	}
	rows, err := conn.Conn().Query(ctx, stmt.Name, connectionId)
	if err != nil {
		return false, models.Connection{}, err
	}
	defer rows.Close()
	connections, err := pgx.CollectRows(rows, pgx.RowToStructByPos[models.Connection])
	if err != nil {
		return false, models.Connection{}, err
	}
	if len(connections) == 0 {
		return false, models.Connection{}, nil
	}
	return true, connections[0], nil
}

func KillRelevantConnections(ids []string) error {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()
	query := `UPDATE connections SET is_active = false WHERE id = ANY($1)`
	stmt, err := conn.Conn().Prepare(ctx, "kill_relevant_connections", query)
	if err != nil {
		return err
	}
	_, err = conn.Conn().Query(ctx, stmt.Name, ids)
	if err != nil {
		return err
	}
	return nil
}

func GetActiveConnections() ([]models.Connection, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return []models.Connection{}, err
	}
	defer conn.Release()
	query := `SELECT * FROM connections WHERE is_active = true`
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
	if err != nil {
		return []models.Connection{}, err
	}
	if len(connections) == 0 {
		return []models.Connection{}, nil
	}
	return connections, nil
}

// Audit Logs Functions
func InsertAuditLogs(auditLogs []interface{}) error {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()

	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return err
	}

	defer conn.Release()
	var auditLog []models.AuditLog

	b, err := json.Marshal(auditLogs)
	if err != nil {
		return err
	}
	err = json.Unmarshal(b, &auditLog)
	if err != nil {
		return err
	}

	stationName := auditLog[0].StationName
	message := auditLog[0].Message
	createdBy := auditLog[0].CreatedBy
	createdAt := auditLog[0].CreatedAt
	createdByUserName := auditLog[0].CreatedByUsername

	query := `INSERT INTO audit_logs ( 
		station_name, 
		message, 
		created_by,
		created_by_username,
		created_at
		) 
    VALUES($1, $2, $3, $4, $5) RETURNING id`

	stmt, err := conn.Conn().Prepare(ctx, "insert_audit_logs", query)
	if err != nil {
		return err
	}

	newAuditLog := models.AuditLog{}
	rows, err := conn.Conn().Query(ctx, stmt.Name,
		stationName, message, createdBy, createdByUserName, createdAt)
	if err != nil {
		return err
	}
	defer rows.Close()
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
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return []models.AuditLog{}, err
	}
	defer conn.Release()
	query := `SELECT a.id, a.station_name, a.message, a.created_by, a.created_by_username, u.type, a.created_at FROM audit_logs AS a 
		LEFT JOIN users AS u ON u.id = a.created_by 
		WHERE a.station_name = $1`
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
	if err != nil {
		return []models.AuditLog{}, err
	}
	if len(auditLogs) == 0 {
		return []models.AuditLog{}, nil
	}
	return auditLogs, nil
}

func RemoveAllAuditLogsByStation(name string) error {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()

	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	removeAuditLogs := `DELETE FROM audit_logs
	WHERE station_name = $1`

	stmt, err := conn.Conn().Prepare(ctx, "remove_audit_logs_by_station", removeAuditLogs)
	if err != nil {
		return err
	}

	_, err = conn.Conn().Exec(ctx, stmt.Name, name)
	if err != nil {
		return err
	}

	return nil
}
func UpdateAuditLogsOfDeletedUser(userId int) error {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()
	query := `UPDATE audit_logs SET created_by = 0, created_by_username = CONCAT(created_by_username, '(deleted)') WHERE created_by = $1 AND created_by_username NOT LIKE '%(deleted)'`
	stmt, err := conn.Conn().Prepare(ctx, "update_audit_logs_of_deleted_user", query)
	if err != nil {
		return err
	}
	_, err = conn.Conn().Query(ctx, stmt.Name, userId)
	if err != nil {
		return err
	}
	return nil
}

// Station Functions
func GetActiveStations() ([]models.Station, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return []models.Station{}, err
	}
	defer conn.Release()
	query := `SELECT * FROM stations AS s WHERE s.is_deleted = false OR s.is_deleted IS NULL`
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
	if err != nil {
		return []models.Station{}, err
	}
	if len(stations) == 0 {
		return []models.Station{}, nil
	}
	return stations, nil
}

func GetStationByName(name string) (bool, models.Station, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return false, models.Station{}, err
	}
	defer conn.Release()
	query := `SELECT * FROM stations WHERE name = $1 AND (is_deleted = false OR is_deleted IS NULL) LIMIT 1`
	stmt, err := conn.Conn().Prepare(ctx, "get_station_by_name", query)
	if err != nil {
		return false, models.Station{}, err
	}
	rows, err := conn.Conn().Query(ctx, stmt.Name, name)
	if err != nil {
		return false, models.Station{}, err
	}
	defer rows.Close()
	stations, err := pgx.CollectRows(rows, pgx.RowToStructByPos[models.Station])
	if err != nil {
		return false, models.Station{}, err
	}
	if len(stations) == 0 {
		return false, models.Station{}, nil
	}
	return true, stations[0], nil
}

func InsertNewStation(
	stationName string,
	userId int,
	username string,
	retentionType string,
	retentionValue int,
	storageType string,
	replicas int,
	schemaName string,
	schemaVersionUpdate int,
	idempotencyWindow int64,
	isNative bool,
	dlsConfiguration models.DlsConfiguration,
	tieredStorageEnabled bool) (models.Station, int64, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()

	conn, err := MetadataDbClient.Client.Acquire(ctx)
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
		created_by_username,
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
    VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17) RETURNING id`

	stmt, err := conn.Conn().Prepare(ctx, "insert_new_station", query)
	if err != nil {
		return models.Station{}, 0, err
	}

	createAt := time.Now()
	updatedAt := time.Now()
	var stationId int

	rows, err := conn.Conn().Query(ctx, stmt.Name,
		stationName, retentionType, retentionValue, storageType, replicas, userId, username, createAt, updatedAt,
		false, schemaName, schemaVersionUpdate, idempotencyWindow, isNative, dlsConfiguration.Poison, dlsConfiguration.Schemaverse, tieredStorageEnabled)
	if err != nil {
		return models.Station{}, 0, err
	}
	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&stationId)
		if err != nil {
			return models.Station{}, 0, err
		}
	}

	if err := rows.Err(); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Detail != "" {
				if strings.Contains(pgErr.Detail, "already exists") {
					return models.Station{}, 0, errors.New("Station" + stationName + " already exists")
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

	newStation := models.Station{
		ID:                          stationId,
		Name:                        stationName,
		CreatedBy:                   userId,
		CreatedByUsername:           username,
		CreatedAt:                   createAt,
		IsDeleted:                   false,
		RetentionType:               retentionType,
		RetentionValue:              retentionValue,
		StorageType:                 storageType,
		Replicas:                    replicas,
		UpdatedAt:                   updatedAt,
		SchemaName:                  schemaName,
		SchemaVersionNumber:         schemaVersionUpdate,
		IdempotencyWindow:           idempotencyWindow,
		IsNative:                    isNative,
		DlsConfigurationPoison:      dlsConfiguration.Poison,
		DlsConfigurationSchemaverse: dlsConfiguration.Schemaverse,
		TieredStorageEnabled:        tieredStorageEnabled,
	}

	rowsAffected := rows.CommandTag().RowsAffected()
	return newStation, rowsAffected, nil
}

func GetAllStationsDetails() ([]models.ExtendedStation, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return []models.ExtendedStation{}, err
	}
	defer conn.Release()
	query := `
	SELECT s.*, COALESCE(p.id, 0),  COALESCE(p.name, ''), COALESCE(p.station_id, 0), COALESCE(p.type, 'application'), COALESCE(p.connection_id, ''), COALESCE(p.created_by, 0), COALESCE(p.created_by_username, ''), COALESCE(p.is_active, false), COALESCE(p.created_at, CURRENT_TIMESTAMP), COALESCE(p.is_deleted, false), COALESCE(c.id, 0),  COALESCE(c.name, ''), COALESCE(c.station_id, 0), COALESCE(c.type, 'application'), COALESCE(c.connection_id, ''),COALESCE(c.consumers_group, ''),COALESCE(c.max_ack_time_ms, 0), COALESCE(c.created_by, 0), COALESCE(c.created_by_username, ''), COALESCE(p.is_active, false), COALESCE(p.created_at, CURRENT_TIMESTAMP), COALESCE(p.is_deleted, false), COALESCE(c.max_msg_deliveries, 0), COALESCE(c.start_consume_from_seq, 0), COALESCE(c.last_msgs, 0) 
	FROM stations AS s
	LEFT JOIN producers AS p
	ON s.id = p.station_id 
	LEFT JOIN consumers AS c 
	ON s.id = c.station_id
	WHERE s.is_deleted = false
	GROUP BY s.id,p.id,c.id`
	stmt, err := conn.Conn().Prepare(ctx, "get_all_stations_details", query)
	if err != nil {
		return []models.ExtendedStation{}, err
	}

	rows, err := conn.Conn().Query(ctx, stmt.Name)
	if err != nil {
		return []models.ExtendedStation{}, err
	}
	if err == pgx.ErrNoRows {
		return []models.ExtendedStation{}, nil
	}
	defer rows.Close()
	stationsMap := map[int]models.ExtendedStation{}
	for rows.Next() {
		var stationRes models.Station
		var producer models.Producer
		var consumer models.Consumer
		if err := rows.Scan(
			&stationRes.ID,
			&stationRes.Name,
			&stationRes.RetentionType,
			&stationRes.RetentionValue,
			&stationRes.StorageType,
			&stationRes.Replicas,
			&stationRes.CreatedBy,
			&stationRes.CreatedByUsername,
			&stationRes.CreatedAt,
			&stationRes.UpdatedAt,
			&stationRes.IsDeleted,
			&stationRes.SchemaName,
			&stationRes.SchemaVersionNumber,
			&stationRes.IdempotencyWindow,
			&stationRes.IsNative,
			&stationRes.DlsConfigurationPoison,
			&stationRes.DlsConfigurationSchemaverse,
			&stationRes.TieredStorageEnabled,
			&producer.ID,
			&producer.Name,
			&producer.StationId,
			&producer.Type,
			&producer.ConnectionId,
			&producer.CreatedBy,
			&producer.CreatedByUsername,
			&producer.IsActive,
			&producer.CreatedAt,
			&producer.IsDeleted,
			&consumer.ID,
			&consumer.Name,
			&consumer.StationId,
			&consumer.Type,
			&consumer.ConnectionId,
			&consumer.ConsumersGroup,
			&consumer.MaxAckTimeMs,
			&consumer.CreatedBy,
			&consumer.CreatedByUsername,
			&consumer.IsActive,
			&consumer.CreatedAt,
			&consumer.IsDeleted,
			&consumer.MaxMsgDeliveries,
			&consumer.StartConsumeFromSeq,
			&consumer.LastMessages,
		); err != nil {
			return []models.ExtendedStation{}, err
		}
		if _, ok := stationsMap[stationRes.ID]; ok {
			tempStation := stationsMap[stationRes.ID]
			if producer.ID != 0 {
				tempStation.Producers = append(tempStation.Producers, producer)
			}
			if consumer.ID != 0 {
				tempStation.Consumers = append(tempStation.Consumers, consumer)
			}
			stationsMap[stationRes.ID] = tempStation
		} else {
			producers := []models.Producer{}
			consumers := []models.Consumer{}
			if producer.ID != 0 {
				producers = append(producers, producer)
			}
			if consumer.ID != 0 {
				consumers = append(consumers, consumer)
			}
			station := models.ExtendedStation{
				ID:                          stationRes.ID,
				Name:                        stationRes.Name,
				RetentionType:               stationRes.RetentionType,
				RetentionValue:              stationRes.RetentionValue,
				StorageType:                 stationRes.StorageType,
				Replicas:                    stationRes.Replicas,
				CreatedBy:                   stationRes.CreatedBy,
				CreatedAt:                   stationRes.CreatedAt,
				UpdatedAt:                   stationRes.UpdatedAt,
				IdempotencyWindow:           stationRes.IdempotencyWindow,
				IsNative:                    stationRes.IsNative,
				DlsConfigurationPoison:      stationRes.DlsConfigurationPoison,
				DlsConfigurationSchemaverse: stationRes.DlsConfigurationSchemaverse,
				Producers:                   producers,
				Consumers:                   consumers,
				TieredStorageEnabled:        stationRes.TieredStorageEnabled,
			}
			stationsMap[station.ID] = station
		}
	}
	if err := rows.Err(); err != nil {
		return []models.ExtendedStation{}, err
	}
	stations := getFilteredExtendedStations(stationsMap)
	return stations, nil
}

func getFilteredExtendedStations(stationsMap map[int]models.ExtendedStation) []models.ExtendedStation {
	stations := []models.ExtendedStation{}
	for _, station := range stationsMap {
		producersMap := map[string]models.Producer{}
		for _, p := range station.Producers {
			if _, ok := producersMap[p.Name]; ok {
				if producersMap[p.Name].CreatedAt.Before(p.CreatedAt) {
					producersMap[p.Name] = p
				}
			} else {
				producersMap[p.Name] = p
			}
		}
		producers := []models.Producer{}
		for _, p := range producersMap {
			producers = append(producers, p)
		}

		consumersMap := map[string]models.Consumer{}
		for _, c := range station.Consumers {
			if _, ok := consumersMap[c.Name]; ok {
				if consumersMap[c.Name].CreatedAt.Before(c.CreatedAt) {
					consumersMap[c.Name] = c
				}
			} else {
				consumersMap[c.Name] = c
			}
		}
		consumers := []models.Consumer{}
		for _, c := range consumersMap {
			consumers = append(consumers, c)
		}
		station.Consumers = consumers
		station.Producers = producers
		stations = append(stations, station)
	}
	return stations
}

func DeleteStationsByNames(stationNames []string) error {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()
	query := `UPDATE stations
	SET is_deleted = true
	WHERE name = ANY($1)
	AND (is_deleted = false)`
	stmt, err := conn.Conn().Prepare(ctx, "delete_stations_by_names", query)
	if err != nil {
		return err
	}
	_, err = conn.Conn().Query(ctx, stmt.Name, stationNames)
	if err != nil {
		return err
	}
	return nil
}

func DeleteStation(name string) error {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()
	query := `UPDATE stations
	SET is_deleted = true
	WHERE name = $1
	AND (is_deleted = false)`
	stmt, err := conn.Conn().Prepare(ctx, "delete_station", query)
	if err != nil {
		return err
	}
	_, err = conn.Conn().Query(ctx, stmt.Name, name)
	if err != nil {
		return err
	}
	return nil
}

func AttachSchemaToStation(stationName string, schemaName string, versionNumber int) error {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()
	query := `UPDATE stations SET schema_name = $2, schema_version_number = $3
	WHERE name = $1 AND is_deleted = false`
	stmt, err := conn.Conn().Prepare(ctx, "attach_schema_to_station", query)
	if err != nil {
		return err
	}
	_, err = conn.Conn().Query(ctx, stmt.Name, stationName, schemaName, versionNumber)
	if err != nil {
		return err
	}
	return nil
}

func DetachSchemaFromStation(stationName string) error {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()
	query := `UPDATE stations SET schema_name = '', schema_version_number = 0
	WHERE name = $1 AND is_deleted = false`
	stmt, err := conn.Conn().Prepare(ctx, "detach_schema_from_station", query)
	if err != nil {
		return err
	}
	_, err = conn.Conn().Query(ctx, stmt.Name, stationName)
	if err != nil {
		return err
	}
	return nil
}

func UpdateStationDlsConfig(stationName string, poison bool, schemaverse bool) error {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()
	query := `UPDATE stations SET dls_configuration_poison = $2, dls_configuration_schemaverse = $3
	WHERE name = $1 AND is_deleted = false`
	stmt, err := conn.Conn().Prepare(ctx, "update_station_dls_config", query)
	if err != nil {
		return err
	}
	_, err = conn.Conn().Query(ctx, stmt.Name, stationName, poison, schemaverse)
	if err != nil {
		return err
	}
	return nil
}

func UpdateStationsOfDeletedUser(userId int) error {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()
	query := `UPDATE stations SET created_by = 0, created_by_username = CONCAT(created_by_username, '(deleted)') WHERE created_by = $1 AND created_by_username NOT LIKE '%(deleted)'`
	stmt, err := conn.Conn().Prepare(ctx, "update_stations_of_deleted_user", query)
	if err != nil {
		return err
	}
	_, err = conn.Conn().Query(ctx, stmt.Name, userId)
	if err != nil {
		return err
	}
	return nil
}

func GetStationNamesUsingSchema(schemaName string) ([]string, error) {
	var stationNames []string
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
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
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
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
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()
	query := `UPDATE stations SET schema_name = '' WHERE schema_name = $1`
	stmt, err := conn.Conn().Prepare(ctx, "remove_schema_from_all_using_stations", query)
	if err != nil {
		return err
	}
	_, err = conn.Conn().Query(ctx, stmt.Name, schemaName)
	if err != nil {
		return err
	}
	return nil
}

// Producer Functions
func GetProducersByConnectionIDWithStationDetails(connectionId string) ([]models.ExtendedProducer, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return []models.ExtendedProducer{}, err
	}
	defer conn.Release()
	query := `
	SELECT p.id, p.name, p.type, p.connection_id, p.created_by, p.created_by_username, s.name, p.created_at, p.is_active, p.is_deleted, c.client_address
	FROM producers AS p
	LEFT JOIN stations AS s
	ON s.id = p.station_id
	LEFT JOIN connections AS c
	ON c.id = p.connection_id
	WHERE p.connection_id = $1 AND p.is_active = true
	GROUP BY p.id, s.id`
	stmt, err := conn.Conn().Prepare(ctx, "get_producers_by_connection_id_with_station_details", query)
	if err != nil {
		return []models.ExtendedProducer{}, err
	}
	rows, err := conn.Conn().Query(ctx, stmt.Name, connectionId)
	if err != nil {
		return []models.ExtendedProducer{}, err
	}
	defer rows.Close()
	producers, err := pgx.CollectRows(rows, pgx.RowToStructByPos[models.ExtendedProducer])
	if err != nil {
		return []models.ExtendedProducer{}, err
	}
	if len(producers) == 0 {
		return []models.ExtendedProducer{}, nil
	}
	return producers, nil

}

func UpdateProducersConnection(connectionId string, isActive bool) error {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()
	query := `UPDATE producers SET is_active = $1 WHERE connection_id = $2`
	stmt, err := conn.Conn().Prepare(ctx, "update_producers_connection", query)
	if err != nil {
		return err
	}
	_, err = conn.Conn().Query(ctx, stmt.Name, isActive, connectionId)
	if err != nil {
		return err
	}
	return nil
}

func GetProducerByID(id int) (bool, models.Producer, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return false, models.Producer{}, err
	}
	defer conn.Release()
	query := `SELECT * FROM producers WHERE id = $1`
	stmt, err := conn.Conn().Prepare(ctx, "get_producer_by_id", query)
	if err != nil {
		return false, models.Producer{}, err
	}
	rows, err := conn.Conn().Query(ctx, stmt.Name, id)
	if err != nil {
		return false, models.Producer{}, err
	}
	defer rows.Close()
	producers, err := pgx.CollectRows(rows, pgx.RowToStructByPos[models.Producer])
	if err != nil {
		return false, models.Producer{}, err
	}
	if len(producers) == 0 {
		return false, models.Producer{}, nil
	}
	return true, producers[0], nil
}

func GetProducerByNameAndConnectionID(name string, connectionId string) (bool, models.Producer, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return false, models.Producer{}, err
	}
	defer conn.Release()
	query := `SELECT * FROM producers WHERE name = $1 AND connection_id = $2`
	stmt, err := conn.Conn().Prepare(ctx, "get_producer_by_name_and_connection_id", query)
	if err != nil {
		return false, models.Producer{}, err
	}
	rows, err := conn.Conn().Query(ctx, stmt.Name, name, connectionId)
	if err != nil {
		return false, models.Producer{}, err
	}
	defer rows.Close()
	producers, err := pgx.CollectRows(rows, pgx.RowToStructByPos[models.Producer])
	if err != nil {
		return false, models.Producer{}, err
	}
	if len(producers) == 0 {
		return false, models.Producer{}, nil
	}
	return true, producers[0], nil
}

func GetProducerByStationIDAndUsername(username string, stationId int, connectionId string) (bool, models.Producer, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return false, models.Producer{}, err
	}
	defer conn.Release()
	query := `SELECT * FROM producers WHERE name = $1 AND station_id = $2 AND connection_id = $3 LIMIT 1`
	stmt, err := conn.Conn().Prepare(ctx, "get_producer_by_station_id_and_username", query)
	if err != nil {
		return false, models.Producer{}, err
	}
	rows, err := conn.Conn().Query(ctx, stmt.Name, username, stationId, connectionId)
	if err != nil {
		return false, models.Producer{}, err
	}
	defer rows.Close()
	producers, err := pgx.CollectRows(rows, pgx.RowToStructByPos[models.Producer])
	if err != nil {
		return false, models.Producer{}, err
	}
	if len(producers) == 0 {
		return false, models.Producer{}, nil
	}
	return true, producers[0], nil
}

func GetActiveProducerByStationID(producerName string, stationId int) (bool, models.Producer, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return false, models.Producer{}, err
	}
	defer conn.Release()

	query := `SELECT * FROM producers WHERE name = $1 AND station_id = $2 AND is_active = true LIMIT 1`
	stmt, err := conn.Conn().Prepare(ctx, "get_active_producer_by_station_id", query)
	if err != nil {
		return false, models.Producer{}, err
	}
	rows, err := conn.Conn().Query(ctx, stmt.Name, producerName, stationId)
	if err != nil {
		return false, models.Producer{}, err
	}
	defer rows.Close()
	producers, err := pgx.CollectRows(rows, pgx.RowToStructByPos[models.Producer])
	if err != nil {
		return false, models.Producer{}, err
	}
	if len(producers) == 0 {
		return false, models.Producer{}, nil
	}
	return true, producers[0], nil
}

func InsertNewProducer(name string, stationId int, producerType string, connectionIdObj string, createdByUser int, createdByUsername string) (models.Producer, int64, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()

	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return models.Producer{}, 0, err
	}
	defer conn.Release()

	query := `INSERT INTO producers ( 
		name, 
		station_id, 
		connection_id,
		created_by, 
		created_by_username,
		is_active, 
		is_deleted, 
		created_at, 
		type) 
    VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING id`

	stmt, err := conn.Conn().Prepare(ctx, "insert_new_producer", query)
	if err != nil {
		return models.Producer{}, 0, err
	}

	var producerId int
	createAt := time.Now()
	isActive := true
	isDeleted := false

	rows, err := conn.Conn().Query(ctx, stmt.Name, name, stationId, connectionIdObj, createdByUser, createdByUsername, isActive, isDeleted, createAt, producerType)
	if err != nil {
		return models.Producer{}, 0, err
	}
	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&producerId)
		if err != nil {
			return models.Producer{}, 0, err
		}
	}

	if err := rows.Err(); err != nil {
		return models.Producer{}, 0, err
	}

	if err := rows.Err(); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Detail != "" {
				return models.Producer{}, 0, errors.New(pgErr.Detail)
			} else {
				return models.Producer{}, 0, errors.New(pgErr.Message)
			}
		} else {
			return models.Producer{}, 0, err
		}
	}

	rowsAffected := rows.CommandTag().RowsAffected()
	newProducer := models.Producer{
		ID:                producerId,
		Name:              name,
		StationId:         stationId,
		Type:              producerType,
		ConnectionId:      connectionIdObj,
		CreatedBy:         createdByUser,
		CreatedByUsername: createdByUsername,
		IsActive:          isActive,
		CreatedAt:         time.Now(),
		IsDeleted:         isDeleted,
	}
	return newProducer, rowsAffected, nil
}

func GetAllProducers() ([]models.ExtendedProducer, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return []models.ExtendedProducer{}, err
	}
	defer conn.Release()
	query := `
		SELECT p.id, p.name, p.type, p.connection_id, p.created_by, p.created_by_username, p.created_at, s.name , p.is_active, p.is_deleted , c.client_address
		FROM producers AS p
		LEFT JOIN stations AS s ON p.station_id = s.id
		LEFT JOIN connections AS c ON p.connection_id = c.id
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
	if err != nil {
		return []models.ExtendedProducer{}, err
	}
	if len(producers) == 0 {
		return []models.ExtendedProducer{}, nil
	}
	return producers, nil
}

func GetNotDeletedProducersByStationID(stationId int) ([]models.Producer, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return []models.Producer{}, err
	}
	defer conn.Release()

	query := `SELECT * FROM producers AS p WHERE p.station_id = $1 AND p.is_deleted = false`
	stmt, err := conn.Conn().Prepare(ctx, "get_producers_by_station_id", query)
	if err != nil {
		return []models.Producer{}, err
	}
	rows, err := conn.Conn().Query(ctx, stmt.Name, stationId)
	if err != nil {
		return []models.Producer{}, err
	}
	defer rows.Close()
	producers, err := pgx.CollectRows(rows, pgx.RowToStructByPos[models.Producer])
	if err != nil {
		return []models.Producer{}, err
	}
	if len(producers) == 0 {
		return []models.Producer{}, nil
	}
	return producers, nil
}
func GetAllProducersByStationID(stationId int) ([]models.ExtendedProducer, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return []models.ExtendedProducer{}, err
	}
	defer conn.Release()
	query := `
	SELECT DISTINCT ON (p.name) p.id, p.name, p.type, p.connection_id, p.created_by, p.created_by_username, p.created_at, s.name, p.is_active, p.is_deleted, c.client_address 
	FROM producers AS p 
	LEFT JOIN stations AS s
	ON s.id = p.station_id
	LEFT JOIN connections AS c
	ON c.id = p.connection_id
	WHERE p.station_id = $1 ORDER BY p.name, p.created_at DESC`
	stmt, err := conn.Conn().Prepare(ctx, "get_producers_by_station_id", query)
	if err != nil {
		return []models.ExtendedProducer{}, err
	}
	rows, err := conn.Conn().Query(ctx, stmt.Name, stationId)
	if err != nil {
		return []models.ExtendedProducer{}, err
	}
	defer rows.Close()

	producers, err := pgx.CollectRows(rows, pgx.RowToStructByPos[models.ExtendedProducer])
	if err != nil {
		return []models.ExtendedProducer{}, err
	}
	if len(producers) == 0 {
		return []models.ExtendedProducer{}, nil
	}
	return producers, nil
}

func DeleteProducerByNameAndStationID(name string, stationId int) (bool, models.Producer, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return false, models.Producer{}, err
	}
	defer conn.Release()
	query := `UPDATE producers SET is_active = false, is_deleted = true WHERE name = $1 AND station_id = $2 AND is_active = true RETURNING *`
	stmt, err := conn.Conn().Prepare(ctx, "delete_producer_by_name_and_station_id", query)
	if err != nil {
		return false, models.Producer{}, err
	}
	rows, err := conn.Conn().Query(ctx, stmt.Name, name, stationId)
	if err != nil {
		return false, models.Producer{}, err
	}
	defer rows.Close()
	producers, err := pgx.CollectRows(rows, pgx.RowToStructByPos[models.Producer])
	if err != nil {
		return false, models.Producer{}, err
	}
	if len(producers) == 0 {
		return false, models.Producer{}, nil
	}
	return true, producers[0], nil
}

func DeleteProducersByStationID(stationId int) error {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()
	query := `UPDATE producers SET is_active = false, is_deleted = true WHERE station_id = $1`
	stmt, err := conn.Conn().Prepare(ctx, "delete_producers_by_station_id", query)
	if err != nil {
		return err
	}
	_, err = conn.Conn().Query(ctx, stmt.Name, stationId)
	if err != nil {
		return err
	}
	return nil
}

func CountActiveProudcersByStationID(stationId int) (int64, error) {
	var activeCount int64
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
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
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
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

func UpdateProducersOfDeletedUser(userId int) error {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()
	query := `UPDATE stations SET created_by = 0, created_by_username = CONCAT(created_by_username, '(deleted)') WHERE created_by = $1 AND created_by_username NOT LIKE '%(deleted)'`
	stmt, err := conn.Conn().Prepare(ctx, "update_producers_of_deleted_user", query)
	if err != nil {
		return err
	}
	_, err = conn.Conn().Query(ctx, stmt.Name, userId)
	if err != nil {
		return err
	}
	return nil
}

func KillProducersByConnections(connectionIds []string) error {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()
	query := `UPDATE producers SET is_active = false WHERE connection_id = ANY($1)`
	stmt, err := conn.Conn().Prepare(ctx, "kill_producers_by_connections", query)
	if err != nil {
		return err
	}
	_, err = conn.Conn().Query(ctx, stmt.Name, connectionIds)
	if err != nil {
		return err
	}
	return nil
}

// Consumer Functions
func GetActiveConsumerByCG(consumersGroup string, stationId int) (bool, models.Consumer, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return false, models.Consumer{}, err
	}
	defer conn.Release()

	query := `SELECT * FROM consumers WHERE consumers_group = $1 AND station_id = $2 AND is_deleted = false LIMIT 1`
	stmt, err := conn.Conn().Prepare(ctx, "get_active_consumer_by_cg", query)
	if err != nil {
		return false, models.Consumer{}, err
	}
	rows, err := conn.Conn().Query(ctx, stmt.Name, consumersGroup, stationId)
	if err != nil {
		return false, models.Consumer{}, err
	}
	defer rows.Close()
	consumers, err := pgx.CollectRows(rows, pgx.RowToStructByPos[models.Consumer])
	if err != nil {
		return false, models.Consumer{}, err
	}
	if len(consumers) == 0 {
		return false, models.Consumer{}, nil
	}
	return true, consumers[0], nil
}

func InsertNewConsumer(name string,
	stationId int,
	consumerType string,
	connectionIdObj string,
	createdBy int,
	createdByUsername string,
	cgName string,
	maxAckTime int,
	maxMsgDeliveries int,
	startConsumeFromSequence uint64,
	lastMessages int64) (models.Consumer, int64, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()

	conn, err := MetadataDbClient.Client.Acquire(ctx)
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
		created_by_username,
		is_active, 
		is_deleted, 
		created_at,
		max_msg_deliveries,
		start_consume_from_seq,
		last_msgs,
		type) 
    VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14) 
	RETURNING id`

	stmt, err := conn.Conn().Prepare(ctx, "insert_new_consumer", query)
	if err != nil {
		return models.Consumer{}, 0, err
	}

	var consumerId int
	createdAt := time.Now()
	isActive := true
	isDeleted := false

	rows, err := conn.Conn().Query(ctx, stmt.Name,
		name, stationId, connectionIdObj, cgName, maxAckTime, createdBy, createdByUsername, isActive, isDeleted, createdAt, maxMsgDeliveries, startConsumeFromSequence, lastMessages, consumerType)
	if err != nil {
		return models.Consumer{}, 0, err
	}
	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&consumerId)
		if err != nil {
			return models.Consumer{}, 0, err
		}
	}

	if err := rows.Err(); err != nil {
		return models.Consumer{}, 0, err
	}

	if err := rows.Err(); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			fmt.Println(pgErr.Detail)
			if pgErr.Detail != "" {
				return models.Consumer{}, 0, errors.New(pgErr.Detail)
			} else {
				return models.Consumer{}, 0, errors.New(pgErr.Message)
			}
		} else {
			return models.Consumer{}, 0, err
		}
	}

	rowsAffected := rows.CommandTag().RowsAffected()
	newConsumer := models.Consumer{
		ID:                  consumerId,
		Name:                name,
		StationId:           stationId,
		Type:                consumerType,
		ConnectionId:        connectionIdObj,
		CreatedBy:           createdBy,
		CreatedByUsername:   createdByUsername,
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

func GetAllConsumers() ([]models.ExtendedConsumer, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return []models.ExtendedConsumer{}, err
	}
	defer conn.Release()
	query := `
		SELECT c.name, c.created_by, c.created_by_username, c.created_at, c.is_active, c.is_deleted, con.client_address, c.consumers_group, c.max_ack_time_ms, c.max_msg_deliveries, s.name 
		FROM consumers AS c
		LEFT JOIN stations AS s ON c.station_id = s.id
		LEFT JOIN connections AS con ON c.connection_id = con.id
	`
	stmt, err := conn.Conn().Prepare(ctx, "get_all_consumers", query)
	if err != nil {
		return []models.ExtendedConsumer{}, err
	}
	rows, err := conn.Conn().Query(ctx, stmt.Name)
	if err != nil {
		return []models.ExtendedConsumer{}, err
	}
	defer rows.Close()
	consumers, err := pgx.CollectRows(rows, pgx.RowToStructByPos[models.ExtendedConsumer])
	if err != nil {
		return []models.ExtendedConsumer{}, err
	}
	if len(consumers) == 0 {
		return []models.ExtendedConsumer{}, nil
	}
	return consumers, nil
}

func GetAllConsumersByStation(stationId int) ([]models.ExtendedConsumer, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return []models.ExtendedConsumer{}, err
	}
	defer conn.Release()
	query := `
		SELECT DISTINCT ON (c.name) c.id, c.name, c.created_by, c.created_by_username, c.created_at, c.is_active, c.is_deleted, con.client_address, c.consumers_group, c.max_ack_time_ms, c.max_msg_deliveries, s.name FROM consumers AS c
		LEFT JOIN stations AS s ON s.id = c.station_id
		LEFT JOIN connections AS con ON con.id = c.connection_id
	WHERE c.station_id = $1 ORDER BY c.name, c.created_at DESC`
	stmt, err := conn.Conn().Prepare(ctx, "get_all_consumers_by_station", query)
	if err != nil {
		return []models.ExtendedConsumer{}, err
	}
	rows, err := conn.Conn().Query(ctx, stmt.Name, stationId)
	if err != nil {
		return []models.ExtendedConsumer{}, err
	}
	defer rows.Close()
	consumers, err := pgx.CollectRows(rows, pgx.RowToStructByPos[models.ExtendedConsumer])
	if err != nil {
		return []models.ExtendedConsumer{}, err
	}
	if len(consumers) == 0 {
		return []models.ExtendedConsumer{}, nil
	}
	return consumers, nil
}

func DeleteConsumer(name string, stationId int) (bool, models.Consumer, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return false, models.Consumer{}, err
	}
	defer conn.Release()
	query1 := ` UPDATE consumers SET is_active = false, is_deleted = true WHERE name = $1 AND station_id = $2 AND is_active = true RETURNING *`
	findAndUpdateStmt, err := conn.Conn().Prepare(ctx, "find_and_update_consumers", query1)
	if err != nil {
		return false, models.Consumer{}, err
	}
	rows, err := conn.Conn().Query(ctx, findAndUpdateStmt.Name, name, stationId)
	if err != nil {
		return false, models.Consumer{}, err
	}
	defer rows.Close()
	consumers, err := pgx.CollectRows(rows, pgx.RowToStructByPos[models.Consumer])
	if err != nil {
		return false, models.Consumer{}, err
	}
	if len(consumers) == 0 {
		return false, models.Consumer{}, err
	}
	query2 := `UPDATE consumers SET is_active = false, is_deleted = true WHERE name = $1 AND station_id = $2`
	updateAllStmt, err := conn.Conn().Prepare(ctx, "update_all_related_consumers", query2)
	if err != nil {
		return false, models.Consumer{}, err
	}
	_, err = conn.Conn().Query(ctx, updateAllStmt.Name, name, stationId)
	if err != nil {
		return false, models.Consumer{}, err
	}
	return true, consumers[0], nil
}

func DeleteConsumersByStationID(stationId int) error {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()
	query := `UPDATE consumers SET is_active = false, is_deleted = true WHERE station_id = $1`
	stmt, err := conn.Conn().Prepare(ctx, "delete_consumers_by_station_id", query)
	if err != nil {
		return err
	}
	_, err = conn.Conn().Query(ctx, stmt.Name, stationId)
	if err != nil {
		return err
	}
	return nil
}

func CountActiveConsumersInCG(consumersGroup string, stationId int) (int64, error) {
	var count int64
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
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
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
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
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
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

func GetConsumerGroupMembers(cgName string, stationId int) ([]models.CgMember, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return []models.CgMember{}, err
	}
	defer conn.Release()
	query := `
		SELECT
			c.name,
			con.client_address,
			c.is_active,
			c.is_deleted,
			c.created_by,
			c.created_by_username,
			c.max_ack_time_ms,
			c.max_msg_deliveries
		FROM
			consumers AS c
			INNER JOIN connections AS con ON c.connection_id = con.id
		WHERE
			c.consumers_group = $1
			AND c.station_id = $2
		ORDER BY
			c.created_at DESC
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

	consumers, err := pgx.CollectRows(rows, pgx.RowToStructByPos[models.CgMember])
	if err != nil {
		return []models.CgMember{}, err
	}
	if len(consumers) == 0 {
		return []models.CgMember{}, nil
	}
	return consumers, nil
}

func GetConsumersByConnectionIDWithStationDetails(connectionId string) ([]models.ExtendedConsumer, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return []models.ExtendedConsumer{}, err
	}
	defer conn.Release()
	query := `
		SELECT c.name, c.created_by, c.created_by_username, c.created_at, c.is_active, c.is_deleted, con.client_address, c.consumers_group, c.max_ack_time_ms, c.max_msg_deliveries, s.name,  
		FROM consumers AS c
		FROM
		consumers AS c
		LEFT JOIN stations AS s ON s.id = c.station_id
		LEFT JOIN connections AS con ON con.id = c.connection_id
	WHERE
		c.connection_id = $1
`
	stmt, err := conn.Conn().Prepare(ctx, "get_all_consumers_by_connection_id_with_station_details", query)
	if err != nil {
		return []models.ExtendedConsumer{}, err
	}
	rows, err := conn.Conn().Query(ctx, stmt.Name, connectionId)
	if err != nil {
		return []models.ExtendedConsumer{}, err
	}
	defer rows.Close()
	consumers, err := pgx.CollectRows(rows, pgx.RowToStructByPos[models.ExtendedConsumer])
	if err != nil {
		return []models.ExtendedConsumer{}, err
	}
	if len(consumers) == 0 {
		return []models.ExtendedConsumer{}, nil
	}
	return consumers, nil
}

func GetActiveConsumerByStationID(consumerName string, stationId int) (bool, models.Consumer, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return false, models.Consumer{}, err
	}
	defer conn.Release()
	query := `SELECT * FROM consumers WHERE name = $1 AND station_id = $2 AND is_active = true LIMIT 1`
	stmt, err := conn.Conn().Prepare(ctx, "get_active_consumer_by_station_id", query)
	if err != nil {
		return false, models.Consumer{}, err
	}
	rows, err := conn.Conn().Query(ctx, stmt.Name, consumerName, stationId)
	if err != nil {
		return false, models.Consumer{}, err
	}
	defer rows.Close()
	consumers, err := pgx.CollectRows(rows, pgx.RowToStructByPos[models.Consumer])
	if err != nil {
		return false, models.Consumer{}, err
	}
	if len(consumers) == 0 {
		return false, models.Consumer{}, nil
	}
	return true, consumers[0], nil
}

func UpdateConsumersConnection(connectionId string, isActive bool) error {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()
	query := `UPDATE consumers SET is_active = $1 WHERE connection_id = $2`
	stmt, err := conn.Conn().Prepare(ctx, "update_consumers_connection", query)
	if err != nil {
		return err
	}
	_, err = conn.Conn().Query(ctx, stmt.Name, isActive, connectionId)
	if err != nil {
		return err
	}
	return nil
}

func UpdateConsumersOfDeletedUser(userId int) error {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()
	query := `UPDATE consumers SET created_by = 0, created_by_username = CONCAT(created_by_username, '(deleted)') WHERE created_by = $1 AND created_by_username NOT LIKE '%(deleted)'`
	stmt, err := conn.Conn().Prepare(ctx, "update_consumers_of_deleted_user", query)
	if err != nil {
		return err
	}
	_, err = conn.Conn().Query(ctx, stmt.Name, userId)
	if err != nil {
		return err
	}
	return nil
}

func KillConsumersByConnections(connectionIds []string) error {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()
	query := `UPDATE consumers SET is_active = false WHERE connection_id = ANY($1)`
	stmt, err := conn.Conn().Prepare(ctx, "kill_consumers_by_connections", query)
	if err != nil {
		return err
	}
	_, err = conn.Conn().Query(ctx, stmt.Name, connectionIds)
	if err != nil {
		return err
	}
	return nil
}

// Schema Functions
func GetSchemaByName(name string) (bool, models.Schema, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return false, models.Schema{}, err
	}
	defer conn.Release()
	query := `SELECT * FROM schemas WHERE name = $1 LIMIT 1`
	stmt, err := conn.Conn().Prepare(ctx, "get_schema_by_name", query)
	if err != nil {
		return false, models.Schema{}, err
	}
	rows, err := conn.Conn().Query(ctx, stmt.Name, name)
	if err != nil {
		return false, models.Schema{}, err
	}
	defer rows.Close()
	schemas, err := pgx.CollectRows(rows, pgx.RowToStructByPos[models.Schema])
	if err != nil {
		return false, models.Schema{}, err
	}
	if len(schemas) == 0 {
		return false, models.Schema{}, nil
	}
	return true, schemas[0], nil
}

func GetSchemaVersionsBySchemaID(id int) ([]models.SchemaVersion, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return []models.SchemaVersion{}, err
	}
	defer conn.Release()
	query := `SELECT * FROM schema_versions WHERE schema_id=$1 ORDER BY created_at DESC`
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
	schemaVersionsRes, err := pgx.CollectRows(rows, pgx.RowToStructByPos[models.SchemaVersionResponse])
	if err != nil {
		return []models.SchemaVersion{}, err
	}
	if len(schemaVersionsRes) == 0 {
		return []models.SchemaVersion{}, nil
	}
	schemaVersions := []models.SchemaVersion{}
	for _, v := range schemaVersionsRes {
		version := models.SchemaVersion{
			ID:                v.ID,
			VersionNumber:     v.VersionNumber,
			Active:            v.Active,
			CreatedBy:         v.CreatedBy,
			CreatedByUsername: v.CreatedByUsername,
			CreatedAt:         v.CreatedAt,
			SchemaContent:     v.SchemaContent,
			SchemaId:          v.SchemaId,
			MessageStructName: v.MessageStructName,
			Descriptor:        string(v.Descriptor),
		}

		schemaVersions = append(schemaVersions, version)
	}
	return schemaVersions, nil
}

func GetActiveVersionBySchemaID(id int) (models.SchemaVersion, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return models.SchemaVersion{}, err
	}
	defer conn.Release()
	query := `SELECT * FROM schema_versions WHERE schema_id=$1 AND active=true LIMIT 1`
	stmt, err := conn.Conn().Prepare(ctx, "get_active_version_by_schema_id", query)
	if err != nil {
		return models.SchemaVersion{}, err
	}
	rows, err := conn.Conn().Query(ctx, stmt.Name, id)
	if err != nil {
		return models.SchemaVersion{}, err
	}
	defer rows.Close()
	schemas, err := pgx.CollectRows(rows, pgx.RowToStructByPos[models.SchemaVersionResponse])
	if err != nil {
		return models.SchemaVersion{}, err
	}
	if len(schemas) == 0 {
		return models.SchemaVersion{}, nil
	}
	schemaVersion := models.SchemaVersion{
		ID:                schemas[0].ID,
		VersionNumber:     schemas[0].VersionNumber,
		Active:            schemas[0].Active,
		CreatedBy:         schemas[0].CreatedBy,
		CreatedByUsername: schemas[0].CreatedByUsername,
		CreatedAt:         schemas[0].CreatedAt,
		SchemaContent:     schemas[0].SchemaContent,
		SchemaId:          schemas[0].SchemaId,
		MessageStructName: schemas[0].MessageStructName,
		Descriptor:        string(schemas[0].Descriptor),
	}

	return schemaVersion, nil
}

func UpdateSchemasOfDeletedUser(userId int) error {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()
	query := ` UPDATE schemas
	SET created_by_username = CONCAT(created_by_username, '(deleted)')
	WHERE created_by_username = (
		SELECT username FROM users WHERE id = $1
	)
	AND created_by_username NOT LIKE '%(deleted)'`
	stmt, err := conn.Conn().Prepare(ctx, "update_schemas_of_deleted_user", query)
	if err != nil {
		return err
	}
	_, err = conn.Conn().Query(ctx, stmt.Name, userId)
	if err != nil {
		return err
	}
	return nil
}

func UpdateSchemaVersionsOfDeletedUser(userId int) error {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()
	query := ` UPDATE schema_versions
	SET created_by_username = CONCAT(created_by_username, '(deleted)')
	WHERE created_by_username = (
		SELECT username FROM users WHERE id = $1
	)
	AND created_by_username NOT LIKE '%(deleted)'`
	stmt, err := conn.Conn().Prepare(ctx, "update_schema_versions_of_deleted_user", query)
	if err != nil {
		return err
	}
	_, err = conn.Conn().Query(ctx, stmt.Name, userId)
	if err != nil {
		return err
	}
	return nil
}

func GetSchemaVersionByNumberAndID(version int, schemaId int) (bool, models.SchemaVersion, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return false, models.SchemaVersion{}, err
	}
	defer conn.Release()
	query := `SELECT * FROM schema_versions WHERE schema_id=$1 AND version_number=$2 LIMIT 1`
	stmt, err := conn.Conn().Prepare(ctx, "get_active_version_by_number_and_id", query)
	if err != nil {
		return false, models.SchemaVersion{}, err
	}
	rows, err := conn.Conn().Query(ctx, stmt.Name, schemaId, version)
	if err != nil {
		return false, models.SchemaVersion{}, err
	}
	defer rows.Close()
	schemas, err := pgx.CollectRows(rows, pgx.RowToStructByPos[models.SchemaVersionResponse])
	if err != nil {
		return false, models.SchemaVersion{}, err
	}
	if len(schemas) == 0 {
		return false, models.SchemaVersion{}, nil
	}
	schemaVersion := models.SchemaVersion{
		ID:                schemas[0].ID,
		VersionNumber:     schemas[0].VersionNumber,
		Active:            schemas[0].Active,
		CreatedBy:         schemas[0].CreatedBy,
		CreatedByUsername: schemas[0].CreatedByUsername,
		CreatedAt:         schemas[0].CreatedAt,
		SchemaContent:     schemas[0].SchemaContent,
		SchemaId:          schemas[0].SchemaId,
		MessageStructName: schemas[0].MessageStructName,
		Descriptor:        string(schemas[0].Descriptor),
	}
	return true, schemaVersion, nil
}

func UpdateSchemaActiveVersion(schemaId int, versionNumber int) error {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()
	query := `UPDATE schema_versions
		SET active = CASE
		WHEN version_number = $2 THEN true
		ELSE false
		END
	WHERE schema_id = $1
`
	stmt, err := conn.Conn().Prepare(ctx, "update_schema_active_version", query)
	if err != nil {
		return err
	}
	_, err = conn.Conn().Query(ctx, stmt.Name, schemaId, versionNumber)
	if err != nil {
		return err
	}
	return nil
}

func GetShcemaVersionsCount(schemaId int) (int, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
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

func GetAllSchemasDetails() ([]models.ExtendedSchema, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return []models.ExtendedSchema{}, err
	}
	defer conn.Release()
	query := `SELECT s.id, s.name, s.type, sv.created_by, s.created_by_username, sv.created_at, asv.version_number
	          FROM schemas AS s
	          LEFT JOIN schema_versions AS sv ON s.id = sv.schema_id AND sv.version_number = 1
	          LEFT JOIN schema_versions AS asv ON s.id = asv.schema_id AND asv.active = true
	          WHERE asv.id IS NOT NULL
	          ORDER BY sv.created_at DESC`
	stmt, err := conn.Conn().Prepare(ctx, "get_all_schemas_details", query)
	if err != nil {
		return []models.ExtendedSchema{}, err
	}

	rows, err := conn.Conn().Query(ctx, stmt.Name)
	if err != nil {
		return []models.ExtendedSchema{}, err
	}
	if err == pgx.ErrNoRows {
		return []models.ExtendedSchema{}, nil
	}
	defer rows.Close()
	schemas := []models.ExtendedSchema{}
	for rows.Next() {
		var sc models.ExtendedSchema
		err := rows.Scan(&sc.ID, &sc.Name, &sc.Type, &sc.CreatedBy, &sc.CreatedByUsername, &sc.CreatedAt, &sc.ActiveVersionNumber)
		if err != nil {
			return []models.ExtendedSchema{}, err
		}
		schemas = append(schemas, sc)
	}
	if len(schemas) == 0 {
		return []models.ExtendedSchema{}, nil
	}
	return schemas, nil
}

func FindAndDeleteSchema(schemaIds []int) error {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()

	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	removeSchemaVersionsQuery := `DELETE FROM schema_versions
	WHERE schema_id = ANY($1)`

	stmt, err := conn.Conn().Prepare(ctx, "remove_schema_versions", removeSchemaVersionsQuery)
	if err != nil {
		return err
	}

	_, err = conn.Conn().Exec(ctx, stmt.Name, schemaIds)
	if err != nil {
		return err
	}

	removeSchemasQuery := `DELETE FROM schemas
	WHERE id = ANY($1)`

	stmt, err = conn.Conn().Prepare(ctx, "remove_schemas", removeSchemasQuery)
	if err != nil {
		return err
	}

	_, err = conn.Conn().Exec(ctx, stmt.Name, schemaIds)
	if err != nil {
		return err
	}
	return nil
}

func InsertNewSchema(schemaName string, schemaType string, createdByUsername string) (models.Schema, int64, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()

	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return models.Schema{}, 0, err
	}
	defer conn.Release()

	query := `INSERT INTO schemas ( 
		name, 
		type,
		created_by_username) 
    VALUES($1, $2, $3) RETURNING id`

	stmt, err := conn.Conn().Prepare(ctx, "insert_new_schema", query)
	if err != nil {
		return models.Schema{}, 0, err
	}

	var schemaId int
	rows, err := conn.Conn().Query(ctx, stmt.Name, schemaName, schemaType, createdByUsername)
	if err != nil {
		return models.Schema{}, 0, err
	}
	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&schemaId)
		if err != nil {
			return models.Schema{}, 0, err
		}
	}

	if err := rows.Err(); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Detail != "" {
				if strings.Contains(pgErr.Detail, "already exists") {
					return models.Schema{}, 0, errors.New("Schema" + schemaName + " already exists")
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
	newSchema := models.Schema{
		ID:                schemaId,
		Name:              schemaName,
		Type:              schemaType,
		CreatedByUsername: createdByUsername,
	}
	return newSchema, rowsAffected, nil
}

func InsertNewSchemaVersion(schemaVersionNumber int, userId int, username string, schemaContent string, schemaId int, messageStructName string, descriptor string, active bool) (models.SchemaVersion, int64, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()

	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return models.SchemaVersion{}, 0, err
	}
	defer conn.Release()

	query := `INSERT INTO schema_versions ( 
		version_number,
		active,
		created_by,
		created_by_username,
		created_at,
		schema_content,
		schema_id,
		msg_struct_name,
		descriptor)
    VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING id`

	stmt, err := conn.Conn().Prepare(ctx, "insert_new_schema_version", query)
	if err != nil {
		return models.SchemaVersion{}, 0, err
	}

	var schemaVersionId int
	createdAt := time.Now()

	rows, err := conn.Conn().Query(ctx, stmt.Name, schemaVersionNumber, active, userId, username, createdAt, schemaContent, schemaId, messageStructName, []byte(descriptor))
	if err != nil {
		return models.SchemaVersion{}, 0, err
	}
	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&schemaVersionId)
		if err != nil {
			return models.SchemaVersion{}, 0, err
		}
	}

	if err := rows.Err(); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
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
	newSchemaVersion := models.SchemaVersion{
		ID:                schemaVersionId,
		VersionNumber:     schemaVersionNumber,
		Active:            active,
		CreatedBy:         userId,
		CreatedByUsername: username,
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
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return false, models.Integration{}, err
	}
	defer conn.Release()
	query := `SELECT * FROM integrations WHERE name=$1 LIMIT 1`
	stmt, err := conn.Conn().Prepare(ctx, "get_integration", query)
	if err != nil {
		return false, models.Integration{}, err
	}
	rows, err := conn.Conn().Query(ctx, stmt.Name, name)
	if err != nil {
		return false, models.Integration{}, err
	}
	defer rows.Close()
	integrations, err := pgx.CollectRows(rows, pgx.RowToStructByPos[models.Integration])
	if err != nil {
		return false, models.Integration{}, err
	}
	if len(integrations) == 0 {
		return false, models.Integration{}, nil
	}
	return true, integrations[0], nil
}

func GetAllIntegrations() (bool, []models.Integration, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return false, []models.Integration{}, err
	}
	defer conn.Release()
	query := `SELECT * FROM integrations`
	stmt, err := conn.Conn().Prepare(ctx, "get_all_integrations", query)
	if err != nil {
		return false, []models.Integration{}, err
	}
	rows, err := conn.Conn().Query(ctx, stmt.Name)
	if err != nil {
		return false, []models.Integration{}, err
	}
	defer rows.Close()
	integrations, err := pgx.CollectRows(rows, pgx.RowToStructByPos[models.Integration])
	if err != nil {
		return false, []models.Integration{}, err
	}
	if len(integrations) == 0 {
		return false, []models.Integration{}, nil
	}
	return true, integrations, nil
}

func DeleteIntegration(name string) error {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()

	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	removeIntegrationQuery := `DELETE FROM integrations
	WHERE name = $1`

	stmt, err := conn.Conn().Prepare(ctx, "remove_integration", removeIntegrationQuery)
	if err != nil {
		return err
	}

	_, err = conn.Conn().Exec(ctx, stmt.Name, name)
	if err != nil {
		return err
	}

	return nil

}

func InsertNewIntegration(name string, keys map[string]string, properties map[string]bool) (models.Integration, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()

	conn, err := MetadataDbClient.Client.Acquire(ctx)
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

	var integrationId int
	rows, err := conn.Conn().Query(ctx, stmt.Name, name, keys, properties)
	if err != nil {
		return models.Integration{}, err
	}
	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&integrationId)
		if err != nil {
			return models.Integration{}, err
		}
	}

	if err := rows.Err(); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Detail != "" {
				if strings.Contains(pgErr.Detail, "already exists") {
					return models.Integration{}, errors.New("Integration" + name + " already exists")
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
	newIntegration := models.Integration{
		ID:         integrationId,
		Name:       name,
		Keys:       keys,
		Properties: properties,
	}
	return newIntegration, nil
}

func UpdateIntegration(name string, keys map[string]string, properties map[string]bool) (models.Integration, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return models.Integration{}, err
	}
	defer conn.Release()
	query := `
	INSERT INTO integrations(name, keys, properties)
	VALUES($1, $2, $3)
	ON CONFLICT(name) DO UPDATE
	SET keys = excluded.keys, properties = excluded.properties
	RETURNING id, name, keys, properties
`
	stmt, err := conn.Conn().Prepare(ctx, "update_integration", query)
	if err != nil {
		return models.Integration{}, err
	}
	rows, err := conn.Conn().Query(ctx, stmt.Name, name, keys, properties)
	if err != nil {
		return models.Integration{}, err
	}
	defer rows.Close()
	integrations, err := pgx.CollectRows(rows, pgx.RowToStructByPos[models.Integration])
	if err != nil {
		return models.Integration{}, err
	}
	if len(integrations) == 0 {
		return models.Integration{}, err
	}
	return integrations[0], nil
}

// User Functions
func CreateUser(username string, userType string, hashedPassword string, fullName string, subscription bool, avatarId int) (models.User, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()

	conn, err := MetadataDbClient.Client.Acquire(ctx)
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

	var userId int
	rows, err := conn.Conn().Query(ctx, stmt.Name, username, hashedPassword, userType, alreadyLoggedIn, createdAt, avatarId, fullName, subscription, skipGetStarted)
	if err != nil {
		return models.User{}, err
	}
	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&userId)
		if err != nil {
			return models.User{}, err
		}
	}

	if err := rows.Err(); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Detail != "" {
				if strings.Contains(pgErr.Detail, "already exists") {
					return models.User{}, errors.New("User " + username + " already exists")
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

	newUser := models.User{
		ID:              userId,
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
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()
	query := `UPDATE users SET password = $2 WHERE username = $1`
	stmt, err := conn.Conn().Prepare(ctx, "change_user_password", query)
	if err != nil {
		return err
	}
	_, err = conn.Conn().Query(ctx, stmt.Name, username, hashedPassword)
	if err != nil {
		return err
	}
	return nil
}

func GetRootUser() (bool, models.User, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return false, models.User{}, err
	}
	defer conn.Release()
	query := `SELECT * FROM users WHERE type = 'root' LIMIT 1`
	stmt, err := conn.Conn().Prepare(ctx, "get_root_user", query)
	if err != nil {
		return false, models.User{}, err
	}
	rows, err := conn.Conn().Query(ctx, stmt.Name)
	if err != nil {
		return false, models.User{}, err
	}
	defer rows.Close()
	users, err := pgx.CollectRows(rows, pgx.RowToStructByPos[models.User])
	if err != nil {
		return false, models.User{}, err
	}
	if len(users) == 0 {
		return false, models.User{}, nil
	}
	return true, users[0], nil
}

func GetUserByUsername(username string) (bool, models.User, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return false, models.User{}, err
	}
	defer conn.Release()
	query := `SELECT * FROM users WHERE username = $1 LIMIT 1`
	stmt, err := conn.Conn().Prepare(ctx, "get_user_by_username", query)
	if err != nil {
		return false, models.User{}, err
	}
	rows, err := conn.Conn().Query(ctx, stmt.Name, username)
	if err != nil {
		return false, models.User{}, err
	}
	defer rows.Close()
	users, err := pgx.CollectRows(rows, pgx.RowToStructByPos[models.User])
	if err != nil {
		return false, models.User{}, err
	}
	if len(users) == 0 {
		return false, models.User{}, nil
	}
	return true, users[0], nil
}

func GetUserByUserId(userId int) (bool, models.User, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return false, models.User{}, err
	}
	defer conn.Release()
	query := `SELECT * FROM users WHERE id = $1 LIMIT 1`
	stmt, err := conn.Conn().Prepare(ctx, "get_user_by_id", query)
	if err != nil {
		return false, models.User{}, err
	}
	rows, err := conn.Conn().Query(ctx, stmt.Name, userId)
	if err != nil {
		return false, models.User{}, err
	}
	defer rows.Close()
	users, err := pgx.CollectRows(rows, pgx.RowToStructByPos[models.User])
	if err != nil {
		return false, models.User{}, err
	}
	if len(users) == 0 {
		return false, models.User{}, nil
	}
	return true, users[0], nil
}

func GetAllUsers() ([]models.FilteredGenericUser, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return []models.FilteredGenericUser{}, err
	}
	defer conn.Release()
	query := `SELECT s.id, s.username, s.type, s.created_at, s.already_logged_in, s.avatar_id FROM users AS s`
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
	if err != nil {
		return []models.FilteredGenericUser{}, err
	}
	if len(users) == 0 {
		return []models.FilteredGenericUser{}, nil
	}
	return users, nil
}

func GetAllApplicationUsers() ([]models.FilteredApplicationUser, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
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
	if err != nil {
		return []models.FilteredApplicationUser{}, err
	}
	if len(users) == 0 {
		return []models.FilteredApplicationUser{}, nil
	}
	return users, nil
}

func UpdateUserAlreadyLoggedIn(userId int) error {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()
	query := `UPDATE users SET already_logged_in = true WHERE id = $1`
	stmt, _ := conn.Conn().Prepare(ctx, "update_user_already_logged_in", query)
	conn.Conn().Query(ctx, stmt.Name, userId)
	return nil
}

func UpdateSkipGetStarted(username string) error {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()
	query := `UPDATE users SET skip_get_started = true WHERE username = $1`
	stmt, err := conn.Conn().Prepare(ctx, "update_skip_get_started", query)
	if err != nil {
		return err
	}
	_, err = conn.Conn().Query(ctx, stmt.Name, username)
	if err != nil {
		return err
	}
	return nil
}

func DeleteUser(username string) error {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()

	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	removeUserQuery := `DELETE FROM users WHERE username = $1`

	stmt, err := conn.Conn().Prepare(ctx, "remove_user", removeUserQuery)
	if err != nil {
		return err
	}

	_, err = conn.Conn().Exec(ctx, stmt.Name, username)
	if err != nil {
		return err
	}

	return nil
}

func EditAvatar(username string, avatarId int) error {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()
	query := `UPDATE users SET avatar_id = $2 WHERE username = $1`
	stmt, err := conn.Conn().Prepare(ctx, "edit_avatar", query)
	if err != nil {
		return err
	}
	_, err = conn.Conn().Query(ctx, stmt.Name, username, avatarId)
	if err != nil {
		return err
	}
	return nil
}

func GetAllActiveUsers() ([]models.FilteredUser, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return []models.FilteredUser{}, err
	}
	defer conn.Release()
	query := `
	SELECT DISTINCT u.username
	FROM users u
	JOIN stations s ON u.id = s.created_by
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
	if err != nil {
		return []models.FilteredUser{}, err
	}
	if len(userList) == 0 {
		return []models.FilteredUser{}, nil
	}
	return userList, nil
}

// Tags Functions
func InsertNewTag(name string, color string, stationArr []int, schemaArr []int, userArr []int) (models.Tag, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()

	conn, err := MetadataDbClient.Client.Acquire(ctx)
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

	stmt, err := conn.Conn().Prepare(ctx, "insert_new_tag", query)
	if err != nil {
		return models.Tag{}, err
	}

	var tagId int
	rows, err := conn.Conn().Query(ctx, stmt.Name, name, color, userArr, stationArr, schemaArr)
	if err != nil {
		return models.Tag{}, err
	}
	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&tagId)
		if err != nil {
			return models.Tag{}, err
		}
	}

	if err := rows.Err(); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Detail != "" {
				if strings.Contains(pgErr.Detail, "already exists") {
					return models.Tag{}, errors.New("Tag" + name + " already exists")
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

	newTag := models.Tag{
		ID:       tagId,
		Name:     name,
		Color:    color,
		Stations: stationArr,
		Schemas:  schemaArr,
		Users:    userArr,
	}
	return newTag, nil

}

func InsertEntityToTag(tagName string, entity string, entity_id int) error {
	var entityDBList string
	switch entity {
	case "station":
		entityDBList = "stations"
	case "schema":
		entityDBList = "schemas"
	case "user":
		entityDBList = "users"
	}
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()
	query := `UPDATE tags SET ` + entityDBList + ` = ARRAY_APPEND(` + entityDBList + `, $1) WHERE name = $2`
	stmt, err := conn.Conn().Prepare(ctx, "insert_entity_to_tag", query)
	if err != nil {
		return err
	}
	_, err = conn.Conn().Query(ctx, stmt.Name, entity_id, tagName)
	if err != nil {
		return err
	}
	return nil
}

func RemoveAllTagsFromEntity(entity string, entity_id int) error {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()
	query := `UPDATE tags SET ` + entity + ` = ARRAY_REMOVE(` + entity + `, $1) WHERE $1 = ANY(` + entity + `)`
	stmt, err := conn.Conn().Prepare(ctx, "remove_all_tags_from_entity", query)
	if err != nil {
		return err
	}
	_, err = conn.Conn().Query(ctx, stmt.Name, entity_id)
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
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()
	query := `UPDATE tags SET ` + entityDBList + ` = ARRAY_REMOVE(` + entityDBList + `, $2) WHERE name = $1`
	stmt, err := conn.Conn().Prepare(ctx, "remove_tag_from_entity", query)
	if err != nil {
		return err
	}
	_, err = conn.Conn().Query(ctx, stmt.Name, tagName, entity_id)
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
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return []models.Tag{}, err
	}
	defer conn.Release()
	uid, err := uuid.NewV4()
	if err != nil {
		return []models.Tag{}, err
	}
	query := `SELECT * FROM tags AS t WHERE $1 = ANY(t.` + entityDBList + `)`
	stmt, err := conn.Conn().Prepare(ctx, "get_tags_by_entity_id"+uid.String(), query)
	if err != nil {
		return []models.Tag{}, err
	}
	rows, err := conn.Conn().Query(ctx, stmt.Name, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	tags, err := pgx.CollectRows(rows, pgx.RowToStructByPos[models.Tag])
	if err != nil {
		return []models.Tag{}, err
	}
	if len(tags) == 0 {
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

	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
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
	if err != nil {
		return []models.Tag{}, err
	}
	if len(tags) == 0 {
		return []models.Tag{}, nil
	}
	return tags, nil
}

func GetAllUsedTags() ([]models.Tag, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
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
	if err != nil {
		return []models.Tag{}, err
	}
	if len(tags) == 0 {
		return []models.Tag{}, nil
	}
	return tags, nil
}

func GetTagByName(name string) (bool, models.Tag, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return false, models.Tag{}, err
	}
	defer conn.Release()
	query := `SELECT * FROM tags WHERE name=$1 LIMIT 1`
	stmt, err := conn.Conn().Prepare(ctx, "get_tag_by_name", query)
	if err != nil {
		return false, models.Tag{}, err
	}
	rows, err := conn.Conn().Query(ctx, stmt.Name, name)
	if err != nil {
		return false, models.Tag{}, err
	}
	defer rows.Close()
	tags, err := pgx.CollectRows(rows, pgx.RowToStructByPos[models.Tag])
	if err != nil {
		return false, models.Tag{}, err
	}
	if len(tags) == 0 {
		return false, models.Tag{}, nil
	}
	return true, tags[0], nil
}

// Sandbox Functions
// func InsertNewSanboxUser(username string, email string, firstName string, lastName string, profilePic string) (models.SandboxUser, error) {
// user := models.SandboxUser{}
// return user, nil
// }

// func UpdateSandboxUserAlreadyLoggedIn(userId int) {
// sandboxUsersCollection.UpdateOne(context.TODO(),
// 	bson.M{"_id": userId},
// 	bson.M{"$set": bson.M{"already_logged_in": true}},
// )
// }

// func GetSandboxUser(username string) (bool, models.SandboxUser, error) {
// 	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
// 	defer cancelfunc()
// 	conn, err := MetadataDbClient.Client.Acquire(ctx)
// 	if err != nil {
// 		return true, models.SandboxUser{}, err
// 	}
// 	defer conn.Release()
// 	query := `SELECT * FROM sandbox_users WHERE username = $1 LIMIT 1`
// 	stmt, err := conn.Conn().Prepare(ctx, "get_sandbox_user", query)
// 	if err != nil {
// 		return true, models.SandboxUser{}, err
// 	}
// 	rows, err := conn.Conn().Query(ctx, stmt.Name, username)
// 	if err != nil {
// 		return true, models.SandboxUser{}, err
// 	}
// 	defer rows.Close()
// 	users, err := pgx.CollectRows(rows, pgx.RowToStructByPos[models.SandboxUser])
// 	if err != nil {
// 		return true, models.SandboxUser{}, err
// 	}
// 	if len(users) == 0 {
// 		return false, models.SandboxUser{}, nil
// 	}
// 	return true, users[0], nil
// }

// func UpdateSkipGetStartedSandbox(username string) error {
// _, err := sandboxUsersCollection.UpdateOne(context.TODO(),
// 	bson.M{"username": username},
// 	bson.M{"$set": bson.M{"skip_get_started": true}},
// )
// if err != nil {
// 	return err
// }
// 	return nil
// }

// Image Functions
func InsertImage(name string, base64Encoding string) error {
	err := InsertConfiguration(name, base64Encoding)
	if err != nil {
		return err
	}
	return nil
}

func DeleteImage(name string) error {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()

	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	removeImageQuery := `DELETE FROM configurations
	WHERE key = $1`

	stmt, err := conn.Conn().Prepare(ctx, "remove_image", removeImageQuery)
	if err != nil {
		return err
	}

	_, err = conn.Conn().Exec(ctx, stmt.Name, name)
	if err != nil {
		return err
	}
	return nil
}

func GetImage(name string) (bool, models.Image, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return false, models.Image{}, err
	}
	defer conn.Release()
	query := `SELECT * FROM configurations WHERE key = $1 LIMIT 1`
	stmt, err := conn.Conn().Prepare(ctx, "get_image", query)
	if err != nil {
		return false, models.Image{}, err
	}
	rows, err := conn.Conn().Query(ctx, stmt.Name, name)
	if err != nil {
		return false, models.Image{}, err
	}
	defer rows.Close()
	images, err := pgx.CollectRows(rows, pgx.RowToStructByPos[models.Image])
	if err != nil {
		return false, models.Image{}, err
	}
	if len(images) == 0 {
		return false, models.Image{}, nil
	}
	return true, images[0], nil
}

// dls Functions
func InsertPoisonedCgMessages(stationId int, messageSeq int, producerId int, poisonedCgs []string, messageDetails models.MessagePayloadPg, updatedAt time.Time, messageType string) (models.DlsMessagePg, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()

	connection, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return models.DlsMessagePg{}, err
	}
	defer connection.Release()

	query := `INSERT INTO dls_messages( 
			station_id,
			message_seq,
			producer_id,
			poisoned_cgs,
			message_details,
			updated_at,
			message_type
			) 
		VALUES($1, $2, $3, $4, $5, $6, $7)
		RETURNING id`

	stmt, err := connection.Conn().Prepare(ctx, "insert_dls_messages", query)
	if err != nil {
		return models.DlsMessagePg{}, err
	}

	rows, err := connection.Conn().Query(ctx, stmt.Name, stationId, messageSeq, producerId, poisonedCgs, messageDetails, updatedAt, messageType)
	if err != nil {
		return models.DlsMessagePg{}, err
	}
	defer rows.Close()
	var messagePaylodId int
	for rows.Next() {
		err := rows.Scan(&messagePaylodId)
		if err != nil {
			return models.DlsMessagePg{}, err
		}
	}

	if err != nil {
		return models.DlsMessagePg{}, err
	}

	msgDetails := models.MessagePayloadDlsPg{
		TimeSent: messageDetails.TimeSent,
		Size:     messageDetails.Size,
		Data:     messageDetails.Data,
		// Headers:  messageDetails.headersJson,
	}

	deadLetterPayload := models.DlsMessagePg{
		ID:             messagePaylodId,
		StationId:      stationId,
		MessageSeq:     messageSeq,
		ProducerId:     producerId,
		PoisonedCgs:    poisonedCgs,
		MessageDetails: msgDetails,
		UpdatedAt:      updatedAt,
	}

	if err := rows.Err(); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Detail != "" {
				if !strings.Contains(pgErr.Detail, "already exists") {
					return models.DlsMessagePg{}, errors.New("messages table already exists")
				} else {
					return models.DlsMessagePg{}, errors.New(pgErr.Detail)
				}
			} else {
				return models.DlsMessagePg{}, errors.New(pgErr.Message)
			}
		} else {
			return models.DlsMessagePg{}, err
		}
	}

	return deadLetterPayload, nil
}

func GetMsgByStationIdAndMsgSeq(stationId, messageSeq int) (bool, models.DlsMessagePg, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()

	connection, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return false, models.DlsMessagePg{}, err
	}
	defer connection.Release()

	query := `SELECT * FROM dls_messages WHERE station_id = $1 AND message_seq = $2 LIMIT 1`

	stmt, err := connection.Conn().Prepare(ctx, "get_dls_messages_by_station_id_and_message_seq", query)
	if err != nil {
		return false, models.DlsMessagePg{}, err
	}

	rows, err := connection.Conn().Query(ctx, stmt.Name, stationId, messageSeq)
	if err != nil {
		return false, models.DlsMessagePg{}, err
	}
	defer rows.Close()

	message, err := pgx.CollectRows(rows, pgx.RowToStructByPos[models.DlsMessagePg])
	if err != nil {
		return false, models.DlsMessagePg{}, err
	}
	if len(message) == 0 {
		return false, models.DlsMessagePg{}, nil
	}

	return true, message[0], nil

}

func UpdatePoisonCgsInDlsMessage(poisonedCgs string, stationId, messageSeq int, updatedAt time.Time) error {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	query := `UPDATE dls_messages SET poisoned_cgs = ARRAY_APPEND(poisoned_cgs, $1), updated_at = $4 WHERE station_id=$2 AND message_seq=$3`
	stmt, err := conn.Conn().Prepare(ctx, "update_poisoned_cgs", query)
	if err != nil {
		return err
	}
	_, err = conn.Conn().Query(ctx, stmt.Name, poisonedCgs, stationId, messageSeq, updatedAt)
	if err != nil {
		return err
	}
	return nil

}

func GetTotalPoisonMsgsPerCg(cgName string) (int, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return 0, err
	}
	defer conn.Release()
	query := `SELECT COUNT(*) FROM dls_messages WHERE $1 = ANY(poisoned_cgs)`
	stmt, err := conn.Conn().Prepare(ctx, "get_total_poison_msgs_per_cg", query)
	if err != nil {
		return 0, err
	}
	var count int
	err = conn.Conn().QueryRow(ctx, stmt.Name, cgName).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}
