package main

import (
	"context"
	"encoding/json"
	"log"
	"my-samples-go/temporal/orchestrator"

	"go.temporal.io/sdk/client"
)

func main() {
	c, err := client.Dial(client.Options{
		HostPort: "passthrough:///localhost:7233",
	})
	if err != nil {
		log.Fatalln("Unable to create Temporal client", err)
	}
	defer c.Close()

	workflowID := orchestrator.OrchestratorWorkflowID
	ctx := context.Background()

	// Query the workflow for registered items
	resp, err := c.QueryWorkflow(ctx, workflowID, "", orchestrator.QueryName)
	if err != nil {
		log.Fatalln("Unable to query workflow", err)
	}

	var queryResult orchestrator.QueryResponse
	if err := resp.Get(&queryResult); err != nil {
		log.Fatalln("Unable to decode query result", err)
	}

	log.Printf("Workflow State:\n")
	log.Printf("  Total Items: %d\n", queryResult.TotalItems)
	log.Printf("  Signals Handled: %d\n", queryResult.SignalsHandled)

	if queryResult.TotalItems > 0 {
		log.Printf("  Orchestrated Items:\n")
		for id, item := range queryResult.OrchestratedItems {
			itemJSON, _ := json.MarshalIndent(item, "    ", "  ")
			log.Printf("    %s: %s\n", id, string(itemJSON))
		}
	} else {
		log.Printf("  No items currently registered\n")
	}
}
