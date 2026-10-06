package experiment

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_ParseBaselineOutput_derives_only_supported_function_facts(t *testing.T) {
	// Given
	raw := []byte(`{
  "breaking": [
    {
      "kind": "function",
      "signature": "probe()",
      "message": "return type changed (uint256 -> bytes)"
    }
  ],
  "additions": [],
  "notes": [],
  "bump": "major"
}`)
	// When
	parsed, err := parseBaselineOutput("D06", raw)
	// Then
	require.NoError(t, err)
	require.True(t, parsed.OutputParsed)
	require.Equal(t, []BaselineSource{
		{Section: "breaking", Index: 0, Kind: "function", Signature: "probe()", Message: "return type changed (uint256 -> bytes)"},
	}, parsed.Sources)
	require.Equal(t, []BaselineFact{{Name: "call_identity_unchanged", Value: true, Source: 0}}, parsed.Capabilities.Facts)
	require.False(t, parsed.Capabilities.DirectionalEquivalent())
}

func Test_ParseBaselineOutput_derives_only_explicit_indexed_layout_facts(t *testing.T) {
	// Given
	raw := []byte("{\"breaking\":[{\"kind\":\"event\",\"signature\":\"Transfer(address,address,uint256)\",\"message\":\"event `indexed` layout changed — log decoding will break\"}],\"additions\":[],\"notes\":[],\"bump\":\"major\"}")
	// When
	parsed, err := parseBaselineOutput("E04", raw)
	// Then
	require.NoError(t, err)
	require.Equal(t, []BaselineFact{
		{Name: "topic_identity", Value: true, Source: 0},
		{Name: "topic_layout_impact", Value: true, Source: 0},
		{Name: "data_layout_impact", Value: true, Source: 0},
		{Name: "cross_decode_impact", Value: true, Source: 0},
	}, parsed.Capabilities.Facts)
	require.False(t, parsed.Capabilities.EventsEquivalent())
}

func Test_ParseBaselineOutput_preserves_unsupported_entries_without_inventing_facts(t *testing.T) {
	// Given
	raw := []byte(`{"breaking":[{"kind":"function","signature":"probe()","message":"function removed"}],"additions":[{"kind":"event","signature":"Changed(uint256)","message":"event added"}],"notes":[],"bump":"major"}`)
	// When
	parsed, err := parseBaselineOutput("C01", raw)
	// Then
	require.NoError(t, err)
	require.Len(t, parsed.Sources, 2)
	require.Empty(t, parsed.Capabilities.Facts)
}

func Test_ParseBaselineOutput_empty_entries_do_not_establish_consumer_facts(t *testing.T) {
	// Given
	raw := []byte(`{"breaking":[],"additions":[],"notes":[],"bump":"none"}`)
	// When
	parsed, err := parseBaselineOutput("N03", raw)
	// Then
	require.NoError(t, err)
	require.True(t, parsed.OutputParsed)
	require.Empty(t, parsed.Sources)
	require.Empty(t, parsed.Capabilities.Facts)
}

func Test_ParseBaselineOutput_rejects_invalid_upstream_shape(t *testing.T) {
	// Given
	inputs := map[string]string{
		"malformed JSON":        `{"breaking":`,
		"missing breaking":      `{"additions":[],"notes":[],"bump":"none"}`,
		"missing additions":     `{"breaking":[],"notes":[],"bump":"none"}`,
		"missing notes":         `{"breaking":[],"additions":[],"bump":"none"}`,
		"missing bump":          `{"breaking":[],"additions":[],"notes":[]}`,
		"unknown top field":     `{"breaking":[],"additions":[],"notes":[],"bump":"none","extra":true}`,
		"consumer facts":        `{"breaking":[],"additions":[],"notes":[],"bump":"none","consumer_facts":{}}`,
		"trailing JSON":         `{"breaking":[],"additions":[],"notes":[],"bump":"none"}{}`,
		"missing kind":          `{"breaking":[{"signature":"probe()","message":"function removed"}],"additions":[],"notes":[],"bump":"major"}`,
		"missing signature":     `{"breaking":[{"kind":"function","message":"function removed"}],"additions":[],"notes":[],"bump":"major"}`,
		"missing message":       `{"breaking":[{"kind":"function","signature":"probe()"}],"additions":[],"notes":[],"bump":"major"}`,
		"empty kind":            `{"breaking":[{"kind":"","signature":"probe()","message":"function removed"}],"additions":[],"notes":[],"bump":"major"}`,
		"empty signature":       `{"breaking":[{"kind":"function","signature":"","message":"function removed"}],"additions":[],"notes":[],"bump":"major"}`,
		"empty message":         `{"breaking":[{"kind":"function","signature":"probe()","message":""}],"additions":[],"notes":[],"bump":"major"}`,
		"unknown entry field":   `{"breaking":[{"kind":"function","signature":"probe()","message":"function removed","extra":true}],"additions":[],"notes":[],"bump":"major"}`,
		"malformed entry value": `{"breaking":[{"kind":1,"signature":"probe()","message":"function removed"}],"additions":[],"notes":[],"bump":"major"}`,
	}
	for name, raw := range inputs {
		t.Run(name, func(t *testing.T) {
			// When
			_, err := parseBaselineOutput("D01", []byte(raw))
			// Then
			require.Error(t, err)
		})
	}
}
