package database

import (
  "context"
  "fmt"
  "os"

  "github.com/mpilhlt/dhamps-vdb/internal/models"

  "github.com/jackc/pgx/v5"
)

// Database initialization

func InitDB(options *models.Options) {
  println("--- Connecting to database ...")

  // urlExample := "postgres://username:password@localhost:5432/database_name"
  url := fmt.Sprintf("postgres://%s:%s@%s:%d/%s",
    options.DBUser, options.DBPassword, options.DBHost, options.DBPort, options.DBName)


  ctx_bg := context.Background()
  conn, err := pgx.Connect(ctx_bg, url)
  if err != nil {
    fmt.Fprintf(os.Stderr, "EEE Unable to connect to database: %v\n", err)
    os.Exit(1)
  }
  defer conn.Close(ctx_bg)
  fmt.Printf("    Successfully connected to postgres database: %s@%s:%d/%s.\n", // alternatively, print conn.ConnInfo().DatabaseName
    options.DBUser, options.DBHost, options.DBPort, options.DBName)


  // Check schema version of database
  println("--- Checking schema version of database ...")
  migrator, err := NewMigrator(ctx_bg, conn)
  if err != nil {
    fmt.Fprintf(os.Stderr, "EEE Unable to initialize migrator: %v\n", err)
    os.Exit(1)
  }
  // get the current migration status
  now, exp, info, err := migrator.Info()
  if err != nil {
    fmt.Fprintf(os.Stderr, "EEE Unable to get schema info: %v\n", err)
    os.Exit(1)
  }
  if now < exp {
    // migration is required, dump out the current state
    // and perform the migration
    println("    Database scheme needs migration, current state: ")
    println(info)

    err = migrator.Migrate(ctx_bg)
    if err != nil {
        fmt.Fprintf(os.Stderr, "EEE Unable to migrate schema: %v\n", err)
        os.Exit(1)
      }
    println("    Database migration successful, schema up to date!")
  } else {
    println("    Database schema up to date, no database migration needed")
  }


  // Test queries: get all users
  println("--- Send test query - list all users:")
  queries := New(conn)
  users, err := queries.GetUsers(ctx_bg)
  if err != nil {
    fmt.Fprintf(os.Stderr, "EEE Unable to get users: %v\n", err)
    os.Exit(1)
  }
  println(users)


  println("--- Database up and initialized.")
}
