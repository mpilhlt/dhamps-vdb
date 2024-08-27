package main

import (
  "context"
  "fmt"
  "net/http"
  "time"

  "github.com/mpilhlt/dhamps-vdb/internal/handlers"
  "github.com/mpilhlt/dhamps-vdb/internal/models"
  "github.com/mpilhlt/dhamps-vdb/internal/database"

  "github.com/danielgtaylor/huma/v2/adapters/humago"
  "github.com/danielgtaylor/huma/v2/humacli"

  huma "github.com/danielgtaylor/huma/v2"
)

func main() {
  // Create a CLI app
  cli := humacli.New(func(hooks humacli.Hooks, options *models.Options) {

    fmt.Printf("Options are debug:%v host:%v port: %v dbhost:%s dbname:%s\n",
      options.Debug, options.Host, options.Port, options.DBHost, options.DBName)

    // Initialize the database
    database.InitDB(options)

    // Create a new router & API
    router := http.NewServeMux()
    api := humago.New(router, huma.DefaultConfig("DHaMPS Vector Database API", "0.0.1"))

    // Add routes to the API
    addRoutes(api)

    // Create the HTTP server
    server := &http.Server{
      Addr:    fmt.Sprintf("%s:%d", options.Host, options.Port),
      Handler: router,
      // TODO: Set up timeouts!
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
      server.Shutdown(ctx)
    })

  })

  // Run the CLI. When passed no commands, it starts the server.
  cli.Run()
}

func addRoutes(api huma.API) {
  handlers.RegisterUsersRoutes(api)
  handlers.RegisterProjectsRoutes(api)
  handlers.RegisterEmbeddingsRoutes(api)
  // handlers.RegisterSimilarRoutes(api)
  // handlers.RegisterLLMProcessRoutes(api)
}
