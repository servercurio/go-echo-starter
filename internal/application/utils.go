package application

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/joomcode/errorx"
	"github.com/labstack/echo/v5"
	"github.com/servercurio/go-echo-starter/internal/database"
	"github.com/servercurio/go-echo-starter/internal/logging"
)

var defaultInsecurePaths = []string{"api/v1/healthz", "api/v1/readyz"}

func configSearchPaths() []string {
	search := []string{fmt.Sprintf("/etc/%s", defaultConfigPathElement)}

	if home, err := os.UserHomeDir(); err == nil {
		search = append(search, absPath(filepath.Join(home, ".config", defaultConfigPathElement)))
	}

	search = append(search, absPath("."))
	return search
}

func absPath(path string) string {
	if p, err := filepath.Abs(path); err == nil {
		return p
	}

	return path
}

func NotifyHttpServerConfig(cfg *HttpConfig) {
	logging.Daemon.Info().
		EmbedObject(cfg).
		Msg("http server configuration")
}

func NotifyHttpsServerConfig(cfg *TlsConfig) {
	if cfg == nil || !cfg.Enabled {
		return
	}

	logging.Daemon.Info().
		EmbedObject(cfg).
		Msg("https server configuration")
}

func NotifyProxySupportConfig(cfg *ProxyConfig) {
	if cfg == nil {
		return
	}

	logging.Daemon.Info().
		EmbedObject(cfg).
		Msg("proxy support configuration")
}

func NotifyDatabaseConfig(cfg *database.Config) {
	if cfg == nil {
		return
	}

	logging.Daemon.Info().
		EmbedObject(cfg).
		Msg("database configuration")
}

func NotifyOpenAPIConfig(cfg *OpenAPIConfig) {
	if cfg == nil {
		return
	}

	logging.Daemon.Info().
		EmbedObject(cfg).
		Msg("openapi configuration")
}

// parseByteSize parses a human-readable byte-size string (e.g. "1MB", "500KB",
// "2GB", "1024") into a byte count. The suffix is case-insensitive; bare
// numbers are treated as bytes. Returns an error for unrecognised suffixes or
// non-numeric magnitudes.
func parseByteSize(s string) (int64, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, errorx.IllegalArgument.New("empty byte size")
	}

	upper := strings.ToUpper(s)
	mult := int64(1)
	switch {
	case strings.HasSuffix(upper, "GB"):
		mult = 1024 * 1024 * 1024
		upper = strings.TrimSuffix(upper, "GB")
	case strings.HasSuffix(upper, "MB"):
		mult = 1024 * 1024
		upper = strings.TrimSuffix(upper, "MB")
	case strings.HasSuffix(upper, "KB"):
		mult = 1024
		upper = strings.TrimSuffix(upper, "KB")
	case strings.HasSuffix(upper, "B"):
		upper = strings.TrimSuffix(upper, "B")
	}

	n, err := strconv.ParseInt(strings.TrimSpace(upper), 10, 64)
	if err != nil {
		return 0, errorx.IllegalArgument.Wrap(err, "invalid byte size %q", s)
	}
	if n < 0 {
		return 0, errorx.IllegalArgument.New("byte size must be non-negative: %q", s)
	}
	return n * mult, nil
}

func HTTPSRedirectWithConfig(cfg *TlsConfig) echo.MiddlewareFunc {
	portSpec := ""
	if cfg.Port != 443 {
		portSpec = fmt.Sprintf(":%d", cfg.Port)
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			if c.Request().TLS != nil {
				return next(c)
			}

			for _, path := range defaultInsecurePaths {
				if strings.Contains(c.Request().URL.Path, path) {
					return next(c)
				}
			}

			hostNoPort := c.Request().Host
			if strings.Contains(hostNoPort, ":") {
				hostNoPort = strings.Split(hostNoPort, ":")[0]
			}

			redirectUrl := fmt.Sprintf("https://%s%s%s",
				hostNoPort, portSpec, c.Request().URL.Path)

			if err := c.Redirect(http.StatusPermanentRedirect, redirectUrl); err != nil {
				return err
			}

			return nil
		}
	}
}
