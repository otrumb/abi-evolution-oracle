package experiment

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_ParseBaselineOutput_actual_abidiff_shape_has_no_consumer_equivalence(t *testing.T) {
	// Given
	raw := []byte(`{"breaking":[{"kind":"function","signature":"probe()","message":"return type changed (uint256 -\u003e bytes)"}],"additions":[],"notes":[],"bump":"major"}`)
	// When
	parsed, err := parseBaselineOutput("D06", raw)
	// Then
	require.NoError(t, err)
	require.False(t, parsed.Capabilities.DirectionalEquivalent())
	require.False(t, parsed.Capabilities.NamesEquivalent())
	require.False(t, parsed.Capabilities.EventsEquivalent())
}

func Test_ParseBaselineOutput_directional_equivalence_requires_all_facts(t *testing.T) {
	// Given
	raw := []byte(`{"breaking":[],"additions":[],"notes":[],"bump":"none","consumer_facts":{"call_identity_unchanged":true,"old_consumer_new_producer_decode":true,"new_consumer_old_producer_decode":true}}`)
	// When
	parsed, err := parseBaselineOutput("D01", raw)
	// Then
	require.NoError(t, err)
	require.True(t, parsed.Capabilities.DirectionalEquivalent())
}

func Test_ParseBaselineOutput_names_equivalence_requires_wire_api_and_compile_facts(t *testing.T) {
	// Given
	raw := []byte(`{"breaking":[],"additions":[],"notes":[],"bump":"none","consumer_facts":{"wire_identity_unchanged":true,"generated_api_impact":true,"source_compile_impact":true}}`)
	// When
	parsed, err := parseBaselineOutput("N03", raw)
	// Then
	require.NoError(t, err)
	require.True(t, parsed.Capabilities.NamesEquivalent())
}

func Test_ParseBaselineOutput_events_equivalence_requires_all_topic_and_decode_facts(t *testing.T) {
	// Given
	raw := []byte(`{"breaking":[],"additions":[],"notes":[],"bump":"none","consumer_facts":{"topic_identity":true,"topic_layout_impact":true,"data_layout_impact":true,"filter_impact":true,"cross_decode_impact":true}}`)
	// When
	parsed, err := parseBaselineOutput("E07", raw)
	// Then
	require.NoError(t, err)
	require.True(t, parsed.Capabilities.EventsEquivalent())
}

func Test_ParseBaselineOutput_rejects_malformed_or_missing_required_shape(t *testing.T) {
	// Given
	inputs := [][]byte{[]byte(`{"breaking":`), []byte(`{"bump":"none"}`)}
	for _, raw := range inputs {
		// When
		_, err := parseBaselineOutput("D01", raw)
		// Then
		require.Error(t, err)
	}
}
