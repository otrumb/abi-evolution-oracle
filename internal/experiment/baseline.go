package experiment

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
	"regexp"

	"local/abi-evolution-oracle-validation/internal/corpus"
)

type baselineOutput struct {
	Breaking  *[]baselineEntry `json:"breaking"`
	Additions *[]baselineEntry `json:"additions"`
	Notes     *[]baselineEntry `json:"notes"`
	Bump      *string          `json:"bump"`
}

type baselineEntry struct {
	Kind      *string `json:"kind"`
	Signature *string `json:"signature"`
	Message   *string `json:"message"`
}

var outputChangePattern = regexp.MustCompile(`^return type changed \(.+ -> .+\)$`)
var mutabilityChangePattern = regexp.MustCompile(`^stateMutability changed \((pure|view|nonpayable|payable) -> (pure|view|nonpayable|payable)\)$`)

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
	decoder := json.NewDecoder(bytes.NewReader(output))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&decoded); err != nil {
		return Baseline{}, fmt.Errorf("decode abidiff %s: %w", fixtureID, err)
	}
	if err := rejectTrailingJSON(decoder); err != nil {
		return Baseline{}, fmt.Errorf("decode abidiff %s: %w", fixtureID, err)
	}
	if decoded.Breaking == nil || decoded.Additions == nil || decoded.Notes == nil || decoded.Bump == nil {
		return Baseline{}, fmt.Errorf("decode abidiff %s: missing required fields", fixtureID)
	}
	if *decoded.Bump != "major" && *decoded.Bump != "minor" && *decoded.Bump != "none" {
		return Baseline{}, fmt.Errorf("decode abidiff %s: invalid bump %q", fixtureID, *decoded.Bump)
	}
	sources, err := baselineSources(*decoded.Breaking, *decoded.Additions, *decoded.Notes)
	if err != nil {
		return Baseline{}, fmt.Errorf("decode abidiff %s: %w", fixtureID, err)
	}
	facts := deriveBaselineFacts(sources)
	return Baseline{FixtureID: fixtureID, Bump: *decoded.Bump, Breaking: len(*decoded.Breaking), Additions: len(*decoded.Additions), OutputParsed: true, Sources: sources, Capabilities: BaselineCapabilities{Facts: facts}}, nil
}

func rejectTrailingJSON(decoder *json.Decoder) error {
	var trailing json.RawMessage
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err == nil {
			return fmt.Errorf("trailing JSON")
		}
		return err
	}
	return nil
}

func baselineSources(groups ...[]baselineEntry) ([]BaselineSource, error) {
	sections := []string{"breaking", "additions", "notes"}
	var sources []BaselineSource
	for groupIndex, entries := range groups {
		for index, entry := range entries {
			if entry.Kind == nil || entry.Signature == nil || entry.Message == nil || *entry.Kind == "" || *entry.Signature == "" || *entry.Message == "" {
				return nil, fmt.Errorf("%s[%d]: missing kind, signature, or message", sections[groupIndex], index)
			}
			if *entry.Kind != "function" && *entry.Kind != "event" && *entry.Kind != "error" {
				return nil, fmt.Errorf("%s[%d]: invalid kind %q", sections[groupIndex], index, *entry.Kind)
			}
			sources = append(sources, BaselineSource{Section: sections[groupIndex], Index: index, Kind: *entry.Kind, Signature: *entry.Signature, Message: *entry.Message})
		}
	}
	return sources, nil
}

func deriveBaselineFacts(sources []BaselineSource) []BaselineFact {
	var facts []BaselineFact
	for sourceIndex, source := range sources {
		if source.Kind == "function" && (outputChangePattern.MatchString(source.Message) || mutabilityChangePattern.MatchString(source.Message)) {
			facts = appendFact(facts, "call_identity_unchanged", sourceIndex)
		}
		if source.Kind == "event" && source.Message == "event `indexed` layout changed — log decoding will break" {
			for _, name := range []string{"topic_identity", "topic_layout_impact", "data_layout_impact", "cross_decode_impact"} {
				facts = appendFact(facts, name, sourceIndex)
			}
		}
	}
	return facts
}

func appendFact(facts []BaselineFact, name string, source int) []BaselineFact {
	for _, fact := range facts {
		if fact.Name == name {
			return facts
		}
	}
	return append(facts, BaselineFact{Name: name, Value: true, Source: source})
}
