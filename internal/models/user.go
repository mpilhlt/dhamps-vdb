package models

import "net/http"

// User represents a user account.
type User struct {
  Handle      string `json:"handle"          doc:"User handle" maxLength:"20"  minLength:"3" example:"jdoe"`
  Name        string `json:"name,omitempty"  doc:"User name"   maxLength:"50"                example:"Jane Doe"`
  Email       string `json:"email"           doc:"User email"  maxLength:"100" minLength:"5" example:"foo@bar.com"`
  APIKey      string `json:"apiKey"          doc:"User API key for dhamps-vdb API" maxLength:"32" minLength:"32" example:"12345678901234567890123456789012"`
  Projects    []Project `json:"projects" doc:"Projects that the user is a member of" default:nil`
}

// Request and Response structs for the user administration API
// The request structs must be structs with fields for the request path/query/header/cookie parameters and/or body.
// The response structs must be structs with fields for the output headers and body of the operation, if any.

// Put/post user request/response
// Path: "/users"

type PutUserRequest struct {
  Body struct {
    User User `json:"user" doc:"User to create or update"`
  }
}

type PutUserResponse struct {
  Header []http.Header `json:"header,omitempty" doc:"Response headers"`
  Body struct {
    Handle string `json:"user" doc:"Handle of created or updated user"`
  }
}

// Get all users request/response
// Path: "/users"

type GetUsersRequest struct {}

type GetUsersResponse struct {
  Header []http.Header `json:"header,omitempty" doc:"Response headers"`
  Body struct {
    Handles []string `json:"handles" doc:"Handles of all registered user accounts"`
  }
}

// Patch user request/response
// Path "/users/{user}"

type PatchUserRequest struct {
  User string `json:"user" path:"user" maxLength:"20" minLength:"3" example:"jdoe" doc:"User handle"`
  Body struct {
    User User `json:"user" doc:"User to update"`
  }
}

type PatchUserResponse struct {
  Header []http.Header `json:"header,omitempty" doc:"Response headers"`
  Body struct {
    User User `json:"user" doc:"Updated user"`
  }
}

// Get user information request/response
// Path "/users/{user}"

type GetUserRequest struct {
  User string `json:"user" path:"user" maxLength:"20" minLength:"3" example:"jdoe" doc:"User handle"`
}

type GetUserResponse struct {
  Header []http.Header `json:"header,omitempty" doc:"Response headers"`
  Body struct {
    User User `json:"user" doc:"User information"`
  }
}

// Delete user request/response
// Path "/users/{user}"

type DeleteUserRequest struct {
  User string `json:"user" path:"user" maxLength:"20" minLength:"3" example:"jdoe" doc:"User handle"`
}

type DeleteUserResponse struct {
  Header []http.Header `json:"header,omitempty" doc:"Response headers"`
  Body string `json:"body" doc:"Status message"`
}
