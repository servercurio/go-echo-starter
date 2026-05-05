package version

import (
	"strings"

	"github.com/Masterminds/semver/v3"
)

func Commit() string {
	return strings.TrimSpace(strings.ToLower(commit))
}

func SemVer() *semver.Version {
	if v, err := semver.NewVersion(version); err == nil {
		return v
	}

	return semver.New(0, 0, 0, "", "")
}

func Number() string {
	return SemVer().String()
}

func Tag() string {
	return "v" + SemVer().String()
}
