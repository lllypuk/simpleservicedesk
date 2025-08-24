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

func (s *TicketsSuite) TestUpdateTicket() {
	s.Run("Update ticket title successfully", func() {
		// Create a test ticket first
		orgID := uuid.New()
		authorID := uuid.New()

		ticketReq := openapi.CreateTicketRequest{
			Title:          "Original Title",
			Description:    "Original description",
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

		// Update the ticket title
		newTitle := "Updated Title"
		updateReq := openapi.UpdateTicketRequest{
			Title: &newTitle,
		}

		updateBody, _ := json.Marshal(updateReq)
		req := httptest.NewRequest(
			http.MethodPut,
			fmt.Sprintf("/tickets/%s", ticketID.String()),
			bytes.NewBuffer(updateBody),
		)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		s.HTTPServer.ServeHTTP(rec, req)
		s.Equal(http.StatusOK, rec.Code)

		var resp openapi.GetTicketResponse
		err = json.Unmarshal(rec.Body.Bytes(), &resp)
		s.NoError(err)
		s.NotNil(resp.Title)
		s.Equal("Updated Title", *resp.Title)
		// Verify other properties unchanged
		s.Equal("Original description", *resp.Description)
		s.Equal(openapi.TicketPriority("normal"), *resp.Priority)
	})

	s.Run("Update ticket description successfully", func() {
		// Create a test ticket
		orgID := uuid.New()
		authorID := uuid.New()

		ticketReq := openapi.CreateTicketRequest{
			Title:          "Test Ticket",
			Description:    "Original description",
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

		// Update the ticket description
		newDescription := "Updated description with more details"
		updateReq := openapi.UpdateTicketRequest{
			Description: &newDescription,
		}

		updateBody, _ := json.Marshal(updateReq)
		req := httptest.NewRequest(
			http.MethodPut,
			fmt.Sprintf("/tickets/%s", ticketID.String()),
			bytes.NewBuffer(updateBody),
		)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		s.HTTPServer.ServeHTTP(rec, req)
		s.Equal(http.StatusOK, rec.Code)

		var resp openapi.GetTicketResponse
		err = json.Unmarshal(rec.Body.Bytes(), &resp)
		s.NoError(err)
		s.NotNil(resp.Description)
		s.Equal("Updated description with more details", *resp.Description)
		// Verify other properties unchanged
		s.Equal("Test Ticket", *resp.Title)
	})

	s.Run("Update ticket priority successfully", func() {
		// Create a test ticket
		orgID := uuid.New()
		authorID := uuid.New()

		ticketReq := openapi.CreateTicketRequest{
			Title:          "Test Ticket",
			Description:    "Test description",
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

		// Update the ticket priority
		newPriority := openapi.TicketPriority("high")
		updateReq := openapi.UpdateTicketRequest{
			Priority: &newPriority,
		}

		updateBody, _ := json.Marshal(updateReq)
		req := httptest.NewRequest(
			http.MethodPut,
			fmt.Sprintf("/tickets/%s", ticketID.String()),
			bytes.NewBuffer(updateBody),
		)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		s.HTTPServer.ServeHTTP(rec, req)
		s.Equal(http.StatusOK, rec.Code)

		var resp openapi.GetTicketResponse
		err = json.Unmarshal(rec.Body.Bytes(), &resp)
		s.NoError(err)
		s.NotNil(resp.Priority)
		s.Equal(openapi.TicketPriority("high"), *resp.Priority)
	})

	s.Run("Update ticket category successfully", func() {
		// Create a test ticket
		orgID := uuid.New()
		authorID := uuid.New()

		ticketReq := openapi.CreateTicketRequest{
			Title:          "Test Ticket",
			Description:    "Test description",
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

		// Update the ticket category
		categoryID := uuid.New()
		updateReq := openapi.UpdateTicketRequest{
			CategoryId: &categoryID,
		}

		updateBody, _ := json.Marshal(updateReq)
		req := httptest.NewRequest(
			http.MethodPut,
			fmt.Sprintf("/tickets/%s", ticketID.String()),
			bytes.NewBuffer(updateBody),
		)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		s.HTTPServer.ServeHTTP(rec, req)
		s.Equal(http.StatusOK, rec.Code)

		var resp openapi.GetTicketResponse
		err = json.Unmarshal(rec.Body.Bytes(), &resp)
		s.NoError(err)
		s.NotNil(resp.CategoryId)
		s.Equal(categoryID, *resp.CategoryId)
	})

	s.Run("Update multiple ticket fields simultaneously", func() {
		// Create a test ticket
		orgID := uuid.New()
		authorID := uuid.New()

		ticketReq := openapi.CreateTicketRequest{
			Title:          "Original Title",
			Description:    "Original description",
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

		// Update multiple fields
		newTitle := "Updated Title"
		newDescription := "Updated description"
		newPriority := openapi.TicketPriority("critical")
		categoryID := uuid.New()

		updateReq := openapi.UpdateTicketRequest{
			Title:       &newTitle,
			Description: &newDescription,
			Priority:    &newPriority,
			CategoryId:  &categoryID,
		}

		updateBody, _ := json.Marshal(updateReq)
		req := httptest.NewRequest(
			http.MethodPut,
			fmt.Sprintf("/tickets/%s", ticketID.String()),
			bytes.NewBuffer(updateBody),
		)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		s.HTTPServer.ServeHTTP(rec, req)
		s.Equal(http.StatusOK, rec.Code)

		var resp openapi.GetTicketResponse
		err = json.Unmarshal(rec.Body.Bytes(), &resp)
		s.NoError(err)
		s.Equal("Updated Title", *resp.Title)
		s.Equal("Updated description", *resp.Description)
		s.Equal(openapi.TicketPriority("critical"), *resp.Priority)
		s.Equal(categoryID, *resp.CategoryId)
	})

	s.Run("Update with empty title returns 400", func() {
		// Create a test ticket
		orgID := uuid.New()
		authorID := uuid.New()

		ticketReq := openapi.CreateTicketRequest{
			Title:          "Valid Title",
			Description:    "Test description",
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

		// Try to update with empty title
		emptyTitle := ""
		updateReq := openapi.UpdateTicketRequest{
			Title: &emptyTitle,
		}

		updateBody, _ := json.Marshal(updateReq)
		req := httptest.NewRequest(
			http.MethodPut,
			fmt.Sprintf("/tickets/%s", ticketID.String()),
			bytes.NewBuffer(updateBody),
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

	s.Run("Update with invalid priority returns 400", func() {
		// Create a test ticket
		orgID := uuid.New()
		authorID := uuid.New()

		ticketReq := openapi.CreateTicketRequest{
			Title:          "Test Ticket",
			Description:    "Test description",
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

		// Try to update with invalid priority
		invalidPriority := openapi.TicketPriority("invalid_priority")
		updateReq := openapi.UpdateTicketRequest{
			Priority: &invalidPriority,
		}

		updateBody, _ := json.Marshal(updateReq)
		req := httptest.NewRequest(
			http.MethodPut,
			fmt.Sprintf("/tickets/%s", ticketID.String()),
			bytes.NewBuffer(updateBody),
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

	s.Run("Update non-existent ticket returns 404", func() {
		nonExistentID := uuid.New()

		newTitle := "Updated Title"
		updateReq := openapi.UpdateTicketRequest{
			Title: &newTitle,
		}

		updateBody, _ := json.Marshal(updateReq)
		req := httptest.NewRequest(
			http.MethodPut,
			fmt.Sprintf("/tickets/%s", nonExistentID.String()),
			bytes.NewBuffer(updateBody),
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

	s.Run("Update ticket with invalid ID returns 400", func() {
		newTitle := "Updated Title"
		updateReq := openapi.UpdateTicketRequest{
			Title: &newTitle,
		}

		updateBody, _ := json.Marshal(updateReq)
		req := httptest.NewRequest(http.MethodPut, "/tickets/invalid-uuid", bytes.NewBuffer(updateBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		s.HTTPServer.ServeHTTP(rec, req)
		s.Equal(http.StatusBadRequest, rec.Code)
	})

	s.Run("Update ticket with invalid JSON returns 400", func() {
		// Create a test ticket
		orgID := uuid.New()
		authorID := uuid.New()

		ticketReq := openapi.CreateTicketRequest{
			Title:          "Test Ticket",
			Description:    "Test description",
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
			http.MethodPut,
			fmt.Sprintf("/tickets/%s", ticketID.String()),
			bytes.NewBufferString(`{"invalid": json}`),
		)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		s.HTTPServer.ServeHTTP(rec, req)
		s.Equal(http.StatusBadRequest, rec.Code)
	})

	s.Run("Update ticket with empty request body returns no changes", func() {
		// Create a test ticket
		orgID := uuid.New()
		authorID := uuid.New()

		ticketReq := openapi.CreateTicketRequest{
			Title:          "Original Title",
			Description:    "Original description",
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

		// Update with empty request (no fields specified)
		updateReq := openapi.UpdateTicketRequest{}

		updateBody, _ := json.Marshal(updateReq)
		req := httptest.NewRequest(
			http.MethodPut,
			fmt.Sprintf("/tickets/%s", ticketID.String()),
			bytes.NewBuffer(updateBody),
		)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		s.HTTPServer.ServeHTTP(rec, req)
		s.Equal(http.StatusOK, rec.Code)

		var resp openapi.GetTicketResponse
		err = json.Unmarshal(rec.Body.Bytes(), &resp)
		s.NoError(err)
		// Verify no changes were made
		s.Equal("Original Title", *resp.Title)
		s.Equal("Original description", *resp.Description)
		s.Equal(openapi.TicketPriority("normal"), *resp.Priority)
	})

	s.Run("Update ticket with empty description is valid", func() {
		// Create a test ticket
		orgID := uuid.New()
		authorID := uuid.New()

		ticketReq := openapi.CreateTicketRequest{
			Title:          "Test Ticket",
			Description:    "Original description",
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

		// Update with empty description (should be valid)
		emptyDescription := ""
		updateReq := openapi.UpdateTicketRequest{
			Description: &emptyDescription,
		}

		updateBody, _ := json.Marshal(updateReq)
		req := httptest.NewRequest(
			http.MethodPut,
			fmt.Sprintf("/tickets/%s", ticketID.String()),
			bytes.NewBuffer(updateBody),
		)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		s.HTTPServer.ServeHTTP(rec, req)
		s.Equal(http.StatusOK, rec.Code)

		var resp openapi.GetTicketResponse
		err = json.Unmarshal(rec.Body.Bytes(), &resp)
		s.NoError(err)
		s.Equal("", *resp.Description)
	})

	s.Run("Update ticket priority to all valid values", func() {
		priorities := []openapi.TicketPriority{"low", "normal", "high", "critical"}

		for _, priority := range priorities {
			// Create a test ticket
			orgID := uuid.New()
			authorID := uuid.New()

			ticketReq := openapi.CreateTicketRequest{
				Title:          fmt.Sprintf("Test Ticket %s Priority", priority),
				Description:    "Test description",
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

			// Update priority
			updateReq := openapi.UpdateTicketRequest{
				Priority: &priority,
			}

			updateBody, _ := json.Marshal(updateReq)
			req := httptest.NewRequest(
				http.MethodPut,
				fmt.Sprintf("/tickets/%s", ticketID.String()),
				bytes.NewBuffer(updateBody),
			)
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()

			s.HTTPServer.ServeHTTP(rec, req)
			s.Equal(http.StatusOK, rec.Code, fmt.Sprintf("Failed to update priority to %s", priority))

			var resp openapi.GetTicketResponse
			err = json.Unmarshal(rec.Body.Bytes(), &resp)
			s.NoError(err)
			s.Equal(priority, *resp.Priority, fmt.Sprintf("Priority not updated to %s", priority))
		}
	})
}
