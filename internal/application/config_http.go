package application

import (
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/servercurio/go-echo-starter/internal/env"
)

type HttpConfig struct {
	Hostname        string        `yaml:"hostname" json:"hostname"`
	BindAddress     string        `yaml:"bindAddress" json:"bindAddress"`
	Port            uint16        `yaml:"port" json:"port"`
	ShutdownTimeout time.Duration `yaml:"shutdownTimeout" json:"shutdownTimeout"`
}

func (h *HttpConfig) MarshalZerologObject(e *zerolog.Event) {
	ba := h.BindAddress

	if strings.TrimSpace(ba) == "" {
		ba = "0.0.0.0"
	}

	e.Str("hostname", h.Hostname)
	e.Str("bindAddress", ba)
	e.Uint16("port", h.Port)
	e.Str("shutdownTimeout", h.ShutdownTimeout.String())
}

func (h *HttpConfig) FromEnv(prefix string) {
	env.SetStringValue(prefix, "hostname", &h.Hostname)
	env.SetStringValue(prefix, "bind_address", &h.BindAddress)
	env.SetUint16Value(prefix, "port", &h.Port)
	env.SetDurationValue(prefix, "shutdown_timeout", &h.ShutdownTimeout)
}

type TlsConfig struct {
	*HttpConfig

	// Enabled indicates whether TLS is enabled.
	Enabled bool `yaml:"enabled" json:"enabled"`
	// Certificate is the path to the TLS certificate file. If omitted or empty, then Let's Encrypt
	// based AutoTLS certificate issuance will be attempted.
	Certificate string `yaml:"certificate" json:"certificate"`
	// Key is the path to the TLS key file. If omitted or empty, then Let's Encrypt based AutoTLS
	// certificate issuance will be attempted.
	Key string `yaml:"key" json:"key"`
	// CertificateCache is the path to the directory where AutoTLS will store its certificates.
	CertificateCache string `yaml:"certificateCache" json:"certificateCache"`
	// UseAcmeIssuer indicates whether to use the ACME issuer for certificate issuance or generate an
	// ephemeral certificate.
	UseAcmeIssuer bool `yaml:"useAcmeIssuer" json:"useAcmeIssuer"`
}

func (t *TlsConfig) MarshalZerologObject(e *zerolog.Event) {
	autoCertIssuance := false
	if strings.TrimSpace(t.Certificate) == "" || strings.TrimSpace(t.Key) == "" {
		autoCertIssuance = true
	}

	e.EmbedObject(t.HttpConfig).
		Bool("enabled", t.Enabled).
		Bool("autoCertIssuance", autoCertIssuance)

	if !autoCertIssuance {
		e.Str("certificate", t.Certificate).
			Str("key", t.Key)
	} else {
		e.Str("certificateCache", t.CertificateCache).
			Bool("useAcmeIssuer", t.UseAcmeIssuer)
	}
}

func (t *TlsConfig) FromEnv(prefix string) {
	t.HttpConfig.FromEnv(prefix)
	env.SetBoolValue(prefix, "enabled", &t.Enabled)
	env.SetStringValue(prefix, "certificate", &t.Certificate)
	env.SetStringValue(prefix, "key", &t.Key)
	env.SetStringValue(prefix, "certificate_cache", &t.CertificateCache)
	env.SetBoolValue(prefix, "use_acme_issuer", &t.UseAcmeIssuer)
}

func DefaultHttpConfig() *HttpConfig {
	return &HttpConfig{
		BindAddress:     "",
		Port:            8080,
		ShutdownTimeout: 10 * time.Second,
	}
}

func DefaultTlsConfig() *TlsConfig {
	return &TlsConfig{
		HttpConfig: &HttpConfig{
			BindAddress:     "",
			Port:            8443,
			ShutdownTimeout: 10 * time.Second,
		},
		Enabled:          false,
		Certificate:      "",
		Key:              "",
		CertificateCache: "",
		UseAcmeIssuer:    false,
	}
}
