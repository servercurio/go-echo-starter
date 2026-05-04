package application

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"log/slog"
	"math/big"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/joomcode/errorx"
	"github.com/labstack/echo/v5"
	mw "github.com/labstack/echo/v5/middleware"
	ce "github.com/servercurio/go-echo-starter/internal/errors"
	"github.com/servercurio/go-echo-starter/internal/logging"
	"golang.org/x/crypto/acme/autocert"
)

type InMemoryCertificate struct {
	PrivateKey  []byte
	PublicKey   []byte
	Certificate []byte
}

func (app *Application) tlsEnabled() bool {
	return (app.config.Server.Https.Enabled && app.config.Server.Https.Certificate != "" &&
		app.config.Server.Https.Key != "") || app.autoTlsEnabled()
}

func (app *Application) autoTlsEnabled() bool {
	return app.config.Server.Https.Enabled &&
		(app.config.Server.Https.Certificate == "" || app.config.Server.Https.Key == "")
}

func (app *Application) configureTlsServer() error {
	if app.config.Server.Https == nil || !app.config.Server.Https.Enabled {
		return nil
	}

	app.tlsServer.Use(app.middleware...)

	if cors := CorsMiddleware(app.config.Server.Cors); cors != nil {
		app.tlsServer.Use(cors)
	}

	bodyLimit, err := parseByteSize(app.config.Server.Https.MaxBodySize)
	if err != nil {
		return errorx.IllegalArgument.Wrap(err, "invalid https max body size %q", app.config.Server.Https.MaxBodySize)
	}
	app.tlsServer.Use(mw.BodyLimit(bodyLimit))

	app.tlsServer.Logger = slog.New(slog.DiscardHandler)

	autoTlsEnabled := app.autoTlsEnabled()

	if autoTlsEnabled {
		if app.config.Server.Https.UseAcmeIssuer {
			logging.Daemon.Debug().
				Msg("auto tls - configuring automatic tls support with an ACME issuer")

			if err := app.configureAutoTlsManager(); err != nil {
				logging.Daemon.Error().
					Err(err).
					Msg("auto tls - failed to configure automatic tls support")
				return err
			}
		} else {
			logging.Daemon.Debug().
				Str("algorithm", "ECDSA").
				Str("curve", "P384").
				Msg("ephemeral tls - generating self-signed certificate")
			var err error
			if app.certificate, err = app.generateTlsCertificate(); err != nil {
				return err
			} else {
				logging.Daemon.Info().
					Msg("ephemeral tls - self-signed certificate generated")
			}
		}
	}

	return nil
}

func (app *Application) generateTlsCertificate() (imc *InMemoryCertificate, err error) {
	imc = &InMemoryCertificate{}
	var pk *ecdsa.PrivateKey

	if pk, err = ecdsa.GenerateKey(elliptic.P384(), rand.Reader); err != nil {
		wErr := errorx.ExternalError.Wrap(err, "failed to generate ephemeral ECDSA private key")
		logging.Daemon.Error().
			Err(wErr).
			Msg("ephemeral tls - failed to generate ECDSA private key")
		return nil, wErr
	}

	// 128-bit random serial per RFC 5280 §4.1.2.2; avoids reusing the same
	// serial across daemon restarts, which would alias entries in client
	// certificate caches.
	serialMax := new(big.Int).Lsh(big.NewInt(1), 128)
	serial, err := rand.Int(rand.Reader, serialMax)
	if err != nil {
		wErr := errorx.ExternalError.Wrap(err, "failed to generate certificate serial number")
		logging.Daemon.Error().
			Err(wErr).
			Msg("ephemeral tls - failed to generate certificate serial number")
		return nil, wErr
	}

	cn := "localhost"
	if app.config.Server.Https.Hostname != "" {
		cn = app.config.Server.Https.Hostname
	}

	var altNames []string

	if app.config.Server.Https.Hostname != "" {
		if !strings.HasPrefix(app.config.Server.Https.Hostname, "www") {
			altNames = append(altNames, "www."+app.config.Server.Https.Hostname)
		}

		if app.config.Server.Https.Hostname == "localhost" {
			altNames = append(altNames, "127.0.0.1")
		}
	} else {
		altNames = append(altNames, "127.0.0.1")
	}

	template := &x509.Certificate{
		SerialNumber: serial,
		Subject:      pkix.Name{CommonName: cn},
		NotBefore:    time.Now().UTC().Add(-24 * time.Hour),
		NotAfter:     time.Now().UTC().Add(365 * 24 * time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		IsCA:         false,
		DNSNames:     altNames,
	}

	var certDerBytes []byte
	if certDerBytes, err = x509.CreateCertificate(rand.Reader, template, template, pk.Public(), pk); err != nil {
		wErr := errorx.ExternalError.Wrap(err, "failed to create certificate")
		logging.Daemon.Error().
			Err(wErr).
			Msg("ephemeral tls - failed to create certificate")
		return nil, wErr
	}

	imc.Certificate = pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDerBytes})

	var pkDerBytes []byte
	if pkDerBytes, err = x509.MarshalPKCS8PrivateKey(pk); err != nil {
		wErr := errorx.ExternalError.Wrap(err, "failed to marshal private key")
		logging.Daemon.Error().
			Err(wErr).
			Msg("ephemeral tls - failed to marshal private key")
		return nil, wErr
	}

	imc.PrivateKey = pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: pkDerBytes})

	var pubDerBytes []byte
	if pubDerBytes, err = x509.MarshalPKIXPublicKey(pk.Public()); err != nil {
		wErr := errorx.ExternalError.Wrap(err, "failed to marshal public key")
		logging.Daemon.Error().
			Err(wErr).
			Msg("ephemeral tls - failed to marshal public key")
		return nil, wErr
	}

	imc.PublicKey = pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubDerBytes})

	return
}

func (app *Application) configureAutoTlsManager() error {
	cacheDir := strings.TrimSpace(app.config.Server.Https.CertificateCache)

	if cacheDir == "" {
		if app.userHomeDirectory != "" && app.userHomeDirectory != "." {
			cacheDir = filepath.Join(app.userHomeDirectory, ".cache", app.Name, "certificates")
		} else {
			cacheDir = filepath.Join(string(filepath.Separator), "etc", app.Name, "certificates")
		}
	}

	if err := os.MkdirAll(cacheDir, 0700); err != nil {
		var wErr error
		if os.IsPermission(err) {
			wErr = ce.FileAccessDenied.Wrap(err, "permission denied: %s", cacheDir)
		} else if os.IsExist(err) {
			wErr = ce.InvalidFilePath.Wrap(err, "file already exists and is a regular file: %s", cacheDir)
		} else {
			wErr = errorx.ExternalError.Wrap(err, "failed to create directory: %s", cacheDir)
		}

		logging.Daemon.Error().
			Err(wErr).
			Str("path", cacheDir).
			Msg("auto tls - failed to create certificate cache directory")
		return wErr
	}

	// MkdirAll honours the mode only when creating; if cacheDir already
	// existed with looser permissions (e.g. from a pre-fix daemon version),
	// tighten it now. The cache holds private keys and ACME account state.
	if err := os.Chmod(cacheDir, 0700); err != nil {
		logging.Daemon.Warn().
			Err(err).
			Str("path", cacheDir).
			Msg("auto tls - could not tighten certificate cache directory permissions")
	}

	app.tlsServer.Logger = nil

	logging.Daemon.Info().
		Str("certificateCache", cacheDir).
		Msg("auto tls - automatic tls support configured")

	return nil
}

func (app *Application) startTlsServer(ctx context.Context) {
	if !app.config.Server.Https.Enabled {
		return
	}

	cacheDir := strings.TrimSpace(app.config.Server.Https.CertificateCache)
	address := fmt.Sprintf("%s:%d", app.config.Server.Https.BindAddress, app.config.Server.Https.Port)

	logging.Daemon.Info().
		Str("address", address).
		Bool("autoCertIssuance", app.autoTlsEnabled()).
		Bool("ephemeralCertIssuance", !app.config.Server.Https.UseAcmeIssuer).
		Msg("https server started")

	tlsCfg := app.config.Server.Https
	sc := &echo.StartConfig{
		HideBanner:      true,
		Address:         address,
		GracefulTimeout: tlsCfg.ShutdownTimeout,
		// Static-cert and self-signed paths use this TLSConfig; ACME path
		// overrides it below with autocert.Manager.TLSConfig() (which also
		// pins MinVersion to TLS 1.2 internally).
		TLSConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
			NextProtos: []string{"h2", "http/1.1"},
		},
		BeforeServeFunc: func(s *http.Server) error {
			s.ReadTimeout = tlsCfg.ReadTimeout
			s.ReadHeaderTimeout = tlsCfg.ReadHeaderTimeout
			s.WriteTimeout = tlsCfg.WriteTimeout
			s.IdleTimeout = tlsCfg.IdleTimeout
			return nil
		},
	}

	if app.autoTlsEnabled() {
		if app.config.Server.Https.UseAcmeIssuer {
			hp := autocert.HostWhitelist("localhost")
			if app.config.Server.Https.Hostname != "" {
				hp = autocert.HostWhitelist(app.config.Server.Https.Hostname)
			}

			acm := &autocert.Manager{
				Prompt:     autocert.AcceptTOS,
				Cache:      autocert.DirCache(cacheDir),
				HostPolicy: hp,
			}

			// autocert.TLSConfig() returns a config with GetCertificate set
			// and MinVersion: TLS 1.2 already; replacing our explicit config
			// with it is safe.
			sc.TLSConfig = acm.TLSConfig()

			if err := sc.Start(ctx, app.tlsServer); err != nil && !errors.Is(err, http.ErrServerClosed) {
				logging.Daemon.Error().
					Err(errorx.EnsureStackTrace(err)).
					Msg("https server shutting down due to an error")
			}
			return
		}

		if err := sc.StartTLS(ctx, app.tlsServer, app.certificate.Certificate, app.certificate.PrivateKey); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logging.Daemon.Error().
				Err(errorx.EnsureStackTrace(err)).
				Msg("https server shutting down due to an error")
		}
		return
	}

	if err := sc.StartTLS(ctx, app.tlsServer, app.config.Server.Https.Certificate, app.config.Server.Https.Key); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logging.Daemon.Error().
			Err(errorx.EnsureStackTrace(err)).
			Msg("https server shutting down due to an error")
	}
}
