package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildTagGroupsUsesConfiguredOrderAndKeepsOtherTags(t *testing.T) {
	groups := buildTagGroups(map[string]int{
		"running": 2,
		"byu":     1,
		"python":  3,
		"unknown": 4,
	})

	assert.Equal(t, "Running", groups[0].Name)
	assert.Equal(t, []string{"running", "byu"}, groups[0].Tags)
	assert.Equal(t, "Programming", groups[1].Name)
	assert.Equal(t, []string{"python"}, groups[1].Tags)
	assert.Equal(t, "Other", groups[2].Name)
	assert.Equal(t, []string{"unknown"}, groups[2].Tags)
}
