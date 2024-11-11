package models

import "net/http"

// User represents a user account.
type User struct {
	UserHandle string    `json:"user_handle"      doc:"User handle" maxLength:"20"  minLength:"3" example:"jdoe"`
	Name       string    `json:"name,omitempty"   doc:"User name"   maxLength:"50"                example:"Jane Doe"`
	Email      string    `json:"email"            doc:"User email"  maxLength:"100" minLength:"5" example:"foo@bar.com"`
	APIKey     string    `json:"apiKey,omitempty" doc:"User API key for dhamps-vdb API" maxLength:"64" minLength:"64" example:"1234567890123456789012345678901212345678901234567890123456789012"`
	Projects   []Project `json:"projects,omitempty" doc:"Projects that the user is a member of"`
}

// Request and Response structs for the user administration API
// The request structs must be structs with fields for the request path/query/header/cookie parameters and/or body.
// The response structs must be structs with fields for the output headers and body of the operation, if any.

// Put/post user
// PUT Path: "/users/{user_handle}"

type PutUserRequest struct {
	UserHandle string `json:"user_handle" path:"user_handle" maxLength:"20" minLength:"3" example:"jdoe" doc:"User handle"`
	Body       User   `json:"user" doc:"User to create or update"`
}

// POST Path: "/users"

type PostUserRequest struct {
	Body User `json:"user" doc:"User to create or update"`
}

type UploadUserResponse struct {
	Header []http.Header `json:"header,omitempty" doc:"Response headers"`
	Body   HandleAPIStruct
}

type HandleAPIStruct struct {
	UserHandle string `json:"user_handle" doc:"Handle of created or updated user"`
	APIKey     string `json:"api_key" doc:"API key for the user"`
}

// Get all users
// Path: "/users"

type GetUsersRequest struct {
	Limit  int `json:"limit,omitempty" query:"limit" minimum:"1" maximum:"200" example:"10" default:"10" doc:"Maximum number of users to return"`
	Offset int `json:"offset,omitempty" query:"offset" minimum:"0" example:"0" default:"0" doc:"Offset into the list of users"`
}

type GetUsersResponse struct {
	Header []http.Header `json:"header,omitempty" doc:"Response headers"`
	Body   []string      `json:"handles" doc:"Handles of all registered user accounts"`
}

// Get single user information
// Path "/users/{user_handle}"

type GetUserRequest struct {
	UserHandle string `json:"user_handle" path:"user_handle" maxLength:"20" minLength:"3" example:"jdoe" doc:"User handle"`
}

type GetUserResponse struct {
	Header []http.Header `json:"header,omitempty" doc:"Response headers"`
	Body   User          `json:"user" doc:"User information"`
}

// Delete user
// Path "/users/{user_handle}"

type DeleteUserRequest struct {
	UserHandle string `json:"user_handle" path:"user_handle" maxLength:"20" minLength:"3" example:"jdoe" doc:"User handle"`
}

type DeleteUserResponse struct {
	Header []http.Header `json:"header,omitempty" doc:"Response headers"`
}
