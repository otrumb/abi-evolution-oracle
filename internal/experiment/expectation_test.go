package experiment

import (
	"strings"
	"testing"

	ethabi "github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/stretchr/testify/require"
	"local/abi-evolution-oracle-validation/internal/corpus"
)

func Test_EventObservation_rejects_unchanged_E01_layout(t *testing.T) {
	// Given
	parsed, err := ethabi.JSON(strings.NewReader(`[{"type":"event","name":"Observed","anonymous":false,"inputs":[{"name":"value","type":"uint256","indexed":false}]}]`))
	require.NoError(t, err)
	entry := corpus.Entry{ID: "E01", Class: "events", Expected: map[string]bool{"layout_changed": true, "consumer_impact": true}}
	// When
	observation := eventObservation(entry, parsed, parsed)
	// Then
	require.Equal(t, "rejected", observation.Status)
}

func Test_CollisionObservation_does_not_invent_two_candidates(t *testing.T) {
	// Given
	parsed, err := ethabi.JSON(strings.NewReader(`[{"type":"function","name":"probe","stateMutability":"view","inputs":[],"outputs":[]}]`))
	require.NoError(t, err)
	entry := corpus.Entry{ID: "C03", Class: "collisions", Expected: map[string]bool{"ambiguity_expected": true}}
	// When
	observation := collisionObservation(entry, parsed, parsed)
	// Then
	require.NotContains(t, observation.Detail, "candidate_count=2")
}

func Test_StructuralObservation_rejects_S06_semantic_addition(t *testing.T) {
	// Given
	entry := corpus.Entry{ID: "S06", Class: "structural", Expected: map[string]bool{"baseline_unchanged": true, "semver_none": true}}
	baseline := Baseline{FixtureID: "S06", Bump: "minor", Additions: 1}
	// When
	observation := structuralObservation(entry, baseline)
	// Then
	require.Equal(t, "rejected", observation.Status)
}
