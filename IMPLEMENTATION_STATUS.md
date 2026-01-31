# LLM Service Architecture Refactoring - Implementation Status

## ✅ Completed Work

### 1. Database Schema (Migration 004)
**File:** `internal/database/migrations/004_refactor_llm_services_architecture.sql`

**What was done:**
- Created `llm_service_definitions` table for reusable service templates
- Renamed `llm_services` → `llm_service_instances` to represent user-specific instances
- Added `_system` user for global definitions
- Added `api_key_encrypted` BYTEA column for encrypted API storage
- Modified `projects` table to have `llm_service_instance_id` (1:1 relationship)
- Removed `projects_llm_services` many-to-many table
- Added `llm_service_instances_shared_with` table for instance sharing (n:m)
- Seeded 5 default LLM service definitions:
  - openai-large (3072 dimensions)
  - openai-small (1536 dimensions)  
  - cohere-multilingual-v3 (1024 dimensions)
  - cohere-v4 (1536 dimensions)
  - gemini-embedding-001 (768 dimensions)

**Migration handles:**
- Automatic renaming of tables and columns
- Data migration: First linked LLM service per project → project.llm_service_instance_id
- Backward and forward migration scripts (create above / drop below)

### 2. Encryption Module
**File:** `internal/crypto/encryption.go`

**Features:**
- AES-256-GCM encryption for API keys
- Uses `ENCRYPTION_KEY` environment variable
- SHA256 hashing of key to ensure 32-byte key size
- Functions:
  - `NewEncryptionKey(keyString)` - Create key from string
  - `GenerateEncryptionKey()` - Generate random key
  - `GetEncryptionKeyFromEnv()` - Read from environment
  - `Encrypt(plaintext) → []byte` - Encrypt to bytes
  - `Decrypt(ciphertext) → string` - Decrypt bytes
  - `EncryptToBase64(plaintext) → string` - Encrypt to base64
  - `DecryptFromBase64(base64) → string` - Decrypt from base64

**Testing:** Full test coverage in `internal/crypto/encryption_test.go`
- All tests passing ✅

### 3. Database Queries (SQLC)
**File:** `internal/database/queries/queries.sql`

**New queries added:**

**LLM Service Definitions:**
- `UpsertLLMDefinition` - Create/update definition
- `DeleteLLMDefinition` - Delete definition
- `RetrieveLLMDefinition` - Get single definition
- `GetLLMDefinitionsByUser` - List user's definitions
- `GetAllLLMDefinitions` - List all definitions
- `GetSystemLLMDefinitions` - List _system definitions

**LLM Service Instances:**
- `UpsertLLMInstance` - Create/update instance (with encryption support)
- `DeleteLLMInstance` - Delete instance
- `RetrieveLLMInstance` - Get single instance
- `RetrieveLLMInstanceByID` - Get instance by ID
- `LinkUserToLLMInstance` - Link user to instance
- `ShareLLMInstance` - Share instance with another user
- `UnshareLLMInstance` - Remove instance sharing
- `GetSharedUsersForInstance` - List users instance is shared with
- `GetLLMInstancesByProject` - Get instance for project (1:1)
- `GetLLMInstancesByUser` - List user's instances
- `GetSharedLLMInstances` - List instances shared with user

**Updated queries:**
- `UpsertProject` - Now includes `llm_service_instance_id`
- `UpsertEmbeddings` - Uses `llm_service_instance_id`
- `RetrieveEmbeddings` - Joins with instances table
- `GetEmbeddingsByProject` - Joins with instances table
- `GetSimilarsByVector` - Uses instances table

**SQLC code generated:** ✅ (`internal/database/models.go`, `internal/database/queries.sql.go`)

## 🚧 Remaining Work

### Phase 1: Go Models (HIGH PRIORITY)
**File:** `internal/models/llm_services.go`

**Need to add:**
```go
// LLM Service Definition (template)
type LLMServiceDefinition struct {
    DefinitionID     int    
    DefinitionHandle string 
    Owner            string 
    Endpoint         string 
    Description      string 
    APIStandard      string 
    Model            string 
    Dimensions       int32  
}

// LLM Service Instance (user-specific with API key)
type LLMServiceInstance struct {
    InstanceID       int     
    InstanceHandle   string  
    Owner            string  
    DefinitionID     *int    // Optional reference to definition
    Endpoint         string  
    Description      string  
    APIKey           string  // Write-only, never returned
    APIStandard      string  
    Model            string  
    Dimensions       int32   
}
```

**Strategy:** Keep existing `LLMService` as alias for `LLMServiceInstance` for backwards compatibility.

### Phase 2: Handlers (HIGH PRIORITY)

**Files to update:**

1. **`internal/handlers/llm_services.go`**
   - Replace `UpsertLLM` → `UpsertLLMInstance`
   - Replace `RetrieveLLM` → `RetrieveLLMInstance`
   - Replace `GetLLMsByUser` → `GetLLMInstancesByUser`
   - Replace `DeleteLLM` → `DeleteLLMInstance`
   - Add encryption/decryption of API keys
   - Update field names: `LLMServiceID` → `InstanceID`, `LLMServiceHandle` → `InstanceHandle`

2. **`internal/handlers/projects.go`**
   - Add `llm_service_instance_id` to project creation
   - Validate instance exists and user has access
   - Remove old `LinkProjectToLLM` calls
   - Update project retrieval to include instance info

3. **`internal/handlers/embeddings.go`**
   - Get `llm_service_instance_id` from project (not from user request)
   - Update dimension validation
   - Update field names in responses

4. **`internal/handlers/admin.go`**
   - Replace `GetLLMsByProject` → `GetLLMInstancesByProject`
   - Update field names: `LLMServiceID` → `LLMServiceInstanceID`

5. **`internal/handlers/llm_definitions.go`** (NEW FILE)
   - Create handlers for definition CRUD operations
   - Routes for `/v1/llm-definitions/...`

### Phase 3: Tests (HIGH PRIORITY)

**All test files need updates:**

1. `internal/handlers/llm_services_test.go`
   - Update to use instance terminology
   - Add encryption tests
   - Add instance sharing tests

2. `internal/handlers/projects_test.go`
   - Test 1:1 relationship
   - Test instance validation

3. `internal/handlers/embeddings_test.go`
   - Update field names
   - Test instance retrieval from project

4. `internal/handlers/admin_test.go`
   - Update field names

### Phase 4: Initialization (MEDIUM PRIORITY)

**File:** `main.go` or new `internal/database/init.go`

**Add:**
- Check for `ENCRYPTION_KEY` environment variable
- Validate _system user exists
- Optional: Provide CLI command to migrate plaintext API keys to encrypted

### Phase 5: Documentation (MEDIUM PRIORITY)

**Files to update:**

1. **`README.md`**
   - Document `ENCRYPTION_KEY` requirement
   - Document new architecture (Definitions vs Instances)
   - Document instance sharing feature
   - Update example workflows

2. **`template.env`**
   - Add `ENCRYPTION_KEY=your-encryption-key-here`

3. **`docs/MIGRATION.md`** (NEW)
   - How to upgrade from old schema
   - How to handle existing API keys
   - Breaking changes (if any)

4. **`docs/API.md`** (NEW or UPDATE)
   - New endpoints for definitions
   - Instance sharing endpoints
   - Updated project creation (requires instance_id)

## 🔥 Known Issues (Build Errors)

Currently, the code does not compile due to:

1. **Missing model types** - Handlers reference `LLMService` but DB returns `LLMServiceInstance`
2. **Missing query methods** - Handlers call `UpsertLLM` but SQLC generated `UpsertLLMInstance`
3. **Field name mismatches** - `LLMServiceID` vs `InstanceID`, `LLMServiceHandle` vs `InstanceHandle`

**Fix order:**
1. Update models to include both Definition and Instance types
2. Update handlers one file at a time
3. Update tests after each handler file
4. Test build after each major file update

## 📋 Implementation Checklist

### Immediate Next Steps
- [ ] Create `LLMServiceDefinition` and `LLMServiceInstance` model types
- [ ] Update `internal/handlers/llm_services.go` to use new queries + encryption
- [ ] Update `internal/handlers/projects.go` for 1:1 instance relationship
- [ ] Update `internal/handlers/embeddings.go` to use project's instance
- [ ] Update `internal/handlers/admin.go` field names
- [ ] Fix compilation errors
- [ ] Run `go build` to verify

### Before Merging
- [ ] All tests passing
- [ ] Documentation updated
- [ ] ENCRYPTION_KEY documented in template.env
- [ ] Migration tested with sample database
- [ ] Manual testing of full workflow

## 🎯 Design Decisions Made

1. **Encryption:** Application-level encryption using Go's crypto package (not PostgreSQL's pgcrypto)
2. **Key Storage:** Environment variable `ENCRYPTION_KEY` (not file-based)
3. **Backwards Compatibility:** Keep existing API endpoints, map to new backend
4. **Default Instances:** Projects MUST specify an instance (no auto-creation)
5. **Sharing Model:** Read-only sharing (shared users can't modify instance)
6. **System Definitions:** Immutable templates owned by `_system` user

## 💡 Recommendations for Completion

1. **Incremental Approach:** Fix one handler file at a time, test after each
2. **Tests First:** Update test for a handler before fixing the handler
3. **Type Aliases:** Use type aliases for backwards compatibility where possible
4. **Deprecation Path:** Mark old patterns as deprecated but still functional
5. **Documentation:** Update docs immediately after code changes

## 📞 Questions for Review

1. Should we auto-create a default instance for new users?
   - Current: No, users must explicitly create instances
   
2. Should instance creation from definition be automatic or require API call?
   - Current: Requires explicit API call (not yet implemented)
   
3. Should we support migration of plaintext API keys to encrypted?
   - Recommended: Yes, provide CLI tool or admin endpoint

4. Should _system definitions be immutable?
   - Current: Yes, seeded in migration, not modifiable via API
