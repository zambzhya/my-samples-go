package orchestrator

import (
	"encoding/json"
	"fmt"
)

func ConvertPayload(payloadData interface{}, target interface{}) error {
	payloadBytes, err := json.Marshal(payloadData)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}
	if err := json.Unmarshal(payloadBytes, target); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}
	return nil
}
