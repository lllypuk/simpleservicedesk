package auth_test

import (
	"testing"

	"simpleservicedesk/internal/application"

	"github.com/stretchr/testify/suite"
)

type AuthSuite struct {
	application.ServerSuite
}

func (s *AuthSuite) SetupTest() {
	s.ServerSuite.SetupTest()
}

func TestAuthSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(AuthSuite))
}
