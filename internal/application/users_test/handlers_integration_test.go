package users_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	"simpleservicedesk/generated/openapi"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func (s *UsersSuite) TestCreateUserIntegration() {
	tests := []struct {
		name           string
		request        interface{}
		expectedStatus int
		expectedError  *string
		validateID     bool
	}{
		{
			name: "valid user creation",
			request: openapi.CreateUserRequest{
				Name:     "John Doe",
				Email:    "john.doe@example.com",
				Password: "password123",
			},
			expectedStatus: http.StatusCreated,
			validateID:     true,
		},
		{
			name: "duplicate email",
			request: openapi.CreateUserRequest{
				Name:     "Jane Doe",
				Email:    "john.doe@example.com", // Same email as above
				Password: "password123",
			},
			expectedStatus: http.StatusConflict,
		},
		{
			name: "invalid email format",
			request: map[string]interface{}{
				"name":     "Invalid User",
				"email":    "invalid-email",
				"password": "password123",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "empty name",
			request: openapi.CreateUserRequest{
				Name:     "",
				Email:    "empty.name@example.com",
				Password: "password123",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "empty email",
			request: map[string]interface{}{
				"name":     "Empty Email",
				"email":    "",
				"password": "password123",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "empty password",
			request: openapi.CreateUserRequest{
				Name:     "Empty Password",
				Email:    "empty.password@example.com",
				Password: "",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "short password",
			request: openapi.CreateUserRequest{
				Name:     "Short Password",
				Email:    "short.password@example.com",
				Password: "12345", // Less than 6 characters
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid JSON",
			request:        `{"invalid": json}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "empty request body",
			request:        "",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			var reqBody []byte
			var err error

			switch v := tt.request.(type) {
			case string:
				reqBody = []byte(v)
			default:
				reqBody, err = json.Marshal(v)
				s.Require().NoError(err)
			}

			req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(reqBody))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			s.HTTPServer.ServeHTTP(rec, req)

			s.Assert().Equal(tt.expectedStatus, rec.Code, "Response: %s", rec.Body.String())

			if tt.expectedStatus == http.StatusCreated {
				var resp openapi.CreateUserResponse
				unmarshalErr := json.Unmarshal(rec.Body.Bytes(), &resp)
				s.Require().NoError(unmarshalErr)
				if tt.validateID {
					s.Assert().NotNil(resp.Id)
					s.Assert().NotEqual(uuid.Nil, *resp.Id)
				}
			} else {
				var errorResp openapi.ErrorResponse
				unmarshalErr := json.Unmarshal(rec.Body.Bytes(), &errorResp)
				s.Require().NoError(unmarshalErr)
				s.Assert().NotNil(errorResp.Message)
				s.Assert().NotEmpty(*errorResp.Message)
			}
		})
	}
}

func (s *UsersSuite) TestGetUserIntegration() {
	// First create a user to test getting
	createReq := openapi.CreateUserRequest{
		Name:     "Test User",
		Email:    "test.user@example.com",
		Password: "password123",
	}
	reqBody, _ := json.Marshal(createReq)
	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	s.HTTPServer.ServeHTTP(rec, req)
	s.Require().Equal(http.StatusCreated, rec.Code)

	var createResp openapi.CreateUserResponse
	err := json.Unmarshal(rec.Body.Bytes(), &createResp)
	s.Require().NoError(err)
	s.Require().NotNil(createResp.Id)
	createdUserID := *createResp.Id

	tests := []struct {
		name           string
		userID         string
		expectedStatus int
		validateResp   bool
	}{
		{
			name:           "get existing user",
			userID:         createdUserID.String(),
			expectedStatus: http.StatusOK,
			validateResp:   true,
		},
		{
			name:           "get non-existent user",
			userID:         uuid.New().String(),
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "invalid UUID format",
			userID:         "invalid-uuid",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			url := fmt.Sprintf("/users/%s", tt.userID)
			testReq := httptest.NewRequest(http.MethodGet, url, nil)
			testRec := httptest.NewRecorder()
			s.HTTPServer.ServeHTTP(testRec, testReq)

			s.Assert().Equal(tt.expectedStatus, testRec.Code, "Response: %s", testRec.Body.String())

			if tt.validateResp && tt.expectedStatus == http.StatusOK {
				var resp openapi.GetUserResponse
				unmarshalErr := json.Unmarshal(testRec.Body.Bytes(), &resp)
				s.Require().NoError(unmarshalErr)
				s.Assert().NotNil(resp.Id)
				s.Assert().Equal(createdUserID, *resp.Id)
				s.Assert().NotNil(resp.Name)
				s.Assert().Equal("Test User", *resp.Name)
				s.Assert().NotNil(resp.Email)
				s.Assert().Equal("test.user@example.com", string(*resp.Email))
			} else if tt.expectedStatus != http.StatusOK {
				var errorResp openapi.ErrorResponse
				unmarshalErr := json.Unmarshal(testRec.Body.Bytes(), &errorResp)
				s.Require().NoError(unmarshalErr)
				s.Assert().NotNil(errorResp.Message)
				s.Assert().NotEmpty(*errorResp.Message)
			}
		})
	}
}

func (s *UsersSuite) TestPingEndpoint() {
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	rec := httptest.NewRecorder()
	s.HTTPServer.ServeHTTP(rec, req)

	s.Assert().Equal(http.StatusOK, rec.Code)
	s.Assert().Equal("pong", rec.Body.String())
}

func (s *UsersSuite) TestContentTypeValidation() {
	userReq := openapi.CreateUserRequest{
		Name:     "Test User",
		Email:    "test@example.com",
		Password: "password123",
	}
	reqBody, _ := json.Marshal(userReq)

	tests := []struct {
		name           string
		contentType    string
		expectedStatus int
	}{
		{
			name:           "valid JSON content type",
			contentType:    echo.MIMEApplicationJSON,
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "missing content type",
			contentType:    "",
			expectedStatus: http.StatusUnsupportedMediaType,
		},
		{
			name:           "invalid content type",
			contentType:    "text/plain",
			expectedStatus: http.StatusUnsupportedMediaType,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(reqBody))
			if tt.contentType != "" {
				req.Header.Set(echo.HeaderContentType, tt.contentType)
			}
			rec := httptest.NewRecorder()
			s.HTTPServer.ServeHTTP(rec, req)

			s.Assert().Equal(tt.expectedStatus, rec.Code, "Response: %s", rec.Body.String())
		})
	}
}

func (s *UsersSuite) TestHTTPMethodValidation() {
	tests := []struct {
		name           string
		method         string
		path           string
		expectedStatus int
	}{
		{
			name:           "POST to /users - allowed",
			method:         http.MethodPost,
			path:           "/users",
			expectedStatus: http.StatusBadRequest, // Bad request due to empty body, but method is allowed
		},
		{
			name:           "GET to /users/{id} - allowed",
			method:         http.MethodGet,
			path:           "/users/" + uuid.New().String(),
			expectedStatus: http.StatusNotFound, // Not found, but method is allowed
		},
		{
			name:           "PUT to /users - not allowed",
			method:         http.MethodPut,
			path:           "/users",
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "DELETE to /users/{id} - not allowed",
			method:         http.MethodDelete,
			path:           "/users/" + uuid.New().String(),
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "PATCH to /users/{id} - not allowed",
			method:         http.MethodPatch,
			path:           "/users/" + uuid.New().String(),
			expectedStatus: http.StatusMethodNotAllowed,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			var body *bytes.Buffer
			if tt.method == http.MethodPost {
				body = bytes.NewBufferString("{}")
			} else {
				body = bytes.NewBuffer(nil)
			}

			req := httptest.NewRequest(tt.method, tt.path, body)
			if tt.method == http.MethodPost {
				req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			}
			rec := httptest.NewRecorder()
			s.HTTPServer.ServeHTTP(rec, req)

			s.Assert().Equal(tt.expectedStatus, rec.Code, "Response: %s", rec.Body.String())
		})
	}
}

func (s *UsersSuite) TestLargePayloadHandling() {
	// Test with a large name (should be accepted since there's no length validation in domain)
	largeString := strings.Repeat("A", 1000) // 1KB string - reasonable size

	userReq := openapi.CreateUserRequest{
		Name:     largeString,
		Email:    "large.payload@example.com",
		Password: "password123",
	}
	reqBody, _ := json.Marshal(userReq)

	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	s.HTTPServer.ServeHTTP(rec, req)

	// Should be accepted since domain only validates non-empty names
	s.Assert().Equal(http.StatusCreated, rec.Code, "Response: %s", rec.Body.String())

	var resp openapi.CreateUserResponse
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	s.Require().NoError(err)
	s.Assert().NotNil(resp.Id)
}

func (s *UsersSuite) TestSpecialCharactersInInput() {
	tests := []struct {
		name     string
		request  openapi.CreateUserRequest
		expected int
	}{
		{
			name: "special characters in name",
			request: openapi.CreateUserRequest{
				Name:     "John O'Connor-Smith",
				Email:    "john.oconnor@example.com",
				Password: "password123",
			},
			expected: http.StatusCreated,
		},
		{
			name: "unicode characters in name",
			request: openapi.CreateUserRequest{
				Name:     "José María García",
				Email:    "jose@example.com",
				Password: "password123",
			},
			expected: http.StatusCreated,
		},
		{
			name: "emoji in name",
			request: openapi.CreateUserRequest{
				Name:     "John 😀 Doe",
				Email:    "john.emoji@example.com",
				Password: "password123",
			},
			expected: http.StatusCreated,
		},
		{
			name: "SQL injection attempt in name",
			request: openapi.CreateUserRequest{
				Name:     "'; DROP TABLE users; --",
				Email:    "injection@example.com",
				Password: "password123",
			},
			expected: http.StatusCreated, // Should be treated as regular text
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			reqBody, _ := json.Marshal(tt.request)
			req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(reqBody))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			s.HTTPServer.ServeHTTP(rec, req)

			s.Assert().Equal(tt.expected, rec.Code, "Response: %s", rec.Body.String())
		})
	}
}
