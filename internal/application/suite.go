package application

import (
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/suite"

	"simpleservicedesk/internal/application/mocks"
)

type ServerSuite struct {
	suite.Suite

	HTTPServer *echo.Echo
	UsersRepo  UserRepository // Используем интерфейс вместо конкретного типа
}

// SetupTest для интеграционных тестов с реальной MongoDB
func (s *ServerSuite) SetupTest() {
	s.UsersRepo = mocks.NewUserRepositoryMock()
	s.HTTPServer = SetupHTTPServer(s.UsersRepo)
}

// GetMockRepo возвращает мок репозитория (для настройки ожиданий в тестах)
func (s *ServerSuite) GetMockRepo() *mocks.UserRepositoryMock {
	if mockRepo, ok := s.UsersRepo.(*mocks.UserRepositoryMock); ok {
		return mockRepo
	}
	return nil
}
