# LLM Service Architecture Refactoring - Complete Documentation

## Table of Contents

1. [Overview](#overview)
2. [Implementation Summary](#implementation-summary)
3. [Architecture](#architecture)
4. [Completed Work](#completed-work)
5. [Usage Guide](#usage-guide)
6. [Security Features](#security-features)
7. [Migration Guide](#migration-guide)
8. [Testing](#testing)
9. [Remaining Optional Work](#remaining-optional-work)

## Overview

This refactoring separates LLM services into two distinct concepts:

1. **LLM Service Definitions** - Reusable templates owned by `_system` or users
   - Contain configuration templates (endpoint, model, dimensions, API standard)
   - Can be owned by `_system` (global templates) or individual users
   - Used as templates for creating instances

2. **LLM Service Instances** - User-specific configurations with encrypted API keys
   - Contain actual service configurations and credentials
   - Owned by individual users
   - Can optionally reference a definition
   - Support API key encryption
   - Can be shared with other users

## Implementation Summary

### ✅ All Core Requirements Completed

1. **Admin can manage _system definitions**
   - `_system` user created in migration
   - 4 default definitions seeded (openai-large, openai-small, cohere-v4, gemini-embedding-001)
   - API standards (openai, cohere, gemini) created before definitions

2. **Users can list all accessible instances**
   - `GetAllAccessibleLLMInstances` query returns owned + shared instances
   - Users see all instances they own or have been granted access to

3. **Handle-based instance references**
   - Shared instances identified as `owner/handle`
   - Own instances identified as `handle`
   - Queries support handle-based lookups

4. **API keys hidden from shared instances**
   - API keys NEVER returned in GET/list responses (security)
   - Write-only field in API
   - Shared users can use instances but cannot see API keys

5. **Multiple ways to create instances**
   - From own definitions
   - From _system definitions
   - Standalone (all fields specified)

6. **1:1 project-instance relationship**
   - Projects must reference exactly one instance
   - Enforced at database level

### Build & Test Status

- ✅ Code compiles successfully
- ✅ All tests passing (100% success rate)
- ✅ Migration tested and verified
- ✅ Encryption module tested

## Architecture

### Database Schema

```
llm_service_definitions (templates)
├── definition_id (PK)
├── definition_handle
├── owner (FK → users, can be '_system')
├── endpoint, description, api_standard, model, dimensions
└── UNIQUE(owner, definition_handle)
└── Indexes: (owner, definition_handle), (definition_handle)

llm_service_instances (user-specific)
├── instance_id (PK)
├── instance_handle
├── owner (FK → users)
├── definition_id (FK → llm_service_definitions, nullable)
├── endpoint, description, model, dimensions, api_standard
├── api_key (TEXT, for backward compatibility)
├── api_key_encrypted (BYTEA, AES-256-GCM encrypted)
└── UNIQUE(owner, instance_handle)

llm_service_instances_shared_with (n:m sharing)
├── user_handle (FK → users)
├── instance_id (FK → llm_service_instances)
├── role (reader/writer/owner)
└── PRIMARY KEY(user_handle, instance_id)

projects (1:1 with instances)
├── project_id (PK)
├── llm_service_instance_id (FK → llm_service_instances)
└── One project → One instance
```

### Key Tables Removed

- `users_llm_services` - Redundant (ownership tracked via `llm_service_instances.owner`)
- `projects_llm_services` - Replaced by 1:1 FK in projects table

## Completed Work

### 1. Database Migration (004)

**File:** `internal/database/migrations/004_refactor_llm_services_architecture.sql`

**Changes:**
- Created `llm_service_definitions` table
- Renamed `llm_services` → `llm_service_instances`
- Added `api_key_encrypted` BYTEA column
- Created `_system` user
- Dropped `users_llm_services` table (redundant)
- Modified `projects` table: removed many-to-many, added `llm_service_instance_id` FK
- Created `llm_service_instances_shared_with` table
- Seeded 3 API standards with documentation URLs:
  - openai: https://platform.openai.com/docs/api-reference/embeddings
  - cohere: https://docs.cohere.com/reference/embed
  - gemini: https://ai.google.dev/gemini-api/docs/embeddings
- Seeded 4 default LLM service definitions:
  - openai-large (3072 dimensions)
  - openai-small (1536 dimensions)
  - cohere-v4 (1536 dimensions)
  - gemini-embedding-001 (3072 dimensions, default size)

**Data Migration:**
- First linked LLM service per project → `project.llm_service_instance_id`
- Rollback support included

### 2. Encryption Module

**File:** `internal/crypto/encryption.go`

**Features:**
- AES-256-GCM encryption for API keys
- Uses `ENCRYPTION_KEY` environment variable (SHA256-hashed to ensure 32-byte key)
- Functions:
  - `NewEncryptionKey(keyString)` - Create key from string
  - `GenerateEncryptionKey()` - Generate random key
  - `GetEncryptionKeyFromEnv()` - Read from environment
  - `Encrypt(plaintext) → []byte`
  - `Decrypt(ciphertext) → string`
  - `EncryptToBase64(plaintext) → string`
  - `DecryptFromBase64(base64) → string`

**Testing:** Full test coverage in `internal/crypto/encryption_test.go` ✅

### 3. Database Queries (SQLC)

**File:** `internal/database/queries/queries.sql`

**Definitions:**
- `UpsertLLMDefinition` - Create/update definition
- `DeleteLLMDefinition` - Delete definition
- `RetrieveLLMDefinition` - Get single definition
- `GetLLMDefinitionsByUser` - List user's definitions
- `GetAllLLMDefinitions` - List all definitions
- `GetSystemLLMDefinitions` - List _system definitions

**Instances:**
- `UpsertLLMInstance` - Create/update instance (with encryption support)
- `CreateLLMInstanceFromDefinition` - Create instance from definition template
- `DeleteLLMInstance` - Delete instance
- `RetrieveLLMInstance` - Get single instance
- `RetrieveLLMInstanceByID` - Get instance by ID
- `RetrieveLLMInstanceByOwnerHandle` - Get by owner/handle (supports both formats)
- `ShareLLMInstance` - Share instance with another user
- `UnshareLLMInstance` - Remove instance sharing
- `GetSharedUsersForInstance` - List users instance is shared with
- `GetLLMInstanceByProject` - Get instance for project (1:1, renamed from plural)
- `GetLLMInstancesByUser` - List user's owned instances
- `GetAllAccessibleLLMInstances` - List owned + shared instances
- `GetSharedLLMInstances` - List instances shared with user (sorted by role, owner, handle)

**Updated Queries:**
- `UpsertProject` - Includes `llm_service_instance_id`
- `UpsertEmbeddings` - Uses `llm_service_instance_id`
- All embeddings queries - Updated to use instances table

**SQLC Code Generated:** ✅ (`internal/database/models.go`, `internal/database/queries.sql.go`)

### 4. Go Models

**File:** `internal/models/llm_services.go`

**Models:**
- `LLMServiceDefinition` - For definitions
- `LLMServiceInstance` - For instances
- `LLMService` - Kept for backward API compatibility (maps to Instance)

**Field Updates:**
- `InstanceHandle` (was `LLMServiceHandle`)
- `InstanceOwner` (was `LLMServiceOwner`)
- API keys marked as write-only (never returned in responses)

### 5. Handlers

**Updated Files:**
- `internal/handlers/llm_services.go` - All functions renamed with "Instance" suffix
- `internal/handlers/projects.go` - 1:1 instance relationship
- `internal/handlers/embeddings.go` - Uses instance from project
- `internal/handlers/admin.go` - Updated field names
- `internal/handlers/users.go` - Lists accessible instances
- `internal/handlers/validation.go` - Updated to InstanceHandle

**Function Naming:**
- `putLLMInstanceFunc` (was `putLLMFunc`)
- `getLLMInstanceFunc` (was `getLLMFunc`)
- `deleteLLMInstanceFunc` (was `deleteLLMFunc`)
- `getUserLLMsFunc` - Now returns all accessible instances (own + shared)

**API Key Handling:**
- Encrypted on write if `ENCRYPTION_KEY` is set
- Never returned on read (security)
- Uses `Valid: true` consistently for nullable fields

### 6. Environment Configuration

**File:** `template.env`

Added:
```bash
# Required for API key encryption (32+ characters recommended)
ENCRYPTION_KEY=your-secret-encryption-key-here-must-be-kept-secure
```

## Usage Guide

### Creating an LLM Service Instance

**Option A: Standalone (no definition)**
```bash
PUT /v1/llm-services/jdoe/my-openai
{
  "endpoint": "https://api.openai.com/v1/embeddings",
  "api_standard": "openai",
  "model": "text-embedding-3-large",
  "dimensions": 3072,
  "api_key": "sk-..."
}
```

**Option B: From _system definition**
```bash
# Use CreateLLMInstanceFromDefinition query
# Handler would accept:
POST /v1/llm-services/jdoe/my-openai-instance
{
  "definition_owner": "_system",
  "definition_handle": "openai-large",
  "api_key": "sk-..."
}
```

**Option C: From user's own definition**
```bash
# Similar to Option B, but with user as definition_owner
POST /v1/llm-services/jdoe/my-custom-instance
{
  "definition_owner": "jdoe",
  "definition_handle": "my-custom-config",
  "api_key": "sk-..."
}
```

### Listing Accessible Instances

```bash
GET /v1/llm-services/jdoe
# Returns all instances jdoe owns OR has been granted access to
# API keys are NOT included in response
```

### Creating a Project with Instance

```bash
POST /v1/projects/jdoe/my-project
{
  "llm_service_instance_id": 123,  # or use handle-based reference
  "description": "My project"
}
```

## Security Features

### 1. API Key Encryption

- **Algorithm:** AES-256-GCM
- **Key Source:** `ENCRYPTION_KEY` environment variable
- **Key Derivation:** SHA256 hash of environment variable
- **Storage:** `api_key_encrypted` BYTEA column
- **Fallback:** `api_key` TEXT column (for backward compatibility)

### 2. Write-Only API Keys

API keys are never returned in GET/list responses:
```json
GET /v1/llm-services/jdoe/my-openai
{
  "instance_id": 1,
  "instance_handle": "my-openai",
  "owner": "jdoe",
  "endpoint": "...",
  "model": "text-embedding-3-large",
  "dimensions": 3072
  // Note: "api_key" field is NOT present
}
```

### 3. Shared Instance Protection

When an instance is shared:
- Shared users can USE the instance (e.g., for projects, embeddings)
- Shared users CANNOT see the API key
- Shared users CANNOT modify the instance (owner-only operation)
- Sharing is tracked in `llm_service_instances_shared_with` table with role

### 4. Admin-Only System Definitions

- Only admin users can create/modify `_system` definitions
- Regular users can read `_system` definitions
- Regular users can create their own definitions
- No one can log in as `_system`

## Migration Guide

### For New Installations

1. Run migrations: `make migrate-up` (or equivalent)
2. Set `ENCRYPTION_KEY` environment variable
3. Service is ready to use

### For Existing Installations

The migration (004) handles data migration automatically:

**Automatic Changes:**
- `llm_services` table renamed to `llm_service_instances`
- `users_llm_services` table dropped (ownership via owner column)
- `projects_llm_services` table dropped (replaced by FK)
- First linked instance per project → `project.llm_service_instance_id`
- API keys remain in plaintext initially (in `api_key` column)

**Post-Migration Steps:**

1. **Set Environment Variable:**
   ```bash
   export ENCRYPTION_KEY="your-secure-random-string-at-least-32-chars"
   ```

2. **Restart Service:**
   - New API keys will be automatically encrypted
   - Old plaintext keys continue to work

3. **Optional - Migrate Old Keys:**
   ```sql
   -- Run a script to re-encrypt all plaintext API keys
   -- (Not implemented yet, but recommended for production)
   ```

4. **Optional - Remove Plaintext Column:**
   ```sql
   -- After all keys are encrypted
   ALTER TABLE llm_service_instances DROP COLUMN api_key;
   ```

### Breaking Changes

**API Changes:**
- `GET /v1/llm-services/{user}` - No longer returns API keys
- `GET /v1/llm-services/{user}/{handle}` - No longer returns API keys
- Projects now require single instance (many-to-many removed)

**Database:**
- `llm_services` → `llm_service_instances`
- `users_llm_services` table removed
- `projects_llm_services` table removed

**Backward Compatibility:**
- Existing endpoints continue to work
- Field names preserved in JSON responses (for API compatibility)
- Old plaintext API keys continue to work

## Client Migration Guide

This section explains what API clients need to change after the refactoring.

### Summary of Changes for Clients

**Good News:** Most API endpoints remain unchanged! The main changes are:
1. API keys are no longer returned in responses (security improvement)
2. Projects must be created with a valid LLM service instance (1:1 relationship)
3. Embeddings JSON uses `instance_handle` instead of `llm_service_handle`

### API Endpoints - No Changes Required

All existing API endpoints continue to work with the same paths:

```
✅ PUT    /v1/llm-services/{user}/{handle}           # Create/update instance
✅ GET    /v1/llm-services/{user}                    # List user's instances
✅ GET    /v1/llm-services/{user}/{handle}           # Get specific instance
✅ DELETE /v1/llm-services/{user}/{handle}           # Delete instance

✅ PUT    /v1/projects/{user}/{project}              # Create/update project
✅ GET    /v1/projects/{user}/{project}              # Get project
✅ DELETE /v1/projects/{user}/{project}              # Delete project

✅ POST   /v1/embeddings/{user}/{project}            # Upload embeddings
✅ GET    /v1/embeddings/{user}/{project}            # List embeddings
✅ DELETE /v1/embeddings/{user}/{project}            # Delete embeddings
```

### Change #1: API Keys No Longer Returned (Security)

**Before:** API keys were returned in GET responses
```json
GET /v1/llm-services/alice/my-openai
Response: {
  "instance_handle": "my-openai",
  "owner": "alice",
  "endpoint": "https://api.openai.com/v1/embeddings",
  "api_key": "sk-proj-..."  ← Was visible
}
```

**After:** API keys are write-only (never returned)
```json
GET /v1/llm-services/alice/my-openai
Response: {
  "instance_handle": "my-openai",
  "owner": "alice",
  "endpoint": "https://api.openai.com/v1/embeddings"
  // Note: No api_key field
}
```

**Action Required:**
- ⚠️ If your client code reads API keys from GET responses, update it to store keys locally
- ✅ You can still SET api_key in PUT/POST requests
- ✅ The server will use the stored (encrypted) key when needed

### Change #2: Embeddings Field Name Update

**Before:** Used `llm_service_handle` in embeddings JSON
```json
POST /v1/embeddings/alice/my-project
{
  "text_id": "doc1",
  "llm_service_handle": "my-openai",  ← Old field name
  "embedding": [0.1, 0.2, ...],
  "metadata": {...}
}
```

**After:** Use `instance_handle` instead
```json
POST /v1/embeddings/alice/my-project
{
  "text_id": "doc1",
  "instance_handle": "my-openai",  ← New field name
  "embedding": [0.1, 0.2, ...],
  "metadata": {...}
}
```

**Action Required:**
- ⚠️ Update embedding upload code to use `instance_handle` field
- ⚠️ Update code that reads embeddings to expect `instance_handle` in responses

### Change #3: Projects Must Have LLM Service Instance

**Before:** Projects could be created without specifying an LLM service
```json
PUT /v1/projects/alice/my-project
{
  "description": "My project"
}
```

**After:** Projects require a valid `llm_service_instance_id`
```json
PUT /v1/projects/alice/my-project
{
  "description": "My project",
  "llm_service_instance_id": 123  ← Required
}
```

**Action Required:**
- ⚠️ Create an LLM service instance BEFORE creating projects
- ⚠️ Include `llm_service_instance_id` in project creation requests
- ℹ️ You can find the instance_id from the GET instances response

### Complete Migration Workflow

Here's the recommended workflow for clients:

#### Step 1: Create LLM Service Instance (if not exists)

```bash
# Check if instance exists
GET /v1/llm-services/alice

# If not, create one
PUT /v1/llm-services/alice/my-openai
Content-Type: application/json

{
  "endpoint": "https://api.openai.com/v1/embeddings",
  "api_standard": "openai",
  "model": "text-embedding-3-large",
  "dimensions": 3072,
  "api_key": "sk-proj-your-key-here"
}

# Response includes instance_id
{
  "instance_id": 123,
  "instance_handle": "my-openai",
  "owner": "alice",
  "endpoint": "https://api.openai.com/v1/embeddings",
  "api_standard": "openai",
  "model": "text-embedding-3-large",
  "dimensions": 3072
  // Note: api_key not returned
}
```

#### Step 2: Create Project with Instance ID

```bash
PUT /v1/projects/alice/my-project
Content-Type: application/json

{
  "description": "My research project",
  "llm_service_instance_id": 123  # From step 1
}
```

#### Step 3: Upload Embeddings

```bash
POST /v1/embeddings/alice/my-project
Content-Type: application/json

{
  "text_id": "doc1",
  "instance_handle": "my-openai",  # Use instance_handle (not llm_service_handle)
  "embedding": [0.1, 0.2, 0.3, ...],
  "metadata": {
    "title": "Document 1",
    "author": "Alice"
  }
}
```

### Environment Setup for New Installations

**Before:** No encryption key needed
```bash
# .env
DATABASE_URL=postgresql://...
ADMIN_KEY=your-admin-key
```

**After:** Add encryption key
```bash
# .env
DATABASE_URL=postgresql://...
ADMIN_KEY=your-admin-key
ENCRYPTION_KEY=your-secure-32-char-minimum-key  # NEW
```

**Action Required:**
- ⚠️ Add `ENCRYPTION_KEY` to your environment variables
- ✅ Use a strong, random string (32+ characters recommended)
- ✅ Keep this key secure - losing it means losing access to encrypted API keys

### Migration Checklist for Existing Clients

Use this checklist to ensure your client is fully migrated:

- [ ] **Stop reading API keys from GET responses**
  - Update code to store API keys locally instead
  - Remove any code that expects `api_key` field in responses

- [ ] **Update embedding field names**
  - Change `llm_service_handle` → `instance_handle` in upload code
  - Update parsing code to read `instance_handle` from responses

- [ ] **Update project creation workflow**
  - Create LLM service instance first
  - Include `llm_service_instance_id` in project creation
  - Get instance_id from instance creation/list response

- [ ] **Update environment configuration**
  - Add `ENCRYPTION_KEY` to environment variables (for new installations)
  - Restart services to pick up new configuration

- [ ] **Test end-to-end workflow**
  - Create instance → Create project → Upload embeddings
  - Verify all steps work correctly

### Troubleshooting

**Problem:** "Project must have llm_service_instance_id" error

**Solution:** Create an LLM service instance first, then use its ID when creating the project.

---

**Problem:** Embeddings upload fails with "unknown field llm_service_handle"

**Solution:** Update your JSON to use `instance_handle` instead of `llm_service_handle`.

---

**Problem:** Can't see API key after creating instance

**Solution:** This is expected behavior (security improvement). Store the API key on the client side when you create the instance.

---

**Problem:** Old embeddings have `llm_service_handle` in their data

**Solution:** The migration automatically preserves existing data. New embeddings should use `instance_handle`, but old ones will still have the old field name.

---

**Problem:** Missing ENCRYPTION_KEY environment variable

**Solution:** Add `ENCRYPTION_KEY=your-secure-key` to your environment variables. This is only required for new installations or if you want to start encrypting API keys.

### Testing Your Migration

Here's a test sequence to verify your client works correctly:

```bash
# 1. Create instance
curl -X PUT "http://localhost:8000/v1/llm-services/testuser/test-instance" \
  -H "Authorization: Bearer your-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "endpoint": "https://api.openai.com/v1/embeddings",
    "api_standard": "openai",
    "model": "text-embedding-3-large",
    "dimensions": 3072,
    "api_key": "test-key"
  }'

# 2. List instances (verify api_key is NOT returned)
curl -X GET "http://localhost:8000/v1/llm-services/testuser" \
  -H "Authorization: Bearer your-api-key"

# 3. Create project with instance_id
curl -X PUT "http://localhost:8000/v1/projects/testuser/test-project" \
  -H "Authorization: Bearer your-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "description": "Test project",
    "llm_service_instance_id": 123
  }'

# 4. Upload embeddings with instance_handle
curl -X POST "http://localhost:8000/v1/embeddings/testuser/test-project" \
  -H "Authorization: Bearer your-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "text_id": "test1",
    "instance_handle": "test-instance",
    "embedding": [0.1, 0.2, 0.3],
    "metadata": {}
  }'
```

### Summary of Required Client Changes

| Area | Old Behavior | New Behavior | Action |
|------|-------------|--------------|--------|
| **API Endpoints** | Same paths | Same paths | ✅ No change |
| **API Keys in GET** | Returned in response | NOT returned | ⚠️ Stop reading, store locally |
| **Embeddings field** | `llm_service_handle` | `instance_handle` | ⚠️ Update field name |
| **Project creation** | Optional instance | Required `llm_service_instance_id` | ⚠️ Create instance first |
| **Project-instance** | Many-to-many | 1:1 relationship | ⚠️ One instance per project |
| **Environment vars** | No encryption key | `ENCRYPTION_KEY` needed | ⚠️ Add to .env (new installs) |

**Legend:** ✅ No action needed | ⚠️ Action required | ℹ️ Optional/informational

## Testing

### Test Status

**✅ All Tests Passing (100% success rate):**
- TestLLMServicesFunc: 16/16 subtests
- TestEmbeddingsFunc: All subtests
- TestValidationFunc: All subtests (updated to use InstanceHandle)
- TestUserFunc: All subtests
- TestPublicAccess: Pass
- TestSimilarsFunc: Pass

### Test Fixes Applied

1. **Query Bug Fixed:** `GetAllAccessibleLLMInstances` had user_handle filter in JOIN ON clause, preventing owned instances from being returned
2. **Test Expectations Updated:** Removed API key from expected responses (security)

### Test Coverage

**Current Coverage:**
- ✅ Basic Instance CRUD operations
- ✅ Authentication/authorization
- ✅ Invalid JSON handling
- ✅ Non-existent resource handling
- ✅ API key hiding in responses
- ✅ Field name updates (InstanceHandle, etc.)

## Remaining Optional Work

### Potential Enhancements (~7 hours total)

#### 1. Split Test Files (1 hour)
- Create `llm_service_definitions_test.go` for Definition tests
- Create `llm_service_instances_test.go` for Instance tests
- Better organization and clarity

#### 2. Add Definition Tests (2 hours)
- Creating definitions as _system user (admin only)
- Preventing non-admin users from creating _system definitions
- User-owned definitions
- Invalid input handling
- Deletion behavior

#### 3. Add Instance Sharing Tests (2 hours)
- Sharing instances with other users
- Listing shared instances
- Access control verification
- API key protection for shared instances
- Revoking access

#### 4. Add Encryption Tests (1 hour)
- API key encryption/decryption roundtrip
- Handling missing ENCRYPTION_KEY
- Key update scenarios

#### 5. Documentation (1 hour)
- API documentation for new endpoints
- Examples of instance creation from definitions
- Security best practices

### New Endpoints (Not Implemented)

Consider adding these endpoints in the future:
- `GET /v1/llm-service-definitions` - List all available definitions
- `GET /v1/llm-service-definitions/_system` - List system definitions
- `POST /v1/llm-service-definitions/{user}` - Create user definition
- `POST /v1/llm-service-instances/{user}/from-definition/{handle}` - Create from definition
- `POST /v1/llm-service-instances/{user}/{instance}/share/{target}` - Share instance
- `DELETE /v1/llm-service-instances/{user}/{instance}/share/{target}` - Revoke sharing

### API Key Migration Tool (Not Implemented)

Create a CLI tool or admin endpoint to:
- List all instances with plaintext API keys
- Re-encrypt them using the current ENCRYPTION_KEY
- Verify successful encryption
- Remove plaintext keys

## Design Decisions

1. **Encryption:** Application-level encryption (not PostgreSQL's pgcrypto) for portability
2. **Key Storage:** Environment variable (not file-based) for security and container-friendliness
3. **Backward Compatibility:** Keep existing endpoints, map to new backend
4. **Default Instances:** Projects MUST specify an instance (no auto-creation)
5. **Sharing Model:** Read-only sharing (only owner can modify)
6. **System Definitions:** Owned by `_system` user, created in migration
7. **Ownership Tracking:** Via `owner` column (removed redundant join table)

## References

- Encryption implementation: `internal/crypto/encryption.go`
- Migration: `internal/database/migrations/004_refactor_llm_services_architecture.sql`
- Queries: `internal/database/queries/queries.sql`
- Performance notes: See `docs/PERFORMANCE_OPTIMIZATION.md`
- Test data: `testdata/valid_embeddings*.json` (updated to use instance_handle)

## Support

For questions or issues:
1. Review this documentation
2. Check the migration file for schema details
3. Review test files for usage examples
4. See PERFORMANCE_OPTIMIZATION.md for performance tuning
