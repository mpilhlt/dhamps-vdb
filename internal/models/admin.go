package models

import "net/http"

// Request and Response structs for the admin API
// The request structs must be structs with fields for the request path/query/header/cookie parameters and/or body.
// The response structs must be structs with fields for the output headers and body of the operation, if any.

// Reset Database
// GET Path: "/admin/reset-db"

type ResetDbRequest struct{}

type ResetDbResponse struct {
	Header []http.Header `json:"header,omitempty" doc:"Response headers"`
}
