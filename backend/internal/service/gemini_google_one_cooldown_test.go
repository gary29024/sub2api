//go:build unit

package service

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestHandleGeminiUpstreamError_GoogleOneUsesTierCooldown(t *testing.T) {
	repo := &geminiErrorPolicyRepo{}
	svc := &GeminiMessagesCompatService{
		accountRepo: repo,
	}

	account := &Account{
		ID:       42,
		Type:     AccountTypeOAuth,
		Platform: PlatformGemini,
		Credentials: map[string]any{
			"oauth_type": "google_one",
			"tier_id":    "google_ai_pro",
		},
	}

	before := time.Now()
	svc.handleGeminiUpstreamError(context.Background(), account, 429, http.Header{}, []byte(`{"error":{"message":"rate limited"}}`))
	after := time.Now()

	require.Equal(t, 1, repo.setRateLimitedCalls)
	require.Equal(t, 0, repo.setErrorCalls)
	require.Equal(t, 0, repo.setTempCalls)
	require.False(t, repo.lastRateLimitedAt.IsZero())

	cooldown := repo.lastRateLimitedAt.Sub(before)
	require.Greater(t, cooldown, 4*time.Minute)
	require.Less(t, cooldown, 10*time.Minute)
	require.True(t, repo.lastRateLimitedAt.Before(after.Add(10*time.Minute)))
}
