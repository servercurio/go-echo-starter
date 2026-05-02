package v1

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/servercurio/go-echo-starter/internal/router"
	asrt "github.com/stretchr/testify/assert"
)

// serve builds a v1 router config with the supplied readiness state, attaches
// the supplied route to a fresh Echo instance, and serves a GET against the
// supplied path. It returns the recorder so each test case can assert on
// status and body. Using the real route/endpoint constructors (rather than
// calling handler closures directly) exercises the routing path and option
// wiring, which is the bit most likely to regress when the std/* helpers
// change.
func serve(t *testing.T, r router.Route, path string) *httptest.ResponseRecorder {
	t.Helper()
	e := echo.New()
	r.AttachGroup(e.Group(""))

	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
	return rec
}

func readinessCfg(ready bool) *router.Config {
	return &router.Config{ReadinessProbe: func() bool { return ready }}
}

// ---------- /livez ----------

func TestLivenessRoute_AlwaysReturns200OK(t *testing.T) {
	assert := asrt.New(t)
	rec := serve(t, LivenessRoute(), "/livez")

	assert.Equal(http.StatusOK, rec.Code)
	assert.JSONEq(`{"status":"ok"}`, rec.Body.String())
}

// ---------- /readyz ----------

func TestReadinessRoute_ReadyReturns200OK(t *testing.T) {
	assert := asrt.New(t)
	rec := serve(t, ReadinessRoute(readinessCfg(true)), "/readyz")

	assert.Equal(http.StatusOK, rec.Code)
	assert.JSONEq(`{"status":"ok"}`, rec.Body.String())
}

func TestReadinessRoute_NotReadyReturns503(t *testing.T) {
	assert := asrt.New(t)
	rec := serve(t, ReadinessRoute(readinessCfg(false)), "/readyz")

	assert.Equal(http.StatusServiceUnavailable, rec.Code)
	assert.JSONEq(`{"status":"not_ready"}`, rec.Body.String())
}

// TestReadinessRoute_NilProbeFailsClosed pins the fail-closed contract for
// the shared readinessEndpoint helper: a misconfigured router config (nil
// probe) must not panic and must report not-ready, so a future refactor
// can't silently turn a missing probe into a misleading 200 OK.
func TestReadinessRoute_NilProbeFailsClosed(t *testing.T) {
	assert := asrt.New(t)

	e := echo.New()
	ReadinessRoute(&router.Config{ReadinessProbe: nil}).AttachGroup(e.Group(""))

	rec := httptest.NewRecorder()
	assert.NotPanics(func() {
		e.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/readyz", nil))
	})

	assert.Equal(http.StatusServiceUnavailable, rec.Code)
	assert.JSONEq(`{"status":"not_ready"}`, rec.Body.String())
}

// ---------- /healthz (legacy alias for readiness) ----------

func TestHealthRoute_ReadyReturns200OK(t *testing.T) {
	assert := asrt.New(t)
	rec := serve(t, HealthRoute(readinessCfg(true)), "/healthz")

	assert.Equal(http.StatusOK, rec.Code)
	assert.JSONEq(`{"status":"ok"}`, rec.Body.String())
}

func TestHealthRoute_NotReadyReturns503(t *testing.T) {
	assert := asrt.New(t)
	rec := serve(t, HealthRoute(readinessCfg(false)), "/healthz")

	assert.Equal(http.StatusServiceUnavailable, rec.Code)
	assert.JSONEq(`{"status":"not_ready"}`, rec.Body.String())
}
