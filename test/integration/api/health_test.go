//go:build integration
// +build integration

package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"simpleservicedesk/internal/application/health"
	"simpleservicedesk/test/integration/shared"

	"github.com/stretchr/testify/suite"
)

type HealthAPITestSuite struct {
	shared.IntegrationSuite
}

func TestHealthAPI(t *testing.T) {
	suite.Run(t, new(HealthAPITestSuite))
}

func (s *HealthAPITestSuite) TestLiveEndpointReturns200() {
	req := httptest.NewRequest(http.MethodGet, "/health/live", nil)
	rec := httptest.NewRecorder()
	s.HTTPServer.ServeHTTP(rec, req)

	s.Equal(http.StatusOK, rec.Code)

	var resp health.Status
	s.Require().NoError(json.Unmarshal(rec.Body.Bytes(), &resp))
	s.Equal("healthy", resp.Status)
	s.Empty(resp.Checks)
}

func (s *HealthAPITestSuite) TestReadyEndpointReturns200WithMongoUp() {
	req := httptest.NewRequest(http.MethodGet, "/health/ready", nil)
	rec := httptest.NewRecorder()
	s.HTTPServer.ServeHTTP(rec, req)

	s.Equal(http.StatusOK, rec.Code)

	var resp health.Status
	s.Require().NoError(json.Unmarshal(rec.Body.Bytes(), &resp))
	s.Equal("healthy", resp.Status)
	s.Require().Len(resp.Checks, 1)
	s.Equal("mongodb", resp.Checks[0].Name)
	s.Equal("up", resp.Checks[0].Status)
	s.GreaterOrEqual(resp.Checks[0].LatencyMs, int64(0))
}

func (s *HealthAPITestSuite) TestHealthEndpointsRequireNoAuth() {
	for _, path := range []string{"/health/live", "/health/ready"} {
		s.Run(path, func() {
			req := httptest.NewRequest(http.MethodGet, path, nil)
			rec := httptest.NewRecorder()
			s.HTTPServer.ServeHTTP(rec, req)

			s.NotEqual(http.StatusUnauthorized, rec.Code, "path %s should not require auth", path)
		})
	}
}
