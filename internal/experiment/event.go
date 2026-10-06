package experiment

import (
	"fmt"

	ethabi "github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/otrumb/abi-evolution-oracle/internal/corpus"
)

func eventObservation(entry corpus.Entry, oldABI, newABI ethabi.ABI) Observation {
	oldEvent, newEvent := oldABI.Events["Observed"], newABI.Events["Observed"]
	oldTopics, oldData := syntheticLog(oldEvent)
	newTopics, newData := syntheticLog(newEvent)
	oldNewErr := decodeEvent(oldEvent, newTopics, newData)
	newOldErr := decodeEvent(newEvent, oldTopics, oldData)
	filterEqual := len(oldTopics) > 0 && len(newTopics) > 0 && oldTopics[0] == newTopics[0]
	layoutChanged := len(oldTopics) != len(newTopics) || len(oldData) != len(newData) || !filterEqual || indexedLayoutDiffers(oldEvent, newEvent)
	consumerImpact := layoutChanged || oldNewErr != nil || newOldErr != nil
	status := "observed"
	if layoutChanged != entry.Expected["layout_changed"] || consumerImpact != entry.Expected["consumer_impact"] {
		status = "rejected"
	}
	detail := fmt.Sprintf("old_topics=%d;new_topics=%d;old_data=%d;new_data=%d;filter_equal=%t;old_new_error=%t;new_old_error=%t;layout_changed=%t;consumer_impact=%t", len(oldTopics), len(newTopics), len(oldData), len(newData), filterEqual, oldNewErr != nil, newOldErr != nil, layoutChanged, consumerImpact)
	return Observation{FixtureID: entry.ID, Class: entry.Class, Probe: "go-ethereum_v1.15.11_event", Status: status, Detail: detail, Assessment: scope(), Actionable: consumerImpact}
}

func indexedLayoutDiffers(oldEvent, newEvent ethabi.Event) bool {
	if oldEvent.Anonymous != newEvent.Anonymous || len(oldEvent.Inputs) != len(newEvent.Inputs) {
		return true
	}
	for index := range oldEvent.Inputs {
		if oldEvent.Inputs[index].Indexed != newEvent.Inputs[index].Indexed || oldEvent.Inputs[index].Type.String() != newEvent.Inputs[index].Type.String() {
			return true
		}
	}
	return false
}

func decodeEvent(event ethabi.Event, topics []common.Hash, data []byte) error {
	values := map[string]any{}
	var indexed ethabi.Arguments
	for _, input := range event.Inputs {
		if input.Indexed {
			indexed = append(indexed, input)
		}
	}
	topicValues := topics
	if !event.Anonymous {
		if len(topics) == 0 || topics[0] != event.ID {
			return fmt.Errorf("topic mismatch")
		}
		topicValues = topics[1:]
	}
	if len(topicValues) != len(indexed) {
		return fmt.Errorf("indexed topic count")
	}
	if err := ethabi.ParseTopicsIntoMap(values, indexed, topicValues); err != nil {
		return err
	}
	return event.Inputs.NonIndexed().UnpackIntoMap(values, data)
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
