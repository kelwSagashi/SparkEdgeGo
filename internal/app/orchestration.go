package app

import (
	"context"
	"log"
	"slices"
	"time"

	"github.com/kelwSagashi/sparkedge-go/internal/domain"
	apperuntime "github.com/kelwSagashi/sparkedge-go/internal/runtime"
)

func (a *App) scheduleDependentTriggers(parentExecution domain.InstanceExecution, parentResult apperuntime.TriggerResult) {
	if a == nil || a.Instances == nil || a.Executions == nil || parentExecution.InstanceID == "" {
		return
	}
	if parentResult.Status != domain.ExecutionSuccess {
		return
	}

	go a.triggerDependentInstances(context.Background(), parentExecution, parentResult)
}

func (a *App) triggerDependentInstances(ctx context.Context, parentExecution domain.InstanceExecution, parentResult apperuntime.TriggerResult) {
	dependents, err := a.findDependentInstances(ctx, parentExecution.InstanceID)
	if err != nil {
		log.Printf("[Workflow] failed to resolve dependents for %s: %v", parentExecution.InstanceID, err)
		return
	}

	for _, instance := range dependents {
		ready, reason, err := a.dependentInstanceReady(ctx, instance)
		if err != nil {
			log.Printf("[Workflow] failed to validate dependent instance %s: %v", instance.ID, err)
			continue
		}
		if !ready {
			if reason != "" {
				log.Printf("[Workflow] skipping dependent instance %s: %s", instance.ID, reason)
			}
			continue
		}
		if !a.beginWorkflowTrigger(instance.ID) {
			continue
		}

		go func(instance domain.Instance) {
			defer a.endWorkflowTrigger(instance.ID)

			input := buildDependentTriggerInput(parentExecution, parentResult, instance)
			if _, _, err := a.triggerInstance(context.Background(), instance.ID, input, domain.TriggerWorkflow); err != nil {
				log.Printf("[Workflow] failed to trigger dependent instance %s: %v", instance.ID, err)
			}
		}(instance)
	}
}

func (a *App) findDependentInstances(ctx context.Context, upstreamInstanceID string) ([]domain.Instance, error) {
	instances, err := a.Instances.ListActive(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]domain.Instance, 0)
	for _, instance := range instances {
		if instance.ID == upstreamInstanceID {
			continue
		}
		if slices.Contains(instance.DependsOn, upstreamInstanceID) {
			result = append(result, instance)
		}
	}
	return result, nil
}

func (a *App) dependentInstanceReady(ctx context.Context, instance domain.Instance) (bool, string, error) {
	if instance.Status == domain.InstanceStatusRunning {
		return false, "instance is already running", nil
	}

	ready, err := a.dependenciesSucceeded(ctx, instance.DependsOn)
	if err != nil {
		return false, "", err
	}
	if !ready {
		return false, "waiting for all dependencies to finish successfully", nil
	}

	debounceSeconds := intFromMapDefault(instance.OrchestrationConfig, "debounce_seconds", "debounceSeconds", 0)
	if debounceSeconds <= 0 {
		return true, "", nil
	}

	executions, err := a.Executions.ListByInstance(ctx, instance.ID, 1)
	if err != nil {
		return false, "", err
	}
	if len(executions) == 0 {
		return true, "", nil
	}

	lastRunAt := executions[0].CreatedAt
	if executions[0].StartedAt != nil && !executions[0].StartedAt.IsZero() {
		lastRunAt = *executions[0].StartedAt
	}
	if time.Since(lastRunAt) < time.Duration(debounceSeconds)*time.Second {
		return false, "debounce window still active", nil
	}
	return true, "", nil
}

func (a *App) dependenciesSucceeded(ctx context.Context, dependsOn []string) (bool, error) {
	for _, dependencyID := range dependsOn {
		executions, err := a.Executions.ListByInstance(ctx, dependencyID, 1)
		if err != nil {
			return false, err
		}
		if len(executions) == 0 {
			return false, nil
		}
		if executions[0].Status != domain.ExecutionSuccess {
			return false, nil
		}
	}
	return true, nil
}

func (a *App) beginWorkflowTrigger(instanceID string) bool {
	if a == nil || instanceID == "" {
		return false
	}
	a.workflowMu.Lock()
	defer a.workflowMu.Unlock()
	if a.workflowInflight == nil {
		a.workflowInflight = map[string]bool{}
	}
	if a.workflowInflight[instanceID] {
		return false
	}
	a.workflowInflight[instanceID] = true
	return true
}

func (a *App) endWorkflowTrigger(instanceID string) {
	if a == nil || instanceID == "" {
		return
	}
	a.workflowMu.Lock()
	defer a.workflowMu.Unlock()
	delete(a.workflowInflight, instanceID)
}

func buildDependentTriggerInput(parentExecution domain.InstanceExecution, parentResult apperuntime.TriggerResult, targetInstance domain.Instance) map[string]any {
	now := time.Now().UTC().Format(time.RFC3339)
	return map[string]any{
		"workflow": map[string]any{
			"triggered_at":        now,
			"source_instance_id":  parentExecution.InstanceID,
			"source_execution_id": parentExecution.ID,
			"source_trigger_type": parentExecution.TriggerType,
			"target_instance_id":  targetInstance.ID,
			"target_instance_name": targetInstance.Name,
			"status":              parentResult.Status,
		},
		"upstream": map[string]any{
			"instance_id":         parentExecution.InstanceID,
			"execution_id":        parentExecution.ID,
			"trigger_type":        parentExecution.TriggerType,
			"status":              parentResult.Status,
			"input":               cloneMap(parentExecution.InputPayload),
			"output":              cloneMap(parentExecution.OutputPayload),
			"script_output":       cloneMap(parentResult.Output),
			"mapped_payloads":     parentResult.MappedPayloads,
			"destination_details": parentResult.DestinationDetails,
			"logs":                parentResult.Logs,
		},
	}
}
