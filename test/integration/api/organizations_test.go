//go:build integration
// +build integration

package api_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"simpleservicedesk/generated/openapi"
	"simpleservicedesk/test/integration/shared"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/suite"
)

type OrganizationAPITestSuite struct {
	shared.IntegrationSuite
}

func TestOrganizationAPI(t *testing.T) {
	suite.Run(t, new(OrganizationAPITestSuite))
}

func (s *OrganizationAPITestSuite) TestCreateOrganizationIntegration() {
	tests := []struct {
		name           string
		request        any
		expectedStatus int
		validateID     bool
	}{
		{
			name:           "valid root organization creation",
			request:        shared.TestOrg1.CreateOrganizationRequest(),
			expectedStatus: http.StatusCreated,
			validateID:     true,
		},
		{
			name:           "valid second organization creation",
			request:        shared.TestOrg2.CreateOrganizationRequest(),
			expectedStatus: http.StatusCreated,
			validateID:     true,
		},
		{
			name: "empty name",
			request: openapi.CreateOrganizationRequest{
				Name:   "",
				Domain: func() *string { s := "empty.com"; return &s }(),
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "missing domain",
			request: openapi.CreateOrganizationRequest{
				Name:   "No Domain Org",
				Domain: nil,
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "empty domain",
			request: openapi.CreateOrganizationRequest{
				Name:   "Empty Domain Org",
				Domain: func() *string { s := ""; return &s }(),
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

			req := httptest.NewRequest(http.MethodPost, "/organizations", bytes.NewBuffer(reqBody))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			s.HTTPServer.ServeHTTP(rec, req)

			s.Assert().Equal(tt.expectedStatus, rec.Code, "Response: %s", rec.Body.String())

			if tt.expectedStatus == http.StatusCreated {
				var resp openapi.CreateOrganizationResponse
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

func (s *OrganizationAPITestSuite) TestCreateSubOrganizationIntegration() {
	// First create a parent organization
	parentReq := shared.TestOrg1.CreateOrganizationRequest()
	reqBody, _ := json.Marshal(parentReq)
	req := httptest.NewRequest(http.MethodPost, "/organizations", bytes.NewBuffer(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	s.HTTPServer.ServeHTTP(rec, req)
	s.Require().Equal(http.StatusCreated, rec.Code)

	var parentResp openapi.CreateOrganizationResponse
	err := json.Unmarshal(rec.Body.Bytes(), &parentResp)
	s.Require().NoError(err)
	s.Require().NotNil(parentResp.Id)
	parentID := *parentResp.Id

	tests := []struct {
		name           string
		request        openapi.CreateOrganizationRequest
		expectedStatus int
		validateID     bool
	}{
		{
			name:           "valid sub-organization creation",
			request:        shared.NewSubOrganization("Sub Corp", "sub.example.com", parentID).CreateOrganizationRequest(),
			expectedStatus: http.StatusCreated,
			validateID:     true,
		},
		{
			name:           "sub-organization with non-existent parent",
			request:        shared.NewSubOrganization("Orphan Corp", "orphan.example.com", uuid.New()).CreateOrganizationRequest(),
			expectedStatus: http.StatusInternalServerError, // Parent validation returns error during creation
			validateID:     false,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			reqBody, _ := json.Marshal(tt.request)
			req := httptest.NewRequest(http.MethodPost, "/organizations", bytes.NewBuffer(reqBody))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			s.HTTPServer.ServeHTTP(rec, req)

			s.Assert().Equal(tt.expectedStatus, rec.Code, "Response: %s", rec.Body.String())

			if tt.expectedStatus == http.StatusCreated && tt.validateID {
				var resp openapi.CreateOrganizationResponse
				unmarshalErr := json.Unmarshal(rec.Body.Bytes(), &resp)
				s.Require().NoError(unmarshalErr)
				s.Assert().NotNil(resp.Id)
				s.Assert().NotEqual(uuid.Nil, *resp.Id)
			}
		})
	}
}

func (s *OrganizationAPITestSuite) TestGetOrganizationIntegration() {
	// First create an organization to test getting
	createReq := shared.TestOrg3.CreateOrganizationRequest()
	reqBody, _ := json.Marshal(createReq)
	req := httptest.NewRequest(http.MethodPost, "/organizations", bytes.NewBuffer(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	s.HTTPServer.ServeHTTP(rec, req)
	s.Require().Equal(http.StatusCreated, rec.Code)

	var createResp openapi.CreateOrganizationResponse
	err := json.Unmarshal(rec.Body.Bytes(), &createResp)
	s.Require().NoError(err)
	s.Require().NotNil(createResp.Id)
	createdOrgID := *createResp.Id

	tests := []struct {
		name           string
		orgID          string
		expectedStatus int
		validateResp   bool
	}{
		{
			name:           "get existing organization",
			orgID:          createdOrgID.String(),
			expectedStatus: http.StatusOK,
			validateResp:   true,
		},
		{
			name:           "get non-existent organization",
			orgID:          uuid.New().String(),
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "invalid UUID format",
			orgID:          "invalid-uuid",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			url := fmt.Sprintf("/organizations/%s", tt.orgID)
			testReq := httptest.NewRequest(http.MethodGet, url, nil)
			testRec := httptest.NewRecorder()
			s.HTTPServer.ServeHTTP(testRec, testReq)

			s.Assert().Equal(tt.expectedStatus, testRec.Code, "Response: %s", testRec.Body.String())

			if tt.validateResp && tt.expectedStatus == http.StatusOK {
				var resp openapi.GetOrganizationResponse
				unmarshalErr := json.Unmarshal(testRec.Body.Bytes(), &resp)
				s.Require().NoError(unmarshalErr)
				s.Assert().NotNil(resp.Id)
				s.Assert().Equal(createdOrgID, *resp.Id)
				s.Assert().NotNil(resp.Name)
				s.Assert().Equal(shared.TestOrg3.Name, *resp.Name)
				s.Assert().NotNil(resp.Domain)
				s.Assert().Equal(shared.TestOrg3.Domain, *resp.Domain)
				s.Assert().NotNil(resp.IsActive)
				s.Assert().True(*resp.IsActive)
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

func (s *OrganizationAPITestSuite) TestListOrganizationsIntegration() {
	// Create multiple organizations for testing
	orgRequests := []openapi.CreateOrganizationRequest{
		shared.TestOrg1.CreateOrganizationRequest(),
		shared.TestOrg2.CreateOrganizationRequest(),
	}

	createdOrgIDs := make([]uuid.UUID, 0, len(orgRequests))
	for _, orgReq := range orgRequests {
		reqBody, _ := json.Marshal(orgReq)
		req := httptest.NewRequest(http.MethodPost, "/organizations", bytes.NewBuffer(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		s.HTTPServer.ServeHTTP(rec, req)
		s.Require().Equal(http.StatusCreated, rec.Code)

		var createResp openapi.CreateOrganizationResponse
		err := json.Unmarshal(rec.Body.Bytes(), &createResp)
		s.Require().NoError(err)
		s.Require().NotNil(createResp.Id)
		createdOrgIDs = append(createdOrgIDs, *createResp.Id)
	}

	tests := []struct {
		name           string
		queryParams    string
		expectedStatus int
		minResults     int
	}{
		{
			name:           "list all organizations",
			queryParams:    "",
			expectedStatus: http.StatusOK,
			minResults:     2,
		},
		{
			name:           "list with limit",
			queryParams:    "?limit=1",
			expectedStatus: http.StatusOK,
			minResults:     0, // May return 0 due to pagination
		},
		{
			name:           "list with pagination",
			queryParams:    "?page=1&limit=10",
			expectedStatus: http.StatusOK,
			minResults:     0,
		},
		{
			name:           "list with large limit",
			queryParams:    "?limit=200",
			expectedStatus: http.StatusOK,
			minResults:     0,
		},
		{
			name:           "list with zero page",
			queryParams:    "?page=0",
			expectedStatus: http.StatusOK,
			minResults:     0,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			url := "/organizations" + tt.queryParams
			req := httptest.NewRequest(http.MethodGet, url, nil)
			rec := httptest.NewRecorder()
			s.HTTPServer.ServeHTTP(rec, req)

			s.Assert().Equal(tt.expectedStatus, rec.Code, "Response: %s", rec.Body.String())

			if tt.expectedStatus == http.StatusOK {
				var resp openapi.ListOrganizationsResponse
				unmarshalErr := json.Unmarshal(rec.Body.Bytes(), &resp)
				s.Require().NoError(unmarshalErr)
				s.Assert().NotNil(resp.Organizations)
				if tt.minResults > 0 {
					s.Assert().GreaterOrEqual(len(*resp.Organizations), tt.minResults)
				}
			}
		})
	}
}

func (s *OrganizationAPITestSuite) TestUpdateOrganizationIntegration() {
	// Create an organization to update
	createReq := shared.TestOrg1.CreateOrganizationRequest()
	reqBody, _ := json.Marshal(createReq)
	req := httptest.NewRequest(http.MethodPost, "/organizations", bytes.NewBuffer(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	s.HTTPServer.ServeHTTP(rec, req)
	s.Require().Equal(http.StatusCreated, rec.Code)

	var createResp openapi.CreateOrganizationResponse
	err := json.Unmarshal(rec.Body.Bytes(), &createResp)
	s.Require().NoError(err)
	s.Require().NotNil(createResp.Id)
	createdOrgID := *createResp.Id

	tests := []struct {
		name           string
		orgID          string
		request        any
		expectedStatus int
		validateResp   bool
	}{
		{
			name:  "update organization name",
			orgID: createdOrgID.String(),
			request: openapi.UpdateOrganizationRequest{
				Name: func() *string { s := "Updated Corp"; return &s }(),
			},
			expectedStatus: http.StatusOK,
			validateResp:   true,
		},
		{
			name:  "update organization domain",
			orgID: createdOrgID.String(),
			request: openapi.UpdateOrganizationRequest{
				Domain: func() *string { s := "updated.com"; return &s }(),
			},
			expectedStatus: http.StatusOK,
			validateResp:   true,
		},
		{
			name:  "update organization status",
			orgID: createdOrgID.String(),
			request: openapi.UpdateOrganizationRequest{
				IsActive: func() *bool { b := false; return &b }(),
			},
			expectedStatus: http.StatusOK,
			validateResp:   true,
		},
		{
			name:  "update non-existent organization",
			orgID: uuid.New().String(),
			request: openapi.UpdateOrganizationRequest{
				Name: func() *string { s := "Non-existent"; return &s }(),
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "update with invalid UUID",
			orgID:          "invalid-uuid",
			request:        openapi.UpdateOrganizationRequest{},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "update with empty request",
			orgID:          createdOrgID.String(),
			request:        openapi.UpdateOrganizationRequest{},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			reqBody, err := json.Marshal(tt.request)
			s.Require().NoError(err)

			url := fmt.Sprintf("/organizations/%s", tt.orgID)
			req := httptest.NewRequest(http.MethodPut, url, bytes.NewBuffer(reqBody))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			s.HTTPServer.ServeHTTP(rec, req)

			s.Assert().Equal(tt.expectedStatus, rec.Code, "Response: %s", rec.Body.String())

			if tt.validateResp && tt.expectedStatus == http.StatusOK {
				var resp openapi.GetOrganizationResponse
				unmarshalErr := json.Unmarshal(rec.Body.Bytes(), &resp)
				s.Require().NoError(unmarshalErr)
				s.Assert().NotNil(resp.Id)
				s.Assert().Equal(createdOrgID, *resp.Id)
			} else if tt.expectedStatus != http.StatusOK {
				var errorResp openapi.ErrorResponse
				unmarshalErr := json.Unmarshal(rec.Body.Bytes(), &errorResp)
				s.Require().NoError(unmarshalErr)
				s.Assert().NotNil(errorResp.Message)
				s.Assert().NotEmpty(*errorResp.Message)
			}
		})
	}
}

func (s *OrganizationAPITestSuite) TestDeleteOrganizationIntegration() {
	// Create an organization to delete
	createReq := shared.TestOrg1.CreateOrganizationRequest()
	reqBody, _ := json.Marshal(createReq)
	req := httptest.NewRequest(http.MethodPost, "/organizations", bytes.NewBuffer(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	s.HTTPServer.ServeHTTP(rec, req)
	s.Require().Equal(http.StatusCreated, rec.Code)

	var createResp openapi.CreateOrganizationResponse
	err := json.Unmarshal(rec.Body.Bytes(), &createResp)
	s.Require().NoError(err)
	s.Require().NotNil(createResp.Id)
	createdOrgID := *createResp.Id

	tests := []struct {
		name           string
		orgID          string
		expectedStatus int
	}{
		{
			name:           "delete existing organization",
			orgID:          createdOrgID.String(),
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "delete already deleted organization",
			orgID:          createdOrgID.String(),
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "delete non-existent organization",
			orgID:          uuid.New().String(),
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "delete with invalid UUID",
			orgID:          "invalid-uuid",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			url := fmt.Sprintf("/organizations/%s", tt.orgID)
			req := httptest.NewRequest(http.MethodDelete, url, nil)
			rec := httptest.NewRecorder()
			s.HTTPServer.ServeHTTP(rec, req)

			s.Assert().Equal(tt.expectedStatus, rec.Code, "Response: %s", rec.Body.String())

			if tt.expectedStatus == http.StatusNoContent {
				s.Assert().Empty(rec.Body.String())
			} else if tt.expectedStatus != http.StatusNoContent {
				var errorResp openapi.ErrorResponse
				unmarshalErr := json.Unmarshal(rec.Body.Bytes(), &errorResp)
				s.Require().NoError(unmarshalErr)
				s.Assert().NotNil(errorResp.Message)
				s.Assert().NotEmpty(*errorResp.Message)
			}
		})
	}
}

func (s *OrganizationAPITestSuite) TestSpecialCharactersInOrganization() {
	tests := []struct {
		name     string
		request  openapi.CreateOrganizationRequest
		expected int
	}{
		{
			name: "special characters in name",
			request: openapi.CreateOrganizationRequest{
				Name:   "O'Reilly & Associates",
				Domain: func() *string { s := "oreilly.com"; return &s }(),
			},
			expected: http.StatusCreated,
		},
		{
			name: "unicode characters in name",
			request: openapi.CreateOrganizationRequest{
				Name:   "Société Générale",
				Domain: func() *string { s := "socgen.com"; return &s }(),
			},
			expected: http.StatusCreated,
		},
		{
			name: "emoji in name",
			request: openapi.CreateOrganizationRequest{
				Name:   "Tech Corp 🚀",
				Domain: func() *string { s := "techcorp.io"; return &s }(),
			},
			expected: http.StatusCreated,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			reqBody, _ := json.Marshal(tt.request)
			req := httptest.NewRequest(http.MethodPost, "/organizations", bytes.NewBuffer(reqBody))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			s.HTTPServer.ServeHTTP(rec, req)

			s.Assert().Equal(tt.expected, rec.Code, "Response: %s", rec.Body.String())
		})
	}
}

func (s *OrganizationAPITestSuite) TestLargePayloadHandling() {
	tests := []struct {
		name           string
		nameLength     int
		expectedStatus int
	}{
		{
			name:           "valid long name within limit",
			nameLength:     90, // Within 100 character limit
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "name exceeds maximum length",
			nameLength:     150, // Exceeds 100 character limit
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			largeString := strings.Repeat("A", tt.nameLength)
			orgReq := openapi.CreateOrganizationRequest{
				Name:   largeString,
				Domain: func() *string { s := "largepayload.com"; return &s }(),
			}
			reqBody, _ := json.Marshal(orgReq)

			req := httptest.NewRequest(http.MethodPost, "/organizations", bytes.NewBuffer(reqBody))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			s.HTTPServer.ServeHTTP(rec, req)

			s.Assert().Equal(tt.expectedStatus, rec.Code, "Response: %s", rec.Body.String())

			if tt.expectedStatus == http.StatusCreated {
				var resp openapi.CreateOrganizationResponse
				err := json.Unmarshal(rec.Body.Bytes(), &resp)
				s.Require().NoError(err)
				s.Assert().NotNil(resp.Id)
			}
		})
	}
}

func (s *OrganizationAPITestSuite) TestNotImplementedEndpoints() {
	// Create an organization first
	createReq := shared.TestOrg1.CreateOrganizationRequest()
	reqBody, _ := json.Marshal(createReq)
	req := httptest.NewRequest(http.MethodPost, "/organizations", bytes.NewBuffer(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	s.HTTPServer.ServeHTTP(rec, req)
	s.Require().Equal(http.StatusCreated, rec.Code)

	var createResp openapi.CreateOrganizationResponse
	err := json.Unmarshal(rec.Body.Bytes(), &createResp)
	s.Require().NoError(err)
	createdOrgID := *createResp.Id

	tests := []struct {
		name           string
		method         string
		path           string
		expectedStatus int
	}{
		{
			name:           "get organization users - not implemented",
			method:         http.MethodGet,
			path:           fmt.Sprintf("/organizations/%s/users", createdOrgID.String()),
			expectedStatus: http.StatusNotImplemented,
		},
		{
			name:           "get organization tickets - not implemented",
			method:         http.MethodGet,
			path:           fmt.Sprintf("/organizations/%s/tickets", createdOrgID.String()),
			expectedStatus: http.StatusNotImplemented,
		},
		{
			name:           "get organization hierarchy - endpoint not found",
			method:         http.MethodGet,
			path:           fmt.Sprintf("/organizations/%s/hierarchy", createdOrgID.String()),
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			rec := httptest.NewRecorder()
			s.HTTPServer.ServeHTTP(rec, req)

			s.Assert().Equal(tt.expectedStatus, rec.Code, "Response: %s", rec.Body.String())
		})
	}
}
