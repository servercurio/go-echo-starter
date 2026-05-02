package application

import (
	"net"
	"strings"

	"github.com/rs/zerolog"
	"github.com/servercurio/go-echo-starter/internal/env"
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
	if trustedIPRangeString != "" {
		ranges := strings.Split(trustedIPRangeString, ",")
		for i := range ranges {
			r := strings.TrimSpace(c.TrustedIPRanges[i])
			if _, network, err := net.ParseCIDR(r); err == nil && network != nil {
				c.TrustedIPRanges[i] = r
			}
		}
	} else {
		c.TrustedIPRanges = []string{}
	}
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
