package policy_test

import (
	"bufio"
	"encoding/json"
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
	if !strings.Contains(release, `--notes-file RELEASE_NOTES.md`) {
		t.Fatal("release creation must use reviewed release notes")
	}
	if !regexp.MustCompile(`(?s)release:\n.*?actions/checkout@[0-9a-f]{40}.*?ref: \$\{\{ github\.sha \}\}.*?--notes-file RELEASE_NOTES\.md`).MatchString(release) {
		t.Fatal("release job must check out the exact tagged source before reading release notes")
	}
	ci := readFile(t, filepath.Join(root, ".github", "workflows", "ci.yml"))
	if !strings.Contains(ci, "go install github.com/ethereum/go-ethereum/cmd/abigen@v1.15.11") {
		t.Fatal("evidence replay must install exact abigen version outside the project module graph")
	}
	if strings.Contains(ci, "go build -trimpath -o \"$RUNNER_TEMP/abigen\" github.com/ethereum/go-ethereum/cmd/abigen") {
		t.Fatal("evidence replay must not build abigen through project go.sum")
	}
}

func Test_Release_notes_preserve_scope_boundaries(t *testing.T) {
	root := repositoryRoot(t)
	notes := readFile(t, filepath.Join(root, "RELEASE_NOTES.md"))
	for _, required := range []string{
		"consumer compatibility evidence",
		"Runtime behavior, storage compatibility, and security were not assessed",
		"Deployment compatibility is unknown",
		"Demand and adoption are unvalidated",
	} {
		if !strings.Contains(notes, required) {
			t.Errorf("release notes missing %q", required)
		}
	}
	for _, forbidden := range []string{"universally compatible", "ABI safe", "non-breaking"} {
		if strings.Contains(notes, forbidden) {
			t.Errorf("release notes contain forbidden claim %q", forbidden)
		}
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

func Test_Publication_authorization_is_explicit_and_bounded(t *testing.T) {
	root := repositoryRoot(t)
	data := readFile(t, filepath.Join(root, "docs", "publication-authorization.json"))
	var receipt struct {
		Schema      int      `json:"schema"`
		State       string   `json:"state"`
		ClosureHead string   `json:"closure_head"`
		Allowed     []string `json:"allowed_public_mutations"`
		Forbidden   []string `json:"still_forbidden"`
	}
	if err := json.Unmarshal([]byte(data), &receipt); err != nil {
		t.Fatal(err)
	}
	if receipt.Schema != 1 || receipt.State != "user_authorized" {
		t.Fatalf("invalid publication authorization: schema=%d state=%q", receipt.Schema, receipt.State)
	}
	if receipt.ClosureHead != "c7e9ca7a0a0cede919f3036df27625bbed0e76ff" {
		t.Fatalf("unexpected closure head %q", receipt.ClosureHead)
	}
	wantAllowed := []string{"create_public_repository", "push_main", "push_v0.1.0_tag", "publish_v0.1.0_release"}
	if strings.Join(receipt.Allowed, "\n") != strings.Join(wantAllowed, "\n") {
		t.Fatalf("unexpected allowed mutations: %v", receipt.Allowed)
	}
	wantForbidden := []string{"contact_maintainers", "claim_adoption", "publish_packages", "rewrite_history"}
	if strings.Join(receipt.Forbidden, "\n") != strings.Join(wantForbidden, "\n") {
		t.Fatalf("unexpected forbidden mutations: %v", receipt.Forbidden)
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
