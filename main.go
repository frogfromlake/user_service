package main

import (
	"context"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"

	db "github.com/Streamfair/streamfair_user_svc/db/sqlc"
	"github.com/Streamfair/streamfair_user_svc/gapi"
	"github.com/Streamfair/streamfair_user_svc/util"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	config, err := util.LoadConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("config: error while loading config:")
	}

	if viper.GetString("ENVIRONMENT") == "development" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	log.Info().Msg("Hello, Streamfair User Service!")

	poolConfig, err := pgxpool.ParseConfig(config.DBSource)
	if err != nil {
		log.Fatal().Err(err).Msg("config: error while parsing config:")
	}

	conn, err := pgxpool.New(context.Background(), poolConfig.ConnString())
	if err != nil {
		log.Fatal().Err(err).Msg("db connection: unable to create connection pool:")
	}

	store := db.NewStore(conn)
	server, err := gapi.NewServer(config, store)
	if err != nil {
		log.Fatal().Err(err).Msg("server: error while creating server:")
	}

	runDBMigration(config.MigrationURL, config.DBSource)

	go server.RunGrpcGatewayServer()
	server.RunGrpcServer()
}

func runDBMigration(migrationURL string, dbSource string) {
	migration, err := migrate.New(migrationURL, dbSource)
	if err != nil {
		log.Fatal().Err(err).Msg("db migration: unable to create migration:")
	}

	if err = migration.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatal().Err(err).Msg("db migration: unable to apply migration:")
	}

	log.Info().Msg("DB migrated successfully")
}
