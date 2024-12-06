package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/mpilhlt/dhamps-vdb/internal/auth"
	"github.com/mpilhlt/dhamps-vdb/internal/database"
	"github.com/mpilhlt/dhamps-vdb/internal/handlers"
	"github.com/mpilhlt/dhamps-vdb/internal/models"

	"github.com/danielgtaylor/huma/v2/adapters/humago"
	"github.com/danielgtaylor/huma/v2/humacli"

	huma "github.com/danielgtaylor/huma/v2"
)

// TODO: Set up limits (e.g. in server definition):
//       <https://huma.rocks/features/request-limits/>

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
		// defer pool.Close()

		// Define standard key generator (for API keys)
		keyGen := handlers.StandardKeyGen{}

		// Create a new router & API
		config := huma.DefaultConfig("DHaMPS Vector Database API", "0.0.1")
		config.Components.SecuritySchemes = auth.Config
		router := http.NewServeMux()
		api := humago.New(router, config)
		api.UseMiddleware(auth.APIKeyAdminAuth(api, options))
		api.UseMiddleware(auth.APIKeyOwnerAuth(api, pool, options))
		api.UseMiddleware(auth.APIKeyReaderAuth(api, pool, options))
		api.UseMiddleware(auth.AuthTermination(api))

		// Add routes to the API
		err = handlers.AddRoutes(pool, keyGen, api)
		if err != nil {
			fmt.Printf("    Unable to add routes: %v\n", err)
			os.Exit(1)
		}

		// Create the HTTP server
		// TODO: Add limits to the server (e.g. timeouts, max header size, etc.)
		server := &http.Server{
			Addr:    fmt.Sprintf("%s:%d", options.Host, options.Port),
			Handler: router,
		}

		// Start server
		hooks.OnStart(func() {
			fmt.Printf("=== Starting API server on port %d...\n\n", options.Port)
			// go func() {
			err := server.ListenAndServe()
			if err != nil && err != http.ErrServerClosed {
				fmt.Printf("listen error: %s\n", err)
			} else {
				fmt.Printf("    API server on port %d stopped.\n", options.Port)
			}
			// }()

			// Keep the program running
			// select {}
		})

		// Gracefully shutdown server
		hooks.OnStop(func() {
			fmt.Printf("\n=== Shutting down API server on port %d...\n", options.Port)

			// Create a context with a timeout for the shutdown process
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			// Attempt to gracefully shut down the server
			if err := server.Shutdown(ctx); err != nil {
				fmt.Printf("Shutdown error: %v\n", err)
			}

			// Close the database pool
			activeConns := pool.Stat().TotalConns()
			fmt.Printf("    Active connections before shutdown: %d\n", activeConns)

			pool.Close()
			fmt.Println("    Database pool successfully closed.")
			fmt.Print("=== DH@MPS Vector Database stopped.\n\n")
		})
	})

	// Run the CLI. When passed no commands, it starts the server.
	cli.Run()
}
