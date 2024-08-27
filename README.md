# dhamps-vdb
Vector Database for the DH at Max Planck Society initiative

[![Go Report Card](https://goreportcard.com/badge/github.com/mpilhlt/dhamps-vdb?style=flat-square)](https://goreportcard.com/report/github.com/mpilhlt/dhamps-vdb)

[![Go Reference](https://pkg.go.dev/badge/github.com/mpilhlt/dhamps-vdb.svg)](https://pkg.go.dev/github.com/mpilhlt/dhamps-vdb)

[![Release](https://img.shields.io/github/release/golang-standards/project-layout.svg?style=flat-square)](https://github.com/golang-standards/project-layout/releases/latest)

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
