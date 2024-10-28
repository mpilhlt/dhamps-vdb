package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/mpilhlt/dhamps-vdb/internal/database"
	"github.com/mpilhlt/dhamps-vdb/internal/handlers"
	"github.com/mpilhlt/dhamps-vdb/internal/models"

	"github.com/danielgtaylor/huma/v2/adapters/humago"
	"github.com/danielgtaylor/huma/v2/humacli"

	huma "github.com/danielgtaylor/huma/v2"
)

// TODO: Set up timeouts (e.g. in server definition)!

func main() {
  // Create a CLI app
  cli := humacli.New(func(hooks humacli.Hooks, options *models.Options) {

    println()
    println("=== Starting DH@MPS Vector Database ...")
    fmt.Printf("    Options are debug:%v host:%v port: %v dbhost:%s dbname:%s\n",
      options.Debug, options.Host, options.Port, options.DBHost, options.DBName)

    // Initialize the database
    pool, err := database.InitDB(options)
    if err != nil {
      fmt.Printf("    Unable to connect to database: %v\n", err)
      os.Exit(1)
    }

    // Define standard key generator (for API keys)
    keyGen := handlers.StandardKeyGen{}

    // Create a new router & API
    router := http.NewServeMux()
    api := humago.New(router, huma.DefaultConfig("DHaMPS Vector Database API", "0.0.1"))

    // Add routes to the API
    err = handlers.AddRoutes(pool, keyGen, api)
    if err != nil {
      fmt.Printf("    Unable to add routes: %v\n", err)
      os.Exit(1)
    }

    // Create the HTTP server
    server := &http.Server{
      Addr:    fmt.Sprintf("%s:%d", options.Host, options.Port),
      Handler: router,
    }

    // Start server
    hooks.OnStart(func() {
      fmt.Printf("Starting API server on port %d...\n", options.Port)
      if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
        fmt.Printf("listen: %s\n", err)
      }
    })

    // Gracefully shutdown server
    hooks.OnStop(func() {
      fmt.Printf("Shutting down API server on port %d...\n", options.Port)
      ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
      defer cancel()
      _ = server.Shutdown(ctx)
    })

  })

  // Run the CLI. When passed no commands, it starts the server.
  cli.Run()
}
