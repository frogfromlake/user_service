package db

import (
	"context"
	"fmt"
	"os"

	"github.com/frogfromlake/user_service/util"
	"github.com/jackc/pgx/v5/pgxpool"

	"testing"
)

var testQueries *Queries
var testDB *pgxpool.Pool

func setupDBConnection() {
	config, err := util.LoadConfig("../..")
	if err != nil {
		fmt.Fprintf(os.Stderr, "config: error while loading config: %v\n", err)
	}

	poolConfig, err := pgxpool.ParseConfig(config.DBSource)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error while parsing config: %v\n", err)
	}

	testDB, err = pgxpool.New(context.Background(), poolConfig.ConnString())
	if err != nil {
		fmt.Fprintf(os.Stderr, "db connection: unable to create connection pool: %v\n", err)
	}

	testQueries = New(testDB)
}

func TestMain(m *testing.M) {
	setupDBConnection()

	defer testDB.Close()

	os.Exit(m.Run())
}
