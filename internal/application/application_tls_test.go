package application

import (
	"crypto/x509"
	"encoding/pem"
	"math/big"
	"testing"
)

// newTestApp builds a minimal Application with the given TLS hostname so
// generateTlsCertificate can run without a full Configure/Initialize.
func newTestApp(hostname string) *Application {
	cfg := DefaultConfig()
	cfg.Server.Https.Hostname = hostname
	cfg.Server.Https.Enabled = true
	return &Application{
		Name:   defaultName,
		config: cfg,
	}
}

func TestGenerateTlsCertificateProducesValidX509(t *testing.T) {
	app := newTestApp("api.example.com")
	imc, err := app.generateTlsCertificate()
	if err != nil {
		t.Fatalf("generateTlsCertificate failed: %v", err)
	}

	if len(imc.Certificate) == 0 {
		t.Fatalf("expected non-empty certificate PEM")
	}
	if len(imc.PrivateKey) == 0 {
		t.Fatalf("expected non-empty private key PEM")
	}

	block, _ := pem.Decode(imc.Certificate)
	if block == nil || block.Type != "CERTIFICATE" {
		t.Fatalf("certificate PEM did not decode")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		t.Fatalf("certificate did not parse: %v", err)
	}

	if cert.Subject.CommonName != "api.example.com" {
		t.Fatalf("CN: got %q, want %q", cert.Subject.CommonName, "api.example.com")
	}

	foundDNS := false
	for _, name := range cert.DNSNames {
		if name == "www.api.example.com" {
			foundDNS = true
			break
		}
	}
	if !foundDNS {
		t.Fatalf("expected www. SAN; got DNSNames=%v", cert.DNSNames)
	}
}

// TestGenerateTlsCertificateSerialIsRandom verifies the PR 2 fix: each call
// produces a unique 128-bit serial, not the constant `1` we had before.
func TestGenerateTlsCertificateSerialIsRandom(t *testing.T) {
	app := newTestApp("localhost")

	a, err := app.generateTlsCertificate()
	if err != nil {
		t.Fatalf("first cert: %v", err)
	}
	b, err := app.generateTlsCertificate()
	if err != nil {
		t.Fatalf("second cert: %v", err)
	}

	parse := func(p []byte) *big.Int {
		block, _ := pem.Decode(p)
		cert, _ := x509.ParseCertificate(block.Bytes)
		return cert.SerialNumber
	}

	sa, sb := parse(a.Certificate), parse(b.Certificate)
	if sa.Cmp(sb) == 0 {
		t.Fatalf("expected different serials across calls; got %s twice", sa.String())
	}
	// Sanity: should be non-trivial (not 1 like the old code)
	if sa.Cmp(big.NewInt(1<<32)) <= 0 {
		t.Fatalf("serial looks too small to be 128-bit random: %s", sa.String())
	}
}
