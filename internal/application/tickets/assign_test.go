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

func (s *TicketsSuite) TestAssignTicket() {
	s.Run("Assign ticket to user successfully", func() {
		// Create a test ticket first
		orgID := uuid.New()
		authorID := uuid.New()
		assigneeID := uuid.New()

		ticketReq := openapi.CreateTicketRequest{
			Title:          "Test Ticket for Assignment",
			Description:    "This ticket will be assigned",
			Priority:       openapi.TicketPriority("normal"),
			OrganizationId: orgID,
			AuthorId:       authorID,
		}

		// Create the ticket
		createBody, _ := json.Marshal(ticketReq)
		createReq := httptest.NewRequest(http.MethodPost, "/tickets", bytes.NewBuffer(createBody))
		createReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		createRec := httptest.NewRecorder()

		s.HTTPServer.ServeHTTP(createRec, createReq)
		s.Equal(http.StatusCreated, createRec.Code)

		var createResp openapi.GetTicketResponse
		err := json.Unmarshal(createRec.Body.Bytes(), &createResp)
		s.NoError(err)
		s.NotNil(createResp.Id)

		ticketID := *createResp.Id

		// Now assign the ticket
		assignReq := openapi.AssignTicketRequest{
			AssigneeId: &assigneeID,
		}

		assignBody, _ := json.Marshal(assignReq)
		req := httptest.NewRequest(
			http.MethodPatch,
			fmt.Sprintf("/tickets/%s/assign", ticketID.String()),
			bytes.NewBuffer(assignBody),
		)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		s.HTTPServer.ServeHTTP(rec, req)
		s.Equal(http.StatusOK, rec.Code)

		var resp openapi.GetTicketResponse
		err = json.Unmarshal(rec.Body.Bytes(), &resp)
		s.NoError(err)
		s.NotNil(resp.AssigneeId)
		s.Equal(assigneeID, *resp.AssigneeId)
	})

	s.Run("Unassign ticket successfully", func() {
		// Create a test ticket first
		orgID := uuid.New()
		authorID := uuid.New()
		assigneeID := uuid.New()

		ticketReq := openapi.CreateTicketRequest{
			Title:          "Test Ticket for Unassignment",
			Description:    "This ticket will be unassigned",
			Priority:       openapi.TicketPriority("normal"),
			OrganizationId: orgID,
			AuthorId:       authorID,
		}

		// Create the ticket
		createBody, _ := json.Marshal(ticketReq)
		createReq := httptest.NewRequest(http.MethodPost, "/tickets", bytes.NewBuffer(createBody))
		createReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		createRec := httptest.NewRecorder()

		s.HTTPServer.ServeHTTP(createRec, createReq)
		s.Equal(http.StatusCreated, createRec.Code)

		var createResp openapi.GetTicketResponse
		err := json.Unmarshal(createRec.Body.Bytes(), &createResp)
		s.NoError(err)
		ticketID := *createResp.Id

		// First assign the ticket
		assignReq := openapi.AssignTicketRequest{
			AssigneeId: &assigneeID,
		}

		assignBody, _ := json.Marshal(assignReq)
		assignReqHTTP := httptest.NewRequest(
			http.MethodPatch,
			fmt.Sprintf("/tickets/%s/assign", ticketID.String()),
			bytes.NewBuffer(assignBody),
		)
		assignReqHTTP.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		assignRec := httptest.NewRecorder()

		s.HTTPServer.ServeHTTP(assignRec, assignReqHTTP)
		s.Equal(http.StatusOK, assignRec.Code)

		// Now unassign the ticket
		unassignReq := openapi.AssignTicketRequest{
			AssigneeId: nil,
		}

		unassignBody, _ := json.Marshal(unassignReq)
		req := httptest.NewRequest(
			http.MethodPatch,
			fmt.Sprintf("/tickets/%s/assign", ticketID.String()),
			bytes.NewBuffer(unassignBody),
		)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		s.HTTPServer.ServeHTTP(rec, req)
		s.Equal(http.StatusOK, rec.Code)

		var resp openapi.GetTicketResponse
		err = json.Unmarshal(rec.Body.Bytes(), &resp)
		s.NoError(err)
		s.Nil(resp.AssigneeId)
	})

	s.Run("Assign non-existent ticket returns 404", func() {
		nonExistentID := uuid.New()
		assigneeID := uuid.New()

		assignReq := openapi.AssignTicketRequest{
			AssigneeId: &assigneeID,
		}

		assignBody, _ := json.Marshal(assignReq)
		req := httptest.NewRequest(
			http.MethodPatch,
			fmt.Sprintf("/tickets/%s/assign", nonExistentID.String()),
			bytes.NewBuffer(assignBody),
		)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		s.HTTPServer.ServeHTTP(rec, req)
		s.Equal(http.StatusNotFound, rec.Code)

		var resp openapi.ErrorResponse
		err := json.Unmarshal(rec.Body.Bytes(), &resp)
		s.NoError(err)
		s.NotNil(resp.Message)
	})

	s.Run("Assign ticket with invalid ID returns 400", func() {
		assigneeID := uuid.New()

		assignReq := openapi.AssignTicketRequest{
			AssigneeId: &assigneeID,
		}

		assignBody, _ := json.Marshal(assignReq)
		req := httptest.NewRequest(http.MethodPatch, "/tickets/invalid-uuid/assign", bytes.NewBuffer(assignBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		s.HTTPServer.ServeHTTP(rec, req)
		s.Equal(http.StatusBadRequest, rec.Code)
	})

	s.Run("Assign ticket with invalid JSON returns 400", func() {
		// Create a test ticket first
		orgID := uuid.New()
		authorID := uuid.New()

		ticketReq := openapi.CreateTicketRequest{
			Title:          "Test Ticket",
			Description:    "Test ticket",
			Priority:       openapi.TicketPriority("normal"),
			OrganizationId: orgID,
			AuthorId:       authorID,
		}

		createBody, _ := json.Marshal(ticketReq)
		createReq := httptest.NewRequest(http.MethodPost, "/tickets", bytes.NewBuffer(createBody))
		createReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		createRec := httptest.NewRecorder()

		s.HTTPServer.ServeHTTP(createRec, createReq)
		s.Equal(http.StatusCreated, createRec.Code)

		var createResp openapi.GetTicketResponse
		err := json.Unmarshal(createRec.Body.Bytes(), &createResp)
		s.NoError(err)
		ticketID := *createResp.Id

		// Try to assign with invalid JSON
		req := httptest.NewRequest(
			http.MethodPatch,
			fmt.Sprintf("/tickets/%s/assign", ticketID.String()),
			bytes.NewBufferString(`{"invalid": json}`),
		)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		s.HTTPServer.ServeHTTP(rec, req)
		s.Equal(http.StatusBadRequest, rec.Code)
	})

	s.Run("Reassign already assigned ticket", func() {
		// Create a test ticket
		orgID := uuid.New()
		authorID := uuid.New()
		firstAssigneeID := uuid.New()
		secondAssigneeID := uuid.New()

		ticketReq := openapi.CreateTicketRequest{
			Title:          "Test Ticket for Reassignment",
			Description:    "This ticket will be reassigned",
			Priority:       openapi.TicketPriority("normal"),
			OrganizationId: orgID,
			AuthorId:       authorID,
		}

		// Create the ticket
		createBody, _ := json.Marshal(ticketReq)
		createReq := httptest.NewRequest(http.MethodPost, "/tickets", bytes.NewBuffer(createBody))
		createReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		createRec := httptest.NewRecorder()

		s.HTTPServer.ServeHTTP(createRec, createReq)
		s.Equal(http.StatusCreated, createRec.Code)

		var createResp openapi.GetTicketResponse
		err := json.Unmarshal(createRec.Body.Bytes(), &createResp)
		s.NoError(err)
		ticketID := *createResp.Id

		// First assignment
		assignReq := openapi.AssignTicketRequest{
			AssigneeId: &firstAssigneeID,
		}

		assignBody, _ := json.Marshal(assignReq)
		firstReq := httptest.NewRequest(
			http.MethodPatch,
			fmt.Sprintf("/tickets/%s/assign", ticketID.String()),
			bytes.NewBuffer(assignBody),
		)
		firstReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		firstRec := httptest.NewRecorder()

		s.HTTPServer.ServeHTTP(firstRec, firstReq)
		s.Equal(http.StatusOK, firstRec.Code)

		// Second assignment (reassignment)
		reassignReq := openapi.AssignTicketRequest{
			AssigneeId: &secondAssigneeID,
		}

		reassignBody, _ := json.Marshal(reassignReq)
		secondReq := httptest.NewRequest(
			http.MethodPatch,
			fmt.Sprintf("/tickets/%s/assign", ticketID.String()),
			bytes.NewBuffer(reassignBody),
		)
		secondReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		secondRec := httptest.NewRecorder()

		s.HTTPServer.ServeHTTP(secondRec, secondReq)
		s.Equal(http.StatusOK, secondRec.Code)

		var resp openapi.GetTicketResponse
		err = json.Unmarshal(secondRec.Body.Bytes(), &resp)
		s.NoError(err)
		s.NotNil(resp.AssigneeId)
		s.Equal(secondAssigneeID, *resp.AssigneeId)
	})
}
