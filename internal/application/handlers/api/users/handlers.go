package users

import (
	"simpleservicedesk/internal/application/services"
)

type UserHandlers struct {
	userService services.UserService
}

func SetupHandlers(userService services.UserService) UserHandlers {
	return UserHandlers{
		userService: userService,
	}
}
