package orchestrator

// QueryResponse represents the response for the orchestrated items query
type QueryResponse struct {
	TotalItems        int                         `json:"totalItems"`
	OrchestratedItems map[string]OrchestratedItem `json:"orchestratedItems"`
	SignalsHandled    int                         `json:"signalsHandled"`
}
