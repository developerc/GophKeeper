package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig(t *testing.T) {
	serverSettings, err := NewServerSettings()
	require.NoError(t, err)
	assert.NotEqual(t, serverSettings, nil, "server settings should not be nil")
}
