package experiment

import (
	"sort"

	"local/abi-evolution-oracle-validation/internal/corpus"
)

func ScoreEvidence(evidence ScoringEvidence) Score {
	byClass := map[string][]Observation{}
	byID := map[string]Baseline{}
	for _, observation := range evidence.Observations {
		byClass[observation.Class] = append(byClass[observation.Class], observation)
	}
	for _, baseline := range evidence.Baselines {
		byID[baseline.FixtureID] = baseline
	}
	consumer := 0
	if classPasses(byClass["directional"], 12, "observed") {
		consumer += 10
	}
	if classPasses(byClass["names"], 12, "observed") {
		consumer += 10
	}
	if classPasses(byClass["events"], 12, "observed") {
		consumer += 10
	}
	if collisionsPass(byClass["collisions"]) {
		consumer += 5
	}
	if structuralPasses(byClass["structural"]) {
		consumer += 5
	}
	corpusScore := 0
	if evidence.Corpus == (corpus.Summary{Total: 50, Directional: 12, Names: 12, Events: 12, Collisions: 8, Structural: 6}) {
		corpusScore += 5
	}
	if evidence.Gates.HashesVerified {
		corpusScore += 5
	}
	if evidence.Gates.ProvenanceVerified {
		corpusScore += 4
	}
	if evidence.Gates.ExpectationsVerified && scopesPass(evidence.Observations) {
		corpusScore += 3
	}
	if len(evidence.Observations) == 50 && len(evidence.Baselines) == 50 {
		corpusScore += 3
	}
	wins := classWins(byClass, byID)
	differentiation := 0
	if len(evidence.Baselines) == 50 {
		differentiation += 5
	}
	if len(evidence.Observations) == 50 && len(byID) == 50 {
		differentiation += 5
	}
	if len(wins) >= 1 {
		differentiation += 5
	}
	if len(wins) >= 2 {
		differentiation += 5
	}
	reproducibility := 0
	if evidence.Gates.PinsVerified {
		reproducibility += 3
	}
	if evidence.Gates.Deterministic {
		reproducibility += 5
	}
	if evidence.Gates.TestsVerified && evidence.Gates.CorpusVerified {
		reproducibility += 5
	}
	if evidence.Gates.EvidencePolicy {
		reproducibility += 2
	}
	scopeScore := 0
	if scopesPass(evidence.Observations) {
		scopeScore += 3
	}
	if evidence.Gates.PublicationLocked {
		scopeScore += 2
	}
	sections := ScoreSections{SectionScore{consumer, 40, 20}, SectionScore{corpusScore, 20, 10}, SectionScore{differentiation, 20, 10}, SectionScore{reproducibility, 15, 8}, SectionScore{scopeScore, 5, 3}}
	total := consumer + corpusScore + differentiation + reproducibility + scopeScore
	failures := hardFailures(evidence, sections, wins, byID)
	verdict := "GO"
	if total < 85 || len(failures) > 0 {
		verdict = "NO-GO"
	}
	return Score{Score: total, TechnicalVerdict: verdict, ClassesBeatingBaseline: wins, HardGateFailures: failures, Sections: sections}
}

func classPasses(values []Observation, count int, status string) bool {
	if len(values) != count {
		return false
	}
	for _, value := range values {
		if value.Status != status || !value.Actionable {
			return false
		}
	}
	return true
}
func collisionsPass(values []Observation) bool {
	if len(values) != 8 {
		return false
	}
	for _, value := range values {
		if value.Status != "ambiguous" || value.CandidateCount < 2 || value.ArbitraryWinner {
			return false
		}
	}
	return true
}
func structuralPasses(values []Observation) bool { return classPasses(values, 6, "observed") }
func scopesPass(values []Observation) bool {
	if len(values) == 0 {
		return false
	}
	for _, value := range values {
		if value.Assessment != scope() {
			return false
		}
	}
	return true
}
func classWins(classes map[string][]Observation, baselines map[string]Baseline) []string {
	var wins []string
	for _, class := range []string{"directional", "names", "events"} {
		values := classes[class]
		if !classPasses(values, 12, "observed") {
			continue
		}
		equivalent := false
		for _, value := range values {
			baseline, ok := baselines[value.FixtureID]
			if !ok || baseline.ConsumerDetail {
				equivalent = true
				break
			}
		}
		if !equivalent {
			wins = append(wins, class)
		}
	}
	sort.Strings(wins)
	return wins
}
func hardFailures(evidence ScoringEvidence, sections ScoreSections, wins []string, baselines map[string]Baseline) []string {
	failures := []string{}
	if len(evidence.Observations) != 50 {
		failures = append(failures, "mandatory_fixture_execution")
	}
	if len(evidence.Baselines) != 50 || len(baselines) != 50 {
		failures = append(failures, "mandatory_baseline_execution")
	}
	if !collisionsPass(filterClass(evidence.Observations, "collisions")) {
		failures = append(failures, "collision_ambiguity")
	}
	if !evidence.Gates.Deterministic {
		failures = append(failures, "determinism")
	}
	if len(wins) < 2 {
		failures = append(failures, "baseline_differentiation")
	}
	for _, section := range []SectionScore{sections.Consumer, sections.Corpus, sections.Differentiation, sections.Reproducibility, sections.Scope} {
		if section.Earned < section.Minimum {
			failures = append(failures, "section_floor")
			break
		}
	}
	return failures
}
func filterClass(values []Observation, class string) []Observation {
	var result []Observation
	for _, value := range values {
		if value.Class == class {
			result = append(result, value)
		}
	}
	return result
}
