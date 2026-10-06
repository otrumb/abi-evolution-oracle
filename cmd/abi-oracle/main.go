package main

import (
	"fmt"
	"os"

	"local/abi-evolution-oracle-validation/internal/corpus"
)

func main() {
	if len(os.Args) != 2 || os.Args[1] != "verify-corpus" {
		fmt.Fprintln(os.Stderr, "usage: abi-oracle verify-corpus")
		os.Exit(2)
	}
	summary, err := corpus.Verify(".")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Printf("corpus verified: total=%d directional=%d names=%d events=%d collisions=%d structural=%d\n", summary.Total, summary.Directional, summary.Names, summary.Events, summary.Collisions, summary.Structural)
}
