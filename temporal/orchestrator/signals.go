package orchestrator

// SignalType represents predefined signal types for the orchestrator workflow
type SignalType string

const (
	SignalChannelName     = "orchestrator-signal-channel" // channel for orchestrator workflow to receive signals from item workflows and starter
	ItemSignalChannelName = "item-signal-channel"         // channel for item workflow to receive "go/no-go" instruction

	RegisterSignal        SignalType = "register"         // register item
	DeregisterSignal      SignalType = "deregister"       // de-register item
	StartProcessingSignal SignalType = "start-processing" // request permission to start processing
	StopProcessingSignal  SignalType = "stop-processing"  // stop processing
	UpdateSignal          SignalType = "update"           // update item
	PingSignal            SignalType = "ping"             // optional, for illustrative purpose of "start-and-signal-workflow"
)

// Signal represents a signal with type and generic payload for the orchestrator workflow
type Signal struct {
	Type    SignalType  `json:"type"`
	Payload interface{} `json:"payload"`
}

// ItemInstructionSignal represents an instruction signal sent to an item workflow
type ItemInstructionSignal struct {
	ID      string `json:"id"`
	Proceed bool   `json:"proceed"`
	Reason  string `json:"reason"`
}

// Example payload implementations
type RegisterPayload struct {
	ID                string      `json:"id"`
	ItemWorkflowID    string      `json:"itemWorkflowId"`
	ItemWorkflowRunID string      `json:"itemWorkflowRunId"`
	Item              interface{} `json:"item,omitempty"`
}

type DeregisterPayload struct {
	ID string `json:"id"`
}

type StartProcessingPayload struct {
	ID string `json:"id"`
}

type StopProcessingPayload struct {
	ID string `json:"id"`
}

type UpdatePayload struct {
	Item interface{} `json:"item,omitempty"`
}
