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

func (s *TicketsSuite) TestListTickets() {
	s.Run("List tickets with default pagination", func() {
		// Create test tickets
		orgID := uuid.New()
		authorID := uuid.New()

		ticket1 := openapi.CreateTicketRequest{
			Title:          "Test Ticket 1",
			Description:    "This is test ticket 1",
			Priority:       openapi.TicketPriority("normal"),
			OrganizationId: orgID,
			AuthorId:       authorID,
		}
		ticket2 := openapi.CreateTicketRequest{
			Title:          "Test Ticket 2",
			Description:    "This is test ticket 2",
			Priority:       openapi.TicketPriority("high"),
			OrganizationId: orgID,
			AuthorId:       authorID,
		}

		tickets := []openapi.CreateTicketRequest{ticket1, ticket2}

		for _, ticket := range tickets {
			reqBody, _ := json.Marshal(ticket)
			req := httptest.NewRequest(http.MethodPost, "/tickets", bytes.NewBuffer(reqBody))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			s.HTTPServer.ServeHTTP(rec, req)
			s.Require().Equal(http.StatusCreated, rec.Code)
		}

		req := httptest.NewRequest(http.MethodGet, "/tickets", nil)
		rec := httptest.NewRecorder()
		s.HTTPServer.ServeHTTP(rec, req)

		s.Require().Equal(http.StatusOK, rec.Code)
		var listResp openapi.ListTicketsResponse
		err := json.Unmarshal(rec.Body.Bytes(), &listResp)
		s.Require().NoError(err)
		s.Require().NotNil(listResp.Tickets)
		s.Require().Len(*listResp.Tickets, 2)
		s.Require().NotNil(listResp.Pagination)
	})

	s.Run("List tickets with custom limit", func() {
		req := httptest.NewRequest(http.MethodGet, "/tickets?limit=1", nil)
		rec := httptest.NewRecorder()
		s.HTTPServer.ServeHTTP(rec, req)

		s.Require().Equal(http.StatusOK, rec.Code)
		var listResp openapi.ListTicketsResponse
		err := json.Unmarshal(rec.Body.Bytes(), &listResp)
		s.Require().NoError(err)
		s.Require().NotNil(listResp.Tickets)
		s.Require().NotNil(listResp.Pagination)
	})

	s.Run("List tickets with pagination", func() {
		req := httptest.NewRequest(http.MethodGet, "/tickets?page=1&limit=10", nil)
		rec := httptest.NewRecorder()
		s.HTTPServer.ServeHTTP(rec, req)

		s.Require().Equal(http.StatusOK, rec.Code)
		var listResp openapi.ListTicketsResponse
		err := json.Unmarshal(rec.Body.Bytes(), &listResp)
		s.Require().NoError(err)
		s.Require().NotNil(listResp.Tickets)
		s.Require().NotNil(listResp.Pagination)
	})

	s.Run("List tickets with status filter", func() {
		req := httptest.NewRequest(http.MethodGet, "/tickets?status=new", nil)
		rec := httptest.NewRecorder()
		s.HTTPServer.ServeHTTP(rec, req)

		s.Require().Equal(http.StatusOK, rec.Code)
		var listResp openapi.ListTicketsResponse
		err := json.Unmarshal(rec.Body.Bytes(), &listResp)
		s.Require().NoError(err)
		s.Require().NotNil(listResp.Tickets)
	})

	s.Run("List tickets with priority filter", func() {
		req := httptest.NewRequest(http.MethodGet, "/tickets?priority=high", nil)
		rec := httptest.NewRecorder()
		s.HTTPServer.ServeHTTP(rec, req)

		s.Require().Equal(http.StatusOK, rec.Code)
		var listResp openapi.ListTicketsResponse
		err := json.Unmarshal(rec.Body.Bytes(), &listResp)
		s.Require().NoError(err)
		s.Require().NotNil(listResp.Tickets)
	})

	s.Run("List tickets with multiple filters", func() {
		orgID := uuid.New()
		authorID := uuid.New()

		req := httptest.NewRequest(http.MethodGet,
			"/tickets?status=new&priority=normal&organization_id="+orgID.String()+"&author_id="+authorID.String(), nil)
		rec := httptest.NewRecorder()
		s.HTTPServer.ServeHTTP(rec, req)

		s.Require().Equal(http.StatusOK, rec.Code)
		var listResp openapi.ListTicketsResponse
		err := json.Unmarshal(rec.Body.Bytes(), &listResp)
		s.Require().NoError(err)
		s.Require().NotNil(listResp.Tickets)
	})

	s.Run("Invalid status filter returns 400", func() {
		req := httptest.NewRequest(http.MethodGet, "/tickets?status=invalid", nil)
		rec := httptest.NewRecorder()
		s.HTTPServer.ServeHTTP(rec, req)

		s.Require().Equal(http.StatusBadRequest, rec.Code)
	})

	s.Run("Invalid priority filter returns 400", func() {
		req := httptest.NewRequest(http.MethodGet, "/tickets?priority=invalid", nil)
		rec := httptest.NewRecorder()
		s.HTTPServer.ServeHTTP(rec, req)

		s.Require().Equal(http.StatusBadRequest, rec.Code)
	})

	s.Run("Invalid page parameter returns 400", func() {
		req := httptest.NewRequest(http.MethodGet, "/tickets?page=0", nil)
		rec := httptest.NewRecorder()
		s.HTTPServer.ServeHTTP(rec, req)

		s.Require().Equal(http.StatusBadRequest, rec.Code)
	})

	s.Run("Invalid limit parameter returns 400", func() {
		req := httptest.NewRequest(http.MethodGet, "/tickets?limit=101", nil)
		rec := httptest.NewRecorder()
		s.HTTPServer.ServeHTTP(rec, req)

		s.Require().Equal(http.StatusBadRequest, rec.Code)
	})

	s.Run("Zero limit parameter returns 400", func() {
		req := httptest.NewRequest(http.MethodGet, "/tickets?limit=0", nil)
		rec := httptest.NewRecorder()
		s.HTTPServer.ServeHTTP(rec, req)

		s.Require().Equal(http.StatusBadRequest, rec.Code)
	})

	s.Run("List tickets with assignee filter", func() {
		assigneeID := uuid.New()

		req := httptest.NewRequest(http.MethodGet, "/tickets?assignee_id="+assigneeID.String(), nil)
		rec := httptest.NewRecorder()
		s.HTTPServer.ServeHTTP(rec, req)

		s.Require().Equal(http.StatusOK, rec.Code)
		var listResp openapi.ListTicketsResponse
		err := json.Unmarshal(rec.Body.Bytes(), &listResp)
		s.Require().NoError(err)
		s.Require().NotNil(listResp.Tickets)
	})

	s.Run("List tickets with category filter", func() {
		categoryID := uuid.New()

		req := httptest.NewRequest(http.MethodGet, "/tickets?category_id="+categoryID.String(), nil)
		rec := httptest.NewRecorder()
		s.HTTPServer.ServeHTTP(rec, req)

		s.Require().Equal(http.StatusOK, rec.Code)
		var listResp openapi.ListTicketsResponse
		err := json.Unmarshal(rec.Body.Bytes(), &listResp)
		s.Require().NoError(err)
		s.Require().NotNil(listResp.Tickets)
	})

	s.Run("Verify pagination response structure", func() {
		req := httptest.NewRequest(http.MethodGet, "/tickets?page=2&limit=5", nil)
		rec := httptest.NewRecorder()
		s.HTTPServer.ServeHTTP(rec, req)

		s.Require().Equal(http.StatusOK, rec.Code)
		var listResp openapi.ListTicketsResponse
		err := json.Unmarshal(rec.Body.Bytes(), &listResp)
		s.Require().NoError(err)

		s.Require().NotNil(listResp.Pagination)
		s.Require().NotNil(listResp.Pagination.Page)
		s.Require().NotNil(listResp.Pagination.Limit)
		s.Require().NotNil(listResp.Pagination.Total)
		s.Require().NotNil(listResp.Pagination.HasNext)

		s.Require().Equal(2, *listResp.Pagination.Page)
		s.Require().Equal(5, *listResp.Pagination.Limit)
	})
}
