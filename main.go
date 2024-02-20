package main

import (
	"context"
	"fmt"
	"log"

	db "github.com/Streamfair/streamfair_user_svc/db/sqlc"
	"github.com/Streamfair/streamfair_user_svc/gapi"
	"github.com/Streamfair/streamfair_user_svc/util"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	fmt.Println("Hello, Streamfair User Management Service!")

	config, err := util.LoadConfig(".")
	if err != nil {
		log.Printf("config: error while loading config: %v\n", err)
	}

	poolConfig, err := pgxpool.ParseConfig(config.DBSource)
	if err != nil {
		log.Printf("config: error while parsing config: %v\n", err)
	}

	conn, err := pgxpool.New(context.Background(), poolConfig.ConnString())
	if err != nil {
		log.Printf("db connection: unable to create connection pool: %v\n", err)
	}

	store := db.NewStore(conn)
	server, err := gapi.NewServer(config, store)
	if err != nil {
		log.Printf("server: error while creating server: %v\n", err)
	}

	runDBMigration(config.MigrationURL, config.DBSource)

	go server.RunGrpcGatewayServer()
	server.RunGrpcServer()
}

func runDBMigration(migrationURL string, dbSource string) {
	migration, err := migrate.New(migrationURL, dbSource)
	if err != nil {
		log.Fatalf("db migration: unable to create migration: %v\n", err)
	}

	if err = migration.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("db migration: unable to apply migration: %v\n", err)
	}

	log.Println("db migrated successfully")
}
