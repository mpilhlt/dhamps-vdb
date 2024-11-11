package handlers

import (
	"context"
	"fmt"
	"net/http"

	"github.com/mpilhlt/dhamps-vdb/internal/database"
	"github.com/mpilhlt/dhamps-vdb/internal/models"

	"github.com/danielgtaylor/huma/v2"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// TODO: Add LLMServices fields

// Create a new project
func putProjectFunc(ctx context.Context, input *models.PutProjectRequest) (*models.UploadProjectResponse, error) {
	if input.ProjectHandle != input.Body.ProjectHandle {
		return nil, huma.Error400BadRequest(fmt.Sprintf("project handle in URL (%s) does not match project handle in body (%s)", input.ProjectHandle, input.Body.ProjectHandle))
	}

	// Get the database connection pool from the context
	pool, err := GetDBPool(ctx)
	if err != nil {
		return nil, err
	} else if pool == nil {
		return nil, huma.Error500InternalServerError("database connection pool is nil")
	}

	// Check if user exists
	queries := database.New(pool)
	_, err = queries.RetrieveUser(ctx, input.UserHandle)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, huma.Error404NotFound(fmt.Sprintf("user %s not found", input.UserHandle))
		}
		return nil, huma.Error500InternalServerError(fmt.Sprintf("unable to check if user %s exists before deleting. %v", input.UserHandle, err))
	}

	// 1. Upload project

	// Build query parameters (project)
	readers := make(map[string]bool)
	for _, user := range input.Body.AuthorizedReaders {
		if user == "*" {
			users, err := getUsersFunc(ctx, &models.GetUsersRequest{})
			if err != nil {
				return nil, huma.Error500InternalServerError(fmt.Sprintf("unable to get user %s", user))
			}
			for _, uu := range users.Body {
				if uu != input.UserHandle {
					readers[uu] = true
				}
			}
		} else {
			u, err := getUserFunc(ctx, &models.GetUserRequest{UserHandle: user})
			if err != nil {
				return nil, huma.Error500InternalServerError(fmt.Sprintf("unable to get user %s", user))
			}
			if u.Body.UserHandle != user {
				return nil, huma.Error404NotFound(fmt.Sprintf("user %s not found", user))
			}
			if user != input.UserHandle {
				readers[user] = true
			}
		}
	}

	project := database.UpsertProjectParams{
		ProjectHandle: input.ProjectHandle,
		Description:   pgtype.Text{String: input.Body.Description, Valid: true},
		Owner:         input.UserHandle,
	}

	p, err := queries.UpsertProject(ctx, project)
	if err != nil {
		return nil, huma.Error500InternalServerError("unable to upload project")
	}

	// 2. Link project and owner
	params := database.LinkProjectToUserParams{ProjectID: p.ProjectID, UserHandle: input.UserHandle, Role: "owner"}
	_, err = queries.LinkProjectToUser(ctx, params)
	if err != nil {
		return nil, huma.Error500InternalServerError(fmt.Sprintf("unable to link project to owner %s", input.UserHandle))
	}

	// 3. Link project and other assigned readers
	for reader := range readers {
		params := database.LinkProjectToUserParams{ProjectID: p.ProjectID, UserHandle: reader, Role: "reader"}
		_, err := queries.LinkProjectToUser(ctx, params)
		if err != nil {
			return nil, huma.Error500InternalServerError(fmt.Sprintf("unable to upload project reader %s", reader))
		}
		// registeredReaders = append(registeredReaders, u.UserHandle)
	}

	// 4. Build the response
	response := &models.UploadProjectResponse{}
	response.Body.ProjectHandle = p.ProjectHandle

	return response, nil
}

// Create a user (without a handle being present in the URL)
func postProjectFunc(ctx context.Context, input *models.PostProjectRequest) (*models.UploadProjectResponse, error) {
	return putProjectFunc(ctx, &models.PutProjectRequest{UserHandle: input.UserHandle, ProjectHandle: input.Body.ProjectHandle, Body: input.Body})
}

// Get all projects for a specific user
func getProjectsFunc(ctx context.Context, input *models.GetProjectsRequest) (*models.GetProjectsResponse, error) {
	// Get the database connection pool from the context
	pool, err := GetDBPool(ctx)
	if err != nil {
		return nil, err
	}

	// Check if user exists
	queries := database.New(pool)
	_, err = queries.RetrieveUser(ctx, input.UserHandle)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, huma.Error404NotFound(fmt.Sprintf("user %s not found", input.UserHandle))
		}
		return nil, huma.Error500InternalServerError(fmt.Sprintf("unable to check if user %s exists before deleting. %v", input.UserHandle, err))
	}

	// Run the queries
	p, err := queries.GetProjectsByUser(ctx, database.GetProjectsByUserParams{UserHandle: input.UserHandle, Limit: int32(input.Limit), Offset: int32(input.Offset)})
	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, huma.Error404NotFound(fmt.Sprintf("no projects found for user %s", input.UserHandle))
		}
		return nil, huma.Error500InternalServerError(fmt.Sprintf("unable to get projects for user %s", input.UserHandle))
	}
	projects := []models.Project{}
	// Get the authorized reader accounts for each project
	for _, project := range p {
		readers := []string{}
		rows, err := queries.GetUsersByProject(ctx, database.GetUsersByProjectParams{Owner: input.UserHandle, ProjectHandle: project.ProjectHandle, Limit: 999, Offset: 0})
		if err != nil {
			return nil, huma.Error500InternalServerError(fmt.Sprintf("unable to get readers for %s's project %s", input.UserHandle, project.ProjectHandle))
		}
		for _, row := range rows {
			readers = append(readers, row.UserHandle)
		}
		projects = append(projects, models.Project{
			ProjectId:         int(project.ProjectID),
			ProjectHandle:     project.ProjectHandle,
			Description:       project.Description.String,
			AuthorizedReaders: readers,
			LLMServices:       nil,
		})
	}

	// Build the response
	response := &models.GetProjectsResponse{}
	response.Body.Projects = projects

	return response, nil
}

// Retrieve a specific project
func getProjectFunc(ctx context.Context, input *models.GetProjectRequest) (*models.GetProjectResponse, error) {
	// Get the database connection pool from the context
	pool, err := GetDBPool(ctx)
	if err != nil {
		return nil, err
	}

	// Check if user exists
	queries := database.New(pool)
	_, err = queries.RetrieveUser(ctx, input.UserHandle)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, huma.Error404NotFound(fmt.Sprintf("user %s not found", input.UserHandle))
		}
		return nil, huma.Error500InternalServerError(fmt.Sprintf("unable to check if user %s exists before deleting. %v", input.UserHandle, err))
	}

	// Build the query parameters
	params := database.RetrieveProjectParams{
		Owner:         input.UserHandle,
		ProjectHandle: input.ProjectHandle,
	}

	// Run the queries
	p, err := queries.RetrieveProject(ctx, params)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, huma.Error404NotFound(fmt.Sprintf("user %s's project %s not found", input.UserHandle, input.ProjectHandle))
		}
		return nil, huma.Error500InternalServerError(fmt.Sprintf("unable to get project %s for user %s", input.ProjectHandle, input.UserHandle))
	}
	// Get the authorized reader accounts for the project
	readers := []string{}
	rows, err := queries.GetUsersByProject(ctx, database.GetUsersByProjectParams{Owner: input.UserHandle, ProjectHandle: input.ProjectHandle, Limit: 999, Offset: 0})
	if err != nil {
		return nil, huma.Error500InternalServerError(fmt.Sprintf("unable to get authorized reader accounts for %s's project %s", input.UserHandle, input.ProjectHandle))
	}
	for _, row := range rows {
		readers = append(readers, row.UserHandle)
	}

	// Build the response
	response := &models.GetProjectResponse{}
	response.Body.Project = models.Project{
		ProjectHandle:     p.ProjectHandle,
		Description:       p.Description.String,
		MetadataScheme:    p.MetadataScheme.String,
		AuthorizedReaders: readers,
		LLMServices:       nil,
	}

	return response, nil
}

func deleteProjectFunc(ctx context.Context, input *models.DeleteProjectRequest) (*models.DeleteProjectResponse, error) {
	// Get the database connection pool from the context
	pool, err := GetDBPool(ctx)
	if err != nil {
		return nil, err
	}

	// Check if user exists
	queries := database.New(pool)
	_, err = queries.RetrieveUser(ctx, input.UserHandle)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, huma.Error404NotFound(fmt.Sprintf("user %s not found", input.UserHandle))
		}
		return nil, huma.Error500InternalServerError(fmt.Sprintf("unable to check if user %s exists before deleting. %v", input.UserHandle, err))
	}

	// Check if project exists
	_, err = queries.RetrieveProject(ctx, database.RetrieveProjectParams{Owner: input.UserHandle, ProjectHandle: input.ProjectHandle})
	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, huma.Error404NotFound(fmt.Sprintf("project %s not found for user %s", input.ProjectHandle, input.UserHandle))
		}
		return nil, huma.Error500InternalServerError(fmt.Sprintf("unable to check if project %s exists before deleting. %v", input.ProjectHandle, err))
	}

	// Build the query parameters
	params := database.DeleteProjectParams{
		Owner:         input.UserHandle,
		ProjectHandle: input.ProjectHandle,
	}

	// Run the query
	err = queries.DeleteProject(ctx, params)
	if err != nil {
		return nil, huma.Error500InternalServerError(fmt.Sprintf("unable to delete project %s for user %s", input.ProjectHandle, input.UserHandle))
	}

	// Build the response
	response := &models.DeleteProjectResponse{}

	return response, nil
}

// RegisterProjectRoutes registers all the project routes with the API
func RegisterProjectsRoutes(pool *pgxpool.Pool, keyGen RandomKeyGenerator, api huma.API) error {
	// Define huma.Operations for each route
	putProjectOp := huma.Operation{
		OperationID:   "putProject",
		Method:        http.MethodPut,
		Path:          "/projects/{user_handle}/{project_handle}",
		DefaultStatus: http.StatusCreated,
		Summary:       "Create or update a project",
		Security: []map[string][]string{
			{"adminAuth": []string{"admin"}},
			{"ownerAuth": []string{"owner"}},
		},
		Tags: []string{"admin", "projects"},
	}
	postProjectOp := huma.Operation{
		OperationID:   "postProject",
		Method:        http.MethodPost,
		Path:          "/projects/{user_handle}",
		DefaultStatus: http.StatusCreated,
		Summary:       "Create or update a project",
		Security: []map[string][]string{
			{"adminAuth": []string{"admin"}},
			{"ownerAuth": []string{"owner"}},
		},
		Tags: []string{"admin", "projects"},
	}
	getProjectsOp := huma.Operation{
		OperationID: "getProjects",
		Method:      http.MethodGet,
		Path:        "/projects/{user_handle}",
		Summary:     "Get all projects for a specific user",
		Security: []map[string][]string{
			{"adminAuth": []string{"admin"}},
			{"ownerAuth": []string{"owner"}},
		},
		Tags: []string{"admin", "projects"},
	}
	getProjectOp := huma.Operation{
		OperationID: "getProject",
		Method:      http.MethodGet,
		Path:        "/projects/{user_handle}/{project_handle}",
		Summary:     "Get a specific project",
		Security: []map[string][]string{
			{"adminAuth": []string{"admin"}},
			{"ownerAuth": []string{"owner"}},
			{"readerAuth": []string{"reader"}},
		},
		Tags: []string{"admin", "projects"},
	}
	deleteProjectOp := huma.Operation{
		OperationID:   "deleteProject",
		Method:        http.MethodDelete,
		Path:          "/projects/{user_handle}/{project_handle}",
		DefaultStatus: http.StatusNoContent,
		Summary:       "Delete a specific project",
		Security: []map[string][]string{
			{"adminAuth": []string{"admin"}},
			{"ownerAuth": []string{"owner"}},
		},
		Tags: []string{"admin", "projects"},
	}

	huma.Register(api, putProjectOp, addPoolToContext(pool, putProjectFunc))
	huma.Register(api, postProjectOp, addPoolToContext(pool, postProjectFunc))
	huma.Register(api, getProjectsOp, addPoolToContext(pool, getProjectsFunc))
	huma.Register(api, getProjectOp, addPoolToContext(pool, getProjectFunc))
	huma.Register(api, deleteProjectOp, addPoolToContext(pool, deleteProjectFunc))
	return nil
}
