package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"local/abi-evolution-oracle-validation/internal/corpus"
	"local/abi-evolution-oracle-validation/internal/experiment"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: abi-oracle verify-corpus|run")
		os.Exit(2)
	}
	switch os.Args[1] {
	case "verify-corpus":
		summary, err := corpus.Verify(".")
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Printf("corpus verified: total=%d directional=%d names=%d events=%d collisions=%d structural=%d\n", summary.Total, summary.Directional, summary.Names, summary.Events, summary.Collisions, summary.Structural)
	case "run":
		if len(os.Args) != 5 {
			fmt.Fprintln(os.Stderr, "usage: abi-oracle run EVIDENCE_ROOT ABIDIFF_CLI ABIGEN")
			os.Exit(2)
		}
		summary, score, err := experiment.Run(".", filepath.Clean(os.Args[2]), filepath.Clean(os.Args[3]), filepath.Clean(os.Args[4]))
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Printf("probes: D=%d N=%d E=%d C=%d S=%d baseline=%d score=%d verdict=%s\n", summary.Directional, summary.Names, summary.Events, summary.Collisions, summary.Structural, summary.Baselines, score.Score, score.TechnicalVerdict)
	case "probe-one":
		if len(os.Args) != 4 {
			fmt.Fprintln(os.Stderr, "usage: abi-oracle probe-one FIXTURE_ID ABIGEN")
			os.Exit(2)
		}
		observation, err := experiment.ProbeOne(".", filepath.Clean(os.Args[3]), os.Args[2])
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		data, err := json.Marshal(observation)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Println(string(data))
	case "score":
		if len(os.Args) != 4 {
			fmt.Fprintln(os.Stderr, "usage: abi-oracle score EVIDENCE_ROOT GATE_JSON")
			os.Exit(2)
		}
		score, err := experiment.ScoreRecorded(".", filepath.Clean(os.Args[2]), filepath.Clean(os.Args[3]))
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Printf("score=%d verdict=%s wins=%d failures=%d\n", score.Score, score.TechnicalVerdict, len(score.ClassesBeatingBaseline), len(score.HardGateFailures))
	default:
		fmt.Fprintln(os.Stderr, "usage: abi-oracle verify-corpus|run")
		os.Exit(2)
	}
}
