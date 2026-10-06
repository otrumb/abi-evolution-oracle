package corpus_test

import (
	"path/filepath"
	"testing"

	"github.com/otrumb/abi-evolution-oracle/internal/corpus"
	"github.com/stretchr/testify/require"
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

func Test_Verify_rejects_placeholder_fixture_semantics(t *testing.T) {
	// Given
	expected := map[string]bool{"executed": true}
	// When
	err := corpus.ValidateExpectations("collisions", expected)
	// Then
	require.Error(t, err)
}

func Test_ValidateExpectations_accepts_explicit_name_observations(t *testing.T) {
	// Given
	expected := map[string]bool{"wire_invariant": true, "signature_invariant": true, "generated_source_change": true, "old_consumer_compile_break": false}
	// When
	err := corpus.ValidateExpectations("names", expected)
	// Then
	require.NoError(t, err)
}
