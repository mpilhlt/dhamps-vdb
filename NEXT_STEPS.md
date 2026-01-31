# Next Steps to Complete LLM Service Architecture Refactoring

## Current Status
✅ Database migration created (004)
✅ Encryption module implemented and tested
✅ SQLC queries updated and generated
❌ Code does not compile (handlers need updates)

## Quick Fix Guide

### Step 1: Update Models (5-10 minutes)

Edit `internal/models/llm_services.go`, keep all existing code and ADD:

```go
// New types for the refactored architecture
type LLMServiceInstance struct {
InstanceID       int    `json:"instance_id,omitempty" readOnly:"true"`
InstanceHandle   string `json:"instance_handle"`
Owner            string `json:"owner" readOnly:"true"`
DefinitionID     *int32 `json:"definition_id,omitempty"`
Endpoint         string `json:"endpoint"`
Description      string `json:"description,omitempty"`
APIKey           string `json:"api_key,omitempty" writeOnly:"true"`
APIStandard      string `json:"api_standard"`
Model            string `json:"model"`
Dimensions       int32  `json:"dimensions"`
}

type LLMServiceDefinition struct {
DefinitionID     int    `json:"definition_id,omitempty" readOnly:"true"`
DefinitionHandle string `json:"definition_handle"`
Owner            string `json:"owner" readOnly:"true"`
Endpoint         string `json:"endpoint"`
Description      string `json:"description,omitempty"`
APIStandard      string `json:"api_standard"`
Model            string `json:"model"`
Dimensions       int32  `json:"dimensions"`
}
```

### Step 2: Update llm_services.go Handler (15-20 minutes)

In `internal/handlers/llm_services.go`, replace database calls:

**Before:**
```go
llm, err := queries.UpsertLLM(ctx, database.UpsertLLMParams{...})
```

**After:**
```go
// Get encryption key
encKey, err := crypto.GetEncryptionKeyFromEnv()
if err != nil {
    // Optionally fallback to unencrypted if key not set
}

// Encrypt API key if provided and key available
var apiKeyEncrypted []byte
if input.Body.APIKey != "" && encKey != nil {
    apiKeyEncrypted, err = encKey.Encrypt(input.Body.APIKey)
    if err != nil {
        return fmt.Errorf("unable to encrypt API key: %v", err)
    }
}

llm, err := queries.UpsertLLMInstance(ctx, database.UpsertLLMInstanceParams{
    Owner:            input.UserHandle,
    InstanceHandle:   input.LLMServiceHandle,
    DefinitionID:     pgtype.Int4{Valid: false}, // No definition ref
    Endpoint:         input.Body.Endpoint,
    Description:      pgtype.Text{String: input.Body.Description, Valid: true},
    APIKey:           pgtype.Text{String: input.Body.APIKey, Valid: input.Body.APIKey != ""},
    APIKeyEncrypted:  apiKeyEncrypted,
    APIStandard:      input.Body.APIStandard,
    Model:            input.Body.Model,
    Dimensions:       int32(input.Body.Dimensions),
})
```

Update response mapping:
```go
llmServiceID = llm.InstanceID
llmServiceHandle = llm.InstanceHandle
```

Similarly update:
- `LinkUserToLLM` → `LinkUserToLLMInstance`
- `RetrieveLLM` → `RetrieveLLMInstance`
- `DeleteLLM` → `DeleteLLMInstance`
- `GetLLMsByUser` → `GetLLMInstancesByUser`

### Step 3: Update projects.go Handler (10-15 minutes)

In `internal/handlers/projects.go`:

**Before:**
```go
err = queries.UpsertProject(ctx, database.UpsertProjectParams{
    ProjectHandle:   input.ProjectHandle,
    Owner:           input.UserHandle,
    Description:     pgtype.Text{...},
    MetadataScheme:  pgtype.Text{...},
    PublicRead:      ...,
})
```

**After:**
```go
// Determine instance ID - could be from input or validation logic
var instanceID pgtype.Int4
// For now, allow nullable (projects can exist without instance initially)
instanceID = pgtype.Int4{Valid: false}

err = queries.UpsertProject(ctx, database.UpsertProjectParams{
    ProjectHandle:          input.ProjectHandle,
    Owner:                  input.UserHandle,
    Description:            pgtype.Text{...},
    MetadataScheme:         pgtype.Text{...},
    PublicRead:             ...,
    LlmServiceInstanceID:   instanceID,
})
```

Remove old `LinkProjectToLLM` calls.

### Step 4: Update embeddings.go Handler (5-10 minutes)

In `internal/handlers/embeddings.go`:

Change field name in `UpsertEmbeddingsParams`:
```go
LLMServiceID: llmServiceID  // OLD
```
to:
```go
LlmServiceInstanceID: llmServiceID  // NEW (note: check exact field name in generated code)
```

Update response field names:
```go
LLMServiceHandle  // OLD
→ InstanceHandle   // NEW
```

### Step 5: Update admin.go Handler (5 minutes)

In `internal/handlers/admin.go`:

```go
GetLLMsByProject → GetLLMInstancesByProject
LLMServiceID → LlmServiceInstanceID (check generated field name)
```

### Step 6: Test Build (2 minutes)

```bash
go build -o /tmp/dhamps-vdb main.go
```

If successful, proceed to Step 7. If not, fix remaining compilation errors.

### Step 7: Update Tests (30-60 minutes)

For each test file:
1. Update struct field names
2. Update query function names  
3. Run tests: `go test -v ./internal/handlers/`

Start with:
- `llm_services_test.go`
- `projects_test.go`
- `embeddings_test.go`
- `admin_test.go`

### Step 8: Add Environment Variable

Add to `template.env`:
```
# Required for encrypting LLM service API keys
ENCRYPTION_KEY=change-this-to-a-long-random-string-at-least-32-characters
```

### Step 9: Update README (10 minutes)

Add to README.md under "Environment Variables":
```markdown
- `ENCRYPTION_KEY`: Required for encrypting API keys stored for LLM service instances. 
  Must be a secure random string. If not set, API keys will be stored in plaintext (not recommended).
```

Add section on new architecture:
```markdown
## LLM Service Architecture

The system now separates:
- **LLM Service Definitions**: Templates (owned by `_system` or users) with service configurations
- **LLM Service Instances**: User-specific instances with optional encrypted API keys

When creating a project, you can specify which LLM service instance to use.
API keys are encrypted using AES-256-GCM before storage.
```

### Step 10: Test Migration (15-20 minutes)

1. Set up test database with old schema
2. Run migration 004
3. Verify:
   - `_system` user created
   - 5 definitions seeded
   - `llm_service_instances` table exists
   - Projects have `llm_service_instance_id` column
4. Test CRUD operations on instances

## Complete Implementation Estimate

- **Minimum (get it compiling)**: 1-2 hours
- **Full (with tests)**: 3-4 hours
- **Production-ready (with docs)**: 4-6 hours

## Key Files Summary

### Already Complete ✅
- `internal/database/migrations/004_refactor_llm_services_architecture.sql`
- `internal/crypto/encryption.go`
- `internal/crypto/encryption_test.go`
- `internal/database/queries/queries.sql`
- Generated SQLC code

### Need Updates ⚠️
- `internal/models/llm_services.go` - Add new types
- `internal/handlers/llm_services.go` - Update to use instances + encryption
- `internal/handlers/projects.go` - Add instance_id handling
- `internal/handlers/embeddings.go` - Update field names
- `internal/handlers/admin.go` - Update field names
- All `*_test.go` files
- `template.env` - Add ENCRYPTION_KEY
- `README.md` - Document new architecture

## Troubleshooting

**"ENCRYPTION_KEY not set" error:**
- Set environment variable: `export ENCRYPTION_KEY="your-key-here"`
- Or modify code to allow fallback to unencrypted (not recommended)

**"field not found" errors:**
- Check generated `internal/database/models.go` for exact field names
- SQLC may have different capitalization (e.g., `LlmServiceInstanceID` vs `LLMServiceInstanceID`)

**Migration fails:**
- Check database user has sufficient privileges
- Verify no foreign key conflicts
- Test rollback script if needed

**Tests fail after update:**
- Update test data to include instance IDs
- Update assertions to use new field names
- Verify test database has migration applied
