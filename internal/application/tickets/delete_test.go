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

func (s *TicketsSuite) TestDeleteTicket() {
	s.Run("Delete existing ticket successfully", func() {
		// Create a test ticket first
		orgID := uuid.New()
		authorID := uuid.New()

		ticketReq := openapi.CreateTicketRequest{
			Title:          "Test Ticket for Deletion",
			Description:    "This ticket will be deleted",
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

		// Delete the ticket
		req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/tickets/%s", ticketID.String()), nil)
		rec := httptest.NewRecorder()

		s.HTTPServer.ServeHTTP(rec, req)
		s.Equal(http.StatusNoContent, rec.Code)
		s.Empty(rec.Body.String())

		// Verify ticket is deleted by trying to get it
		getReq := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/tickets/%s", ticketID.String()), nil)
		getRec := httptest.NewRecorder()

		s.HTTPServer.ServeHTTP(getRec, getReq)
		s.Equal(http.StatusNotFound, getRec.Code)
	})

	s.Run("Delete non-existent ticket returns 404", func() {
		nonExistentID := uuid.New()

		req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/tickets/%s", nonExistentID.String()), nil)
		rec := httptest.NewRecorder()

		s.HTTPServer.ServeHTTP(rec, req)
		s.Equal(http.StatusNotFound, rec.Code)

		var resp openapi.ErrorResponse
		err := json.Unmarshal(rec.Body.Bytes(), &resp)
		s.NoError(err)
		s.NotNil(resp.Message)
	})

	s.Run("Delete ticket with invalid ID returns 400", func() {
		req := httptest.NewRequest(http.MethodDelete, "/tickets/invalid-uuid", nil)
		rec := httptest.NewRecorder()

		s.HTTPServer.ServeHTTP(rec, req)
		s.Equal(http.StatusBadRequest, rec.Code)
	})

	s.Run("Delete ticket with comments", func() {
		// Create a test ticket first
		orgID := uuid.New()
		authorID := uuid.New()
		commentAuthorID := uuid.New()

		ticketReq := openapi.CreateTicketRequest{
			Title:          "Test Ticket with Comments for Deletion",
			Description:    "This ticket has comments and will be deleted",
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

		// Add a comment to the ticket
		commentReq := openapi.CreateCommentRequest{
			AuthorId: commentAuthorID,
			Content:  "This is a comment before deletion",
		}

		commentBody, _ := json.Marshal(commentReq)
		commentReqHTTP := httptest.NewRequest(
			http.MethodPost,
			fmt.Sprintf("/tickets/%s/comments", ticketID.String()),
			bytes.NewBuffer(commentBody),
		)
		commentReqHTTP.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		commentRec := httptest.NewRecorder()

		s.HTTPServer.ServeHTTP(commentRec, commentReqHTTP)
		s.Equal(http.StatusCreated, commentRec.Code)

		// Delete the ticket
		req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/tickets/%s", ticketID.String()), nil)
		rec := httptest.NewRecorder()

		s.HTTPServer.ServeHTTP(rec, req)
		s.Equal(http.StatusNoContent, rec.Code)

		// Verify ticket and its comments are deleted
		getReq := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/tickets/%s", ticketID.String()), nil)
		getRec := httptest.NewRecorder()

		s.HTTPServer.ServeHTTP(getRec, getReq)
		s.Equal(http.StatusNotFound, getRec.Code)

		// Verify comments are also inaccessible
		getCommentsReq := httptest.NewRequest(
			http.MethodGet,
			fmt.Sprintf("/tickets/%s/comments", ticketID.String()),
			nil,
		)
		getCommentsRec := httptest.NewRecorder()

		s.HTTPServer.ServeHTTP(getCommentsRec, getCommentsReq)
		s.Equal(http.StatusNotFound, getCommentsRec.Code)
	})

	s.Run("Delete assigned ticket", func() {
		// Create a test ticket first
		orgID := uuid.New()
		authorID := uuid.New()
		assigneeID := uuid.New()

		ticketReq := openapi.CreateTicketRequest{
			Title:          "Assigned Ticket for Deletion",
			Description:    "This assigned ticket will be deleted",
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

		// Assign the ticket
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

		// Delete the assigned ticket
		req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/tickets/%s", ticketID.String()), nil)
		rec := httptest.NewRecorder()

		s.HTTPServer.ServeHTTP(rec, req)
		s.Equal(http.StatusNoContent, rec.Code)

		// Verify ticket is deleted
		getReq := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/tickets/%s", ticketID.String()), nil)
		getRec := httptest.NewRecorder()

		s.HTTPServer.ServeHTTP(getRec, getReq)
		s.Equal(http.StatusNotFound, getRec.Code)
	})

	s.Run("Delete ticket with different statuses", func() {
		statuses := []string{"open", "in_progress", "resolved", "closed"}

		for _, status := range statuses {
			// Create a test ticket
			orgID := uuid.New()
			authorID := uuid.New()

			ticketReq := openapi.CreateTicketRequest{
				Title:          fmt.Sprintf("Test Ticket with %s Status", status),
				Description:    fmt.Sprintf("This ticket has %s status and will be deleted", status),
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

			// Change status if not "open" (default status)
			if status != "open" {
				statusReq := openapi.UpdateTicketStatusRequest{
					Status: openapi.TicketStatus(status),
				}

				statusBody, _ := json.Marshal(statusReq)
				statusReqHTTP := httptest.NewRequest(
					http.MethodPatch,
					fmt.Sprintf("/tickets/%s/status", ticketID.String()),
					bytes.NewBuffer(statusBody),
				)
				statusReqHTTP.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
				statusRec := httptest.NewRecorder()

				s.HTTPServer.ServeHTTP(statusRec, statusReqHTTP)
				// Some status transitions might not be valid, so we don't assert success here
			}

			// Delete the ticket regardless of status
			req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/tickets/%s", ticketID.String()), nil)
			rec := httptest.NewRecorder()

			s.HTTPServer.ServeHTTP(rec, req)
			s.Equal(http.StatusNoContent, rec.Code, fmt.Sprintf("Failed to delete ticket with status %s", status))

			// Verify ticket is deleted
			getReq := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/tickets/%s", ticketID.String()), nil)
			getRec := httptest.NewRecorder()

			s.HTTPServer.ServeHTTP(getRec, getReq)
			s.Equal(
				http.StatusNotFound,
				getRec.Code,
				fmt.Sprintf("Ticket with status %s was not properly deleted", status),
			)
		}
	})

	s.Run("Multiple delete operations on same ticket", func() {
		// Create a test ticket first
		orgID := uuid.New()
		authorID := uuid.New()

		ticketReq := openapi.CreateTicketRequest{
			Title:          "Test Ticket for Multiple Deletions",
			Description:    "This ticket will be deleted multiple times",
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

		// First deletion should succeed
		req1 := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/tickets/%s", ticketID.String()), nil)
		rec1 := httptest.NewRecorder()

		s.HTTPServer.ServeHTTP(rec1, req1)
		s.Equal(http.StatusNoContent, rec1.Code)

		// Second deletion should return 404
		req2 := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/tickets/%s", ticketID.String()), nil)
		rec2 := httptest.NewRecorder()

		s.HTTPServer.ServeHTTP(rec2, req2)
		s.Equal(http.StatusNotFound, rec2.Code)

		var resp openapi.ErrorResponse
		err = json.Unmarshal(rec2.Body.Bytes(), &resp)
		s.NoError(err)
		s.NotNil(resp.Message)
	})
}
