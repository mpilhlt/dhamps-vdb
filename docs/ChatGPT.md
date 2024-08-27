# ChatGPT Output

I had ChatGPT (4o) write code implementing an OpenAPI spec developed in dialogue with it.

## OpenAPI spec

```yaml
openapi: 3.0.0
info:
  title: Example API
  version: 1.0.0

servers:
  - url: 'https://example.com/'

paths:
  /admin/users:
    put:
      summary: Create or update a user
      operationId: putAdminUsers
      tags:
        - Admin
      security:
        - bearerAuth: []
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              properties:
                username:
                  type: string
                email:
                  type: string
                apiKey:
                  type: string
      responses:
        '200':
          description: User created or updated
          content:
            application/json:
              schema:
                type: object
                properties:
                  id:
                    type: string
                  username:
                    type: string
                  email:
                    type: string
        '401':
          description: Unauthorized
        '400':
          description: Bad Request

  /admin/users/{user}:
    patch:
      summary: Update a specific user
      operationId: patchAdminUser
      tags:
        - Admin
      security:
        - bearerAuth: []
      parameters:
        - name: user
          in: path
          required: true
          schema:
            type: string
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              properties:
                username:
                  type: string
                email:
                  type: string
                apiKey:
                  type: string
      responses:
        '200':
          description: User updated
        '401':
          description: Unauthorized
        '404':
          description: User not found

    # Repeating this pattern for other methods like GET and DELETE

  /{user}/projects:
    put:
      summary: Create or update a project
      operationId: putUserProjects
      tags:
        - Projects
      security:
        - bearerAuth: []
      # Parameters and request body definition goes here
      responses:
        '200':
          description: Project created or updated
        '401':
          description: Unauthorized

    # Similar for PATCH, GET, DELETE on /{user}/projects/:project

  /{user}/{project}/embeddings:
    put:
      summary: Create or update embeddings for a project
      operationId: putProjectEmbeddings
      tags:
        - Embeddings
      security:
        - bearerAuth: []
      # Define request and response structure
    post:
      summary: Create embeddings for provided text
      operationId: postProjectEmbeddings
      tags:
        - Embeddings
      security:
        - bearerAuth: []
      # Define request and response structure

    # Similar for GET, DELETE on /{user}/{project}/embeddings/:id

  /{user}/{project}/similars/{id}:
    get:
      summary: Retrieve similar items
      operationId: getProjectSimilars
      tags:
        - Similars
      security:
        - bearerAuth: []
      parameters:
        - name: count
          in: query
          schema:
            type: integer
        - name: threshold
          in: query
          schema:
            type: number
      # Define response structure

  /llm-process:
    post:
      summary: Process text with LLM service
      operationId: llmProcess
      tags:
        - LLM
      security:
        - bearerAuth: []
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              properties:
                serviceId:
                  type: string
                projectId:
                  type: string
                contextId:
                  type: string
                textFields:
                  type: array
                  items:
                    type: string
      responses:
        '200':
          description: Processed successfully
        '401':
          description: Unauthorized

components:
  securitySchemes:
    bearerAuth:
      type: http
      scheme: bearer
      bearerFormat: JWT
```

> Missing information and assumptions:
>
> - Detailed request and response schemas for each endpoint are needed.
> - Error response codes need to be defined for each operation.
> - Authentication method is assumed to be Bearer Token (JWT), which needs to be confirmed.
> - Path and query parameters need to be fully specified with types and descriptions.
> - For the llm-process endpoint, the exact nature of "multiple text fields or text-id parameters" needs clarification for accurate schema definition.
> This is a basic structure, and you will need to fill in the details based on the actual requirements and data structures of your API.

Then, requesting PASETO tokens:

```yaml
openapi: 3.0.0
info:
  title: Example API
  version: 1.0.0

servers:
  - url: 'https://example.com/'

paths:
  /admin/users:
    put:
      summary: Create or update a user
      operationId: putAdminUsers
      tags:
        - Admin
      security:
        - pasetoAuth: []
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              properties:
                username:
                  type: string
                email:
                  type: string
                apiKey:
                  type: string
      responses:
        '200':
          description: User created or updated
        '401':
          description: Unauthorized
        # Define other responses as needed

    # Other endpoints definitions go here...

components:
  securitySchemes:
    pasetoAuth:
      type: http
      scheme: bearer
      bearerFormat: PASETO
```

> Please note that while OpenAPI doesn't natively distinguish between JWT and PASETO as they both use Bearer tokens, specifying bearerFormat helps indicate the expected token format. It's also important to document this in the API documentation so that consumers know that the API uses PASETO tokens for authentication.

## Golang implementation

Here is the golang implementation using the huma framework and go's own `http.ServeMux()` router:

Project structure:

```default
example-api/
├── main.go
├── handlers/
│   ├── admin.go
│   ├── projects.go
│   ├── embeddings.go
│   ├── similars.go
│   ├── llm_process.go
└── models/
    ├── user.go
    ├── project.go
    ├── embedding.go
    ├── similar.go
    ├── llm_process.go
```

```{code} go
:filename: main.go
:linenos:
package main

import (
    "net/http"

    "github.com/danielgtaylor/huma/v2"
    "example-api/handlers"
)

func main() {
    router := http.NewServeMux()
    api := huma.NewRouter("Example API", "1.0.0")

    handlers.RegisterAdminRoutes(api)
    handlers.RegisterProjectRoutes(api)
    handlers.RegisterEmbeddingRoutes(api)
    handlers.RegisterSimilarRoutes(api)
    handlers.RegisterLLMProcessRoutes(api)

    router.Handle("/", api)
    http.ListenAndServe(":8080", router)
}
```

```{code} go
:filename: handlers/admin.go
:linenos:
package handlers

import (
    "github.com/danielgtaylor/huma/v2"
    "example-api/models"
)

func RegisterAdminRoutes(api *huma.Router) {
    api.Resource("/admin/users").
        Put("Create or update a user", func(ctx huma.Context, input models.User) {
            // Implement your logic here
        })

    api.Resource("/admin/users/{user}").
        Patch("Update a specific user", func(ctx huma.Context, input models.User) {
            // Implement your logic here
        }).
        Get("Get a specific user", func(ctx huma.Context) {
            // Implement your logic here
        }).
        Delete("Delete a specific user", func(ctx huma.Context) {
            // Implement your logic here
        })
}
```

```{code} go
:filename: handlers/projects.go
:linenos:
package handlers

import (
    "github.com/danielgtaylor/huma/v2"
    "example-api/models"
)

func RegisterProjectRoutes(api *huma.Router) {
    api.Resource("/{user}/projects").
        Put("Create or update a project", func(ctx huma.Context, input models.Project) {
            // Implement your logic here
        })

    api.Resource("/{user}/projects/{project}").
        Patch("Update a specific project", func(ctx huma.Context, input models.Project) {
            // Implement your logic here
        }).
        Get("Get a specific project", func(ctx huma.Context) {
            // Implement your logic here
        }).
        Delete("Delete a specific project", func(ctx huma.Context) {
            // Implement your logic here
        })
}
```

```{code} go
:filename: handlers/embeddings.go
:linenos:
package handlers

import (
    "github.com/danielgtaylor/huma/v2"
    "example-api/models"
)

func RegisterEmbeddingRoutes(api *huma.Router) {
    api.Resource("/{user}/{project}/embeddings").
        Put("Create or update embeddings for a project", func(ctx huma.Context, input models.Embedding) {
            // Implement your logic here
        }).
        Post("Create embeddings for provided text", func(ctx huma.Context, input models.Embedding) {
            // Implement your logic here
        })

    api.Resource("/{user}/{project}/embeddings/{id}").
        Get("Get a specific embedding", func(ctx huma.Context) {
            // Implement your logic here
        }).
        Delete("Delete a specific embedding", func(ctx huma.Context) {
            // Implement your logic here
        })
}
```

```{code} go
:filename: handlers/similars.go
:linenos:
package handlers

import (
    "github.com/danielgtaylor/huma/v2"
    "example-api/models"
)

func RegisterSimilarRoutes(api *huma.Router) {
    api.Resource("/{user}/{project}/similars/{id}").
        Get("Retrieve similar items", func(ctx huma.Context) {
            // Implement your logic here
        })

    api.Resource("/{user}/{project}/similars").
        Post("Retrieve similar items with query", func(ctx huma.Context, input models.SimilarQuery) {
            // Implement your logic here
        })
}
```

```{code} go
:filename: handlers/llm_process.go
:linenos:
package handlers

import (
    "github.com/danielgtaylor/huma/v2"
    "example-api/models"
)

func RegisterLLMProcessRoutes(api *huma.Router) {
    api.Resource("/llm-process").
        Post("Process text with LLM service", func(ctx huma.Context, input models.LLMProcessRequest) {
            // Implement your logic here
        })
}
```

```{code} go
:filename: models/user.go
:linenos:
package models

type User struct {
    Username string `json:"username"`
    Email    string `json:"email"`
    APIKey   string `json:"apiKey"`
}
```

```{code} go
:filename: models/project.go
:linenos:
package models

type Project struct {
    // Define project fields here
}
```

```{code} go
:filename: models/embedding.go
:linenos:
package models

type Embedding struct {
    // Define embedding fields here
}
```

```{code} go
:filename: models/similar.go
:linenos:
package models

type SimilarQuery struct {
    Count     int     `json:"count"`
    Threshold float64 `json:"threshold"`
}
```

```{code} go
:filename: models/llm_process.go
:linenos:
package models

type LLMProcessRequest struct {
    ServiceID  string   `json:"serviceId"`
    ProjectID  string   `json:"projectId"`
    ContextID  string   `json:"contextId"`
    TextFields []string `json:"textFields"`
}
```

> This code sets up the basic structure and routing for your API using the Huma v2 framework and Go's http.ServeMux. Each handler file contains the route definitions for a specific part of the API, and each model file defines the data structures used in the API.
>
> You will need to fill in the implementation details for each handler function and define the full structure of the models as required by your API.
