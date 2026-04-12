package handler

import (
	"testing"

	"github.com/senran-N/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestBackendModeBlocksLogin(t *testing.T) {
	t.Run("allows admin in backend mode", func(t *testing.T) {
		require.False(t, backendModeBlocksLogin(true, &service.User{Role: service.RoleAdmin}))
	})

	t.Run("blocks non admin in backend mode", func(t *testing.T) {
		require.True(t, backendModeBlocksLogin(true, &service.User{Role: service.RoleUser}))
	})

	t.Run("ignores backend restriction when mode disabled", func(t *testing.T) {
		require.False(t, backendModeBlocksLogin(false, &service.User{Role: service.RoleUser}))
	})

	t.Run("ignores nil user", func(t *testing.T) {
		require.False(t, backendModeBlocksLogin(true, nil))
	})
}
