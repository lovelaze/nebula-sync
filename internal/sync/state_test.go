package sync

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestState_Add(t *testing.T) {
	s := NewState()
	require.Empty(t, s.Stack)

	outcome := *NewOutcome(true)
	s.Add(outcome)

	assert.Len(t, s.Stack, 1)
	assert.Contains(t, s.Stack, outcome)
}

func TestState_OnSuccess(t *testing.T) {
	s := NewState()
	require.Empty(t, s.Stack)

	s.OnSuccess()

	assert.Len(t, s.Stack, 1)
	assert.True(t, s.Stack[0].Success)
	now := time.Now()
	assert.WithinRange(t, s.Stack[0].Timestamp, now.Add(-1*time.Second), now.Add(1*time.Second))
}

func TestState_OnFailure(t *testing.T) {
	s := NewState()
	require.Empty(t, s.Stack)

	s.OnFailure(errors.New("test error"))

	assert.Len(t, s.Stack, 1)
	assert.False(t, s.Stack[0].Success)
	now := time.Now()
	assert.WithinRange(t, s.Stack[0].Timestamp, now.Add(-1*time.Second), now.Add(1*time.Second))
}
