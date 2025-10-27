package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRun(t *testing.T) {
	app, err := InitializeApp()

	require.NoError(t, err)
	assert.NotNil(t, app)
}
