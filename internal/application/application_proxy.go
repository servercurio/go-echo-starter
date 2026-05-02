package application

import (
	"net"
	"strings"

	"github.com/joomcode/errorx"
	"github.com/labstack/echo/v5"
	"github.com/servercurio/go-echo-starter/internal/logging"
)

func (app *Application) configureProxySupport() error {
	pc := app.config.Proxy
	if pc == nil {
		return nil
	}

	if err := app.validateProxyFlags(); err != nil {
		return err
	}

	var extractor echo.IPExtractor
	if pc.UseXRealIPHeader {
		extractor = echo.ExtractIPFromRealIPHeader(app.resolveProxyTrustOptions()...)
	}

	if pc.UseXFFHeader {
		extractor = echo.ExtractIPFromXFFHeader(app.resolveProxyTrustOptions()...)
	}

	if extractor == nil {
		extractor = echo.ExtractIPDirect()
	}

	app.httpServer.IPExtractor = extractor
	app.tlsServer.IPExtractor = extractor
	return nil
}

func (app *Application) validateProxyFlags() error {
	pc := app.config.Proxy
	if pc == nil {
		return nil
	}

	flagCnt := 0
	if pc.UseDirectIP {
		flagCnt++
	}

	if pc.UseXRealIPHeader {
		flagCnt++
	}

	if pc.UseXFFHeader {
		flagCnt++
	}

	if flagCnt > 1 {
		return errorx.IllegalState.New("only one of useDirectIP, useXRealIPHeader, or useXFFHeader can be enabled")
	}

	return nil
}

func (app *Application) resolveProxyTrustOptions() []echo.TrustOption {
	pc := app.config.Proxy
	if pc == nil {
		return []echo.TrustOption{}
	}

	var invalidRanges = make([]string, 0)
	var opts = make([]echo.TrustOption, 0)
	for _, trustedRange := range pc.TrustedIPRanges {
		trustedRange = strings.TrimSpace(trustedRange)
		if trustedRange == "" {
			continue
		}

		if _, network, err := net.ParseCIDR(trustedRange); err == nil && network != nil {
			opts = append(opts, echo.TrustIPRange(network))
		} else {
			invalidRanges = append(invalidRanges, trustedRange)
		}
	}

	if len(invalidRanges) > 0 {
		logging.Daemon.
			Warn().
			Strs("cidrAddresses", invalidRanges).
			Msg("proxy config contains invalid CIDR addresses")
	}

	opts = append(opts,
		echo.TrustPrivateNet(true),
		echo.TrustLoopback(true),
		echo.TrustLinkLocal(true))

	return opts
}
