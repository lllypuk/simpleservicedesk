package categories_test

import (
	"testing"

	"simpleservicedesk/internal/application"

	"github.com/stretchr/testify/suite"
)

type CategoriesSuite struct {
	suite.Suite
	application.ServerSuite
}

func (s *CategoriesSuite) SetupTest() {
	s.ServerSuite.SetupTest()
}

func TestCategoriesSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(CategoriesSuite))
}
