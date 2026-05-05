package application

import (
	"testing"
)

// TestProxyConfig_FromEnv_TrustedIPRanges pins the parsing contract for
// APP_PROXY_TRUSTED_IP_RANGES. The pre-fix implementation panicked on
// any non-empty value — it indexed the destination slice rather than
// the freshly-split source — and even when the index was lucky it
// never appended results. This table covers the matrix that broke
// previously:
//
//   - empty → no-op, slice stays empty
//   - single valid CIDR → singleton list
//   - multiple valid CIDRs (with surrounding whitespace, blank segment
//     from a trailing comma) → all retained, in source order
//   - mixed valid + invalid → only valid entries retained
//   - all invalid → empty list, no panic
//   - IPv6 → retained (net.ParseCIDR accepts both families)
func TestProxyConfig_FromEnv_TrustedIPRanges(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  []string
	}{
		{name: "empty", value: "", want: []string{}},
		{name: "single valid", value: "10.0.0.0/8", want: []string{"10.0.0.0/8"}},
		{
			name:  "multiple valid with whitespace and trailing comma",
			value: " 10.0.0.0/8 , 192.168.0.0/16 , ",
			want:  []string{"10.0.0.0/8", "192.168.0.0/16"},
		},
		{
			name:  "mixed valid and invalid",
			value: "10.0.0.0/8,not-a-cidr,192.168.0.0/16",
			want:  []string{"10.0.0.0/8", "192.168.0.0/16"},
		},
		{name: "all invalid", value: "not-a-cidr, also-bogus", want: []string{}},
		{name: "ipv6", value: "2001:db8::/32", want: []string{"2001:db8::/32"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("APP_PROXY_TRUSTED_IP_RANGES", tt.value)

			cfg := DefaultProxyConfig()
			cfg.FromEnv("APP_PROXY")

			if len(cfg.TrustedIPRanges) != len(tt.want) {
				t.Fatalf("TrustedIPRanges length: got %v (%d), want %v (%d)",
					cfg.TrustedIPRanges, len(cfg.TrustedIPRanges), tt.want, len(tt.want))
			}
			for i, got := range cfg.TrustedIPRanges {
				if got != tt.want[i] {
					t.Fatalf("TrustedIPRanges[%d]: got %q, want %q", i, got, tt.want[i])
				}
			}
		})
	}
}
