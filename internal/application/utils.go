package application

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/servercurio/go-echo-starter/internal/logging"
)

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

func HTTPSRedirectWithConfig(cfg *TlsConfig) echo.MiddlewareFunc {
	portSpec := ""
	if cfg.Port != 443 {
		portSpec = fmt.Sprintf(":%d", cfg.Port)
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if c.Request().TLS != nil {
				return next(c)
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
