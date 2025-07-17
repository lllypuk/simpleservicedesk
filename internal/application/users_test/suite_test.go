package users_test

import (
	"testing"

	"simpleservicedesk/internal/application"

	"github.com/stretchr/testify/suite"
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
