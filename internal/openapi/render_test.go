// SPDX-License-Identifier: Apache-2.0

package openapi_test

import (
	"strings"
	"testing"

	asrt "github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"

	"github.com/servercurio/go-echo-starter/internal/openapi"
)

func TestMarshalYAMLPrefixesLicenseHeader(t *testing.T) {
	t.Parallel()
	assert := asrt.New(t)

	data, err := openapi.MarshalYAML(&openapi.Spec{OpenAPI: "3.0.3", Info: openapi.Info{Title: "Test", Version: "1.0.0"}})
	assert.NoError(err)

	out := string(data)
	assert.True(strings.HasPrefix(out, openapi.LicenseHeader), "rendered YAML must start with the SPDX header")

	var decoded map[string]any
	assert.NoError(yaml.Unmarshal(data, &decoded))
	assert.Equal("3.0.3", decoded["openapi"])
	assert.Equal(1, strings.Count(out, "SPDX-License-Identifier"))
}
