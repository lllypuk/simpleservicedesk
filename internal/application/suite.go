package application

import (
	usersInfra "simpleservicedesk/internal/infrastructure/users"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/suite"
)

type ServerSuite struct {
	suite.Suite

	HTTPServer *echo.Echo
	UsersRepo  *usersInfra.InMemoryRepo
}

func (s *ServerSuite) SetupTest() {
	s.UsersRepo = usersInfra.NewInMemoryRepo()
	s.HTTPServer = SetupHTTPServer(s.UsersRepo)
}
