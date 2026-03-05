package contextkeys

type CtxKey string

const RequestIDCtxKey CtxKey = "request_id"
const TraceIDCtxKey CtxKey = "trace_id"
const AuthClaimsCtxKey CtxKey = "auth_claims"
