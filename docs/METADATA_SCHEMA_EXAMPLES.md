# Metadata Schema Validation Examples

This document provides practical examples of using metadata schema validation in the dhamps-vdb API.

## Example 1: Creating a Project with a Metadata Schema

```bash
# Create a project with a metadata schema
curl -X POST http://localhost:8080/v1/projects/alice \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "project_handle": "literary-texts",
    "description": "Literary texts with structured metadata",
    "metadataScheme": "{\"type\":\"object\",\"properties\":{\"author\":{\"type\":\"string\"},\"year\":{\"type\":\"integer\"},\"genre\":{\"type\":\"string\",\"enum\":[\"poetry\",\"prose\",\"drama\"]}},\"required\":[\"author\",\"year\"]}"
  }'
```

## Example 2: Uploading Embeddings with Valid Metadata

```bash
# Upload embeddings that conform to the schema
curl -X POST http://localhost:8080/v1/embeddings/alice/literary-texts \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "embeddings": [{
      "text_id": "kant-critique-pure-reason",
      "llm_service_handle": "openai-large",
      "text": "Critique of Pure Reason excerpt",
      "vector": [0.1, 0.2, 0.3, 0.4, 0.5],
      "vector_dim": 5,
      "metadata": {
        "author": "Immanuel Kant",
        "year": 1781,
        "genre": "prose"
      }
    }]
  }'
```

## Example 3: Validation Error - Missing Required Field

```bash
# This will fail because "year" is required but missing
curl -X POST http://localhost:8080/v1/embeddings/alice/literary-texts \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "embeddings": [{
      "text_id": "some-text",
      "llm_service_handle": "openai-large",
      "vector": [0.1, 0.2, 0.3, 0.4, 0.5],
      "vector_dim": 5,
      "metadata": {
        "author": "John Doe"
      }
    }]
  }'
```

Expected error response:
```json
{
  "$schema": "http://localhost:8080/schemas/ErrorModel.json",
  "title": "Bad Request",
  "status": 400,
  "detail": "metadata validation failed for text_id 'some-text': metadata validation failed:\n  - year is required"
}
```

## Example 4: Validation Error - Wrong Type

```bash
# This will fail because "year" should be an integer, not a string
curl -X POST http://localhost:8080/v1/embeddings/alice/literary-texts \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "embeddings": [{
      "text_id": "some-text",
      "llm_service_handle": "openai-large",
      "vector": [0.1, 0.2, 0.3, 0.4, 0.5],
      "vector_dim": 5,
      "metadata": {
        "author": "John Doe",
        "year": "1781"
      }
    }]
  }'
```

## Example 5: Dimension Validation Error

```bash
# This will fail because vector has 3 elements but declares 5 dimensions
curl -X POST http://localhost:8080/v1/embeddings/alice/literary-texts \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "embeddings": [{
      "text_id": "some-text",
      "llm_service_handle": "openai-large",
      "vector": [0.1, 0.2, 0.3],
      "vector_dim": 5,
      "metadata": {
        "author": "John Doe",
        "year": 1781
      }
    }]
  }'
```

Expected error response:
```json
{
  "$schema": "http://localhost:8080/schemas/ErrorModel.json",
  "title": "Bad Request",
  "status": 400,
  "detail": "dimension validation failed: vector length mismatch for text_id 'some-text': actual vector has 3 elements but vector_dim declares 5"
}
```

## Common Metadata Schema Patterns

### Simple Required Fields
```json
{
  "type": "object",
  "properties": {
    "author": {"type": "string"},
    "year": {"type": "integer"}
  },
  "required": ["author"]
}
```

### With Enums
```json
{
  "type": "object",
  "properties": {
    "genre": {
      "type": "string",
      "enum": ["poetry", "prose", "drama", "essay"]
    },
    "language": {
      "type": "string",
      "enum": ["en", "de", "fr", "es", "la"]
    }
  },
  "required": ["genre"]
}
```

### Nested Objects
```json
{
  "type": "object",
  "properties": {
    "author": {
      "type": "object",
      "properties": {
        "name": {"type": "string"},
        "birth_year": {"type": "integer"},
        "nationality": {"type": "string"}
      },
      "required": ["name"]
    }
  },
  "required": ["author"]
}
```

### Arrays
```json
{
  "type": "object",
  "properties": {
    "keywords": {
      "type": "array",
      "items": {"type": "string"},
      "minItems": 1,
      "maxItems": 10
    },
    "categories": {
      "type": "array",
      "items": {
        "type": "string",
        "enum": ["philosophy", "literature", "science", "history"]
      }
    }
  }
}
```

### With Constraints
```json
{
  "type": "object",
  "properties": {
    "title": {
      "type": "string",
      "minLength": 1,
      "maxLength": 200
    },
    "page_count": {
      "type": "integer",
      "minimum": 1
    },
    "rating": {
      "type": "number",
      "minimum": 0,
      "maximum": 5
    }
  }
}
```

## Tips

1. **Escape JSON for command line**: When passing JSON schemas in curl commands, make sure to properly escape quotes or use single quotes for the outer JSON.

2. **Use schema validators**: Before setting up your schema in the API, test it with online JSON Schema validators like [jsonschemavalidator.net](https://www.jsonschemavalidator.net/).

3. **Start simple**: Begin with a simple schema and add more constraints as needed. You can always update the schema using a PATCH or PUT request to the project endpoint.

4. **Optional metadata**: If you don't provide a `metadataScheme` when creating a project, metadata validation is skipped, and you can upload any JSON metadata.

5. **Schema updates**: When you update a project's metadata schema, existing embeddings are not revalidated. The schema only applies to new or updated embeddings.
