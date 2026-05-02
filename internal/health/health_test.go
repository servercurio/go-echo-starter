package health

import (
	"testing"

	asrt "github.com/stretchr/testify/assert"
)

func TestRegistry_EmptyRegistryIsUp(t *testing.T) {
	// "No declared dependencies" is, by definition, healthy.
	assert := asrt.New(t)
	r := NewRegistry()

	rep := r.Snapshot()
	assert.Equal(StatusUp, rep.Status)
	assert.Empty(rep.Components)
}

func TestRegistry_AllUpAggregatesUp(t *testing.T) {
	assert := asrt.New(t)
	r := NewRegistry()
	r.Register("a", func() ComponentResult { return ComponentResult{Status: StatusUp} })
	r.Register("b", func() ComponentResult { return ComponentResult{Status: StatusUp} })

	rep := r.Snapshot()
	assert.Equal(StatusUp, rep.Status)
	assert.Len(rep.Components, 2)
}

func TestRegistry_AnyDownAggregatesDown(t *testing.T) {
	// Pin the conjunction semantics: a single DOWN component flips the
	// overall report. Mirrors how Spring Boot Actuator and Quarkus health
	// aggregate component status.
	assert := asrt.New(t)
	r := NewRegistry()
	r.Register("a", func() ComponentResult { return ComponentResult{Status: StatusUp} })
	r.Register("b", func() ComponentResult { return ComponentResult{Status: StatusDown} })

	rep := r.Snapshot()
	assert.Equal(StatusDown, rep.Status)
	assert.Equal(StatusUp, rep.Components["a"].Status)
	assert.Equal(StatusDown, rep.Components["b"].Status)
}

func TestRegistry_RegisterReplacesExisting(t *testing.T) {
	assert := asrt.New(t)
	r := NewRegistry()
	r.Register("x", func() ComponentResult { return ComponentResult{Status: StatusUp} })
	r.Register("x", func() ComponentResult { return ComponentResult{Status: StatusDown} })

	rep := r.Snapshot()
	assert.Equal(StatusDown, rep.Status)
}

func TestRegistry_UnregisterRemovesCheck(t *testing.T) {
	assert := asrt.New(t)
	r := NewRegistry()
	r.Register("flapper", func() ComponentResult { return ComponentResult{Status: StatusDown} })
	r.Unregister("flapper")

	rep := r.Snapshot()
	assert.Equal(StatusUp, rep.Status, "removing the only DOWN component should restore overall UP")
	assert.NotContains(rep.Components, "flapper")
}

func TestRegistry_NilReceiverIsSafe(t *testing.T) {
	// Defensive: the v1 handler holds the registry by pointer; if a future
	// caller hands us a nil pointer we'd rather report "ready, no
	// components" than panic. Pin that here.
	assert := asrt.New(t)
	var r *Registry

	rep := r.Snapshot()
	assert.Equal(StatusUp, rep.Status)
	assert.Empty(rep.Components)

	// Register/Unregister on nil receiver are no-ops, not panics.
	assert.NotPanics(func() {
		r.Register("never", func() ComponentResult { return ComponentResult{Status: StatusDown} })
		r.Unregister("never")
	})
}

func TestFormatFromAccept(t *testing.T) {
	assert := asrt.New(t)

	cases := []struct {
		accept string
		want   Format
	}{
		{"", FormatJSON},
		{"*/*", FormatJSON},
		{"application/json", FormatJSON},
		{"application/json, text/plain", FormatJSON},
		{"application/yaml", FormatYAML},
		{"application/x-yaml", FormatYAML},
		{"text/yaml", FormatYAML},
		{"application/health+yaml", FormatYAML},
		{"APPLICATION/YAML", FormatYAML}, // case-insensitive
	}

	for _, c := range cases {
		assert.Equal(c.want, FormatFromAccept(c.accept), "Accept=%q", c.accept)
	}
}
