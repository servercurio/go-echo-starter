package application

import (
	"errors"
	"net"
	"strings"

	"github.com/rs/zerolog"
	"github.com/servercurio/go-echo-starter/internal/env"
	"github.com/servercurio/go-echo-starter/internal/logging"
)

type ProxyConfig struct {
	UseDirectIP      bool     `yaml:"useDirectIP" json:"useDirectIP"`
	UseXFFHeader     bool     `yaml:"useXFFHeader" json:"useXFFHeader"`
	UseXRealIPHeader bool     `yaml:"useXRealIPHeader" json:"useXRealIPHeader"`
	TrustedIPRanges  []string `yaml:"trustedIPRanges" json:"trustedIPRanges"`
}

func (c *ProxyConfig) FromEnv(prefix string) {
	env.SetBoolValue(prefix, "use_direct_ip", &c.UseDirectIP)
	env.SetBoolValue(prefix, "use_xff_header", &c.UseXFFHeader)
	env.SetBoolValue(prefix, "use_x_real_ip_header", &c.UseXRealIPHeader)

	var trustedIPRangeString string
	env.SetStringValue(prefix, "trusted_ip_ranges", &trustedIPRangeString)

	trustedIPRangeString = strings.TrimSpace(trustedIPRangeString)
	if trustedIPRangeString == "" {
		c.TrustedIPRanges = []string{}
		return
	}

	parts := strings.Split(trustedIPRangeString, ",")
	parsed := make([]string, 0, len(parts))
	for _, p := range parts {
		r := strings.TrimSpace(p)
		if r == "" {
			continue
		}
		if _, network, err := net.ParseCIDR(r); err == nil && network != nil {
			parsed = append(parsed, r)
			continue
		}
		logging.Daemon.Warn().
			Str("range", r).
			Msg("ignoring invalid CIDR in trusted IP ranges")
	}
	c.TrustedIPRanges = parsed
}

// Validate refuses configurations that mix mutually-exclusive IP-extraction
// strategies. The proxy module reads from one source — direct, XFF, or
// X-Real-IP — and stitching multiple together would cause spoofing
// surprises (an attacker who controls one header could override the IP
// the rate limiter / log line attributes the request to). Pre-fix this
// was caught at Initialize and downgraded to a Warn; that left a daemon
// running with whatever extractor won the race.
func (c *ProxyConfig) Validate() error {
	if c == nil {
		return nil
	}
	count := 0
	if c.UseDirectIP {
		count++
	}
	if c.UseXFFHeader {
		count++
	}
	if c.UseXRealIPHeader {
		count++
	}
	if count > 1 {
		return errors.New("proxy: only one of useDirectIP, useXFFHeader, useXRealIPHeader may be enabled")
	}
	return nil
}

func (c *ProxyConfig) MarshalZerologObject(e *zerolog.Event) {
	e.Bool("useDirectIP", c.UseDirectIP).
		Bool("useXFFHeader", c.UseXFFHeader).
		Bool("useXRealIPHeader", c.UseXRealIPHeader).
		Strs("trustedIPRanges", c.TrustedIPRanges)
}

func DefaultProxyConfig() *ProxyConfig {
	return &ProxyConfig{
		UseDirectIP:      true,
		UseXFFHeader:     false,
		UseXRealIPHeader: false,
		TrustedIPRanges:  []string{},
	}

}
