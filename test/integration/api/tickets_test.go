//go:build integration
// +build integration

package api_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"simpleservicedesk/generated/openapi"
	"simpleservicedesk/test/integration/shared"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/suite"
)

type TicketAPITestSuite struct {
	shared.IntegrationSuite
}

func TestTicketAPI(t *testing.T) {
	suite.Run(t, new(TicketAPITestSuite))
}

func (s *TicketAPITestSuite) TestCreateTicketIntegration() {
	// Create organization and author for the ticket
	orgID := uuid.New()
	authorID := uuid.New()

	tests := []struct {
		name           string
		request        openapi.CreateTicketRequest
		expectedStatus int
		expectedError  *string
		validateID     bool
	}{
		{
			name: "valid ticket creation",
			request: openapi.CreateTicketRequest{
				Title:          "Integration Test Ticket",
				Description:    "This is an integration test ticket",
				Priority:       openapi.TicketPriority("normal"),
				OrganizationId: orgID,
				AuthorId:       authorID,
			},
			expectedStatus: http.StatusCreated,
			validateID:     true,
		},
		{
			name: "ticket with category",
			request: openapi.CreateTicketRequest{
				Title:          "Ticket with Category",
				Description:    "This ticket has a category",
				Priority:       openapi.TicketPriority("high"),
				OrganizationId: orgID,
				AuthorId:       authorID,
				CategoryId:     func() *uuid.UUID { id := uuid.New(); return &id }(),
			},
			expectedStatus: http.StatusCreated,
			validateID:     true,
		},
		{
			name: "empty title",
			request: openapi.CreateTicketRequest{
				Title:          "",
				Description:    "This ticket has no title",
				Priority:       openapi.TicketPriority("normal"),
				OrganizationId: orgID,
				AuthorId:       authorID,
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  func() *string { s := "title too short"; return &s }(),
		},
		{
			name: "invalid priority",
			request: openapi.CreateTicketRequest{
				Title:          "Invalid Priority Ticket",
				Description:    "This ticket has invalid priority",
				Priority:       openapi.TicketPriority("invalid"),
				OrganizationId: orgID,
				AuthorId:       authorID,
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  func() *string { s := "invalid priority"; return &s }(),
		},
		{
			name: "empty description is valid",
			request: openapi.CreateTicketRequest{
				Title:          "No Description Ticket",
				Description:    "",
				Priority:       openapi.TicketPriority("low"),
				OrganizationId: orgID,
				AuthorId:       authorID,
			},
			expectedStatus: http.StatusCreated,
			validateID:     true,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			reqBody, err := json.Marshal(tt.request)
			s.Require().NoError(err)

			req := httptest.NewRequest(http.MethodPost, "/tickets", bytes.NewBuffer(reqBody))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()

			s.HTTPServer.ServeHTTP(rec, req)
			s.Equal(tt.expectedStatus, rec.Code)

			if tt.expectedStatus == http.StatusCreated && tt.validateID {
				var resp openapi.GetTicketResponse
				err := json.Unmarshal(rec.Body.Bytes(), &resp)
				s.Require().NoError(err)
				s.NotNil(resp.Id)
				s.NotEqual(uuid.Nil, *resp.Id)
			}

			if tt.expectedError != nil {
				var errorResp openapi.ErrorResponse
				err := json.Unmarshal(rec.Body.Bytes(), &errorResp)
				s.Require().NoError(err)
				s.Contains(*errorResp.Message, *tt.expectedError)
			}
		})
	}
}

func (s *TicketAPITestSuite) TestGetTicketIntegration() {
	// Create a ticket first
	orgID := uuid.New()
	authorID := uuid.New()
	categoryID := uuid.New()

	createReq := openapi.CreateTicketRequest{
		Title:          "Get Test Ticket",
		Description:    "This ticket is for get testing",
		Priority:       openapi.TicketPriority("critical"),
		OrganizationId: orgID,
		AuthorId:       authorID,
		CategoryId:     &categoryID,
	}

	reqBody, err := json.Marshal(createReq)
	s.Require().NoError(err)

	req := httptest.NewRequest(http.MethodPost, "/tickets", bytes.NewBuffer(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	s.HTTPServer.ServeHTTP(rec, req)
	s.Require().Equal(http.StatusCreated, rec.Code)

	var createResp openapi.GetTicketResponse
	err = json.Unmarshal(rec.Body.Bytes(), &createResp)
	s.Require().NoError(err)
	s.Require().NotNil(createResp.Id)

	ticketID := *createResp.Id

	s.Run("get existing ticket", func() {
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/tickets/%s", ticketID.String()), nil)
		rec := httptest.NewRecorder()
		s.HTTPServer.ServeHTTP(rec, req)

		s.Equal(http.StatusOK, rec.Code)

		var getResp openapi.GetTicketResponse
		err := json.Unmarshal(rec.Body.Bytes(), &getResp)
		s.Require().NoError(err)

		// Verify all fields
		s.NotNil(getResp.Id)
		s.Equal(ticketID, *getResp.Id)
		s.NotNil(getResp.Title)
		s.Equal("Get Test Ticket", *getResp.Title)
		s.NotNil(getResp.Description)
		s.Equal("This ticket is for get testing", *getResp.Description)
		s.NotNil(getResp.Priority)
		s.Equal(openapi.TicketPriority("critical"), *getResp.Priority)
		s.NotNil(getResp.Status)
		s.Equal(openapi.TicketStatus("new"), *getResp.Status)
		s.NotNil(getResp.OrganizationId)
		s.Equal(orgID, *getResp.OrganizationId)
		s.NotNil(getResp.AuthorId)
		s.Equal(authorID, *getResp.AuthorId)
		s.NotNil(getResp.CategoryId)
		s.Equal(categoryID, *getResp.CategoryId)
		s.NotNil(getResp.CreatedAt)
		s.NotNil(getResp.UpdatedAt)
	})

	s.Run("get non-existent ticket", func() {
		nonExistentID := uuid.New()
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/tickets/%s", nonExistentID.String()), nil)
		rec := httptest.NewRecorder()
		s.HTTPServer.ServeHTTP(rec, req)

		s.Equal(http.StatusNotFound, rec.Code)
	})

	s.Run("get ticket with invalid ID", func() {
		req := httptest.NewRequest(http.MethodGet, "/tickets/invalid-uuid", nil)
		rec := httptest.NewRecorder()
		s.HTTPServer.ServeHTTP(rec, req)

		s.Equal(http.StatusBadRequest, rec.Code)
	})
}

func (s *TicketAPITestSuite) TestListTicketsIntegration() {
	// Create multiple tickets for testing
	orgID := uuid.New()
	authorID := uuid.New()

	tickets := []openapi.CreateTicketRequest{
		{
			Title:          "List Test Ticket 1",
			Description:    "First ticket for list testing",
			Priority:       openapi.TicketPriority("low"),
			OrganizationId: orgID,
			AuthorId:       authorID,
		},
		{
			Title:          "List Test Ticket 2",
			Description:    "Second ticket for list testing",
			Priority:       openapi.TicketPriority("high"),
			OrganizationId: orgID,
			AuthorId:       authorID,
		},
		{
			Title:          "List Test Ticket 3",
			Description:    "Third ticket for list testing",
			Priority:       openapi.TicketPriority("normal"),
			OrganizationId: orgID,
			AuthorId:       authorID,
		},
	}

	// Create the tickets
	for _, ticket := range tickets {
		reqBody, err := json.Marshal(ticket)
		s.Require().NoError(err)

		req := httptest.NewRequest(http.MethodPost, "/tickets", bytes.NewBuffer(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		s.HTTPServer.ServeHTTP(rec, req)
		s.Require().Equal(http.StatusCreated, rec.Code)
	}

	s.Run("list tickets with default pagination", func() {
		req := httptest.NewRequest(http.MethodGet, "/tickets", nil)
		rec := httptest.NewRecorder()
		s.HTTPServer.ServeHTTP(rec, req)

		s.Equal(http.StatusOK, rec.Code)

		var listResp openapi.ListTicketsResponse
		err := json.Unmarshal(rec.Body.Bytes(), &listResp)
		s.Require().NoError(err)

		s.NotNil(listResp.Tickets)
		s.GreaterOrEqual(len(*listResp.Tickets), 3) // At least our 3 test tickets
		s.NotNil(listResp.Pagination)
	})

	s.Run("list tickets with custom limit", func() {
		req := httptest.NewRequest(http.MethodGet, "/tickets?limit=2", nil)
		rec := httptest.NewRecorder()
		s.HTTPServer.ServeHTTP(rec, req)

		s.Equal(http.StatusOK, rec.Code)

		var listResp openapi.ListTicketsResponse
		err := json.Unmarshal(rec.Body.Bytes(), &listResp)
		s.Require().NoError(err)

		s.NotNil(listResp.Tickets)
		s.NotNil(listResp.Pagination)
		s.NotNil(listResp.Pagination.Limit)
		s.Equal(2, *listResp.Pagination.Limit)
	})

	s.Run("list tickets with status filter", func() {
		req := httptest.NewRequest(http.MethodGet, "/tickets?status=new", nil)
		rec := httptest.NewRecorder()
		s.HTTPServer.ServeHTTP(rec, req)

		s.Equal(http.StatusOK, rec.Code)

		var listResp openapi.ListTicketsResponse
		err := json.Unmarshal(rec.Body.Bytes(), &listResp)
		s.Require().NoError(err)

		s.NotNil(listResp.Tickets)
		// All created tickets should have "new" status
		for _, ticket := range *listResp.Tickets {
			s.NotNil(ticket.Status)
			s.Equal(openapi.TicketStatus("new"), *ticket.Status)
		}
	})

	s.Run("list tickets with priority filter", func() {
		// Use organization filter to limit results to our test tickets
		req := httptest.NewRequest(http.MethodGet, "/tickets?priority=high&organization_id="+orgID.String(), nil)
		rec := httptest.NewRecorder()
		s.HTTPServer.ServeHTTP(rec, req)

		s.Equal(http.StatusOK, rec.Code)

		var listResp openapi.ListTicketsResponse
		err := json.Unmarshal(rec.Body.Bytes(), &listResp)
		s.Require().NoError(err)

		s.NotNil(listResp.Tickets)
		// We should get exactly 1 ticket with "high" priority from our test set
		s.Len(*listResp.Tickets, 1)
		ticket := (*listResp.Tickets)[0]
		s.NotNil(ticket.Priority)
		s.Equal(openapi.TicketPriority("high"), *ticket.Priority)
		s.Equal(orgID, *ticket.OrganizationId)
	})

	s.Run("list tickets with organization filter", func() {
		req := httptest.NewRequest(http.MethodGet, "/tickets?organization_id="+orgID.String(), nil)
		rec := httptest.NewRecorder()
		s.HTTPServer.ServeHTTP(rec, req)

		s.Equal(http.StatusOK, rec.Code)

		var listResp openapi.ListTicketsResponse
		err := json.Unmarshal(rec.Body.Bytes(), &listResp)
		s.Require().NoError(err)

		s.NotNil(listResp.Tickets)
		// All returned tickets should belong to our organization
		for _, ticket := range *listResp.Tickets {
			s.NotNil(ticket.OrganizationId)
			s.Equal(orgID, *ticket.OrganizationId)
		}
	})

	s.Run("invalid status filter", func() {
		req := httptest.NewRequest(http.MethodGet, "/tickets?status=invalid", nil)
		rec := httptest.NewRecorder()
		s.HTTPServer.ServeHTTP(rec, req)

		s.Equal(http.StatusBadRequest, rec.Code)
	})

	s.Run("invalid priority filter", func() {
		req := httptest.NewRequest(http.MethodGet, "/tickets?priority=invalid", nil)
		rec := httptest.NewRecorder()
		s.HTTPServer.ServeHTTP(rec, req)

		s.Equal(http.StatusBadRequest, rec.Code)
	})

	s.Run("invalid pagination parameters", func() {
		// Test invalid page
		req := httptest.NewRequest(http.MethodGet, "/tickets?page=0", nil)
		rec := httptest.NewRecorder()
		s.HTTPServer.ServeHTTP(rec, req)
		s.Equal(http.StatusBadRequest, rec.Code)

		// Test invalid limit
		req = httptest.NewRequest(http.MethodGet, "/tickets?limit=101", nil)
		rec = httptest.NewRecorder()
		s.HTTPServer.ServeHTTP(rec, req)
		s.Equal(http.StatusBadRequest, rec.Code)

		// Test zero limit
		req = httptest.NewRequest(http.MethodGet, "/tickets?limit=0", nil)
		rec = httptest.NewRecorder()
		s.HTTPServer.ServeHTTP(rec, req)
		s.Equal(http.StatusBadRequest, rec.Code)
	})
}

func (s *TicketAPITestSuite) TestTicketPriorityLevels() {
	orgID := uuid.New()
	authorID := uuid.New()

	priorities := []openapi.TicketPriority{"low", "normal", "high", "critical"}

	for _, priority := range priorities {
		s.Run(fmt.Sprintf("create ticket with %s priority", priority), func() {
			createReq := openapi.CreateTicketRequest{
				Title:          fmt.Sprintf("Priority Test - %s", priority),
				Description:    fmt.Sprintf("Testing %s priority ticket", priority),
				Priority:       priority,
				OrganizationId: orgID,
				AuthorId:       authorID,
			}

			reqBody, err := json.Marshal(createReq)
			s.Require().NoError(err)

			req := httptest.NewRequest(http.MethodPost, "/tickets", bytes.NewBuffer(reqBody))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			s.HTTPServer.ServeHTTP(rec, req)

			s.Equal(http.StatusCreated, rec.Code)

			var createResp openapi.GetTicketResponse
			err = json.Unmarshal(rec.Body.Bytes(), &createResp)
			s.Require().NoError(err)
			s.NotNil(createResp.Id)

			// Verify by getting the ticket
			req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/tickets/%s", createResp.Id.String()), nil)
			rec = httptest.NewRecorder()
			s.HTTPServer.ServeHTTP(rec, req)

			s.Equal(http.StatusOK, rec.Code)

			var getResp openapi.GetTicketResponse
			err = json.Unmarshal(rec.Body.Bytes(), &getResp)
			s.Require().NoError(err)
			s.NotNil(getResp.Priority)
			s.Equal(priority, *getResp.Priority)
		})
	}
}

func (s *TicketAPITestSuite) TestTicketValidationEdgeCases() {
	orgID := uuid.New()
	authorID := uuid.New()

	s.Run("title length validation", func() {
		// Test minimum title length (should be at least 3 characters)
		createReq := openapi.CreateTicketRequest{
			Title:          "AB", // Too short
			Description:    "Testing title validation",
			Priority:       openapi.TicketPriority("normal"),
			OrganizationId: orgID,
			AuthorId:       authorID,
		}

		reqBody, err := json.Marshal(createReq)
		s.Require().NoError(err)

		req := httptest.NewRequest(http.MethodPost, "/tickets", bytes.NewBuffer(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		s.HTTPServer.ServeHTTP(rec, req)

		s.Equal(http.StatusBadRequest, rec.Code)
	})

	s.Run("valid minimum title length", func() {
		createReq := openapi.CreateTicketRequest{
			Title:          "ABC", // Minimum valid length
			Description:    "Testing minimum title validation",
			Priority:       openapi.TicketPriority("normal"),
			OrganizationId: orgID,
			AuthorId:       authorID,
		}

		reqBody, err := json.Marshal(createReq)
		s.Require().NoError(err)

		req := httptest.NewRequest(http.MethodPost, "/tickets", bytes.NewBuffer(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		s.HTTPServer.ServeHTTP(rec, req)

		s.Equal(http.StatusCreated, rec.Code)
	})
}
