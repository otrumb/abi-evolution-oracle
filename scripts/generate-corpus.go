//go:build ignore

package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

type fixture struct {
	id, class, scenario, oldABI, newABI string
	expected                            map[string]bool
}
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
type probe struct {
	ID         string          `json:"id"`
	Scenario   string          `json:"scenario"`
	Expected   map[string]bool `json:"expected_observations"`
	Assessment assessment      `json:"assessment"`
}
type assessment struct {
	Runtime    string `json:"runtime_behavior"`
	Storage    string `json:"storage_compatibility"`
	Security   string `json:"security"`
	Deployment string `json:"deployment_compatibility"`
}

func main() {
	fixtures := append(append(append(append(directionalFixtures(), nameFixtures()...), eventFixtures()...), collisionFixtures()...), structuralFixtures()...)
	manifests := map[string][]entry{}
	for _, item := range fixtures {
		dir := filepath.Join("corpus", item.class, item.id)
		must(os.MkdirAll(dir, 0755))
		probeData, err := json.MarshalIndent(probe{item.id, item.scenario, item.expected, assessment{"not_assessed", "not_assessed", "not_assessed", "unknown"}}, "", "  ")
		must(err)
		oldData, newData, probeData := []byte(item.oldABI+"\n"), []byte(item.newABI+"\n"), append(probeData, '\n')
		must(os.WriteFile(filepath.Join(dir, "old.json"), oldData, 0644))
		must(os.WriteFile(filepath.Join(dir, "new.json"), newData, 0644))
		must(os.WriteFile(filepath.Join(dir, "probe.json"), probeData, 0644))
		base := filepath.ToSlash(dir)
		manifests[item.class] = append(manifests[item.class], entry{item.id, item.class, base + "/old.json", base + "/new.json", base + "/probe.json", "CC0-1.0", "original", item.expected, hashes{digest(oldData), digest(newData), digest(probeData)}})
	}
	must(os.MkdirAll(filepath.Join("corpus", "manifests"), 0755))
	for class, records := range manifests {
		sort.Slice(records, func(i, j int) bool { return records[i].ID < records[j].ID })
		var data []byte
		for _, record := range records {
			line, err := json.Marshal(record)
			must(err)
			data = append(data, append(line, '\n')...)
		}
		must(os.WriteFile(filepath.Join("corpus", "manifests", class+".jsonl"), data, 0644))
	}
}

func directionalFixtures() []fixture {
	pairs := [][3]string{{"uint256", "address", "scalar_word_reinterpretation"}, {"uint256", "bool", "bool_validation"}, {"bytes32", "uint256", "fixed_word_reinterpretation"}, {"uint256", "uint256,address", "output_added"}, {"uint256,address", "uint256", "output_removed"}, {"uint256", "bytes", "static_to_dynamic"}, {"bytes", "string", "dynamic_type_change"}, {"int256", "uint256", "signedness_change"}, {"uint256[2]", "uint256[3]", "fixed_array_length"}, {"tuple(uint256)", "tuple(address)", "tuple_member_type"}, {"tuple(uint256,tuple(address))", "tuple(uint256,tuple(bytes32))", "nested_tuple_shape"}, {"", "uint256", "output_introduced"}}
	result := make([]fixture, 0, len(pairs))
	for index, pair := range pairs {
		result = append(result, fixture{fmt.Sprintf("D%02d", index+1), "directional", pair[2], function("probe", "[]", outputs(pair[0])), function("probe", "[]", outputs(pair[1])), expected("selector_invariant", "decode_both_directions", "consumer_observation")})
	}
	return result
}

func nameFixtures() []fixture {
	tuple := func(field string) string {
		return fmt.Sprintf(`{"name":"result","type":"tuple","components":[{"name":"%s","type":"uint256"}]}`, field)
	}
	nested := func(field string) string {
		return fmt.Sprintf(`{"name":"result","type":"tuple","components":[{"name":"inner","type":"tuple","components":[{"name":"%s","type":"uint256"}]}]}`, field)
	}
	array := func(field string) string {
		return fmt.Sprintf(`{"name":"result","type":"tuple[]","components":[{"name":"%s","type":"uint256"}]}`, field)
	}
	return []fixture{
		{"N01", "names", "single_named_output", function("probe", "[]", `[{"name":"before","type":"uint256"}]`), function("probe", "[]", `[{"name":"after","type":"uint256"}]`), nameExpected(false)},
		{"N02", "names", "two_named_outputs", function("probe", "[]", `[{"name":"left","type":"uint256"},{"name":"right","type":"address"}]`), function("probe", "[]", `[{"name":"first","type":"uint256"},{"name":"second","type":"address"}]`), nameExpected(false)},
		{"N03", "names", "output_tuple_member", function("probe", "[]", "["+tuple("before")+"]"), function("probe", "[]", "["+tuple("after")+"]"), nameExpected(true)},
		{"N04", "names", "nested_output_tuple_member", function("probe", "[]", "["+nested("before")+"]"), function("probe", "[]", "["+nested("after")+"]"), nameExpected(true)},
		{"N05", "names", "output_tuple_array_member", function("probe", "[]", "["+array("before")+"]"), function("probe", "[]", "["+array("after")+"]"), nameExpected(true)},
		{"N06", "names", "input_tuple_member", function("probe", "["+tuple("before")+"]", "[]"), function("probe", "["+tuple("after")+"]", "[]"), nameExpected(true)},
		{"N07", "names", "nested_input_tuple_member", function("probe", "["+nested("before")+"]", "[]"), function("probe", "["+nested("after")+"]", "[]"), nameExpected(true)},
		{"N08", "names", "generated_type_name_collision_after_sanitization", function("probe", `[ {"name":"foo_bar","type":"tuple","components":[{"name":"x","type":"uint256"}]} ]`, "[]"), function("probe", `[ {"name":"fooBar","type":"tuple","components":[{"name":"x","type":"uint256"}]} ]`, "[]"), nameExpected(true)},
		{"N09", "names", "acronym_case_normalization", function("probe", `[{"name":"urlValue","type":"uint256"}]`, "[]"), function("probe", `[{"name":"URLValue","type":"uint256"}]`, "[]"), nameExpected(false)},
		{"N10", "names", "unnamed_becomes_named", function("probe", "[]", `[{"name":"","type":"uint256"}]`), function("probe", "[]", `[{"name":"value","type":"uint256"}]`), nameExpected(false)},
		{"N11", "names", "event_argument_name", event("Observed", false, arg("before", "uint256", false)), event("Observed", false, arg("after", "uint256", false)), nameExpected(true)},
		{"N12", "names", "custom_error_argument_name", customError("Failure", "before", "uint256"), customError("Failure", "after", "uint256"), nameExpected(true)},
	}
}

func eventFixtures() []fixture {
	return []fixture{
		{"E01", "events", "scalar_nonindexed_to_indexed", event("Observed", false, arg("value", "uint256", false)), event("Observed", false, arg("value", "uint256", true)), eventExpected()},
		{"E02", "events", "scalar_indexed_to_nonindexed", event("Observed", false, arg("value", "address", true)), event("Observed", false, arg("value", "address", false)), eventExpected()},
		{"E03", "events", "second_argument_to_indexed", event("Observed", false, arg("first", "address", true)+","+arg("second", "uint256", false)), event("Observed", false, arg("first", "address", true)+","+arg("second", "uint256", true)), eventExpected()},
		{"E04", "events", "indexed_argument_order", event("Observed", false, arg("first", "address", true)+","+arg("second", "uint256", true)), event("Observed", false, arg("second", "uint256", true)+","+arg("first", "address", true)), eventExpected()},
		{"E05", "events", "dynamic_string_becomes_indexed", event("Observed", false, arg("text", "string", false)), event("Observed", false, arg("text", "string", true)), eventExpected()},
		{"E06", "events", "indexed_dynamic_bytes_becomes_nonindexed", event("Observed", false, arg("data", "bytes", true)), event("Observed", false, arg("data", "bytes", false)), eventExpected()},
		{"E07", "events", "nonanonymous_to_anonymous", event("Observed", false, arg("value", "uint256", true)), event("Observed", true, arg("value", "uint256", true)), eventExpected()},
		{"E08", "events", "anonymous_to_nonanonymous", event("Observed", true, arg("value", "address", true)), event("Observed", false, arg("value", "address", true)), eventExpected()},
		{"E09", "events", "anonymous_one_indexed", event("Observed", true, arg("value", "uint256", false)), event("Observed", true, arg("value", "uint256", true)), eventExpected()},
		{"E10", "events", "anonymous_four_indexed", event("Observed", true, arg("a", "uint256", true)+","+arg("b", "uint256", true)+","+arg("c", "uint256", false)+","+arg("d", "uint256", false)), event("Observed", true, arg("a", "uint256", true)+","+arg("b", "uint256", true)+","+arg("c", "uint256", true)+","+arg("d", "uint256", true)), eventExpected()},
		{"E11", "events", "indexed_and_anonymous_together", event("Observed", false, arg("value", "bytes32", false)), event("Observed", true, arg("value", "bytes32", true)), eventExpected()},
		{"E12", "events", "same_signature_incompatible_layout", event("Observed", false, arg("first", "uint256", true)+","+arg("second", "address", false)), event("Observed", false, arg("first", "uint256", false)+","+arg("second", "address", true)), eventExpected()},
	}
}

func collisionFixtures() []fixture {
	return []fixture{
		{"C01", "collisions", "known_selector_collision", function("burn", `[{"name":"value","type":"uint256"}]`, "[]"), function("collate_propagate_storage", `[{"name":"value","type":"bytes16"}]`, "[]"), collisionExpected()},
		{"C02", "collisions", "same_signature_different_outputs", function("probe", "[]", `[{"name":"value","type":"uint256"}]`), function("probe", "[]", `[{"name":"value","type":"bytes32"}]`), collisionExpected()},
		{"C03", "collisions", "overloaded_function_name", function("probe", `[{"name":"value","type":"uint256"}]`, "[]"), `[{"type":"function","name":"probe","stateMutability":"view","inputs":[{"name":"value","type":"uint256"}],"outputs":[]},{"type":"function","name":"probe","stateMutability":"view","inputs":[{"name":"value","type":"address"}],"outputs":[]}]`, collisionExpected()},
		{"C04", "collisions", "function_custom_error_selector", function("Shared", `[{"name":"value","type":"uint256"}]`, "[]"), customError("Shared", "value", "uint256"), collisionExpected()},
		{"C05", "collisions", "duplicate_event_signature_indexed_layout", event("Observed", false, arg("value", "uint256", true)), event("Observed", false, arg("value", "uint256", false)), collisionExpected()},
		{"C06", "collisions", "anonymous_indistinguishable_layout", event("Alpha", true, arg("value", "uint256", true)), event("Beta", true, arg("value", "uint256", true)), collisionExpected()},
		{"C07", "collisions", "generated_identifier_normalization", function("foo_bar", "[]", "[]"), function("fooBar", "[]", "[]"), collisionExpected()},
		{"C08", "collisions", "tuple_alias_canonical_signature", functionInternal("probe", "struct Alpha.Payload"), functionInternal("probe", "struct Beta.Payload"), collisionExpected()},
	}
}

func structuralFixtures() []fixture {
	return []fixture{
		{"S01", "structural", "function_added", "[]", function("probe", "[]", "[]"), expected("baseline_addition", "semver_minor", "semantic_change")},
		{"S02", "structural", "function_removed", function("probe", "[]", "[]"), "[]", expected("baseline_breaking", "semver_major", "semantic_change")},
		{"S03", "structural", "input_type_changed", function("set", `[{"name":"value","type":"uint256"}]`, "[]"), function("set", `[{"name":"value","type":"address"}]`, "[]"), expected("baseline_breaking", "semver_major", "semantic_change")},
		{"S04", "structural", "payable_to_nonpayable", payable("pay", "payable"), payable("pay", "nonpayable"), expected("baseline_breaking", "semver_major", "semantic_change")},
		{"S05", "structural", "tuple_array_canonical_change", function("probe", "[]", `[{"name":"values","type":"uint256[2]"}]`), function("probe", "[]", `[{"name":"values","type":"uint256[3]"}]`), expected("baseline_breaking", "semver_major", "semantic_change")},
		{"S06", "structural", "json_rewrapped_semantically_equal", function("probe", "[]", `[{"name":"value","type":"uint256"}]`), `{"abi":` + function("probe", "[]", `[{"name":"value","type":"uint256"}]`) + `}`, expected("baseline_unchanged", "semver_none", "semantic_equal")},
	}
}

func expected(keys ...string) map[string]bool {
	result := make(map[string]bool, len(keys))
	for _, key := range keys {
		result[key] = true
	}
	return result
}
func nameExpected(compileBreak bool) map[string]bool {
	result := expected("wire_invariant", "signature_invariant", "generated_source_change")
	result["old_consumer_compile_break"] = compileBreak
	return result
}
func eventExpected() map[string]bool {
	return expected("synthetic_log", "cross_version_filter", "cross_version_decode")
}
func collisionExpected() map[string]bool {
	return expected("candidate_set_preserved", "ambiguity_expected", "no_arbitrary_winner")
}
func function(name, inputs, outputs string) string {
	return fmt.Sprintf(`[{"type":"function","name":"%s","stateMutability":"view","inputs":%s,"outputs":%s}]`, name, inputs, outputs)
}
func functionInternal(name, internal string) string {
	return fmt.Sprintf(`[{"type":"function","name":"%s","stateMutability":"view","inputs":[{"name":"value","type":"tuple","internalType":"%s","components":[{"name":"x","type":"uint256"}]}],"outputs":[]}]`, name, internal)
}
func payable(name, mutability string) string {
	return fmt.Sprintf(`[{"type":"function","name":"%s","stateMutability":"%s","inputs":[],"outputs":[]}]`, name, mutability)
}
func event(name string, anonymous bool, args string) string {
	return fmt.Sprintf(`[{"type":"event","name":"%s","anonymous":%t,"inputs":[%s]}]`, name, anonymous, args)
}
func customError(name, field, kind string) string {
	return fmt.Sprintf(`[{"type":"error","name":"%s","inputs":[{"name":"%s","type":"%s"}]}]`, name, field, kind)
}
func arg(name, kind string, indexed bool) string {
	return fmt.Sprintf(`{"name":"%s","type":"%s","indexed":%t}`, name, kind, indexed)
}
func outputs(kind string) string {
	switch kind {
	case "":
		return "[]"
	case "tuple(uint256)":
		return `[{"name":"value","type":"tuple","components":[{"name":"a","type":"uint256"}]}]`
	case "tuple(address)":
		return `[{"name":"value","type":"tuple","components":[{"name":"a","type":"address"}]}]`
	case "tuple(uint256,tuple(address))":
		return `[{"name":"value","type":"tuple","components":[{"name":"a","type":"uint256"},{"name":"inner","type":"tuple","components":[{"name":"b","type":"address"}]}]}]`
	case "tuple(uint256,tuple(bytes32))":
		return `[{"name":"value","type":"tuple","components":[{"name":"a","type":"uint256"},{"name":"inner","type":"tuple","components":[{"name":"b","type":"bytes32"}]}]}]`
	default:
		return fmt.Sprintf(`[{"name":"value","type":"%s"}]`, kind)
	}
}
func digest(data []byte) string { sum := sha256.Sum256(data); return hex.EncodeToString(sum[:]) }
func must(err error) {
	if err != nil {
		panic(err)
	}
}
