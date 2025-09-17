package main

import (
	"log"
	"my-samples-go/temporal/orchestrator"
	"time"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"
)

const (
	SignalTimeout = 120 * time.Second
)

// OrchestratorWorkflow is the main workflow for the orchestrator.
func OrchestratorWorkflow(ctx workflow.Context, state orchestrator.OrchestratorState) error {
	signalCh := workflow.GetSignalChannel(ctx, orchestrator.SignalChannelName)
	logger := workflow.GetLogger(ctx)

	logger.Info("Orchestrator workflow started", "signalsHandled", state.SignalsHandled, "orchestratedItems", len(state.Orchestrator.AllItems()))

	// Set up query handler
	err := workflow.SetQueryHandler(ctx, orchestrator.QueryName, func() (orchestrator.QueryResponse, error) {
		return buildQueryResponse(state), nil
	})
	if err != nil {
		logger.Error("Failed to set query handler", "error", err)
		return err
	}

	for {
		timerCtx, cancelTimer := workflow.WithCancel(ctx)
		timerFuture := workflow.NewTimer(timerCtx, SignalTimeout)

		selector := workflow.NewSelector(ctx)
		selector.AddFuture(timerFuture, func(f workflow.Future) {
			// This callback executes when the timer fires.
		})

		var sig orchestrator.Signal
		selector.AddReceive(signalCh, func(c workflow.ReceiveChannel, more bool) {
			c.Receive(ctx, &sig)
			cancelTimer() // Cancel the timer because a signal was received.
		})

		selector.Select(ctx) // Wait for a signal or the timer.

		if sig.Type != "" { // A signal was received
			state.HandleSignal(logger, sig, ctx)
		} else { // The timer fired
			if len(state.Orchestrator.RegisteredItems()) == 0 {
				logger.Info("Timer fired and no registered items, finishing workflow.")
				return nil
			}
			logger.Info("Timer fired, but registered items exist. Resetting timer.", "registeredCount", len(state.Orchestrator.RegisteredItems()))
		}

		// Check for ContinueAsNew after processing the signal or timer.
		if workflow.GetInfo(ctx).GetContinueAsNewSuggested() {
			logger.Info("Continuing as new due to Temporal suggestion", "orchestratedItems", len(state.Orchestrator.AllItems()))
			return workflow.NewContinueAsNewError(ctx, OrchestratorWorkflow, state)
		}
	}
}

func buildQueryResponse(state orchestrator.OrchestratorState) orchestrator.QueryResponse {
	return orchestrator.QueryResponse{
		TotalItems:        len(state.Orchestrator.AllItems()),
		OrchestratedItems: state.Orchestrator.AllItems(),
		SignalsHandled:    state.SignalsHandled,
	}
}

func main() {
	c, err := client.Dial(client.Options{
		HostPort: "passthrough:///localhost:7233",
	})
	if err != nil {
		log.Fatalln("Unable to create Temporal client", err)
	}
	defer c.Close()

	w := worker.New(c, "orchestrator-task-queue", worker.Options{})
	w.RegisterWorkflowWithOptions(OrchestratorWorkflow, workflow.RegisterOptions{Name: orchestrator.OrchestratorWorkflowName})
	w.RegisterWorkflowWithOptions(ItemWorkflowA, workflow.RegisterOptions{Name: orchestrator.ItemWorkflowAName})
	w.RegisterWorkflowWithOptions(ItemWorkflowB, workflow.RegisterOptions{Name: orchestrator.ItemWorkflowBName})

	log.Println("Starting Orchestrator and Item Workflow worker...")
	if err := w.Run(worker.InterruptCh()); err != nil {
		log.Fatalln("Unable to start worker", err)
	}
}
