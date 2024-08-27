package database

import (
  "context"
  "fmt"
  "os"

  "github.com/mpilhlt/dhamps-vdb/internal/models"

  "github.com/jackc/pgx/v5"

  // "github.com/pgvector/pgvector-go"

  pgxvector "github.com/pgvector/pgvector-go/pgx"
)

// Database initialization

func InitDB(options *models.Options) {
  // urlExample := "postgres://username:password@localhost:5432/database_name"
  url := fmt.Sprintf("postgres://%s:%s@%s:%d/%s",
    options.DBUser, options.DBPassword, options.DBHost, options.DBPort, options.DBName)

  // fmt.Printf("Connecting to postgres database: %s\n", os.Getenv("SERVICE_DB_HOST"))
  ctx_bg := context.Background()
  conn, err := pgx.Connect(ctx_bg, url)
  if err != nil {
    fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
    os.Exit(1)
  }
  defer conn.Close(ctx_bg)
  fmt.Printf("Connected to postgres database: %s@%s:%d/%s\n", // alternatively, print conn.ConnInfo().DatabaseName
    options.DBUser, options.DBHost, options.DBPort, options.DBName)

  // Enable extension
  _, err = conn.Exec(context.TODO(), "CREATE EXTENSION IF NOT EXISTS vector")
  if err != nil {
    fmt.Fprintf(os.Stderr, "Unable to enable vector extension: %v\n", err)
    os.Exit(1)
  }
  // Register Types
  err = pgxvector.RegisterTypes(context.TODO(), conn)
  if err != nil {
    fmt.Fprintf(os.Stderr, "Unable to register types: %v\n", err)
    os.Exit(1)
  }

  // Make sure tables and indices are created
  err = SetubDB(conn)
  if err != nil {
    fmt.Fprintf(os.Stderr, "Unable to setup database: %v\n", err)
    os.Exit(1)
  }
  fmt.Printf("Database setup completed\n")
}

func SetubDB(conn *pgx.Conn) (error) {

  // Create tables
  _, err := conn.Exec(context.TODO(), "CREATE TABLE IF NOT EXISTS test (id bigserial PRIMARY KEY, embedding vector(3))")
  if err != nil {
    return err
  }

  // Create indices
  _, err = conn.Exec(context.TODO(), "CREATE INDEX IF NOT EXISTS idx_test_embedding ON test USING  hnsw (embedding vector_cosine_ops)")
  if err != nil {
    return err
  }

// Add an approximate index
//  _, err := conn.Exec(ctx, "CREATE INDEX ON items USING hnsw (embedding vector_l2_ops)")
//  or
//  _, err := conn.Exec(ctx, "CREATE INDEX ON items USING ivfflat (embedding vector_l2_ops) WITH (lists = 100)")
//
// Use vector_ip_ops for inner product and vector_cosine_ops for cosine distance

  return nil
}

// Vector functions

func RetrieveSimilarsByID(id string) {
}

func RetrieveSimilarsByVector(v models.Vector) {
// Get the nearest neighbors to a vector
// rows, err := conn.Query(ctx, "SELECT id FROM items ORDER BY embedding <-> $1 LIMIT 5", pgvector.NewVector([]float32{1, 2, 3}))
}

// Embeddings handling

func UploadEmbeddings(e models.Embeddings) {
// Insert a vector:
// _, err := conn.Exec(ctx, "INSERT INTO items (embedding) VALUES ($1)", pgvector.NewVector([]float32{1, 2, 3}))

}

func RetrieveEmbeddings(id string) {
}

// User handling

func UploadUser(u models.User) {
}

func RetrieveUser(id string) {
}

func UpdateUser(id string, u models.User) {
}

func DeleteUser(id string) {
}

// Project handling

func UploadProject(p models.Project) {
}

func RetrieveProject(id string) {
}

func UpdateProject(id string, p models.Project) {
}

func DeleteProject(id string) {
}
