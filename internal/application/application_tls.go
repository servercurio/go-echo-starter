package application

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/joomcode/errorx"
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
	app.tlsServer.HideBanner = true
	app.tlsServer.Logger.SetOutput(io.Discard)
	app.tlsServer.StdLogger.SetOutput(io.Discard)

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
		SerialNumber: big.NewInt(1),
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

	if err := os.MkdirAll(cacheDir, 0755); err != nil {
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

	hp := autocert.HostWhitelist("localhost")
	if app.config.Server.Https.Hostname != "" {
		hp = autocert.HostWhitelist(app.config.Server.Https.Hostname)
	}

	app.tlsServer.AutoTLSManager.HostPolicy = hp
	app.tlsServer.AutoTLSManager.Cache = autocert.DirCache(cacheDir)
	app.tlsServer.AutoTLSManager.Prompt = autocert.AcceptTOS

	app.tlsServer.Logger.SetOutput(os.Stderr)
	app.tlsServer.StdLogger.SetOutput(os.Stderr)

	logging.Daemon.Info().
		Str("certificateCache", cacheDir).
		Msg("auto tls - automatic tls support configured")

	return nil
}

func (app *Application) startTlsServer() {
	if !app.config.Server.Https.Enabled {
		return
	}

	address := fmt.Sprintf("%s:%d", app.config.Server.Https.BindAddress, app.config.Server.Https.Port)

	logging.Daemon.Info().
		Str("address", address).
		Bool("autoCertIssuance", app.autoTlsEnabled()).
		Bool("ephemeralCertIssuance", !app.config.Server.Https.UseAcmeIssuer).
		Msg("https server started")

	if app.autoTlsEnabled() {
		if app.config.Server.Https.UseAcmeIssuer {
			if err := app.tlsServer.StartAutoTLS(address); !errors.Is(err, http.ErrServerClosed) {
				logging.Daemon.Error().
					Err(errorx.EnsureStackTrace(err)).
					Msg("https server shutting down due to an error")
			}
			return
		}

		if err := app.tlsServer.StartTLS(address, app.certificate.Certificate, app.certificate.PrivateKey); !errors.Is(err, http.ErrServerClosed) {
			logging.Daemon.Error().
				Err(errorx.EnsureStackTrace(err)).
				Msg("https server shutting down due to an error")
		}
		return
	}

	if err := app.tlsServer.StartTLS(address, app.config.Server.Https.Certificate, app.config.Server.Https.Key); !errors.Is(err, http.ErrServerClosed) {
		logging.Daemon.Error().
			Err(errorx.EnsureStackTrace(err)).
			Msg("https server shutting down due to an error")
	}
}

func (app *Application) shutdownTlsServer() {
	if !app.config.Server.Https.Enabled {
		return
	}

	var ctx context.Context
	var cancel context.CancelFunc

	if app.config.Server.Https.ShutdownTimeout > 0 {
		ctx, cancel = context.WithTimeout(context.Background(), app.config.Server.Https.ShutdownTimeout)
	} else {
		ctx, cancel = context.WithCancel(context.Background())
	}

	defer cancel()
	if err := app.tlsServer.Shutdown(ctx); err != nil {
		logging.Daemon.Error().
			Err(err).
			Msg("https server shutdown failed")
	}
}
