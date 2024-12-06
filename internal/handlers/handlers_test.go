package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/mpilhlt/dhamps-vdb/internal/auth"
	"github.com/mpilhlt/dhamps-vdb/internal/database"
	"github.com/mpilhlt/dhamps-vdb/internal/handlers"
	"github.com/mpilhlt/dhamps-vdb/internal/models"

	huma "github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humago"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

// TODO: Use values from env or config to override default options.
// TODO: Set up timeouts!

// Each package ("handlers", in this case) can have its own TestMain function.
// This function is executed before any tests in the package are run and can
// be used to set up resources needed by the tests. It can also be used to
// run setup code that should only be run once for the entire package.
// It has a signature of func TestMain(m *testing.M), where m has a single
// method Run() that runs all the tests in the package. It should call os.Exit
// with the result of m.Run() to signal the test runner whether the tests
// passed or failed.

// While there is the humago package to set up a testing API against which we
// could register our routes and run requests, we use an actual API connecting
// to a real database. We still can choose between a live postgres database and
// a testcontainer spun up just for testing.

var (
	options = models.Options{
		Debug:      true,
		Host:       "localhost",
		Port:       8080,
		DBHost:     "localhost",
		DBName:     "testdb",
		DBUser:     "test",
		DBPassword: "test",
		AdminKey:   "Password123",
	}
	connPool *pgxpool.Pool
	teardown func()
)

// TestMain function initializes the database container.
// Then it runs all the tests. Setup of api, router and server
// is done in the tests themselves to provide better isolation.
func TestMain(m *testing.M) {
	// Get a database connection pool
	var err error
	connPool, err, teardown = getTestDatabase()
	if err != nil {
		fmt.Printf("Unable to get database connection pool: %v", err)
		teardown()
		os.Exit(1)
	}
	if connPool == nil {
		fmt.Print("Database connection pool is nil")
		teardown()
		os.Exit(1)
	}
	defer connPool.Close()
	defer teardown()
	fmt.Print("\n    Database ready\n    Running tests ...\n\n")

	// Run the tests
	code := m.Run() // Execute all the tests

	os.Exit(code)
}

// --- Helper functions and types ---

// GetTestDatabase spins up a new Postgres container and returns
// a connection pool, an error value and a closure.
// Please always make sure to call the closure as it is the teardown
// function terminating the container.
func getTestDatabase() (*pgxpool.Pool, error, func()) {
	ctx := context.Background()

	// 1. Run PostgreSQL container
	pgVectorContainer, err := postgres.Run(ctx,
		// "pgvector/pgvector:pg16",
		"pgvector/pgvector:0.7.4-pg16",
		postgres.WithDatabase(options.DBName),
		postgres.WithUsername(options.DBUser),
		postgres.WithPassword(options.DBPassword),
		postgres.WithInitScripts(filepath.Join("..", "..", "testdata", "postgres", "enable-vector.sql")),
		testcontainers.WithWaitStrategy(
			// First, we wait for the container to log readiness twice.
			// This is because it will restart itself after the first startup.
			wait.ForLog("database system is ready to accept connections").WithOccurrence(2).WithStartupTimeout(120*time.Second),
			// Then, we wait for docker to actually serve the port on localhost.
			// For non-linux OSes like Mac and Windows, Docker or Rancher Desktop will have to start a separate proxy.
			// Without this, the tests will be flaky on those OSes!
			wait.ForListeningPort("5432/tcp").WithStartupTimeout(120*time.Second),
		),
	)
	if err != nil {
		fmt.Printf("Error creating container: %v\n", err)
		return nil, err, nil
	}
	connStr, err := pgVectorContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		fmt.Printf("Error reading connection string: %v\n", err)
		return nil, err, nil
	}

	// 2. Connect to db
	connPool, err := pgxpool.New(context.Background(), connStr)
	if err != nil {
		fmt.Printf("Error creating connection pool: %v\n", err)
		return nil, err, nil
	}
	err = connPool.Ping(context.Background())
	if err != nil {
		fmt.Printf("Error pinging connection pool: %v\n", err)
		return nil, err, nil
	}
	fmt.Printf("Connection pool of database %v/%v established.\n", options.DBHost, options.DBName)

	// 3. Prepare test database
	err = database.VerifySchema(ctx, connStr)
	if err != nil {
		fmt.Printf("Error preparing test database: %v\n", err)
		return nil, err, nil
	}

	return connPool, nil, func() {
		err := pgVectorContainer.Terminate(context.Background())
		if err != nil {
			fmt.Printf("Error terminating container: %v\n", err)
		}
	}
}

// setupServer sets up server, router and API for testing.
// It returns an error value and a closure function that
// should be called to clean up.
// It is supposed to be called from the various tests.
func startTestServer(t *testing.T, pool *pgxpool.Pool, keyGen handlers.RandomKeyGenerator) (error, func()) {
	/* Use transactions to isolate tests from each other.

	   // Get a database connection
	   conn, err := pool.Acquire(context.Background())
	   if err != nil {
	       t.Fatal(err)
	   }

	   // Start a transaction
	   _, err = conn.Exec(context.Background(), "BEGIN")
	   if err != nil {
	       t.Fatal(err)
	   }
	*/

	// Create a new router & API
	config := huma.DefaultConfig("DHaMPS Vector Database API", "0.0.1")
	config.Components.SecuritySchemes = auth.Config
	router := http.NewServeMux()
	api := humago.New(router, config)
	api.UseMiddleware(auth.APIKeyAdminAuth(api, &options))
	api.UseMiddleware(auth.APIKeyOwnerAuth(api, pool, &options))
	api.UseMiddleware(auth.APIKeyReaderAuth(api, pool, &options))
	api.UseMiddleware(auth.AuthTermination(api))

	err := handlers.AddRoutes(pool, keyGen, api)
	if err != nil {
		fmt.Printf("Unable to add routes to API: %v", err)
		return err, func() {}
	}
	fmt.Print("    Router ready\n")

	/* HTTP Server setup (we set up httptest.Server below instead)
	   // Create the HTTP server
	   server := &http.Server{
	     Addr:    fmt.Sprintf("%s:%d", options.Host, options.Port),
	     Handler: router,
	   }
	   // Start the server
	   go func() {
	     if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
	       fmt.Printf("Unable to start server: %v", err)
	       server.Close()    // Close the server
	       connPool.Close()  // Close the database connection
	       teardown()
	       os.Exit(1)
	     }
	   }()
	   time.Sleep(1 * time.Second) // Wait for the server to start
	*/

	// For testing, we use a httptest.Server instead of a real server.
	// Running this on our custom port requires setting up a listener...
	// create a listener with the desired port.
	l, err := net.Listen("tcp", fmt.Sprintf("%s:%d", options.Host, options.Port))
	assert.NoError(t, err)
	if err != nil {
		fmt.Printf("Error setting up listener: %v", err)
		return err, func() {}
	}
	// Create a new server with the router.
	server := httptest.NewUnstartedServer(router)
	// NewUnstartedServer creates a server-cum-listener.
	// Close that listener and replace with the one we created.
	server.Listener.Close()
	server.Listener = l
	// Start the server.
	server.Start()
	fmt.Printf("    Server listening on %s:%d\n", options.Host, options.Port)

	cleanup := func() {
		// Close the server
		server.Close()
		/* Clean up transactions.
		   _, err := conn.Exec(context.Background(), "ROLLBACK")
		   if err != nil {
		       t.Fatal(err)
		   }
		   conn.Release()
		*/
	}

	return nil, cleanup
}

// MockKeyGen is a mock implementation of the RandomKeyGenerator interface.
type MockKeyGen struct{ mock.Mock }

// Implement mock's randomKey method
func (m *MockKeyGen) RandomKey(len int) (string, error) {
	args := m.Called(len)
	return args.String(0), args.Error(1)
}

// createUser creates a user and returns the API key and an error value
// it accepts a JSON string encoding the user object as input
func createUser(t *testing.T, userJSON string) (string, error) {
	fmt.Print("    Creating user (\"alice\") for testing ...\n")
	requestURL := fmt.Sprintf("http://%s:%d/v1/users/alice", options.Host, options.Port)
	requestBody := bytes.NewReader([]byte(userJSON))
	req, err := http.NewRequest(http.MethodPut, requestURL, requestBody)
	assert.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+options.AdminKey)
	resp, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	defer resp.Body.Close()

	// get API key for user alice from response body
	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	// fmt.Printf("Response body: %v\n", string(body))
	userInfo := models.HandleAPIStruct{}
	err = json.Unmarshal(body, &userInfo)
	assert.NoError(t, err)
	fmt.Printf("        Successfully created user (handle: \"%s\", apiKey: \"%s\").\n", userInfo.UserHandle, userInfo.APIKey)
	// fmt.Printf("        User info: %v\n", userInfo)
	return userInfo.APIKey, nil
}

// createProject creates a project and returns the project ID and an error value
// it accepts a JSON string encoding the project object as input
func createProject(t *testing.T, projectJSON string, user string, apiKey string) (int, error) {
	fmt.Print("    Creating project ")
	jsonInput := &struct {
		Handle      string `json:"project_handle" doc:"Handle of created or updated project"`
		Description string `json:"description" doc:"Description of the project"`
	}{}
	err := json.Unmarshal([]byte(projectJSON), jsonInput)
	if err != nil {
		fmt.Printf("Error unmarshalling project JSON: %v\n", err)
	}
	assert.NoError(t, err)
	fmt.Printf("(\"%s/%s\") for testing ...\n", user, jsonInput.Handle)

	requestURL := fmt.Sprintf("http://%s:%d/v1/projects/%s/%s", options.Host, options.Port, user, jsonInput.Handle)
	requestBody := bytes.NewReader([]byte(projectJSON))
	req, err := http.NewRequest(http.MethodPut, requestURL, requestBody)
	req.Header.Set("Authorization", "Bearer "+apiKey)
	assert.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	defer resp.Body.Close()

	// get project ID from response body
	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)

	projectInfo := &struct {
		ProjectHandle string `json:"project_handle" doc:"Handle of created or updated project"`
		ProjectID     int    `json:"project_id" doc:"Unique project identifier"`
	}{}
	err = json.Unmarshal(body, &projectInfo)
	if err != nil {
		fmt.Printf("Error unmarshalling project info: %v\nbody: %v", err, string(body))
	}
	assert.NoError(t, err)

	fmt.Printf("        Successfully created project (handle \"%s/%s\", id \"%d\").\n", user, projectInfo.ProjectHandle, projectInfo.ProjectID)
	return projectInfo.ProjectID, nil
}

// createAPIStandard creates an API standard defintion for testing and returns the handle and an error value
// it accepts a JSON string encoding the API standard object as input
func createAPIStandard(t *testing.T, apiStandardJSON string, apiKey string) (string, error) {
	fmt.Print("    Creating API standard ")
	jsonInput := &struct {
		APIStandardHandle string `json:"api_standard_handle" doc:"Handle of created or updated api standard"`
		Description       string `json:"description" doc:"Description of the api standard"`
		KeyMethod         string `json:"key_method" doc:"Method used to authenticate the API standard"`
		KeyField          string `json:"key_field" doc:"Field in the request used to authenticate the API standard"`
	}{}
	err := json.Unmarshal([]byte(apiStandardJSON), jsonInput)
	if err != nil {
		fmt.Printf("\nError unmarshalling api standard JSON: %v\n", err)
	}
	assert.NoError(t, err)
	fmt.Printf("(\"%s\") for testing ...\n", jsonInput.APIStandardHandle)

	requestURL := fmt.Sprintf("http://%s:%d/v1/api-standards/%s", options.Host, options.Port, jsonInput.APIStandardHandle)
	requestBody := bytes.NewReader([]byte(apiStandardJSON))
	req, err := http.NewRequest(http.MethodPut, requestURL, requestBody)
	req.Header.Set("Authorization", "Bearer "+apiKey)
	assert.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	defer resp.Body.Close()

	// get API handle from response body
	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)

	APIStandardInfo := &struct {
		APIStandardHandle string `json:"api_standard_handle" doc:"Handle of created or updated api standard"`
	}{}
	err = json.Unmarshal(body, &APIStandardInfo)
	if err != nil {
		fmt.Printf("Error unmarshalling api standard info: %v\nbody: %v", err, string(body))
	}
	assert.NoError(t, err)

	fmt.Printf("        Successfully created API Standard (handle \"%s\").\n", APIStandardInfo.APIStandardHandle)
	return APIStandardInfo.APIStandardHandle, nil
}

// createLLMService creates an LLM service definition for testing and returns the handle and an error value
// it accepts a JSON string encoding the LLM service object as input
func createLLMService(t *testing.T, llmServiceJSON string, user, apiKey string) (string, error) {
	fmt.Print("    Creating LLM service ")
	jsonInput := &struct {
		LLMServiceHandle string `json:"llm_service_handle" doc:"Handle of created or updated LLM service"`
		Description      string `json:"description" doc:"Description of the LLM service"`
	}{}
	err := json.Unmarshal([]byte(llmServiceJSON), jsonInput)
	if err != nil {
		fmt.Printf("Error unmarshalling LLM service JSON: %v\n", err)
	}
	assert.NoError(t, err)
	fmt.Printf("(\"%s\") for testing ...\n", jsonInput.LLMServiceHandle)

	requestURL := fmt.Sprintf("http://%s:%d/v1/llm-services/%s/%s", options.Host, options.Port, user, jsonInput.LLMServiceHandle)
	requestBody := bytes.NewReader([]byte(llmServiceJSON))
	req, err := http.NewRequest(http.MethodPut, requestURL, requestBody)
	req.Header.Set("Authorization", "Bearer "+apiKey)
	assert.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	defer resp.Body.Close()

	// get LLM service handle from response body
	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)

	LLMServiceInfo := &struct {
		LLMServiceHandle string `json:"llm_service_handle" doc:"Handle of created or updated LLM service"`
	}{}
	err = json.Unmarshal(body, &LLMServiceInfo)
	if err != nil {
		fmt.Printf("Error unmarshalling LLM service info: %v\nbody: %v", err, string(body))
	}
	assert.NoError(t, err)

	fmt.Printf("        Successfully created LLM Service (handle \"%s\").\n", LLMServiceInfo.LLMServiceHandle)
	return LLMServiceInfo.LLMServiceHandle, nil
}

// isJSON checks if a string is a valid JSON.
func isJSON(str string) bool {
	var js json.RawMessage
	return json.Unmarshal([]byte(str), &js) == nil
}
