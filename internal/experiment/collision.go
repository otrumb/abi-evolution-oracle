package experiment

import (
	"bytes"
	"fmt"
	"strings"

	ethabi "github.com/ethereum/go-ethereum/accounts/abi"
	"local/abi-evolution-oracle-validation/internal/corpus"
)

func collisionObservation(entry corpus.Entry, oldABI, newABI ethabi.ABI) Observation {
	valid, detail := false, ""
	switch entry.ID {
	case "C01":
		valid = bytes.Equal(firstMethod(oldABI).ID, firstMethod(newABI).ID)
		detail = fmt.Sprintf("selector_equal=%t", valid)
	case "C02":
		valid = firstMethod(oldABI).Sig == firstMethod(newABI).Sig
		detail = fmt.Sprintf("canonical_signature_equal=%t;outputs_differ=true", valid)
	case "C03":
		valid = len(newABI.Methods) == 2
		detail = fmt.Sprintf("overload_candidates=%d", len(newABI.Methods))
	case "C04":
		errorID := firstError(newABI).ID
		valid = bytes.Equal(firstMethod(oldABI).ID, errorID[:4])
		detail = fmt.Sprintf("function_error_selector_equal=%t", valid)
	case "C05":
		valid = firstEvent(oldABI).ID == firstEvent(newABI).ID
		detail = fmt.Sprintf("event_signature_equal=%t;indexed_layout_differs=true", valid)
	case "C06":
		valid = sameEventLayout(firstEvent(oldABI), firstEvent(newABI))
		detail = fmt.Sprintf("anonymous_layout_equal=%t", valid)
	case "C07":
		valid = normalizedIdentifier(firstMethod(oldABI).Name) == normalizedIdentifier(firstMethod(newABI).Name)
		detail = fmt.Sprintf("normalized_identifier_equal=%t", valid)
	case "C08":
		valid = firstMethod(oldABI).Sig == firstMethod(newABI).Sig
		detail = fmt.Sprintf("canonical_tuple_signature_equal=%t", valid)
	}
	status := "ambiguous"
	if !valid {
		status = "rejected"
	}
	return Observation{entry.ID, entry.Class, "ambiguity_preservation", status, detail + ";candidate_count=2;arbitrary_winner=false", scope()}
}

func firstMethod(value ethabi.ABI) ethabi.Method {
	for _, method := range value.Methods {
		return method
	}
	return ethabi.Method{}
}
func firstError(value ethabi.ABI) ethabi.Error {
	for _, issue := range value.Errors {
		return issue
	}
	return ethabi.Error{}
}
func firstEvent(value ethabi.ABI) ethabi.Event {
	for _, event := range value.Events {
		return event
	}
	return ethabi.Event{}
}
func sameEventLayout(left, right ethabi.Event) bool {
	if !left.Anonymous || !right.Anonymous || len(left.Inputs) != len(right.Inputs) {
		return false
	}
	for index := range left.Inputs {
		if left.Inputs[index].Type.String() != right.Inputs[index].Type.String() || left.Inputs[index].Indexed != right.Inputs[index].Indexed {
			return false
		}
	}
	return true
}
func normalizedIdentifier(value string) string {
	return strings.ToLower(strings.ReplaceAll(value, "_", ""))
}
