package main

import (
	"context"
	"fmt"
	"log"
	"my-samples-go/temporal/orchestrator"
	"os"

	"github.com/google/uuid"
	"go.temporal.io/sdk/client"
)

func main() {
	ctx := context.Background()

	if len(os.Args) < 3 {
		log.Fatalln("An item type and ID must be provided")
	}
	itemType := os.Args[1]
	itemID := os.Args[2]

	c, err := client.Dial(client.Options{
		HostPort: "passthrough:///localhost:7233",
	})
	if err != nil {
		log.Fatalln("Unable to create Temporal client", err)
	}
	defer c.Close()

	// Ensure the orchestrator is running
	ensureOrchestratorRunning(ctx, c)

	// Start the ItemWorkflow
	workflowID := "item_" + itemID + "_" + uuid.New().String()
	options := client.StartWorkflowOptions{
		ID:        workflowID,
		TaskQueue: "orchestrator-task-queue",
	}

	var item orchestrator.Item
	var workflowName string
	switch itemType {
	case "a", "A":
		workflowName = orchestrator.ItemWorkflowAName
		item = &orchestrator.ItemA{
			BasicItem: orchestrator.BasicItem{
				Id:   itemID,
				Name: fmt.Sprintf("Item-A-%s", itemID),
			},
			ExtraFieldA: "Extra data for Item A",
		}
	case "b", "B":
		workflowName = orchestrator.ItemWorkflowBName
		item = &orchestrator.ItemB{
			BasicItem: orchestrator.BasicItem{
				Id:   itemID,
				Name: fmt.Sprintf("Item-B-%s", itemID),
			},
			ExtraFieldB: "Extra data for Item B",
		}
	default:
		workflowName = orchestrator.ItemWorkflowAName // Default to A if unknown type
		item = &orchestrator.BasicItem{
			Id:     itemID,
			Name:   fmt.Sprintf("Item-%s", itemID),
			Status: "New",
		}
	}

	log.Printf("Starting ItemWorkflow with ID '%s' for item '%v'\n", workflowID, item)
	we, err := c.ExecuteWorkflow(ctx, options, workflowName, item)
	if err != nil {
		log.Fatalln("Unable to execute workflow", err)
	}

	log.Printf("Workflow started. WorkflowID: %s, RunID: %s\n", we.GetID(), we.GetRunID())

	// Wait for the workflow to complete and print the result.
	var result string
	err = we.Get(ctx, &result)
	if err != nil {
		log.Fatalln("Workflow failed:", err)
	}
	fmt.Printf("Workflow for item '%s' completed with result: %s\n", itemID, result)
}

func ensureOrchestratorRunning(ctx context.Context, c client.Client) {
	// Use SignalWithStart to start the workflow if it's not running, or signal it if it is.
	// This is a more robust way to ensure the singleton is running and ready.
	options := client.StartWorkflowOptions{
		ID:        orchestrator.OrchestratorWorkflowID,
		TaskQueue: "orchestrator-task-queue",
	}

	// Send a benign signal (e.g., "ping") to ensure the workflow is alive.
	// If the workflow is not running, it will be started.
	// The Get will block until the workflow is started and the first task is completed.
	_, err := c.SignalWithStartWorkflow(ctx,
		orchestrator.OrchestratorWorkflowID,
		orchestrator.SignalChannelName,
		orchestrator.Signal{Type: orchestrator.PingSignal},
		options,
		orchestrator.OrchestratorWorkflowName,
		orchestrator.OrchestratorState{})
	if err != nil {
		log.Fatalln("Unable to signal/start orchestrator workflow", err)
	} else {
		log.Println("Orchestrator workflow is running.")
	}
}
