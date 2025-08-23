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

func (s *TicketsSuite) TestCreateComment() {
	s.Run("Create comment successfully", func() {
		// Create a test ticket first
		orgID := uuid.New()
		authorID := uuid.New()
		commentAuthorID := uuid.New()

		ticketReq := openapi.CreateTicketRequest{
			Title:          "Test Ticket for Comments",
			Description:    "This ticket will have comments",
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

		// Create a comment
		commentReq := openapi.CreateCommentRequest{
			AuthorId: commentAuthorID,
			Content:  "This is a test comment",
		}

		commentBody, _ := json.Marshal(commentReq)
		req := httptest.NewRequest(
			http.MethodPost,
			fmt.Sprintf("/tickets/%s/comments", ticketID.String()),
			bytes.NewBuffer(commentBody),
		)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		s.HTTPServer.ServeHTTP(rec, req)
		s.Equal(http.StatusCreated, rec.Code)

		var resp openapi.TicketComment
		err = json.Unmarshal(rec.Body.Bytes(), &resp)
		s.NoError(err)
		s.NotNil(resp.Id)
		s.NotNil(resp.TicketId)
		s.NotNil(resp.AuthorId)
		s.NotNil(resp.Content)
		s.Equal(ticketID, *resp.TicketId)
		s.Equal(commentAuthorID, *resp.AuthorId)
		s.Equal("This is a test comment", *resp.Content)
		s.NotNil(resp.IsInternal)
		s.False(*resp.IsInternal) // Default should be false
	})

	s.Run("Create internal comment successfully", func() {
		// Create a test ticket first
		orgID := uuid.New()
		authorID := uuid.New()
		commentAuthorID := uuid.New()

		ticketReq := openapi.CreateTicketRequest{
			Title:          "Test Ticket for Internal Comments",
			Description:    "This ticket will have internal comments",
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

		// Create an internal comment
		isInternal := true
		commentReq := openapi.CreateCommentRequest{
			AuthorId:   commentAuthorID,
			Content:    "This is an internal comment",
			IsInternal: &isInternal,
		}

		commentBody, _ := json.Marshal(commentReq)
		req := httptest.NewRequest(
			http.MethodPost,
			fmt.Sprintf("/tickets/%s/comments", ticketID.String()),
			bytes.NewBuffer(commentBody),
		)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		s.HTTPServer.ServeHTTP(rec, req)
		s.Equal(http.StatusCreated, rec.Code)

		var resp openapi.TicketComment
		err = json.Unmarshal(rec.Body.Bytes(), &resp)
		s.NoError(err)
		s.NotNil(resp.IsInternal)
		s.True(*resp.IsInternal)
		s.Equal("This is an internal comment", *resp.Content)
	})

	s.Run("Create comment on non-existent ticket returns 404", func() {
		nonExistentID := uuid.New()
		commentAuthorID := uuid.New()

		commentReq := openapi.CreateCommentRequest{
			AuthorId: commentAuthorID,
			Content:  "This comment won't be created",
		}

		commentBody, _ := json.Marshal(commentReq)
		req := httptest.NewRequest(
			http.MethodPost,
			fmt.Sprintf("/tickets/%s/comments", nonExistentID.String()),
			bytes.NewBuffer(commentBody),
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

	s.Run("Create comment with invalid ticket ID returns 400", func() {
		commentAuthorID := uuid.New()

		commentReq := openapi.CreateCommentRequest{
			AuthorId: commentAuthorID,
			Content:  "This comment won't be created",
		}

		commentBody, _ := json.Marshal(commentReq)
		req := httptest.NewRequest(http.MethodPost, "/tickets/invalid-uuid/comments", bytes.NewBuffer(commentBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		s.HTTPServer.ServeHTTP(rec, req)
		s.Equal(http.StatusBadRequest, rec.Code)
	})

	s.Run("Create comment with invalid JSON returns 400", func() {
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

		// Try to create comment with invalid JSON
		req := httptest.NewRequest(
			http.MethodPost,
			fmt.Sprintf("/tickets/%s/comments", ticketID.String()),
			bytes.NewBufferString(`{"invalid": json}`),
		)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		s.HTTPServer.ServeHTTP(rec, req)
		s.Equal(http.StatusBadRequest, rec.Code)
	})

	s.Run("Create comment with empty content should fail", func() {
		// Create a test ticket first
		orgID := uuid.New()
		authorID := uuid.New()
		commentAuthorID := uuid.New()

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

		// Try to create comment with empty content
		commentReq := openapi.CreateCommentRequest{
			AuthorId: commentAuthorID,
			Content:  "",
		}

		commentBody, _ := json.Marshal(commentReq)
		req := httptest.NewRequest(
			http.MethodPost,
			fmt.Sprintf("/tickets/%s/comments", ticketID.String()),
			bytes.NewBuffer(commentBody),
		)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		s.HTTPServer.ServeHTTP(rec, req)
		s.Equal(http.StatusBadRequest, rec.Code)
	})
}

func (s *TicketsSuite) TestGetComments() {
	s.Run("Get all comments from ticket", func() {
		// Create a test ticket first
		orgID := uuid.New()
		authorID := uuid.New()
		commentAuthorID := uuid.New()

		ticketReq := openapi.CreateTicketRequest{
			Title:          "Test Ticket for Getting Comments",
			Description:    "This ticket will have comments to retrieve",
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

		// Add a few comments
		comments := []string{"First comment", "Second comment", "Third comment"}
		for _, content := range comments {
			commentReq := openapi.CreateCommentRequest{
				AuthorId: commentAuthorID,
				Content:  content,
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
		}

		// Get all comments
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/tickets/%s/comments", ticketID.String()), nil)
		rec := httptest.NewRecorder()

		s.HTTPServer.ServeHTTP(rec, req)
		s.Equal(http.StatusOK, rec.Code)

		var resp []openapi.TicketComment
		err = json.Unmarshal(rec.Body.Bytes(), &resp)
		s.NoError(err)
		s.Len(resp, 3)

		// Verify comment contents
		for i, comment := range resp {
			s.Equal(comments[i], *comment.Content)
			s.Equal(commentAuthorID, *comment.AuthorId)
			s.Equal(ticketID, *comment.TicketId)
		}
	})

	s.Run("Get comments including internal ones", func() {
		// Create a test ticket first
		orgID := uuid.New()
		authorID := uuid.New()
		commentAuthorID := uuid.New()

		ticketReq := openapi.CreateTicketRequest{
			Title:          "Test Ticket for Internal Comments",
			Description:    "This ticket will have internal and public comments",
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

		// Add public comment
		publicCommentReq := openapi.CreateCommentRequest{
			AuthorId: commentAuthorID,
			Content:  "Public comment",
		}
		publicBody, _ := json.Marshal(publicCommentReq)
		publicReqHTTP := httptest.NewRequest(
			http.MethodPost,
			fmt.Sprintf("/tickets/%s/comments", ticketID.String()),
			bytes.NewBuffer(publicBody),
		)
		publicReqHTTP.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		publicRec := httptest.NewRecorder()

		s.HTTPServer.ServeHTTP(publicRec, publicReqHTTP)
		s.Equal(http.StatusCreated, publicRec.Code)

		// Add internal comment
		isInternal := true
		internalCommentReq := openapi.CreateCommentRequest{
			AuthorId:   commentAuthorID,
			Content:    "Internal comment",
			IsInternal: &isInternal,
		}
		internalBody, _ := json.Marshal(internalCommentReq)
		internalReqHTTP := httptest.NewRequest(
			http.MethodPost,
			fmt.Sprintf("/tickets/%s/comments", ticketID.String()),
			bytes.NewBuffer(internalBody),
		)
		internalReqHTTP.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		internalRec := httptest.NewRecorder()

		s.HTTPServer.ServeHTTP(internalRec, internalReqHTTP)
		s.Equal(http.StatusCreated, internalRec.Code)

		// Get comments without including internal ones (default)
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/tickets/%s/comments", ticketID.String()), nil)
		rec := httptest.NewRecorder()

		s.HTTPServer.ServeHTTP(rec, req)
		s.Equal(http.StatusOK, rec.Code)

		var resp []openapi.TicketComment
		err = json.Unmarshal(rec.Body.Bytes(), &resp)
		s.NoError(err)
		s.Len(resp, 1) // Only public comment
		s.Equal("Public comment", *resp[0].Content)

		// Get comments including internal ones
		reqWithInternal := httptest.NewRequest(
			http.MethodGet,
			fmt.Sprintf("/tickets/%s/comments?include_internal=true", ticketID.String()),
			nil,
		)
		recWithInternal := httptest.NewRecorder()

		s.HTTPServer.ServeHTTP(recWithInternal, reqWithInternal)
		s.Equal(http.StatusOK, recWithInternal.Code)

		var respWithInternal []openapi.TicketComment
		err = json.Unmarshal(recWithInternal.Body.Bytes(), &respWithInternal)
		s.NoError(err)
		s.Len(respWithInternal, 2) // Both public and internal comments
	})

	s.Run("Get comments from non-existent ticket returns 404", func() {
		nonExistentID := uuid.New()

		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/tickets/%s/comments", nonExistentID.String()), nil)
		rec := httptest.NewRecorder()

		s.HTTPServer.ServeHTTP(rec, req)
		s.Equal(http.StatusNotFound, rec.Code)

		var resp openapi.ErrorResponse
		err := json.Unmarshal(rec.Body.Bytes(), &resp)
		s.NoError(err)
		s.NotNil(resp.Message)
	})

	s.Run("Get comments with invalid ticket ID returns 400", func() {
		req := httptest.NewRequest(http.MethodGet, "/tickets/invalid-uuid/comments", nil)
		rec := httptest.NewRecorder()

		s.HTTPServer.ServeHTTP(rec, req)
		s.Equal(http.StatusBadRequest, rec.Code)
	})

	s.Run("Get comments from ticket with no comments returns empty array", func() {
		// Create a test ticket without comments
		orgID := uuid.New()
		authorID := uuid.New()

		ticketReq := openapi.CreateTicketRequest{
			Title:          "Empty Ticket",
			Description:    "This ticket has no comments",
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

		// Get comments
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/tickets/%s/comments", ticketID.String()), nil)
		rec := httptest.NewRecorder()

		s.HTTPServer.ServeHTTP(rec, req)
		s.Equal(http.StatusOK, rec.Code)

		var resp []openapi.TicketComment
		err = json.Unmarshal(rec.Body.Bytes(), &resp)
		s.NoError(err)
		s.Len(resp, 0)
	})
}
