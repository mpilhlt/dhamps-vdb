# Public Access to Embeddings

## Overview

Projects can be configured to allow unauthenticated (public) read access to embeddings and similar documents by including the special value `"*"` in the `authorizedReaders` array when creating or updating a project.

## Usage

### Creating a Public Project

When creating or updating a project, include `"*"` in the `authorizedReaders` field:

```json
{
  "project_handle": "my-public-project",
  "description": "A publicly accessible project",
  "authorizedReaders": ["*"]
}
```

### Endpoints with Public Access

When a project has public read access enabled, the following endpoints can be accessed without authentication:

- `GET /v1/projects/{user}/{project}` - Retrieve project metadata (including owner and authorizedReaders)
- `GET /v1/embeddings/{user}/{project}` - Retrieve all embeddings for the project
- `GET /v1/embeddings/{user}/{project}/{text_id}` - Retrieve a specific embedding
- `GET /v1/similars/{user}/{project}/{text_id}` - Find similar documents

### Endpoints Requiring Authentication

Even for public projects, the following operations still require authentication:

- `POST /v1/embeddings/{user}/{project}` - Create new embeddings
- `DELETE /v1/embeddings/{user}/{project}` - Delete all embeddings
- `DELETE /v1/embeddings/{user}/{project}/{text_id}` - Delete a specific embedding

## Implementation Details

### Database Schema

A `public_read` boolean flag is stored in the `projects` table to indicate whether a project allows public access.

### Authentication Flow

1. When a request is made to a reader-protected endpoint, the middleware checks if authentication is required
2. If the project has `public_read` set to true, the request is allowed without an Authorization header
3. Unauthenticated requests are logged with the user set to "public"
4. If `public_read` is false or not set, normal authentication rules apply

### Backwards Compatibility

For backwards compatibility, when `"*"` is included in `authorizedReaders`:
- The `public_read` flag is set to true (enabling unauthenticated access)
- All existing users are still added to the `users_projects` table as readers
- This ensures that existing authentication mechanisms continue to work

### Project Metadata Display

When a project has `public_read` enabled:
- The `authorizedReaders` field will display `["*"]` instead of an expanded list of all users
- This makes it clear that the project is publicly accessible
- Anonymous users can view project metadata including owner, description, and the `["*"]` indicator

## Security Considerations

- Public access only applies to read operations (GET requests)
- Write operations (POST, PUT, DELETE) always require authentication
- Project metadata and ownership information is publicly visible for public projects
- The admin and owner authentication mechanisms are unaffected

## Examples

### Accessing a Public Project Without Authentication

```bash
# Get project metadata without authentication
curl http://localhost:8080/v1/projects/alice/public-project
# Returns: {"project_handle": "public-project", "owner": "alice", "authorizedReaders": ["*"], ...}

# Get all embeddings without authentication
curl http://localhost:8080/v1/embeddings/alice/public-project

# Get a specific embedding without authentication
curl http://localhost:8080/v1/embeddings/alice/public-project/text123

# Find similar documents without authentication
curl http://localhost:8080/v1/similars/alice/public-project/text123
```

### Creating Embeddings Still Requires Authentication

```bash
# This will fail with 401 Unauthorized
curl -X POST http://localhost:8080/v1/embeddings/alice/public-project \
  -H "Content-Type: application/json" \
  -d '{"embeddings": [...]}'

# This will succeed with a valid API key
curl -X POST http://localhost:8080/v1/embeddings/alice/public-project \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"embeddings": [...]}'
```

## Migration

Existing projects are not affected. The `public_read` flag defaults to `false`, so all existing projects continue to require authentication for read operations unless explicitly updated to include `"*"` in their `authorizedReaders`.
