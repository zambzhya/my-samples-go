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
	// IdleTimeout is the maximum duration the orchestrator workflow will wait without receiving any signals
	// before checking if it should continue (e.g. if there are still registered items to orchestrate) or finish.
	IdleTimeout = 120 * time.Second
)

type OW[O orchestrator.OrchestratorStateManager] struct {
	workflowAlias          string
	createStateManagerFunc orchestrator.CreateOrchestratorStateManagerFunc[O]
}

func NewOW[O orchestrator.OrchestratorStateManager](workflowAlias string, CreateStateManagerFunc orchestrator.CreateOrchestratorStateManagerFunc[O]) *OW[O] {
	return &OW[O]{
		workflowAlias:          workflowAlias,
		createStateManagerFunc: CreateStateManagerFunc,
	}
}

// OrchestratorWorkflow is the main workflow for the orchestrator.
func (ow *OW[O]) OrchestratorWorkflow(ctx workflow.Context, state orchestrator.OrchestratorState) error {
	signalCh := workflow.GetSignalChannel(ctx, orchestrator.SignalChannelName)
	logger := workflow.GetLogger(ctx)

	stateManager := ow.createStateManagerFunc(&state)
	logger.Info("Orchestrator workflow started", "signalsHandled", stateManager.GetState().GetSignalsHandled(), "orchestratedItems", len(stateManager.AllItems()))

	// Set up query handler
	err := workflow.SetQueryHandler(ctx, orchestrator.QueryName, func() (orchestrator.QueryResponse, error) {
		return ow.buildQueryResponse(stateManager), nil
	})
	if err != nil {
		logger.Error("Failed to set query handler", "error", err)
		return err
	}

	for {
		idleTimerCtx, cancelIdleTimer := workflow.WithCancel(ctx)
		idleTimerFuture := workflow.NewTimer(idleTimerCtx, IdleTimeout)

		selector := workflow.NewSelector(ctx)
		selector.AddFuture(idleTimerFuture, func(f workflow.Future) {
			// This callback executes when the timer fires.
		})

		var sig orchestrator.Signal
		selector.AddReceive(signalCh, func(c workflow.ReceiveChannel, more bool) {
			c.Receive(ctx, &sig)
			cancelIdleTimer() // Cancel the timer because a signal was received.
		})

		selector.Select(ctx) // Wait for a signal or the timer.

		if sig.Type != "" { // A signal was received
			ow.HandleSignal(ctx, stateManager, sig)
		} else { // The timer fired
			if len(stateManager.RegisteredItems()) == 0 {
				logger.Info("Timer fired and no registered items, finishing workflow.")
				return nil
			}
			logger.Info("Timer fired, but registered items exist. Resetting timer.", "registeredCount", len(stateManager.RegisteredItems()))
		}

		// Check for ContinueAsNew after processing the signal or timer.
		if workflow.GetInfo(ctx).GetContinueAsNewSuggested() {
			logger.Info("Continuing as new due to Temporal suggestion", "orchestratedItems", len(stateManager.AllItems()))
			state := orchestrator.OrchestratorState{}
			if stateManager.GetState() != nil {
				state = *stateManager.GetState()
			}
			return workflow.NewContinueAsNewError(ctx, ow.workflowAlias, state)
		}
	}
}

func (ow *OW[O]) HandleSignal(ctx workflow.Context, stateManager O, sig orchestrator.Signal) {
	logger := workflow.GetLogger(ctx)

	stateManager.GetState().SignalsHandled++
	switch sig.Type {
	case orchestrator.RegisterSignal:
		var p orchestrator.RegisterPayload
		if err := orchestrator.ConvertPayload(sig.Payload, &p); err != nil {
			logger.Error("Failed to convert register payload", "error", err)
			return
		}
		logger.Info("Handling register signal", "id", p.ID)

		canProceed := true
		reason := "Registration accepted."
		if err := stateManager.RegisterItem(p.ID, p.ItemWorkflowID, p.ItemWorkflowRunID, p.Item); err != nil {
			logger.Error("Failed to register item", "error", err)
			canProceed = false
			reason = "Registration denied: " + err.Error()
		}

		itemSignal := orchestrator.ItemInstructionSignal{ID: p.ID, Proceed: canProceed, Reason: reason}

		logger.Info("Sending signal to item workflow", "workflowID", p.ItemWorkflowID, "proceed", canProceed)
		err := workflow.SignalExternalWorkflow(ctx, p.ItemWorkflowID, p.ItemWorkflowRunID, orchestrator.ItemSignalChannelName, itemSignal).Get(ctx, nil)
		if err != nil {
			logger.Error("Failed to send signal to item workflow", "error", err, "itemWorkflowID", p.ItemWorkflowID)
			return
		}

	case orchestrator.StartProcessingSignal:
		var p orchestrator.StartProcessingPayload
		if err := orchestrator.ConvertPayload(sig.Payload, &p); err != nil {
			logger.Error("Failed to convert start-processing payload", "error", err)
			return
		}
		logger.Info("Handling start-processing request", "id", p.ID)

		canProceed := true
		reason := "Start processing permitted."
		item, err := stateManager.StartProcessing(p.ID)
		if err != nil {
			logger.Error("Failed to start processing", "error", err)
			canProceed = false
			reason = "Start processing denied: " + err.Error()
		}

		if item == nil {
			return
		}

		itemSignal := orchestrator.ItemInstructionSignal{ID: p.ID, Proceed: canProceed, Reason: reason}

		logger.Info("Sending signal to item workflow", "workflowID", item.ItemWorkflowID, "proceed", canProceed)
		err = workflow.SignalExternalWorkflow(ctx, item.ItemWorkflowID, item.ItemWorkflowRunID, orchestrator.ItemSignalChannelName, itemSignal).Get(ctx, nil)
		if err != nil {
			logger.Error("Failed to send signal to item workflow", "error", err, "itemWorkflowID", item.ItemWorkflowID)
			return
		}

	case orchestrator.StopProcessingSignal:
		var p orchestrator.StopProcessingPayload
		if err := orchestrator.ConvertPayload(sig.Payload, &p); err != nil {
			logger.Error("Failed to convert stop-processing payload", "error", err)
			return
		}
		logger.Info("Handling stop-processing signal", "id", p.ID)

		_, err := stateManager.StopProcessing(p.ID)
		if err != nil {
			logger.Error("Failed to stop processing item", "error", err)
			return
		}

	case orchestrator.DeregisterSignal:
		var p orchestrator.DeregisterPayload
		if err := orchestrator.ConvertPayload(sig.Payload, &p); err != nil {
			logger.Error("Failed to convert deregister payload", "error", err)
			return
		}
		logger.Info("Handling de-register signal", "id", p.ID)

		if err := stateManager.Deregister(p.ID); err != nil {
			logger.Error("Failed to stop de-register item", "error", err)
			return
		}

	case orchestrator.UpdateSignal:
		var p orchestrator.UpdatePayload
		if err := orchestrator.ConvertPayload(sig.Payload, &p); err != nil {
			logger.Error("Failed to convert update payload", "error", err)
			return
		}
		logger.Info("Handling update signal", "id", p.ID)

		if err := stateManager.UpdateItem(p.ID, p.Item); err != nil {
			logger.Error("Failed to update item", "error", err)
			return
		}

	case orchestrator.PingSignal:
		logger.Info("Handling ping signal")
	default:
		logger.Warn("Received unknown signal type", "type", sig.Type)
	}
}

func (ow *OW[O]) buildQueryResponse(stateManager O) orchestrator.QueryResponse {
	return orchestrator.QueryResponse{
		TotalItems:        len(stateManager.AllItems()),
		OrchestratedItems: stateManager.AllItems(),
		SignalsHandled:    stateManager.GetState().SignalsHandled,
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

	itemOrchestratorWorkflow := NewOW(orchestrator.OrchestratorWorkflowName, orchestrator.NewItemOrchestratorStateManager)
	w.RegisterWorkflowWithOptions(itemOrchestratorWorkflow.OrchestratorWorkflow, workflow.RegisterOptions{Name: orchestrator.OrchestratorWorkflowName})
	w.RegisterWorkflowWithOptions(ItemWorkflowA, workflow.RegisterOptions{Name: orchestrator.ItemWorkflowAName})
	w.RegisterWorkflowWithOptions(ItemWorkflowB, workflow.RegisterOptions{Name: orchestrator.ItemWorkflowBName})

	log.Println("Starting Orchestrator and Item Workflow worker...")
	if err := w.Run(worker.InterruptCh()); err != nil {
		log.Fatalln("Unable to start worker", err)
	}
}
