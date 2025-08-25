package users

import (
	"net/http"

	"simpleservicedesk/generated/openapi"
	"simpleservicedesk/internal/queries"

	"github.com/labstack/echo/v4"
)

func (h UserHandlers) GetUsers(c echo.Context, params openapi.GetUsersParams) error {
	ctx := c.Request().Context()

	// Convert OpenAPI params to filter using the centralized converter
	filter, err := queries.FromOpenAPIUserParams(params)
	if err != nil {
		msg := err.Error()
		return c.JSON(http.StatusBadRequest, openapi.ErrorResponse{Message: &msg})
	}

	// Validate filter with business rules
	filter, validateErr := filter.ValidateAndSetDefaults()
	if validateErr != nil {
		msg := validateErr.Error()
		return c.JSON(http.StatusBadRequest, openapi.ErrorResponse{Message: &msg})
	}
	// Get users from repository
	usersList, err := h.repo.ListUsers(ctx, filter)
	if err != nil {
		return handleUserError(c, err)
	}

	// Получаем общее количество для пагинации
	total, err := h.repo.CountUsers(ctx, filter)
	if err != nil {
		return handleUserError(c, err)
	}

	// Преобразуем в response
	var userResponses []openapi.GetUserResponse
	for _, user := range usersList {
		userResponses = append(userResponses, userToResponse(user))
	}

	// Создаем пагинацию
	page := 1
	if params.Page != nil {
		page = *params.Page
	}
	limit := filter.Limit
	totalPages := int((total + int64(limit) - 1) / int64(limit))
	hasNext := page < totalPages

	totalInt := int(total)
	pagination := openapi.PaginationResponse{
		Total:   &totalInt,
		Page:    &page,
		Limit:   &limit,
		HasNext: &hasNext,
	}

	response := openapi.ListUsersResponse{
		Users:      &userResponses,
		Pagination: &pagination,
	}

	return c.JSON(http.StatusOK, response)
}
