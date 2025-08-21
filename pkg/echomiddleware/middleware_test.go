package echomiddleware_test

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"simpleservicedesk/pkg/contextkeys"
	echoMw "simpleservicedesk/pkg/echomiddleware"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock logger for testing
type mockLogger struct {
	logs []logEntry
}

type logEntry struct {
	level   slog.Level
	message string
	attrs   []slog.Attr
}

func (m *mockLogger) LogAttrs(_ context.Context, level slog.Level, msg string, attrs ...slog.Attr) {
	m.logs = append(m.logs, logEntry{
		level:   level,
		message: msg,
		attrs:   attrs,
	})
}

func (m *mockLogger) getLogs() []logEntry {
	return m.logs
}

func TestSlogLoggerMiddleware(t *testing.T) {
	t.Run("successful request", func(t *testing.T) {
		testSlogLoggerMiddleware(t, "GET", "/test", map[string]string{}, 200, nil, slog.LevelInfo,
			[]string{"path", "status_code", "method", "protocol", "remote_ip", "user_agent", "exec_time"})
	})

	t.Run("client error request", func(t *testing.T) {
		testSlogLoggerMiddleware(t, "POST", "/error", map[string]string{}, 400, echo.NewHTTPError(400, "Bad Request"),
			slog.LevelInfo, []string{"path", "status_code", "method", "err"})
	})

	t.Run("server error request", func(t *testing.T) {
		testSlogLoggerMiddleware(t, "POST", "/server-error", map[string]string{}, 500,
			echo.NewHTTPError(500, "Internal Server Error"),
			slog.LevelError, []string{"path", "status_code", "method", "err"})
	})

	t.Run("request with headers", func(t *testing.T) {
		headers := map[string]string{
			echoMw.RequestIDHeader:   "req-123",
			echoMw.TraceParentHeader: "00-12345678901234567890123456789012-1234567890123456-01",
		}
		testSlogLoggerMiddleware(t, "GET", "/with-headers", headers, 200, nil, slog.LevelInfo,
			[]string{"path", "status_code", "method", "request_id", "trace_id"})
	})
}

func testSlogLoggerMiddleware(t *testing.T, method, path string, headers map[string]string,
	handlerStatus int, handlerError error, expectedLevel slog.Level, expectedFields []string) {
	mockLogger := &mockLogger{}

	e := echo.New()
	e.Use(echoMw.SlogLoggerMiddleware(mockLogger))

	// Set up handler that returns the expected status and error
	e.Any("/*", func(c echo.Context) error {
		if handlerError != nil {
			return handlerError
		}
		return c.String(handlerStatus, "response")
	})

	req := httptest.NewRequest(method, path, nil)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	// Verify logging occurred
	logs := mockLogger.getLogs()
	require.Len(t, logs, 1)

	log := logs[0]
	assert.Equal(t, expectedLevel, log.level)

	// Check that expected fields are present in the log attributes
	require.Len(t, log.attrs, 1) // Should have one group attribute named "context"

	contextGroup := log.attrs[0]
	assert.Equal(t, "context", contextGroup.Key)

	// Extract the group value and check for expected fields
	groupValue := contextGroup.Value
	groupAttrs := groupValue.Group()

	fieldMap := make(map[string]bool)
	for _, attr := range groupAttrs {
		fieldMap[attr.Key] = true
	}

	for _, expectedField := range expectedFields {
		assert.True(t, fieldMap[expectedField], "Expected field %s not found", expectedField)
	}

	// Verify specific header values if provided
	verifyHeadersInLogs(t, headers, groupAttrs)
}

func verifyHeadersInLogs(t *testing.T, headers map[string]string, groupAttrs []slog.Attr) {
	if reqID, ok := headers[echoMw.RequestIDHeader]; ok {
		found := false
		for _, attr := range groupAttrs {
			if attr.Key == "request_id" && attr.Value.String() == reqID {
				found = true
				break
			}
		}
		assert.True(t, found, "Expected request_id not found in logs")
	}

	if traceParent, ok := headers[echoMw.TraceParentHeader]; ok {
		expectedTraceID := strings.Split(traceParent, "-")[1]
		found := false
		for _, attr := range groupAttrs {
			if attr.Key == "trace_id" && attr.Value.String() == expectedTraceID {
				found = true
				break
			}
		}
		assert.True(t, found, "Expected trace_id not found in logs")
	}
}

func TestPutRequestIDContext(t *testing.T) {
	tests := []struct {
		name          string
		headers       map[string]string
		expectedReqID string
		expectedTrace string
	}{
		{
			name:          "no headers",
			headers:       map[string]string{},
			expectedReqID: "",
			expectedTrace: "0",
		},
		{
			name: "with request ID",
			headers: map[string]string{
				echoMw.RequestIDHeader: "test-request-id",
			},
			expectedReqID: "test-request-id",
			expectedTrace: "0",
		},
		{
			name: "with trace parent",
			headers: map[string]string{
				echoMw.TraceParentHeader: "00-12345678901234567890123456789012-1234567890123456-01",
			},
			expectedReqID: "",
			expectedTrace: "12345678901234567890123456789012",
		},
		{
			name: "with both headers",
			headers: map[string]string{
				echoMw.RequestIDHeader:   "test-request-id",
				echoMw.TraceParentHeader: "00-abcdef1234567890abcdef1234567890-abcdef1234567890-01",
			},
			expectedReqID: "test-request-id",
			expectedTrace: "abcdef1234567890abcdef1234567890",
		},
		{
			name: "invalid trace parent format",
			headers: map[string]string{
				echoMw.TraceParentHeader: "invalid-format",
			},
			expectedReqID: "",
			expectedTrace: "0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var capturedReqID interface{}
			var capturedTraceID interface{}

			middleware := echoMw.PutRequestIDContext(func(c echo.Context) error {
				ctx := c.Request().Context()
				capturedReqID = ctx.Value(contextkeys.RequestIDCtxKey)
				capturedTraceID = ctx.Value(contextkeys.TraceIDCtxKey)
				return c.String(200, "OK")
			})

			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/test", nil)

			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}

			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			err := middleware(c)
			require.NoError(t, err)

			// Check context values
			assert.Equal(t, tt.expectedReqID, capturedReqID)
			assert.Equal(t, tt.expectedTrace, capturedTraceID)
		})
	}
}

func TestGetRequestID(t *testing.T) {
	tests := []struct {
		name     string
		headers  map[string][]string
		expected string
	}{
		{
			name:     "no headers",
			headers:  map[string][]string{},
			expected: "",
		},
		{
			name: "with request ID",
			headers: map[string][]string{
				echoMw.RequestIDHeader: {"test-id"},
			},
			expected: "test-id",
		},
		{
			name: "with multiple request IDs",
			headers: map[string][]string{
				echoMw.RequestIDHeader: {"first-id", "second-id"},
			},
			expected: "first-id", // Should take the first one
		},
		{
			name: "case insensitive header",
			headers: map[string][]string{
				"X-Request-Id": {"case-test-id"}, // Canonical form
			},
			expected: "case-test-id",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := echoMw.GetRequestID(tt.headers)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetTraceID(t *testing.T) {
	tests := []struct {
		name     string
		headers  map[string][]string
		expected string
	}{
		{
			name:     "no headers",
			headers:  map[string][]string{},
			expected: "0",
		},
		{
			name: "valid trace parent",
			headers: map[string][]string{
				echoMw.TraceParentHeader: {"00-12345678901234567890123456789012-1234567890123456-01"},
			},
			expected: "12345678901234567890123456789012",
		},
		{
			name: "invalid trace parent format",
			headers: map[string][]string{
				echoMw.TraceParentHeader: {"invalid-format"},
			},
			expected: "0",
		},
		{
			name: "trace parent with wrong separator count",
			headers: map[string][]string{
				echoMw.TraceParentHeader: {"00-12345-01"},
			},
			expected: "0",
		},
		{
			name: "empty trace parent",
			headers: map[string][]string{
				echoMw.TraceParentHeader: {""},
			},
			expected: "0",
		},
		{
			name: "multiple trace parents",
			headers: map[string][]string{
				echoMw.TraceParentHeader: {
					"00-abcdef1234567890abcdef1234567890-abcdef1234567890-01",
					"00-second-trace-01",
				},
			},
			expected: "abcdef1234567890abcdef1234567890", // Should take the first one
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := echoMw.GetTraceID(tt.headers)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMiddlewareIntegration(t *testing.T) {
	// Test that both middlewares work together correctly
	mockLogger := &mockLogger{}

	e := echo.New()
	e.Use(echoMw.PutRequestIDContext)
	e.Use(echoMw.SlogLoggerMiddleware(mockLogger))

	// Handler that accesses context values
	e.GET("/test", func(c echo.Context) error {
		ctx := c.Request().Context()
		reqID := ctx.Value(contextkeys.RequestIDCtxKey)
		traceID := ctx.Value(contextkeys.TraceIDCtxKey)

		response := map[string]interface{}{
			"request_id": reqID,
			"trace_id":   traceID,
		}

		return c.JSON(200, response)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set(echoMw.RequestIDHeader, "integration-test-id")
	req.Header.Set(echoMw.TraceParentHeader, "00-integration1234567890123456789012-integration123456-01")

	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, 200, rec.Code)

	// Verify response contains context values
	var response map[string]interface{}
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "integration-test-id", response["request_id"])
	assert.Equal(t, "integration1234567890123456789012", response["trace_id"])

	// Verify logging occurred with correct values
	logs := mockLogger.getLogs()
	require.Len(t, logs, 1)

	log := logs[0]
	assert.Equal(t, slog.LevelInfo, log.level)
	assert.Equal(t, "REQUEST", log.message)
}

func TestMiddlewareErrorHandling(t *testing.T) {
	mockLogger := &mockLogger{}

	e := echo.New()
	e.Use(echoMw.PutRequestIDContext)
	e.Use(echoMw.SlogLoggerMiddleware(mockLogger))

	// Handler that panics
	e.GET("/panic", func(_ echo.Context) error {
		panic("test panic")
	})

	// Handler that returns HTTP error
	e.GET("/error", func(_ echo.Context) error {
		return echo.NewHTTPError(400, "test error")
	})

	// Test panic handling (should be caught by Echo's recover middleware if present)
	req := httptest.NewRequest(http.MethodGet, "/error", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, 400, rec.Code)

	// Verify error was logged
	logs := mockLogger.getLogs()
	require.Len(t, logs, 1)

	log := logs[0]
	assert.Equal(t, slog.LevelInfo, log.level) // 400 errors are logged as INFO, not ERROR

	// Check that error is included in log attributes
	contextGroup := log.attrs[0]
	groupAttrs := contextGroup.Value.Group()

	errorFound := false
	for _, attr := range groupAttrs {
		if attr.Key == "err" {
			errorFound = true
			assert.Contains(t, attr.Value.String(), "test error")
			break
		}
	}
	assert.True(t, errorFound, "Error not found in log attributes")
}
