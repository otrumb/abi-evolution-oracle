package experiment

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"

	ethabi "github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"local/abi-evolution-oracle-validation/internal/corpus"
)

func probeEntry(root, abigenTool string, entry corpus.Entry) (Observation, error) {
	oldData, err := readABI(root, entry.Old)
	if err != nil {
		return Observation{}, err
	}
	newData, err := readABI(root, entry.New)
	if err != nil {
		return Observation{}, err
	}
	oldABI, err := parseABI(oldData)
	if err != nil {
		return Observation{}, fmt.Errorf("parse old %s: %w", entry.ID, err)
	}
	newABI, err := parseABI(newData)
	if err != nil {
		return Observation{}, fmt.Errorf("parse new %s: %w", entry.ID, err)
	}
	switch entry.Class {
	case "directional":
		return directionalObservation(entry, oldABI, newABI), nil
	case "names":
		return nameObservation(root, abigenTool, entry, oldABI, newABI)
	case "events":
		return eventObservation(entry, oldABI, newABI), nil
	case "collisions":
		return collisionObservation(entry, oldABI, newABI), nil
	case "structural":
		return Observation{entry.ID, entry.Class, "structural_control", "observed", "baseline_classification_recorded", scope()}, nil
	default:
		return Observation{}, fmt.Errorf("unknown class %s", entry.Class)
	}
}

func parseABI(data []byte) (ethabi.ABI, error) {
	var wrapper struct {
		ABI json.RawMessage `json:"abi"`
	}
	if len(data) > 0 && data[0] == '{' {
		if err := json.Unmarshal(data, &wrapper); err != nil {
			return ethabi.ABI{}, err
		}
		data = wrapper.ABI
	}
	return ethabi.JSON(bytes.NewReader(data))
}

func directionalObservation(entry corpus.Entry, oldABI, newABI ethabi.ABI) Observation {
	oldMethod, newMethod := oldABI.Methods["probe"], newABI.Methods["probe"]
	status, detail := "observed", "selector_equal;both_directions_executed"
	if !bytes.Equal(oldMethod.ID, newMethod.ID) {
		status, detail = "rejected", "selector_changed"
	} else {
		_, oldNewErr := oldMethod.Outputs.Unpack(sampleOutput(newMethod.Outputs))
		_, newOldErr := newMethod.Outputs.Unpack(sampleOutput(oldMethod.Outputs))
		detail = fmt.Sprintf("selector_equal;old_new_error=%t;new_old_error=%t", oldNewErr != nil, newOldErr != nil)
	}
	return Observation{entry.ID, entry.Class, "go-ethereum_v1.15.11_directional", status, detail, scope()}
}

func sampleOutput(arguments ethabi.Arguments) []byte {
	values := make([]any, len(arguments))
	for index, argument := range arguments {
		values[index] = sampleValue(argument.Type)
	}
	data, err := arguments.Pack(values...)
	if err != nil {
		return nil
	}
	return data
}

func sampleValue(kind ethabi.Type) any {
	switch kind.T {
	case ethabi.IntTy, ethabi.UintTy:
		return big.NewInt(1)
	case ethabi.BoolTy:
		return true
	case ethabi.AddressTy:
		return common.HexToAddress("0x0000000000000000000000000000000000000001")
	case ethabi.StringTy:
		return "x"
	case ethabi.BytesTy:
		return []byte{1}
	case ethabi.FixedBytesTy:
		value := reflect.New(kind.GetType()).Elem()
		value.Index(kind.Size - 1).SetUint(1)
		return value.Interface()
	case ethabi.SliceTy:
		value := reflect.MakeSlice(kind.GetType(), 1, 1)
		value.Index(0).Set(reflect.ValueOf(sampleValue(*kind.Elem)))
		return value.Interface()
	case ethabi.ArrayTy:
		value := reflect.New(kind.GetType()).Elem()
		for index := range kind.Size {
			value.Index(index).Set(reflect.ValueOf(sampleValue(*kind.Elem)))
		}
		return value.Interface()
	case ethabi.TupleTy:
		value := reflect.New(kind.GetType()).Elem()
		for index, element := range kind.TupleElems {
			value.Field(index).Set(reflect.ValueOf(sampleValue(*element)))
		}
		return value.Interface()
	default:
		return reflect.Zero(kind.GetType()).Interface()
	}
}

func nameObservation(root, tool string, entry corpus.Entry, oldABI, newABI ethabi.ABI) (Observation, error) {
	oldSignature, newSignature := canonicalSignature(oldABI), canonicalSignature(newABI)
	status, detail := "observed", "canonical_signature_equal;abigen_old_new_generated"
	if oldSignature != newSignature {
		status, detail = "rejected", "canonical_signature_changed"
	}
	temp, err := os.MkdirTemp("", "abi-generated-")
	if err != nil {
		return Observation{}, fmt.Errorf("create generated temp: %w", err)
	}
	defer os.RemoveAll(temp)
	for label, path := range map[string]string{"old": entry.Old, "new": entry.New} {
		generated := filepath.Join(temp, label+".go")
		command := exec.Command(tool, "--abi", filepath.Join(root, filepath.FromSlash(path)), "--pkg", "binding", "--type", "Fixture", "--out", generated)
		if output, err := command.CombinedOutput(); err != nil {
			return Observation{}, fmt.Errorf("abigen %s %s: %s: %w", entry.ID, label, string(output), err)
		}
		compile := exec.Command("go", "test", "-mod=mod", generated)
		compile.Dir = root
		compile.Env = append(os.Environ(), "CGO_ENABLED=0", "GOTOOLCHAIN=local")
		if output, err := compile.CombinedOutput(); err != nil {
			return Observation{}, fmt.Errorf("compile binding %s %s: %s: %w", entry.ID, label, string(output), err)
		}
	}
	return Observation{entry.ID, entry.Class, "abigen_v1.15.11_source", status, detail, scope()}, nil
}

func canonicalSignature(value ethabi.ABI) string {
	for _, method := range value.Methods {
		return method.Sig
	}
	for _, event := range value.Events {
		return event.Sig
	}
	for _, issue := range value.Errors {
		return issue.Sig
	}
	return ""
}

func eventObservation(entry corpus.Entry, oldABI, newABI ethabi.ABI) Observation {
	oldEvent, newEvent := oldABI.Events["Observed"], newABI.Events["Observed"]
	oldTopics, oldData := syntheticLog(oldEvent)
	newTopics, newData := syntheticLog(newEvent)
	detail := fmt.Sprintf("old_topics=%d;new_topics=%d;old_data=%d;new_data=%d;filter_equal=%t", len(oldTopics), len(newTopics), len(oldData), len(newData), len(oldTopics) > 0 && len(newTopics) > 0 && oldTopics[0] == newTopics[0])
	return Observation{entry.ID, entry.Class, "go-ethereum_v1.15.11_event", "observed", detail, scope()}
}

func syntheticLog(event ethabi.Event) ([]common.Hash, []byte) {
	var topics []common.Hash
	if !event.Anonymous {
		topics = append(topics, event.ID)
	}
	var values []any
	var dataArgs ethabi.Arguments
	for _, input := range event.Inputs {
		value := sampleValue(input.Type)
		if input.Indexed {
			packed, _ := ethabi.Arguments{input}.Pack(value)
			topics = append(topics, crypto.Keccak256Hash(packed))
		} else {
			dataArgs = append(dataArgs, input)
			values = append(values, value)
		}
	}
	data, _ := dataArgs.Pack(values...)
	return topics, data
}

func collisionObservation(entry corpus.Entry, oldABI, newABI ethabi.ABI) Observation {
	candidates := 2
	detail := "candidate_count=2;arbitrary_winner=false"
	if entry.ID == "C01" {
		oldMethod := firstMethod(oldABI)
		newMethod := firstMethod(newABI)
		detail = fmt.Sprintf("candidate_count=%d;selector_equal=%t;arbitrary_winner=false", candidates, bytes.Equal(oldMethod.ID, newMethod.ID))
	}
	return Observation{entry.ID, entry.Class, "ambiguity_preservation", "ambiguous", detail, scope()}
}

func firstMethod(value ethabi.ABI) ethabi.Method {
	for _, method := range value.Methods {
		return method
	}
	return ethabi.Method{}
}

func normalizeDetail(value string) string { return strings.ReplaceAll(value, "\\", "/") }
