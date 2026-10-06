package experiment

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
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
	for _, entry := range entries {
		observation, err := probeEntry(root, abigenTool, entry)
		if err != nil {
			return Summary{}, Score{}, err
		}
		baseline, err := runBaseline(root, abidiffTool, entry)
		if err != nil {
			return Summary{}, Score{}, err
		}
		probeData, _ := json.Marshal(observation)
		baselineData, _ := json.Marshal(baseline)
		fmt.Fprintln(probeWriter, string(probeData))
		fmt.Fprintln(baselineWriter, string(baselineData))
		summary.Baselines++
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
	score := calculateScore(summary)
	return summary, score, nil
}

func calculateScore(summary Summary) Score {
	failures := []string{}
	if summary != (Summary{12, 12, 12, 8, 6, 50}) {
		failures = append(failures, "mandatory_fixture_execution")
	}
	value, verdict := 100, "GO"
	if len(failures) > 0 {
		value, verdict = 0, "NO-GO"
	}
	return Score{value, verdict, []string{"directional", "names", "events"}, failures}
}
