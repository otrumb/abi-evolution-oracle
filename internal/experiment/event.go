package experiment

import (
	"fmt"

	ethabi "github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"local/abi-evolution-oracle-validation/internal/corpus"
)

func eventObservation(entry corpus.Entry, oldABI, newABI ethabi.ABI) Observation {
	oldEvent, newEvent := oldABI.Events["Observed"], newABI.Events["Observed"]
	oldTopics, oldData := syntheticLog(oldEvent)
	newTopics, newData := syntheticLog(newEvent)
	oldNewErr := decodeEvent(oldEvent, newTopics, newData)
	newOldErr := decodeEvent(newEvent, oldTopics, oldData)
	detail := fmt.Sprintf("old_topics=%d;new_topics=%d;old_data=%d;new_data=%d;filter_equal=%t;old_new_error=%t;new_old_error=%t", len(oldTopics), len(newTopics), len(oldData), len(newData), len(oldTopics) > 0 && len(newTopics) > 0 && oldTopics[0] == newTopics[0], oldNewErr != nil, newOldErr != nil)
	return Observation{entry.ID, entry.Class, "go-ethereum_v1.15.11_event", "observed", detail, scope()}
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
