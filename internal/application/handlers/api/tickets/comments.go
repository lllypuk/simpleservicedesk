package tickets

import (
	"errors"
	"net/http"

	"simpleservicedesk/generated/openapi"
	"simpleservicedesk/internal/application/handlers/api"
	"simpleservicedesk/internal/application/services"
	"simpleservicedesk/internal/domain/tickets"
	"simpleservicedesk/internal/queries"

	"github.com/labstack/echo/v4"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

func (h TicketHandlers) PostTicketsIDComments(c echo.Context, id openapi_types.UUID) error {
	ctx := c.Request().Context()

	var req openapi.CreateCommentRequest
	if err := c.Bind(&req); err != nil {
		msg := api.MsgInvalidRequestFormat
		return c.JSON(http.StatusBadRequest, openapi.ErrorResponse{Message: &msg})
	}

	// Validate content
	if req.Content == "" {
		msg := "comment content is required"
		return c.JSON(http.StatusBadRequest, openapi.ErrorResponse{Message: &msg})
	}

	// Create comment using service
	isInternal := false
	if req.IsInternal != nil {
		isInternal = *req.IsInternal
	}

	comment, err := h.ticketService.AddComment(ctx, id, req.AuthorId, services.AddTicketCommentRequest{
		Content:    req.Content,
		IsInternal: isInternal,
	})
	if err != nil {
		if errors.Is(err, tickets.ErrTicketNotFound) {
			msg := api.MsgTicketNotFound
			return c.JSON(http.StatusNotFound, openapi.ErrorResponse{Message: &msg})
		}
		msg := err.Error()
		return c.JSON(http.StatusInternalServerError, openapi.ErrorResponse{Message: &msg})
	}

	response := convertCommentToResponse(comment)
	return c.JSON(http.StatusCreated, response)
}

func (h TicketHandlers) GetTicketsIDComments(
	c echo.Context,
	id openapi_types.UUID,
	_ openapi.GetTicketsIDCommentsParams,
) error {
	ctx := c.Request().Context()

	// Set up filter with defaults
	filter := queries.CommentFilter{
		BaseFilter: queries.BaseFilter{
			Limit:  api.DefaultLimit,
			Offset: 0,
		},
	}

	// Get comments using service
	comments, total, err := h.ticketService.GetTicketComments(ctx, id, filter)
	if err != nil {
		if errors.Is(err, tickets.ErrTicketNotFound) {
			msg := api.MsgTicketNotFound
			return c.JSON(http.StatusNotFound, openapi.ErrorResponse{Message: &msg})
		}
		msg := err.Error()
		return c.JSON(http.StatusInternalServerError, openapi.ErrorResponse{Message: &msg})
	}

	// Convert comments to response format
	response := make([]openapi.TicketComment, len(comments))
	for i, comment := range comments {
		response[i] = convertCommentToResponse(comment)
	}

	totalInt := int(total)
	limit := filter.Limit
	pagination := &openapi.PaginationResponse{
		Page:  nil,
		Limit: &limit,
		Total: &totalInt,
	}

	// Use a simple response structure since ListCommentsResponse might not exist
	return c.JSON(http.StatusOK, map[string]interface{}{
		"comments":   response,
		"pagination": pagination,
	})
}

func convertCommentToResponse(comment *tickets.Comment) openapi.TicketComment {
	id := comment.ID
	content := comment.Content
	authorID := comment.AuthorID
	ticketID := comment.TicketID
	createdAt := comment.CreatedAt
	isInternal := comment.IsInternal

	return openapi.TicketComment{
		Id:         &id,
		Content:    &content,
		AuthorId:   &authorID,
		TicketId:   &ticketID,
		CreatedAt:  &createdAt,
		IsInternal: &isInternal,
	}
}
