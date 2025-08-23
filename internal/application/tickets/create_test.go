package tickets_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"

	"simpleservicedesk/generated/openapi"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func (s *TicketsSuite) TestCreateTicket() {
	s.Run("Create valid ticket", func() {
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
		var resp openapi.CreateTicketResponse
		err := json.Unmarshal(rec.Body.Bytes(), &resp)
		s.Require().NoError(err)
		s.Require().NotNil(resp.Id)
		s.Require().NotEqual(uuid.Nil, *resp.Id)
	})

	s.Run("Create ticket with category", func() {
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
		var resp openapi.CreateTicketResponse
		err := json.Unmarshal(rec.Body.Bytes(), &resp)
		s.Require().NoError(err)
		s.Require().NotNil(resp.Id)
		s.Require().NotEqual(uuid.Nil, *resp.Id)
	})

	s.Run("Missing title returns 400", func() {
		orgID := uuid.New()
		authorID := uuid.New()

		ticketReq := openapi.CreateTicketRequest{
			Title:          "",
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

		s.Require().Equal(http.StatusBadRequest, rec.Code)
	})

	s.Run("Invalid priority returns 400", func() {
		orgID := uuid.New()
		authorID := uuid.New()

		ticketReq := openapi.CreateTicketRequest{
			Title:          "Test Ticket",
			Description:    "This is a test ticket",
			Priority:       openapi.TicketPriority("invalid"),
			OrganizationId: orgID,
			AuthorId:       authorID,
		}
		reqBody, _ := json.Marshal(ticketReq)

		req := httptest.NewRequest(http.MethodPost, "/tickets", bytes.NewBuffer(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		s.HTTPServer.ServeHTTP(rec, req)

		s.Require().Equal(http.StatusBadRequest, rec.Code)
	})

	s.Run("Empty description is valid", func() {
		orgID := uuid.New()
		authorID := uuid.New()

		ticketReq := openapi.CreateTicketRequest{
			Title:          "Test Ticket",
			Description:    "",
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
		var resp openapi.CreateTicketResponse
		err := json.Unmarshal(rec.Body.Bytes(), &resp)
		s.Require().NoError(err)
		s.Require().NotNil(resp.Id)
	})

	s.Run("Invalid JSON returns 400", func() {
		req := httptest.NewRequest(http.MethodPost, "/tickets", bytes.NewBufferString(`{"invalid": json}`))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		s.HTTPServer.ServeHTTP(rec, req)

		s.Require().Equal(http.StatusBadRequest, rec.Code)
	})

	s.Run("Different priority levels", func() {
		priorities := []openapi.TicketPriority{"low", "normal", "high", "critical"}

		for _, priority := range priorities {
			orgID := uuid.New()
			authorID := uuid.New()

			ticketReq := openapi.CreateTicketRequest{
				Title:          "Test Ticket - " + string(priority),
				Description:    "This is a test ticket with " + string(priority) + " priority",
				Priority:       priority,
				OrganizationId: orgID,
				AuthorId:       authorID,
			}
			reqBody, _ := json.Marshal(ticketReq)

			req := httptest.NewRequest(http.MethodPost, "/tickets", bytes.NewBuffer(reqBody))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			s.HTTPServer.ServeHTTP(rec, req)

			s.Require().Equal(http.StatusCreated, rec.Code, "Failed for priority: "+string(priority))
		}
	})
}
