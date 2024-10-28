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
	if input.Project != input.Body.Handle {
		return nil, huma.Error400BadRequest(fmt.Sprintf("project handle in URL (%s) does not match project handle in body (%s)", input.Project, input.Body.Handle))
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
	_, err = queries.RetrieveUser(ctx, input.User)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, huma.Error404NotFound(fmt.Sprintf("user %s not found", input.User))
		}
		return nil, huma.Error500InternalServerError(fmt.Sprintf("unable to check if user %s exists before deleting. %v", input.User, err))
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
				if uu != input.User {
					readers[uu] = true
				}
			}
		} else {
			u, err := getUserFunc(ctx, &models.GetUserRequest{Handle: user})
			if err != nil {
				return nil, huma.Error500InternalServerError(fmt.Sprintf("unable to get user %s", user))
			}
			if u.Body.Handle != user {
				return nil, huma.Error404NotFound(fmt.Sprintf("user %s not found", user))
			}
			if user != input.User {
				readers[user] = true
			}
		}
	}

	project := database.UpsertProjectParams{
		Handle:      input.Project,
		Description: pgtype.Text{String: input.Body.Description, Valid: true},
		Owner:       input.User,
	}

	p, err := queries.UpsertProject(ctx, project)
	if err != nil {
		return nil, huma.Error500InternalServerError("unable to upload project")
	}

	// 2. Link project and owner
	params := database.LinkProjectToUserParams{ProjectID: p.ProjectID, UserHandle: input.User, Role: "owner"}
	_, err = queries.LinkProjectToUser(ctx, params)
	if err != nil {
		return nil, huma.Error500InternalServerError(fmt.Sprintf("unable to link project to owner %s", input.User))
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
	response.Body.Handle = p.Handle

	return response, nil
}

// Create a user (without a handle being present in the URL)
func postProjectFunc(ctx context.Context, input *models.PostProjectRequest) (*models.UploadProjectResponse, error) {
	return putProjectFunc(ctx, &models.PutProjectRequest{User: input.User, Project: input.Body.Handle, Body: input.Body})
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
	_, err = queries.RetrieveUser(ctx, input.User)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, huma.Error404NotFound(fmt.Sprintf("user %s not found", input.User))
		}
		return nil, huma.Error500InternalServerError(fmt.Sprintf("unable to check if user %s exists before deleting. %v", input.User, err))
	}

	// Run the queries
	p, err := queries.GetProjectsByUser(ctx, database.GetProjectsByUserParams{UserHandle: input.User, Limit: int32(input.Limit), Offset: int32(input.Offset)})
	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, huma.Error404NotFound(fmt.Sprintf("no projects found for user %s", input.User))
		}
		return nil, huma.Error500InternalServerError(fmt.Sprintf("unable to get projects for user %s", input.User))
	}
	projects := []models.Project{}
	// Get the authorized reader accounts for each project
	for _, project := range p {
		readers := []string{}
		rows, err := queries.GetUsersByProject(ctx, database.GetUsersByProjectParams{Owner: input.User, Handle: project.Handle, Limit: 999, Offset: 0})
		if err != nil {
			return nil, huma.Error500InternalServerError(fmt.Sprintf("unable to get readers for %s's project %s", input.User, project.Handle))
		}
		for _, row := range rows {
			readers = append(readers, row.Handle)
		}
		projects = append(projects, models.Project{
			Id:                int(project.ProjectID),
			Handle:            project.Handle,
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
	_, err = queries.RetrieveUser(ctx, input.User)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, huma.Error404NotFound(fmt.Sprintf("user %s not found", input.User))
		}
		return nil, huma.Error500InternalServerError(fmt.Sprintf("unable to check if user %s exists before deleting. %v", input.User, err))
	}

	// Build the query parameters
	params := database.RetrieveProjectParams{
		Owner:  input.User,
		Handle: input.Project,
	}

	// Run the queries
	p, err := queries.RetrieveProject(ctx, params)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, huma.Error404NotFound(fmt.Sprintf("user %s's project %s not found", input.User, input.Project))
		}
		return nil, huma.Error500InternalServerError(fmt.Sprintf("unable to get project %s for user %s", input.Project, input.User))
	}
	// Get the authorized reader accounts for the project
	readers := []string{}
	rows, err := queries.GetUsersByProject(ctx, database.GetUsersByProjectParams{Owner: input.User, Handle: input.Project, Limit: 999, Offset: 0})
	if err != nil {
		return nil, huma.Error500InternalServerError(fmt.Sprintf("unable to get authorized reader accounts for %s's project %s", input.User, input.Project))
	}
	for _, row := range rows {
		readers = append(readers, row.Handle)
	}

	// Build the response
	response := &models.GetProjectResponse{}
	response.Body.Project = models.Project{
		Handle:            p.Handle,
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
	_, err = queries.RetrieveUser(ctx, input.User)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, huma.Error404NotFound(fmt.Sprintf("user %s not found", input.User))
		}
		return nil, huma.Error500InternalServerError(fmt.Sprintf("unable to check if user %s exists before deleting. %v", input.User, err))
	}

	// Check if project exists
	_, err = queries.RetrieveProject(ctx, database.RetrieveProjectParams{Owner: input.User, Handle: input.Project})
	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, huma.Error404NotFound(fmt.Sprintf("project %s not found for user %s", input.Project, input.User))
		}
		return nil, huma.Error500InternalServerError(fmt.Sprintf("unable to check if project %s exists before deleting. %v", input.Project, err))
	}

	// Build the query parameters
	params := database.DeleteProjectParams{
		Owner:  input.User,
		Handle: input.Project,
	}

	// Run the query
	err = queries.DeleteProject(ctx, params)
	if err != nil {
		return nil, huma.Error500InternalServerError(fmt.Sprintf("unable to delete project %s for user %s", input.Project, input.User))
	}

	// Build the response
	response := &models.DeleteProjectResponse{}

	return response, nil
}

// RegisterProjectRoutes registers all the project routes with the API
func RegisterProjectsRoutes(pool *pgxpool.Pool, api huma.API) error {
	// Define huma.Operations for each route
	putProjectOp := huma.Operation{
		OperationID:   "putProject",
		Method:        http.MethodPut,
		Path:          "/projects/{user}/{project}",
		DefaultStatus: http.StatusCreated,
		Summary:       "Create or update a project",
		Tags:          []string{"admin", "projects"},
	}
	postProjectOp := huma.Operation{
		OperationID:   "postProject",
		Method:        http.MethodPost,
		Path:          "/projects/{user}",
		DefaultStatus: http.StatusCreated,
		Summary:       "Create or update a project",
		Tags:          []string{"admin", "projects"},
	}
	getProjectsOp := huma.Operation{
		OperationID: "getProjects",
		Method:      http.MethodGet,
		Path:        "/projects/{user}",
		Summary:     "Get all projects for a specific user",
		Tags:        []string{"admin", "projects"},
	}
	getProjectOp := huma.Operation{
		OperationID: "getProject",
		Method:      http.MethodGet,
		Path:        "/projects/{user}/{project}",
		Summary:     "Get a specific project",
		Tags:        []string{"admin", "projects"},
	}
	deleteProjectOp := huma.Operation{
		OperationID:   "deleteProject",
		Method:        http.MethodDelete,
		Path:          "/projects/{user}/{project}",
		DefaultStatus: http.StatusNoContent,
		Summary:       "Delete a specific project",
		Tags:          []string{"admin", "projects"},
	}

	huma.Register(api, putProjectOp, addPoolToContext(pool, putProjectFunc))
	huma.Register(api, postProjectOp, addPoolToContext(pool, postProjectFunc))
	huma.Register(api, getProjectsOp, addPoolToContext(pool, getProjectsFunc))
	huma.Register(api, getProjectOp, addPoolToContext(pool, getProjectFunc))
	huma.Register(api, deleteProjectOp, addPoolToContext(pool, deleteProjectFunc))
	return nil
}
