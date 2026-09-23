package contextx

type CtxKey string

const (
	CtxKeyTracer         CtxKey = "tracer"
	CtxKeyLogger         CtxKey = "logger"
	CtxKeyIdempotencyKey CtxKey = "idempotency_key"
	CtxKeyCorrelationID  CtxKey = "correlation_id"
)
