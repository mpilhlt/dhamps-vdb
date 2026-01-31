# LLM Service Architecture Refactoring Guide

## Overview

This refactoring separates LLM services into two concepts:
1. **LLM Service Definitions** - Templates/configurations that can be shared (owned by `_system` or users)
2. **LLM Service Instances** - User-specific instances with optional encrypted API keys

## Changes Made So Far

### 1. Database Migration (004_refactor_llm_services_architecture.sql)
- Created `llm_service_definitions` table for templates
- Renamed `llm_services` to `llm_service_instances`
- Added `api_key_encrypted` column for encrypted API keys
- Created `_system` user for global definitions
- Seeded 5 default LLM service definitions (OpenAI, Cohere, Gemini)
- Changed projects from many-to-many to 1:1 relationship with instances
- Added `llm_service_instances_shared_with` table for instance sharing

### 2. Encryption Module (internal/crypto/encryption.go)
- AES-256-GCM encryption for API keys
- Uses ENCRYPTION_KEY environment variable
- Functions: Encrypt(), Decrypt(), EncryptToBase64(), DecryptFromBase64()

### 3. Database Queries (internal/database/queries/queries.sql)
- Added queries for LLM Service Definitions (UpsertLLMDefinition, GetLLMDefinitionsByUser, GetSystemLLMDefinitions, etc.)
- Updated/added queries for LLM Service Instances (UpsertLLMInstance, ShareLLMInstance, etc.)
- Updated embeddings queries to use `llm_service_instance_id`
- Updated project queries to include `llm_service_instance_id`

## Remaining Work

### Phase 1: Update Models (internal/models/llm_services.go)

Need to create separate model types:

```go
// LLM Service Definition (template)
type LLMServiceDefinition struct {
    DefinitionID     int    `json:"definition_id,omitempty" readOnly:"true"`
    DefinitionHandle string `json:"definition_handle" minLength:"3" maxLength:"20"`
    Owner            string `json:"owner" readOnly:"true"`
    Endpoint         string `json:"endpoint"`
    Description      string `json:"description,omitempty"`
    APIStandard      string `json:"api_standard"`
    Model            string `json:"model"`
    Dimensions       int32  `json:"dimensions"`
}

// LLM Service Instance (user-specific)
type LLMServiceInstance struct {
    InstanceID       int     `json:"instance_id,omitempty" readOnly:"true"`
    InstanceHandle   string  `json:"instance_handle" minLength:"3" maxLength:"20"`
    Owner            string  `json:"owner" readOnly:"true"`
    DefinitionID     *int    `json:"definition_id,omitempty"`  // nullable, can be standalone
    Endpoint         string  `json:"endpoint"`
    Description      string  `json:"description,omitempty"`
    APIKey           string  `json:"api_key,omitempty" writeOnly:"true"`  // Never returned in responses
    APIStandard      string  `json:"api_standard"`
    Model            string  `json:"model"`
    Dimensions       int32   `json:"dimensions"`
    SharedWith       []string `json:"shared_with,omitempty" readOnly:"true"`
}
```

Keep existing `LLMService` for backwards compatibility, mapping to `LLMServiceInstance`.

### Phase 2: Update Handlers

#### a) internal/handlers/llm_services.go
- Update `putLLMFunc` to use `UpsertLLMInstance` and handle encryption
- Update `getLLMFunc` to use `RetrieveLLMInstance` (decrypt API key if needed)
- Update `getUserLLMsFunc` to use `GetLLMInstancesByUser`
- Update `deleteLLMFunc` to use `DeleteLLMInstance`
- Add handlers for sharing instances: `shareLLMInstanceFunc`, `unshareLLMInstanceFunc`

#### b) internal/handlers/projects.go
- Update project creation/update to handle single `llm_service_instance_id`
- Validate that the instance exists and user has access
- Update project retrieval to include instance information
- Remove old `LinkProjectToLLM` calls

#### c) internal/handlers/embeddings.go
- Update embedding creation to get `llm_service_instance_id` from project (not from request)
- Update dimension validation to use instance dimensions
- Update responses to use `instance_handle` instead of `llm_service_handle`

#### d) internal/handlers/admin.go
- Update to use `GetLLMInstancesByProject` (returns single instance)
- Update field names from `LLMServiceID` to `LLMServiceInstanceID`

#### e) New: internal/handlers/llm_definitions.go
Create handlers for LLM Service Definitions:
- `putLLMDefinitionFunc` - Create/update definitions
- `getLLMDefinitionFunc` - Get single definition  
- `getUserLLMDefinitionsFunc` - List user's definitions
- `getSystemLLMDefinitionsFunc` - List _system definitions
- `deleteLLMDefinitionFunc` - Delete definition

### Phase 3: Add Initialization Logic

In `main.go` or `internal/database/db.go`, add initialization after migration:

```go
func InitializeDefaultData(ctx context.Context, pool *pgxpool.Pool) error {
    queries := database.New(pool)
    
    // Check if _system user exists (migration should have created it)
    _, err := queries.RetrieveUser(ctx, "_system")
    if err != nil {
        // Handle error
    }
    
    // Definitions are already seeded in migration 004
    // No additional initialization needed
    
    return nil
}
```

### Phase 4: Environment Variables

Add to `template.env`:
```
# Encryption key for API keys (32+ characters recommended)
ENCRYPTION_KEY=your-secret-encryption-key-here-must-be-kept-secure
```

Update README to document this requirement.

### Phase 5: Update Tests

All test files need updates:

#### internal/handlers/llm_services_test.go
- Update to use Instance terminology
- Test encryption/decryption of API keys
- Test instance sharing
- Test creating instances from definitions

#### internal/handlers/projects_test.go
- Test 1:1 relationship with instances
- Test that projects require an instance
- Test updating project's instance

#### internal/handlers/embeddings_test.go
- Update to use `llm_service_instance_id`
- Test that embeddings use project's instance

#### internal/handlers/admin_test.go
- Update field names and queries

### Phase 6: Migration Strategy for Existing Installations

The migration (004) handles data migration automatically:
- Existing `llm_services` → `llm_service_instances` (renamed)
- First linked instance per project → `project.llm_service_instance_id`
- API keys remain in plaintext initially (in `api_key` column)

**Post-migration steps for admins:**
1. Set `ENCRYPTION_KEY` environment variable
2. Restart service (will now encrypt new API keys)
3. Optional: Manually migrate existing plaintext API keys to encrypted format
4. Optional: Remove old `api_key` column after full migration

### Recommended Implementation Order

1. Create/update model files with new types
2. Update llm_services.go handlers (core functionality)
3. Update projects.go handlers (1:1 relationship)
4. Update embeddings.go handlers (use project's instance)
5. Update admin.go handlers (field names)
6. Create llm_definitions.go handlers (new functionality)
7. Update all tests
8. Test full workflow:
   - Create user
   - Create LLM service instance (from definition or standalone)
   - Create project with instance
   - Upload embeddings
   - Query similar embeddings
9. Test instance sharing
10. Update documentation

## API Changes

### Backwards Compatibility

To maintain backwards compatibility, existing endpoints can continue to work:
- `PUT /v1/llm-services/{user}/{handle}` → Creates/updates an instance
- `GET /v1/llm-services/{user}/{handle}` → Gets an instance (no API key in response)
- `DELETE /v1/llm-services/{user}/{handle}` → Deletes an instance

### New Endpoints (Optional)

Consider adding:
- `GET /v1/llm-definitions` → List all available definitions (system + user's own)
- `GET /v1/llm-definitions/_system` → List system definitions
- `POST /v1/llm-definitions/{user}` → Create user definition
- `POST /v1/llm-instances/{user}/from-definition/{definition_handle}` → Create instance from definition
- `POST /v1/llm-instances/{user}/{handle}/share` → Share instance with another user
- `DELETE /v1/llm-instances/{user}/{handle}/share/{shared_user}` → Unshare

## Security Considerations

1. **API Key Encryption**: Always encrypt API keys before storing
2. **API Key Response**: Never return decrypted API keys in API responses
3. **Instance Sharing**: Shared users can use the instance but cannot see API keys
4. **Encryption Key Management**: `ENCRYPTION_KEY` must be kept secure and backed up
5. **Key Rotation**: Changing `ENCRYPTION_KEY` requires re-encrypting all API keys

## Testing the Refactoring

1. Run unit tests: `go test ./internal/crypto/`
2. Run integration tests: `go test -v ./...` (requires testcontainers)
3. Manual testing:
   - Create definitions as _system user
   - Create instances as regular users
   - Test instance sharing
   - Test project-instance 1:1 relationship
   - Test encryption/decryption

## Questions to Address

1. **Default Instance**: Should new projects require an instance ID, or should we create a default?
   - Recommendation: Require instance ID to enforce explicit choice
   - Alternative: Auto-create a default instance for the user if none exists

2. **Instance Creation**: Should instance creation from definition copy all fields?
   - Recommendation: Yes, but allow overriding endpoint, model, dimensions
   - User only needs to provide their API key

3. **Shared Instance Access**: Can shared users modify the instance?
   - Recommendation: No, only owner can modify. Shared users have read-only access.

4. **Definition Immutability**: Should _system definitions be immutable?
   - Recommendation: Yes, only admin can modify. Users can create their own definitions.
