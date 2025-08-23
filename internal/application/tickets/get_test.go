package tickets_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"

	"simpleservicedesk/generated/openapi"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func (s *TicketsSuite) TestGetTicket() {
	s.Run("Get existing ticket", func() {
		orgID := uuid.New()
		authorID := uuid.New()

		ticketReq := openapi.CreateTicketRequest{
			Title:          "Test Ticket",
			Description:    "This is a test ticket",
			Priority:       openapi.TicketPriority("normal"),
			OrganizationId: orgID,
			AuthorId:       authorID,
		}
		reqBody, _ := json.Marshal(ticketReq)

		req := httptest.NewRequest(http.MethodPost, "/tickets", bytes.NewBuffer(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		s.HTTPServer.ServeHTTP(rec, req)

		s.Require().Equal(http.StatusCreated, rec.Code)
		var createResp openapi.GetTicketResponse
		err := json.Unmarshal(rec.Body.Bytes(), &createResp)
		s.Require().NoError(err)

		req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/tickets/%s", createResp.Id.String()), nil)
		rec = httptest.NewRecorder()
		s.HTTPServer.ServeHTTP(rec, req)

		s.Require().Equal(http.StatusOK, rec.Code)
		var getResp openapi.GetTicketResponse
		err = json.Unmarshal(rec.Body.Bytes(), &getResp)
		s.Require().NoError(err)
		s.Require().NotNil(getResp.Id)
		s.Require().Equal(*createResp.Id, *getResp.Id)
		s.Require().Equal("Test Ticket", *getResp.Title)
		s.Require().Equal("This is a test ticket", *getResp.Description)
		s.Require().Equal(openapi.TicketPriority("normal"), *getResp.Priority)
		s.Require().Equal(openapi.TicketStatus("new"), *getResp.Status)
	})

	s.Run("Get ticket with category", func() {
		orgID := uuid.New()
		authorID := uuid.New()
		categoryID := uuid.New()

		ticketReq := openapi.CreateTicketRequest{
			Title:          "Test Ticket with Category",
			Description:    "This is a test ticket with category",
			Priority:       openapi.TicketPriority("high"),
			OrganizationId: orgID,
			AuthorId:       authorID,
			CategoryId:     &categoryID,
		}
		reqBody, _ := json.Marshal(ticketReq)

		req := httptest.NewRequest(http.MethodPost, "/tickets", bytes.NewBuffer(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		s.HTTPServer.ServeHTTP(rec, req)

		s.Require().Equal(http.StatusCreated, rec.Code)
		var createResp openapi.GetTicketResponse
		err := json.Unmarshal(rec.Body.Bytes(), &createResp)
		s.Require().NoError(err)

		req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/tickets/%s", createResp.Id.String()), nil)
		rec = httptest.NewRecorder()
		s.HTTPServer.ServeHTTP(rec, req)

		s.Require().Equal(http.StatusOK, rec.Code)
		var getResp openapi.GetTicketResponse
		err = json.Unmarshal(rec.Body.Bytes(), &getResp)
		s.Require().NoError(err)
		s.Require().NotNil(getResp.CategoryId)
		s.Require().Equal(categoryID, *getResp.CategoryId)
	})

	s.Run("Get non-existent ticket returns 404", func() {
		nonExistentID := uuid.New()
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/tickets/%s", nonExistentID.String()), nil)
		rec := httptest.NewRecorder()
		s.HTTPServer.ServeHTTP(rec, req)

		s.Require().Equal(http.StatusNotFound, rec.Code)
	})

	s.Run("Get ticket with invalid ID returns 400", func() {
		req := httptest.NewRequest(http.MethodGet, "/tickets/invalid-uuid", nil)
		rec := httptest.NewRecorder()
		s.HTTPServer.ServeHTTP(rec, req)

		s.Require().Equal(http.StatusBadRequest, rec.Code)
	})

	s.Run("Verify ticket response structure", func() {
		orgID := uuid.New()
		authorID := uuid.New()

		ticketReq := openapi.CreateTicketRequest{
			Title:          "Structured Ticket",
			Description:    "Testing response structure",
			Priority:       openapi.TicketPriority("critical"),
			OrganizationId: orgID,
			AuthorId:       authorID,
		}
		reqBody, _ := json.Marshal(ticketReq)

		req := httptest.NewRequest(http.MethodPost, "/tickets", bytes.NewBuffer(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		s.HTTPServer.ServeHTTP(rec, req)

		s.Require().Equal(http.StatusCreated, rec.Code)
		var createResp openapi.GetTicketResponse
		err := json.Unmarshal(rec.Body.Bytes(), &createResp)
		s.Require().NoError(err)

		req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/tickets/%s", createResp.Id.String()), nil)
		rec = httptest.NewRecorder()
		s.HTTPServer.ServeHTTP(rec, req)

		s.Require().Equal(http.StatusOK, rec.Code)
		var getResp openapi.GetTicketResponse
		err = json.Unmarshal(rec.Body.Bytes(), &getResp)
		s.Require().NoError(err)

		// Verify all required fields are present
		s.Require().NotNil(getResp.Id)
		s.Require().NotNil(getResp.Title)
		s.Require().NotNil(getResp.Description)
		s.Require().NotNil(getResp.Status)
		s.Require().NotNil(getResp.Priority)
		s.Require().NotNil(getResp.OrganizationId)
		s.Require().NotNil(getResp.AuthorId)
		s.Require().NotNil(getResp.CreatedAt)
		s.Require().NotNil(getResp.UpdatedAt)

		// Verify values
		s.Require().Equal(orgID, *getResp.OrganizationId)
		s.Require().Equal(authorID, *getResp.AuthorId)
		s.Require().Equal("critical", string(*getResp.Priority))
		s.Require().Equal("new", string(*getResp.Status))
	})
}
