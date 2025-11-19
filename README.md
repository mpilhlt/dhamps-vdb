# dhamps-vdb
Vector Database for the DH at Max Planck Society initiative

[![Go Report Card](https://goreportcard.com/badge/github.com/mpilhlt/dhamps-vdb?style=flat-square)](https://goreportcard.com/report/github.com/mpilhlt/dhamps-vdb) [![Release](https://img.shields.io/github/v/release/mpilhlt/dhamps-vdb.svg?style=flat-square&include_prereleases)](https://github.com/mpilhlt/dhamps-vdb/releases/latest)

<!--
[![Go Reference](https://pkg.go.dev/badge/github.com/mpilhlt/dhamps-vdb.svg)](https://pkg.go.dev/github.com/mpilhlt/dhamps-vdb)
-->

## Introduction

This is an application serving an API to handle embeddings. It stores embeddings in a PostgreSQL backend and uses its vector support, but allows you to manage different users, projects, and LLM configurations via a simple Restful API.

The typical use case is as a component of a Retrieval Augmented Generation (RAG) workflow: You create embeddings for a collection of text snippets and upload them to this API. For each text snippet, you upload a text identifier, the embeddings vector and, optionally, metadata or the text itself. Then, you can

- `GET` the most similar texts for a text that is already in the database by specifying the text's identifier in an URL, and eventually later
- `POST` another text and get the most similar texts in the database

In both cases, the service returns a list of text identifiers that you can then use in your own processing, perhaps based on other means of providing the respective texts.

## Features

- OpenAPI documentation
- Supports different embeddings configurations (e.g. dimensions)
- Rights management (authentication via API token)

## Getting started

### Compiling

For **compiling**, you should run `go get ./... ; sqlc generate --no-remote ; go build -o build/dhamps-vdb main.go` (or in place of the last command you can also run it directly with `go run main.go`).

#### Run tests

Actual (mostly integration) tests are run like this:

```bash
$> systemctl --user start podman.socket
$> export DOCKER_HOST=unix://$XDG_RUNTIME_DIR/podman/podman.sock
$> go test -v ./...
```

These tests do not contact any separately installed/launched backend, and instead have a container managed by the testing itself (via [testcontainers](https://testcontainers.com/guides/getting-started-with-testcontainers-for-go/)).

### Running

For **running**, you need to set an admin key and a couple of other variables. Have a look in [options.go](./internal/models/options.go) for the full list and documentation.

You can specify them as command line options, with environment values or in an `.env` file. This last option is recommended because this way, the sensitive information will not be in the command history; but you have to take care of this file's security. For instance, it will/should not be synced to your versioning platform, it is thus listed in `.gitignore` by default.

If you authenticate with this key, you can create users by `POST`ing to the `/v1/users` endpoint. (The response to user creation will contain an API key for the new user. **Keep it safe, it is not stored anywhere and cannot be recovered.**) Then, either the admin or the user can create projects, llm services and finally post embeddings and get similar elements.

When launching, the application checks and migrates the database schema to the appropriate version if possible. It presupposes however, that a suitable database and user (with appropriate privileges) have been created beforehand. SQL commands to prepare the database are listed below. For other ways of running, e.g. for tests or with a container instead of an external database, see below as well.

#### Run with local container

A local container with a pg_vector-enabled postgresql can be run like this:

```bash
$> podman run -p 8888:5432 -e POSTGRES_PASSWORD=password pgvector/pgvector:0.7.4-pg16
```

But be aware that the filesystem is not persisted if you run it like this. That means that when you stop and remove the container, you will have to repeat the following database setup when you run it again later on. (And of course any data you may have saved inside the container is lost, too.)

You can connect to it from a second terminal like so:

```bash
$> psql -p 8888 -h localhost -U postgres -d postgres
```

And then set up the database like this:

```sql
postgres=# CREATE DATABASE my_vectors;
postgres=# CREATE USER my_user WITH PASSWORD 'my-password';
postgres=# GRANT ALL PRIVILEGES ON DATABASE "my_vectors" to my_user;
postgres=# \c my_vectors
postgres=# GRANT ALL ON SCHEMA public TO my_user;
postgres=# CREATE EXTENSION IF NOT EXISTS vector;
```

### Client Authentication

Clients should communicate the API key in the `Authorization` header with a `Bearer` prefix, e.g. `Bearer 024v2013621509245f2e24`. Most operations can only be done by the (admin or the) owner of the resource in question. For projects and their embeddings, you can define other user accounts that should be authorized as readers, too. 

**Public Access**: Projects can be made publicly accessible (allowing unauthenticated read access to embeddings and similars) by including `"*"` in the `authorizedReaders` array when creating or updating the project. See [docs/PUBLIC_ACCESS.md](./docs/PUBLIC_ACCESS.md) for details.

## API documentation

### API Versioning

We are at `v1`. The first path component to all the endpoints (except for the OpenAPI file) is the version number, e.g. `POST https://<hostname>/v1/embeddings/<user>/<project>`.

### Endpoints

In the following table, the version number is skipped for readibility reasons. Nevertheless, it is the first component of all these endpoints.

For a more detailed, and always up-to-date documentation of the endpoints, including query parameters, return values and data schemes, see the automatically generated live OpenAPI document at `/openapi.yaml` or the browsable version at `/docs`.

| Endpoint | Method | Description | Allowed Users |
|----------|--------|-------------|---------------|
| /admin/footgun | GET | Reset Database: Remove all records from database and reset serials/counters | admin |
| /users | GET  | Get all users (list of handles) registered with the Db | admin |
| /users | POST | Register a new user with the Db | admin |
| /users/\<username\> | GET | Get information about user \<username\> | admin, \<username\> |
| /users/\<username\> | PUT | Register a new user with the Db | admin |
| /users/\<username\> | DELETE | Delete a user and all their projects/llm services from the Db | admin, \<username\> |
| /projects/\<username\> | GET  | Get all projects (objects) for user \<username\> | admin, \<username\> |
| /projects/\<username\> | POST | Register a new project for user \<username\> | admin, \<username\> |
| /projects/\<username\>/\<projectname\> | GET | Get project information for \<username\>'s project \<projectname\> | admin, \<username\>, authorized readers |
| /projects/\<username\>/\<projectname\> | PUT | Register a new project calles \<projectname\> for user \<username\> | admin, \<username\> |
| /projects/\<username\>/\<projectname\> | DELETE | Delete \<username\>'s project \<projectname\> | admin, \<username\> |
| /llm-services/\<username\> | GET  | Get all LLM services (objects) for user \<username\> | admin, \<username\> |
| /llm-services/\<username\> | POST | Register a new LLM service for user \<username\> | admin, \<username\> |
| /llm-services/\<username\>/<llm_servicename> | GET | Get information about LLM service <llm_servicename> of user \<username\> | admin, \<username\> |
| /llm-services/\<username\>/<llm_servicename> | PUT | Register a new LLM service called <llm_servicename> for user \<username\> | admin, \<username\> |
| /llm-services/\<username\>/<llm_servicename> | DELETE | Delete \<username\>'s LLM service <llm_servicename> | admin, \<username\> |
| /api-standards | GET  | Get all defined API standards* | public |
| /api-standards | POST | Register a new API standard* | admin |
| /api-standards/\<standardname\> | GET | Get information about API standard* \<standardname\> | public |
| /api-standards/\<standardname\> | PUT | Register a new API standard* \<standardname\> | admin |
| /api-standards/\<standardname\> | DELETE | Delete API standard* \<standardname\> | admin |
| /embeddings/\<username\>/\<projectname\> | GET  | Get all embeddings for \<username\>'s project \<projectname\> (use `limit` and `offset` for paging) | admin, \<username\>, authorized readers |
| /embeddings/\<username\>/\<projectname\> | POST | Register a new record with an embeddings vector for \<username\>'s project \<projectname\> | admin, \<username\> |
| /embeddings/\<username\>/\<projectname\> | DELETE | Delete ***all*** embeddings for \<username\>'s project \<projectname\> | admin, \<username\> |
| /embeddings/\<username\>/\<projectname\>/\<identifier\> | GET | Get embeddings and other information about text \<identifier\> from \<username\>'s project \<projectname\> | admin, \<username\>, authorized readers |
| /embeddings/\<username\>/\<projectname\>/\<identifier\> | DELETE | Delete record \<identifier\> from \<username\>'s project \<projectname\> | admin, \<username\> |
| /similars/\<username\>/\<projectname\>/\<identifier\> | GET | Get a list of identifiers for records that are similar to the text \<identifier\> in \<username\>'s project \<projectname\> | admin, \<username\>, authorized readers |

\* API standards are definitions of how to access an LLM Service: API endpoints, authentication mechanism etc. They are referred to from LLM Service definitions. When LLM Processing will be attempted, this is what will be implemented. Examples are the Cohere Embed API, Version 2, as documented in <https://docs.cohere.com/reference/embed>, or the OpenAI Embeddings API, Version 1, as documented in <https://platform.openai.com/docs/api-reference/embeddings>. You can find these examples in the [valid_api_standard\*.json](./testdata/) files in the `testdata` directory.

## Code creation and structure

This API is programmed in go and uses the [huma](https://huma.rocks/) framework with go's stock `http.ServeMux()` routing.

Some initial code and some later bugfixes have been developed in dialogue with [ChatGPT](./docs/ChatGPT.md). After manual inspection and correction, this is the project structure:

```default
dhamps-vdb/
├── .env           // This is not distributed because it's in .gitignore
├── .gitignore
├── .repopackignore
├── LICENSE
├── README.md
├── go.mod
├── go.sum
├── main.go
├── repopack-output.xml
├── repopack.config.json
├── sqlc.yaml
├── template.env
├── api/
│   └── openapi.yml          // OpenAPI spec file, not up to date
├── docs/
│   └── ChatGPT.md           // Code as suggested by ChatGPT (GPT4 turbo and GPT4o) on 2024-06-09
├── internal/
│   ├── auth/
│   │   └── authenticate.go
│   ├── database/
│   │   ├── migrations/
│   │   │   ├── 001_create_initial_scheme.sql
│   │   │   ├── 002_create_emb_index.sql
│   │   │   ├── tern.conf        // This is not distributed because it's in .gitignore
│   │   │   └── tern.conf.tpl
│   │   ├── queries/
│   │   │   └── queries.sql
│   │   ├── database.go
│   │   ├── db.go                // This is auto-generated by sqlc
│   │   ├── migrations.go
│   │   ├── models.go            // This is auto-generated by sqlc
│   │   └── queries.sql.go       // This is auto-generated by sqlc
│   ├── handlers/
│   │   ├── admin.go
│   │   ├── admin_test.go
│   │   ├── api_standards.go
│   │   ├── api_standards_test.go
│   │   ├── embeddings.go
│   │   ├── embeddings_test.go
│   │   ├── handlers.go
│   │   ├── handlers_test.go
│   │   ├── llm_processes.go
│   │   ├── llm_services.go
│   │   ├── llm_services_test.go
│   │   ├── projects.go
│   │   ├── projects_test.go
│   │   ├── similars.go
│   │   ├── users.go
│   │   └── users_test.go
│   └── models/
│       ├── admin.go
│       ├── api_standards.go
│       ├── embeddings.go
│       ├── llm_processes.go
│       ├── llm_services.go
│       ├── options.go
│       ├── projects.go
│       ├── similars.go
│       └── users.go
├── testdata/
│   ├── postgres/
│   │   ├── enable-vector.sql
│   │   └── users.yml
│   ├── invalid_api_standard.json
│   ├── invalid_embeddings.json
│   ├── ...
│   ├── valid_api_standard_cohere_v2.json
│   ├── valid_api_standard_ollama.json
│   ├── valid_api_standard_openai_v1.json
│   ├── valid_embeddings.json
│   ├── valid_llm_service_cohere-multilingual-3.json
│   ├── valid_llm_service_openai-large-full.json
│   ├── ...
│   └── valid_user.json
└── web/                      // Web resources for the html response (in the future)
```

## Roadmap

- [x] Tests
  - [x] When testing, check cleanup by adding a new query/function to see if all tables are empty
  - [ ] Make sure pagination is supported consistently
  - [ ] Make sure input is validated consistently
- [x] Catch POST to existing resources
- [x] User authentication & restrictions on some API calls
- [x] API versioning
- [x] better **options** handling (<https://huma.rocks/features/cli/>)
- [x] handle **metadata**
  - [ ] Validation with metadata schema
- [x] Allow to filter similar passages by metadata field (so as to exclude e.g. documents from the same author)
  - [ ] Add documentation (the GET query parameters are called `metadata_path` and `metadata_value` as in: `https://xy.org/vdb-api/v1/similars/sal/sal-openai-large/https%3A%2F%2Fid.myproject.net%2Ftexts%2FW0011%3A1.3.1.3.1?threshold=0.7&limit=5&metadata_path=author_id&metadata_value=A0083`)
- [ ] Implement and make consequent use of **max_idle** (5), **max_concurr** (5), **timeouts**, and **cancellations**
- [ ] **Concurrency** (leaky bucket approach) and **Rate limiting** (redis, sliding window, implement headers)
- [ ] Use **transactions** (most importantly, when an action requires several queries, e.g. projects being added and then linked to several read-authorized users)
- [ ] Use PATCH method to change existing resources
- [x] Add mechanism to allow anonymous/public reading access to embeddings (via `"*"` in `authorizedReaders`)
- [ ] **Dockerization**
- [ ] **Batch mode**
- [ ] **Link or unlink** users/LLMs as standalone operations
- [ ] **Transfer** of projects from one owner to another as new operation
- [ ] Fix automatically generated documentation
- [ ] Proper **logging** with `--verbose` and `--quiet` modes
- [ ] Caching
- [ ] HTML UI?
- [ ] LLM handling processing (receive text and send it to an llm service on the user's behalf, then store the results)
  - [ ] allow API keys for services to be read from env variables (on the server, but still maybe useful)
  - [ ] calls to LLM services

## License

[MIT License](./LICENSE)

## Versions

- 2024-12-10 **v0.0.1**: Initial public release (still work in progress) of API v1
