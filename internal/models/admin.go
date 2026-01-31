package models

import "net/http"

// Request and Response structs for the admin API
// The request structs must be structs with fields for the request path/query/header/cookie parameters and/or body.
// The response structs must be structs with fields for the output headers and body of the operation, if any.

// Reset Database
// GET Path: "/v1/admin/footgun"

type ResetDbRequest struct{}

type ResetDbResponse struct {
	Header []http.Header `json:"header,omitempty" doc:"Response headers"`
}

// Sanity Check
// GET Path: "/v1/admin/sanity-check"

type SanityCheckRequest struct{}

type SanityCheckResponse struct {
	Header []http.Header `json:"header,omitempty" doc:"Response headers"`
	Body   struct {
		Status        string   `json:"status" doc:"Overall status: PASSED, WARNING, or FAILED"`
		TotalProjects int      `json:"total_projects" doc:"Total number of projects checked"`
		IssuesCount   int      `json:"issues_count" doc:"Number of validation issues found"`
		WarningsCount int      `json:"warnings_count" doc:"Number of warnings found"`
		Issues        []string `json:"issues,omitempty" doc:"List of validation issues"`
		Warnings      []string `json:"warnings,omitempty" doc:"List of warnings"`
	}
}
