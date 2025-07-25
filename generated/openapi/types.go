// Package openapi provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/deepmap/oapi-codegen version v1.16.3 DO NOT EDIT.
package openapi

import (
	openapi_types "github.com/oapi-codegen/runtime/types"
)

// CreateUserRequest defines model for CreateUserRequest.
type CreateUserRequest struct {
	Email    openapi_types.Email `json:"email"`
	Name     string              `json:"name"`
	Password string              `json:"password"`
}

// CreateUserResponse defines model for CreateUserResponse.
type CreateUserResponse struct {
	Id *openapi_types.UUID `json:"id,omitempty"`
}

// ErrorResponse defines model for ErrorResponse.
type ErrorResponse struct {
	Message *string `json:"message,omitempty"`
}

// GetUserResponse defines model for GetUserResponse.
type GetUserResponse struct {
	Email *openapi_types.Email `json:"email,omitempty"`
	Id    *openapi_types.UUID  `json:"id,omitempty"`
	Name  *string              `json:"name,omitempty"`
}

// PostUsersJSONRequestBody defines body for PostUsers for application/json ContentType.
type PostUsersJSONRequestBody = CreateUserRequest
