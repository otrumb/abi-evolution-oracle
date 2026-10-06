package experiment

import (
	"path/filepath"
	"strings"
	"testing"

	ethabi "github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/stretchr/testify/require"
)

func Test_Run_executes_all_mandatory_fixtures(t *testing.T) {
	// Given
	root := filepath.Join("..", "..")
	tool := filepath.Join(t.TempDir(), "missing.js")
	// When
	_, _, err := Run(root, t.TempDir(), tool, tool)
	// Then
	require.Error(t, err)
}

func Test_sampleOutput_packs_composite_outputs(t *testing.T) {
	// Given
	parsed, err := ethabi.JSON(strings.NewReader(`[{"type":"function","name":"probe","inputs":[],"outputs":[{"name":"items","type":"tuple[]","components":[{"name":"value","type":"uint256"}]}]}]`))
	require.NoError(t, err)
	// When
	data := sampleOutput(parsed.Methods["probe"].Outputs)
	// Then
	require.NotEmpty(t, data)
}
