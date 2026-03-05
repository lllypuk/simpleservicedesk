package auth_test

import (
	"net/http"
	"testing"

	"simpleservicedesk/generated/openapi"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/require"
)

func TestOpenAPISecurityConfiguration(t *testing.T) {
	t.Parallel()

	swagger, err := openapi.GetSwagger()
	require.NoError(t, err)

	require.NotNil(t, swagger.Components.SecuritySchemes)

	bearerAuth, exists := swagger.Components.SecuritySchemes["bearerAuth"]
	require.True(t, exists)
	require.NotNil(t, bearerAuth)
	require.NotNil(t, bearerAuth.Value)
	require.Equal(t, "http", bearerAuth.Value.Type)
	require.Equal(t, "bearer", bearerAuth.Value.Scheme)
	require.Equal(t, "JWT", bearerAuth.Value.BearerFormat)

	require.False(t, operationUsesBearerAuth(swagger, "/login", http.MethodPost))
	require.True(t, operationUsesBearerAuth(swagger, "/users", http.MethodGet))
	require.True(t, operationUsesBearerAuth(swagger, "/tickets", http.MethodGet))
}

func operationUsesBearerAuth(swagger *openapi3.T, path string, method string) bool {
	pathItem := swagger.Paths.Value(path)
	if pathItem == nil {
		return false
	}

	operation := operationForMethod(pathItem, method)
	if operation == nil {
		return false
	}

	if operation.Security != nil {
		return securityRequirementsContainBearerAuth(*operation.Security)
	}

	if len(swagger.Security) == 0 {
		return false
	}

	return securityRequirementsContainBearerAuth(swagger.Security)
}

func operationForMethod(pathItem *openapi3.PathItem, method string) *openapi3.Operation {
	switch method {
	case http.MethodGet:
		return pathItem.Get
	case http.MethodPost:
		return pathItem.Post
	case http.MethodPut:
		return pathItem.Put
	case http.MethodPatch:
		return pathItem.Patch
	case http.MethodDelete:
		return pathItem.Delete
	default:
		return nil
	}
}

func securityRequirementsContainBearerAuth(securityRequirements openapi3.SecurityRequirements) bool {
	for _, requirement := range securityRequirements {
		_, exists := requirement["bearerAuth"]
		if exists {
			return true
		}
	}

	return false
}
