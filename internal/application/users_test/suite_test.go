package users_test

import (
	"testing"

	"github.com/stretchr/testify/suite"
	"simpleservicedesk/internal/application"
)

type UsersSuite struct {
	suite.Suite
	application.ServerSuite
}

func (s *UsersSuite) SetupTest() {
	s.ServerSuite.SetupTest()
}

func TestUsersSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(UsersSuite))
}
