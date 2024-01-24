package db

import (
	"context"
	"errors"
	"time"
)

// CustomPing checks the database connection by executing a simple query with a timeout.
func (store *SQLStore) Ping(ctx context.Context, timeout time.Duration) error {
	// Create a context with a timeout
	pingCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Execute a simple query to check the connection
	err := store.ExecTx(pingCtx, func(queries *Queries) error {
		// Use a simple SELECT 1 query to check the connection
		row := queries.db.QueryRow(pingCtx, "SELECT 1")
		var result int
		if err := row.Scan(&result); err != nil {
			return err
		}

		// Ensure that the result is as expected (1 in this case)
		if result != 1 {
			return errors.New("unexpected result from SELECT 1")
		}

		return nil
	})

	return err
}
