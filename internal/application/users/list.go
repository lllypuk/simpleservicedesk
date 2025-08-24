package users

import (
	"net/http"

	"simpleservicedesk/generated/openapi"
	"simpleservicedesk/internal/domain/users"
	infraUsers "simpleservicedesk/internal/infrastructure/users"

	"github.com/labstack/echo/v4"
)

func (h UserHandlers) GetUsers(c echo.Context, params openapi.GetUsersParams) error {
	ctx := c.Request().Context()

	const defaultLimit = 20
	// Создаем фильтр на основе параметров
	filter := infraUsers.UserFilter{
		Offset: 0,
		Limit:  defaultLimit, // Значение по умолчанию
	}

	if params.Name != nil {
		filter.Name = *params.Name
	}
	if params.Email != nil {
		filter.Email = *params.Email
	}
	if params.Role != nil {
		role, err := users.ParseRole(string(*params.Role))
		if err == nil {
			filter.Role = &role
		}
	}
	if params.OrganizationId != nil {
		filter.OrganizationID = params.OrganizationId
	}
	if params.IsActive != nil {
		filter.IsActive = params.IsActive
	}

	// Обрабатываем пагинацию
	if params.Page != nil && *params.Page > 1 {
		filter.Offset = (*params.Page - 1) * filter.Limit
	}
	if params.Limit != nil && *params.Limit > 0 {
		filter.Limit = *params.Limit
		if params.Page != nil && *params.Page > 1 {
			filter.Offset = (*params.Page - 1) * filter.Limit
		}
	}

	// Получаем список пользователей
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
