package orchestrator

import (
	tlog "go.temporal.io/sdk/log"
	"go.temporal.io/sdk/workflow"
)

type OrchestratorState struct {
	SignalsHandled int
	Orchestrator   ItemOrchestrator
}

func (s *OrchestratorState) HandleSignal(logger tlog.Logger, sig Signal, ctx workflow.Context) {
	s.SignalsHandled++
	switch sig.Type {
	case RegisterSignal:
		var p RegisterPayload
		if err := ConvertPayload(sig.Payload, &p); err != nil {
			logger.Error("Failed to convert register payload", "error", err)
			return
		}
		logger.Info("Handling register signal", "id", p.ID)

		canProceed := true
		reason := "Registration accepted."
		if err := s.Orchestrator.RegisterItem(p.ID, p.ItemWorkflowID, p.ItemWorkflowRunID, p.Item); err != nil {
			logger.Error("Failed to register item", "error", err)
			canProceed = false
			reason = "Registration denied: " + err.Error()
		}

		itemSignal := ItemInstructionSignal{ID: p.ID, Proceed: canProceed, Reason: reason}

		logger.Info("Sending signal to item workflow", "workflowID", p.ItemWorkflowID, "proceed", canProceed)
		err := workflow.SignalExternalWorkflow(ctx, p.ItemWorkflowID, p.ItemWorkflowRunID, ItemSignalChannelName, itemSignal).Get(ctx, nil)
		if err != nil {
			logger.Error("Failed to send signal to item workflow", "error", err, "itemWorkflowID", p.ItemWorkflowID)
			return
		}

	case StartProcessingSignal:
		var p StartProcessingPayload
		if err := ConvertPayload(sig.Payload, &p); err != nil {
			logger.Error("Failed to convert start-processing payload", "error", err)
			return
		}
		logger.Info("Handling start-processing request", "id", p.ID)

		canProceed := true
		reason := "Start processing permitted."
		item, err := s.Orchestrator.StartProcessing(p.ID)
		if err != nil {
			logger.Error("Failed to start processing", "error", err)
			canProceed = false
			reason = "Start processing denied: " + err.Error()
		}

		if item == nil {
			return
		}

		itemSignal := ItemInstructionSignal{ID: p.ID, Proceed: canProceed, Reason: reason}

		logger.Info("Sending signal to item workflow", "workflowID", item.ItemWorkflowID, "proceed", canProceed)
		err = workflow.SignalExternalWorkflow(ctx, item.ItemWorkflowID, item.ItemWorkflowRunID, ItemSignalChannelName, itemSignal).Get(ctx, nil)
		if err != nil {
			logger.Error("Failed to send signal to item workflow", "error", err, "itemWorkflowID", item.ItemWorkflowID)
			return
		}

	case StopProcessingSignal:
		var p StopProcessingPayload
		if err := ConvertPayload(sig.Payload, &p); err != nil {
			logger.Error("Failed to convert stop-processing payload", "error", err)
			return
		}
		logger.Info("Handling stop-processing signal", "id", p.ID)

		_, err := s.Orchestrator.StopProcessing(p.ID)
		if err != nil {
			logger.Error("Failed to stop processing item", "error", err)
			return
		}

	case DeregisterSignal:
		var p DeregisterPayload
		if err := ConvertPayload(sig.Payload, &p); err != nil {
			logger.Error("Failed to convert deregister payload", "error", err)
			return
		}
		logger.Info("Handling de-register signal", "id", p.ID)

		if err := s.Orchestrator.Deregister(p.ID); err != nil {
			logger.Error("Failed to stop de-register item", "error", err)
			return
		}

	case PingSignal:
		logger.Info("Handling ping signal")
	default:
		logger.Warn("Received unknown signal type", "type", sig.Type)
	}
}

func (s OrchestratorState) IncrementSignalsHandled() {
	s.SignalsHandled++
}
