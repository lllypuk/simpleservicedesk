//go:build integration
// +build integration

package shared

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"

	"simpleservicedesk/generated/openapi"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type E2ESeedData struct {
	OrganizationID uuid.UUID
	CategoryIDs    []uuid.UUID
}

func (s *IntegrationSuite) MustSeedE2EData() E2ESeedData {
	s.Require().NotEmpty(s.defaultAdminToken)

	orgName := fmt.Sprintf("E2E Org %s", uuid.NewString()[:8])
	orgDomain := fmt.Sprintf("e2e-%s.example.com", uuid.NewString()[:8])
	createOrgReq := openapi.CreateOrganizationRequest{
		Name:   orgName,
		Domain: &orgDomain,
	}

	orgReqBody, err := json.Marshal(createOrgReq)
	s.Require().NoError(err)

	orgReq := httptest.NewRequest(http.MethodPost, "/organizations", bytes.NewBuffer(orgReqBody))
	orgReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	orgRec := httptest.NewRecorder()
	s.ServeAuthenticatedHTTP(orgRec, orgReq)
	s.Require().Equal(http.StatusCreated, orgRec.Code, "response: %s", orgRec.Body.String())

	var orgResp openapi.CreateOrganizationResponse
	err = json.Unmarshal(orgRec.Body.Bytes(), &orgResp)
	s.Require().NoError(err)
	s.Require().NotNil(orgResp.Id)

	defaultCategories := []struct {
		name        string
		description string
	}{
		{name: "General Support", description: "Default category for general requests"},
		{name: "Platform Incidents", description: "Default category for incident tickets"},
	}
	categoryIDs := make([]uuid.UUID, 0, len(defaultCategories))

	for _, category := range defaultCategories {
		createCategoryReq := openapi.CreateCategoryRequest{
			Name:           category.name,
			Description:    &category.description,
			OrganizationId: *orgResp.Id,
		}
		categoryReqBody, marshalErr := json.Marshal(createCategoryReq)
		s.Require().NoError(marshalErr)

		categoryReq := httptest.NewRequest(http.MethodPost, "/categories", bytes.NewBuffer(categoryReqBody))
		categoryReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		categoryRec := httptest.NewRecorder()
		s.ServeAuthenticatedHTTP(categoryRec, categoryReq)
		s.Require().Equal(http.StatusCreated, categoryRec.Code, "response: %s", categoryRec.Body.String())

		var categoryResp openapi.CreateCategoryResponse
		unmarshalErr := json.Unmarshal(categoryRec.Body.Bytes(), &categoryResp)
		s.Require().NoError(unmarshalErr)
		s.Require().NotNil(categoryResp.Id)
		categoryIDs = append(categoryIDs, *categoryResp.Id)
	}

	return E2ESeedData{
		OrganizationID: *orgResp.Id,
		CategoryIDs:    categoryIDs,
	}
}
