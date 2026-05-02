// Package health implements an in-process health-check registry and the
// response model returned by the /api/v1/livez, /readyz, and /healthz
// endpoints.
//
// The model intentionally mirrors what Spring Boot Actuator, Quarkus
// SmallRye Health, and Micronaut return: an overall status plus a map of
// per-component statuses with optional details. Components register
// themselves with a Registry; the v1 handlers ask the registry for a
// Snapshot per request.
//
// The Registry is an explicit dependency (passed via router.Config) rather
// than a package-level singleton, in keeping with the no-global-state
// convention documented in CLAUDE.md.
package health

import "sync"

// Status is the overall or component-level health state.
type Status string

const (
	// StatusUp means the component (or aggregated report) is healthy.
	StatusUp Status = "UP"

	// StatusDown means the component (or aggregated report) is unhealthy and
	// callers should treat it as not-ready.
	StatusDown Status = "DOWN"
)

// ComponentResult is the per-component status returned by a CheckFunc.
// Details is optional — leave it nil when there's nothing useful to surface.
type ComponentResult struct {
	Status  Status         `json:"status" yaml:"status"`
	Details map[string]any `json:"details,omitempty" yaml:"details,omitempty"`
}

// Report is the aggregated response body returned by /readyz and /healthz.
// Status is the conjunction of every component's status: UP iff all UP.
type Report struct {
	Status     Status                     `json:"status" yaml:"status"`
	Components map[string]ComponentResult `json:"components" yaml:"components"`
}

// CheckFunc is the contract a component implements to participate in health
// reports. Implementations should be cheap (the function may be called on
// every readyz request) and must not block indefinitely.
type CheckFunc func() ComponentResult

// Registry is a thread-safe collection of named CheckFuncs.
type Registry struct {
	mu     sync.RWMutex
	checks map[string]CheckFunc
}

// NewRegistry returns an empty Registry. Lifecycle code should construct one
// per Application instance and pass it to consumers (router.Config, the
// v1 handlers, etc.).
func NewRegistry() *Registry {
	return &Registry{checks: map[string]CheckFunc{}}
}

// Register associates a check with a component name. A subsequent Register
// call with the same name replaces the previous check.
func (r *Registry) Register(name string, check CheckFunc) {
	if r == nil || check == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.checks[name] = check
}

// Unregister removes a previously-registered check. No-op if absent.
func (r *Registry) Unregister(name string) {
	if r == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.checks, name)
}

// Snapshot runs every registered check and aggregates the results into a
// Report. Overall Status is UP iff every component reports UP. An empty
// registry returns Status=UP with an empty components map (a server with no
// declared dependencies is, by definition, ready).
func (r *Registry) Snapshot() Report {
	if r == nil {
		return Report{Status: StatusUp, Components: map[string]ComponentResult{}}
	}
	r.mu.RLock()
	defer r.mu.RUnlock()

	components := make(map[string]ComponentResult, len(r.checks))
	overall := StatusUp
	for name, check := range r.checks {
		result := check()
		components[name] = result
		if result.Status != StatusUp {
			overall = StatusDown
		}
	}
	return Report{Status: overall, Components: components}
}
