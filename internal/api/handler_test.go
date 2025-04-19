package api

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lovelaze/nebula-sync/internal/sync"
)

func TestHealthHandler_healthy(t *testing.T) {
	state := sync.NewState()
	state.OnSuccess()

	require.Len(t, state.Stack, 1)
	require.True(t, state.Stack[0].Success)

	server := NewServer(state)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	resp := httptest.NewRecorder()

	server.router.ServeHTTP(resp, req)

	result := resp.Result()
	defer result.Body.Close()

	assert.Equal(t, 200, result.StatusCode)
}

func TestHealthHandler_unhealthy(t *testing.T) {
	state := sync.NewState()
	state.OnFailure(errors.New(("test error")))

	require.Len(t, state.Stack, 1)
	require.False(t, state.Stack[0].Success)

	server := NewServer(state)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	resp := httptest.NewRecorder()

	server.router.ServeHTTP(resp, req)

	result := resp.Result()
	defer result.Body.Close()

	assert.Equal(t, 500, result.StatusCode)
}
