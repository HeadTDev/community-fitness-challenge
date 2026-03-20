package jwt

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJWTManager(t *testing.T) {
	secretKey := "test-secret-key"
	accessTTL := 1 * time.Minute
	refreshTTL := 5 * time.Minute
	manager := NewJWTManager(secretKey, accessTTL, refreshTTL)

	userID := "user-123"
	role := "admin"

	t.Run("Generate and Validate Access Token", func(t *testing.T) {
		token, err := manager.GenerateAccessToken(userID, role)
		require.NoError(t, err)
		assert.NotEmpty(t, token)

		claims, err := manager.ValidateToken(token)
		assert.NoError(t, err)
		assert.Equal(t, userID, claims.UserID)
		assert.Equal(t, role, claims.Role)
	})

	t.Run("Generate and Validate Refresh Token", func(t *testing.T) {
		token, err := manager.GenerateRefreshToken(userID)
		require.NoError(t, err)
		assert.NotEmpty(t, token)

		claims, err := manager.ValidateToken(token)
		assert.NoError(t, err)
		assert.Equal(t, userID, claims.UserID)
		assert.Empty(t, claims.Role) // Refresh tokenben nincs role
	})

	t.Run("Expired Token", func(t *testing.T) {
		shortManager := NewJWTManager(secretKey, -1*time.Second, refreshTTL)
		token, err := shortManager.GenerateAccessToken(userID, role)
		require.NoError(t, err)

		claims, err := shortManager.ValidateToken(token)
		assert.ErrorIs(t, err, ErrExpiredToken)
		assert.Nil(t, claims)
	})

	t.Run("Invalid Token", func(t *testing.T) {
		claims, err := manager.ValidateToken("invalid.token.string")
		assert.ErrorIs(t, err, ErrInvalidToken)
		assert.Nil(t, claims)
	})

	t.Run("Wrong Secret Key", func(t *testing.T) {
		token, err := manager.GenerateAccessToken(userID, role)
		require.NoError(t, err)

		wrongManager := NewJWTManager("wrong-secret", accessTTL, refreshTTL)
		claims, err := wrongManager.ValidateToken(token)
		assert.ErrorIs(t, err, ErrInvalidToken)
		assert.Nil(t, claims)
	})
}
