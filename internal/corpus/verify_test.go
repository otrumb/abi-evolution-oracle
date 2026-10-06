package corpus_test

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"local/abi-evolution-oracle-validation/internal/corpus"
)

func Test_Verify_accepts_frozen_50_pair_corpus(t *testing.T) {
	// Given
	root := filepath.Join("..", "..")
	// When
	summary, err := corpus.Verify(root)
	// Then
	require.NoError(t, err)
	require.Equal(t, 50, summary.Total)
}
