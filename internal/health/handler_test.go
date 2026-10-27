package health_test

import (
	"context"
	"encoding/json"
	"errors"
	"go-web-template/internal/apperr"
	"go-web-template/internal/health"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type stub struct {
	name string
	err  error
}

func newStub(name string, err error) *stub {
	return &stub{name: name, err: err}
}

func (s *stub) Name() string                  { return s.name }
func (s *stub) Check(_ context.Context) error { return s.err }

func waitReported(t *testing.T, svc *health.Service, names ...string) {
	t.Helper()
	require.Eventually(t, func() bool {
		checks := svc.Report().Checks
		for _, name := range names {
			if _, ok := checks[name]; !ok {
				return false
			}
		}
		return true
	}, time.Second, 5*time.Millisecond, "checks %v not reported within 1s", names)
}

func serve(svc *health.Service) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	handler := health.NewHealthHandler(svc, apperr.NewResponder(zap.NewNop()))
	handler.Report(rec, httptest.NewRequest(http.MethodGet, "/health", nil))
	return rec
}

func newSvc(t *testing.T) *health.Service {
	t.Helper()
	svc, err := health.New(zap.NewNop())
	require.NoError(t, err)
	return svc
}

func startSvc(t *testing.T, svc *health.Service) {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	go func() { _ = svc.Run(ctx) }()
}

func TestHandler_AllHealthy(t *testing.T) {
	svc := newSvc(t)
	svc.Add(newStub("db", nil))
	startSvc(t, svc)
	waitReported(t, svc, "db")

	rec := serve(svc)

	assert.Equal(t, http.StatusOK, rec.Code)

	var body map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	assert.Equal(t, "ok", body["status"])
	dbCheck := body["checks"].(map[string]any)["db"].(map[string]any)
	assert.Equal(t, "ok", dbCheck["status"])
}

func TestHandler_Degraded(t *testing.T) {
	svc := newSvc(t)
	svc.Add(newStub("db", errors.New("connection refused")))
	startSvc(t, svc)
	waitReported(t, svc, "db")

	rec := serve(svc)

	assert.Equal(t, http.StatusServiceUnavailable, rec.Code)
	assert.NotContains(t, rec.Body.String(), "connection refused")

	var body map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	assert.Equal(t, "degraded", body["status"])
	dbCheck := body["checks"].(map[string]any)["db"].(map[string]any)
	assert.Equal(t, "error", dbCheck["status"])
}

func TestHandler_PartialFailure(t *testing.T) {
	svc := newSvc(t)
	svc.Add(newStub("db", nil))
	svc.Add(newStub("redis", errors.New("connection refused")))
	startSvc(t, svc)
	waitReported(t, svc, "db", "redis")

	rec := serve(svc)

	assert.Equal(t, http.StatusServiceUnavailable, rec.Code)

	var body map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	assert.Equal(t, "degraded", body["status"])
	checks := body["checks"].(map[string]any)
	assert.Equal(t, "ok", checks["db"].(map[string]any)["status"])
	assert.Equal(t, "error", checks["redis"].(map[string]any)["status"])
}
