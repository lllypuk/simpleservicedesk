package echomiddleware

import (
	"net/http"
	"strings"
)

const (
	RequestIDHeader   = "X-Request-Id"
	TraceParentHeader = "Traceparent"
)

func GetRequestID(headers map[string][]string) string {
	var reqID string
	if val, ok := headers[http.CanonicalHeaderKey(RequestIDHeader)]; ok && len(val) > 0 {
		reqID = val[0]
	}
	return reqID
}

func GetTraceID(headers map[string][]string) string {
	traceID := "0"
	if val, ok := headers[http.CanonicalHeaderKey(TraceParentHeader)]; ok && len(val) > 0 {
		parentVal := val[0]
		if strings.Count(parentVal, "-") == parentSeparatorNumber {
			traceID = strings.Split(parentVal, "-")[1]
		}
	}
	return traceID
}
