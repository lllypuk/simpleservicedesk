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

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/stretchr/testify/suite"
)

type E2ETestSuite struct {
	shared.IntegrationSuite
	seedData shared.E2ESeedData
}

type e2eUserCredentials struct {
	ID       openapi_types.UUID
	Email    string
	Password string
}

func TestE2EInfrastructure(t *testing.T) {
	suite.Run(t, new(E2ETestSuite))
}

func (s *E2ETestSuite) SetupTest() {
	s.IntegrationSuite.SetupTest()
	s.seedData = s.MustSeedE2EData()
}

func (s *E2ETestSuite) TestSeedDataIncludesAdminOrganizationAndCategories() {
	s.Require().NotEmpty(s.DefaultAdminToken())
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

	invalidLoginRec := s.Login(customer.Email, customer.Passphrase+"-wrong")
	s.Require().Equal(http.StatusUnauthorized, invalidLoginRec.Code, "response: %s", invalidLoginRec.Body.String())
	s.Empty(s.GetTokenFromLoginResponse(invalidLoginRec))
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
	resolvedAt := resolvedTicket.ResolvedAt.UTC()

	closeResp := s.mustUpdateTicketStatusWithToken(*createdTicket.Id, openapi.Closed, s.DefaultAdminToken())
	s.Require().Equal(http.StatusOK, closeResp.Code, "response: %s", closeResp.Body.String())
	var closedTicket openapi.GetTicketResponse
	err = json.Unmarshal(closeResp.Body.Bytes(), &closedTicket)
	s.Require().NoError(err)
	s.Require().NotNil(closedTicket.Status)
	s.Equal(openapi.Closed, *closedTicket.Status)
	s.Require().NotNil(closedTicket.ResolvedAt)
	s.Require().NotNil(closedTicket.ClosedAt)
	closedResolvedAt := closedTicket.ResolvedAt.UTC()
	closedAt := closedTicket.ClosedAt.UTC()
	s.Equal(resolvedAt, closedResolvedAt)
	s.False(closedAt.Before(resolvedAt))
}

func (s *E2ETestSuite) TestUserManagementWorkflow() {
	customerA := s.mustCreateUserWithAdmin("E2E Customer A", "customer.a", "customerApass123")
	customerB := s.mustCreateUserWithAdmin("E2E Customer B", "customer.b", "customerBpass123")
	agent := s.mustCreateUserWithAdmin("E2E Agent", "agent.user", "agentPass123")

	agentRoleReqBody, err := json.Marshal(openapi.UpdateUserRoleRequest{Role: openapi.Agent})
	s.Require().NoError(err)
	agentRoleReq := httptest.NewRequest(
		http.MethodPatch,
		fmt.Sprintf("/users/%s/role", agent.ID.String()),
		bytes.NewReader(agentRoleReqBody),
	)
	agentRoleReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	agentRoleRec := httptest.NewRecorder()
	s.ServeAuthenticatedHTTP(agentRoleRec, agentRoleReq)
	s.Require().Equal(http.StatusOK, agentRoleRec.Code, "response: %s", agentRoleRec.Body.String())

	customerAToken, customerALoginRec := s.LoginAndGetToken(customerA.Email, customerA.Password)
	s.Require().Equal(http.StatusOK, customerALoginRec.Code, "response: %s", customerALoginRec.Body.String())
	s.Require().NotEmpty(customerAToken)

	customerBToken, customerBLoginRec := s.LoginAndGetToken(customerB.Email, customerB.Password)
	s.Require().Equal(http.StatusOK, customerBLoginRec.Code, "response: %s", customerBLoginRec.Body.String())
	s.Require().NotEmpty(customerBToken)

	agentToken, agentLoginRec := s.LoginAndGetToken(agent.Email, agent.Password)
	s.Require().Equal(http.StatusOK, agentLoginRec.Code, "response: %s", agentLoginRec.Body.String())
	s.Require().NotEmpty(agentToken)

	customerATicket := s.mustCreateTicketWithToken(customerAToken, customerA.ID, "E2E visibility ticket A")
	customerBTicketForVisibility := s.mustCreateTicketWithToken(customerBToken, customerB.ID, "E2E visibility ticket B")
	customerBTicketForRoleChange := s.mustCreateTicketWithToken(customerBToken, customerB.ID, "E2E role update ticket B")

	customerAListRec := s.mustListOwnTicketsWithToken(customerAToken)
	s.Require().Equal(http.StatusOK, customerAListRec.Code, "response: %s", customerAListRec.Body.String())
	var customerAList openapi.ListTicketsResponse
	err = json.Unmarshal(customerAListRec.Body.Bytes(), &customerAList)
	s.Require().NoError(err)
	s.Require().NotNil(customerAList.Tickets)
	s.Len(*customerAList.Tickets, 1)
	s.Require().NotNil((*customerAList.Tickets)[0].Id)
	s.Require().NotNil((*customerAList.Tickets)[0].AuthorId)
	s.Equal(*customerATicket.Id, *(*customerAList.Tickets)[0].Id)
	s.Equal(customerA.ID, *(*customerAList.Tickets)[0].AuthorId)

	customerAFilteredAsOtherRec := s.mustListTicketsWithToken(customerAToken, customerB.ID.String())
	s.Require().Equal(http.StatusOK, customerAFilteredAsOtherRec.Code, "response: %s", customerAFilteredAsOtherRec.Body.String())
	var customerAFilteredAsOther openapi.ListTicketsResponse
	err = json.Unmarshal(customerAFilteredAsOtherRec.Body.Bytes(), &customerAFilteredAsOther)
	s.Require().NoError(err)
	s.Require().NotNil(customerAFilteredAsOther.Tickets)
	s.Len(*customerAFilteredAsOther.Tickets, 1)
	s.Require().NotNil((*customerAFilteredAsOther.Tickets)[0].AuthorId)
	s.Equal(customerA.ID, *(*customerAFilteredAsOther.Tickets)[0].AuthorId)

	customerBListRec := s.mustListOwnTicketsWithToken(customerBToken)
	s.Require().Equal(http.StatusOK, customerBListRec.Code, "response: %s", customerBListRec.Body.String())
	var customerBList openapi.ListTicketsResponse
	err = json.Unmarshal(customerBListRec.Body.Bytes(), &customerBList)
	s.Require().NoError(err)
	s.Require().NotNil(customerBList.Tickets)
	s.Len(*customerBList.Tickets, 2)
	for _, ticket := range *customerBList.Tickets {
		s.Require().NotNil(ticket.AuthorId)
		s.Equal(customerB.ID, *ticket.AuthorId)
	}

	customerAOldRoleUpdateRec := s.mustUpdateTicketStatusWithToken(*customerBTicketForRoleChange.Id, openapi.InProgress, customerAToken)
	s.Require().Equal(http.StatusForbidden, customerAOldRoleUpdateRec.Code, "response: %s", customerAOldRoleUpdateRec.Body.String())

	customerARoleReqBody, err := json.Marshal(openapi.UpdateUserRoleRequest{Role: openapi.Agent})
	s.Require().NoError(err)
	customerARoleReq := httptest.NewRequest(
		http.MethodPatch,
		fmt.Sprintf("/users/%s/role", customerA.ID.String()),
		bytes.NewReader(customerARoleReqBody),
	)
	customerARoleReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	customerARoleRec := httptest.NewRecorder()
	s.ServeAuthenticatedHTTP(customerARoleRec, customerARoleReq)
	s.Require().Equal(http.StatusOK, customerARoleRec.Code, "response: %s", customerARoleRec.Body.String())

	customerAOldTokenAfterRoleChangeRec := s.mustUpdateTicketStatusWithToken(
		*customerBTicketForRoleChange.Id,
		openapi.InProgress,
		customerAToken,
	)
	s.Require().Equal(
		http.StatusForbidden,
		customerAOldTokenAfterRoleChangeRec.Code,
		"response: %s",
		customerAOldTokenAfterRoleChangeRec.Body.String(),
	)

	customerANewToken, customerANewLoginRec := s.LoginAndGetToken(customerA.Email, customerA.Password)
	s.Require().Equal(http.StatusOK, customerANewLoginRec.Code, "response: %s", customerANewLoginRec.Body.String())
	s.Require().NotEmpty(customerANewToken)

	customerANewRoleUpdateRec := s.mustUpdateTicketStatusWithToken(*customerBTicketForRoleChange.Id, openapi.InProgress, customerANewToken)
	s.Require().Equal(http.StatusOK, customerANewRoleUpdateRec.Code, "response: %s", customerANewRoleUpdateRec.Body.String())
	var updatedByNewRole openapi.GetTicketResponse
	err = json.Unmarshal(customerANewRoleUpdateRec.Body.Bytes(), &updatedByNewRole)
	s.Require().NoError(err)
	s.Require().NotNil(updatedByNewRole.Status)
	s.Equal(openapi.InProgress, *updatedByNewRole.Status)

	agentUpdateRec := s.mustUpdateTicketStatusWithToken(*customerATicket.Id, openapi.InProgress, agentToken)
	s.Require().Equal(http.StatusOK, agentUpdateRec.Code, "response: %s", agentUpdateRec.Body.String())

	agentGetsCustomerBTicketRec := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/tickets/%s", customerBTicketForVisibility.Id.String()), nil)
	agentGetsCustomerBTicketRec.Header.Set(echo.HeaderAuthorization, "Bearer "+agentToken)
	agentGetsCustomerBTicketResponse := httptest.NewRecorder()
	s.HTTPServer.ServeHTTP(agentGetsCustomerBTicketResponse, agentGetsCustomerBTicketRec)
	s.Require().Equal(http.StatusOK, agentGetsCustomerBTicketResponse.Code, "response: %s", agentGetsCustomerBTicketResponse.Body.String())

	customerAGetsCustomerBTicketReq := httptest.NewRequest(
		http.MethodGet,
		fmt.Sprintf("/tickets/%s", customerBTicketForVisibility.Id.String()),
		nil,
	)
	customerAGetsCustomerBTicketReq.Header.Set(echo.HeaderAuthorization, "Bearer "+customerAToken)
	customerAGetsCustomerBTicketRec := httptest.NewRecorder()
	s.HTTPServer.ServeHTTP(customerAGetsCustomerBTicketRec, customerAGetsCustomerBTicketReq)
	s.Require().Equal(http.StatusForbidden, customerAGetsCustomerBTicketRec.Code, "response: %s", customerAGetsCustomerBTicketRec.Body.String())

	customerBGetsCustomerATicketReq := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/tickets/%s", customerATicket.Id.String()), nil)
	customerBGetsCustomerATicketReq.Header.Set(echo.HeaderAuthorization, "Bearer "+customerBToken)
	customerBGetsCustomerATicketRec := httptest.NewRecorder()
	s.HTTPServer.ServeHTTP(customerBGetsCustomerATicketRec, customerBGetsCustomerATicketReq)
	s.Require().Equal(http.StatusForbidden, customerBGetsCustomerATicketRec.Code, "response: %s", customerBGetsCustomerATicketRec.Body.String())
}

func (s *E2ETestSuite) TestOrganizationWorkflow() {
	orgName := fmt.Sprintf("E2E Organization %s", s.seedData.OrganizationID.String()[:8])
	orgDomain := fmt.Sprintf("e2e-org-%s.example.com", s.seedData.OrganizationID.String()[:8])
	organizationID := s.mustCreateOrganizationWithAdmin(orgName, orgDomain)
	categoryID := s.mustCreateCategoryWithAdmin(organizationID, "E2E Organization Category")

	customer := s.mustCreateUserWithAdmin("E2E Org Customer", "org.customer", "orgCustomerPass123")
	agent := s.mustCreateUserWithAdmin("E2E Org Agent", "org.agent", "orgAgentPass123")
	s.mustUpdateUserRoleWithAdmin(agent.ID, openapi.Agent)

	s.mustAssignUserToOrganizationWithAdmin(customer.ID, organizationID)
	s.mustAssignUserToOrganizationWithAdmin(agent.ID, organizationID)

	customerRec := s.mustGetUserWithAdmin(customer.ID)
	var customerResp openapi.GetUserResponse
	err := json.Unmarshal(customerRec.Body.Bytes(), &customerResp)
	s.Require().NoError(err)
	s.Require().NotNil(customerResp.OrganizationId)
	s.Equal(organizationID, *customerResp.OrganizationId)

	agentRec := s.mustGetUserWithAdmin(agent.ID)
	var agentResp openapi.GetUserResponse
	err = json.Unmarshal(agentRec.Body.Bytes(), &agentResp)
	s.Require().NoError(err)
	s.Require().NotNil(agentResp.OrganizationId)
	s.Equal(organizationID, *agentResp.OrganizationId)

	customerToken, customerLoginRec := s.LoginAndGetToken(customer.Email, customer.Password)
	s.Require().Equal(http.StatusOK, customerLoginRec.Code, "response: %s", customerLoginRec.Body.String())
	s.Require().NotEmpty(customerToken)

	agentToken, agentLoginRec := s.LoginAndGetToken(agent.Email, agent.Password)
	s.Require().Equal(http.StatusOK, agentLoginRec.Code, "response: %s", agentLoginRec.Body.String())
	s.Require().NotEmpty(agentToken)

	orgTicketBody, err := json.Marshal(openapi.CreateTicketRequest{
		Title:          "E2E organization workflow ticket",
		Description:    "Ticket created under dedicated organization",
		Priority:       openapi.Normal,
		OrganizationId: organizationID,
		AuthorId:       customer.ID,
		CategoryId:     &categoryID,
	})
	s.Require().NoError(err)

	orgTicketReq := httptest.NewRequest(http.MethodPost, "/tickets", bytes.NewReader(orgTicketBody))
	orgTicketReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	orgTicketReq.Header.Set(echo.HeaderAuthorization, "Bearer "+customerToken)
	orgTicketRec := httptest.NewRecorder()
	s.HTTPServer.ServeHTTP(orgTicketRec, orgTicketReq)
	s.Require().Equal(http.StatusCreated, orgTicketRec.Code, "response: %s", orgTicketRec.Body.String())

	var orgTicketResp openapi.GetTicketResponse
	err = json.Unmarshal(orgTicketRec.Body.Bytes(), &orgTicketResp)
	s.Require().NoError(err)
	s.Require().NotNil(orgTicketResp.Id)
	s.Require().NotNil(orgTicketResp.OrganizationId)
	s.Equal(organizationID, *orgTicketResp.OrganizationId)

	otherOrgTicketBody, err := json.Marshal(openapi.CreateTicketRequest{
		Title:          "E2E other organization ticket",
		Description:    "Ticket created in seed organization for filtering checks",
		Priority:       openapi.Normal,
		OrganizationId: s.seedData.OrganizationID,
		AuthorId:       customer.ID,
		CategoryId:     &s.seedData.CategoryIDs[0],
	})
	s.Require().NoError(err)

	otherOrgTicketReq := httptest.NewRequest(http.MethodPost, "/tickets", bytes.NewReader(otherOrgTicketBody))
	otherOrgTicketReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	otherOrgTicketReq.Header.Set(echo.HeaderAuthorization, "Bearer "+customerToken)
	otherOrgTicketRec := httptest.NewRecorder()
	s.HTTPServer.ServeHTTP(otherOrgTicketRec, otherOrgTicketReq)
	s.Require().Equal(http.StatusCreated, otherOrgTicketRec.Code, "response: %s", otherOrgTicketRec.Body.String())

	filteredByOrganizationRec := s.mustListTicketsByOrganizationWithToken(agentToken, organizationID)
	s.Require().Equal(http.StatusOK, filteredByOrganizationRec.Code, "response: %s", filteredByOrganizationRec.Body.String())
	var filteredByOrganization openapi.ListTicketsResponse
	err = json.Unmarshal(filteredByOrganizationRec.Body.Bytes(), &filteredByOrganization)
	s.Require().NoError(err)
	s.Require().NotNil(filteredByOrganization.Tickets)
	s.Require().Len(*filteredByOrganization.Tickets, 1)
	s.Require().NotNil((*filteredByOrganization.Tickets)[0].Id)
	s.Require().NotNil((*filteredByOrganization.Tickets)[0].OrganizationId)
	s.Equal(*orgTicketResp.Id, *(*filteredByOrganization.Tickets)[0].Id)
	s.Equal(organizationID, *(*filteredByOrganization.Tickets)[0].OrganizationId)
}

func (s *E2ETestSuite) TestCategoryAndTicketClassificationWorkflow() {
	organizationID := s.mustCreateOrganizationWithAdmin(
		fmt.Sprintf("E2E Category Org %s", s.seedData.OrganizationID.String()[:8]),
		fmt.Sprintf("e2e-category-org-%s.example.com", s.seedData.OrganizationID.String()[:8]),
	)
	rootCategoryA := s.mustCreateCategoryWithAdmin(organizationID, "E2E Root Category A")
	rootCategoryB := s.mustCreateCategoryWithAdmin(organizationID, "E2E Root Category B")
	childCategory := s.mustCreateCategoryWithParentWithAdmin(organizationID, rootCategoryA, "E2E Child Category")

	customer := s.mustCreateUserWithAdmin("E2E Category Customer", "category.customer", "categoryCustomerPass123")
	s.mustAssignUserToOrganizationWithAdmin(customer.ID, organizationID)

	customerToken, customerLoginRec := s.LoginAndGetToken(customer.Email, customer.Password)
	s.Require().Equal(http.StatusOK, customerLoginRec.Code, "response: %s", customerLoginRec.Body.String())
	s.Require().NotEmpty(customerToken)

	ticketInRootA := s.mustCreateTicketWithCategoryAndOrganization(
		customerToken,
		customer.ID,
		organizationID,
		rootCategoryA,
		"E2E ticket in root category A",
	)
	ticketInChild := s.mustCreateTicketWithCategoryAndOrganization(
		customerToken,
		customer.ID,
		organizationID,
		childCategory,
		"E2E ticket in child category",
	)
	ticketInRootB := s.mustCreateTicketWithCategoryAndOrganization(
		customerToken,
		customer.ID,
		organizationID,
		rootCategoryB,
		"E2E ticket in root category B",
	)

	rootADirectRec := s.mustListTicketsByCategoryWithToken(customerToken, rootCategoryA)
	s.Require().Equal(http.StatusOK, rootADirectRec.Code, "response: %s", rootADirectRec.Body.String())
	rootADirectTickets := s.mustDecodeTicketsResponse(rootADirectRec)
	s.Require().Len(rootADirectTickets, 1)
	s.Require().NotNil(rootADirectTickets[0].Id)
	s.Equal(*ticketInRootA.Id, *rootADirectTickets[0].Id)

	rootAWithChildrenBeforeMoveRec := s.mustListCategoryTicketsWithToken(customerToken, rootCategoryA, true)
	s.Require().Equal(http.StatusOK, rootAWithChildrenBeforeMoveRec.Code, "response: %s", rootAWithChildrenBeforeMoveRec.Body.String())
	rootAWithChildrenBeforeMove := s.mustDecodeTicketsResponse(rootAWithChildrenBeforeMoveRec)
	s.Require().Len(rootAWithChildrenBeforeMove, 2)
	s.Contains(s.ticketIDs(rootAWithChildrenBeforeMove), *ticketInRootA.Id)
	s.Contains(s.ticketIDs(rootAWithChildrenBeforeMove), *ticketInChild.Id)

	rootADefaultBeforeMoveRec := s.mustListCategoryTicketsDefaultWithToken(customerToken, rootCategoryA)
	s.Require().Equal(http.StatusOK, rootADefaultBeforeMoveRec.Code, "response: %s", rootADefaultBeforeMoveRec.Body.String())
	rootADefaultBeforeMove := s.mustDecodeTicketsResponse(rootADefaultBeforeMoveRec)
	s.Require().Len(rootADefaultBeforeMove, 1)
	s.Contains(s.ticketIDs(rootADefaultBeforeMove), *ticketInRootA.Id)

	rootAWithoutChildrenBeforeMoveRec := s.mustListCategoryTicketsWithToken(customerToken, rootCategoryA, false)
	s.Require().Equal(http.StatusOK, rootAWithoutChildrenBeforeMoveRec.Code, "response: %s", rootAWithoutChildrenBeforeMoveRec.Body.String())
	rootAWithoutChildrenBeforeMove := s.mustDecodeTicketsResponse(rootAWithoutChildrenBeforeMoveRec)
	s.Require().Len(rootAWithoutChildrenBeforeMove, 1)
	s.Contains(s.ticketIDs(rootAWithoutChildrenBeforeMove), *ticketInRootA.Id)

	s.mustUpdateCategoryParentWithAdmin(childCategory, rootCategoryB)

	childCategoryRec := s.mustGetCategoryWithAdmin(childCategory)
	var childCategoryResp openapi.GetCategoryResponse
	err := json.Unmarshal(childCategoryRec.Body.Bytes(), &childCategoryResp)
	s.Require().NoError(err)
	s.Require().NotNil(childCategoryResp.ParentId)
	s.Equal(rootCategoryB, *childCategoryResp.ParentId)

	childDirectRec := s.mustListTicketsByCategoryWithToken(customerToken, childCategory)
	s.Require().Equal(http.StatusOK, childDirectRec.Code, "response: %s", childDirectRec.Body.String())
	childDirectTickets := s.mustDecodeTicketsResponse(childDirectRec)
	s.Require().Len(childDirectTickets, 1)
	s.Require().NotNil(childDirectTickets[0].Id)
	s.Equal(*ticketInChild.Id, *childDirectTickets[0].Id)

	rootAWithChildrenAfterMoveRec := s.mustListCategoryTicketsWithToken(customerToken, rootCategoryA, true)
	s.Require().Equal(http.StatusOK, rootAWithChildrenAfterMoveRec.Code, "response: %s", rootAWithChildrenAfterMoveRec.Body.String())
	rootAWithChildrenAfterMove := s.mustDecodeTicketsResponse(rootAWithChildrenAfterMoveRec)
	s.Require().Len(rootAWithChildrenAfterMove, 1)
	s.Contains(s.ticketIDs(rootAWithChildrenAfterMove), *ticketInRootA.Id)

	rootBWithChildrenAfterMoveRec := s.mustListCategoryTicketsWithToken(customerToken, rootCategoryB, true)
	s.Require().Equal(http.StatusOK, rootBWithChildrenAfterMoveRec.Code, "response: %s", rootBWithChildrenAfterMoveRec.Body.String())
	rootBWithChildrenAfterMove := s.mustDecodeTicketsResponse(rootBWithChildrenAfterMoveRec)
	s.Require().Len(rootBWithChildrenAfterMove, 2)
	s.Contains(s.ticketIDs(rootBWithChildrenAfterMove), *ticketInRootB.Id)
	s.Contains(s.ticketIDs(rootBWithChildrenAfterMove), *ticketInChild.Id)

	rootBWithoutChildrenAfterMoveRec := s.mustListCategoryTicketsWithToken(customerToken, rootCategoryB, false)
	s.Require().Equal(http.StatusOK, rootBWithoutChildrenAfterMoveRec.Code, "response: %s", rootBWithoutChildrenAfterMoveRec.Body.String())
	rootBWithoutChildrenAfterMove := s.mustDecodeTicketsResponse(rootBWithoutChildrenAfterMoveRec)
	s.Require().Len(rootBWithoutChildrenAfterMove, 1)
	s.Contains(s.ticketIDs(rootBWithoutChildrenAfterMove), *ticketInRootB.Id)
}

func (s *E2ETestSuite) TestErrorScenariosWorkflow() {
	customer := s.MustCreateTestUser(userdomain.RoleCustomer)
	customerToken, loginRec := s.LoginAndGetToken(customer.Email, customer.Passphrase)
	s.Require().Equal(http.StatusOK, loginRec.Code, "response: %s", loginRec.Body.String())
	s.Require().NotEmpty(customerToken)

	ticketResp := s.mustCreateTicketWithToken(customerToken, customer.UserID, "E2E error scenario ticket")
	invalidTransitionRec := s.mustUpdateTicketStatusWithToken(*ticketResp.Id, openapi.Resolved, s.DefaultAdminToken())
	s.Require().Equal(http.StatusBadRequest, invalidTransitionRec.Code, "response: %s", invalidTransitionRec.Body.String())
	s.Contains(strings.ToLower(invalidTransitionRec.Body.String()), "invalid status transition")

	email := openapi_types.Email(fmt.Sprintf("duplicate-%s@example.com", s.seedData.OrganizationID.String()[:8]))
	createUserBody, err := json.Marshal(openapi.CreateUserRequest{
		Name:     "Duplicate Email User",
		Email:    email,
		Password: "duplicatePass123",
	})
	s.Require().NoError(err)

	firstCreateReq := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(createUserBody))
	firstCreateReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	firstCreateRec := httptest.NewRecorder()
	s.ServeAuthenticatedHTTP(firstCreateRec, firstCreateReq)
	s.Require().Equal(http.StatusCreated, firstCreateRec.Code, "response: %s", firstCreateRec.Body.String())

	secondCreateReq := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(createUserBody))
	secondCreateReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	secondCreateRec := httptest.NewRecorder()
	s.ServeAuthenticatedHTTP(secondCreateRec, secondCreateReq)
	s.Require().Equal(http.StatusConflict, secondCreateRec.Code, "response: %s", secondCreateRec.Body.String())

	parentOrgID := s.mustCreateOrganizationWithAdmin(
		fmt.Sprintf("E2E Circular Parent Org %s", s.seedData.OrganizationID.String()[:8]),
		fmt.Sprintf("e2e-circular-parent-%s.example.com", s.seedData.OrganizationID.String()[:8]),
	)
	childOrgDomain := fmt.Sprintf("e2e-circular-child-%s.example.com", s.seedData.OrganizationID.String()[:8])
	childOrgBody, err := json.Marshal(openapi.CreateOrganizationRequest{
		Name:     fmt.Sprintf("E2E Circular Child Org %s", s.seedData.OrganizationID.String()[:8]),
		Domain:   &childOrgDomain,
		ParentId: &parentOrgID,
	})
	s.Require().NoError(err)

	createChildOrgReq := httptest.NewRequest(http.MethodPost, "/organizations", bytes.NewReader(childOrgBody))
	createChildOrgReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	createChildOrgRec := httptest.NewRecorder()
	s.ServeAuthenticatedHTTP(createChildOrgRec, createChildOrgReq)
	s.Require().Equal(http.StatusCreated, createChildOrgRec.Code, "response: %s", createChildOrgRec.Body.String())

	var childOrgResp openapi.CreateOrganizationResponse
	err = json.Unmarshal(createChildOrgRec.Body.Bytes(), &childOrgResp)
	s.Require().NoError(err)
	s.Require().NotNil(childOrgResp.Id)

	makeParentChildBody, err := json.Marshal(openapi.UpdateOrganizationRequest{
		ParentId: childOrgResp.Id,
	})
	s.Require().NoError(err)
	makeParentChildReq := httptest.NewRequest(
		http.MethodPut,
		"/organizations/"+parentOrgID.String(),
		bytes.NewReader(makeParentChildBody),
	)
	makeParentChildReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	makeParentChildRec := httptest.NewRecorder()
	s.ServeAuthenticatedHTTP(makeParentChildRec, makeParentChildReq)
	s.Require().Equal(http.StatusBadRequest, makeParentChildRec.Code, "response: %s", makeParentChildRec.Body.String())
	s.Contains(strings.ToLower(makeParentChildRec.Body.String()), "circular")

	rootCategoryID := s.mustCreateCategoryWithAdmin(parentOrgID, "E2E Circular Root Category")
	childCategoryID := s.mustCreateCategoryWithParentWithAdmin(parentOrgID, rootCategoryID, "E2E Circular Child Category")

	makeCategoryParentChildBody, err := json.Marshal(openapi.UpdateCategoryRequest{
		ParentId: &childCategoryID,
	})
	s.Require().NoError(err)
	makeCategoryParentChildReq := httptest.NewRequest(
		http.MethodPut,
		"/categories/"+rootCategoryID.String(),
		bytes.NewReader(makeCategoryParentChildBody),
	)
	makeCategoryParentChildReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	makeCategoryParentChildRec := httptest.NewRecorder()
	s.ServeAuthenticatedHTTP(makeCategoryParentChildRec, makeCategoryParentChildReq)
	s.Require().Equal(http.StatusBadRequest, makeCategoryParentChildRec.Code, "response: %s", makeCategoryParentChildRec.Body.String())
	s.Contains(strings.ToLower(makeCategoryParentChildRec.Body.String()), "circular")

	nonExistentID := uuid.NewString()
	getNonExistentTicketReq := httptest.NewRequest(http.MethodGet, "/tickets/"+nonExistentID, nil)
	getNonExistentTicketRec := httptest.NewRecorder()
	s.ServeAuthenticatedHTTP(getNonExistentTicketRec, getNonExistentTicketReq)
	s.Require().Equal(http.StatusNotFound, getNonExistentTicketRec.Code, "response: %s", getNonExistentTicketRec.Body.String())

	getNonExistentUserReq := httptest.NewRequest(http.MethodGet, "/users/"+uuid.NewString(), nil)
	getNonExistentUserRec := httptest.NewRecorder()
	s.ServeAuthenticatedHTTP(getNonExistentUserRec, getNonExistentUserReq)
	s.Require().Equal(http.StatusNotFound, getNonExistentUserRec.Code, "response: %s", getNonExistentUserRec.Body.String())
}

func (s *E2ETestSuite) mustCreateUserWithAdmin(name, emailPrefix, password string) e2eUserCredentials {
	email := openapi_types.Email(fmt.Sprintf("%s-%s@example.com", emailPrefix, s.seedData.OrganizationID.String()[:8]))
	reqBody, err := json.Marshal(openapi.CreateUserRequest{
		Name:     name,
		Email:    email,
		Password: password,
	})
	s.Require().NoError(err)

	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	s.ServeAuthenticatedHTTP(rec, req)
	s.Require().Equal(http.StatusCreated, rec.Code, "response: %s", rec.Body.String())

	var resp openapi.CreateUserResponse
	err = json.Unmarshal(rec.Body.Bytes(), &resp)
	s.Require().NoError(err)
	s.Require().NotNil(resp.Id)

	return e2eUserCredentials{
		ID:       *resp.Id,
		Email:    string(email),
		Password: password,
	}
}

func (s *E2ETestSuite) mustCreateTicketWithToken(
	token string,
	authorID openapi_types.UUID,
	title string,
) openapi.GetTicketResponse {
	reqBody, err := json.Marshal(openapi.CreateTicketRequest{
		Title:          title,
		Description:    "E2E user management workflow ticket",
		Priority:       openapi.Normal,
		OrganizationId: s.seedData.OrganizationID,
		AuthorId:       authorID,
		CategoryId:     &s.seedData.CategoryIDs[0],
	})
	s.Require().NoError(err)

	req := httptest.NewRequest(http.MethodPost, "/tickets", bytes.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(echo.HeaderAuthorization, "Bearer "+token)
	rec := httptest.NewRecorder()
	s.HTTPServer.ServeHTTP(rec, req)
	s.Require().Equal(http.StatusCreated, rec.Code, "response: %s", rec.Body.String())

	var ticketResp openapi.GetTicketResponse
	err = json.Unmarshal(rec.Body.Bytes(), &ticketResp)
	s.Require().NoError(err)
	s.Require().NotNil(ticketResp.Id)
	s.Require().NotNil(ticketResp.AuthorId)

	return ticketResp
}

func (s *E2ETestSuite) mustListTicketsWithToken(token, authorID string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, "/tickets?author_id="+authorID, nil)
	req.Header.Set(echo.HeaderAuthorization, "Bearer "+token)
	rec := httptest.NewRecorder()
	s.HTTPServer.ServeHTTP(rec, req)
	return rec
}

func (s *E2ETestSuite) mustListOwnTicketsWithToken(token string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, "/tickets", nil)
	req.Header.Set(echo.HeaderAuthorization, "Bearer "+token)
	rec := httptest.NewRecorder()
	s.HTTPServer.ServeHTTP(rec, req)
	return rec
}

func (s *E2ETestSuite) mustListTicketsByOrganizationWithToken(
	token string,
	organizationID openapi_types.UUID,
) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, "/tickets?organization_id="+organizationID.String(), nil)
	req.Header.Set(echo.HeaderAuthorization, "Bearer "+token)
	rec := httptest.NewRecorder()
	s.HTTPServer.ServeHTTP(rec, req)
	return rec
}

func (s *E2ETestSuite) mustCreateOrganizationWithAdmin(name, domain string) openapi_types.UUID {
	reqBody, err := json.Marshal(openapi.CreateOrganizationRequest{
		Name:   name,
		Domain: &domain,
	})
	s.Require().NoError(err)

	req := httptest.NewRequest(http.MethodPost, "/organizations", bytes.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	s.ServeAuthenticatedHTTP(rec, req)
	s.Require().Equal(http.StatusCreated, rec.Code, "response: %s", rec.Body.String())

	var resp openapi.CreateOrganizationResponse
	err = json.Unmarshal(rec.Body.Bytes(), &resp)
	s.Require().NoError(err)
	s.Require().NotNil(resp.Id)
	return *resp.Id
}

func (s *E2ETestSuite) mustCreateCategoryWithAdmin(
	organizationID openapi_types.UUID,
	name string,
) openapi_types.UUID {
	reqBody, err := json.Marshal(openapi.CreateCategoryRequest{
		Name:           name,
		OrganizationId: organizationID,
	})
	s.Require().NoError(err)

	req := httptest.NewRequest(http.MethodPost, "/categories", bytes.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	s.ServeAuthenticatedHTTP(rec, req)
	s.Require().Equal(http.StatusCreated, rec.Code, "response: %s", rec.Body.String())

	var resp openapi.CreateCategoryResponse
	err = json.Unmarshal(rec.Body.Bytes(), &resp)
	s.Require().NoError(err)
	s.Require().NotNil(resp.Id)
	return *resp.Id
}

func (s *E2ETestSuite) mustCreateCategoryWithParentWithAdmin(
	organizationID openapi_types.UUID,
	parentID openapi_types.UUID,
	name string,
) openapi_types.UUID {
	reqBody, err := json.Marshal(openapi.CreateCategoryRequest{
		Name:           name,
		OrganizationId: organizationID,
		ParentId:       &parentID,
	})
	s.Require().NoError(err)

	req := httptest.NewRequest(http.MethodPost, "/categories", bytes.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	s.ServeAuthenticatedHTTP(rec, req)
	s.Require().Equal(http.StatusCreated, rec.Code, "response: %s", rec.Body.String())

	var resp openapi.CreateCategoryResponse
	err = json.Unmarshal(rec.Body.Bytes(), &resp)
	s.Require().NoError(err)
	s.Require().NotNil(resp.Id)
	return *resp.Id
}

func (s *E2ETestSuite) mustUpdateCategoryParentWithAdmin(categoryID, parentID openapi_types.UUID) {
	reqBody, err := json.Marshal(openapi.UpdateCategoryRequest{
		ParentId: &parentID,
	})
	s.Require().NoError(err)

	req := httptest.NewRequest(http.MethodPut, "/categories/"+categoryID.String(), bytes.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	s.ServeAuthenticatedHTTP(rec, req)
	s.Require().Equal(http.StatusOK, rec.Code, "response: %s", rec.Body.String())
}

func (s *E2ETestSuite) mustGetCategoryWithAdmin(categoryID openapi_types.UUID) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, "/categories/"+categoryID.String(), nil)
	rec := httptest.NewRecorder()
	s.ServeAuthenticatedHTTP(rec, req)
	s.Require().Equal(http.StatusOK, rec.Code, "response: %s", rec.Body.String())
	return rec
}

func (s *E2ETestSuite) mustCreateTicketWithCategoryAndOrganization(
	token string,
	authorID openapi_types.UUID,
	organizationID openapi_types.UUID,
	categoryID openapi_types.UUID,
	title string,
) openapi.GetTicketResponse {
	reqBody, err := json.Marshal(openapi.CreateTicketRequest{
		Title:          title,
		Description:    "E2E category workflow ticket",
		Priority:       openapi.Normal,
		OrganizationId: organizationID,
		AuthorId:       authorID,
		CategoryId:     &categoryID,
	})
	s.Require().NoError(err)

	req := httptest.NewRequest(http.MethodPost, "/tickets", bytes.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(echo.HeaderAuthorization, "Bearer "+token)
	rec := httptest.NewRecorder()
	s.HTTPServer.ServeHTTP(rec, req)
	s.Require().Equal(http.StatusCreated, rec.Code, "response: %s", rec.Body.String())

	var ticketResp openapi.GetTicketResponse
	err = json.Unmarshal(rec.Body.Bytes(), &ticketResp)
	s.Require().NoError(err)
	s.Require().NotNil(ticketResp.Id)

	return ticketResp
}

func (s *E2ETestSuite) mustListTicketsByCategoryWithToken(
	token string,
	categoryID openapi_types.UUID,
) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, "/tickets?category_id="+categoryID.String(), nil)
	req.Header.Set(echo.HeaderAuthorization, "Bearer "+token)
	rec := httptest.NewRecorder()
	s.HTTPServer.ServeHTTP(rec, req)
	return rec
}

func (s *E2ETestSuite) mustListCategoryTicketsWithToken(
	token string,
	categoryID openapi_types.UUID,
	includeSubcategories bool,
) *httptest.ResponseRecorder {
	req := httptest.NewRequest(
		http.MethodGet,
		fmt.Sprintf("/categories/%s/tickets?include_subcategories=%t", categoryID.String(), includeSubcategories),
		nil,
	)
	req.Header.Set(echo.HeaderAuthorization, "Bearer "+token)
	rec := httptest.NewRecorder()
	s.HTTPServer.ServeHTTP(rec, req)
	return rec
}

func (s *E2ETestSuite) mustListCategoryTicketsDefaultWithToken(
	token string,
	categoryID openapi_types.UUID,
) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/categories/%s/tickets", categoryID.String()), nil)
	req.Header.Set(echo.HeaderAuthorization, "Bearer "+token)
	rec := httptest.NewRecorder()
	s.HTTPServer.ServeHTTP(rec, req)
	return rec
}

func (s *E2ETestSuite) mustDecodeTicketsResponse(rec *httptest.ResponseRecorder) []openapi.GetTicketResponse {
	var ticketsResp openapi.ListTicketsResponse
	err := json.Unmarshal(rec.Body.Bytes(), &ticketsResp)
	s.Require().NoError(err)
	s.Require().NotNil(ticketsResp.Tickets)
	return *ticketsResp.Tickets
}

func (s *E2ETestSuite) ticketIDs(tickets []openapi.GetTicketResponse) []openapi_types.UUID {
	ids := make([]openapi_types.UUID, 0, len(tickets))
	for _, ticket := range tickets {
		s.Require().NotNil(ticket.Id)
		ids = append(ids, *ticket.Id)
	}
	return ids
}

func (s *E2ETestSuite) mustAssignUserToOrganizationWithAdmin(userID, organizationID openapi_types.UUID) {
	reqBody, err := json.Marshal(openapi.UpdateUserRequest{
		OrganizationId: &organizationID,
	})
	s.Require().NoError(err)

	req := httptest.NewRequest(http.MethodPut, "/users/"+userID.String(), bytes.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	s.ServeAuthenticatedHTTP(rec, req)
	s.Require().Equal(http.StatusOK, rec.Code, "response: %s", rec.Body.String())
}

func (s *E2ETestSuite) mustUpdateUserRoleWithAdmin(userID openapi_types.UUID, role openapi.UserRole) {
	reqBody, err := json.Marshal(openapi.UpdateUserRoleRequest{
		Role: role,
	})
	s.Require().NoError(err)

	req := httptest.NewRequest(http.MethodPatch, "/users/"+userID.String()+"/role", bytes.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	s.ServeAuthenticatedHTTP(rec, req)
	s.Require().Equal(http.StatusOK, rec.Code, "response: %s", rec.Body.String())
}

func (s *E2ETestSuite) mustGetUserWithAdmin(userID openapi_types.UUID) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, "/users/"+userID.String(), nil)
	rec := httptest.NewRecorder()
	s.ServeAuthenticatedHTTP(rec, req)
	s.Require().Equal(http.StatusOK, rec.Code, "response: %s", rec.Body.String())
	return rec
}

func (s *E2ETestSuite) mustUpdateTicketStatusWithToken(
	ticketID openapi_types.UUID,
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
