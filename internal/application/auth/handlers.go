package auth

import "context"

type LoginService interface {
	Login(ctx context.Context, email, password string) (string, error)
}

type Handlers struct {
	service LoginService
}

func SetupHandlers(service LoginService) Handlers {
	return Handlers{
		service: service,
	}
}
