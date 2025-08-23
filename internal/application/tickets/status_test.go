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

func (s *TicketsSuite) TestUpdateTicketStatus() {
	s.Run("Update status from open to in_progress", func() {
		// Create a test ticket first
		orgID := uuid.New()
		authorID := uuid.New()

		ticketReq := openapi.CreateTicketRequest{
			Title:          "Test Ticket for Status Update",
			Description:    "This ticket's status will be updated",
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
		s.NotNil(createResp.Status)
		s.Equal(openapi.TicketStatus("new"), *createResp.Status)

		ticketID := *createResp.Id

		// Update status to in_progress
		statusReq := openapi.UpdateTicketStatusRequest{
			Status: openapi.TicketStatus("in_progress"),
		}

		statusBody, _ := json.Marshal(statusReq)
		req := httptest.NewRequest(
			http.MethodPatch,
			fmt.Sprintf("/tickets/%s/status", ticketID.String()),
			bytes.NewBuffer(statusBody),
		)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		s.HTTPServer.ServeHTTP(rec, req)
		s.Equal(http.StatusOK, rec.Code)

		var resp openapi.GetTicketResponse
		err = json.Unmarshal(rec.Body.Bytes(), &resp)
		s.NoError(err)
		s.NotNil(resp.Status)
		s.Equal(openapi.TicketStatus("in_progress"), *resp.Status)
	})

	s.Run("Update status from in_progress to resolved", func() {
		// Create a test ticket
		orgID := uuid.New()
		authorID := uuid.New()

		ticketReq := openapi.CreateTicketRequest{
			Title:          "Test Ticket for Resolution",
			Description:    "This ticket will be resolved",
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

		// First update to in_progress
		statusReq1 := openapi.UpdateTicketStatusRequest{
			Status: openapi.TicketStatus("in_progress"),
		}

		statusBody1, _ := json.Marshal(statusReq1)
		req1 := httptest.NewRequest(
			http.MethodPatch,
			fmt.Sprintf("/tickets/%s/status", ticketID.String()),
			bytes.NewBuffer(statusBody1),
		)
		req1.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec1 := httptest.NewRecorder()

		s.HTTPServer.ServeHTTP(rec1, req1)
		s.Equal(http.StatusOK, rec1.Code)

		// Then update to resolved
		statusReq2 := openapi.UpdateTicketStatusRequest{
			Status: openapi.TicketStatus("resolved"),
		}

		statusBody2, _ := json.Marshal(statusReq2)
		req2 := httptest.NewRequest(
			http.MethodPatch,
			fmt.Sprintf("/tickets/%s/status", ticketID.String()),
			bytes.NewBuffer(statusBody2),
		)
		req2.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec2 := httptest.NewRecorder()

		s.HTTPServer.ServeHTTP(rec2, req2)
		s.Equal(http.StatusOK, rec2.Code)

		var resp openapi.GetTicketResponse
		err = json.Unmarshal(rec2.Body.Bytes(), &resp)
		s.NoError(err)
		s.NotNil(resp.Status)
		s.Equal(openapi.TicketStatus("resolved"), *resp.Status)
	})

	s.Run("Update status from resolved to closed", func() {
		// Create and progress ticket to resolved
		orgID := uuid.New()
		authorID := uuid.New()

		ticketReq := openapi.CreateTicketRequest{
			Title:          "Test Ticket for Closure",
			Description:    "This ticket will be closed",
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

		// Update to in_progress then resolved
		statuses := []openapi.TicketStatus{"in_progress", "resolved", "closed"}
		for _, status := range statuses {
			statusReq := openapi.UpdateTicketStatusRequest{
				Status: status,
			}

			statusBody, _ := json.Marshal(statusReq)
			req := httptest.NewRequest(
				http.MethodPatch,
				fmt.Sprintf("/tickets/%s/status", ticketID.String()),
				bytes.NewBuffer(statusBody),
			)
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()

			s.HTTPServer.ServeHTTP(rec, req)
			s.Equal(http.StatusOK, rec.Code)

			var resp openapi.GetTicketResponse
			err = json.Unmarshal(rec.Body.Bytes(), &resp)
			s.NoError(err)
			s.NotNil(resp.Status)
			s.Equal(status, *resp.Status)
		}
	})

	s.Run("Invalid status transition returns 400", func() {
		// Create a test ticket
		orgID := uuid.New()
		authorID := uuid.New()

		ticketReq := openapi.CreateTicketRequest{
			Title:          "Test Ticket for Invalid Transition",
			Description:    "This ticket will test invalid transitions",
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

		// Try to jump directly from open to closed (should be invalid)
		statusReq := openapi.UpdateTicketStatusRequest{
			Status: openapi.TicketStatus("resolved"),
		}

		statusBody, _ := json.Marshal(statusReq)
		req := httptest.NewRequest(
			http.MethodPatch,
			fmt.Sprintf("/tickets/%s/status", ticketID.String()),
			bytes.NewBuffer(statusBody),
		)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		s.HTTPServer.ServeHTTP(rec, req)
		s.Equal(http.StatusBadRequest, rec.Code)

		var resp openapi.ErrorResponse
		err = json.Unmarshal(rec.Body.Bytes(), &resp)
		s.NoError(err)
		s.NotNil(resp.Message)
	})

	s.Run("Invalid status value returns 400", func() {
		// Create a test ticket
		orgID := uuid.New()
		authorID := uuid.New()

		ticketReq := openapi.CreateTicketRequest{
			Title:          "Test Ticket for Invalid Status",
			Description:    "This ticket will test invalid status",
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

		// Try to set an invalid status
		statusReq := openapi.UpdateTicketStatusRequest{
			Status: openapi.TicketStatus("invalid_status"),
		}

		statusBody, _ := json.Marshal(statusReq)
		req := httptest.NewRequest(
			http.MethodPatch,
			fmt.Sprintf("/tickets/%s/status", ticketID.String()),
			bytes.NewBuffer(statusBody),
		)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		s.HTTPServer.ServeHTTP(rec, req)
		s.Equal(http.StatusBadRequest, rec.Code)

		var resp openapi.ErrorResponse
		err = json.Unmarshal(rec.Body.Bytes(), &resp)
		s.NoError(err)
		s.NotNil(resp.Message)
		s.Contains(*resp.Message, "invalid status")
	})

	s.Run("Update status of non-existent ticket returns 404", func() {
		nonExistentID := uuid.New()

		statusReq := openapi.UpdateTicketStatusRequest{
			Status: openapi.TicketStatus("in_progress"),
		}

		statusBody, _ := json.Marshal(statusReq)
		req := httptest.NewRequest(
			http.MethodPatch,
			fmt.Sprintf("/tickets/%s/status", nonExistentID.String()),
			bytes.NewBuffer(statusBody),
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

	s.Run("Update status with invalid ticket ID returns 400", func() {
		statusReq := openapi.UpdateTicketStatusRequest{
			Status: openapi.TicketStatus("in_progress"),
		}

		statusBody, _ := json.Marshal(statusReq)
		req := httptest.NewRequest(http.MethodPatch, "/tickets/invalid-uuid/status", bytes.NewBuffer(statusBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		s.HTTPServer.ServeHTTP(rec, req)
		s.Equal(http.StatusBadRequest, rec.Code)
	})

	s.Run("Update status with invalid JSON returns 400", func() {
		// Create a test ticket
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

		// Try to update with invalid JSON
		req := httptest.NewRequest(
			http.MethodPatch,
			fmt.Sprintf("/tickets/%s/status", ticketID.String()),
			bytes.NewBufferString(`{"invalid": json}`),
		)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		s.HTTPServer.ServeHTTP(rec, req)
		s.Equal(http.StatusBadRequest, rec.Code)
	})

	s.Run("Update status maintains other ticket properties", func() {
		// Create a test ticket
		orgID := uuid.New()
		authorID := uuid.New()

		ticketReq := openapi.CreateTicketRequest{
			Title:          "Test Ticket Properties",
			Description:    "This ticket tests property preservation",
			Priority:       openapi.TicketPriority("high"),
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

		// Update status
		statusReq := openapi.UpdateTicketStatusRequest{
			Status: openapi.TicketStatus("in_progress"),
		}

		statusBody, _ := json.Marshal(statusReq)
		req := httptest.NewRequest(
			http.MethodPatch,
			fmt.Sprintf("/tickets/%s/status", ticketID.String()),
			bytes.NewBuffer(statusBody),
		)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		s.HTTPServer.ServeHTTP(rec, req)
		s.Equal(http.StatusOK, rec.Code)

		var resp openapi.GetTicketResponse
		err = json.Unmarshal(rec.Body.Bytes(), &resp)
		s.NoError(err)

		// Verify status changed but other properties preserved
		s.Equal(openapi.TicketStatus("in_progress"), *resp.Status)
		s.Equal("Test Ticket Properties", *resp.Title)
		s.Equal("This ticket tests property preservation", *resp.Description)
		s.Equal(openapi.TicketPriority("high"), *resp.Priority)
		s.Equal(orgID, *resp.OrganizationId)
		s.Equal(authorID, *resp.AuthorId)
	})

	s.Run("Multiple status updates in sequence", func() {
		// Create a test ticket
		orgID := uuid.New()
		authorID := uuid.New()

		ticketReq := openapi.CreateTicketRequest{
			Title:          "Test Multiple Status Updates",
			Description:    "This ticket will go through multiple status changes",
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

		// Test full lifecycle: open -> in_progress -> resolved -> closed
		statusSequence := []openapi.TicketStatus{"in_progress", "resolved", "closed"}

		for _, status := range statusSequence {
			statusReq := openapi.UpdateTicketStatusRequest{
				Status: status,
			}

			statusBody, _ := json.Marshal(statusReq)
			req := httptest.NewRequest(
				http.MethodPatch,
				fmt.Sprintf("/tickets/%s/status", ticketID.String()),
				bytes.NewBuffer(statusBody),
			)
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()

			s.HTTPServer.ServeHTTP(rec, req)
			s.Equal(http.StatusOK, rec.Code, fmt.Sprintf("Failed to update to status: %s", status))

			var resp openapi.GetTicketResponse
			err = json.Unmarshal(rec.Body.Bytes(), &resp)
			s.NoError(err)
			s.Equal(status, *resp.Status, fmt.Sprintf("Status not updated to: %s", status))
		}
	})
}
