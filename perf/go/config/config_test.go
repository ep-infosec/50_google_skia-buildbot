package config

import (
	"path/filepath"

	"testing"

	"github.com/stretchr/testify/require"
)

func TestInstanceConfigBytes_AllExistingConfigs_ShouldBeValid(t *testing.T) {

	allExistingConfigs, err := filepath.Glob("../../configs/*.json")
	require.Greater(t, len(allExistingConfigs), 0)
	require.NoError(t, err)
	for _, filename := range allExistingConfigs {
		_, schemaErrors, err := InstanceConfigFromFile(filename)
		require.Len(t, schemaErrors, 0)
		require.NoError(t, err, filename)
	}
}

func TestInstanceConfigBytes_EmptyJSONObject_ShouldBeInValid(t *testing.T) {

	_, _, err := InstanceConfigFromFile("./testdata/empty.json")
	require.Error(t, err)
}
