package database

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/mpilhlt/dhamps-vdb/internal/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Database initialization

func InitDB(options *models.Options) (*pgxpool.Pool, error) {
	println("--- Connecting to database ...")

	// urlExample := "postgres://username:password@localhost:5432/database_name"
	url := fmt.Sprintf("postgres://%s:%s@%s:%d/%s",
		options.DBUser, options.DBPassword, options.DBHost, options.DBPort, options.DBName)

	// Connect to the database, first without concurrency to check the schema version
	ctx_cancel, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err := VerifySchema(ctx_cancel, url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "EEE Unable to verify schema: %v\n", err)
		return nil, err
	}

	// For the actual application, connect to the db using a connection pool
	ctx_bg := context.Background()
	pool, err := pgxpool.New(ctx_bg, url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "EEE Unable to get connection pool from database: %v\n", err)
		return nil, err
	}
	defer pool.Close()
	fmt.Printf("    Successfully got connection pool from postgres database: %s@%s:%d/%s.\n", // alternatively, print conn.ConnInfo().DatabaseName
		options.DBUser, options.DBHost, options.DBPort, options.DBName)

	// Run test query (get all users)
	conn, err := pool.Acquire(ctx_bg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "EEE Unable to get connection from pool: %v\n", err)
		return nil, err
	}
	err = testQuery(ctx_bg, conn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "EEE Unable to run test query: %v\n", err)
		return nil, err
	}

	// Done, everything has been set up. Return connection pool.
	println("--- Database up and initialized.")
	return pool, nil
}

func VerifySchema(ctx context.Context, url string) error {
	conn, err := pgx.Connect(ctx, url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "EEE Unable to connect to database: %v\n", err)
		return err
	}
	// Check schema version of database
	println("--- Checking schema version of database ...")
	migrator, err := NewMigrator(ctx, conn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "EEE Unable to initialize migrator: %v\n", err)
		return err
	}
	// get the current migration status
	now, exp, info, err := migrator.Info()
	if err != nil {
		fmt.Fprintf(os.Stderr, "EEE Unable to get schema info: %v\n", err)
		return err
	}
	if now < exp {
		// migration is required, dump out the current state
		// and perform the migration
		println("    Database scheme needs migration, current state: ")
		println(info)

		ctx_cancel, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
		err = migrator.Migrate(ctx_cancel)
		if err != nil {
			fmt.Fprintf(os.Stderr, "EEE Unable to migrate schema: %v\n", err)
			return err
		}
		println("    Database migration successful, schema up to date!")
	} else {
		println("    Database schema up to date, no database migration needed")
	}

	conn.Close(ctx)
	return nil
}

func testQuery(ctx context.Context, conn *pgxpool.Conn) error {
	println("--- Send test query - list all users:")

	ctx_cancel, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	queries := New(conn)
	users, err := queries.GetUsers(ctx_cancel, GetUsersParams{Limit: 10, Offset: 0})
	if err != nil {
		fmt.Fprintf(os.Stderr, "EEE Unable to get users: %v\n", err)
		return err
	}
	println(users)
	return nil
}
