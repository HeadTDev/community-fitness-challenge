package services

import (
	"testing"

	"github.com/HeadTDev/fitchallenge/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScoringServiceReferenceCase(t *testing.T) {
	svc := NewScoringService()

	result, err := svc.Calculate(ScoreInput{
		Steps:         12000,
		Calories:      650,
		ActiveMinutes: 45,
	})
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, ScoringVersionV1, result.Version)
	assert.Equal(t, 72.50, result.Score)
}

func TestScoringServiceValidation(t *testing.T) {
	svc := NewScoringService()

	t.Run("rejects all zero", func(t *testing.T) {
		_, err := svc.Calculate(ScoreInput{})
		assert.ErrorIs(t, err, domain.ErrInvalidInput)
	})

	t.Run("rejects negatives", func(t *testing.T) {
		_, err := svc.Calculate(ScoreInput{Steps: -1})
		assert.ErrorIs(t, err, domain.ErrInvalidInput)
	})
}

func TestScoringServiceCapsToHundred(t *testing.T) {
	svc := NewScoringService()

	result, err := svc.Calculate(ScoreInput{
		Steps:         30000,
		Calories:      3000,
		ActiveMinutes: 300,
	})
	require.NoError(t, err)
	assert.Equal(t, 100.0, result.Score)
}
