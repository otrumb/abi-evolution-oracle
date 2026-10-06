package experiment

import (
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
	evidence.Baselines = evidence.Baselines[:49]
	// When
	score := ScoreEvidence(evidence)
	// Then
	require.Contains(t, score.HardGateFailures, "mandatory_baseline_execution")
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
			baselines = append(baselines, Baseline{FixtureID: id, Bump: "none"})
		}
	}
	return ScoringEvidence{Observations: observations, Baselines: baselines, Corpus: corpus.Summary{Total: 50, Directional: 12, Names: 12, Events: 12, Collisions: 8, Structural: 6}, Gates: GateEvidence{CorpusVerified: true, HashesVerified: true, ProvenanceVerified: true, ExpectationsVerified: true, PinsVerified: true, TestsVerified: true, Deterministic: true, EvidencePolicy: true, PublicationLocked: true}}
}
