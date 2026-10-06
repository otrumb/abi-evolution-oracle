package policy_test

import (
	"bufio"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

const moduleName = "github.com/otrumb/abi-evolution-oracle"

var actionPattern = regexp.MustCompile(`(?m)^\s*- uses: [^@\s]+@([0-9a-f]{40})(?:\s+# .+)?$`)
var usesPattern = regexp.MustCompile(`(?m)^\s*- uses:`)

func Test_PublicIdentity_matches_release_contract(t *testing.T) {
	root := repositoryRoot(t)
	goMod := readFile(t, filepath.Join(root, "go.mod"))
	if !strings.HasPrefix(goMod, "module "+moduleName+"\n") {
		t.Fatalf("module identity mismatch")
	}
	if _, err := os.Stat(filepath.Join(root, "cmd", "abi-evolution-oracle", "main.go")); err != nil {
		t.Fatalf("public command missing: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "cmd", "abi-oracle", "main.go")); !os.IsNotExist(err) {
		t.Fatalf("legacy command entrypoint exists")
	}
}

func Test_Readme_states_bounded_evidence_and_licenses(t *testing.T) {
	root := repositoryRoot(t)
	readme := readFile(t, filepath.Join(root, "README.md"))
	required := []string{
		"Reports consumer compatibility evidence only for tested corpus,",
		"Runtime behavior, storage compatibility,",
		"and security are not assessed. Deployment compatibility is unknown.",
		"Demand and\nadoption are unvalidated.",
		"universal\nsafe/breaking claims",
		"100/100 for this experiment",
		"not market validation",
	}
	for _, text := range required {
		if !strings.Contains(readme, text) {
			t.Errorf("README missing required scope text %q", text)
		}
	}
	if strings.Count(readme, "Project source is MIT licensed") != 1 {
		t.Fatal("README must contain one source MIT notice")
	}
	if strings.Count(readme, "separately dedicated under CC0-1.0") != 1 {
		t.Fatal("README must contain one corpus CC0 notice")
	}
}

func Test_Workflows_pin_actions_and_enforce_least_privilege(t *testing.T) {
	root := repositoryRoot(t)
	for _, name := range []string{"ci.yml", "release.yml"} {
		workflow := readFile(t, filepath.Join(root, ".github", "workflows", name))
		if strings.Contains(workflow, "pull_request_target:") {
			t.Fatalf("%s uses pull_request_target", name)
		}
		if uses := len(usesPattern.FindAllString(workflow, -1)); uses == 0 || uses != len(actionPattern.FindAllStringSubmatch(workflow, -1)) {
			t.Fatalf("%s contains missing or non-SHA action pins", name)
		}
		if !strings.Contains(workflow, "permissions:\n  contents: read") {
			t.Fatalf("%s lacks read-only workflow permission", name)
		}
	}

	release := readFile(t, filepath.Join(root, ".github", "workflows", "release.yml"))
	if strings.Count(release, "contents: write") != 1 {
		t.Fatal("release workflow must grant write once")
	}
	if !regexp.MustCompile(`(?s)release:\n.*?needs: \[validate-ubuntu, validate-windows, package\].*?permissions:\n\s+contents: write`).MatchString(release) {
		t.Fatal("final release must depend on both OS validations and package job")
	}
	if !strings.Contains(release, `gh release create "${GITHUB_REF_NAME}" --draft --verify-tag`) {
		t.Fatal("release creation must remain draft and verify tag")
	}
}

func Test_Release_assets_and_checksums_are_exact(t *testing.T) {
	root := repositoryRoot(t)
	script := readFile(t, filepath.Join(root, "scripts", "package-release.go"))
	release := readFile(t, filepath.Join(root, ".github", "workflows", "release.yml"))
	assets := []string{
		"abi-evolution-oracle_0.1.0_linux_amd64.tar.gz",
		"abi-evolution-oracle_0.1.0_windows_amd64.zip",
	}
	for _, asset := range assets {
		if strings.Count(script, asset) != 1 || strings.Count(release, asset) != 1 {
			t.Errorf("asset contract mismatch for %s", asset)
		}
	}
	if strings.Contains(script, `current.archive: "SHA256SUMS"`) || strings.Contains(script, `hash.Sum(nil), "SHA256SUMS"`) {
		t.Fatal("checksum manifest must not hash itself")
	}
	for _, required := range []string{"LICENSE", "README.md", "CGO_ENABLED=0", "GOARCH=amd64", "-trimpath", "v0.1.0"} {
		if !strings.Contains(script, required) {
			t.Errorf("packager missing %q", required)
		}
	}
}

func Test_Text_contracts_materialize_as_LF_with_autocrlf(t *testing.T) {
	root := repositoryRoot(t)
	paths := []string{
		"go.mod",
		"cmd/abi-evolution-oracle/main.go",
		"corpus/directional/D01/old.json",
		"evidence/observations.jsonl",
		"README.md",
		"LICENSE",
		"VERSION",
		".github/workflows/ci.yml",
	}
	arguments := append([]string{"-c", "core.autocrlf=true", "check-attr", "eol", "--"}, paths...)
	command := exec.Command("git", arguments...)
	command.Dir = root
	output, err := command.Output()
	if err != nil {
		t.Fatalf("git check-attr: %v", err)
	}
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		if !strings.HasSuffix(scanner.Text(), ": lf") {
			t.Errorf("non-LF text contract: %s", scanner.Text())
		}
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
}

func repositoryRoot(t *testing.T) string {
	t.Helper()
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	return root
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}
