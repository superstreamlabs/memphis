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
	"errors"
	"io/ioutil"

	"memphis/conf"
	"strings"

	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
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
	Client *pgxpool.Pool
	Ctx    context.Context
	Cancel context.CancelFunc
	Conn   *pgxpool.Conn
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
		db.Client.Close()
	}()
}

func JoinTable(dbPostgre DbPostgreSQLInstance) error {
	ctx, cancelfunc := context.WithTimeout(context.Background(), dbOperationTimeout*time.Second)
	defer cancelfunc()
	query := `SELECT users
    FROM users AS q
    JOIN tags AS a ON q.id = a.users
    WHERE q.id = $1`

	conn := dbPostgre.Conn

	defer conn.Release()
	_, err := conn.Conn().Prepare(ctx, "join", query)
	if err != nil {
		return err
	}
	_, err = conn.Conn().Exec(ctx, "join", 1)
	if err != nil {
		return err
	}

	return nil
}

func InsertToTable(dbPostgre DbPostgreSQLInstance) error {
	ctx, cancelfunc := context.WithTimeout(context.Background(), dbOperationTimeout*time.Second)
	defer cancelfunc()
	conn := dbPostgre.Conn
	defer conn.Release()
	_, err := conn.Conn().Prepare(ctx, "insert into", `INSERT INTO tags (name, color, users, stations, schemas) 
    VALUES($1, $2, $3, $4, $5);`)
	if err != nil {
		return err
	}
	_, err = conn.Conn().Exec(ctx, "insert into", "sname", "red", "{1}", "{}", "{}")
	if err != nil {
		return err
	}

	return nil
}

func SelectFromTable(dbPostgre DbPostgreSQLInstance) error {
	ctx, cancelfunc := context.WithTimeout(context.Background(), dbOperationTimeout*time.Second)
	defer cancelfunc()

	conn := dbPostgre.Conn
	defer conn.Release()
	_, err := conn.Conn().Prepare(ctx, "select from", `SELECT username FROM users WHERE username = $1`)
	if err != nil {
		return err
	}
	var username string
	rows := conn.Conn().QueryRow(ctx, "select from", "test")

	err = rows.Scan(&username)
	if err != nil {
		return err
	}

	return nil

}

func updateFieldInTable(dbPostgre DbPostgreSQLInstance) error {
	ctx, cancelfunc := context.WithTimeout(context.Background(), dbOperationTimeout*time.Second)
	defer cancelfunc()

	conn := dbPostgre.Conn
	defer conn.Release()
	_, err := conn.Conn().Prepare(ctx, "update", `UPDATE users
	SET username = $2
	WHERE id = $1
	RETURNING id, username;`)
	if err != nil {
		return err
	}
	var username string
	var id int
	rows := conn.Conn().QueryRow(ctx, "update", 7, "test")
	err = rows.Scan(&id, &username)
	if err != nil {
		return err
	}

	return nil
}

func dropRowInTable(dbPostgre DbPostgreSQLInstance) error {
	ctx, cancelfunc := context.WithTimeout(context.Background(), dbOperationTimeout*time.Second)
	defer cancelfunc()

	conn := dbPostgre.Conn
	defer conn.Release()
	_, err := conn.Conn().Prepare(ctx, "drop", `DELETE FROM users WHERE id = $1;`)
	if err != nil {
		return err
	}
	_, err = conn.Conn().Exec(ctx, "drop", 7)
	if err != nil {
		return err
	}
	return nil
}

func AddInexToTable(indexName, tableName, field string, dbPostgreSQL DbPostgreSQLInstance) error {
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

func createTablesInDb(dbPostgreSQL DbPostgreSQLInstance) error {
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
		id SERIAL NOT NULL,
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
		name VARCHAR NOT NULL,
		keys JSON,
		properties JSON,
		PRIMARY KEY (id)
		);`

	schemasTable := `
	CREATE TYPE enum_type AS ENUM ('json', 'graphql', 'protobuf');
	CREATE TABLE IF NOT EXISTS schemas(
		id SERIAL NOT NULL,
		name VARCHAR NOT NULL,
		type enum_type NOT NULL DEFAULT 'protobuf',
		PRIMARY KEY (id)
		);
		CREATE INDEX name
		ON schemas (name);`

	tagsTable := `CREATE TABLE IF NOT EXISTS tags(
		id SERIAL NOT NULL,
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
		id SERIAL NOT NULL,
		name VARCHAR NOT NULL,
		station_id INTEGER NOT NULL,
		connection_id INTEGER NOT NULL,
		consumers_group VARCHAR NOT NULL,
		max_ack_time_ms SERIAL NOT NULL,
		created_by INTEGER NOT NULL,
		is_active BOOL NOT NULL DEFAULT false,
		is_deleted BOOL NOT NULL DEFAULT false,
		created_at TIMESTAMP NOT NULL,
		max_msg_deliveries SERIAL NOT NULL,
		start_consume_from_seq SERIAL NOT NULL,
		last_msgs SERIAL NOT NULL,
		type enum_type_consumer NOT NULL DEFAULT 'application',
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
		ON consumers (station_id);`

	stationsTable := `
	CREATE TYPE enum_retention_type AS ENUM ('message_age_sec', 'messages', 'bytes');
	CREATE TYPE enum_storage_type AS ENUM ('file', 'memory');
	CREATE TABLE IF NOT EXISTS stations(
		id SERIAL NOT NULL,
		name VARCHAR NOT NULL,
		retention_type enum_retention_type NOT NULL DEFAULT 'message_age_sec',
		storage_type enum_storage_type NOT NULL DEFAULT 'file',
		replicas SERIAL NOT NULL,
		created_by INTEGER NOT NULL,
		created_at TIMESTAMP NOT NULL,
		updated_at TIMESTAMP NOT NULL,
		is_deleted BOOL NOT NULL,
		idempotency_window_ms SERIAL NOT NULL,
		is_native BOOL NOT NULL ,
		tiered_storage_enabled BOOL NOT NULL,
		dls_config JSON NOT NULL,
		schema_name VARCHAR,
		schema_version_number SERIAL,
		PRIMARY KEY (id),
		CONSTRAINT fk_created_by
			FOREIGN KEY(created_by)
			REFERENCES users(id)
		);
		CREATE UNIQUE INDEX unique_station_name_deleted ON stations (name, is_deleted) WHERE is_deleted = false;`

	schemaVersionsTable := `CREATE TABLE IF NOT EXISTS schema_versions(
		id SERIAL NOT NULL,
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
		id SERIAL NOT NULL,
		name VARCHAR NOT NULL,
		station_id INTEGER NOT NULL,
		connection_id INTEGER NOT NULL,	
		created_by INTEGER NOT NULL,
		is_active BOOL NOT NULL DEFAULT false,
		is_deleted BOOL NOT NULL DEFAULT false,
		created_at TIMESTAMP NOT NULL,
		type enum_producer_type NOT NULL DEFAULT 'application',
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

	_, err := db.Exec(ctx, usersTable)
	if err != nil {
		var pgErr *pgconn.PgError
		errPg := errors.As(err, &pgErr)
		if errPg && !strings.Contains(pgErr.Message, "already exists") {
			cancelfunc()
			return err
		}
	}

	_, err = db.Exec(ctx, connectionsTable)
	if err != nil {
		var pgErr *pgconn.PgError
		errPg := errors.As(err, &pgErr)
		if errPg && !strings.Contains(pgErr.Message, "already exists") {
			cancelfunc()
			return err
		}
	}

	_, err = db.Exec(ctx, auditLogsTable)
	if err != nil {
		var pgErr *pgconn.PgError
		errPg := errors.As(err, &pgErr)
		if errPg && !strings.Contains(pgErr.Message, "already exists") {
			cancelfunc()
			return err
		}
	}

	_, err = db.Exec(ctx, configurationsTable)
	if err != nil {
		var pgErr *pgconn.PgError
		errPg := errors.As(err, &pgErr)
		if errPg && !strings.Contains(pgErr.Message, "already exists") {
			cancelfunc()
			return err
		}
	}

	_, err = db.Exec(ctx, integrationsTable)
	if err != nil {
		var pgErr *pgconn.PgError
		errPg := errors.As(err, &pgErr)
		if errPg && !strings.Contains(pgErr.Message, "already exists") {
			cancelfunc()
			return err
		}
	}

	_, err = db.Exec(ctx, schemasTable)
	if err != nil {
		var pgErr *pgconn.PgError
		errPg := errors.As(err, &pgErr)
		if errPg && !strings.Contains(pgErr.Message, "already exists") {
			cancelfunc()
			return err
		}
	}

	_, err = db.Exec(ctx, tagsTable)
	if err != nil {
		var pgErr *pgconn.PgError
		errPg := errors.As(err, &pgErr)
		if errPg && !strings.Contains(pgErr.Message, "already exists") {
			cancelfunc()
			return err
		}
	}

	_, err = db.Exec(ctx, stationsTable)
	if err != nil {
		var pgErr *pgconn.PgError
		errPg := errors.As(err, &pgErr)
		if errPg && !strings.Contains(pgErr.Message, "already exists") {
			cancelfunc()
			return err
		}
	}
	_, err = db.Exec(ctx, consumersTable)
	if err != nil {
		var pgErr *pgconn.PgError
		errPg := errors.As(err, &pgErr)
		if errPg && !strings.Contains(pgErr.Message, "already exists") {
			cancelfunc()
			return err
		}
	}

	_, err = db.Exec(ctx, schemaVersionsTable)
	if err != nil {
		var pgErr *pgconn.PgError
		errPg := errors.As(err, &pgErr)
		if errPg && !strings.Contains(pgErr.Message, "already exists") {
			cancelfunc()
			return err
		}
	}
	_, err = db.Exec(ctx, producersTable)
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

	postgreSqlUser := configuration.POSTGRESQL_USER
	postgreSqlPassword := configuration.POSTGRESQL_PASS
	postgreSqlDbName := configuration.POSTGRESQL_DBNAME
	postgreSqlServiceName := configuration.POSTGRESQL_SERVICE
	postgreSqlPort := configuration.POSTGRESQL_PORT
	var postgreSqlUrl string
	if configuration.POSTGRESQL_TLS_ENABLED {
		postgreSqlUrl = "postgres://" + postgreSqlUser + "@" + postgreSqlServiceName + ":" + postgreSqlPort + "/" + postgreSqlDbName + "?sslmode=verify-full"
	} else {
		postgreSqlUrl = "postgres://" + postgreSqlUser + ":" + postgreSqlPassword + "@" + postgreSqlServiceName + ":" + postgreSqlPort + "/" + postgreSqlDbName + "?sslmode=disable"
	}

	config, err := pgxpool.ParseConfig(postgreSqlUrl)
	config.MaxConns = 5

	cert, err := tls.LoadX509KeyPair(configuration.POSTGRESQL_TLS_CRT, configuration.POSTGRESQL_TLS_KEY)
	if err != nil {
		cancelfunc()
		return DbPostgreSQLInstance{}, err
	}

	CACert, err := ioutil.ReadFile(configuration.POSTGRESQL_TLS_CA)
	if err != nil {
		cancelfunc()
		return DbPostgreSQLInstance{}, err
	}

	CACertPool := x509.NewCertPool()
	CACertPool.AppendCertsFromPEM(CACert)

	if configuration.POSTGRESQL_TLS_ENABLED {
		config.ConnConfig.TLSConfig = &tls.Config{Certificates: []tls.Certificate{cert}, RootCAs: CACertPool, InsecureSkipVerify: true}

	}

	dbPostgreSQL, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		cancelfunc()
		return DbPostgreSQLInstance{}, err
	}

	err = dbPostgreSQL.Ping(ctx)
	if err != nil {
		cancelfunc()
		return DbPostgreSQLInstance{}, err
	}
	l.Noticef("Established connection with the PostgreSQL DB")
	conn, err := dbPostgreSQL.Acquire(ctx)
	if err != nil {
		cancelfunc()
		return DbPostgreSQLInstance{}, err
	}
	dbPostgre := DbPostgreSQLInstance{Ctx: ctx, Cancel: cancelfunc, Client: dbPostgreSQL, Conn: conn}
	err = createTablesInDb(dbPostgre)
	if err != nil {
		cancelfunc()
		return DbPostgreSQLInstance{}, err
	}

	// err = AddInexToTable("username_index_2", "users", "username", dbPostgre)
	// if err != nil {
	// 	return DbPostgreSQLInstance{}, err
	// }

	// err = InsertToTable(dbPostgre)
	// if err != nil {
	// 	return DbPostgreSQLInstance{}, err
	// }

	// err = SelectFromTable(dbPostgre)
	// if err != nil {
	// 	return DbPostgreSQLInstance{}, err
	// }

	// err = updateFieldInTable(dbPostgre)
	// if err != nil {
	// 	return DbPostgreSQLInstance{}, err
	// }

	// err = dropRowInTable(dbPostgre)
	// if err != nil {
	// 	return DbPostgreSQLInstance{}, err
	// }

	// err = JoinTable(dbPostgre)
	// if err != nil {
	// 	return DbPostgreSQLInstance{}, err
	// }

	return DbPostgreSQLInstance{Client: dbPostgreSQL, Ctx: ctx, Cancel: cancelfunc}, nil

}
