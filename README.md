# dhamps-vdb
Vector Database for the DH at Max Planck Society initiative

[![Go Report Card](https://goreportcard.com/badge/github.com/mpilhlt/dhamps-vdb?style=flat-square)](https://goreportcard.com/report/github.com/mpilhlt/dhamps-vdb)

[![Go Reference](https://pkg.go.dev/badge/github.com/mpilhlt/dhamps-vdb.svg)](https://pkg.go.dev/github.com/mpilhlt/dhamps-vdb)

[![Release](https://img.shields.io/github/release/golang-standards/project-layout.svg?style=flat-square)](https://github.com/golang-standards/project-layout/releases/latest)

## Introduction

This is an application serving an API to handle embeddings. It stores embeddings in a PostgreSQL backend and uses its vector support, but allows you to manage different users, projects, and LLM configurations via a simple Restful API.

The typical use case is as a component of a Retrieval Augmented Generation (RAG) workflow: You create embeddings for a collection of text snippets and upload them to this API. For each text snippet, you upload a text identifier, the embeddings vector and, optionally, metadata or the text itself. Then, you can later

- POST another text and get the most similar texts in the database, or
- GET the most similar texts for a text that is already in the database by specifying the text's identifier in an URL

In both cases, the service returns a list of text identifiers that you can then use in your own processing, perhaps based on other means of providing the respective texts.

## Features

- OpenAPI documentation
- Supports different embeddings configurations (e.g. dimensions)
- Rights management (authentication via API token)

## Getting started

### Authenticating (providing the API key)

### Rate limiting

### API Versioning

## Endpoints

| Endpoint | Method | Description | Allowed Users |
|----------|--------|-------------|---------------|

For a more detailed, and always up-to-date documentation of the endpoints, see the automatically generated OpenAPI document.

## Code creation and structure

This API is programmed in go and uses the [huma](https://huma.rocks/) framework with go's stock `http.ServeMux()` routing.

The Code has been developed in dialogue with [ChatGPT](./docs/ChatGPT.md). After manual inspection and correction, this is the project structure:

```default
dhamps-vdb/
├── LICENSE
├── README.md
├── go.mod
├── go.sum
├── main.go
├── main_test.go
├── api/
│   └── openapi.yml         // OpenAPI spec file
├── docs/
│   └── ChatGPT.md          // Code as suggested by ChatGPT (GPT4 turbo and GPT4o) on 2024-06-09
├── internal/
│   ├── auth/
│   ├── database/
│   ├── handlers/
│   │   ├── admin.go
│   │   ├── projects.go
│   │   ├── embeddings.go
│   │   ├── similars.go
│   │   └── llm_process.go
│   └── models/
│       ├── user.go
│       ├── project.go
│       ├── embeddings.go
│       ├── similar.go
│       └── llm_process.go
└── web/                      // web resources for the html response
```

The application checks and migrates the database schema to the appropriate version if possible. It presupposes however, that a suitable database and user (with appropriate privileges) have been created.

### Run tests

Actual (mostly integration) tests are run like this:

```bash
$> systemctl --user start podman.socket
$> export DOCKER_HOST=unix://$XDG_RUNTIME_DIR/podman/podman.sock
$> go test -v ./...
```

These tests do not contact any separately installed/launched backend, and instead have a container managed by the testing itself (via [testcontainers](https://testcontainers.com/guides/getting-started-with-testcontainers-for-go/)).

### Run with local container

A local container with a pg_vector-enabled postgresql can be run like this:

```bash
$> podman run -p 8888:5432 -e POSTGRES_PASSWORD=password pgvector/pgvector:0.7.4-pg16
```

But be aware that the filesystem is not persisted if you run it like this. That means that when you stop and restart the container, you will have to re-setup the database as described below.

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

For testing (i.e. without compiling and deploying), you can go to the main directory of the git repository and launch the vdb app like this:

```bash
$> go run main.go --port=8880 --db-port=8888 --db-user=my_user --db-password=my-password --db-name=my_vectors
```

## Roadmap

- [√] Tests
- [√] Catch post to existing resources
- [√] User authentication & restrictions on some API calls
- [√] When testing, check cleanup by adding a new query/function to see if all tables are empty
- [√] API versioning
- [√] handle **metadata**
  - [ ] Validation with metadata schema
- [ ] Implement and make consequent use of **max_idle** (5), **max_concurr** (5), **timeouts**, and **cancellations**
- [ ] Use **transactions** (most importantly, when an action requires several queries, e.g. projects being added and then linked to several read-authorized users)
- [ ] **Dockerization**
- [ ] **Rate limiting** (redis, sliding window, implement headers)
- [ ] **Concurrency** (leaky bucket)
- [ ] **Batch mode**
- [ ] Check if pagination is supported consistently
- [ ] Check if input is validated consistently
- [ ] **Link or unlink** users/LLMs as standalone operations
- [ ] **Transfer** of projects from one owner to another as new operation
- [ ] better **options** handling (<https://huma.rocks/features/cli/>)
- [ ] proper **logging** with `--verbose` and `--quiet` modes
- [ ] Caching
- [ ] HTML UI?
- [ ] LLM handling
  - [ ] allow API keys for services to be read from env variables (on the server, but still maybe useful)
  - [ ] calls to LLM services

## License

## Versions
