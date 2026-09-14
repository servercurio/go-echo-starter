package openapi

import (
	"testing"

	asrt "github.com/stretchr/testify/assert"
)

// TestSortTags_ReordersUnsortedInput pins sortTags against a reversed slice.
// Build feeds it tags from map iteration, so the swap branch is only hit when
// the random order happens to be unsorted; exercising it directly keeps
// coverage deterministic.
func TestSortTags_ReordersUnsortedInput(t *testing.T) {
	assert := asrt.New(t)

	tags := []Tag{{Name: "v1"}, {Name: "users"}, {Name: "api"}}
	sortTags(tags)

	assert.Equal([]Tag{{Name: "api"}, {Name: "users"}, {Name: "v1"}}, tags)
}
