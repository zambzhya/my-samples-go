package main

import (
	"errors"
	"fmt"
	"my-samples-go/temporal/orchestrator"
	"time"

	"go.temporal.io/sdk/workflow"
)

var errUnableToProceed = errors.New("Unable to proceed")

type ItemWorkflow[T orchestrator.Item] struct {
}

func NewItemWorkflowA(ctx workflow.Context) ItemWorkflow[orchestrator.ItemA] {
	return NewItemWorkflow[orchestrator.ItemA](ctx)
}

func ItemWorkflowA(ctx workflow.Context, item orchestrator.ItemA) (string, error) {
	w := NewItemWorkflowA(ctx)

	logger := workflow.GetLogger(ctx)
	logger.Info("ItemWorkflowA started", "ItemID", item.ID())

	item.Status = orchestrator.ItemStatusNew

	// 1. Signal the Orchestrator Workflow to register this item,
	// and Wait for the "go/no-go" signal from the orchestrator.
	err := w.RegisterAndWaitForInstructions(ctx, item)
	if errors.Is(err, errUnableToProceed) {
		item.Status = orchestrator.ItemStatusCancelled
		err = w.SendUpdate(ctx, item)
		if err != nil {
			item.Status = orchestrator.ItemStatusFailed
			logger.Error("Failed to send update signal", "error", err)
			return "Failed to send update signal", err
		}
		logger.Warn("Received 'no-go' signal from orchestrator. Completing workflow without processing.")
		return "Halted by orchestrator", nil
	}
	if err != nil {
		item.Status = orchestrator.ItemStatusFailed
		logger.Error("Failed to send register signal to orchestrator workflow", "error", err)
		return "Failed to register", err
	}

	logger.Info("Successfully registered. Waiting 30s before requesting to start processing...")

	if err := workflow.Sleep(ctx, 30*time.Second); err != nil {
		return "Failed to sleep before processing request", err
	}

	// 2. Send a StartProcessingSignal to request to start processing,
	// and Wait for the second "go/no-go" signal for processing.
	err = w.StartProcessingAndWaitForInstructions(ctx, item)
	if errors.Is(err, errUnableToProceed) {
		item.Status = orchestrator.ItemStatusCancelled
		err = w.SendUpdate(ctx, item)
		if err != nil {
			item.Status = orchestrator.ItemStatusFailed
			logger.Error("Failed to send update signal", "error", err)
			return "Failed to send update signal", err
		}
		logger.Warn("Received 'no-go' signal from orchestrator. Completing workflow without processing.")
		return "Processing denied", nil
	}
	if err != nil {
		item.Status = orchestrator.ItemStatusFailed
		logger.Error("Failed to start processing", "error", err)
		return "Failed to start processing", err
	}

	logger.Info("Request to process was approved. Starting work...")
	item.Status = orchestrator.ItemStatusProcessing

	// 3. Signal the Orchestrator Workflow with the update (status update).
	err = w.SendUpdate(ctx, item)
	if err != nil {
		item.Status = orchestrator.ItemStatusFailed
		logger.Error("Failed to send update signal", "error", err)
		return "Failed to send update signal", err
	}

	// Simulate doing work for 30 seconds
	logger.Info("Starting processing...")
	if err := workflow.Sleep(ctx, 30*time.Second); err != nil {
		item.Status = orchestrator.ItemStatusFailed
		return "Failed to sleep", err
	}

	logger.Info("Item processing complete. Stopping processing.")
	item.Status = orchestrator.ItemStatusCompleted

	// 4. Signal the Orchestrator Workflow with the update (status update).
	err = w.SendUpdate(ctx, item)
	if err != nil {
		item.Status = orchestrator.ItemStatusFailed
		logger.Error("Failed to send update signal", "error", err)
		return "Failed to send update signal", err
	}

	// 5. Signal the Orchestrator Workflow to stop processing.
	err = w.StopProcessingAndWaitForInstructions(ctx, item)
	if err != nil {
		item.Status = orchestrator.ItemStatusFailed
		logger.Error("Failed to send stop-processing signal", "error", err)
		return "Failed to stop processing", err
	}

	// 6. Signal the Orchestrator Workflow to deregister this item.
	err = w.Deregister(ctx, item)
	if err != nil {
		item.Status = orchestrator.ItemStatusFailed
		logger.Error("Failed to send deregister signal", "error", err)
		return "Failed to deregister", err
	}

	return "Finished Successfully", nil
}

func NewItemWorkflowB(ctx workflow.Context) ItemWorkflow[orchestrator.ItemB] {
	return NewItemWorkflow[orchestrator.ItemB](ctx)
}

func ItemWorkflowB(ctx workflow.Context, item orchestrator.ItemB) (string, error) {
	w := NewItemWorkflowB(ctx)

	logger := workflow.GetLogger(ctx)
	logger.Info("ItemWorkflowB started", "ItemID", item.ID())

	item.Status = orchestrator.ItemStatusNew

	// 1. Signal the Orchestrator Workflow to register this item,
	// and Wait for the "go/no-go" signal from the orchestrator.
	err := w.RegisterAndWaitForInstructions(ctx, item)
	if errors.Is(err, errUnableToProceed) {
		item.Status = orchestrator.ItemStatusCancelled
		err = w.SendUpdate(ctx, item)
		if err != nil {
			item.Status = orchestrator.ItemStatusFailed
			logger.Error("Failed to send update signal", "error", err)
			return "Failed to send update signal", err
		}
		logger.Warn("Received 'no-go' signal from orchestrator. Completing workflow without processing.")
		return "Halted by orchestrator", nil
	}
	if err != nil {
		item.Status = orchestrator.ItemStatusFailed
		logger.Error("Failed to send register signal to orchestrator workflow", "error", err)
		return "Failed to register", err
	}

	logger.Info("Successfully registered. Waiting 30s before requesting to start processing...")
	if err := workflow.Sleep(ctx, 30*time.Second); err != nil {
		return "Failed to sleep before processing request", err
	}

	// 2. Send a StartProcessingSignal to request to start processing,
	// and Wait for the second "go/no-go" signal for processing.
	err = w.StartProcessingAndWaitForInstructions(ctx, item)
	if errors.Is(err, errUnableToProceed) {
		item.Status = orchestrator.ItemStatusCancelled
		err = w.SendUpdate(ctx, item)
		if err != nil {
			item.Status = orchestrator.ItemStatusFailed
			logger.Error("Failed to send update signal", "error", err)
			return "Failed to send update signal", err
		}
		logger.Warn("Received 'no-go' signal from orchestrator. Completing workflow without processing.")
		return "Processing denied", nil
	}
	if err != nil {
		item.Status = orchestrator.ItemStatusFailed
		logger.Error("Failed to start processing", "error", err)
		return "Failed to start processing", err
	}

	logger.Info("Request to process was approved. Starting work...")
	item.Status = orchestrator.ItemStatusProcessing

	// 3. Signal the Orchestrator Workflow with the update (status update).
	err = w.SendUpdate(ctx, item)
	if err != nil {
		item.Status = orchestrator.ItemStatusFailed
		logger.Error("Failed to send update signal", "error", err)
		return "Failed to send update signal", err
	}

	// Simulate doing work for 30 seconds
	logger.Info("Starting processing...")
	if err := workflow.Sleep(ctx, 30*time.Second); err != nil {
		item.Status = orchestrator.ItemStatusFailed
		return "Failed to sleep", err
	}

	logger.Info("Item processing complete. Stopping processing.")
	item.Status = orchestrator.ItemStatusCompleted

	// 4. Signal the Orchestrator Workflow with the update (status update).
	err = w.SendUpdate(ctx, item)
	if err != nil {
		item.Status = orchestrator.ItemStatusFailed
		logger.Error("Failed to send update signal", "error", err)
		return "Failed to send update signal", err
	}

	// 5. Signal the Orchestrator Workflow to stop processing.
	err = w.StopProcessingAndWaitForInstructions(ctx, item)
	if err != nil {
		item.Status = orchestrator.ItemStatusFailed
		logger.Error("Failed to send stop-processing signal", "error", err)
		return "Failed to stop processing", err
	}

	// 6. Signal the Orchestrator Workflow to deregister this item.
	err = w.Deregister(ctx, item)
	if err != nil {
		item.Status = orchestrator.ItemStatusFailed
		logger.Error("Failed to send deregister signal", "error", err)
		return "Failed to deregister", err
	}

	return "Finished Successfully", nil
}

// ItemWorkflow is the workflow that processes a single item.
func NewItemWorkflow[T orchestrator.Item](ctx workflow.Context) ItemWorkflow[T] {
	return ItemWorkflow[T]{}
}

func (w ItemWorkflow[T]) RegisterAndWaitForInstructions(ctx workflow.Context, item T) error {
	// Signal the Orchestrator Workflow to register this item.
	info := workflow.GetInfo(ctx)
	registerPayload := orchestrator.RegisterPayload{
		ID:                item.ID(),
		ItemWorkflowID:    info.WorkflowExecution.ID,
		ItemWorkflowRunID: info.WorkflowExecution.RunID,
		Item:              item,
	}
	registerSignal := orchestrator.Signal{
		Type:    orchestrator.RegisterSignal,
		Payload: registerPayload,
	}

	// Use SignalWorkflow from within the workflow to signal the orchestrator
	err := workflow.SignalExternalWorkflow(ctx, orchestrator.OrchestratorWorkflowID, "", orchestrator.SignalChannelName, registerSignal).Get(ctx, nil)
	if err != nil {
		return fmt.Errorf("Failed to send register signal to orchestrator workflow: %w", err)
	}

	// Wait for the "go/no-go" signal from the orchestrator.
	signalCh := workflow.GetSignalChannel(ctx, orchestrator.ItemSignalChannelName)
	var processSignal orchestrator.ItemInstructionSignal
	signalCh.Receive(ctx, &processSignal) // This will block until the signal is received

	// Decide whether to proceed based on the signal.
	if !processSignal.Proceed {
		return errors.Join(errUnableToProceed, fmt.Errorf("Received 'no-go' signal from orchestrator. Reason: %s", processSignal.Reason))
	}

	return nil
}

func (w ItemWorkflow[T]) StartProcessingAndWaitForInstructions(ctx workflow.Context, item T) error {
	// Signal the Orchestrator Workflow to request permission to start processing this item.
	startProcessingPayload := orchestrator.StartProcessingPayload{ID: item.ID()}
	startProcessingSignal := orchestrator.Signal{
		Type:    orchestrator.StartProcessingSignal,
		Payload: startProcessingPayload,
	}
	err := workflow.SignalExternalWorkflow(ctx, orchestrator.OrchestratorWorkflowID, "", orchestrator.SignalChannelName, startProcessingSignal).Get(ctx, nil)
	if err != nil {
		return fmt.Errorf("Failed to send start-processing signal to orchestrator workflow: %w", err)
	}

	// Wait for the "go/no-go" signal for processing.
	signalCh := workflow.GetSignalChannel(ctx, orchestrator.ItemSignalChannelName)
	var processSignal orchestrator.ItemInstructionSignal
	signalCh.Receive(ctx, &processSignal) // Block until the signal is received
	if !processSignal.Proceed {
		// Also deregister since we are not proceeding.
		deregisterPayload := orchestrator.DeregisterPayload{ID: item.ID()}
		deregisterSignal := orchestrator.Signal{Type: orchestrator.DeregisterSignal, Payload: deregisterPayload}
		err = workflow.SignalExternalWorkflow(ctx, orchestrator.OrchestratorWorkflowID, "", orchestrator.SignalChannelName, deregisterSignal).Get(ctx, nil)
		if err != nil {
			return fmt.Errorf("Failed to send deregister signal after processing denial: %w", err)
		}
		return errors.Join(errUnableToProceed, fmt.Errorf("Request to process was denied by orchestrator. Reason: %s", processSignal.Reason))
	}

	return nil
}

func (w ItemWorkflow[T]) StopProcessingAndWaitForInstructions(ctx workflow.Context, item T) error {
	// Signal the Orchestrator Workflow to request permission to stop processing this item.
	stopProcessingPayload := orchestrator.StopProcessingPayload{ID: item.ID()}
	stopProcessingSignal := orchestrator.Signal{
		Type:    orchestrator.StopProcessingSignal,
		Payload: stopProcessingPayload,
	}
	err := workflow.SignalExternalWorkflow(ctx, orchestrator.OrchestratorWorkflowID, "", orchestrator.SignalChannelName, stopProcessingSignal).Get(ctx, nil)
	if err != nil {
		return fmt.Errorf("Failed to send stop-processing signal to orchestrator workflow: %w", err)
	}

	return nil
}

func (w ItemWorkflow[T]) Deregister(ctx workflow.Context, item T) error {
	// Signal the Orchestrator Workflow to deregister this item.
	deregisterPayload := orchestrator.DeregisterPayload{
		ID: item.ID(),
	}
	deregisterSignal := orchestrator.Signal{
		Type:    orchestrator.DeregisterSignal,
		Payload: deregisterPayload,
	}
	err := workflow.SignalExternalWorkflow(ctx, orchestrator.OrchestratorWorkflowID, "", orchestrator.SignalChannelName, deregisterSignal).Get(ctx, nil)
	if err != nil {
		return fmt.Errorf("Failed to send deregister signal to orchestrator workflow: %w", err)
	}

	return nil
}

func (w ItemWorkflow[T]) SendUpdate(ctx workflow.Context, item T) error {
	updatePayload := orchestrator.UpdatePayload{
		ID:   item.ID(),
		Item: item,
	}
	updateSignal := orchestrator.Signal{
		Type:    orchestrator.UpdateSignal,
		Payload: updatePayload,
	}
	err := workflow.SignalExternalWorkflow(ctx, orchestrator.OrchestratorWorkflowID, "", orchestrator.SignalChannelName, updateSignal).Get(ctx, nil)
	if err != nil {
		return fmt.Errorf("Failed to send update signal to orchestrator workflow: %w", err)
	}

	return nil
}
