//go:build integration && e2e
// +build integration,e2e

package e2e_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"simpleservicedesk/generated/openapi"
	userdomain "simpleservicedesk/internal/domain/users"
	"simpleservicedesk/test/integration/shared"

	"github.com/labstack/echo/v4"
	openapi_types "github.com/oapi-codegen/runtime/types"
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

func (s *E2ETestSuite) TestTicketLifecycleWorkflow() {
	agentName := fmt.Sprintf("E2E Agent %s", s.seedData.OrganizationID.String()[:8])
	agentEmail := fmt.Sprintf("agent-%s@example.com", s.seedData.OrganizationID.String()[:8])
	agentPassword := "agentpass123"

	createUserBody, err := json.Marshal(openapi.CreateUserRequest{
		Name:     agentName,
		Email:    openapi_types.Email(agentEmail),
		Password: agentPassword,
	})
	s.Require().NoError(err)

	createAgentReq := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(createUserBody))
	createAgentReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	createAgentRec := httptest.NewRecorder()
	s.ServeAuthenticatedHTTP(createAgentRec, createAgentReq)
	s.Require().Equal(http.StatusCreated, createAgentRec.Code, "response: %s", createAgentRec.Body.String())

	var createUserResp openapi.CreateUserResponse
	err = json.Unmarshal(createAgentRec.Body.Bytes(), &createUserResp)
	s.Require().NoError(err)
	s.Require().NotNil(createUserResp.Id)

	assignRoleBody, err := json.Marshal(openapi.UpdateUserRoleRequest{Role: openapi.Agent})
	s.Require().NoError(err)

	assignRoleReq := httptest.NewRequest(
		http.MethodPatch,
		fmt.Sprintf("/users/%s/role", createUserResp.Id.String()),
		bytes.NewReader(assignRoleBody),
	)
	assignRoleReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	assignRoleRec := httptest.NewRecorder()
	s.ServeAuthenticatedHTTP(assignRoleRec, assignRoleReq)
	s.Require().Equal(http.StatusOK, assignRoleRec.Code, "response: %s", assignRoleRec.Body.String())

	var roleResp openapi.GetUserResponse
	err = json.Unmarshal(assignRoleRec.Body.Bytes(), &roleResp)
	s.Require().NoError(err)
	s.Require().NotNil(roleResp.Role)
	s.Equal(openapi.Agent, *roleResp.Role)

	loginRec := s.Login(agentEmail, agentPassword)
	s.Require().Equal(http.StatusOK, loginRec.Code, "response: %s", loginRec.Body.String())
	agentToken := s.GetTokenFromLoginResponse(loginRec)
	s.Require().NotEmpty(agentToken)

	createTicketBody, err := json.Marshal(openapi.CreateTicketRequest{
		Title:          "E2E lifecycle ticket",
		Description:    "Ticket for full lifecycle workflow test",
		Priority:       openapi.High,
		OrganizationId: s.seedData.OrganizationID,
		AuthorId:       *createUserResp.Id,
		CategoryId:     &s.seedData.CategoryIDs[0],
	})
	s.Require().NoError(err)

	createTicketReq := httptest.NewRequest(http.MethodPost, "/tickets", bytes.NewReader(createTicketBody))
	createTicketReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	createTicketReq.Header.Set(echo.HeaderAuthorization, "Bearer "+agentToken)
	createTicketRec := httptest.NewRecorder()
	s.HTTPServer.ServeHTTP(createTicketRec, createTicketReq)
	s.Require().Equal(http.StatusCreated, createTicketRec.Code, "response: %s", createTicketRec.Body.String())

	var createdTicket openapi.GetTicketResponse
	err = json.Unmarshal(createTicketRec.Body.Bytes(), &createdTicket)
	s.Require().NoError(err)
	s.Require().NotNil(createdTicket.Id)
	s.Require().NotNil(createdTicket.Status)
	s.Equal(openapi.New, *createdTicket.Status)
	s.Nil(createdTicket.ResolvedAt)
	s.Nil(createdTicket.ClosedAt)

	invalidTransitionResp := s.mustUpdateTicketStatusWithToken(*createdTicket.Id, openapi.Resolved, agentToken)
	s.Require().Equal(http.StatusBadRequest, invalidTransitionResp.Code, "response: %s", invalidTransitionResp.Body.String())
	s.Contains(strings.ToLower(invalidTransitionResp.Body.String()), "invalid status transition")

	inProgressResp := s.mustUpdateTicketStatusWithToken(*createdTicket.Id, openapi.InProgress, agentToken)
	s.Require().Equal(http.StatusOK, inProgressResp.Code, "response: %s", inProgressResp.Body.String())
	var inProgressTicket openapi.GetTicketResponse
	err = json.Unmarshal(inProgressResp.Body.Bytes(), &inProgressTicket)
	s.Require().NoError(err)
	s.Require().NotNil(inProgressTicket.Status)
	s.Equal(openapi.InProgress, *inProgressTicket.Status)
	s.Nil(inProgressTicket.ResolvedAt)
	s.Nil(inProgressTicket.ClosedAt)

	resolvedResp := s.mustUpdateTicketStatusWithToken(*createdTicket.Id, openapi.Resolved, agentToken)
	s.Require().Equal(http.StatusOK, resolvedResp.Code, "response: %s", resolvedResp.Body.String())
	var resolvedTicket openapi.GetTicketResponse
	err = json.Unmarshal(resolvedResp.Body.Bytes(), &resolvedTicket)
	s.Require().NoError(err)
	s.Require().NotNil(resolvedTicket.Status)
	s.Equal(openapi.Resolved, *resolvedTicket.Status)
	s.Require().NotNil(resolvedTicket.ResolvedAt)
	s.Nil(resolvedTicket.ClosedAt)

	closeResp := s.mustUpdateTicketStatusWithToken(*createdTicket.Id, openapi.Closed, s.seedData.AdminToken)
	s.Require().Equal(http.StatusOK, closeResp.Code, "response: %s", closeResp.Body.String())
	var closedTicket openapi.GetTicketResponse
	err = json.Unmarshal(closeResp.Body.Bytes(), &closedTicket)
	s.Require().NoError(err)
	s.Require().NotNil(closedTicket.Status)
	s.Equal(openapi.Closed, *closedTicket.Status)
	s.Require().NotNil(closedTicket.ResolvedAt)
	s.Require().NotNil(closedTicket.ClosedAt)
}

func (s *E2ETestSuite) mustUpdateTicketStatusWithToken(
	ticketID fmt.Stringer,
	status openapi.TicketStatus,
	token string,
) *httptest.ResponseRecorder {
	reqBody, err := json.Marshal(openapi.UpdateTicketStatusRequest{Status: status})
	s.Require().NoError(err)

	req := httptest.NewRequest(
		http.MethodPatch,
		fmt.Sprintf("/tickets/%s/status", ticketID.String()),
		bytes.NewReader(reqBody),
	)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(echo.HeaderAuthorization, "Bearer "+token)
	rec := httptest.NewRecorder()
	s.HTTPServer.ServeHTTP(rec, req)

	return rec
}
