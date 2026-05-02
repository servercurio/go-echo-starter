package version

import (
	_ "embed"
)

//go:embed commit.txt
var commit string

//go:embed version.txt
var version string
