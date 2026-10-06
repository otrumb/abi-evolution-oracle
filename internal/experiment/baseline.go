package experiment

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"

	"local/abi-evolution-oracle-validation/internal/corpus"
)

type baselineOutput struct {
	Breaking  []json.RawMessage `json:"breaking"`
	Additions []json.RawMessage `json:"additions"`
	Bump      string            `json:"bump"`
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
	var decoded baselineOutput
	if err := json.Unmarshal(output, &decoded); err != nil {
		return Baseline{}, fmt.Errorf("decode abidiff %s: %w", entry.ID, err)
	}
	return Baseline{FixtureID: entry.ID, ExitCode: exitCode, Bump: decoded.Bump, Breaking: len(decoded.Breaking), Additions: len(decoded.Additions), ConsumerDetail: false}, nil
}
