//go:build ignore

package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type fixture struct{ id, class, change, oldABI, newABI string }
type hashes struct {
	Old   string `json:"old"`
	New   string `json:"new"`
	Probe string `json:"probe"`
}
type entry struct {
	ID       string          `json:"id"`
	Class    string          `json:"class"`
	Old      string          `json:"old"`
	New      string          `json:"new"`
	Probe    string          `json:"probe"`
	License  string          `json:"license"`
	Origin   string          `json:"origin"`
	Expected map[string]bool `json:"expected_observations"`
	Hashes   hashes          `json:"sha256"`
}

func function(outputs string) string {
	return fmt.Sprintf(`[{"type":"function","name":"probe","stateMutability":"view","inputs":[],"outputs":%s}]`, outputs)
}
func event(inputs string, anonymous bool) string {
	return fmt.Sprintf(`[{"type":"event","name":"Observed","anonymous":%t,"inputs":%s}]`, anonymous, inputs)
}
func arg(name, kind string, indexed bool) string {
	return fmt.Sprintf(`{"name":"%s","type":"%s","indexed":%t}`, name, kind, indexed)
}
func tuple(name, component string) string {
	return fmt.Sprintf(`[{"name":"%s","type":"tuple","components":[{"name":"%s","type":"uint256"}]}]`, name, component)
}

func main() {
	fixtures := buildFixtures()
	manifests := map[string][]byte{}
	for _, item := range fixtures {
		dir := filepath.Join("corpus", item.class, item.id)
		must(os.MkdirAll(dir, 0755))
		probe, _ := json.MarshalIndent(map[string]any{"id": item.id, "change": item.change, "assessment": map[string]string{"runtime_behavior": "not_assessed", "storage_compatibility": "not_assessed", "security": "not_assessed", "deployment_compatibility": "unknown"}}, "", "  ")
		oldData := []byte(item.oldABI + "\n")
		newData := []byte(item.newABI + "\n")
		probe = append(probe, '\n')
		must(os.WriteFile(filepath.Join(dir, "old.json"), oldData, 0644))
		must(os.WriteFile(filepath.Join(dir, "new.json"), newData, 0644))
		must(os.WriteFile(filepath.Join(dir, "probe.json"), probe, 0644))
		oldPath := filepath.ToSlash(filepath.Join("corpus", item.class, item.id, "old.json"))
		newPath := filepath.ToSlash(filepath.Join("corpus", item.class, item.id, "new.json"))
		probePath := filepath.ToSlash(filepath.Join("corpus", item.class, item.id, "probe.json"))
		record := entry{item.id, item.class, oldPath, newPath, probePath, "CC0-1.0", "original", map[string]bool{"executed": true}, hashes{digest(oldData), digest(newData), digest(probe)}}
		line, _ := json.Marshal(record)
		manifests[item.class] = append(manifests[item.class], append(line, '\n')...)
	}
	must(os.MkdirAll(filepath.Join("corpus", "manifests"), 0755))
	for class, data := range manifests {
		must(os.WriteFile(filepath.Join("corpus", "manifests", class+".jsonl"), data, 0644))
	}
}

func digest(data []byte) string { sum := sha256.Sum256(data); return hex.EncodeToString(sum[:]) }
func must(err error) {
	if err != nil {
		panic(err)
	}
}

func buildFixtures() []fixture {
	var result []fixture
	dTypes := [][2]string{{"uint256", "address"}, {"uint256", "bool"}, {"bytes32", "uint256"}, {"uint256", "uint256,address"}, {"uint256,address", "uint256"}, {"uint256", "bytes"}, {"bytes", "string"}, {"int256", "uint256"}, {"uint256[2]", "uint256[3]"}, {"(uint256)", "(address)"}, {"(uint256,(address))", "(uint256,(bytes32))"}, {"", "uint256"}}
	for i, pair := range dTypes {
		oldOut := outputs(pair[0])
		newOut := outputs(pair[1])
		result = append(result, fixture{fmt.Sprintf("D%02d", i+1), "directional", pair[0] + " to " + pair[1], function(oldOut), function(newOut)})
	}
	for i := 1; i <= 12; i++ {
		oldName := fmt.Sprintf("oldField%d", i)
		newName := fmt.Sprintf("newField%d", i)
		oldOut := fmt.Sprintf(`[{"name":"%s","type":"uint256"}]`, oldName)
		newOut := fmt.Sprintf(`[{"name":"%s","type":"uint256"}]`, newName)
		if i >= 3 && i <= 9 {
			oldOut = tuple("result", oldName)
			newOut = tuple("result", newName)
		}
		result = append(result, fixture{fmt.Sprintf("N%02d", i), "names", "generated identifier rename", function(oldOut), function(newOut)})
	}
	for i := 1; i <= 12; i++ {
		oldIndexed := i%2 == 0
		newIndexed := !oldIndexed
		oldAnon := i == 8 || i == 9 || i == 10
		newAnon := i == 7 || i == 9 || i == 10
		kind := "uint256"
		if i == 5 {
			kind = "string"
		}
		if i == 6 {
			kind = "bytes"
		}
		oldInputs := "[" + arg("value", kind, oldIndexed) + "]"
		newInputs := "[" + arg("value", kind, newIndexed) + "]"
		result = append(result, fixture{fmt.Sprintf("E%02d", i), "events", "event layout evolution", event(oldInputs, oldAnon), event(newInputs, newAnon)})
	}
	for i := 1; i <= 8; i++ {
		old := function(`[{"name":"value","type":"uint256"}]`)
		newer := function(`[{"name":"value","type":"bytes32"}]`)
		if i == 1 {
			old = `[{"type":"function","name":"burn","stateMutability":"nonpayable","inputs":[{"name":"value","type":"uint256"}],"outputs":[]}]`
			newer = `[{"type":"function","name":"collate_propagate_storage","stateMutability":"nonpayable","inputs":[{"name":"value","type":"bytes16"}],"outputs":[]}]`
		}
		result = append(result, fixture{fmt.Sprintf("C%02d", i), "collisions", "candidate ambiguity", old, newer})
	}
	structural := []fixture{{"S01", "structural", "function added", "[]", function("[]")}, {"S02", "structural", "function removed", function("[]"), "[]"}, {"S03", "structural", "input changed", `[{"type":"function","name":"set","stateMutability":"nonpayable","inputs":[{"name":"x","type":"uint256"}],"outputs":[]}]`, `[{"type":"function","name":"set","stateMutability":"nonpayable","inputs":[{"name":"x","type":"address"}],"outputs":[]}]`}, {"S04", "structural", "payable restriction", `[{"type":"function","name":"pay","stateMutability":"payable","inputs":[],"outputs":[]}]`, `[{"type":"function","name":"pay","stateMutability":"nonpayable","inputs":[],"outputs":[]}]`}, {"S05", "structural", "tuple array changed", function(`[{"name":"x","type":"uint256[2]"}]`), function(`[{"name":"x","type":"uint256[3]"}]`)}, {"S06", "structural", "semantic no change", function(`[{"name":"x","type":"uint256"}]`), "{\"abi\":" + function(`[{"name":"x","type":"uint256"}]`) + "}"}}
	return append(result, structural...)
}

func outputs(types string) string {
	if types == "" {
		return "[]"
	}
	parts := []byte("[")
	start := 0
	depth := 0
	index := 0
	for i := 0; i <= len(types); i++ {
		if i < len(types) {
			if types[i] == '(' {
				depth++
			}
			if types[i] == ')' {
				depth--
			}
		}
		if i == len(types) || (types[i] == ',' && depth == 0) {
			kind := types[start:i]
			if index > 0 {
				parts = append(parts, ',')
			}
			parts = append(parts, []byte(fmt.Sprintf(`{"name":"value%d","type":"%s"}`, index, kind))...)
			start = i + 1
			index++
		}
	}
	return string(append(parts, ']'))
}
