package tickets_test

import (
	"testing"

	"simpleservicedesk/internal/application"

	"github.com/stretchr/testify/suite"
)

type TicketsSuite struct {
	suite.Suite
	application.ServerSuite
}

func (s *TicketsSuite) SetupTest() {
	s.ServerSuite.SetupTest()
}

func TestTicketsSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(TicketsSuite))
}
