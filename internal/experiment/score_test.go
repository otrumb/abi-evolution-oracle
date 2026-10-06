package experiment

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"local/abi-evolution-oracle-validation/internal/corpus"
)

func Test_ScoreEvidence_rejects_mutated_observation_status(t *testing.T) {
	// Given
	evidence := completeScoringEvidence()
	evidence.Observations[0].Status = "rejected"
	// When
	score := ScoreEvidence(evidence)
	// Then
	require.Less(t, score.Score, 100)
	require.Equal(t, "NO-GO", score.TechnicalVerdict)
}

func Test_ScoreEvidence_rejects_missing_baseline_record(t *testing.T) {
	// Given
	evidence := completeScoringEvidence()
	evidence.Baselines = append(evidence.Baselines[:8], evidence.Baselines[9:]...)
	// When
	score := ScoreEvidence(evidence)
	// Then
	require.Contains(t, score.HardGateFailures, "mandatory_baseline_execution")
	require.NotContains(t, score.ClassesBeatingBaseline, "directional")
	require.Equal(t, "NO-GO", score.TechnicalVerdict)
}

func Test_ScoreEvidence_rejects_duplicate_baseline_record_for_class_win(t *testing.T) {
	// Given
	evidence := completeScoringEvidence()
	evidence.Baselines = append(evidence.Baselines, evidence.Baselines[8])
	// When
	score := ScoreEvidence(evidence)
	// Then
	require.Contains(t, score.HardGateFailures, "mandatory_baseline_execution")
	require.NotContains(t, score.ClassesBeatingBaseline, "directional")
	require.Equal(t, "NO-GO", score.TechnicalVerdict)
}

func Test_ScoreEvidence_rejects_unparsed_baseline_capabilities(t *testing.T) {
	// Given
	evidence := completeScoringEvidence()
	evidence.Baselines[8].OutputParsed = false
	// When
	score := ScoreEvidence(evidence)
	// Then
	require.Contains(t, score.HardGateFailures, "baseline_capability_parse")
	require.NotContains(t, score.ClassesBeatingBaseline, "directional")
	require.Equal(t, "NO-GO", score.TechnicalVerdict)
}

func Test_ScoreEvidence_requires_actionable_consumer_detail_for_class_win(t *testing.T) {
	// Given
	evidence := completeScoringEvidence()
	for index := range evidence.Observations {
		if evidence.Observations[index].Class == "events" {
			evidence.Observations[index].Actionable = false
		}
	}
	// When
	score := ScoreEvidence(evidence)
	// Then
	require.NotContains(t, score.ClassesBeatingBaseline, "events")
}

func Test_ScoreEvidence_all_equivalent_baselines_produce_zero_wins(t *testing.T) {
	// Given
	evidence := completeScoringEvidence()
	for index := range evidence.Baselines {
		evidence.Baselines[index].Capabilities = allConsumerCapabilities()
	}
	// When
	score := ScoreEvidence(evidence)
	// Then
	require.Empty(t, score.ClassesBeatingBaseline)
	require.Equal(t, "NO-GO", score.TechnicalVerdict)
}

func Test_ScoreEvidence_incomplete_equivalent_facts_permit_class_wins(t *testing.T) {
	// Given
	evidence := completeScoringEvidence()
	for index := range evidence.Baselines {
		evidence.Baselines[index].Capabilities.Facts = nil
	}
	// When
	score := ScoreEvidence(evidence)
	// Then
	require.ElementsMatch(t, []string{"directional", "events", "names"}, score.ClassesBeatingBaseline)
	require.Equal(t, "GO", score.TechnicalVerdict)
}

func Test_ScoreEvidence_one_equivalent_class_removes_only_that_win(t *testing.T) {
	// Given
	evidence := completeScoringEvidence()
	for index := range evidence.Baselines {
		if strings.HasPrefix(evidence.Baselines[index].FixtureID, "events") {
			evidence.Baselines[index].Capabilities = allConsumerCapabilities()
		}
	}
	// When
	score := ScoreEvidence(evidence)
	// Then
	require.ElementsMatch(t, []string{"directional", "names"}, score.ClassesBeatingBaseline)
}

func Test_ScoreEvidence_enforces_section_floor_and_hard_gate(t *testing.T) {
	// Given
	evidence := completeScoringEvidence()
	evidence.Gates.Deterministic = false
	evidence.Gates.PinsVerified = false
	evidence.Gates.TestsVerified = false
	// When
	score := ScoreEvidence(evidence)
	// Then
	require.Contains(t, score.HardGateFailures, "determinism")
	require.Less(t, score.Sections.Reproducibility.Earned, score.Sections.Reproducibility.Minimum)
	require.Contains(t, score.HardGateFailures, "section_floor")
	require.Equal(t, "NO-GO", score.TechnicalVerdict)
}

func completeScoringEvidence() ScoringEvidence {
	classes := []struct {
		name   string
		count  int
		status string
	}{{"collisions", 8, "ambiguous"}, {"directional", 12, "observed"}, {"events", 12, "observed"}, {"names", 12, "observed"}, {"structural", 6, "observed"}}
	var observations []Observation
	var baselines []Baseline
	for _, class := range classes {
		for index := range class.count {
			id := class.name + string(rune('A'+index))
			observations = append(observations, Observation{FixtureID: id, Class: class.name, Status: class.status, Actionable: true, Assessment: scope(), CandidateCount: 2})
			baselines = append(baselines, Baseline{FixtureID: id, Bump: "none", OutputParsed: true, Capabilities: noConsumerCapabilities()})
		}
	}
	return ScoringEvidence{Observations: observations, Baselines: baselines, Corpus: corpus.Summary{Total: 50, Directional: 12, Names: 12, Events: 12, Collisions: 8, Structural: 6}, Gates: GateEvidence{CorpusVerified: true, HashesVerified: true, ProvenanceVerified: true, ExpectationsVerified: true, PinsVerified: true, TestsVerified: true, Deterministic: true, EvidencePolicy: true, PublicationLocked: true}}
}

func allConsumerCapabilities() BaselineCapabilities {
	names := []string{"call_identity_unchanged", "old_consumer_new_producer_decode", "new_consumer_old_producer_decode", "wire_identity_unchanged", "generated_api_impact", "source_compile_impact", "topic_identity", "topic_layout_impact", "data_layout_impact", "filter_impact", "cross_decode_impact"}
	facts := make([]BaselineFact, 0, len(names))
	for _, name := range names {
		facts = append(facts, BaselineFact{Name: name, Value: true})
	}
	return BaselineCapabilities{Facts: facts}
}

func noConsumerCapabilities() BaselineCapabilities {
	capabilities := allConsumerCapabilities()
	for index := range capabilities.Facts {
		capabilities.Facts[index].Value = false
	}
	return capabilities
}
