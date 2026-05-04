package application

import (
	"testing"
)

func TestValidateProxyFlags_MultipleEnabled(t *testing.T) {
	t.Parallel()

	cfg := DefaultConfig()
	cfg.Proxy.UseDirectIP = true
	cfg.Proxy.UseXFFHeader = true

	app := &Application{config: cfg}
	if err := app.validateProxyFlags(); err == nil {
		t.Fatalf("expected error when multiple proxy modes are enabled")
	}
}

func TestValidateProxyFlags_OneEnabled(t *testing.T) {
	t.Parallel()

	cfg := DefaultConfig()
	cfg.Proxy.UseDirectIP = false
	cfg.Proxy.UseXFFHeader = true
	cfg.Proxy.UseXRealIPHeader = false

	app := &Application{config: cfg}
	if err := app.validateProxyFlags(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestResolveProxyTrustOptions_MixedCIDR(t *testing.T) {
	t.Parallel()

	cfg := DefaultConfig()
	cfg.Proxy.TrustedIPRanges = []string{
		"10.0.0.0/8",   // valid
		"not-a-cidr",   // invalid — should be dropped, not crash
		"  ",           // empty after trim — should be skipped
		"192.168.0.0/16", // valid
	}

	app := &Application{config: cfg}
	opts := app.resolveProxyTrustOptions()

	// We can't easily inspect the TrustOption values, but we can assert the
	// total count: 2 valid CIDRs + 3 always-appended (private/loopback/link-local).
	if len(opts) != 5 {
		t.Fatalf("expected 5 trust options (2 CIDR + 3 builtin), got %d", len(opts))
	}
}

func TestResolveProxyTrustOptions_NilProxy(t *testing.T) {
	t.Parallel()

	app := &Application{config: &Config{}}
	opts := app.resolveProxyTrustOptions()
	if len(opts) != 0 {
		t.Fatalf("expected 0 options when Proxy is nil, got %d", len(opts))
	}
}
