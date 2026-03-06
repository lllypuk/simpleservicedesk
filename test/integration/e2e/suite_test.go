//go:build integration && e2e
// +build integration,e2e

package e2e_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"simpleservicedesk/generated/openapi"
	userdomain "simpleservicedesk/internal/domain/users"
	"simpleservicedesk/test/integration/shared"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/suite"
)

type E2ETestSuite struct {
	shared.IntegrationSuite
	seedData shared.E2ESeedData
}

func TestE2EInfrastructure(t *testing.T) {
	suite.Run(t, new(E2ETestSuite))
}

func (s *E2ETestSuite) SetupTest() {
	s.IntegrationSuite.SetupTest()
	s.seedData = s.MustSeedE2EData()
}

func (s *E2ETestSuite) TestSeedDataIncludesAdminOrganizationAndCategories() {
	s.Require().NotEmpty(s.seedData.AdminToken)
	s.Require().NotZero(s.seedData.OrganizationID)
	s.Require().Len(s.seedData.CategoryIDs, 2)

	getOrgReq := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/organizations/%s", s.seedData.OrganizationID.String()), nil)
	getOrgRec := httptest.NewRecorder()
	s.ServeAuthenticatedHTTP(getOrgRec, getOrgReq)
	s.Require().Equal(http.StatusOK, getOrgRec.Code, "response: %s", getOrgRec.Body.String())

	listCategoriesReq := httptest.NewRequest(
		http.MethodGet,
		fmt.Sprintf("/categories?organization_id=%s", s.seedData.OrganizationID.String()),
		nil,
	)
	listCategoriesRec := httptest.NewRecorder()
	s.ServeAuthenticatedHTTP(listCategoriesRec, listCategoriesReq)
	s.Require().Equal(http.StatusOK, listCategoriesRec.Code, "response: %s", listCategoriesRec.Body.String())

	var categoriesResp openapi.ListCategoriesResponse
	err := json.Unmarshal(listCategoriesRec.Body.Bytes(), &categoriesResp)
	s.Require().NoError(err)
	s.Require().NotNil(categoriesResp.Categories)
	s.GreaterOrEqual(len(*categoriesResp.Categories), 2)
}

func (s *E2ETestSuite) TestLoginAndTokenHelpers() {
	customer := s.MustCreateTestUser(userdomain.RoleCustomer)
	loginRec := s.Login(customer.Email, customer.Passphrase)
	s.Require().Equal(http.StatusOK, loginRec.Code, "response: %s", loginRec.Body.String())

	token := s.GetTokenFromLoginResponse(loginRec)
	s.Require().NotEmpty(token)

	reqBody, err := json.Marshal(openapi.CreateTicketRequest{
		Title:          "E2E helper validation ticket",
		Description:    "Created through login/token helper flow",
		Priority:       openapi.Normal,
		OrganizationId: s.seedData.OrganizationID,
		AuthorId:       customer.UserID,
		CategoryId:     &s.seedData.CategoryIDs[0],
	})
	s.Require().NoError(err)

	createTicketReq := httptest.NewRequest(http.MethodPost, "/tickets", bytes.NewReader(reqBody))
	createTicketReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	createTicketReq.Header.Set(echo.HeaderAuthorization, "Bearer "+token)
	createTicketRec := httptest.NewRecorder()

	s.HTTPServer.ServeHTTP(createTicketRec, createTicketReq)
	s.Require().Equal(http.StatusCreated, createTicketRec.Code, "response: %s", createTicketRec.Body.String())
}
