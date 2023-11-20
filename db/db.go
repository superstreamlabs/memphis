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

	"strings"

	"github.com/memphisdev/memphis/conf"

	"github.com/memphisdev/memphis/models"

	"context"
	"time"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var configuration = conf.GetConfig()

var MetadataDbClient MetadataStorage

const (
	DbOperationTimeout = 40
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

	alterTenantsTable := `
	DO $$
	BEGIN
		IF EXISTS (
			SELECT 1 FROM information_schema.tables WHERE table_name = 'tenants' AND table_schema = 'public'
		) THEN
			ALTER TABLE tenants ADD COLUMN IF NOT EXISTS firebase_organization_id VARCHAR NOT NULL DEFAULT '';
			ALTER TABLE tenants ADD COLUMN IF NOT EXISTS organization_name VARCHAR NOT NULL DEFAULT '';
			UPDATE tenants SET organization_name = name WHERE organization_name = ''; 
		END IF;
	END $$;`

	tenantsTable := `CREATE TABLE IF NOT EXISTS tenants(
		id SERIAL NOT NULL,
		name VARCHAR NOT NULL UNIQUE DEFAULT '$memphis',
		firebase_organization_id VARCHAR NOT NULL DEFAULT '',
		internal_ws_pass VARCHAR NOT NULL,
		organization_name VARCHAR NOT NULL DEFAULT '',
		PRIMARY KEY (id));`

	alterAuditLogsTable := `
	DO $$
	BEGIN
		IF EXISTS (
			SELECT 1 FROM information_schema.tables WHERE table_name = 'audit_logs' AND table_schema = 'public'
		) THEN
			ALTER TABLE audit_logs ADD COLUMN IF NOT EXISTS tenant_name VARCHAR NOT NULL DEFAULT '$memphis';
			DROP INDEX IF EXISTS station_name;
			CREATE INDEX audit_logs_station_tenant_name ON audit_logs (station_name, tenant_name);
		END IF;
	END $$;`

	auditLogsTable := `CREATE TABLE IF NOT EXISTS audit_logs(
		id SERIAL NOT NULL,
		station_name VARCHAR NOT NULL,
		message TEXT NOT NULL,
		created_by INTEGER NOT NULL,
		created_by_username VARCHAR NOT NULL,
		created_at TIMESTAMPTZ NOT NULL,
		tenant_name VARCHAR NOT NULL DEFAULT '$memphis',
		PRIMARY KEY (id));
	CREATE INDEX IF NOT EXISTS station_name ON audit_logs (station_name, tenant_name);`

	alterUsersTable := `
	DO $$
	BEGIN
		IF EXISTS (
			SELECT 1 FROM information_schema.tables WHERE table_name = 'users' AND table_schema = 'public'
		) THEN
		ALTER TABLE users ADD COLUMN IF NOT EXISTS tenant_name VARCHAR NOT NULL DEFAULT '$memphis';
		ALTER TABLE users ADD COLUMN IF NOT EXISTS pending BOOL NOT NULL DEFAULT false;
		ALTER TABLE users ADD COLUMN IF NOT EXISTS team VARCHAR NOT NULL DEFAULT '';
		ALTER TABLE users ADD COLUMN IF NOT EXISTS position VARCHAR NOT NULL DEFAULT '';
		ALTER TABLE users ADD COLUMN IF NOT EXISTS owner VARCHAR NOT NULL DEFAULT '';
		ALTER TABLE users ADD COLUMN IF NOT EXISTS description VARCHAR NOT NULL DEFAULT '';
		ALTER TABLE users DROP CONSTRAINT IF EXISTS users_username_key;
		ALTER TABLE users DROP CONSTRAINT IF EXISTS users_username_tenant_name_key;
		ALTER TABLE users ADD CONSTRAINT users_username_tenant_name_key UNIQUE(username, tenant_name);
		END IF;
	END $$;`

	usersTable := `
	CREATE TYPE enum AS ENUM ('root', 'management', 'application');
	CREATE TABLE IF NOT EXISTS users(
		id SERIAL NOT NULL,
		username VARCHAR NOT NULL,
		password TEXT NOT NULL,
		type enum NOT NULL DEFAULT 'root',
		already_logged_in BOOL NOT NULL DEFAULT false,
		created_at TIMESTAMPTZ NOT NULL,
		avatar_id SERIAL NOT NULL,
		full_name VARCHAR,
		subscription BOOL NOT NULL DEFAULT false,
		skip_get_started BOOL NOT NULL DEFAULT false,
		tenant_name VARCHAR NOT NULL DEFAULT '$memphis',
		pending BOOL NOT NULL DEFAULT false,
		team VARCHAR NOT NULL DEFAULT '',
		position VARCHAR NOT NULL DEFAULT '',
		owner VARCHAR NOT NULL DEFAULT '',
		description VARCHAR NOT NULL DEFAULT '',
		PRIMARY KEY (id),
		CONSTRAINT fk_tenant_name
			FOREIGN KEY(tenant_name)
			REFERENCES tenants(name),
		UNIQUE(username, tenant_name)
		);`

	alterConfigurationsTable := `
	DO $$
	BEGIN
		IF EXISTS (
			SELECT 1 FROM information_schema.tables WHERE table_name = 'configurations' AND table_schema = 'public'
		) THEN
		ALTER TABLE configurations ADD COLUMN IF NOT EXISTS tenant_name VARCHAR NOT NULL DEFAULT '$memphis';
		ALTER TABLE configurations DROP CONSTRAINT IF EXISTS configurations_key_key;
		ALTER TABLE configurations DROP CONSTRAINT IF EXISTS key_tenant_name;
		ALTER TABLE configurations ADD CONSTRAINT key_tenant_name UNIQUE(key, tenant_name);
		END IF;
	END $$;`

	configurationsTable := `CREATE TABLE IF NOT EXISTS configurations(
		id SERIAL NOT NULL,
		key VARCHAR NOT NULL,
		value TEXT NOT NULL,
		tenant_name VARCHAR NOT NULL DEFAULT '$memphis',
		PRIMARY KEY (id),
		UNIQUE(key, tenant_name),
		CONSTRAINT fk_tenant_name_configurations
			FOREIGN KEY(tenant_name)
			REFERENCES tenants(name)
		);`

	alterConnectionsTable := `DROP TABLE IF EXISTS connections;`

	alterIntegrationsTable := `
	DO $$
	BEGIN
		IF EXISTS (
			SELECT 1 FROM information_schema.tables WHERE table_name = 'integrations' AND table_schema = 'public'
		) THEN
		ALTER TABLE integrations ADD COLUMN IF NOT EXISTS tenant_name VARCHAR NOT NULL DEFAULT '$memphis';
		ALTER TABLE integrations ADD COLUMN IF NOT EXISTS is_valid BOOL NOT NULL DEFAULT true;
	ALTER TABLE integrations DROP CONSTRAINT IF EXISTS integrations_name_key;
	ALTER TABLE integrations DROP CONSTRAINT IF EXISTS tenant_name_name;
	ALTER TABLE integrations ADD CONSTRAINT tenant_name_name UNIQUE(name, tenant_name);
		END IF;
	END $$;`

	integrationsTable := `CREATE TABLE IF NOT EXISTS integrations(
		id SERIAL NOT NULL,
		name VARCHAR NOT NULL,
		keys JSON NOT NULL DEFAULT '{}',
		properties JSON NOT NULL DEFAULT '{}',
		tenant_name VARCHAR NOT NULL DEFAULT '$memphis',
		is_valid BOOL NOT NULL DEFAULT true,
		PRIMARY KEY (id),
		CONSTRAINT fk_tenant_name
			FOREIGN KEY(tenant_name)
			REFERENCES tenants(name),
		UNIQUE(name, tenant_name)
		);`

	alterSchemasTable := `
	DO $$
	BEGIN
		IF EXISTS (
			SELECT 1 FROM information_schema.tables WHERE table_name = 'schemas' AND table_schema = 'public'
		) THEN
		ALTER TABLE schemas ADD COLUMN IF NOT EXISTS tenant_name VARCHAR NOT NULL DEFAULT '$memphis';
		ALTER TABLE schemas DROP CONSTRAINT IF EXISTS name;
		ALTER TABLE schemas DROP CONSTRAINT IF EXISTS schemas_name_tenant_name_key;
		ALTER TABLE schemas ADD CONSTRAINT schemas_name_tenant_name_key UNIQUE(name, tenant_name);
		ALTER TYPE enum_type ADD VALUE 'avro';
		END IF;
	END $$;`

	schemasTable := `
	CREATE TYPE enum_type AS ENUM ('json', 'graphql', 'protobuf', 'avro');
	CREATE TABLE IF NOT EXISTS schemas(
		id SERIAL NOT NULL,
		name VARCHAR NOT NULL,
		type enum_type NOT NULL DEFAULT 'protobuf',
		created_by_username VARCHAR NOT NULL,
		tenant_name VARCHAR NOT NULL DEFAULT '$memphis',
		PRIMARY KEY (id),
		CONSTRAINT fk_tenant_name_schemas
			FOREIGN KEY(tenant_name)
			REFERENCES tenants(name),
		UNIQUE(name, tenant_name)
		);
		CREATE INDEX IF NOT EXISTS name ON schemas (name);`

	alterTagsTable := `
	DO $$
	BEGIN
		IF EXISTS (
			SELECT 1 FROM information_schema.tables WHERE table_name = 'tags' AND table_schema = 'public'
		) THEN
		ALTER TABLE tags ADD COLUMN IF NOT EXISTS tenant_name VARCHAR NOT NULL DEFAULT '$memphis';
		ALTER TABLE tags DROP CONSTRAINT IF EXISTS name;
		ALTER TABLE tags DROP CONSTRAINT IF EXISTS tags_name_tenant_name_key;
		ALTER TABLE tags ADD CONSTRAINT tags_name_tenant_name_key UNIQUE(name, tenant_name);
		END IF;
	END $$;`

	tagsTable := `CREATE TABLE IF NOT EXISTS tags(
		id SERIAL NOT NULL,
		name VARCHAR NOT NULL,
		color VARCHAR NOT NULL,
		users INTEGER[],
		stations INTEGER[],
		schemas INTEGER[],
		tenant_name VARCHAR NOT NULL DEFAULT '$memphis',
		PRIMARY KEY (id),
		CONSTRAINT fk_tenant_name_tags
			FOREIGN KEY(tenant_name)
			REFERENCES tenants(name),
		UNIQUE(name, tenant_name)
		);
		CREATE INDEX IF NOT EXISTS name_tag ON tags (name);`

	alterConsumersTable := `
	DO $$
	BEGIN
		IF EXISTS (
			SELECT 1 FROM information_schema.tables WHERE table_name = 'consumers' AND table_schema = 'public'
		) THEN
			IF EXISTS (
				SELECT 1
				FROM information_schema.columns
				WHERE table_name = 'consumers'
				AND column_name = 'is_deleted'
			) THEN
				DELETE FROM consumers WHERE is_deleted = true;
			END IF;
			ALTER TABLE consumers DROP COLUMN IF EXISTS created_by;
			ALTER TABLE consumers DROP COLUMN IF EXISTS created_by_username;
			ALTER TABLE consumers DROP COLUMN IF EXISTS is_deleted;
			ALTER TABLE consumers ADD COLUMN IF NOT EXISTS tenant_name VARCHAR NOT NULL DEFAULT '$memphis';
			DROP INDEX IF EXISTS unique_consumer_table;
			ALTER TABLE consumers DROP CONSTRAINT IF EXISTS fk_connection_id;
			CREATE INDEX IF NOT EXISTS consumer_tenant_name ON consumers(tenant_name);
			CREATE INDEX IF NOT EXISTS consumer_connection_id ON consumers(connection_id);
			ALTER TABLE consumers ADD COLUMN IF NOT EXISTS partitions INTEGER[];
			ALTER TABLE consumers ADD COLUMN IF NOT EXISTS version INTEGER NOT NULL DEFAULT 2;
			ALTER TABLE consumers ADD COLUMN IF NOT EXISTS sdk VARCHAR NOT NULL DEFAULT 'unknown';
			ALTER TABLE consumers ADD COLUMN IF NOT EXISTS app_id VARCHAR NOT NULL DEFAULT 'unknown';
			UPDATE consumers SET app_id = connection_id WHERE app_id = 'unknown';
			IF EXISTS (
				SELECT 1
				FROM information_schema.columns
				WHERE table_name = 'consumers'
				AND column_name = 'created_at'
			) THEN
				ALTER TABLE consumers RENAME COLUMN created_at TO updated_at;
			END IF;
		END IF;
		
	END $$;`

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
		is_active BOOL NOT NULL DEFAULT true,
		updated_at TIMESTAMPTZ NOT NULL,
		max_msg_deliveries SERIAL NOT NULL,
		start_consume_from_seq SERIAL NOT NULL,
		last_msgs SERIAL NOT NULL,
		tenant_name VARCHAR NOT NULL DEFAULT '$memphis',
		partitions INTEGER[],
		version INTEGER NOT NULL DEFAULT 2,
		sdk VARCHAR NOT NULL DEFAULT 'unknown',
		app_id VARCHAR NOT NULL,
		PRIMARY KEY (id),
		CONSTRAINT fk_station_id
			FOREIGN KEY(station_id)
			REFERENCES stations(id),
		CONSTRAINT fk_tenant_name_consumers
			FOREIGN KEY(tenant_name)
			REFERENCES tenants(name)
		);
		CREATE INDEX IF NOT EXISTS station_id ON consumers (station_id);
		CREATE INDEX IF NOT EXISTS consumer_name ON consumers (name);
		CREATE INDEX IF NOT EXISTS consumer_tenant_name ON consumers(tenant_name);
		CREATE INDEX IF NOT EXISTS consumer_connection_id ON consumers(connection_id);`

	alterStationsTable := `
	DO $$
	BEGIN
		IF EXISTS (
			SELECT 1 FROM information_schema.tables WHERE table_name = 'stations' AND table_schema = 'public'
		) THEN
		ALTER TYPE enum_retention_type ADD VALUE IF NOT EXISTS 'ack_based';
		ALTER TABLE stations ADD COLUMN IF NOT EXISTS tenant_name VARCHAR NOT NULL DEFAULT '$memphis';
		ALTER TABLE stations ADD COLUMN IF NOT EXISTS resend_disabled BOOL NOT NULL DEFAULT false;
		ALTER TABLE stations ADD COLUMN IF NOT EXISTS partitions INTEGER[];
		ALTER TABLE stations ADD COLUMN IF NOT EXISTS version INTEGER NOT NULL DEFAULT 0;
		ALTER TABLE stations ADD COLUMN IF NOT EXISTS dls_station VARCHAR NOT NULL DEFAULT '';
		ALTER TABLE stations ADD COLUMN IF NOT EXISTS functions_lock_held BOOL NOT NULL DEFAULT false;
		ALTER TABLE stations ADD COLUMN IF NOT EXISTS functions_locked_at TIMESTAMPTZ NOT NULL DEFAULT NOW();
		DROP INDEX IF EXISTS unique_station_name_deleted;
		CREATE UNIQUE INDEX unique_station_name_deleted ON stations(name, is_deleted, tenant_name) WHERE is_deleted = false;
		END IF;
	END $$;`

	stationsTable := `
	CREATE TYPE enum_retention_type AS ENUM ('message_age_sec', 'messages', 'bytes', 'ack_based');
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
		created_at TIMESTAMPTZ NOT NULL,
		updated_at TIMESTAMPTZ NOT NULL,
		is_deleted BOOL NOT NULL,
		schema_name VARCHAR,
		schema_version_number SERIAL,
		idempotency_window_ms SERIAL NOT NULL,
		is_native BOOL NOT NULL ,
		dls_configuration_poison BOOL NOT NULL DEFAULT true,
		dls_configuration_schemaverse BOOL NOT NULL DEFAULT true,
		tiered_storage_enabled BOOL NOT NULL,
		tenant_name VARCHAR NOT NULL DEFAULT '$memphis',
		resend_disabled BOOL NOT NULL DEFAULT false,
		partitions INTEGER[],
		version INTEGER NOT NULL DEFAULT 0,
		dls_station VARCHAR NOT NULL DEFAULT '',
		functions_lock_held BOOL NOT NULL DEFAULT false,
		functions_locked_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		PRIMARY KEY (id),
		CONSTRAINT fk_tenant_name_stations
			FOREIGN KEY(tenant_name)
			REFERENCES tenants(name)
		);
		CREATE UNIQUE INDEX IF NOT EXISTS unique_station_name_deleted ON stations(name, is_deleted, tenant_name) WHERE is_deleted = false;`

	alterSchemaVerseTable := `
	DO $$
	BEGIN
		IF EXISTS (
			SELECT 1 FROM information_schema.tables WHERE table_name = 'schema_versions' AND table_schema = 'public'
		) THEN
		ALTER TABLE schema_versions ADD COLUMN IF NOT EXISTS tenant_name VARCHAR NOT NULL DEFAULT '$memphis';
		END IF;
	END $$;`

	schemaVersionsTable := `CREATE TABLE IF NOT EXISTS schema_versions(
		id SERIAL NOT NULL,
		version_number SERIAL NOT NULL,
		active BOOL NOT NULL DEFAULT false,
		created_by INTEGER NOT NULL,
		created_by_username VARCHAR NOT NULL,
		created_at TIMESTAMPTZ NOT NULL,
		schema_content TEXT NOT NULL,
		schema_id INTEGER NOT NULL,
		msg_struct_name VARCHAR DEFAULT '',
		descriptor bytea,
		tenant_name VARCHAR NOT NULL DEFAULT '$memphis',
		PRIMARY KEY (id),
		UNIQUE(version_number, schema_id),
		CONSTRAINT fk_schema_id
			FOREIGN KEY(schema_id)
			REFERENCES schemas(id),
		CONSTRAINT fk_tenant_name_schemaverse
			FOREIGN KEY(tenant_name)
			REFERENCES tenants(name)
		);`

	alterProducersTable := `
	DO $$
	BEGIN
		IF EXISTS (
			SELECT 1 FROM information_schema.tables WHERE table_name = 'producers' AND table_schema = 'public'
		) THEN
			IF EXISTS (
				SELECT 1
				FROM information_schema.columns
				WHERE table_name = 'producers'
				AND column_name = 'is_deleted'
			) THEN
				DELETE FROM producers WHERE is_deleted = true;
			END IF;
			ALTER TABLE producers ADD COLUMN IF NOT EXISTS tenant_name VARCHAR NOT NULL DEFAULT '$memphis';
			ALTER TABLE producers ADD COLUMN IF NOT EXISTS connection_id INTEGER NOT NULL;
			ALTER TABLE producers DROP COLUMN IF EXISTS created_by;
			ALTER TABLE producers DROP COLUMN IF EXISTS created_by_username;
			ALTER TABLE producers DROP COLUMN IF EXISTS is_deleted;
			ALTER TABLE producers DROP CONSTRAINT IF EXISTS fk_connection_id;
			DROP INDEX IF EXISTS unique_producer_table;
			DROP INDEX IF EXISTS producer_connection_id;
			CREATE INDEX IF NOT EXISTS producer_name ON producers(name);
			CREATE INDEX IF NOT EXISTS producer_tenant_name ON producers(tenant_name);
			CREATE INDEX IF NOT EXISTS producer_connection_id ON producers(connection_id);
			ALTER TABLE producers ADD COLUMN IF NOT EXISTS partitions INTEGER[];
			ALTER TABLE producers ADD COLUMN IF NOT EXISTS version INTEGER NOT NULL DEFAULT 2;
			ALTER TABLE producers ADD COLUMN IF NOT EXISTS sdk VARCHAR NOT NULL DEFAULT 'unknown';
			ALTER TABLE producers ADD COLUMN IF NOT EXISTS app_id VARCHAR NOT NULL DEFAULT 'unknown';
			UPDATE producers SET app_id = connection_id WHERE app_id = 'unknown';
			IF EXISTS (
				SELECT 1
				FROM information_schema.columns
				WHERE table_name = 'producers'
				AND column_name = 'created_at'
			) THEN
				ALTER TABLE producers RENAME COLUMN created_at TO updated_at;
			END IF;
		END IF;
	END $$;`

	producersTable := `
	CREATE TYPE enum_producer_type AS ENUM ('application', 'connector');
	CREATE TABLE IF NOT EXISTS producers(
		id SERIAL NOT NULL,
		name VARCHAR NOT NULL,
		station_id INTEGER NOT NULL,
		type enum_producer_type NOT NULL DEFAULT 'application',
		connection_id VARCHAR NOT NULL, 
		is_active BOOL NOT NULL DEFAULT true,
		updated_at TIMESTAMPTZ NOT NULL,
		tenant_name VARCHAR NOT NULL DEFAULT '$memphis',
		partitions INTEGER[],
		version INTEGER NOT NULL DEFAULT 2,
		sdk VARCHAR NOT NULL DEFAULT 'unknown',
		app_id VARCHAR NOT NULL,
		PRIMARY KEY (id),
		CONSTRAINT fk_station_id
			FOREIGN KEY(station_id)
			REFERENCES stations(id),
		CONSTRAINT fk_tenant_name_producers
			FOREIGN KEY(tenant_name)
			REFERENCES tenants(name)
		);
		CREATE INDEX IF NOT EXISTS producer_station_id ON producers(station_id);
		CREATE INDEX IF NOT EXISTS producer_name ON producers(name);
		CREATE INDEX IF NOT EXISTS producer_tenant_name ON producers(tenant_name);
		CREATE INDEX IF NOT EXISTS producer_connection_id ON producers(connection_id);`

	alterDlsMsgsTable := `DO $$
	BEGIN
		IF EXISTS (
			SELECT 1 FROM information_schema.tables WHERE table_name = 'dls_messages' AND table_schema = 'public'
		) THEN
			ALTER TABLE dls_messages ADD COLUMN IF NOT EXISTS tenant_name VARCHAR NOT NULL DEFAULT '$memphis';
			ALTER TABLE dls_messages ADD COLUMN IF NOT EXISTS producer_name VARCHAR NOT NULL DEFAULT '';
			ALTER TABLE dls_messages ADD COLUMN IF NOT EXISTS partition_number INTEGER NOT NULL DEFAULT -1;
			ALTER TABLE dls_messages ADD COLUMN IF NOT EXISTS attached_function_id INT NOT NULL DEFAULT -1;
			DROP INDEX IF EXISTS dls_producer_id;
			IF EXISTS (
				SELECT 1 FROM information_schema.columns WHERE table_name = 'dls_messages' AND column_name = 'producer_id'
			) THEN
				UPDATE dls_messages
				SET producer_name = p.name
				FROM producers p
				WHERE dls_messages.producer_id = p.id AND EXISTS (SELECT 1 FROM producers WHERE id = dls_messages.producer_id);
				ALTER TABLE dls_messages DROP COLUMN IF EXISTS producer_id;
			END IF;
		END IF;
	END $$;`

	// we remove fk_producer_id from dls_msgs table because dls msgs with nats compatibility
	// TODO: add fk_producer_id and solve dls messages nats compatibility
	// CONSTRAINT fk_producer_id
	// FOREIGN KEY(producer_id)
	// REFERENCES producers(id),

	dlsMessagesTable := `
	CREATE TABLE IF NOT EXISTS dls_messages(
		id SERIAL NOT NULL,    
		station_id INT NOT NULL,
		message_seq INT NOT NULL, 
		poisoned_cgs VARCHAR[],
		message_details JSON NOT NULL,    
		updated_at TIMESTAMPTZ NOT NULL,
		message_type VARCHAR NOT NULL,
		validation_error VARCHAR DEFAULT '',
		tenant_name VARCHAR NOT NULL DEFAULT '$memphis',
		producer_name VARCHAR NOT NULL,
		partition_number INTEGER NOT NULL DEFAULT -1,
		attached_function_id INT NOT NULL DEFAULT -1,
		PRIMARY KEY (id),
		CONSTRAINT fk_station_id
			FOREIGN KEY(station_id)
			REFERENCES stations(id),
		CONSTRAINT fk_tenant_name_dls_msgs
			FOREIGN KEY(tenant_name)
			REFERENCES tenants(name)
	);
	CREATE INDEX IF NOT EXISTS dls_station_id
		ON dls_messages(station_id);`

	asyncTasksTable := `
        CREATE TABLE IF NOT EXISTS async_tasks(
            id SERIAL NOT NULL,    
            name VARCHAR NOT NULL DEFAULT '',
            broker_in_charge VARCHAR NOT NULL DEFAULT 'memphis-0',
            created_at TIMESTAMPTZ NOT NULL,
            updated_at TIMESTAMPTZ NOT NULL,
            meta_data JSON NOT NULL DEFAULT '{}',
			tenant_name VARCHAR NOT NULL DEFAULT '$memphis',
			station_id INT NOT NULL,
			created_by VARCHAR NOT NULL,
			PRIMARY KEY (id)
        );`

	sharedLocksTable := `
		CREATE TABLE IF NOT EXISTS shared_locks(
			id SERIAL NOT NULL,
			name VARCHAR NOT NULL,
			tenant_name VARCHAR NOT NULL,
			lock_held BOOL NOT NULL DEFAULT false,
			locked_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			PRIMARY KEY (id),
			UNIQUE(name, tenant_name),
		CONSTRAINT fk_tenant_name_shared_locks
			FOREIGN KEY(tenant_name)
			REFERENCES tenants(name)
		);`

	alterAsyncTasks := `DO $$
	BEGIN
		IF EXISTS (
			SELECT 1 FROM information_schema.tables WHERE table_name = 'async_tasks' AND table_schema = 'public'
		) THEN
		ALTER TABLE async_tasks ADD COLUMN IF NOT EXISTS created_by VARCHAR NOT NULL;
		END IF;
		IF EXISTS (
			SELECT 1 FROM information_schema.table_constraints
			WHERE table_name = 'async_tasks' AND constraint_type = 'UNIQUE'
			AND constraint_name = 'async_tasks_name_tenant_name_station_id_key'
		) THEN
			ALTER TABLE async_tasks DROP CONSTRAINT async_tasks_name_tenant_name_station_id_key;
		END IF;

		IF NOT EXISTS (
			SELECT 1 FROM information_schema.table_constraints
			WHERE table_name = 'async_tasks' AND constraint_type = 'PRIMARY KEY'
		) THEN
			ALTER TABLE async_tasks ADD PRIMARY KEY (id);
		END IF;
	END $$;`

	db := MetadataDbClient.Client
	ctx := MetadataDbClient.Ctx

	tables := []string{alterTenantsTable, tenantsTable, alterUsersTable, usersTable, alterAuditLogsTable, auditLogsTable, alterConfigurationsTable, configurationsTable, alterIntegrationsTable, integrationsTable, alterSchemasTable, schemasTable, alterTagsTable, tagsTable, alterStationsTable, stationsTable, alterDlsMsgsTable, dlsMessagesTable, alterConsumersTable, consumersTable, alterSchemaVerseTable, schemaVersionsTable, alterProducersTable, producersTable, alterConnectionsTable, asyncTasksTable, alterAsyncTasks, testEventsTable, functionsTable, attachedFunctionsTable, sharedLocksTable, functionsEngineWorkersTable, scheduledFunctionWorkersTable}

	for _, table := range tables {
		_, err := db.Exec(ctx, table)
		if err != nil {
			var pgErr *pgconn.PgError
			errPg := errors.As(err, &pgErr)
			if errPg && !strings.Contains(pgErr.Message, "already exists") {
				return err
			}
		}
	}
	return nil
}

func InitalizeMetadataDbConnection() (MetadataStorage, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)

	defer cancelfunc()
	metadataDbUser := configuration.METADATA_DB_USER
	metadataDbPassword := configuration.METADATA_DB_PASS
	metadataDbName := configuration.METADATA_DB_DBNAME
	metadataDbHost := configuration.METADATA_DB_HOST
	metadataDbPort := configuration.METADATA_DB_PORT
	var metadataDbUrl string
	if configuration.METADATA_DB_TLS_ENABLED {
		metadataAuth := ""
		if !configuration.METADATA_DB_TLS_MUTUAL {
			metadataAuth = ":" + metadataDbPassword
		}
		metadataDbUrl = "postgres://" + metadataDbUser + metadataAuth + "@" + metadataDbHost + ":" + metadataDbPort + "/" + metadataDbName + "?sslmode=verify-full"
	} else {
		metadataDbUrl = "postgres://" + metadataDbUser + ":" + metadataDbPassword + "@" + metadataDbHost + ":" + metadataDbPort + "/" + metadataDbName + "?sslmode=prefer"
	}

	config, err := pgxpool.ParseConfig(metadataDbUrl)
	if err != nil {
		return MetadataStorage{}, err
	}
	config.MaxConns = int32(configuration.METADATA_DB_MAX_CONNS)

	if configuration.METADATA_DB_TLS_ENABLED {
		CACert, err := os.ReadFile(configuration.METADATA_DB_TLS_CA)
		if err != nil {
			return MetadataStorage{}, err
		}

		CACertPool := x509.NewCertPool()
		CACertPool.AppendCertsFromPEM(CACert)

		cert, err := tls.LoadX509KeyPair(configuration.METADATA_DB_TLS_CRT, configuration.METADATA_DB_TLS_KEY)
		if err != nil {
			return MetadataStorage{}, err
		}

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
	client := MetadataStorage{Ctx: ctx, Cancel: cancelfunc, Client: pool}
	err = createTables(client)
	if err != nil {
		return MetadataStorage{}, err
	}
	MetadataDbClient = MetadataStorage{Client: pool, Ctx: ctx, Cancel: cancelfunc}
	return MetadataDbClient, nil
}

// System Keys Functions
func GetSystemKey(key string, tenantName string) (bool, models.SystemKey, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return false, models.SystemKey{}, err
	}
	defer conn.Release()
	query := `SELECT * FROM configurations WHERE key = $1 AND tenant_name = $2 LIMIT 1`
	stmt, err := conn.Conn().Prepare(ctx, "get_system_key", query)
	if err != nil {
		return false, models.SystemKey{}, err
	}
	if tenantName != conf.GlobalAccount {
		tenantName = strings.ToLower(tenantName)
	}
	rows, err := conn.Conn().Query(ctx, stmt.Name, key, tenantName)
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

func InsertSystemKey(key string, value string, tenantName string) error {
	if tenantName != conf.GlobalAccount {
		tenantName = strings.ToLower(tenantName)
	}
	err := InsertConfiguration(key, value, tenantName)
	if err != nil {
		return err
	}
	return nil
}

func EditConfigurationValue(key string, value string, tenantName string) error {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()
	query := `UPDATE configurations SET value = $2 WHERE key = $1 AND tenant_name=$3`
	stmt, err := conn.Conn().Prepare(ctx, "edit_configuration_value", query)
	if err != nil {
		return err
	}
	if tenantName != conf.GlobalAccount {
		tenantName = strings.ToLower(tenantName)
	}
	_, err = conn.Conn().Query(ctx, stmt.Name, key, value, tenantName)
	if err != nil {
		return err
	}
	return nil
}

// Configuration Functions
func GetAllConfigurations() (bool, []models.ConfigurationsValue, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return false, []models.ConfigurationsValue{}, err
	}
	defer conn.Release()
	query := `SELECT * FROM configurations`
	stmt, err := conn.Conn().Prepare(ctx, "get_all_configurations", query)
	if err != nil {
		return false, []models.ConfigurationsValue{}, err
	}
	rows, err := conn.Conn().Query(ctx, stmt.Name)
	if err != nil {
		return false, []models.ConfigurationsValue{}, err
	}
	defer rows.Close()
	configurations, err := pgx.CollectRows(rows, pgx.RowToStructByPos[models.ConfigurationsValue])
	if err != nil {
		return false, []models.ConfigurationsValue{}, err
	}
	if len(configurations) == 0 {
		return false, []models.ConfigurationsValue{}, nil
	}

	return true, configurations, nil
}

func InsertConfiguration(key string, value string, tenantName string) error {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()

	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	query := `INSERT INTO configurations( 
			key, 
			value,
			tenant_name) 
		VALUES($1, $2, $3) 
		RETURNING id`

	stmt, err := conn.Conn().Prepare(ctx, "insert_new_configuration", query)
	if err != nil {
		return err
	}

	newConfiguration := models.ConfigurationsValue{}
	if tenantName != conf.GlobalAccount {
		tenantName = strings.ToLower(tenantName)
	}
	rows, err := conn.Conn().Query(ctx, stmt.Name,
		key, value, tenantName)
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

	if err := rows.Err(); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Detail != "" {
				if strings.Contains(pgErr.Detail, "already exists") {
					return errors.New("configuration " + key + " already exists")
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

func UpsertConfiguration(key string, value string, tenantName string) error {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()
	query := `INSERT INTO configurations (key, value, tenant_name) VALUES($1, $2, $3)
	ON CONFLICT(key, tenant_name) DO UPDATE SET value = EXCLUDED.value`
	stmt, err := conn.Conn().Prepare(ctx, "update_configuration", query)
	if err != nil {
		return err
	}
	if tenantName != conf.GlobalAccount {
		tenantName = strings.ToLower(tenantName)
	}
	_, err = conn.Conn().Query(ctx, stmt.Name, key, value, tenantName)
	if err != nil {
		return err
	}
	return nil
}

// Connection Functions
func UpdateProducersCounsumersConnection(connectionId string, isActive bool) (bool, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return false, err
	}
	defer conn.Release()
	query := `WITH updated_producers AS (
		UPDATE producers
		SET is_active = $1, updated_at = NOW()
		WHERE connection_id = $2
	  )
	  UPDATE consumers
	  SET is_active = $1, updated_at = NOW()
	  WHERE connection_id = $2;`
	stmt, err := conn.Conn().Prepare(ctx, "update_connection", query)
	if err != nil {
		return false, err
	}
	rows, err := conn.Conn().Query(ctx, stmt.Name, isActive, connectionId)
	if err != nil {
		return false, err
	}
	defer rows.Close()
	affectedRows := 0
	for rows.Next() {
		affectedRows++
		if affectedRows > 0 {
			return true, nil
		}
	}
	return false, nil
}

func GetActiveConnections() ([]string, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return []string{}, err
	}
	defer conn.Release()
	query := `SELECT connection_id FROM producers WHERE is_active = true UNION SELECT connection_id FROM consumers WHERE is_active = true;`
	stmt, err := conn.Conn().Prepare(ctx, "get_active_connection", query)
	if err != nil {
		return []string{}, err
	}
	rows, err := conn.Conn().Query(ctx, stmt.Name)
	if err != nil {
		return []string{}, err
	}
	defer rows.Close()
	var connections []string
	for rows.Next() {
		var connectionID string
		err = rows.Scan(&connectionID)
		if err != nil {
			return nil, err
		}
		connections = append(connections, connectionID)
	}
	if len(connections) == 0 {
		return []string{}, nil
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
	tenantName := auditLog[0].TenantName

	query := `INSERT INTO audit_logs ( 
		station_name, 
		message, 
		created_by,
		created_by_username,
		created_at,
		tenant_name
		) 
    VALUES($1, $2, $3, $4, $5, $6) RETURNING id`

	stmt, err := conn.Conn().Prepare(ctx, "insert_audit_logs", query)
	if err != nil {
		return err
	}

	newAuditLog := models.AuditLog{}
	rows, err := conn.Conn().Query(ctx, stmt.Name,
		stationName, message, createdBy, createdByUserName, createdAt, tenantName)
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

func GetAuditLogsByStation(name string, tenantName string) ([]models.AuditLog, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return []models.AuditLog{}, err
	}
	defer conn.Release()
	query := `SELECT * FROM audit_logs AS a
		WHERE a.station_name = $1 AND a.tenant_name = $2
		ORDER BY a.created_at DESC
		LIMIT 1000;`
	stmt, err := conn.Conn().Prepare(ctx, "get_audit_logs_by_station", query)
	if err != nil {
		return []models.AuditLog{}, err
	}
	rows, err := conn.Conn().Query(ctx, stmt.Name, name, tenantName)
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

func RemoveAllAuditLogsByStation(name string, tenantName string) error {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()

	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	removeAuditLogs := `DELETE FROM audit_logs
	WHERE station_name = $1 AND tenant_name = $2`

	stmt, err := conn.Conn().Prepare(ctx, "remove_audit_logs_by_station", removeAuditLogs)
	if err != nil {
		return err
	}

	_, err = conn.Conn().Exec(ctx, stmt.Name, name, tenantName)
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

func RemoveAuditLogsByTenant(tenantName string) error {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()
	query := `DELETE FROM audit_logs WHERE tenant_name = $1`
	stmt, err := conn.Conn().Prepare(ctx, "remove_audit_logs_by_tenant", query)
	if err != nil {
		return err
	}
	_, err = conn.Conn().Query(ctx, stmt.Name, tenantName)
	if err != nil {
		return err
	}
	return nil
}

// Station Functions
func GetActiveStationsPerTenant(tenantName string) ([]models.Station, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return []models.Station{}, err
	}
	defer conn.Release()
	query := `SELECT * FROM stations AS s WHERE (s.is_deleted = false) AND tenant_name = $1`
	stmt, err := conn.Conn().Prepare(ctx, "get_active_stations_per_tenant", query)
	if err != nil {
		return []models.Station{}, err
	}
	if tenantName != conf.GlobalAccount {
		tenantName = strings.ToLower(tenantName)
	}
	rows, err := conn.Conn().Query(ctx, stmt.Name, tenantName)
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

func GetActiveStations() ([]models.Station, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return []models.Station{}, err
	}
	defer conn.Release()
	query := `SELECT * FROM stations AS s WHERE s.is_deleted = false`
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

func GetStationByName(name string, tenantName string) (bool, models.Station, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return false, models.Station{}, err
	}
	defer conn.Release()
	query := `SELECT * FROM stations WHERE name = $1 AND (is_deleted = false) AND tenant_name = $2 LIMIT 1`
	stmt, err := conn.Conn().Prepare(ctx, "get_station_by_name", query)
	if err != nil {
		return false, models.Station{}, err
	}
	if tenantName != conf.GlobalAccount {
		tenantName = strings.ToLower(tenantName)
	}
	rows, err := conn.Conn().Query(ctx, stmt.Name, name, tenantName)
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

func GetStationById(stationId int, tenantName string) (bool, models.Station, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return false, models.Station{}, err
	}
	defer conn.Release()
	query := `SELECT * FROM stations WHERE id = $1 AND (is_deleted = false) AND tenant_name = $2 LIMIT 1`
	stmt, err := conn.Conn().Prepare(ctx, "get_station_by_id", query)
	if err != nil {
		return false, models.Station{}, err
	}
	if tenantName != conf.GlobalAccount {
		tenantName = strings.ToLower(tenantName)
	}
	rows, err := conn.Conn().Query(ctx, stmt.Name, stationId, tenantName)
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
	tieredStorageEnabled bool,
	tenantName string,
	partitionsList []int,
	version int,
	dlsStationName string) (models.Station, int64, error) {
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
		tiered_storage_enabled,
		tenant_name,
		partitions,
		version,
		dls_station
		) 
    VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21) RETURNING id`

	stmt, err := conn.Conn().Prepare(ctx, "insert_new_station", query)
	if err != nil {
		return models.Station{}, 0, err
	}

	createAt := time.Now()
	updatedAt := time.Now()
	var stationId int
	if tenantName != conf.GlobalAccount {
		tenantName = strings.ToLower(tenantName)
	}
	rows, err := conn.Conn().Query(ctx, stmt.Name,
		stationName, retentionType, retentionValue, storageType, replicas, userId, username, createAt, updatedAt,
		false, schemaName, schemaVersionUpdate, idempotencyWindow, isNative, dlsConfiguration.Poison, dlsConfiguration.Schemaverse, tieredStorageEnabled, tenantName, partitionsList, version, dlsStationName)
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
	if tenantName != conf.GlobalAccount {
		tenantName = strings.ToLower(tenantName)
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
		TenantName:                  tenantName,
		PartitionsList:              partitionsList,
		Version:                     version,
		DlsStation:                  dlsStationName,
	}

	rowsAffected := rows.CommandTag().RowsAffected()
	return newStation, rowsAffected, nil
}

func GetAllStationsWithNoHA3() ([]models.Station, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return []models.Station{}, err
	}
	defer conn.Release()
	query := `SELECT * FROM stations WHERE replicas = 1 OR replicas = 5`
	stmt, err := conn.Conn().Prepare(ctx, "get_all_stations_with_no_ha_3", query)
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
		return []models.Station{}, err
	}
	return stations, nil
}

func GetAllStationsDetailsPerTenant(tenantName string) ([]models.ExtendedStation, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return []models.ExtendedStation{}, err
	}
	defer conn.Release()
	query := `
	SELECT s.*, COALESCE(p.id, 0),  
	COALESCE(p.name, ''), 
	COALESCE(p.station_id, 0), 
	COALESCE(p.type, 'application'), 
	COALESCE(p.connection_id, ''), 
	COALESCE(p.is_active, false), 
	COALESCE(p.updated_at, CURRENT_TIMESTAMP), 
	COALESCE(c.id, 0),  
	COALESCE(c.name, ''), 
	COALESCE(c.station_id, 0), 
	COALESCE(c.type, 'application'), 
	COALESCE(c.connection_id, ''),
	COALESCE(c.consumers_group, ''),
	COALESCE(c.max_ack_time_ms, 0), 
	COALESCE(c.is_active, false), 
	COALESCE(c.updated_at, CURRENT_TIMESTAMP), 
	COALESCE(c.max_msg_deliveries, 0), 
	COALESCE(c.start_consume_from_seq, 0),
	COALESCE(c.last_msgs, 0) 
	FROM stations AS s
	LEFT JOIN producers AS p
	ON s.id = p.station_id 
	LEFT JOIN consumers AS c 
	ON s.id = c.station_id
	WHERE s.is_deleted = false AND s.tenant_name = $1
	GROUP BY s.id,p.id,c.id`
	stmt, err := conn.Conn().Prepare(ctx, "get_all_stations_details_per_tenant", query)
	if err != nil {
		return []models.ExtendedStation{}, err
	}

	rows, err := conn.Conn().Query(ctx, stmt.Name, tenantName)
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
			&stationRes.TenantName,
			&stationRes.ResendDisabled,
			&producer.ID,
			&producer.Name,
			&producer.StationId,
			&producer.Type,
			&producer.ConnectionId,
			&producer.IsActive,
			&producer.UpdatedAt,
			&consumer.ID,
			&consumer.Name,
			&consumer.StationId,
			&consumer.Type,
			&consumer.ConnectionId,
			&consumer.ConsumersGroup,
			&consumer.MaxAckTimeMs,
			&consumer.IsActive,
			&consumer.UpdatedAt,
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
			if tenantName != conf.GlobalAccount {
				tenantName = strings.ToLower(tenantName)
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
				TenantName:                  tenantName,
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

func GetAllStationsDetailsLight(tenantName string) ([]models.ExtendedStationLight, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return []models.ExtendedStationLight{}, err
	}
	defer conn.Release()

	query := `
	SELECT s.*,
	false AS is_active
	FROM stations AS s
	WHERE s.is_deleted = false AND s.tenant_name = $1
	ORDER BY s.updated_at DESC
	LIMIT 5000;`
	stmt, err := conn.Conn().Prepare(ctx, "get_all_stations_details_for_main_overview", query)
	if err != nil {
		return []models.ExtendedStationLight{}, err
	}

	rows, err := conn.Conn().Query(ctx, stmt.Name, tenantName)
	if err != nil {
		return []models.ExtendedStationLight{}, err
	}
	if err == pgx.ErrNoRows {
		return []models.ExtendedStationLight{}, nil
	}
	defer rows.Close()
	stations := []models.ExtendedStationLight{}
	for rows.Next() {
		var stationRes models.ExtendedStationLight
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
			&stationRes.TenantName,
			&stationRes.ResendDisabled,
			&stationRes.PartitionsList,
			&stationRes.Version,
			&stationRes.DlsStation,
			&stationRes.FunctionsLockHeld,
			&stationRes.FunctionsLockedAt,
			&stationRes.Activity,
		); err != nil {
			return []models.ExtendedStationLight{}, err
		}
		stationRes.TenantName = tenantName
		stations = append(stations, stationRes)
	}
	if err := rows.Err(); err != nil {
		return []models.ExtendedStationLight{}, err
	}
	return stations, nil
}

func GetStationsLight(tenantName string) ([]models.StationLight, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return []models.StationLight{}, err
	}
	defer conn.Release()

	query := `
	SELECT s.id, s.name, s.schema_name,
	(SELECT COUNT(*) FROM dls_messages dm WHERE dm.station_id = s.id) AS dls_count
	FROM stations AS s
	WHERE s.is_deleted = false AND s.tenant_name = $1
	ORDER BY s.updated_at DESC
	LIMIT 40000;`
	stmt, err := conn.Conn().Prepare(ctx, "get_active_stations_light", query)
	if err != nil {
		return []models.StationLight{}, err
	}

	rows, err := conn.Conn().Query(ctx, stmt.Name, tenantName)
	if err != nil {
		return []models.StationLight{}, err
	}
	if err == pgx.ErrNoRows {
		return []models.StationLight{}, nil
	}
	defer rows.Close()
	stations := []models.StationLight{}
	for rows.Next() {
		var stationRes models.StationLight
		if err := rows.Scan(
			&stationRes.ID,
			&stationRes.Name,
			&stationRes.SchemaName,
			&stationRes.DlsMsgs,
		); err != nil {
			return []models.StationLight{}, err
		}
		stations = append(stations, stationRes)
	}
	if err := rows.Err(); err != nil {
		return []models.StationLight{}, err
	}
	return stations, nil
}

func GetAllStationsWithActiveProducersConsumersPerTenant(tenantName string) ([]models.ActiveProducersConsumersDetails, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return []models.ActiveProducersConsumersDetails{}, err
	}
	defer conn.Release()
	query := `
	SELECT s.id,
			   COUNT(DISTINCT CASE WHEN p.is_active THEN p.id END) FILTER (WHERE p.is_active = true) AS active_producers_count,
			   COUNT(DISTINCT CASE WHEN c.is_active THEN c.id END) FILTER (WHERE c.is_active = true) AS active_consumers_count
		FROM stations AS s
		LEFT JOIN producers AS p ON s.id = p.station_id AND p.is_active = true
		LEFT JOIN consumers AS c ON s.id = c.station_id AND c.is_active = true
		WHERE s.is_deleted = false AND s.tenant_name = $1
	GROUP BY s.id;
`
	stmt, err := conn.Conn().Prepare(ctx, "get_all_active_producers_consumers_of_station", query)
	if err != nil {
		return []models.ActiveProducersConsumersDetails{}, err
	}

	rows, err := conn.Conn().Query(ctx, stmt.Name, tenantName)
	if err != nil {
		return []models.ActiveProducersConsumersDetails{}, err
	}
	if err == pgx.ErrNoRows {
		return []models.ActiveProducersConsumersDetails{}, nil
	}
	defer rows.Close()
	stations := []models.ActiveProducersConsumersDetails{}
	for rows.Next() {
		var stationId int
		var activeProducersCount int
		var activeConsumersCount int
		if err := rows.Scan(
			&stationId,
			&activeProducersCount,
			&activeConsumersCount,
		); err != nil {
			return []models.ActiveProducersConsumersDetails{}, err
		}
		station := models.ActiveProducersConsumersDetails{
			ID:                   stationId,
			ActiveProducersCount: activeProducersCount,
			ActiveConsumersCount: activeConsumersCount,
		}
		stations = append(stations, station)
	}
	if err := rows.Err(); err != nil {
		return []models.ActiveProducersConsumersDetails{}, err
	}
	return stations, nil
}

func GetAllStations() ([]models.Station, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return []models.Station{}, err
	}
	defer conn.Release()
	query := `SELECT * FROM stations`
	stmt, err := conn.Conn().Prepare(ctx, "get_all_stations", query)
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
		return []models.Station{}, err
	}
	return stations, nil
}

func CountStationsByTenant(tenantName string) (int, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return 0, err
	}
	defer conn.Release()
	query := `SELECT COUNT(*) FROM stations where tenant_name = $1`
	stmt, err := conn.Conn().Prepare(ctx, "get_count_stations_by_tenant", query)
	if err != nil {
		return 0, err
	}
	var count int
	err = conn.Conn().QueryRow(ctx, stmt.Name, tenantName).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
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
	SELECT s.*, COALESCE(p.id, 0),  
	COALESCE(p.name, ''), 
	COALESCE(p.station_id, 0), 
	COALESCE(p.type, 'application'), 
	COALESCE(p.connection_id, ''), 
	COALESCE(p.created_by, 0), 
	COALESCE(p.created_by_username, ''), 
	COALESCE(p.is_active, false), 
	COALESCE(p.created_at, CURRENT_TIMESTAMP), 
	COALESCE(p.is_deleted, false),
	COALESCE(p.tenant_name, ''),  
	COALESCE(c.id, 0),  
	COALESCE(c.name, ''), 
	COALESCE(c.station_id, 0), 
	COALESCE(c.type, 'application'), 
	COALESCE(c.connection_id, ''),
	COALESCE(c.consumers_group, ''),
	COALESCE(c.max_ack_time_ms, 0), 
	COALESCE(c.created_by, 0), 
	COALESCE(c.created_by_username, ''), 
	COALESCE(c.is_active, false), 
	COALESCE(c.created_at, CURRENT_TIMESTAMP), 
	COALESCE(c.is_deleted, false), 
	COALESCE(c.max_msg_deliveries, 0), 
	COALESCE(c.start_consume_from_seq, 0),
	COALESCE(c.last_msgs, 0),
	COALESCE(c.tenant_name, '') 
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
			&stationRes.TenantName,
			&stationRes.FunctionsLockHeld,
			&stationRes.FunctionsLockedAt,
			&producer.ID,
			&producer.Name,
			&producer.StationId,
			&producer.Type,
			&producer.ConnectionId,
			&producer.IsActive,
			&producer.UpdatedAt,
			&producer.TenantName,
			&consumer.ID,
			&consumer.Name,
			&consumer.StationId,
			&consumer.Type,
			&consumer.ConnectionId,
			&consumer.ConsumersGroup,
			&consumer.MaxAckTimeMs,
			&consumer.IsActive,
			&consumer.UpdatedAt,
			&consumer.MaxMsgDeliveries,
			&consumer.StartConsumeFromSeq,
			&consumer.LastMessages,
			&consumer.TenantName,
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
				if producersMap[p.Name].UpdatedAt.Before(p.UpdatedAt) {
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
				if consumersMap[c.Name].UpdatedAt.Before(c.UpdatedAt) {
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

func DeleteStationsByNames(stationNames []string, tenantName string) error {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()
	query := `DELETE FROM stations
	WHERE name = ANY($1)
	AND (is_deleted = false)
	AND tenant_name=$2`
	stmt, err := conn.Conn().Prepare(ctx, "delete_stations_by_names", query)
	if err != nil {
		return err
	}
	if tenantName != conf.GlobalAccount {
		tenantName = strings.ToLower(tenantName)
	}
	_, err = conn.Conn().Query(ctx, stmt.Name, stationNames, tenantName)
	if err != nil {
		return err
	}
	return nil
}

func RemoveDeletedStations() error {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()
	query := `DELETE FROM stations WHERE is_deleted = true`
	stmt, err := conn.Conn().Prepare(ctx, "remove_deleted_stations", query)
	if err != nil {
		return err
	}

	_, err = conn.Conn().Query(ctx, stmt.Name)
	if err != nil {
		return err
	}
	return nil
}

func DeleteStation(name string, tenantName string) error {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()
	query := `DELETE FROM stations
	WHERE name = $1
	AND tenant_name=$2`
	stmt, err := conn.Conn().Prepare(ctx, "delete_station", query)
	if err != nil {
		return err
	}
	if tenantName != conf.GlobalAccount {
		tenantName = strings.ToLower(tenantName)
	}
	_, err = conn.Conn().Query(ctx, stmt.Name, name, tenantName)
	if err != nil {
		return err
	}
	return nil
}

func AttachSchemaToStation(stationName string, schemaName string, versionNumber int, tenantName string) error {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()
	query := `UPDATE stations SET schema_name = $2, schema_version_number = $3
	WHERE name = $1 AND is_deleted = false AND tenant_name=$4`
	stmt, err := conn.Conn().Prepare(ctx, "attach_schema_to_station", query)
	if err != nil {
		return err
	}
	if tenantName != conf.GlobalAccount {
		tenantName = strings.ToLower(tenantName)
	}
	_, err = conn.Conn().Query(ctx, stmt.Name, stationName, schemaName, versionNumber, tenantName)
	if err != nil {
		return err
	}
	return nil
}

func DetachSchemaFromStation(stationName string, tenantName string) error {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()
	query := `UPDATE stations SET schema_name = '', schema_version_number = 0
	WHERE name = $1 AND is_deleted = false AND tenant_name=$2`
	stmt, err := conn.Conn().Prepare(ctx, "detach_schema_from_station", query)
	if err != nil {
		return err
	}
	if tenantName != conf.GlobalAccount {
		tenantName = strings.ToLower(tenantName)
	}
	_, err = conn.Conn().Query(ctx, stmt.Name, stationName, tenantName)
	if err != nil {
		return err
	}
	return nil
}

func UpdateStationDlsConfig(stationName string, poison bool, schemaverse bool, tenantName string) error {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()
	query := `UPDATE stations SET dls_configuration_poison = $2, dls_configuration_schemaverse = $3
	WHERE name = $1 AND is_deleted = false AND tenant_name=$4`
	stmt, err := conn.Conn().Prepare(ctx, "update_station_dls_config", query)
	if err != nil {
		return err
	}
	if tenantName != conf.GlobalAccount {
		tenantName = strings.ToLower(tenantName)
	}
	_, err = conn.Conn().Query(ctx, stmt.Name, stationName, poison, schemaverse, tenantName)
	if err != nil {
		return err
	}
	return nil
}

func UpdateStationsOfDeletedUser(userId int, tenantName string) error {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()
	query := `UPDATE stations SET created_by = 0, created_by_username = CONCAT(created_by_username, '(deleted)') WHERE created_by = $1 AND created_by_username NOT LIKE '%(deleted)' AND tenant_name=$2`
	stmt, err := conn.Conn().Prepare(ctx, "update_stations_of_deleted_user", query)
	if err != nil {
		return err
	}
	if tenantName != conf.GlobalAccount {
		tenantName = strings.ToLower(tenantName)
	}
	_, err = conn.Conn().Query(ctx, stmt.Name, userId, tenantName)
	if err != nil {
		return err
	}
	return nil
}

func UpdateStationsWithNoHA3() error {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()
	query := `UPDATE stations SET replicas = 3 WHERE replicas = 1 OR replicas = 5`
	stmt, err := conn.Conn().Prepare(ctx, "update_stations_with_no_ha_3", query)
	if err != nil {
		return err
	}
	_, err = conn.Conn().Query(ctx, stmt.Name)
	if err != nil {
		return err
	}
	return nil
}

func UpdateResendDisabledInStations(resendDisabled bool, stationId []int) error {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()
	query := `UPDATE stations SET resend_disabled = $1 WHERE id = ANY($2)`
	stmt, err := conn.Conn().Prepare(ctx, "update_resend_disabled_in_stations", query)
	if err != nil {
		return err
	}
	_, err = conn.Conn().Query(ctx, stmt.Name, resendDisabled, stationId)
	if err != nil {
		return err
	}
	return nil
}

func RemoveStationsByTenant(tenantName string) error {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()
	query := `DELETE FROM stations WHERE tenant_name=$1`
	stmt, err := conn.Conn().Prepare(ctx, "remove_stations_by_tenant", query)
	if err != nil {
		return err
	}
	_, err = conn.Conn().Query(ctx, stmt.Name, tenantName)
	if err != nil {
		return err
	}
	return nil
}

func GetStationNamesUsingSchema(schemaName string, tenantName string) ([]string, error) {
	stationNames := []string{}
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Release()
	query := `
		SELECT name FROM stations
		WHERE schema_name = $1 AND is_deleted = false AND tenant_name = $2
	`
	stmt, err := conn.Conn().Prepare(ctx, "get_station_names_using_schema", query)
	if err != nil {
		return nil, err
	}
	if tenantName != conf.GlobalAccount {
		tenantName = strings.ToLower(tenantName)
	}
	rows, err := conn.Conn().Query(ctx, stmt.Name, schemaName, tenantName)
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

func GetCountStationsUsingSchema(schemaName string, tenantName string) (int, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return 0, err
	}
	defer conn.Release()
	query := `SELECT COUNT(*) FROM stations WHERE schema_name = $1 AND is_deleted = false AND tenant_name=$2`
	stmt, err := conn.Conn().Prepare(ctx, "get_count_stations_using_schema", query)
	if err != nil {
		return 0, err
	}
	var count int
	if tenantName != conf.GlobalAccount {
		tenantName = strings.ToLower(tenantName)
	}
	err = conn.Conn().QueryRow(ctx, stmt.Name, schemaName, tenantName).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func RemoveSchemaFromAllUsingStations(schemaName string, tenantName string) error {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()
	query := `UPDATE stations SET schema_name = '' WHERE schema_name = $1 AND tenant_name=$2`
	stmt, err := conn.Conn().Prepare(ctx, "remove_schema_from_all_using_stations", query)
	if err != nil {
		return err
	}
	if tenantName != conf.GlobalAccount {
		tenantName = strings.ToLower(tenantName)
	}
	_, err = conn.Conn().Query(ctx, stmt.Name, schemaName, tenantName)
	if err != nil {
		return err
	}
	return nil
}

func GetDeletedStations() ([]models.Station, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return []models.Station{}, err
	}
	defer conn.Release()
	query := `SELECT * FROM stations WHERE is_deleted = true`
	stmt, err := conn.Conn().Prepare(ctx, "get_not_active_stations", query)
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
		return []models.Station{}, err
	}
	return stations, nil
}

func UpdateStationsDls(stationNames []string, dlsName, tenantName string) error {
	if len(stationNames) == 0 {
		return nil
	}
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	query := `UPDATE stations SET dls_station = $1 WHERE name = ANY($2) AND tenant_name=$3`
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()
	stmt, err := conn.Conn().Prepare(ctx, "update_stations_dls", query)
	if err != nil {
		return err
	}
	if tenantName != conf.GlobalAccount {
		tenantName = strings.ToLower(tenantName)
	}
	_, err = conn.Conn().Query(ctx, stmt.Name, dlsName, stationNames, tenantName)
	if err != nil {
		return err
	}
	return nil
}

func RemoveDlsStationFromAllStations(name, tenantName string) error {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	query := `UPDATE stations SET dls_station = '' WHERE dls_station = $1 AND tenant_name=$2`
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()
	stmt, err := conn.Conn().Prepare(ctx, "remove_dls_station_from_all_stations", query)
	if err != nil {
		return err
	}
	if tenantName != conf.GlobalAccount {
		tenantName = strings.ToLower(tenantName)
	}
	_, err = conn.Conn().Query(ctx, stmt.Name, name, tenantName)
	if err != nil {
		return err
	}
	return nil
}

// Producer Functions
func UpdateProducersActiveAndGetDetails(connectionId string, isActive bool) ([]models.LightProducer, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return []models.LightProducer{}, err
	}
	defer conn.Release()
	query := `
	WITH updated_producers AS (
		UPDATE producers
		SET is_active = $2
		WHERE connection_id = $1 AND is_active = true
		RETURNING *
	)
	SELECT DISTINCT ON (p.name) p.name, s.name, COUNT(*) OVER (PARTITION BY p.name)
	FROM updated_producers AS p
	LEFT JOIN stations AS s ON s.id = p.station_id
	GROUP BY p.name, p.id, s.id
	LIMIT 5000;`
	stmt, err := conn.Conn().Prepare(ctx, "get_producers_by_connection_id_with_station_details", query)
	if err != nil {
		return []models.LightProducer{}, err
	}
	rows, err := conn.Conn().Query(ctx, stmt.Name, connectionId, isActive)
	if err != nil {
		return []models.LightProducer{}, err
	}
	defer rows.Close()
	producers, err := pgx.CollectRows(rows, pgx.RowToStructByPos[models.LightProducer])
	if err != nil {
		return []models.LightProducer{}, err
	}
	if len(producers) == 0 {
		return []models.LightProducer{}, nil
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
	query := `SELECT * FROM producers WHERE id = $1 LIMIT 1`
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

func GetProducerByStationIDAndConnectionId(name string, stationId int, connectionId string) (bool, models.Producer, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return false, models.Producer{}, err
	}
	defer conn.Release()
	query := `SELECT * FROM producers WHERE name = $1 AND station_id = $2 AND connection_id = $3 ORDER BY is_active DESC LIMIT 1`
	stmt, err := conn.Conn().Prepare(ctx, "get_producer_by_station_id_and_connection_id", query)
	if err != nil {
		return false, models.Producer{}, err
	}
	rows, err := conn.Conn().Query(ctx, stmt.Name, name, stationId, connectionId)
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

func GetProducerByNameAndStationID(name string, stationId int) (bool, models.Producer, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return false, models.Producer{}, err
	}
	defer conn.Release()
	query := `SELECT * FROM producers WHERE name = $1 AND station_id = $2 ORDER BY is_active DESC LIMIT 1`
	stmt, err := conn.Conn().Prepare(ctx, "get_producer_by_name_and_station_id", query)
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

func GetProducersForGraph(tenantName string) ([]models.ProducerForGraph, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return []models.ProducerForGraph{}, err
	}
	defer conn.Release()
	query := `SELECT p.name, p.station_id, p.app_id
				FROM producers AS p
				WHERE p.tenant_name = $1 AND p.is_active = true
				ORDER BY p.name, p.station_id DESC
				LIMIT 10000;`
	stmt, err := conn.Conn().Prepare(ctx, "get_producers_for_graph", query)
	if err != nil {
		return []models.ProducerForGraph{}, err
	}
	rows, err := conn.Conn().Query(ctx, stmt.Name, tenantName)
	if err != nil {
		return []models.ProducerForGraph{}, err
	}
	defer rows.Close()
	producers, err := pgx.CollectRows(rows, pgx.RowToStructByPos[models.ProducerForGraph])
	if err != nil {
		return []models.ProducerForGraph{}, err
	}
	if len(producers) == 0 {
		return []models.ProducerForGraph{}, nil
	}
	return producers, nil
}

func InsertNewProducer(name string, stationId int, producerType string, connectionIdObj string, tenantName string, partitionsList []int, version int, sdk string, appId string) (models.Producer, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()

	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return models.Producer{}, err
	}
	defer conn.Release()

	query := `INSERT INTO producers ( 
		name, 
		station_id, 
		connection_id,
		is_active, 
		updated_at, 
		type,
		tenant_name,
		partitions,
		version,
		sdk,
		app_id) 
    VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11) RETURNING id`

	stmt, err := conn.Conn().Prepare(ctx, "insert_new_producer", query)
	if err != nil {
		return models.Producer{}, err
	}

	var producerId int
	updatedAt := time.Now()
	isActive := true
	if tenantName != conf.GlobalAccount {
		tenantName = strings.ToLower(tenantName)
	}
	rows, err := conn.Conn().Query(ctx, stmt.Name, name, stationId, connectionIdObj, isActive, updatedAt, producerType, tenantName, partitionsList, version, sdk, appId)
	if err != nil {
		return models.Producer{}, err
	}
	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&producerId)
		if err != nil {
			return models.Producer{}, err
		}
	}

	if err := rows.Err(); err != nil {
		return models.Producer{}, err
	}

	if err := rows.Err(); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Detail != "" {
				return models.Producer{}, errors.New(pgErr.Detail)
			} else {
				return models.Producer{}, errors.New(pgErr.Message)
			}
		} else {
			return models.Producer{}, err
		}
	}

	newProducer := models.Producer{
		ID:             producerId,
		Name:           name,
		StationId:      stationId,
		Type:           producerType,
		ConnectionId:   connectionIdObj,
		IsActive:       isActive,
		UpdatedAt:      time.Now(),
		PartitionsList: partitionsList,
	}
	return newProducer, nil
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
	stmt, err := conn.Conn().Prepare(ctx, "get_not_deleted_producers_by_station_id", query)
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
	query := `SELECT
			p.id,
			p.name,
			p.type,
			p.connection_id,
			p.updated_at,
			s.name,
			p.is_active,
			COUNT(CASE WHEN p.is_active THEN 1 END) OVER (PARTITION BY p.name) AS count_producers
			FROM producers AS p
			LEFT JOIN stations AS s ON s.id = p.station_id
			WHERE p.station_id = $1
			ORDER BY p.name, p.is_active DESC, p.updated_at DESC
			LIMIT 5000;`
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

func DeleteProducerByNameAndStationID(name string, stationId int) (bool, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return false, err
	}
	defer conn.Release()
	query := `DELETE FROM producers WHERE name = $1 AND station_id = $2 LIMIT 1`
	stmt, err := conn.Conn().Prepare(ctx, "delete_producer_by_name_and_station_id", query)
	if err != nil {
		return false, err
	}
	_, err = conn.Conn().Query(ctx, stmt.Name, name, stationId)
	if err != nil {
		return false, err
	}
	return true, nil
}

func DeleteProducerByNameStationIDAndConnID(name string, stationId int, connId string) (bool, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return false, err
	}
	defer conn.Release()
	query := `DELETE FROM producers WHERE name = $1 AND station_id = $2 AND connection_id = $3
	AND EXISTS (
		SELECT 1 FROM producers
		WHERE name = $1 AND station_id = $2 AND connection_id = $3
		FETCH FIRST 1 ROW ONLY
	);`
	stmt, err := conn.Conn().Prepare(ctx, "delete_producer_by_name_and_station_id", query)
	if err != nil {
		return false, err
	}
	_, err = conn.Conn().Query(ctx, stmt.Name, name, stationId, connId)
	if err != nil {
		return false, err
	}
	return true, nil
}

func DeleteProducersByStationID(stationId int) error {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()
	query := `DELETE FROM producers WHERE station_id = $1`
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
	query := `UPDATE producers SET created_by = 0, created_by_username = CONCAT(created_by_username, '(deleted)') WHERE created_by = $1 AND created_by_username NOT LIKE '%(deleted)'`
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

func RemoveProducersByTenant(tenantName string) error {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()
	query := `DELETE FROM producers WHERE tenant_name = $1`
	stmt, err := conn.Conn().Prepare(ctx, "delete_producers_by_tenant", query)
	if err != nil {
		return err
	}
	_, err = conn.Conn().Query(ctx, stmt.Name, tenantName)
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

	query := `SELECT * FROM consumers WHERE consumers_group = $1 AND station_id = $2 LIMIT 1`
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
	cgName string,
	maxAckTime int,
	maxMsgDeliveries int,
	startConsumeFromSequence uint64,
	lastMessages int64,
	tenantName string,
	partitionsList []int,
	version int, sdk string, appId string) (models.Consumer, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()

	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return models.Consumer{}, err
	}
	defer conn.Release()

	query := `INSERT INTO consumers ( 
		name, 
		station_id,
		connection_id,
		consumers_group,
		max_ack_time_ms,
		is_active, 
		updated_at,
		max_msg_deliveries,
		start_consume_from_seq,
		last_msgs,
		type,
		tenant_name, 
		partitions,
		version,
		sdk,
		app_id)
    VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16) 
	RETURNING id`

	stmt, err := conn.Conn().Prepare(ctx, "insert_new_consumer", query)
	if err != nil {
		return models.Consumer{}, err
	}

	var consumerId int
	updatedAt := time.Now()
	isActive := true

	rows, err := conn.Conn().Query(ctx, stmt.Name,
		name, stationId, connectionIdObj, cgName, maxAckTime, isActive, updatedAt, maxMsgDeliveries, startConsumeFromSequence, lastMessages, consumerType, tenantName, partitionsList, version, sdk, appId)
	if err != nil {
		return models.Consumer{}, err
	}
	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&consumerId)
		if err != nil {
			return models.Consumer{}, err
		}
	}

	if err := rows.Err(); err != nil {
		return models.Consumer{}, err
	}

	if err := rows.Err(); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Detail != "" {
				return models.Consumer{}, errors.New(pgErr.Detail)
			} else {
				return models.Consumer{}, errors.New(pgErr.Message)
			}
		} else {
			return models.Consumer{}, err
		}
	}

	newConsumer := models.Consumer{
		ID:                  consumerId,
		Name:                name,
		StationId:           stationId,
		Type:                consumerType,
		ConnectionId:        connectionIdObj,
		ConsumersGroup:      cgName,
		IsActive:            isActive,
		UpdatedAt:           time.Now(),
		MaxAckTimeMs:        int64(maxAckTime),
		MaxMsgDeliveries:    maxMsgDeliveries,
		StartConsumeFromSeq: startConsumeFromSequence,
		LastMessages:        lastMessages,
		TenantName:          tenantName,
		PartitionsList:      partitionsList,
	}
	return newConsumer, nil
}

func GetConsumers() ([]models.Consumer, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return []models.Consumer{}, err
	}
	defer conn.Release()
	query := `SELECT * From consumers`
	stmt, err := conn.Conn().Prepare(ctx, "get_consumers", query)
	if err != nil {
		return []models.Consumer{}, err
	}
	rows, err := conn.Conn().Query(ctx, stmt.Name)
	if err != nil {
		return []models.Consumer{}, err
	}
	defer rows.Close()
	consumers, err := pgx.CollectRows(rows, pgx.RowToStructByPos[models.Consumer])
	if err != nil {
		return []models.Consumer{}, err
	}
	if len(consumers) == 0 {
		return []models.Consumer{}, err
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
	query := `SELECT DISTINCT ON (c.name, c.consumers_group) c.id, c.name, c.updated_at, c.is_active, c.consumers_group, c.max_ack_time_ms, c.max_msg_deliveries, s.name, c.partitions,
				COUNT (CASE WHEN c.is_active THEN 1 END) OVER (PARTITION BY c.name) AS count
				FROM consumers AS c
				LEFT JOIN stations AS s ON s.id = c.station_id
				WHERE c.station_id = $1 ORDER BY c.name, c.consumers_group, c.updated_at DESC
				LIMIT 5000;`
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

func GetConsumersForGraph(tenantName string) ([]models.ConsumerForGraph, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return []models.ConsumerForGraph{}, err
	}
	defer conn.Release()
	query := `SELECT c.name, c.consumers_group, c.station_id, c.app_id
				FROM consumers AS c
				WHERE c.tenant_name = $1 AND c.is_active = true
				ORDER BY c.name, c.station_id DESC
				LIMIT 10000;`
	stmt, err := conn.Conn().Prepare(ctx, "get_consumers_for_graph", query)
	if err != nil {
		return []models.ConsumerForGraph{}, err
	}
	rows, err := conn.Conn().Query(ctx, stmt.Name, tenantName)
	if err != nil {
		return []models.ConsumerForGraph{}, err
	}
	defer rows.Close()
	cgs, err := pgx.CollectRows(rows, pgx.RowToStructByPos[models.ConsumerForGraph])
	if err != nil {
		return []models.ConsumerForGraph{}, err
	}
	if len(cgs) == 0 {
		return []models.ConsumerForGraph{}, nil
	}
	return cgs, nil
}

func DeleteConsumerByNameStationIDAndConnID(connectionId, name string, stationId int) (bool, models.Consumer, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return false, models.Consumer{}, err
	}
	defer conn.Release()
	query := ` DELETE FROM consumers WHERE ctid = ( SELECT ctid FROM consumers WHERE connection_id = $1 AND name = $2 AND station_id = $3 LIMIT 1) RETURNING *`
	deleteStmt, err := conn.Conn().Prepare(ctx, "delete_consumers", query)
	if err != nil {
		return false, models.Consumer{}, err
	}
	rows, err := conn.Conn().Query(ctx, deleteStmt.Name, connectionId, name, stationId)
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
	return true, consumers[0], nil
}

func DeleteConsumerByNameAndStationId(name string, stationId int) (bool, models.Consumer, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return false, models.Consumer{}, err
	}
	defer conn.Release()
	query := ` DELETE FROM consumers WHERE ctid = ( SELECT ctid FROM consumers WHERE name = $1 AND station_id = $ LIMIT 1) RETURNING *`
	deleteStmt, err := conn.Conn().Prepare(ctx, "delete_consumers", query)
	if err != nil {
		return false, models.Consumer{}, err
	}
	rows, err := conn.Conn().Query(ctx, deleteStmt.Name, name, stationId)
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
	return true, consumers[0], nil
}

func DeleteAllConsumersByStationID(stationId int) error {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()
	query := `DELETE FROM consumers WHERE station_id = $1`
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

func DeleteDLSMessagesByStationID(stationId int) error {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()
	query := `DELETE FROM dls_messages WHERE station_id = $1`
	stmt, err := conn.Conn().Prepare(ctx, "delete_messages_from_chosen_dls", query)
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
	query := `SELECT COUNT(*) FROM consumers WHERE station_id = $1 AND consumers_group = $2 AND is_active = true`
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
	stmt, err := conn.Conn().Prepare(ctx, "count_all_active_consumers", query)
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
			c.connection_id,
			c.is_active,
			c.max_msg_deliveries,
			c.max_ack_time_ms,
			c.partitions,
			COUNT (CASE WHEN c.is_active THEN 1 END) OVER (PARTITION BY c.name) AS count
		FROM
			consumers AS c
		WHERE
			c.consumers_group = $1
			AND c.station_id = $2
		ORDER BY
			c.name, c.updated_at DESC
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

func UpdateCosnumersActiveAndGetDetails(connectionId string, isActive bool) ([]models.LightConsumer, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return []models.LightConsumer{}, err
	}
	defer conn.Release()
	query := `
	WITH updated_consumers AS (
		UPDATE consumers
		SET is_active = $2
		WHERE connection_id = $1 AND is_active = true
		RETURNING *
	)
	SELECT DISTINCT ON (c.name) c.name, s.name, COUNT(*) OVER (PARTITION BY c.name)
	FROM updated_consumers AS c
	LEFT JOIN stations AS s ON s.id = c.station_id
	GROUP BY c.name, c.id, s.id
	LIMIT 5000;`
	stmt, err := conn.Conn().Prepare(ctx, "get_all_consumers_by_connection_id_with_station_details", query)
	if err != nil {
		return []models.LightConsumer{}, err
	}
	rows, err := conn.Conn().Query(ctx, stmt.Name, connectionId, isActive)
	if err != nil {
		return []models.LightConsumer{}, err
	}
	defer rows.Close()
	consumers, err := pgx.CollectRows(rows, pgx.RowToStructByPos[models.LightConsumer])
	if err != nil {
		return []models.LightConsumer{}, err
	}
	if len(consumers) == 0 {
		return []models.LightConsumer{}, nil
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

func RemoveConsumersByTenant(tenantName string) error {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()
	query := `DELETE FROM consumers WHERE tenant_name = $1`
	stmt, err := conn.Conn().Prepare(ctx, "remove_consumers_by_tenant", query)
	if err != nil {
		return err
	}
	_, err = conn.Conn().Query(ctx, stmt.Name, tenantName)
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

func GetActiveCgsByName(names []string, tenantName string) ([]models.LightConsumer, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return []models.LightConsumer{}, err
	}
	defer conn.Release()
	query := `
		SELECT c.consumers_group, s.name, COUNT(*)
		FROM consumers AS c
		LEFT JOIN stations AS s ON s.id = c.station_id
		WHERE c.tenant_name = $1 AND c.consumers_group = ANY($2) AND c.is_active = true
		GROUP BY c.consumers_group, s.name, c.station_id;`
	stmt, err := conn.Conn().Prepare(ctx, "get_active_consumers_by_name", query)
	if err != nil {
		return []models.LightConsumer{}, err
	}
	rows, err := conn.Conn().Query(ctx, stmt.Name, tenantName, names)
	if err != nil {
		return []models.LightConsumer{}, err
	}
	defer rows.Close()
	consumers, err := pgx.CollectRows(rows, pgx.RowToStructByPos[models.LightConsumer])
	if err != nil {
		return []models.LightConsumer{}, err
	}
	if len(consumers) == 0 {
		return []models.LightConsumer{}, nil
	}
	return consumers, nil
}

// Schema Functions
func GetSchemaByName(name string, tenantName string) (bool, models.Schema, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return false, models.Schema{}, err
	}
	defer conn.Release()
	query := `SELECT * FROM schemas WHERE name = $1 AND tenant_name = $2 LIMIT 1`
	stmt, err := conn.Conn().Prepare(ctx, "get_schema_by_name", query)
	if err != nil {
		return false, models.Schema{}, err
	}
	if tenantName != conf.GlobalAccount {
		tenantName = strings.ToLower(tenantName)
	}
	rows, err := conn.Conn().Query(ctx, stmt.Name, name, tenantName)
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
			TenantName:        strings.ToLower(v.TenantName),
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
		TenantName:        strings.ToLower(schemas[0].TenantName),
	}

	return schemaVersion, nil
}

func UpdateSchemasOfDeletedUser(userId int, tenantName string) error {
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
	AND created_by_username NOT LIKE '%(deleted)'
	AND tenant_name = $2`
	stmt, err := conn.Conn().Prepare(ctx, "update_schemas_of_deleted_user", query)
	if err != nil {
		return err
	}
	if tenantName != conf.GlobalAccount {
		tenantName = strings.ToLower(tenantName)
	}
	_, err = conn.Conn().Query(ctx, stmt.Name, userId, tenantName)
	if err != nil {
		return err
	}
	return nil
}

func RemoveSchemasByTenant(tenantName string) error {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()
	query := `DELETE FROM schemas WHERE tenant_name = $1`
	stmt, err := conn.Conn().Prepare(ctx, "remove_schemas_by_tenant", query)
	if err != nil {
		return err
	}
	_, err = conn.Conn().Exec(ctx, stmt.Name, tenantName)
	if err != nil {
		return err
	}
	return nil
}

func UpdateSchemaVersionsOfDeletedUser(userId int, tenantName string) error {
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
	AND created_by_username NOT LIKE '%(deleted)'
	AND tenant_name = $2`
	stmt, err := conn.Conn().Prepare(ctx, "update_schema_versions_of_deleted_user", query)
	if err != nil {
		return err
	}
	if tenantName != conf.GlobalAccount {
		tenantName = strings.ToLower(tenantName)
	}
	_, err = conn.Conn().Query(ctx, stmt.Name, userId, tenantName)
	if err != nil {
		return err
	}
	return nil
}

func RemoveSchemaVersionsByTenant(tenantName string) error {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	query := `DELETE FROM schema_versions WHERE tenant_name = $1`
	stmt, err := conn.Conn().Prepare(ctx, "remove_schema_versions_by_tenant", query)
	if err != nil {
		return err
	}
	_, err = conn.Conn().Query(ctx, stmt.Name, tenantName)
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
		TenantName:        strings.ToLower(schemas[0].TenantName),
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

func GetShcemaVersionsCount(schemaId int, tenantName string) (int, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return 0, err
	}
	defer conn.Release()
	query := `SELECT COUNT(*) FROM schema_versions WHERE schema_id=$1 AND tenant_name=$2`
	stmt, err := conn.Conn().Prepare(ctx, "get_schema_versions_count", query)
	if err != nil {
		return 0, err
	}
	var count int
	if tenantName != conf.GlobalAccount {
		tenantName = strings.ToLower(tenantName)
	}
	err = conn.Conn().QueryRow(ctx, stmt.Name, schemaId, tenantName).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func GetAllSchemasDetails(tenantName string) ([]models.ExtendedSchema, error) {
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
	          WHERE asv.id IS NOT NULL AND s.tenant_name = $1
	          ORDER BY sv.created_at DESC`
	stmt, err := conn.Conn().Prepare(ctx, "get_all_schemas_details", query)
	if err != nil {
		return []models.ExtendedSchema{}, err
	}
	if tenantName != conf.GlobalAccount {
		tenantName = strings.ToLower(tenantName)
	}
	rows, err := conn.Conn().Query(ctx, stmt.Name, tenantName)
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

func InsertNewSchema(schemaName string, schemaType string, createdByUsername string, tenantName string) (models.Schema, int64, error) {
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
		created_by_username,
		tenant_name) 
    VALUES($1, $2, $3, $4) RETURNING id`

	stmt, err := conn.Conn().Prepare(ctx, "insert_new_schema", query)
	if err != nil {
		return models.Schema{}, 0, err
	}

	var schemaId int
	if tenantName != conf.GlobalAccount {
		tenantName = strings.ToLower(tenantName)
	}
	rows, err := conn.Conn().Query(ctx, stmt.Name, schemaName, schemaType, createdByUsername, tenantName)
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
					return models.Schema{}, 0, errors.New("Schema " + schemaName + " already exists")
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

func InsertNewSchemaVersion(schemaVersionNumber int, userId int, username string, schemaContent string, schemaId int, messageStructName string, descriptor string, active bool, tenantName string) (models.SchemaVersion, int64, error) {
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
		descriptor,
		tenant_name)
    VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) RETURNING id`

	stmt, err := conn.Conn().Prepare(ctx, "insert_new_schema_version", query)
	if err != nil {
		return models.SchemaVersion{}, 0, err
	}

	var schemaVersionId int
	createdAt := time.Now()
	if tenantName != conf.GlobalAccount {
		tenantName = strings.ToLower(tenantName)
	}
	rows, err := conn.Conn().Query(ctx, stmt.Name, schemaVersionNumber, active, userId, username, createdAt, schemaContent, schemaId, messageStructName, []byte(descriptor), tenantName)
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

func CountAllSchemasByTenant(tenantName string) (int64, error) {
	var count int64
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return 0, err
	}
	defer conn.Release()
	query := `SELECT COUNT(*) FROM schemas WHERE tenant_name = $1`
	stmt, err := conn.Conn().Prepare(ctx, "get_total_schemas_by_tenant", query)
	if err != nil {
		return 0, err
	}
	err = conn.Conn().QueryRow(ctx, stmt.Name, tenantName).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

// Integration Functions
func GetIntegration(name string, tenantName string) (bool, models.Integration, error) {
	if tenantName != conf.GlobalAccount {
		tenantName = strings.ToLower(tenantName)
	}
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return false, models.Integration{}, err
	}
	defer conn.Release()
	query := `SELECT * FROM integrations WHERE name=$1 AND tenant_name=$2 LIMIT 1`
	stmt, err := conn.Conn().Prepare(ctx, "get_integration", query)
	if err != nil {
		return false, models.Integration{}, err
	}
	rows, err := conn.Conn().Query(ctx, stmt.Name, name, tenantName)
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

func GetAllIntegrationsByTenant(tenantName string) (bool, []models.Integration, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return false, []models.Integration{}, err
	}
	defer conn.Release()
	query := `SELECT * FROM integrations WHERE tenant_name = $1`
	stmt, err := conn.Conn().Prepare(ctx, "get_all_integrations_by_tenant", query)
	if err != nil {
		return false, []models.Integration{}, err
	}
	rows, err := conn.Conn().Query(ctx, stmt.Name, tenantName)
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

func DeleteIntegration(name string, tenantName string) error {
	if tenantName != conf.GlobalAccount {
		tenantName = strings.ToLower(tenantName)
	}
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()

	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	removeIntegrationQuery := `DELETE FROM integrations WHERE name = $1 AND tenant_name = $2`

	stmt, err := conn.Conn().Prepare(ctx, "remove_integration", removeIntegrationQuery)
	if err != nil {
		return err
	}

	_, err = conn.Conn().Exec(ctx, stmt.Name, name, tenantName)
	if err != nil {
		return err
	}

	return nil
}

func InsertNewIntegration(tenantName string, name string, keys map[string]interface{}, properties map[string]bool) (models.Integration, error) {
	if tenantName != conf.GlobalAccount {
		tenantName = strings.ToLower(tenantName)
	}
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
		properties,
		tenant_name,
		is_valid) 
    VALUES($1, $2, $3, $4, $5) RETURNING id`

	stmt, err := conn.Conn().Prepare(ctx, "insert_new_integration", query)
	if err != nil {
		return models.Integration{}, err
	}

	var integrationId int
	rows, err := conn.Conn().Query(ctx, stmt.Name, name, keys, properties, tenantName, true)
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
					return models.Integration{}, errors.New("Integration " + name + " already exists")
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
		TenantName: tenantName,
		IsValid:    true,
	}
	return newIntegration, nil
}

func UpdateIntegration(tenantName string, name string, keys map[string]interface{}, properties map[string]bool) (models.Integration, error) {
	if tenantName != conf.GlobalAccount {
		tenantName = strings.ToLower(tenantName)
	}
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return models.Integration{}, err
	}
	defer conn.Release()
	query := `
	INSERT INTO integrations(name, keys, properties, tenant_name)
	VALUES($1, $2, $3, $4)
	ON CONFLICT(name, tenant_name) DO UPDATE
	SET keys = excluded.keys, properties = excluded.properties
	RETURNING id, name, keys, properties, tenant_name, is_valid
`
	stmt, err := conn.Conn().Prepare(ctx, "update_integration", query)
	if err != nil {
		return models.Integration{}, err
	}
	rows, err := conn.Conn().Query(ctx, stmt.Name, name, keys, properties, tenantName)
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

func UpdateIsValidIntegration(tenantName, integrationName string, isValid bool) error {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()
	query := `UPDATE integrations SET is_valid = $2 WHERE name = $1 AND tenant_name=$3`
	stmt, err := conn.Conn().Prepare(ctx, "update_is_valid_integration", query)
	if err != nil {
		return err
	}
	if tenantName != conf.GlobalAccount {
		tenantName = strings.ToLower(tenantName)
	}
	_, err = conn.Conn().Query(ctx, stmt.Name, integrationName, isValid, tenantName)
	if err != nil {
		return err
	}
	return nil
}

// User Functions
func UpdatePendingUser(tenantName, username string, pending bool) error {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()
	query := `UPDATE users SET pending = $2 WHERE username = $1 AND tenant_name=$3`
	stmt, err := conn.Conn().Prepare(ctx, "update_pending_user", query)
	if err != nil {
		return err
	}
	if tenantName != conf.GlobalAccount {
		tenantName = strings.ToLower(tenantName)
	}
	_, err = conn.Conn().Query(ctx, stmt.Name, username, pending, tenantName)
	if err != nil {
		return err
	}

	return nil

}
func CreateUser(username string, userType string, hashedPassword string, fullName string, subscription bool, avatarId int, tenantName string, pending bool, team, position, owner, description string) (models.User, error) {
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
		skip_get_started,
		tenant_name,
		pending,
		team, 
		position,
		owner,
		description) 
    VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15) RETURNING id`

	stmt, err := conn.Conn().Prepare(ctx, "create_new_user", query)
	if err != nil {
		return models.User{}, err
	}
	createdAt := time.Now()
	skipGetStarted := false
	alreadyLoggedIn := false

	var userId int
	if tenantName != conf.GlobalAccount {
		tenantName = strings.ToLower(tenantName)
	}
	rows, err := conn.Conn().Query(ctx, stmt.Name, username, hashedPassword, userType, alreadyLoggedIn, createdAt, avatarId, fullName, subscription, skipGetStarted, tenantName, pending, team, position, owner, description)
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
	if tenantName != conf.GlobalAccount {
		tenantName = strings.ToLower(tenantName)
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
		TenantName:      tenantName,
		Pending:         pending,
		Team:            team,
		Position:        position,
		Owner:           owner,
		Description:     description,
	}
	return newUser, nil
}

func UpsertUserUpdatePassword(username string, userType string, hashedPassword string, fullName string, subscription bool, avatarId int, tenantName string) (bool, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()

	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return false, err
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
		skip_get_started,
		tenant_name
	) 
	VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) 
	ON CONFLICT (username, tenant_name) DO UPDATE SET password = excluded.password
	RETURNING id, CASE WHEN xmax = 0 THEN 'insert' ELSE 'update' END AS action`

	stmt, err := conn.Conn().Prepare(ctx, "upsert_new_user_update_password", query)
	if err != nil {
		return false, err
	}
	createdAt := time.Now()
	skipGetStarted := false
	alreadyLoggedIn := false

	var userId int
	var op string
	if tenantName != conf.GlobalAccount {
		tenantName = strings.ToLower(tenantName)
	}
	rows, err := conn.Conn().Query(ctx, stmt.Name, username, hashedPassword, userType, alreadyLoggedIn, createdAt, avatarId, fullName, subscription, skipGetStarted, tenantName)
	if err != nil {
		return false, err
	}
	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&userId, &op)
		if err != nil {
			return false, err
		}
	}

	didCreate := false
	if op == "insert" {
		didCreate = true
	}

	if err := rows.Err(); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Detail != "" {
				if strings.Contains(pgErr.Detail, "already exists") {
					return false, errors.New("User " + username + " already exists")
				} else {
					return false, errors.New(pgErr.Detail)
				}
			} else {
				return false, errors.New(pgErr.Message)
			}
		} else {
			return false, err
		}
	}

	return didCreate, nil
}

func ChangeUserPassword(username string, hashedPassword string, tenantName string) error {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()
	query := `UPDATE users SET password = $2 WHERE username = $1 AND tenant_name=$3`
	stmt, err := conn.Conn().Prepare(ctx, "change_user_password", query)
	if err != nil {
		return err
	}
	if tenantName != conf.GlobalAccount {
		tenantName = strings.ToLower(tenantName)
	}
	_, err = conn.Conn().Query(ctx, stmt.Name, username, hashedPassword, tenantName)
	if err != nil {
		return err
	}
	return nil
}

func GetRootUser(tenantName string) (bool, models.User, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return false, models.User{}, err
	}
	defer conn.Release()
	query := `SELECT * FROM users WHERE type = 'root' and tenant_name =$1 LIMIT 1`
	stmt, err := conn.Conn().Prepare(ctx, "get_root_user", query)
	if err != nil {
		return false, models.User{}, err
	}
	if tenantName != conf.GlobalAccount {
		tenantName = strings.ToLower(tenantName)
	}
	rows, err := conn.Conn().Query(ctx, stmt.Name, tenantName)
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

func GetUserByUsername(username string, tenantName string) (bool, models.User, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return false, models.User{}, err
	}
	defer conn.Release()
	query := `SELECT * FROM users WHERE username = $1 AND tenant_name = $2 LIMIT 1`
	stmt, err := conn.Conn().Prepare(ctx, "get_user_by_username", query)
	if err != nil {
		return false, models.User{}, err
	}
	if tenantName != conf.GlobalAccount {
		tenantName = strings.ToLower(tenantName)
	}
	rows, err := conn.Conn().Query(ctx, stmt.Name, username, tenantName)
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

func GetUserForLogin(username string) (bool, models.User, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return false, models.User{}, err
	}
	defer conn.Release()
	query := `SELECT * FROM users WHERE username = $1 AND NOT type = 'application' LIMIT 1`
	stmt, err := conn.Conn().Prepare(ctx, "get_user_for_login", query)
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

func GetUserForLoginByUsernameAndTenant(username, tenantname string) (bool, models.User, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return false, models.User{}, err
	}
	defer conn.Release()
	query := `SELECT * FROM users WHERE username = $1 AND tenant_name = $2  AND NOT type = 'application' LIMIT 1`
	stmt, err := conn.Conn().Prepare(ctx, "get_user_for_login_by_username_and_tenant", query)
	if err != nil {
		return false, models.User{}, err
	}
	rows, err := conn.Conn().Query(ctx, stmt.Name, username, tenantname)
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

func GetAllUsers(tenantName string) ([]models.FilteredGenericUser, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return []models.FilteredGenericUser{}, err
	}
	defer conn.Release()
	query := `SELECT s.id, s.username, s.type, s.created_at, s.avatar_id, s.full_name, s.pending, s.position, s.team, s.owner, s.description FROM users AS s WHERE tenant_name=$1 AND username NOT LIKE '$%'` // filter memphis internal users
	stmt, err := conn.Conn().Prepare(ctx, "get_all_users", query)
	if err != nil {
		return []models.FilteredGenericUser{}, err
	}
	if tenantName != conf.GlobalAccount {
		tenantName = strings.ToLower(tenantName)
	}
	rows, err := conn.Conn().Query(ctx, stmt.Name, tenantName)
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

func CountAllUsers() (int64, error) {
	var count int64
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return 0, err
	}
	defer conn.Release()
	query := `SELECT COUNT(*) FROM users WHERE username NOT LIKE '$%'` // filter memphis internal users
	stmt, err := conn.Conn().Prepare(ctx, "get_total_users", query)
	if err != nil {
		return 0, err
	}
	err = conn.Conn().QueryRow(ctx, stmt.Name).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func CountAllUsersByTenant(tenantName string) (int64, error) {
	var count int64
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return 0, err
	}
	defer conn.Release()
	query := `SELECT COUNT(*) FROM users WHERE tenant_name = $1 AND username NOT LIKE '$%'` // filter memphis internal users`
	stmt, err := conn.Conn().Prepare(ctx, "get_total_users_by_tenant", query)
	if err != nil {
		return 0, err
	}
	err = conn.Conn().QueryRow(ctx, stmt.Name, tenantName).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func GetAllUsersByTypeAndTenantName(userType []string, tenantName string) ([]models.User, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return []models.User{}, err
	}
	defer conn.Release()
	var rows pgx.Rows
	query := `SELECT * FROM users WHERE type=ANY($1) AND tenant_name=$2 AND username NOT LIKE '$%'` // filter memphis internal users
	stmt, err := conn.Conn().Prepare(ctx, "get_all_users_by_type_and_tenant_name", query)
	if err != nil {
		return []models.User{}, err
	}
	if tenantName != conf.GlobalAccount {
		tenantName = strings.ToLower(tenantName)
	}
	rows, err = conn.Conn().Query(ctx, stmt.Name, userType, tenantName)
	if err != nil {
		return []models.User{}, err
	}

	defer rows.Close()
	users, err := pgx.CollectRows(rows, pgx.RowToStructByPos[models.User])
	if err != nil {
		return []models.User{}, err
	}
	if len(users) == 0 {
		return []models.User{}, nil
	}
	return users, nil
}

func GetAllUsersByTenantName(tenantName string) ([]models.User, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return []models.User{}, err
	}
	defer conn.Release()
	var rows pgx.Rows
	query := `SELECT * FROM users WHERE tenant_name=$1 AND username NOT LIKE '$%'` // filter memphis internal users
	stmt, err := conn.Conn().Prepare(ctx, "get_all_users_by_type_and_tenant_name", query)
	if err != nil {
		return []models.User{}, err
	}

	rows, err = conn.Conn().Query(ctx, stmt.Name, tenantName)
	if err != nil {
		return []models.User{}, err
	}

	defer rows.Close()
	users, err := pgx.CollectRows(rows, pgx.RowToStructByPos[models.User])
	if err != nil {
		return []models.User{}, err
	}
	if len(users) == 0 {
		return []models.User{}, nil
	}
	return users, nil
}

func GetAllUsersByType(userType []string) ([]models.User, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return []models.User{}, err
	}
	defer conn.Release()
	var rows pgx.Rows
	query := `SELECT * FROM users WHERE type=ANY($1) AND username NOT LIKE '$%'` // filter memphis internal users
	stmt, err := conn.Conn().Prepare(ctx, "get_all_users_by_type", query)
	if err != nil {
		return []models.User{}, err
	}
	rows, err = conn.Conn().Query(ctx, stmt.Name, userType)
	if err != nil {
		return []models.User{}, err
	}

	defer rows.Close()
	users, err := pgx.CollectRows(rows, pgx.RowToStructByPos[models.User])
	if err != nil {
		return []models.User{}, err
	}
	if len(users) == 0 {
		return []models.User{}, nil
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

func UpdateSkipGetStarted(username string, tenantName string) error {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()
	query := `UPDATE users SET skip_get_started = true WHERE username = $1 AND tenant_name=$2`
	stmt, err := conn.Conn().Prepare(ctx, "update_skip_get_started", query)
	if err != nil {
		return err
	}
	if tenantName != conf.GlobalAccount {
		tenantName = strings.ToLower(tenantName)
	}
	_, err = conn.Conn().Query(ctx, stmt.Name, username, tenantName)
	if err != nil {
		return err
	}
	return nil
}

func DeleteUser(username string, tenantName string) error {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()

	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	removeUserQuery := `DELETE FROM users WHERE username = $1 AND tenant_name=$2`

	stmt, err := conn.Conn().Prepare(ctx, "remove_user", removeUserQuery)
	if err != nil {
		return err
	}
	if tenantName != conf.GlobalAccount {
		tenantName = strings.ToLower(tenantName)
	}
	_, err = conn.Conn().Exec(ctx, stmt.Name, username, tenantName)
	if err != nil {
		return err
	}

	return nil
}

func DeleteUsersByTenant(tenantName string) ([]string, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()

	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	removeUserQuery := `DELETE FROM users WHERE tenant_name = $1 RETURNING username`
	rows, err := conn.Query(ctx, removeUserQuery, tenantName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users_list []string

	for rows.Next() {
		var usermame string
		err := rows.Scan(&usermame)
		if err != nil {
			return nil, err
		}
		if !strings.HasPrefix(usermame, "$") { // skip memphis internal users
			users_list = append(users_list, usermame)
		}
	}

	return users_list, err
}

func EditAvatar(username string, avatarId int, tenantName string) error {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()
	query := `UPDATE users SET avatar_id = $2 WHERE username = $1 AND tenant_name=$3`
	stmt, err := conn.Conn().Prepare(ctx, "edit_avatar", query)
	if err != nil {
		return err
	}
	if tenantName != conf.GlobalAccount {
		tenantName = strings.ToLower(tenantName)
	}
	_, err = conn.Conn().Query(ctx, stmt.Name, username, avatarId, tenantName)
	if err != nil {
		return err
	}
	return nil
}

func GetAllActiveUsersStations(tenantName string) ([]models.FilteredUser, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return []models.FilteredUser{}, err
	}
	defer conn.Release()
	query := `
	SELECT DISTINCT u.username
	FROM users AS u
	JOIN stations AS s ON u.id = s.created_by
	WHERE s.tenant_name=$1 AND username NOT LIKE '$%'` // filter memphis internal users

	stmt, err := conn.Conn().Prepare(ctx, "get_all_active_users_stations", query)
	if err != nil {
		return []models.FilteredUser{}, err
	}
	if tenantName != conf.GlobalAccount {
		tenantName = strings.ToLower(tenantName)
	}
	rows, err := conn.Conn().Query(ctx, stmt.Name, tenantName)
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

func GetAllActiveUsersSchemaVersions(tenantName string) ([]models.FilteredUser, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return []models.FilteredUser{}, err
	}
	defer conn.Release()
	query := `
	SELECT DISTINCT u.username
	FROM users AS u
	JOIN schema_versions AS s ON u.id = s.created_by
	WHERE s.tenant_name=$1 AND username NOT LIKE '$%'` // filter memphis internal users

	stmt, err := conn.Conn().Prepare(ctx, "get_all_active_users_schema_versions", query)
	if err != nil {
		return []models.FilteredUser{}, err
	}
	if tenantName != conf.GlobalAccount {
		tenantName = strings.ToLower(tenantName)
	}
	rows, err := conn.Conn().Query(ctx, stmt.Name, tenantName)
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

func UpsertBatchOfUsers(users []models.User) error {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()
	valueStrings := make([]string, 0, len(users))
	valueArgs := make([]interface{}, 0, len(users)*6)
	for i, user := range users {
		valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d)", i*6+1, i*6+2, i*6+3, i*6+4, i*6+5, i*6+6))
		valueArgs = append(valueArgs, user.Username)
		valueArgs = append(valueArgs, user.Password)
		valueArgs = append(valueArgs, user.UserType)
		valueArgs = append(valueArgs, user.CreatedAt)
		valueArgs = append(valueArgs, user.AvatarId)
		valueArgs = append(valueArgs, user.FullName)
	}
	query := fmt.Sprintf("INSERT INTO users (username, password, type, created_at, avatar_id, full_name) VALUES %s ON CONFLICT (username) DO UPDATE SET password = EXCLUDED.password", strings.Join(valueStrings, ","))
	stmt, err := conn.Conn().Prepare(ctx, "update_batch_of_users", query)
	if err != nil {
		return err
	}
	_, err = conn.Conn().Query(ctx, stmt.Name, valueArgs...)
	if err != nil {
		return err
	}
	return nil
}

// Tags Functions
func InsertNewTag(name string, color string, stationArr []int, schemaArr []int, userArr []int, tenantName string) (models.Tag, error) {
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
		schemas,
		tenant_name) 
    VALUES($1, $2, $3, $4, $5, $6) RETURNING id`

	stmt, err := conn.Conn().Prepare(ctx, "insert_new_tag", query)
	if err != nil {
		return models.Tag{}, err
	}

	var tagId int
	if tenantName != conf.GlobalAccount {
		tenantName = strings.ToLower(tenantName)
	}
	rows, err := conn.Conn().Query(ctx, stmt.Name, name, color, userArr, stationArr, schemaArr, tenantName)
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
	if tenantName != conf.GlobalAccount {
		tenantName = strings.ToLower(tenantName)
	}
	newTag := models.Tag{
		ID:         tagId,
		Name:       name,
		Color:      color,
		Stations:   stationArr,
		Schemas:    schemaArr,
		Users:      userArr,
		TenantName: tenantName,
	}
	return newTag, nil
}

func InsertEntityToTag(tagName string, entity string, entity_id int, tenantName, color string) error {
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
	if color == "" {
		query := `UPDATE tags SET ` + entityDBList + ` = ARRAY_APPEND(` + entityDBList + `, $1) WHERE name = $2 AND tenant_name = $3`
		stmt, err := conn.Conn().Prepare(ctx, "insert_entity_to_tag", query)
		if err != nil {
			return err
		}
		_, err = conn.Conn().Query(ctx, stmt.Name, entity_id, tagName, tenantName)
		if err != nil {
			return err
		}
	} else {
		query := `UPDATE tags SET ` + entityDBList + ` = ARRAY_APPEND(` + entityDBList + `, $1) , color = $4 WHERE name = $2 AND tenant_name = $3`
		stmt, err := conn.Conn().Prepare(ctx, "insert_entity_to_tag", query)
		if err != nil {
			return err
		}
		_, err = conn.Conn().Query(ctx, stmt.Name, entity_id, tagName, tenantName, color)
		if err != nil {
			return err
		}
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
	query := `SELECT * FROM tags AS t WHERE $1 = ANY(t.` + entityDBList + `) LIMIT 1000;`
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

func GetTagsByEntityIDLight(entity string, id int) ([]models.CreateTag, error) {
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
		return []models.CreateTag{}, err
	}
	defer conn.Release()
	uid, err := uuid.NewV4()
	if err != nil {
		return []models.CreateTag{}, err
	}
	query := `SELECT t.name, t.color FROM tags AS t WHERE $1 = ANY(t.` + entityDBList + `) LIMIT 1000;`
	stmt, err := conn.Conn().Prepare(ctx, "get_tags_by_entity_id_light"+uid.String(), query)
	if err != nil {
		return []models.CreateTag{}, err
	}
	rows, err := conn.Conn().Query(ctx, stmt.Name, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	tags, err := pgx.CollectRows(rows, pgx.RowToStructByPos[models.CreateTag])
	if err != nil {
		return []models.CreateTag{}, err
	}
	if len(tags) == 0 {
		return []models.CreateTag{}, err
	}
	return tags, nil
}

func GetTagsByEntityType(entity string, tenantName string) ([]models.Tag, error) {
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
		query = `SELECT * FROM tags WHERE tenant_name=$1`
	} else {
		query = fmt.Sprintf(`SELECT * FROM tags WHERE %s IS NOT NULL AND array_length(%s, 1) > 0 AND tenant_name=$1`, entityDBList, entityDBList)
	}
	stmt, err := conn.Conn().Prepare(ctx, "get_tags_by_entity_type", query)
	if err != nil {
		return []models.Tag{}, err
	}
	if tenantName != conf.GlobalAccount {
		tenantName = strings.ToLower(tenantName)
	}
	rows, err := conn.Conn().Query(ctx, stmt.Name, tenantName)
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

func GetAllUsedTags(tenantName string) ([]models.Tag, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Release()
	query := `SELECT * FROM tags WHERE (ARRAY_LENGTH(schemas, 1) > 0 OR ARRAY_LENGTH(stations, 1) > 0 OR ARRAY_LENGTH(users, 1) > 0) AND tenant_name=$1`
	stmt, err := conn.Conn().Prepare(ctx, "get_all_used_tags", query)
	if err != nil {
		return nil, err
	}
	if tenantName != conf.GlobalAccount {
		tenantName = strings.ToLower(tenantName)
	}
	rows, err := conn.Conn().Query(ctx, stmt.Name, tenantName)
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

func GetAllUsedStationsTags(tenantName string) ([]models.Tag, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Release()
	query := `SELECT * FROM tags WHERE (ARRAY_LENGTH(stations, 1) > 0) AND tenant_name=$1`
	stmt, err := conn.Conn().Prepare(ctx, "get_all_used_stations_tags", query)
	if err != nil {
		return nil, err
	}
	if tenantName != conf.GlobalAccount {
		tenantName = strings.ToLower(tenantName)
	}
	rows, err := conn.Conn().Query(ctx, stmt.Name, tenantName)
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

func GetAllUsedSchemasTags(tenantName string) ([]models.Tag, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Release()
	query := `SELECT * FROM tags WHERE (ARRAY_LENGTH(schemas, 1) > 0) AND tenant_name=$1`
	stmt, err := conn.Conn().Prepare(ctx, "get_all_used_schemas_tags", query)
	if err != nil {
		return nil, err
	}
	if tenantName != conf.GlobalAccount {
		tenantName = strings.ToLower(tenantName)
	}
	rows, err := conn.Conn().Query(ctx, stmt.Name, tenantName)
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

func GetTagByName(name string, tenantName string) (bool, models.Tag, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return false, models.Tag{}, err
	}
	defer conn.Release()
	query := `SELECT * FROM tags WHERE name=$1 AND tenant_name=$2 LIMIT 1`
	stmt, err := conn.Conn().Prepare(ctx, "get_tag_by_name", query)
	if err != nil {
		return false, models.Tag{}, err
	}
	if tenantName != conf.GlobalAccount {
		tenantName = strings.ToLower(tenantName)
	}
	rows, err := conn.Conn().Query(ctx, stmt.Name, name, tenantName)
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

// Image Functions
func InsertImage(name string, base64Encoding string, tenantName string) error {
	if tenantName != conf.GlobalAccount {
		tenantName = strings.ToLower(tenantName)
	}
	err := UpsertConfiguration(name, base64Encoding, tenantName)
	if err != nil {
		return err
	}
	return nil
}

func DeleteImage(name string, tenantName string) error {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()

	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	removeImageQuery := `DELETE FROM configurations
	WHERE key = $1 AND tenant_name=$2`

	stmt, err := conn.Conn().Prepare(ctx, "remove_image", removeImageQuery)
	if err != nil {
		return err
	}
	if tenantName != conf.GlobalAccount {
		tenantName = strings.ToLower(tenantName)
	}
	_, err = conn.Conn().Exec(ctx, stmt.Name, name, tenantName)
	if err != nil {
		return err
	}
	return nil
}

func GetImage(name string, tenantName string) (bool, models.Image, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return false, models.Image{}, err
	}
	defer conn.Release()
	query := `SELECT id, key, value FROM configurations WHERE key = $1 AND tenant_name =$2 LIMIT 1`
	stmt, err := conn.Conn().Prepare(ctx, "get_image", query)
	if err != nil {
		return false, models.Image{}, err
	}
	if tenantName != conf.GlobalAccount {
		tenantName = strings.ToLower(tenantName)
	}
	rows, err := conn.Conn().Query(ctx, stmt.Name, name, tenantName)
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
func InsertSchemaverseDlsMsg(stationId int, messageSeq int, producerName string, poisonedCgs []string, messageDetails models.MessagePayload, validationError string, tenantName string, partitionNumber int) (models.DlsMessage, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	connection, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return models.DlsMessage{}, err
	}
	defer connection.Release()

	query := `INSERT INTO dls_messages( 
			station_id,
			message_seq,
			poisoned_cgs,
			message_details,
			updated_at,
			message_type,
			validation_error,
			tenant_name,
			producer_name,
			partition_number) 
		VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id`

	stmt, err := connection.Conn().Prepare(ctx, "insert_dls_messages", query)
	if err != nil {
		return models.DlsMessage{}, err
	}
	updatedAt := time.Now()
	if tenantName != conf.GlobalAccount {
		tenantName = strings.ToLower(tenantName)
	}
	rows, err := connection.Conn().Query(ctx, stmt.Name, stationId, messageSeq, poisonedCgs, messageDetails, updatedAt, "schema", validationError, tenantName, producerName, partitionNumber)
	if err != nil {
		return models.DlsMessage{}, err
	}
	defer rows.Close()
	var messagePayloadId int
	for rows.Next() {
		err := rows.Scan(&messagePayloadId)
		if err != nil {
			return models.DlsMessage{}, err
		}
	}

	if err != nil {
		return models.DlsMessage{}, err
	}
	if tenantName != conf.GlobalAccount {
		tenantName = strings.ToLower(tenantName)
	}
	deadLetterPayload := models.DlsMessage{
		ID:           messagePayloadId,
		StationId:    stationId,
		MessageSeq:   messageSeq,
		ProducerName: producerName,
		PoisonedCgs:  poisonedCgs,
		MessageDetails: models.MessagePayload{
			TimeSent: messageDetails.TimeSent,
			Size:     messageDetails.Size,
			Data:     messageDetails.Data,
			Headers:  messageDetails.Headers,
		},
		UpdatedAt:       updatedAt,
		TenantName:      tenantName,
		PartitionNumber: partitionNumber,
	}

	if err := rows.Err(); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Detail != "" {
				if !strings.Contains(pgErr.Detail, "already exists") {
					return models.DlsMessage{}, errors.New("schemaverse dls already exists")
				} else {
					return models.DlsMessage{}, errors.New(pgErr.Detail)
				}
			} else {
				return models.DlsMessage{}, errors.New(pgErr.Message)
			}
		} else {
			return models.DlsMessage{}, err
		}
	}

	return deadLetterPayload, nil
}

func GetMsgByStationIdAndMsgSeq(stationId, messageSeq, partitionNumber int) (bool, models.DlsMessage, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()

	connection, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return false, models.DlsMessage{}, err
	}
	defer connection.Release()

	query := `SELECT * FROM dls_messages WHERE station_id = $1 AND message_seq = $2 AND partition_number = $3 LIMIT 1`

	stmt, err := connection.Conn().Prepare(ctx, "get_dls_messages_by_station_id_and_message_seq", query)
	if err != nil {
		return false, models.DlsMessage{}, err
	}

	rows, err := connection.Conn().Query(ctx, stmt.Name, stationId, messageSeq, partitionNumber)
	if err != nil {
		return false, models.DlsMessage{}, err
	}
	defer rows.Close()

	message, err := pgx.CollectRows(rows, pgx.RowToStructByPos[models.DlsMessage])
	if err != nil {
		return false, models.DlsMessage{}, err
	}
	if len(message) == 0 {
		return false, models.DlsMessage{}, nil
	}

	return true, message[0], nil
}

func StorePoisonMsg(stationId, messageSeq int, cgName string, producerName string, poisonedCgs []string, messageDetails models.MessagePayload, tenantName string, partitionNumber int) (int, bool, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	updated := false
	connection, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return 0, updated, err
	}
	defer connection.Release()

	tx, err := connection.Conn().Begin(ctx)
	if err != nil {
		return 0, updated, err
	}
	defer tx.Rollback(ctx)

	query := `SELECT EXISTS(SELECT 1 FROM consumers WHERE tenant_name = $1 AND consumers_group = $2)`
	stmt, err := tx.Prepare(ctx, "check_if_consumer_exists", query)
	if err != nil {
		return 0, updated, err
	}
	if tenantName != conf.GlobalAccount {
		tenantName = strings.ToLower(tenantName)
	}
	checkRows, err := tx.Query(ctx, stmt.Name, tenantName, cgName)
	if err != nil {
		return 0, updated, err
	}
	var exists bool
	for checkRows.Next() {
		checkRows.Scan(&exists)
	}
	if !exists {
		return 0, updated, err
	}
	checkRows.Close()

	query = `SELECT * FROM dls_messages WHERE station_id = $1 AND message_seq = $2 AND tenant_name =$3 AND partition_number = $4 LIMIT 1 FOR UPDATE`
	stmt, err = tx.Prepare(ctx, "handle_insert_dls_message", query)
	if err != nil {
		return 0, updated, err
	}
	rows, err := tx.Query(ctx, stmt.Name, stationId, messageSeq, tenantName, partitionNumber)
	if err != nil {
		return 0, updated, err
	}

	message, err := pgx.CollectRows(rows, pgx.RowToStructByPos[models.DlsMessage])
	if err != nil {
		return 0, updated, err
	}
	defer rows.Close()

	var dlsMsgId int
	if len(message) == 0 { // then insert
		query = `INSERT INTO dls_messages( 
			station_id,
			message_seq,
			producer_name,
			poisoned_cgs,
			message_details,
			updated_at,
			message_type,
			validation_error,
			tenant_name,
			partition_number
			) 
		VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id`

		stmt, err := tx.Prepare(ctx, "insert_dls_message", query)
		if err != nil {
			return 0, updated, err
		}
		updatedAt := time.Now()
		if tenantName != conf.GlobalAccount {
			tenantName = strings.ToLower(tenantName)
		}
		rows, err := tx.Query(ctx, stmt.Name, stationId, messageSeq, producerName, poisonedCgs, messageDetails, updatedAt, "poison", "", tenantName, partitionNumber)
		if err != nil {
			return 0, updated, err
		}
		defer rows.Close()

		for rows.Next() {
			err := rows.Scan(&dlsMsgId)
			if err != nil {
				return 0, updated, err
			}
		}
		if err := rows.Err(); err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) {
				if pgErr.Detail != "" {
					if strings.Contains(pgErr.Detail, "already exists") {
						return 0, updated, errors.New("dls_messages row already exists")
					} else {
						return 0, updated, errors.New(pgErr.Detail)
					}
				} else {
					return 0, updated, errors.New(pgErr.Message)
				}
			} else {
				return 0, updated, err
			}
		}
	} else { // then update
		query = `UPDATE dls_messages SET poisoned_cgs = ARRAY_APPEND(poisoned_cgs, $1), updated_at = $4 WHERE station_id=$2 AND message_seq=$3 AND not($1 = ANY(poisoned_cgs)) AND tenant_name=$5 RETURNING id`
		stmt, err := tx.Prepare(ctx, "update_poisoned_cgs", query)
		if err != nil {
			return 0, updated, err
		}
		updatedAt := time.Now()
		if tenantName != conf.GlobalAccount {
			tenantName = strings.ToLower(tenantName)
		}
		rows, err = tx.Query(ctx, stmt.Name, poisonedCgs[0], stationId, messageSeq, updatedAt, tenantName)
		if err != nil {
			return 0, updated, err
		}
		defer rows.Close()

		for rows.Next() {
			err := rows.Scan(&dlsMsgId)
			if err != nil {
				return 0, updated, err
			}
		}
		if err := rows.Err(); err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) {
				return 0, updated, errors.New(pgErr.Message)
			} else {
				return 0, updated, err
			}
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		return 0, updated, err
	}

	return dlsMsgId, updated, nil
}

func GetTotalPoisonMsgsPerCg(cgName string, stationId int) (int, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return 0, err
	}
	defer conn.Release()
	query := `SELECT COUNT(*) FROM dls_messages WHERE $1 = ANY(poisoned_cgs) AND station_id = $2`
	stmt, err := conn.Conn().Prepare(ctx, "get_total_poison_msgs_per_cg", query)
	if err != nil {
		return 0, err
	}
	var count int
	err = conn.Conn().QueryRow(ctx, stmt.Name, cgName, stationId).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func DeleteOldDlsMessageByRetention(updatedAt time.Time, tenantName string) error {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	query := `DELETE FROM dls_messages WHERE tenant_name = $1 AND updated_at < $2`
	stmt, err := conn.Conn().Prepare(ctx, "delete_old_dls_messages", query)
	if err != nil {
		return err
	}
	_, err = conn.Conn().Exec(ctx, stmt.Name, tenantName, updatedAt)
	if err != nil {
		return err
	}
	return nil
}

func DropDlsMessages(messageIds []int) error {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return errors.New("dropSchemaDlsMsg: " + err.Error())
	}
	defer conn.Release()

	query := `DELETE FROM dls_messages where id=ANY($1)`
	stmt, err := conn.Conn().Prepare(ctx, "drop_dls_schema_msg", query)
	if err != nil {
		return err
	}

	_, err = conn.Conn().Exec(ctx, stmt.Name, messageIds)
	if err != nil {
		return errors.New("dropSchemaDlsMsg: " + err.Error())
	}
	return nil
}

func PurgeDlsMsgsFromStation(station_id int) error {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return errors.New("PurgeDlsMsgsFromStation: " + err.Error())
	}
	defer conn.Release()

	query := `DELETE FROM dls_messages where station_id=$1`
	stmt, err := conn.Conn().Prepare(ctx, "purge_dls_messages", query)
	if err != nil {
		return err
	}

	_, err = conn.Conn().Exec(ctx, stmt.Name, station_id)
	if err != nil {
		return errors.New("PurgeDlsMsgsFromStation: " + err.Error())
	}
	return nil
}

func PurgeDlsMsgsFromPartition(station_id, partitionNumber int) error {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return errors.New("PurgeDlsMsgsFromPartition: " + err.Error())
	}
	defer conn.Release()

	query := `DELETE FROM dls_messages where station_id=$1 AND partition_number = $2`
	stmt, err := conn.Conn().Prepare(ctx, "purge_dls_messages", query)
	if err != nil {
		return err
	}

	_, err = conn.Conn().Exec(ctx, stmt.Name, station_id, partitionNumber)
	if err != nil {
		return errors.New("PurgeDlsMsgsFromPartition: " + err.Error())
	}
	return nil
}

func RemoveCgFromDlsMsg(msgId int, cgName string, tenantName string) error {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	query := `WITH removed_value AS (UPDATE dls_messages SET poisoned_cgs = ARRAY_REMOVE(poisoned_cgs, $1) WHERE id = $2 AND tenant_name = $3 RETURNING *) 
	DELETE FROM dls_messages WHERE (SELECT array_length(poisoned_cgs, 1)) <= 1 AND id = $2;`

	stmt, err := conn.Conn().Prepare(ctx, "get_msg_by_id_and_remove msg", query)
	if err != nil {
		return err
	}
	_, err = conn.Conn().Query(ctx, stmt.Name, cgName, msgId, tenantName)
	if err != nil {
		return err
	}

	return nil
}

func CountDlsMsgsByStationAndPartition(stationId, partitionNumber int) (int, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return 0, err
	}
	defer conn.Release()
	query := `SELECT COUNT(*) from dls_messages where station_id=$1`
	stmtName := "count_dls_msgs_by_station_and_partition"
	if partitionNumber > -1 {
		query = `SELECT COUNT(*) from dls_messages where station_id=$1 AND partition_number = $2`
		stmtName = "count_all_dls_msgs_by_station"
	}

	stmt, err := conn.Conn().Prepare(ctx, stmtName, query)
	if err != nil {
		return 0, err
	}

	var count uint64
	if partitionNumber == -1 {
		err = conn.Conn().QueryRow(ctx, stmt.Name, stationId).Scan(&count)
	} else {
		err = conn.Conn().QueryRow(ctx, stmt.Name, stationId, partitionNumber).Scan(&count)
	}
	if err != nil {
		return 0, err
	}

	return int(count), nil
}

func GetDlsMessageById(messageId int) (bool, models.DlsMessage, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return false, models.DlsMessage{}, err
	}
	defer conn.Release()
	query := `SELECT * from dls_messages where id=$1 LIMIT 1`
	stmt, err := conn.Conn().Prepare(ctx, "get_dls_msg_by_id", query)
	if err != nil {
		return false, models.DlsMessage{}, err
	}
	rows, err := conn.Conn().Query(ctx, stmt.Name, messageId)
	if err != nil {
		return false, models.DlsMessage{}, err
	}
	defer rows.Close()
	dlsMsgs, err := pgx.CollectRows(rows, pgx.RowToStructByPos[models.DlsMessage])
	if err != nil {
		return false, models.DlsMessage{}, err
	}
	if len(dlsMsgs) == 0 {
		return false, models.DlsMessage{}, nil
	}
	return true, dlsMsgs[0], nil
}

func GetTotalDlsMessages(tenantName string) (uint64, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return 0, err
	}
	defer conn.Release()
	query := `SELECT COUNT(*) FROM dls_messages WHERE tenant_name=$1`
	stmt, err := conn.Conn().Prepare(ctx, "get_total_dls_msgs", query)
	if err != nil {
		return 0, err
	}
	var count uint64
	if tenantName != conf.GlobalAccount {
		tenantName = strings.ToLower(tenantName)
	}
	err = conn.Conn().QueryRow(ctx, stmt.Name, tenantName).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func GetStationIdsFromDlsMsgs(tenantName string) (map[int]string, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	stationIds := map[int]string{}
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return stationIds, err
	}
	defer conn.Release()
	query := `SELECT DISTINCT station_id FROM dls_messages WHERE tenant_name=$1`
	stmt, err := conn.Conn().Prepare(ctx, "get_station_ids_in_dls_messages", query)
	if err != nil {
		return stationIds, err
	}
	if tenantName != conf.GlobalAccount {
		tenantName = strings.ToLower(tenantName)
	}
	rows, err := conn.Conn().Query(ctx, stmt.Name, tenantName)
	if err != nil {
		return stationIds, err
	}
	defer rows.Close()

	for rows.Next() {
		var stationId int
		err := rows.Scan(&stationId)
		if err != nil {
			return stationIds, err
		}
		stationIds[stationId] = ""
	}
	return stationIds, nil
}

func DeleteDlsMsgsByTenant(tenantName string) error {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()

	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	removeUserQuery := `DELETE FROM dls_messages WHERE tenant_name = $1`

	stmt, err := conn.Conn().Prepare(ctx, "remove_dls_msgs_by_tenant", removeUserQuery)
	if err != nil {
		return err
	}
	_, err = conn.Conn().Exec(ctx, stmt.Name, tenantName)
	if err != nil {
		return err
	}

	return nil

}

func GetMinMaxIdsOfDlsMsgsByUpdatedAt(tenantName string, updatedAt time.Time, stationId int) (int, int, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return -1, -1, err
	}
	defer conn.Release()
	query := `
		WITH limited_rows AS (
		SELECT id
		FROM dls_messages
		WHERE tenant_name = $1 AND updated_at <= $2 AND message_type = 'poison' AND station_id = $3
		)
		SELECT
		(SELECT MIN(id) FROM limited_rows) AS min_id,
		(SELECT MAX(id) FROM limited_rows) AS max_id
		FROM limited_rows;`
	stmt, err := conn.Conn().Prepare(ctx, "get_min_max_id_dls_msgs_by_updated_at", query)
	if err != nil {
		return -1, -1, err
	}
	if tenantName != conf.GlobalAccount {
		tenantName = strings.ToLower(tenantName)
	}
	rows, err := conn.Conn().Query(ctx, stmt.Name, tenantName, updatedAt, stationId)
	if err != nil {
		return -1, -1, err
	}
	defer rows.Close()

	var minId, maxId int
	if rows.Next() {
		err := rows.Scan(&minId, &maxId)
		if err != nil {
			return -1, -1, err
		}
	}
	return minId, maxId, nil
}

func GetDlsMsgsBatch(tenantName string, min, max, stationId int) (bool, []models.DlsMessage, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return false, []models.DlsMessage{}, err
	}
	defer conn.Release()
	query := `
		SELECT *
		FROM dls_messages
		WHERE tenant_name = $1 AND message_type = 'poison' AND station_id = $2 AND id > $3 AND id <= $4 ORDER BY id ASC LIMIT 10
		`
	stmt, err := conn.Conn().Prepare(ctx, "get_dls_msgs_batch_between_min_max_id", query)
	if err != nil {
		return false, []models.DlsMessage{}, err
	}
	if tenantName != conf.GlobalAccount {
		tenantName = strings.ToLower(tenantName)
	}
	rows, err := conn.Conn().Query(ctx, stmt.Name, tenantName, stationId, min, max)
	if err != nil {
		return false, []models.DlsMessage{}, err
	}
	defer rows.Close()

	dlsMsgs, err := pgx.CollectRows(rows, pgx.RowToStructByPos[models.DlsMessage])
	if err != nil {
		return false, []models.DlsMessage{}, err
	}
	if len(dlsMsgs) == 0 {
		return false, []models.DlsMessage{}, nil
	}
	return true, dlsMsgs, nil
}

// Tenants functions
func UpsertTenant(name string, encryptrdInternalWSPass string) (models.Tenant, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()

	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return models.Tenant{}, err
	}
	defer conn.Release()

	query := `INSERT INTO tenants (name, internal_ws_pass) VALUES($1, $2) ON CONFLICT (name) DO NOTHING`

	stmt, err := conn.Conn().Prepare(ctx, "upsert_tenant", query)
	if err != nil {
		return models.Tenant{}, err
	}

	var tenantId int
	rows, err := conn.Conn().Query(ctx, stmt.Name, name, encryptrdInternalWSPass)
	if err != nil {
		return models.Tenant{}, err
	}
	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&tenantId)
		if err != nil {
			return models.Tenant{}, err
		}
	}

	if err := rows.Err(); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Detail != "" {
				if strings.Contains(pgErr.Detail, "already exists") {
					return models.Tenant{}, errors.New("Tenant " + name + " already exists")
				} else {
					return models.Tenant{}, errors.New(pgErr.Detail)
				}
			} else {
				return models.Tenant{}, errors.New(pgErr.Message)
			}
		} else {
			return models.Tenant{}, err
		}
	}

	newTenant := models.Tenant{
		Name: name,
	}

	//After creation a tenant we update the users table and create fk
	//(we can't do alter and create fk in tables before global tenant exists in tenants table)
	queryAlterUsersTable := `
	ALTER TABLE IF EXISTS users ADD CONSTRAINT fk_tenant_name_users FOREIGN KEY (tenant_name) REFERENCES tenants (name);
	ALTER TABLE IF EXISTS configurations ADD CONSTRAINT fk_tenant_name_configurations FOREIGN KEY(tenant_name) REFERENCES tenants(name);
	ALTER TABLE IF EXISTS schemas ADD CONSTRAINT fk_tenant_name_schemas FOREIGN KEY(tenant_name) REFERENCES tenants(name);
	ALTER TABLE IF EXISTS tags ADD CONSTRAINT fk_tenant_name_tags FOREIGN KEY(tenant_name) REFERENCES tenants(name);
	ALTER TABLE IF EXISTS consumers ADD CONSTRAINT fk_tenant_name_consumers FOREIGN KEY(tenant_name) REFERENCES tenants(name);
	ALTER TABLE IF EXISTS stations ADD CONSTRAINT fk_tenant_name_stations FOREIGN KEY(tenant_name) REFERENCES tenants(name);
	ALTER TABLE IF EXISTS schema_versions ADD CONSTRAINT fk_tenant_name_schemaverse FOREIGN KEY(tenant_name) REFERENCES tenants(name);
	ALTER TABLE IF EXISTS producers ADD CONSTRAINT fk_tenant_name_producers FOREIGN KEY(tenant_name) REFERENCES tenants(name);`
	// we remove alter from dls_msgs table because dls msgs with nats compatability
	// TODO: add this alter and solve dls messages nats compatability
	// ALTER TABLE IF EXISTS dls_messages ADD CONSTRAINT fk_tenant_name_dls_msgs FOREIGN KEY(tenant_name) REFERENCES tenants(name);

	_, err = conn.Conn().Exec(ctx, queryAlterUsersTable)
	if err != nil {
		var pgErr *pgconn.PgError
		errPg := errors.As(err, &pgErr)
		if errPg && !strings.Contains(pgErr.Message, "already exists") {
			return models.Tenant{}, err
		}
	}
	return newTenant, nil
}

func UpsertBatchOfTenants(tenants []models.TenantForUpsert) error {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	valueStrings := make([]string, 0, len(tenants))
	valueArgs := make([]interface{}, 0, len(tenants))
	for i, tenant := range tenants {
		valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d)", i*2+1, i*2+2))
		valueArgs = append(valueArgs, tenant.Name)
		valueArgs = append(valueArgs, tenant.InternalWSPass)
	}
	query := fmt.Sprintf("INSERT INTO tenants (name, internal_ws_pass) VALUES %s ON CONFLICT (name) DO NOTHING", strings.Join(valueStrings, ","))
	stmt, err := conn.Conn().Prepare(ctx, "upsert_tenants", query)
	if err != nil {
		return err
	}
	_, err = conn.Conn().Query(ctx, stmt.Name, valueArgs...)
	if err != nil {
		return err
	}

	return nil
}

func GetGlobalTenant() (bool, models.Tenant, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return false, models.Tenant{}, err
	}
	defer conn.Release()
	query := `SELECT * FROM tenants WHERE name = '$memphis' LIMIT 1`
	stmt, err := conn.Conn().Prepare(ctx, "get_global_tenant", query)
	if err != nil {
		return false, models.Tenant{}, err
	}
	rows, err := conn.Conn().Query(ctx, stmt.Name)
	if err != nil {
		return false, models.Tenant{}, err
	}
	defer rows.Close()
	tenants, err := pgx.CollectRows(rows, pgx.RowToStructByPos[models.Tenant])
	if err != nil {
		return false, models.Tenant{}, err
	}
	if len(tenants) == 0 {
		return false, models.Tenant{}, nil
	}
	return true, tenants[0], nil
}

func GetAllTenants() ([]models.Tenant, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return []models.Tenant{}, err
	}
	defer conn.Release()
	query := `SELECT * FROM tenants`
	stmt, err := conn.Conn().Prepare(ctx, "get_all_tenants", query)
	if err != nil {
		return []models.Tenant{}, err
	}
	rows, err := conn.Conn().Query(ctx, stmt.Name)
	if err != nil {
		return []models.Tenant{}, err
	}
	defer rows.Close()
	tenants, err := pgx.CollectRows(rows, pgx.RowToStructByPos[models.Tenant])
	if err != nil {
		return []models.Tenant{}, err
	}
	if len(tenants) == 0 {
		return []models.Tenant{}, nil
	}
	return tenants, nil
}

func GetAllTenantsWithoutGlobal() ([]models.Tenant, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return []models.Tenant{}, err
	}
	defer conn.Release()
	query := `SELECT * FROM tenants WHERE name != $1 AND name != $2`
	stmt, err := conn.Conn().Prepare(ctx, "get_all_tenants_without_global", query)
	if err != nil {
		return []models.Tenant{}, err
	}
	rows, err := conn.Conn().Query(ctx, stmt.Name, conf.MemphisGlobalAccountName, conf.GlobalAccount)
	if err != nil {
		return []models.Tenant{}, err
	}
	defer rows.Close()
	tenants, err := pgx.CollectRows(rows, pgx.RowToStructByPos[models.Tenant])
	if err != nil {
		return []models.Tenant{}, err
	}
	if len(tenants) == 0 {
		return []models.Tenant{}, nil
	}
	return tenants, nil
}

func IsTenantExists(tenantName string) (bool, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()

	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return false, err
	}
	defer conn.Release()

	query := `SELECT EXISTS(SELECT 1 FROM tenants WHERE name = $1 LIMIT 1)`
	stmt, err := conn.Conn().Prepare(ctx, "is_tenant_exists", query)
	if err != nil {
		return false, err
	}

	var exists bool
	err = conn.Conn().QueryRow(ctx, stmt.Name, tenantName).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

func GetTenantById(id int) (bool, models.Tenant, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return false, models.Tenant{}, err
	}
	defer conn.Release()
	query := `SELECT * FROM tenants AS c WHERE id = $1 LIMIT 1`
	stmt, err := conn.Conn().Prepare(ctx, "get_tenant_by_id", query)
	if err != nil {
		return false, models.Tenant{}, err
	}
	rows, err := conn.Conn().Query(ctx, stmt.Name, id)
	if err != nil {
		return false, models.Tenant{}, err
	}
	defer rows.Close()
	tenants, err := pgx.CollectRows(rows, pgx.RowToStructByPos[models.Tenant])
	if err != nil {
		return false, models.Tenant{}, err
	}
	if len(tenants) == 0 {
		return false, models.Tenant{}, nil
	}
	return true, tenants[0], nil
}

func GetTenantByName(name string) (bool, models.Tenant, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return false, models.Tenant{}, err
	}
	defer conn.Release()
	query := `SELECT * FROM tenants AS c WHERE name = $1 LIMIT 1`
	stmt, err := conn.Conn().Prepare(ctx, "get_tenant_by_name", query)
	if err != nil {
		return false, models.Tenant{}, err
	}

	if name != conf.GlobalAccount {
		name = strings.ToLower(name)
	}

	rows, err := conn.Conn().Query(ctx, stmt.Name, name)
	if err != nil {
		return false, models.Tenant{}, err
	}
	defer rows.Close()
	tenants, err := pgx.CollectRows(rows, pgx.RowToStructByPos[models.Tenant])
	if err != nil {
		return false, models.Tenant{}, err
	}
	if len(tenants) == 0 {
		return false, models.Tenant{}, nil
	}
	return true, tenants[0], nil
}

func CreateTenant(name, firebaseOrganizationId, encryptrdInternalWSPass, organizationName string) (models.Tenant, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()

	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return models.Tenant{}, err
	}
	defer conn.Release()

	query := `INSERT INTO tenants (id, name, firebase_organization_id, internal_ws_pass, organization_name) VALUES(nextval('tenants_seq'),$1, $2, $3, $4)`

	stmt, err := conn.Conn().Prepare(ctx, "create_new_tenant", query)
	if err != nil {
		return models.Tenant{}, err
	}

	var tenantId int
	rows, err := conn.Conn().Query(ctx, stmt.Name, name, firebaseOrganizationId, encryptrdInternalWSPass, organizationName)
	if err != nil {
		return models.Tenant{}, err
	}
	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&tenantId)
		if err != nil {
			return models.Tenant{}, err
		}
	}

	if err := rows.Err(); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Detail != "" {
				if strings.Contains(pgErr.Detail, "already exists") {
					return models.Tenant{}, errors.New("Tenant " + name + " already exists")
				} else {
					return models.Tenant{}, errors.New(pgErr.Detail)
				}
			} else {
				return models.Tenant{}, errors.New(pgErr.Message)
			}
		} else {
			return models.Tenant{}, err
		}

	}

	newTenant := models.Tenant{
		Name: name,
	}

	return newTenant, nil
}

func RemoveTenant(tenantName string) error {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()

	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	removeTenants := `DELETE FROM tenants WHERE name = $1`

	stmt, err := conn.Conn().Prepare(ctx, "remove_tenant", removeTenants)
	if err != nil {
		return err
	}

	if tenantName != conf.GlobalAccount {
		tenantName = strings.ToLower(tenantName)
	}

	_, err = conn.Conn().Exec(ctx, stmt.Name, tenantName)
	if err != nil {
		return err
	}

	return nil
}

func RemoveTagsResourcesByTenant(tenantName string) error {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	query := `DELETE FROM tags WHERE tenant_name = $1`
	stmt, err := conn.Conn().Prepare(ctx, "remove_tenant_resources_by_tenant", query)
	if err != nil {
		return err
	}

	_, err = conn.Conn().Exec(ctx, stmt.Name, tenantName)
	if err != nil {
		return err
	}
	return nil
}

func SetTenantSequence(sequence int) error {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	tenant_seq := fmt.Sprintf("CREATE SEQUENCE IF NOT EXISTS tenants_seq start %v INCREMENT 1;", sequence)

	_, err = conn.Conn().Exec(ctx, tenant_seq)
	if err != nil {
		return err
	}
	return nil
}

func GetAllUsersInDB() (bool, []models.User, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return false, nil, err
	}

	defer conn.Release()
	query := `SELECT * FROM users`
	stmt, err := conn.Conn().Prepare(ctx, "get_user", query)
	if err != nil {
		return false, nil, err
	}

	rows, err := conn.Conn().Query(ctx, stmt.Name)
	if err != nil {
		return false, nil, err
	}
	defer rows.Close()

	users, err := pgx.CollectRows(rows, pgx.RowToStructByPos[models.User])
	if err != nil {
		return false, nil, err
	}
	if len(users) == 0 {
		return false, nil, nil
	}
	return true, users, nil
}

func DeleteOldProducersAndConsumers(timeInterval time.Time) ([]models.LightCG, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return []models.LightCG{}, err
	}
	defer conn.Release()

	var queries []string
	queries = append(queries, "DELETE FROM producers WHERE is_active = false AND updated_at < $1")
	queries = append(queries, "WITH deleted AS (DELETE FROM consumers WHERE is_active = false AND updated_at < $1 RETURNING *) SELECT deleted.consumers_group, s.name as station_name, deleted.station_id , deleted.tenant_name, deleted.partitions FROM deleted INNER JOIN stations s ON deleted.station_id = s.id GROUP BY deleted.consumers_group, s.name, deleted.station_id, deleted.tenant_name, deleted.partitions")

	batch := &pgx.Batch{}
	for _, q := range queries {
		batch.Queue(q, timeInterval)
	}

	br := conn.SendBatch(ctx, batch)

	_, err = br.Exec()
	if err != nil {
		return []models.LightCG{}, err
	}

	rows, err := br.Query()
	if err != nil {
		return []models.LightCG{}, err
	}

	cgs, err := pgx.CollectRows(rows, pgx.RowToStructByPos[models.LightCG])
	if err != nil {
		return []models.LightCG{}, err
	}

	return cgs, err
}

func GetAllDeletedConsumersFromList(consumers []string) ([]string, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return []string{}, err
	}
	defer conn.Release()
	query := "SELECT consumers_group FROM consumers WHERE consumers_group = ANY($1) GROUP BY station_id, consumers_group"

	rows, err := conn.Query(ctx, query, consumers)
	if err != nil {
		return []string{}, err
	}
	defer rows.Close()
	var remainingCG []string
	for rows.Next() {
		var cgName string
		err := rows.Scan(&cgName)
		if err != nil {
			return []string{}, err
		}
		remainingCG = append(remainingCG, cgName)
	}

	return remainingCG, nil
}

func RemovePoisonedCg(stationId int, cgName string) error {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()

	connection, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return err
	}
	defer connection.Release()

	tx, err := connection.Conn().Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	query := `UPDATE dls_messages SET poisoned_cgs = ARRAY_REMOVE(poisoned_cgs, $1) WHERE station_id=$2 ;`

	stmt, err := tx.Prepare(ctx, "update_dls_message", query)
	if err != nil {
		return err
	}

	rows, err := tx.Query(ctx, stmt.Name, cgName, stationId)
	if err != nil {
		return err
	}
	rows.Close()

	query = `DELETE FROM dls_messages WHERE message_type = 'poison' AND poisoned_cgs = '{}' OR poisoned_cgs IS NULL;`
	stmt, err = tx.Prepare(ctx, "delete_dls_message", query)
	if err != nil {
		return err
	}

	rows, err = tx.Query(ctx, stmt.Name)
	if err != nil {
		return err
	}
	rows.Close()

	err = tx.Commit(ctx)
	if err != nil {
		return err
	}

	return nil
}

func DeleteConfByTenantName(tenantName string) error {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()

	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	query := `DELETE FROM configurations
	WHERE tenant_name=$1`

	stmt, err := conn.Conn().Prepare(ctx, "remove_conf_by_tenant_name", query)
	if err != nil {
		return err
	}
	tenantName = strings.ToLower(tenantName)
	_, err = conn.Conn().Exec(ctx, stmt.Name, tenantName)
	if err != nil {
		return err
	}
	return nil
}

func DeleteIntegrationsByTenantName(tenantName string) error {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()

	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	query := `DELETE FROM integrations
	WHERE tenant_name=$1`

	stmt, err := conn.Conn().Prepare(ctx, "remove_integrations_by_tenant_name", query)
	if err != nil {
		return err
	}
	tenantName = strings.ToLower(tenantName)
	_, err = conn.Conn().Exec(ctx, stmt.Name, tenantName)
	if err != nil {
		return err
	}
	return nil
}

// Async tasks functions
func UpsertAsyncTask(task, brokerInCharge string, createdAt time.Time, tenantName string, stationId int, username string) (models.AsyncTask, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()

	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return models.AsyncTask{}, err
	}
	defer conn.Release()

	query := `INSERT INTO async_tasks (name, broker_in_charge, created_at, updated_at, tenant_name, station_id, created_by) VALUES($1, $2, $3, $4, $5, $6, $7) RETURNING *`
	stmt, err := conn.Conn().Prepare(ctx, "upsert_async_task", query)
	if err != nil {
		return models.AsyncTask{}, err
	}

	updatedAt := createdAt
	var asyncTask models.AsyncTask
	err = conn.Conn().QueryRow(ctx, stmt.Name, task, brokerInCharge, createdAt, updatedAt, tenantName, stationId, username).Scan(
		&asyncTask.ID,
		&asyncTask.Name,
		&asyncTask.BrokrInCharge,
		&asyncTask.CreatedAt,
		&asyncTask.UpdatedAt,
		&asyncTask.Data,
		&asyncTask.TenantName,
		&asyncTask.StationId,
		&asyncTask.CreatedBy,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			ctx, cancelfunc = context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
			defer cancelfunc()

			conn, err = MetadataDbClient.Client.Acquire(ctx)
			if err != nil {
				return models.AsyncTask{}, err
			}
			defer conn.Release()
			// The task already exists, retrieve the existing task
			queryExistingTask := `
				SELECT *
				FROM async_tasks
				WHERE name = $1 AND tenant_name = $2 AND station_id = $3;
			`
			existingStmt, err := conn.Conn().Prepare(ctx, "select_existing_async_task", queryExistingTask)
			if err != nil {
				return models.AsyncTask{}, err
			}

			err = conn.Conn().QueryRow(ctx, existingStmt.Name, task, tenantName, stationId).Scan(
				&asyncTask.ID,
				&asyncTask.Name,
				&asyncTask.BrokrInCharge,
				&asyncTask.CreatedAt,
				&asyncTask.UpdatedAt,
				&asyncTask.Data,
				&asyncTask.TenantName,
				&asyncTask.StationId,
				&asyncTask.CreatedBy,
			)
			if err != nil {
				return models.AsyncTask{}, err
			}
		} else {
			return models.AsyncTask{}, err
		}
	}

	return asyncTask, nil
}

func GetAsyncTasksByName(task string) (bool, []models.AsyncTask, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()

	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return false, []models.AsyncTask{}, err
	}
	defer conn.Release()

	query := `SELECT * FROM async_tasks WHERE name = $1`
	stmt, err := conn.Conn().Prepare(ctx, "get_async_tasks_by_name", query)
	if err != nil {
		return false, []models.AsyncTask{}, err
	}

	rows, err := conn.Conn().Query(ctx, stmt.Name, task)
	if err != nil {
		return false, []models.AsyncTask{}, err
	}
	defer rows.Close()
	asyncTask, err := pgx.CollectRows(rows, pgx.RowToStructByPos[models.AsyncTask])
	if err != nil {
		return false, []models.AsyncTask{}, err
	}
	if len(asyncTask) == 0 {
		return false, []models.AsyncTask{}, nil
	}
	return true, asyncTask, nil
}

func GetAsyncTaskByNameAndBrokerName(task, brokerName string) (bool, []models.AsyncTask, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()

	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return false, []models.AsyncTask{}, err
	}
	defer conn.Release()

	query := `SELECT * FROM async_tasks WHERE name = $1 AND broker_in_charge = $2`
	stmt, err := conn.Conn().Prepare(ctx, "get_async_task_by_name_and_broker_name", query)
	if err != nil {
		return false, []models.AsyncTask{}, err
	}

	rows, err := conn.Conn().Query(ctx, stmt.Name, task, brokerName)
	if err != nil {
		return false, []models.AsyncTask{}, err
	}
	defer rows.Close()
	asyncTask, err := pgx.CollectRows(rows, pgx.RowToStructByPos[models.AsyncTask])
	if err != nil {
		return false, []models.AsyncTask{}, err
	}
	if len(asyncTask) == 0 {
		return false, []models.AsyncTask{}, nil
	}
	return true, asyncTask, nil
}

func GetAllAsyncTasks(tenantName string) ([]models.AsyncTaskRes, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()

	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return []models.AsyncTaskRes{}, err
	}
	defer conn.Release()
	query := `SELECT a.id, a.name, a.created_at, a.created_by, s.name, a.meta_data
	FROM async_tasks AS a
	LEFT JOIN stations AS s ON a.station_id = s.id
	WHERE a.tenant_name = $1
	ORDER BY a.created_at DESC;`
	stmt, err := conn.Conn().Prepare(ctx, "get_all_async_tasks", query)
	if err != nil {
		return []models.AsyncTaskRes{}, err
	}

	rows, err := conn.Conn().Query(ctx, stmt.Name, tenantName)
	if err != nil {
		return []models.AsyncTaskRes{}, err
	}
	defer rows.Close()

	var asyncTasks []models.AsyncTaskRes
	for rows.Next() {
		var task models.AsyncTaskRes
		var sName pgtype.Varchar

		err := rows.Scan(
			&task.ID,
			&task.Name,
			&task.CreatedAt,
			&task.CreatedBy,
			&sName,
			&task.Data,
		)
		if err != nil {
			return []models.AsyncTaskRes{}, err
		}

		if sName.Status == pgtype.Present {
			task.StationName = sName.String
		} else {
			task.StationName = "" // Handle NULL value
		}

		asyncTasks = append(asyncTasks, task)
	}

	if err := rows.Err(); err != nil {
		return []models.AsyncTaskRes{}, err
	}
	if len(asyncTasks) == 0 {
		return []models.AsyncTaskRes{}, nil
	}
	return asyncTasks, nil
}

func UpdateAsyncTask(task, tenantName string, updatedAt time.Time, metaData interface{}, stationId int) error {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()
	query := `UPDATE async_tasks SET updated_at = $1 ,meta_data = $2 WHERE name = $3 AND tenant_name=$4 AND station_id = $5`
	stmt, err := conn.Conn().Prepare(ctx, "edit_async_task_by_task_and_tenant_name_and_station_id", query)
	if err != nil {
		return err
	}
	if tenantName != conf.GlobalAccount {
		tenantName = strings.ToLower(tenantName)
	}
	_, err = conn.Conn().Query(ctx, stmt.Name, updatedAt, metaData, task, tenantName, stationId)
	if err != nil {
		return err
	}
	return nil
}

func RemoveAsyncTask(task, tenantName string, stationId int) error {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()
	query := `DELETE FROM async_tasks WHERE name = $1 AND tenant_name=$2 AND station_id = $3`
	stmt, err := conn.Conn().Prepare(ctx, "remove_async_task_by_name_and_tenant_name_and_station_id", query)
	if err != nil {
		return err
	}
	_, err = conn.Conn().Exec(ctx, stmt.Name, task, tenantName, stationId)
	if err != nil {
		return err
	}
	return nil
}

func RemoveAllAsyncTasks(duration time.Duration) ([]int, error) {
	sub := time.Now().Add(-duration)
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return []int{}, err
	}
	defer conn.Release()
	query := `WITH deleted_rows AS (
		DELETE FROM async_tasks
		WHERE updated_at <= $1 AND name='resend_all_dls_msgs'
		RETURNING station_id
	  )
	  SELECT ARRAY_AGG(station_id) AS deleted_station_ids
	  FROM deleted_rows;`
	stmt, err := conn.Conn().Prepare(ctx, "remove_all_async_tasks", query)
	if err != nil {
		return []int{}, err
	}

	var stationIds []int
	rows, err := conn.Conn().Query(ctx, stmt.Name, sub)
	if err != nil {
		return []int{}, err
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&stationIds)
		if err != nil {
			return []int{}, err
		}
	}
	return stationIds, nil
}

func CreateUserIfNotExist(username string, userType string, hashedPassword string, fullName string, subscription bool, avatarId int, tenantName string, pending bool, team, position, owner, description string) (models.User, error) {
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
		skip_get_started,
		tenant_name,
		pending,
		team, 
		position,
		owner,
		description) 
    VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15) 
	ON CONFLICT (username, tenant_name) DO NOTHING
	RETURNING id`

	stmt, err := conn.Conn().Prepare(ctx, "create_new_user_if_not_exist", query)
	if err != nil {
		return models.User{}, err
	}
	createdAt := time.Now()
	skipGetStarted := false
	alreadyLoggedIn := false

	var userId int
	if tenantName != conf.GlobalAccount {
		tenantName = strings.ToLower(tenantName)
	}
	rows, err := conn.Conn().Query(ctx, stmt.Name, username, hashedPassword, userType, alreadyLoggedIn, createdAt, avatarId, fullName, subscription, skipGetStarted, tenantName, pending, team, position, owner, description)
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
	if tenantName != conf.GlobalAccount {
		tenantName = strings.ToLower(tenantName)
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
		TenantName:      tenantName,
		Pending:         pending,
		Team:            team,
		Position:        position,
		Owner:           owner,
		Description:     description,
	}
	return newUser, nil
}

func CountProudcersForStation(stationId int) (int64, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return 0, err
	}
	defer conn.Release()
	query := `SELECT COUNT(*) FROM producers WHERE station_id=$1`
	stmt, err := conn.Conn().Prepare(ctx, "get_count_producers_for_station", query)
	if err != nil {
		return 0, err
	}
	var count int64
	err = conn.Conn().QueryRow(ctx, stmt.Name, stationId).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

// Shared Locks Functions
func GetAndLockSharedLock(name string, tenantName string) (bool, bool, models.SharedLock, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return false, false, models.SharedLock{}, err
	}
	defer conn.Release()
	tx, err := conn.Begin(ctx)
	if err != nil {
		return false, false, models.SharedLock{}, err
	}
	defer tx.Rollback(ctx)

	if tenantName != conf.GlobalAccount {
		tenantName = strings.ToLower(tenantName)
	}

	selectQuery := `SELECT * FROM shared_locks WHERE name = $1 AND tenant_name = $2 FOR UPDATE LIMIT 1`
	stmt, err := tx.Prepare(ctx, "get_and_lock_shared_lock", selectQuery)
	if err != nil {
		return false, false, models.SharedLock{}, err
	}
	rows, err := tx.Query(ctx, stmt.Name, name, tenantName)
	if err != nil {
		if err == pgx.ErrNoRows {
			return false, false, models.SharedLock{}, nil
		} else {
			return false, false, models.SharedLock{}, err
		}
	}
	defer rows.Close()
	sharedLocks, err := pgx.CollectRows(rows, pgx.RowToStructByPos[models.SharedLock])
	if err != nil && err != pgx.ErrNoRows {
		return false, false, models.SharedLock{}, err
	}
	var sharedLock models.SharedLock
	lockTime := time.Now()
	newLock := false
	if len(sharedLocks) == 0 || err == pgx.ErrNoRows {
		insterQuery := `INSERT INTO shared_locks(name, tenant_name, lock_held, locked_at) VALUES($1, $2, $3, $4) RETURNING *`

		stmt, err := conn.Conn().Prepare(ctx, "insert_new_shared_lock_and_lock", insterQuery)
		if err != nil {
			return false, false, models.SharedLock{}, err
		}

		newSharedLock := models.SharedLock{}
		if tenantName != conf.GlobalAccount {
			tenantName = strings.ToLower(tenantName)
		}
		rows, err := conn.Conn().Query(ctx, stmt.Name, name, tenantName, true, lockTime)
		if err != nil {
			return false, false, models.SharedLock{}, err
		}
		defer rows.Close()

		for rows.Next() {
			err := rows.Scan(&newSharedLock.ID, &newSharedLock.Name, &newSharedLock.TenantName, &newSharedLock.LockHeld, &newSharedLock.LockedAt)
			if err != nil {
				return false, false, models.SharedLock{}, err
			}
		}
		sharedLock = newSharedLock
		newLock = true
	} else {
		sharedLock = sharedLocks[0]
	}

	if !sharedLock.LockHeld {
		updateQuery := "UPDATE shared_locks SET lock_held = TRUE, locked_at = $1 WHERE id = $2"
		lockStmt, err := tx.Prepare(ctx, "lock_shared_lock", updateQuery)
		if err != nil {
			return false, false, models.SharedLock{}, err
		}
		_, err = tx.Exec(ctx, lockStmt.Name, lockTime, sharedLock.ID)
		if err != nil {
			return false, false, models.SharedLock{}, err
		}
		sharedLock.LockHeld = true
		sharedLock.LockedAt = lockTime
	} else if !newLock {
		return true, false, sharedLock, nil
	}

	err = tx.Commit(ctx)
	if err != nil {
		return false, false, models.SharedLock{}, err
	}
	return true, true, sharedLock, nil
}

func SharedLockUnlock(name, tenantName string) error {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	query := `UPDATE shared_locks SET lock_held = FALSE WHERE name = $1 AND tenant_name=$2`
	stmt, err := conn.Conn().Prepare(ctx, "unlock_shared_lock", query)
	if err != nil {
		return err
	}
	tenantName = strings.ToLower(tenantName)
	_, err = conn.Conn().Query(ctx, stmt.Name, name, tenantName)
	if err != nil {
		return err
	}
	return nil
}

func DeleteAllSharedLocks(tenantName string) error {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()

	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()
	query := `DELETE FROM shared_locks WHERE tenant_name=$1`
	stmt, err := conn.Conn().Prepare(ctx, "delete_all_shared_locks", query)
	if err != nil {
		return err
	}
	_, err = conn.Conn().Query(ctx, stmt.Name, tenantName)
	if err != nil {
		return err
	}
	return nil
}

func releaseStuckedSharedLocks(lockedAt time.Time) error {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()
	query := `UPDATE shared_locks SET lock_held = false WHERE locked_at < $1 AND lock_held = true`
	stmt, err := conn.Conn().Prepare(ctx, "set_lock_held_by_name_and_by_locked_at_shared_lock", query)
	if err != nil {
		return err
	}
	_, err = conn.Conn().Query(ctx, stmt.Name, lockedAt)
	if err != nil {
		return err
	}
	return nil
}

func releaseStuckedStationLocks(lockedAt time.Time) error {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()
	query := `UPDATE stations SET functions_lock_held = false WHERE functions_locked_at < $1 AND functions_lock_held = true`
	stmt, err := conn.Conn().Prepare(ctx, "set_functions_lock_held_at_station", query)
	if err != nil {
		return err
	}
	_, err = conn.Conn().Query(ctx, stmt.Name, lockedAt)
	if err != nil {
		return err
	}
	return nil
}

func UnlockStuckLocks(lockedAt time.Time) error {
	err := releaseStuckedStationLocks(lockedAt)
	if err != nil {
		return fmt.Errorf("releaseStuckedStationLocks: %v", err)
	}
	err = releaseStuckedSharedLocks(lockedAt)
	if err != nil {
		return fmt.Errorf("releaseStuckedSharedLocks: %v", err)
	}
	return nil
}

func GetMemphisFunctionsByMemphis() ([]models.Function, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return []models.Function{}, err
	}
	defer conn.Release()
	query := `SELECT * FROM functions WHERE by_memphis = true`
	stmt, err := conn.Conn().Prepare(ctx, "get_memphis_functions_by_memphis", query)
	if err != nil {
		return []models.Function{}, err
	}
	rows, err := conn.Conn().Query(ctx, stmt.Name)
	if err != nil {
		return []models.Function{}, err
	}

	defer rows.Close()
	functions, err := pgx.CollectRows(rows, pgx.RowToStructByPos[models.Function])
	if err != nil {
		return []models.Function{}, err
	}
	if len(functions) == 0 {
		return []models.Function{}, err
	}
	return functions, nil
}
