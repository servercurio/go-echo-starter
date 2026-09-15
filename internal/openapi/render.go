// SPDX-License-Identifier: Apache-2.0

package openapi

import (
	"gopkg.in/yaml.v3"
)

// LicenseHeader is written at the top of rendered YAML specs so a regenerated
// docs/openapi.yaml matches the committed file and passes the License Headers
// check.
const LicenseHeader = "# SPDX-License-Identifier: Apache-2.0\n\n"

// MarshalYAML renders spec as YAML, prefixed with LicenseHeader.
func MarshalYAML(spec *Spec) ([]byte, error) {
	data, err := yaml.Marshal(spec)
	if err != nil {
		return nil, err
	}

	return append([]byte(LicenseHeader), data...), nil
}
