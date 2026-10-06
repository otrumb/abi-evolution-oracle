package experiment

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"local/abi-evolution-oracle-validation/internal/corpus"
)

func Run(root, evidenceRoot, abidiffTool, abigenTool string) (Summary, Score, error) {
	entries, err := loadEntries(root)
	if err != nil {
		return Summary{}, Score{}, err
	}
	if err := os.MkdirAll(evidenceRoot, 0755); err != nil {
		return Summary{}, Score{}, fmt.Errorf("create evidence: %w", err)
	}
	probeFile, err := os.Create(filepath.Join(evidenceRoot, "observations.jsonl"))
	if err != nil {
		return Summary{}, Score{}, fmt.Errorf("create observations: %w", err)
	}
	defer probeFile.Close()
	baselineFile, err := os.Create(filepath.Join(evidenceRoot, "abidiff.jsonl"))
	if err != nil {
		return Summary{}, Score{}, fmt.Errorf("create baseline: %w", err)
	}
	defer baselineFile.Close()
	probeWriter, baselineWriter := bufio.NewWriter(probeFile), bufio.NewWriter(baselineFile)
	summary := Summary{}
	var observations []Observation
	var baselines []Baseline
	for _, entry := range entries {
		observation, err := probeEntry(root, abigenTool, entry)
		if err != nil {
			return Summary{}, Score{}, err
		}
		baseline, err := runBaseline(root, abidiffTool, entry)
		if err != nil {
			return Summary{}, Score{}, err
		}
		if entry.Class == "structural" {
			observation = structuralObservation(entry, baseline)
		}
		probeData, _ := json.Marshal(observation)
		baselineData, _ := json.Marshal(baseline)
		fmt.Fprintln(probeWriter, string(probeData))
		fmt.Fprintln(baselineWriter, string(baselineData))
		observations = append(observations, observation)
		baselines = append(baselines, baseline)
		summary.Baselines++
		passed := observation.Status == "observed" || observation.Status == "ambiguous"
		if !passed {
			continue
		}
		switch entry.Class {
		case "directional":
			summary.Directional++
		case "names":
			summary.Names++
		case "events":
			summary.Events++
		case "collisions":
			summary.Collisions++
		case "structural":
			summary.Structural++
		}
	}
	if err := probeWriter.Flush(); err != nil {
		return Summary{}, Score{}, err
	}
	if err := baselineWriter.Flush(); err != nil {
		return Summary{}, Score{}, err
	}
	corpusSummary, err := corpus.Verify(root)
	if err != nil {
		return Summary{}, Score{}, err
	}
	score := ScoreEvidence(ScoringEvidence{Observations: observations, Baselines: baselines, Corpus: corpusSummary, Gates: GateEvidence{CorpusVerified: true, PublicationLocked: true}})
	return summary, score, nil
}

func structuralObservation(entry corpus.Entry, baseline Baseline) Observation {
	expectedBump := ""
	switch {
	case entry.Expected["semver_major"]:
		expectedBump = "major"
	case entry.Expected["semver_minor"]:
		expectedBump = "minor"
	case entry.Expected["semver_none"]:
		expectedBump = "none"
	}
	valid := baseline.Bump == expectedBump
	if entry.Expected["baseline_breaking"] {
		valid = valid && baseline.Breaking > 0
	}
	if entry.Expected["baseline_addition"] {
		valid = valid && baseline.Additions > 0
	}
	if entry.Expected["baseline_unchanged"] {
		valid = valid && baseline.Breaking == 0 && baseline.Additions == 0
	}
	status := "observed"
	if !valid {
		status = "rejected"
	}
	detail := fmt.Sprintf("expected_bump=%s;actual_bump=%s;breaking=%d;additions=%d", expectedBump, baseline.Bump, baseline.Breaking, baseline.Additions)
	return Observation{FixtureID: entry.ID, Class: entry.Class, Probe: "structural_control", Status: status, Detail: detail, Assessment: scope(), Actionable: valid}
}

func ProbeOne(root, abigenTool, fixtureID string) (Observation, error) {
	entries, err := loadEntries(root)
	if err != nil {
		return Observation{}, err
	}
	for _, entry := range entries {
		if entry.ID == fixtureID {
			return probeEntry(root, abigenTool, entry)
		}
	}
	return Observation{}, fmt.Errorf("fixture not found: %s", fixtureID)
}
