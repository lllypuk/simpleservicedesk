package organizations_test

import (
	"testing"

	"simpleservicedesk/internal/application"

	"github.com/stretchr/testify/suite"
)

type OrganizationsSuite struct {
	application.ServerSuite
}

func (s *OrganizationsSuite) SetupTest() {
	s.ServerSuite.SetupTest()
}

func TestOrganizationsSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(OrganizationsSuite))
}
