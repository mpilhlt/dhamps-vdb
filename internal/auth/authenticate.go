package auth

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"

	"github.com/mpilhlt/dhamps-vdb/internal/database"
	"github.com/mpilhlt/dhamps-vdb/internal/models"

	"github.com/danielgtaylor/huma/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	AuthUserKey = "authUser"
	IsAdminKey  = "isAdmin"
	IsOwnerKey  = "isOwner"
)

// Config is the security scheme configuration for the API.
var Config = map[string]*huma.SecurityScheme{
	"adminAuth": {
		Type:   "apiKey",
		In:     "header",
		Scheme: "bearer",
		Name:   "Authorization",
	},
	"ownerAuth": {
		Type:   "apiKey",
		In:     "header",
		Scheme: "bearer",
		Name:   "Authorization",
	},
	"readerAuth": {
		Type:   "apiKey",
		In:     "header",
		Scheme: "bearer",
		Name:   "Authorization",
	},
}

// APITermination returns a middleware function that evaluates if any of the preceding
//
//	authentication middleware functions were successful. If not, it rejects the request,
//	otherwise it calls the next middleware (or the final handler) function.
//	This is supposed to be called as the last auth middleware function in
//	the chain.
func AuthTermination(api huma.API) func(ctx huma.Context, next func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {
		// Check if the current operation requires authentication
		isAuthRequired := false
		for _, securityScheme := range ctx.Operation().Security {
			if len(securityScheme) > 0 {
				isAuthRequired = true
				break
			}
		}

		if !isAuthRequired {
			// No authentication required for this operation
			next(ctx)
			return
		}

		// Check if any authentication middleware has set AuthUserKey
		if _, ok := ctx.Context().Value(AuthUserKey).(string); ok {
			next(ctx)
			return
		}
		fmt.Print("        Authentication failed.\n")
		_ = huma.WriteErr(api, ctx, http.StatusUnauthorized, "Authentication failed. Perhaps a missing or incorrect API key?")
	}
}

// APIKey... functions return a middleware function that checks for a valid API key.

// APIKeyAdminAuth checks for an admin API key in the Authorization header.
func APIKeyAdminAuth(api huma.API, options *models.Options) func(ctx huma.Context, next func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {

		// Check if adminAuth is applicable
		isAuthorizationRequired := false
		for _, opScheme := range ctx.Operation().Security {
			var ok bool
			if _, ok = opScheme["adminAuth"]; ok {
				isAuthorizationRequired = true
				break
			}
		}
		if !isAuthorizationRequired {
			next(ctx)
			return
		}

		token := strings.TrimPrefix(ctx.Header("Authorization"), "Bearer ")

		if token == options.AdminKey {
			ctx = huma.WithValue(ctx, IsAdminKey, true)
			ctx = huma.WithValue(ctx, AuthUserKey, "admin")
			fmt.Print("        Admin authentication successful\n")
			next(ctx)
			return
		}

		next(ctx)
	}
}

// APIKeyOwnerAuth checks for an owner API key in the Authorization header.
func APIKeyOwnerAuth(api huma.API, pool *pgxpool.Pool, options *models.Options) func(ctx huma.Context, next func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {

		// Check if ownerAuth is applicable
		isAuthorizationRequired := false
		for _, opScheme := range ctx.Operation().Security {
			var ok bool
			if _, ok = opScheme["ownerAuth"]; ok {
				isAuthorizationRequired = true
				break
			}
		}
		if !isAuthorizationRequired {
			next(ctx)
			return
		}

		// Check if adminAuth has already authenticated the request
		if isAdmin, ok := ctx.Context().Value(IsAdminKey).(bool); ok && isAdmin {
			next(ctx)
			return
		}

		owner := ctx.Param("user_handle")
		token := strings.TrimPrefix(ctx.Header("Authorization"), "Bearer ")

		if len(owner) == 0 {
			next(ctx)
			return
		}

		queries := database.New(pool)
		storedHash, err := queries.GetKeyByUser(ctx.Context(), owner)
		if err != nil && err.Error() == "no rows in result set" {
			next(ctx)
			return
		}
		if err != nil && err.Error() != "no rows in result set" {
			_ = huma.WriteErr(api, ctx, http.StatusInternalServerError, "unable to check if owner exists")
			return
		}
		if storedHash == "" {
			next(ctx)
			return
		}

		// fmt.Printf("        check owner hash against API token: %s/%s ...\n", storedHash, token)
		if apiKeyIsValid(token, storedHash) {
			ctx = huma.WithValue(ctx, IsOwnerKey, true)
			ctx = huma.WithValue(ctx, AuthUserKey, owner)
			fmt.Printf("        Owner authentication successful: %s\n", owner)
			next(ctx)
			return
		}

		next(ctx)
	}
}

// APIKeyReaderAuth checks for a reader API key in the Authorization header.
func APIKeyReaderAuth(api huma.API, pool *pgxpool.Pool, options *models.Options) func(ctx huma.Context, next func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {
		// Check if readerAuth is applicable
		isAuthorizationRequired := false
		for _, opScheme := range ctx.Operation().Security {
			var ok bool
			if _, ok = opScheme["readerAuth"]; ok {
				isAuthorizationRequired = true
				break
			}
		}
		if !isAuthorizationRequired {
			next(ctx)
			return
		}
		// Check if adminAuth or ownerAuth has already authenticated the request
		if isAdmin, ok := ctx.Context().Value(IsAdminKey).(bool); ok && isAdmin {
			next(ctx)
			return
		}
		if isOwner, ok := ctx.Context().Value(IsOwnerKey).(bool); ok && isOwner {
			next(ctx)
			return
		}

		owner := ctx.Param("user_handle")
		project := ctx.Param("project_handle")
		token := strings.TrimPrefix(ctx.Header("Authorization"), "Bearer ")

		if len(owner) == 0 || len(project) == 0 {
			next(ctx)
			return
		}

		// Check if the project has public_read enabled
		queries := database.New(pool)
		publicReadParams := database.IsProjectPubliclyReadableParams{
			Owner:         owner,
			ProjectHandle: project,
		}
		publicRead, err := queries.IsProjectPubliclyReadable(ctx.Context(), publicReadParams)
		// If project exists and public_read is true, allow unauthenticated access
		if err == nil && publicRead.Valid && publicRead.Bool {
			// Public read is enabled, allow unauthenticated access
			fmt.Print("        Public read access granted (no authentication required)\n")
			ctx = huma.WithValue(ctx, AuthUserKey, "public")
			next(ctx)
			return
		}
		// If there's an error (e.g., project not found), continue to check authorized readers
		// The project existence check will happen in the handler

		// If not public, check for authorized readers
		getKeysParams := database.GetKeysByLinkedUsersParams{
			Owner:         owner,
			ProjectHandle: project,
			Limit:         50,
			Offset:        0,
		}
		allowedUsers, err := queries.GetKeysByLinkedUsers(ctx.Context(), getKeysParams)
		if err != nil && err.Error() != "no rows in result set" {
			_ = huma.WriteErr(api, ctx, http.StatusInternalServerError, "unable to get linked users")
			return
		}
		if err != nil && err.Error() == "no rows in result set" {
			next(ctx)
			return
		}
		for _, authUser := range allowedUsers {
			storedHash := authUser.VdbAPIKey

			if apiKeyIsValid(token, storedHash) {
				fmt.Print("        Reader authentication successful\n")
				ctx = huma.WithValue(ctx, AuthUserKey, authUser.UserHandle)
				next(ctx)
				return
			}
		}

		next(ctx)
	}
}

// apiKeyIsValid checks if the given API key is valid
func apiKeyIsValid(rawKey string, storedHash string) bool {
	hash := sha256.Sum256([]byte(rawKey))
	hashedKey := hex.EncodeToString(hash[:])

	contentEqual := subtle.ConstantTimeCompare([]byte(storedHash), []byte(hashedKey)) == 1
	return contentEqual
}

// CORSMiddleware handles CORS for the API
func CORSMiddleware(api huma.API) func(ctx huma.Context, next func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {
		// Set CORS headers
		for key, value := range map[string]string{
			"Access-Control-Allow-Origin":  "*",
			"Access-Control-Allow-Methods": "GET, POST, PUT, DELETE, OPTIONS, PATCH, UPDATE, QUERY",
			"Access-Control-Allow-Headers": "Accept, Authorization, Content-Type, Content-Disposition, Origin, X-Requested-With",
		} {
			ctx.SetHeader(key, value)
		}

		// If this is a preflight OPTIONS request, return immediately with 200 OK
		if ctx.Operation().Method == "OPTIONS" {
			// fmt.Print("    OPTIONS request received, handled in CORS middleware.\n")
			ctx.SetStatus(http.StatusOK)
			return
		}

		// Otherwise, continue processing the request
		next(ctx)
	}
}
