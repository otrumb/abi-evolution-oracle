package experiment

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"

	"local/abi-evolution-oracle-validation/internal/corpus"
)

type baselineOutput struct {
	Breaking      *[]json.RawMessage   `json:"breaking"`
	Additions     *[]json.RawMessage   `json:"additions"`
	Notes         *[]json.RawMessage   `json:"notes"`
	Bump          *string              `json:"bump"`
	ConsumerFacts BaselineCapabilities `json:"consumer_facts"`
}

func runBaseline(root, tool string, entry corpus.Entry) (Baseline, error) {
	command := exec.Command("node", tool, filepath.FromSlash(entry.Old), filepath.FromSlash(entry.New), "--json")
	command.Dir = root
	output, runErr := command.Output()
	exitCode := 0
	if runErr != nil {
		if exitErr, ok := runErr.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			return Baseline{}, fmt.Errorf("run abidiff: %w", runErr)
		}
	}
	baseline, err := parseBaselineOutput(entry.ID, output)
	if err != nil {
		return Baseline{}, err
	}
	baseline.ExitCode = exitCode
	return baseline, nil
}

func parseBaselineOutput(fixtureID string, output []byte) (Baseline, error) {
	var decoded baselineOutput
	if err := json.Unmarshal(output, &decoded); err != nil {
		return Baseline{}, fmt.Errorf("decode abidiff %s: %w", fixtureID, err)
	}
	if decoded.Breaking == nil || decoded.Additions == nil || decoded.Notes == nil || decoded.Bump == nil {
		return Baseline{}, fmt.Errorf("decode abidiff %s: missing required fields", fixtureID)
	}
	return Baseline{FixtureID: fixtureID, Bump: *decoded.Bump, Breaking: len(*decoded.Breaking), Additions: len(*decoded.Additions), Capabilities: decoded.ConsumerFacts, CapabilitiesParsed: true}, nil
}
