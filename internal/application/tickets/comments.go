package tickets

import (
	"errors"
	"net/http"

	"simpleservicedesk/generated/openapi"
	"simpleservicedesk/internal/domain/tickets"

	"github.com/labstack/echo/v4"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

func (h TicketHandlers) PostTicketsIDComments(c echo.Context, id openapi_types.UUID) error {
	ctx := c.Request().Context()
	var req openapi.CreateCommentRequest
	if err := c.Bind(&req); err != nil {
		return err
	}

	// Convert author ID
	authorID := req.AuthorId

	// Determine if comment is internal
	isInternal := false
	if req.IsInternal != nil {
		isInternal = *req.IsInternal
	}

	ticket, err := h.repo.UpdateTicket(ctx, id, func(ticket *tickets.Ticket) (bool, error) {
		if err := ticket.AddComment(authorID, req.Content, isInternal); err != nil {
			return false, err
		}
		return true, nil
	})

	if err != nil {
		msg := err.Error()
		if errors.Is(err, tickets.ErrTicketNotFound) {
			return c.JSON(http.StatusNotFound, openapi.ErrorResponse{Message: &msg})
		}
		if errors.Is(err, tickets.ErrTicketValidation) {
			return c.JSON(http.StatusBadRequest, openapi.ErrorResponse{Message: &msg})
		}
		return c.JSON(http.StatusInternalServerError, openapi.ErrorResponse{Message: &msg})
	}

	// Get the last added comment
	comments := ticket.Comments()
	if len(comments) == 0 {
		msg := "failed to add comment"
		return c.JSON(http.StatusInternalServerError, openapi.ErrorResponse{Message: &msg})
	}

	lastComment := comments[len(comments)-1]
	response := convertCommentToResponse(lastComment)

	return c.JSON(http.StatusCreated, response)
}

func (h TicketHandlers) GetTicketsIDComments(
	c echo.Context, id openapi_types.UUID, params openapi.GetTicketsIDCommentsParams,
) error {
	ctx := c.Request().Context()

	ticket, err := h.repo.GetTicket(ctx, id)
	if err != nil {
		msg := err.Error()
		if errors.Is(err, tickets.ErrTicketNotFound) {
			return c.JSON(http.StatusNotFound, openapi.ErrorResponse{Message: &msg})
		}
		return c.JSON(http.StatusInternalServerError, openapi.ErrorResponse{Message: &msg})
	}

	// Determine which comments to include
	includeInternal := false
	if params.IncludeInternal != nil {
		includeInternal = *params.IncludeInternal
	}

	var comments []tickets.Comment
	if includeInternal {
		comments = ticket.Comments()
	} else {
		comments = ticket.GetPublicComments()
	}

	// Convert to OpenAPI responses
	commentResponses := make([]openapi.TicketComment, len(comments))
	for i, comment := range comments {
		commentResponses[i] = convertCommentToResponse(comment)
	}

	return c.JSON(http.StatusOK, commentResponses)
}

// convertCommentToResponse converts domain comment to OpenAPI response
func convertCommentToResponse(comment tickets.Comment) openapi.TicketComment {
	id := comment.ID
	ticketID := comment.TicketID
	authorID := comment.AuthorID
	content := comment.Content
	isInternal := comment.IsInternal
	createdAt := comment.CreatedAt

	return openapi.TicketComment{
		Id:         &id,
		TicketId:   &ticketID,
		AuthorId:   &authorID,
		Content:    &content,
		IsInternal: &isInternal,
		CreatedAt:  &createdAt,
	}
}
