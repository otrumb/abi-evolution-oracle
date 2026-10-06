package experiment

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"local/abi-evolution-oracle-validation/internal/corpus"
)

func ScoreRecorded(root, evidenceRoot, gatePath string) (Score, error) {
	observations, err := readJSONL[Observation](filepath.Join(evidenceRoot, "observations.jsonl"))
	if err != nil {
		return Score{}, err
	}
	baselines, err := readJSONL[Baseline](filepath.Join(evidenceRoot, "abidiff.jsonl"))
	if err != nil {
		return Score{}, err
	}
	gateData, err := os.ReadFile(gatePath)
	if err != nil {
		return Score{}, fmt.Errorf("read gates: %w", err)
	}
	var gates GateEvidence
	if err := json.Unmarshal(gateData, &gates); err != nil {
		return Score{}, fmt.Errorf("decode gates: %w", err)
	}
	corpusSummary, err := corpus.Verify(root)
	if err != nil {
		return Score{}, err
	}
	score := ScoreEvidence(ScoringEvidence{Observations: observations, Baselines: baselines, Corpus: corpusSummary, Gates: gates})
	data, err := json.MarshalIndent(score, "", "  ")
	if err != nil {
		return Score{}, fmt.Errorf("encode score: %w", err)
	}
	if err := os.WriteFile(filepath.Join(evidenceRoot, "score.json"), append(data, '\n'), 0644); err != nil {
		return Score{}, fmt.Errorf("write score: %w", err)
	}
	return score, nil
}

func readJSONL[T any](path string) ([]T, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open records: %w", err)
	}
	defer file.Close()
	var values []T
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var value T
		if err := json.Unmarshal(scanner.Bytes(), &value); err != nil {
			return nil, fmt.Errorf("decode record: %w", err)
		}
		values = append(values, value)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan records: %w", err)
	}
	return values, nil
}
