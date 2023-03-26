package application

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewConfig(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("no envs", func(t *testing.T) {

		config, err := NewConfig(ctx)
		require.NoError(t, err)

		assert.Equal(t, "./db", config.DB.Path)
	})

	t.Run("env without prefix", func(t *testing.T) {
		require.NoError(t, os.Setenv("DB_PATH", "./old_db"))

		config, err := NewConfig(ctx)
		require.NoError(t, err)

		assert.Equal(t, "./old_db", config.DB.Path)
	})

	t.Run("prefix env override", func(t *testing.T) {
		require.NoError(t, os.Setenv("WEBARCHIVE_DB_PATH", "./new_db"))

		config, err := NewConfig(ctx)
		require.NoError(t, err)

		assert.Equal(t, "./new_db", config.DB.Path)
	})
}
