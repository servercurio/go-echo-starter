package v1

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/servercurio/go-echo-starter/internal/router"
	asrt "github.com/stretchr/testify/assert"
)

// healthRequest builds a v1 router config with the supplied readiness state,
// attaches the health route to a fresh Echo instance, and serves a GET against
// /healthz. It returns the recorder so each test case can assert on status and
// body. Using the real route/endpoint constructors (rather than calling the
// handler closure directly) exercises the routing path and option wiring,
// which is the bit most likely to regress when the std/* helpers change.
func healthRequest(ready bool) *httptest.ResponseRecorder {
	e := echo.New()
	HealthRoute(&router.Config{ReadinessProbe: func() bool { return ready }}).
		AttachGroup(e.Group(""))

	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/healthz", nil))
	return rec
}

func TestHealthRoute_ReadyReturns200OK(t *testing.T) {
	assert := asrt.New(t)
	rec := healthRequest(true)

	assert.Equal(http.StatusOK, rec.Code)
	assert.JSONEq(`{"status":"ok"}`, rec.Body.String())
}

func TestHealthRoute_NotReadyReturns503(t *testing.T) {
	assert := asrt.New(t)
	rec := healthRequest(false)

	assert.Equal(http.StatusServiceUnavailable, rec.Code)
	assert.JSONEq(`{"status":"not_ready"}`, rec.Body.String())
}

// TestHealthRoute_NilProbeFailsClosed makes sure a misconfigured router config
// (nil probe) doesn't panic and instead reports not-ready. This codifies the
// fail-closed default in the handler so a future refactor can't silently turn
// a missing probe into a misleading 200 OK.
func TestHealthRoute_NilProbeFailsClosed(t *testing.T) {
	assert := asrt.New(t)

	e := echo.New()
	HealthRoute(&router.Config{ReadinessProbe: nil}).
		AttachGroup(e.Group(""))

	rec := httptest.NewRecorder()
	assert.NotPanics(func() {
		e.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/healthz", nil))
	})

	assert.Equal(http.StatusServiceUnavailable, rec.Code)
	assert.JSONEq(`{"status":"not_ready"}`, rec.Body.String())
}
