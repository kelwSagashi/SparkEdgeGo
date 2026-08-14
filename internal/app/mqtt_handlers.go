package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	goruntime "runtime"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/kelwSagashi/sparkedge-go/internal/domain"
	apperuntime "github.com/kelwSagashi/sparkedge-go/internal/runtime"
	"github.com/kelwSagashi/sparkedge-go/internal/sqlite"
)

func (a *App) registerMqttCommandHandlers() {
	if a == nil || a.MQTT == nil {
		return
	}
	a.MQTT.RegisterHandler("get_stats", func(context.Context, map[string]any) (map[string]any, error) {
		return a.collectOperationalSnapshot(context.Background()), nil
	})
	a.MQTT.RegisterHandler("run_script", a.handleMqttRunScript)
	a.MQTT.RegisterHandler("remote_job", a.handleMqttRemoteJob)
	a.MQTT.RegisterHandler("CONFIG", a.handleMqttConfig)
	a.MQTT.RegisterHandler("restart", a.handleMqttRestart)
	a.MQTT.RegisterHandler("REBOOT", a.handleMqttRestart)
}

func (a *App) handleMqttRemoteJob(ctx context.Context, payload map[string]any) (map[string]any, error) {
	jobID := firstString(payload, "job_id", "cloud_job_id")
	if jobID == "" {
		jobID = fmt.Sprintf("job-%d", time.Now().UTC().UnixNano())
	}
	action := firstString(payload, "action")
	if action == "" {
		action = "trigger_instance"
	}

	_ = a.enqueueRemoteJobCloudSync(ctx, jobID, action, "accepted", remoteJobPayload(payload, "accepted", "job_received", 5), nil)
	_ = a.enqueueRemoteJobCloudSync(ctx, jobID, action, "running", remoteJobPayload(payload, "running", "job_started", 15), nil)

	if action == "run_workflow" {
		response, err := a.runRemoteWorkflow(ctx, jobID, payload)
		if err != nil {
			_ = a.enqueueRemoteJobCloudSync(ctx, jobID, action, "failed", remoteJobPayload(payload, "failed", "workflow_failed", 100), err)
			return nil, err
		}
		_ = a.enqueueRemoteJobCloudSync(ctx, jobID, action, "completed", remoteJobPayload(response, "completed", "workflow_completed", 100), nil)
		return response, nil
	}

	response, err := a.executeRemoteJobAction(ctx, jobID, action, payload)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (a *App) executeRemoteJobAction(ctx context.Context, jobID string, action string, payload map[string]any) (map[string]any, error) {
	switch action {
	case "trigger_instance":
		instanceID := firstString(payload, "instance_id", "script_name")
		if instanceID == "" {
			err := errors.New("missing instance_id or script_name in remote job payload")
			_ = a.enqueueRemoteJobCloudSync(ctx, jobID, action, "failed", remoteJobPayload(payload, "failed", "validate_input", 15), err)
			return nil, err
		}
		_ = a.enqueueRemoteJobCloudSync(ctx, jobID, action, "running", remoteJobPayload(map[string]any{"instance_id": instanceID}, "running", "instance_triggering", 45), nil)
		execution, result, err := a.triggerInstance(ctx, instanceID, mapValue(payload, "input", "inputs", "data"), domain.TriggerManual)
		if err != nil {
			_ = a.enqueueRemoteJobCloudSync(ctx, jobID, action, "failed", remoteJobPayload(map[string]any{"instance_id": instanceID}, "failed", "instance_triggering", 45), err)
			return nil, err
		}
		if result.Status != domain.ExecutionSuccess {
			jobErr := errors.New(result.Error)
			if result.Error == "" {
				jobErr = errors.New("remote job execution failed")
			}
			_ = a.enqueueRemoteJobCloudSync(ctx, jobID, action, "failed", remoteJobPayload(map[string]any{"instance_id": instanceID, "execution_id": execution.ID}, "failed", "instance_finished", 90), jobErr)
			return nil, jobErr
		}
		response := map[string]any{
			"job_id":       jobID,
			"action":       action,
			"execution_id": execution.ID,
			"instance_id":  execution.InstanceID,
			"status":       "completed",
		}
		_ = a.enqueueRemoteJobCloudSync(ctx, jobID, action, "completed", remoteJobPayload(response, "completed", "instance_finished", 100), nil)
		return response, nil
	case "collect_meta":
		_ = a.enqueueRemoteJobCloudSync(ctx, jobID, action, "running", remoteJobPayload(payload, "running", "metadata_collecting", 40), nil)
		response, err := a.runRemoteCollectMeta(ctx, jobID, action)
		if err != nil {
			_ = a.enqueueRemoteJobCloudSync(ctx, jobID, action, "failed", remoteJobPayload(payload, "failed", "metadata_collecting", 40), err)
			return nil, err
		}
		_ = a.enqueueRemoteJobCloudSync(ctx, jobID, action, "completed", remoteJobPayload(response, "completed", "metadata_published", 100), nil)
		return response, nil
	case "sync_now":
		_ = a.enqueueRemoteJobCloudSync(ctx, jobID, action, "running", remoteJobPayload(payload, "running", "cloudsync_flushing", 50), nil)
		response, err := a.runRemoteSyncNow(ctx, jobID, action)
		if err != nil {
			_ = a.enqueueRemoteJobCloudSync(ctx, jobID, action, "failed", remoteJobPayload(payload, "failed", "cloudsync_flushing", 50), err)
			return nil, err
		}
		_ = a.enqueueRemoteJobCloudSync(ctx, jobID, action, "completed", remoteJobPayload(response, "completed", "cloudsync_finished", 100), nil)
		return response, nil
	case "refresh_edge_state":
		_ = a.enqueueRemoteJobCloudSync(ctx, jobID, action, "running", remoteJobPayload(payload, "running", "metadata_collecting", 25), nil)
		metaResponse, err := a.runRemoteCollectMeta(ctx, jobID, action)
		if err != nil {
			_ = a.enqueueRemoteJobCloudSync(ctx, jobID, action, "failed", remoteJobPayload(payload, "failed", "metadata_collecting", 25), err)
			return nil, err
		}
		_ = a.enqueueRemoteJobCloudSync(ctx, jobID, action, "running", remoteJobPayload(payload, "running", "stats_publishing", 55), nil)
		if a.MQTT != nil {
			_ = a.MQTT.PublishHeartbeat(ctx)
			_ = a.MQTT.PublishStats(ctx)
		}
		snapshot := a.collectOperationalSnapshot(ctx)
		_ = a.enqueueRemoteJobCloudSync(ctx, jobID, action, "running", remoteJobPayload(snapshot, "running", "cloudsync_flushing", 80), nil)
		syncResponse, err := a.runRemoteSyncNow(ctx, jobID, action)
		if err != nil {
			_ = a.enqueueRemoteJobCloudSync(ctx, jobID, action, "failed", remoteJobPayload(snapshot, "failed", "cloudsync_flushing", 80), err)
			return nil, err
		}
		response := map[string]any{
			"job_id":    jobID,
			"action":    action,
			"status":    "completed",
			"meta":      metaResponse,
			"stats":     snapshot,
			"cloudsync": syncResponse,
		}
		_ = a.enqueueRemoteJobCloudSync(ctx, jobID, action, "completed", remoteJobPayload(response, "completed", "edge_state_refreshed", 100), nil)
		return response, nil
	default:
		err := fmt.Errorf("unsupported remote job action: %s", action)
		_ = a.enqueueRemoteJobCloudSync(ctx, jobID, action, "failed", remoteJobPayload(payload, "failed", "unsupported_action", 10), err)
		return nil, err
	}
}

func (a *App) runRemoteWorkflow(ctx context.Context, jobID string, payload map[string]any) (map[string]any, error) {
	rawSteps, ok := payload["steps"].([]any)
	if !ok || len(rawSteps) == 0 {
		return nil, errors.New("remote workflow requires a non-empty steps array")
	}

	mode := strings.ToLower(firstString(payload, "mode", "execution_mode", "executionMode"))
	if mode == "" {
		mode = "sequential"
	}
	maxParallelism := intValue(payload, "max_parallelism", "maxParallelism")
	if maxParallelism <= 0 {
		maxParallelism = len(rawSteps)
	}

	steps := make([]workflowStepSpec, 0, len(rawSteps))
	for index, rawStep := range rawSteps {
		stepPayload, ok := rawStep.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("workflow step %d is invalid", index+1)
		}
		stepSpec, err := buildWorkflowStepSpec(payload, stepPayload, index+1, len(rawSteps))
		if err != nil {
			return nil, err
		}
		steps = append(steps, stepSpec)
	}

	var (
		results   []map[string]any
		hadErrors bool
		err       error
	)
	if mode == "parallel" {
		results, hadErrors, err = a.runRemoteWorkflowParallel(ctx, jobID, steps, maxParallelism)
	} else {
		mode = "sequential"
		results, hadErrors, err = a.runRemoteWorkflowSequential(ctx, jobID, steps)
	}
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"job_id":          jobID,
		"action":          "run_workflow",
		"mode":            mode,
		"max_parallelism": maxParallelism,
		"status":          "completed",
		"had_errors":      hadErrors,
		"steps":           results,
	}, nil
}

type workflowStepSpec struct {
	Index           int
	Total           int
	Key             string
	Action          string
	StepName        string
	Payload         map[string]any
	DependsOn       []string
	ContinueOnError bool
	TimeoutSeconds  int
	MaxAttempts     int
	ProgressBase    int
}

func buildWorkflowStepSpec(workflowPayload map[string]any, stepPayload map[string]any, index int, total int) (workflowStepSpec, error) {
	stepAction := firstString(stepPayload, "action")
	if stepAction == "" {
		return workflowStepSpec{}, fmt.Errorf("workflow step %d is missing action", index)
	}
	continueOnError := boolValue(stepPayload, "continue_on_error", "continueOnError")
	timeoutSeconds := intValue(stepPayload, "timeout_seconds", "timeoutSeconds")
	maxAttempts := intValue(stepPayload, "max_attempts", "maxAttempts")
	if maxAttempts <= 0 {
		retries := intValue(stepPayload, "retries", "retry_count", "retryCount")
		if retries > 0 {
			maxAttempts = retries + 1
		}
	}
	if maxAttempts <= 0 {
		maxAttempts = 1
	}
	progressBase := 20
	if total > 0 {
		progressBase += int(float64(index-1) / float64(total) * 70)
	}
	stepKey := firstString(stepPayload, "id", "key", "name")
	if stepKey == "" {
		stepKey = fmt.Sprintf("step_%d", index)
	}
	mergedPayload := map[string]any{}
	for key, value := range workflowPayload {
		if key == "steps" {
			continue
		}
		mergedPayload[key] = value
	}
	for key, value := range stepPayload {
		mergedPayload[key] = value
	}
	mergedPayload["action"] = stepAction

	return workflowStepSpec{
		Index:           index,
		Total:           total,
		Key:             stepKey,
		Action:          stepAction,
		StepName:        fmt.Sprintf("workflow_step_%d_%s", index, stepAction),
		Payload:         mergedPayload,
		DependsOn:       dependencyRefs(stepPayload),
		ContinueOnError: continueOnError,
		TimeoutSeconds:  timeoutSeconds,
		MaxAttempts:     maxAttempts,
		ProgressBase:    progressBase,
	}, nil
}

func (a *App) runRemoteWorkflowSequential(ctx context.Context, jobID string, steps []workflowStepSpec) ([]map[string]any, bool, error) {
	if err := resolveWorkflowDependencies(steps); err != nil {
		return nil, false, err
	}
	results := make([]map[string]any, 0, len(steps))
	hadErrors := false
	stepStates := map[string]string{}
	for _, step := range steps {
		if depErr := dependencyStateError(step, stepStates); depErr != nil {
			hadErrors = true
			if !step.ContinueOnError {
				return nil, hadErrors, depErr
			}
			results = append(results, blockedStepResult(step, depErr))
			stepStates[step.Key] = "blocked_dependency"
			continue
		}
		result, err := a.executeWorkflowStep(ctx, jobID, step)
		if err != nil {
			hadErrors = true
			if !step.ContinueOnError {
				return nil, hadErrors, fmt.Errorf("workflow step %d (%s) failed: %w", step.Index, step.Action, err)
			}
			results = append(results, failedContinueStepResult(step, err))
			stepStates[step.Key] = "failed_continue"
			continue
		}
		results = append(results, result)
		stepStates[step.Key] = "completed"
	}
	return results, hadErrors, nil
}

func (a *App) runRemoteWorkflowParallel(ctx context.Context, jobID string, steps []workflowStepSpec, maxParallelism int) ([]map[string]any, bool, error) {
	if err := resolveWorkflowDependencies(steps); err != nil {
		return nil, false, err
	}

	type workflowStepOutcome struct {
		index  int
		result map[string]any
		err    error
		step   workflowStepSpec
	}

	hadErrors := false
	fatalErrors := make([]error, 0)
	resultsByIndex := map[int]map[string]any{}
	stepStates := map[string]string{}
	pending := make(map[string]workflowStepSpec, len(steps))
	for _, step := range steps {
		pending[step.Key] = step
	}

	for len(pending) > 0 {
		ready := make([]workflowStepSpec, 0)
		blocked := make([]workflowStepSpec, 0)
		for _, step := range pending {
			state, depErr := dependencyResolutionState(step, stepStates)
			switch state {
			case "ready":
				ready = append(ready, step)
			case "blocked":
				blocked = append(blocked, step)
				hadErrors = true
				resultsByIndex[step.Index] = blockedStepResult(step, depErr)
				stepStates[step.Key] = "blocked_dependency"
			}
		}

		for _, step := range blocked {
			delete(pending, step.Key)
		}

		if len(ready) == 0 {
			if len(blocked) > 0 {
				continue
			}
			return nil, true, errors.New("workflow dependencies could not be resolved; check for cycles or unsatisfied prerequisites")
		}

		batch := ready
		if len(batch) > maxParallelism {
			batch = batch[:maxParallelism]
		}

		outcomes := make(chan workflowStepOutcome, len(batch))
		var wg sync.WaitGroup
		for _, step := range batch {
			delete(pending, step.Key)
			wg.Add(1)
			go func(step workflowStepSpec) {
				defer wg.Done()
				result, err := a.executeWorkflowStep(ctx, jobID, step)
				outcomes <- workflowStepOutcome{
					index:  step.Index,
					result: result,
					err:    err,
					step:   step,
				}
			}(step)
		}
		wg.Wait()
		close(outcomes)

		for outcome := range outcomes {
			if outcome.err != nil {
				hadErrors = true
				if !outcome.step.ContinueOnError {
					fatalErrors = append(fatalErrors, fmt.Errorf("workflow step %d (%s) failed: %w", outcome.step.Index, outcome.step.Action, outcome.err))
					stepStates[outcome.step.Key] = "failed"
					continue
				}
				resultsByIndex[outcome.index] = failedContinueStepResult(outcome.step, outcome.err)
				stepStates[outcome.step.Key] = "failed_continue"
				continue
			}
			resultsByIndex[outcome.index] = outcome.result
			stepStates[outcome.step.Key] = "completed"
		}
	}

	if len(fatalErrors) > 0 {
		return nil, hadErrors, fatalErrors[0]
	}

	orderedIndexes := make([]int, 0, len(resultsByIndex))
	for index := range resultsByIndex {
		orderedIndexes = append(orderedIndexes, index)
	}
	sort.Ints(orderedIndexes)

	results := make([]map[string]any, 0, len(orderedIndexes))
	for _, index := range orderedIndexes {
		results = append(results, resultsByIndex[index])
	}
	return results, hadErrors, nil
}

func (a *App) executeWorkflowStep(ctx context.Context, jobID string, step workflowStepSpec) (map[string]any, error) {
	stepSyncPayload := map[string]any{
		"workflow_index":    step.Index,
		"workflow_total":    step.Total,
		"workflow_key":      step.Key,
		"workflow_step":     step.StepName,
		"step_action":       step.Action,
		"depends_on":        step.DependsOn,
		"max_attempts":      step.MaxAttempts,
		"timeout_seconds":   step.TimeoutSeconds,
		"continue_on_error": step.ContinueOnError,
		"mode":              step.Payload["mode"],
		"max_parallelism":   step.Payload["max_parallelism"],
		"steps":             nil,
	}
	for key, value := range step.Payload {
		if key == "steps" {
			continue
		}
		stepSyncPayload[key] = value
	}
	_ = a.enqueueRemoteJobCloudSync(ctx, jobID, "run_workflow", "running", remoteJobPayload(stepSyncPayload, "running", step.StepName, step.ProgressBase), nil)

	var (
		stepResult map[string]any
		err        error
	)
	for attempt := 1; attempt <= step.MaxAttempts; attempt++ {
		attemptPayload := cloneMap(step.Payload)
		attemptPayload["attempt"] = attempt
		attemptPayload["max_attempts"] = step.MaxAttempts
		attemptCtx := ctx
		var cancel context.CancelFunc
		if step.TimeoutSeconds > 0 {
			attemptCtx, cancel = context.WithTimeout(ctx, time.Duration(step.TimeoutSeconds)*time.Second)
		}
		_ = a.enqueueRemoteJobCloudSync(
			attemptCtx,
			jobID,
			"run_workflow",
			"running",
			remoteJobPayload(map[string]any{
				"workflow_index":    step.Index,
				"workflow_total":    step.Total,
				"workflow_key":      step.Key,
				"workflow_step":     step.StepName,
				"step_action":       step.Action,
				"depends_on":        step.DependsOn,
				"attempt":           attempt,
				"max_attempts":      step.MaxAttempts,
				"continue_on_error": step.ContinueOnError,
				"timeout_seconds":   step.TimeoutSeconds,
				"mode":              step.Payload["mode"],
				"max_parallelism":   step.Payload["max_parallelism"],
			}, "running", fmt.Sprintf("%s_attempt_%d", step.StepName, attempt), step.ProgressBase),
			nil,
		)
		stepResult, err = a.executeRemoteJobAction(attemptCtx, jobID, step.Action, attemptPayload)
		if cancel != nil {
			cancel()
		}
		if err == nil {
			break
		}
		if errors.Is(err, context.DeadlineExceeded) {
			err = fmt.Errorf("step timed out after %ds", step.TimeoutSeconds)
		}
		if attempt < step.MaxAttempts {
			_ = a.enqueueRemoteJobCloudSync(
				ctx,
				jobID,
				"run_workflow",
				"running",
				remoteJobPayload(map[string]any{
					"workflow_index":  step.Index,
					"workflow_total":  step.Total,
					"workflow_key":    step.Key,
					"workflow_step":   step.StepName,
					"step_action":     step.Action,
					"depends_on":      step.DependsOn,
					"attempt":         attempt,
					"max_attempts":    step.MaxAttempts,
					"error":           err.Error(),
					"mode":            step.Payload["mode"],
					"max_parallelism": step.Payload["max_parallelism"],
				}, "running", fmt.Sprintf("%s_retrying", step.StepName), step.ProgressBase),
				nil,
			)
		}
	}
	if err != nil {
		if step.ContinueOnError {
			_ = a.enqueueRemoteJobCloudSync(
				ctx,
				jobID,
				"run_workflow",
				"running",
				remoteJobPayload(map[string]any{
					"workflow_index":    step.Index,
					"workflow_total":    step.Total,
					"workflow_key":      step.Key,
					"workflow_step":     step.StepName,
					"step_action":       step.Action,
					"depends_on":        step.DependsOn,
					"continue_on_error": true,
					"error":             err.Error(),
					"mode":              step.Payload["mode"],
					"max_parallelism":   step.Payload["max_parallelism"],
				}, "running", fmt.Sprintf("%s_failed_continue", step.StepName), step.ProgressBase),
				nil,
			)
		}
		return nil, err
	}

	return map[string]any{
		"index":             step.Index,
		"key":               step.Key,
		"action":            step.Action,
		"status":            "completed",
		"depends_on":        step.DependsOn,
		"continue_on_error": step.ContinueOnError,
		"result":            stepResult,
	}, nil
}

func (a *App) runRemoteCollectMeta(ctx context.Context, jobID string, action string) (map[string]any, error) {
	edgeInfo, err := a.Edge.Load(ctx)
	if err != nil {
		return nil, err
	}
	config, _, _ := a.Edge.GetOnboarding(ctx)
	metadata := map[string]any{
		"edge_id":         edgeInfo.EdgeID,
		"edge_name":       edgeInfo.EdgeName,
		"description":     emptyStringToNil(config.Description),
		"lat":             emptyStringToNil(config.Lat),
		"lng":             emptyStringToNil(config.Lng),
		"tags":            config.Tags,
		"location_source": emptyStringToNil(config.LocationSource),
		"os":              goruntime.GOOS,
		"os_version":      goruntime.GOOS + "/" + goruntime.GOARCH,
		"edge_version":    envOrDefault("SPARKEDGE_VERSION", "go-dev"),
		"hardware":        envOrDefault("SPARKEDGE_HARDWARE", goruntime.GOARCH),
		"environment":     envOrDefault("SPARKEDGE_ENV", "production"),
	}
	if a.MQTT != nil {
		_ = a.MQTT.PublishMeta(ctx, metadata)
		_ = a.MQTT.PublishHeartbeat(ctx)
		_ = a.MQTT.PublishStats(ctx)
	}
	response := map[string]any{
		"job_id":     jobID,
		"action":     action,
		"edge_id":    edgeInfo.EdgeID,
		"edge_name":  edgeInfo.EdgeName,
		"status":     "completed",
		"meta_sent":  true,
		"stats_sent": true,
	}
	return response, nil
}

func (a *App) runRemoteSyncNow(ctx context.Context, jobID string, action string) (map[string]any, error) {
	if a.CloudSync == nil {
		return nil, errors.New("cloud sync is not configured")
	}
	result, err := a.CloudSync.Flush(ctx, 50)
	if err != nil {
		return nil, err
	}
	response := map[string]any{
		"job_id": jobID,
		"action": action,
		"status": "completed",
		"flush":  result,
	}
	return response, nil
}

func (a *App) handleMqttRunScript(ctx context.Context, payload map[string]any) (map[string]any, error) {
	instanceID := firstString(payload, "script_name", "instance_id", "instanceId")
	if instanceID == "" {
		return nil, errors.New("missing script_name or instance_id in payload")
	}
	_ = a.MQTT.PublishLog(ctx, "info", "Executing script: "+instanceID)
	execution, result, err := a.triggerInstance(ctx, instanceID, mapValue(payload, "input", "inputs", "data"), domain.TriggerManual)
	if err != nil {
		return nil, fmt.Errorf("script execution failed: %w", err)
	}
	if result.Status != domain.ExecutionSuccess {
		if result.Error != "" {
			return nil, errors.New(result.Error)
		}
		return nil, errors.New("script execution failed")
	}
	return map[string]any{"execution_id": execution.ID, "status": result.Status, "message": "Script triggered successfully"}, nil
}

func (a *App) handleMqttConfig(ctx context.Context, payload map[string]any) (map[string]any, error) {
	config := payload
	if nested, ok := payload["data"].(map[string]any); ok {
		config = nested
	}
	if len(config) == 0 {
		return nil, errors.New("missing configuration data")
	}
	if _, err := a.Edge.UpsertConfigMap(ctx, config); err != nil {
		return nil, err
	}
	_ = a.MQTT.PublishLog(ctx, "info", "Local configuration updated via MQTT")
	return map[string]any{"message": "Configuration synchronized"}, nil
}

func (a *App) handleMqttRestart(ctx context.Context, _ map[string]any) (map[string]any, error) {
	_ = a.MQTT.PublishLog(ctx, "warn", "Edge rebooting via remote command...")
	if strings.EqualFold(os.Getenv("SPARKEDGE_ALLOW_RESTART"), "true") {
		go func() {
			time.Sleep(2 * time.Second)
			command, args := restartCommand()
			_ = exec.Command(command, args...).Start()
		}()
		return map[string]any{"message": "Reboot initiated"}, nil
	}
	return map[string]any{"message": "Reboot command acknowledged but not executed; set SPARKEDGE_ALLOW_RESTART=true to enable system restart"}, nil
}

func (a *App) triggerInstance(ctx context.Context, instanceID string, input map[string]any, triggerType domain.TriggerType) (domain.InstanceExecution, apperuntime.TriggerResult, error) {
	instanceWithDestinations, err := a.Instances.GetWithDestinations(ctx, instanceID)
	if err != nil {
		return domain.InstanceExecution{}, apperuntime.TriggerResult{}, err
	}
	instance := instanceWithDestinations.Instance
	script, err := a.Scripts.FindByID(ctx, instance.ScriptID)
	if err != nil {
		return domain.InstanceExecution{}, apperuntime.TriggerResult{}, err
	}

	startedAt := time.Now().UTC()
	execution, err := a.Executions.Create(ctx, sqlite.CreateInstanceExecutionParams{
		InstanceID:  instance.ID,
		Status:      domain.ExecutionRunning,
		TriggerType: triggerType,
		StartedAt:   &startedAt,
		Logs:        []domain.ExecutionLog{{Level: "info", Message: "Queued MQTT trigger", Timestamp: startedAt}},
	})
	if err != nil {
		return domain.InstanceExecution{}, apperuntime.TriggerResult{}, err
	}

	_, _ = a.Instances.UpdateStatus(ctx, instance.ID, domain.InstanceStatusRunning)
	result, runErr := a.Runtime.Trigger(ctx, apperuntime.TriggerRequest{
		ExecutionID:  execution.ID,
		Instance:     instance,
		Script:       script,
		Destinations: instanceWithDestinations.Destinations,
		Trigger:      triggerType,
		Input:        input,
	})
	updated, updateErr := a.finishExecution(ctx, execution.ID, instance.ID, startedAt, result)
	if updateErr != nil {
		return updated, result, updateErr
	}
	if runErr != nil {
		return updated, result, runErr
	}
	return updated, result, nil
}

func (a *App) finishExecution(ctx context.Context, executionID string, instanceID string, startedAt time.Time, result apperuntime.TriggerResult) (domain.InstanceExecution, error) {
	finishedAt := result.FinishedAt
	if finishedAt.IsZero() {
		finishedAt = time.Now().UTC()
	}
	duration := result.DurationMS
	if duration == 0 {
		duration = int(finishedAt.Sub(startedAt).Milliseconds())
	}
	output := result.RawOutput
	if output == "" && result.Output != nil {
		if data, err := json.Marshal(result.Output); err == nil {
			output = string(data)
		}
	}
	errorMessage := result.Error
	destinationSent := result.DestinationSent
	fallbackUsed := result.FallbackUsed
	inputPayload := result.InputPayload
	outputPayload := result.Output
	logs := result.Logs
	destinationDetails := result.DestinationDetails
	if len(logs) == 0 {
		logs = []domain.ExecutionLog{{Level: "info", Message: "Execution finished", Timestamp: finishedAt}}
	}
	updated, err := a.Executions.UpdateStatus(ctx, executionID, sqlite.UpdateInstanceExecutionStatusParams{
		Status:             result.Status,
		FinishedAt:         &finishedAt,
		DurationMS:         &duration,
		ErrorMessage:       &errorMessage,
		Output:             &output,
		DestinationSent:    &destinationSent,
		FallbackUsed:       &fallbackUsed,
		InputPayload:       &inputPayload,
		OutputPayload:      &outputPayload,
		Logs:               &logs,
		DestinationDetails: &destinationDetails,
	})
	if result.Status == domain.ExecutionSuccess {
		_, _ = a.Instances.UpdateStatus(ctx, instanceID, domain.InstanceStatusIdle)
	} else {
		_, _ = a.Instances.UpdateStatus(ctx, instanceID, domain.InstanceStatusError)
	}
	a.enqueueExecutionCloudSync(ctx, updated, result)
	return updated, err
}

func firstString(payload map[string]any, keys ...string) string {
	for _, key := range keys {
		if value, ok := payload[key].(string); ok && strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func mapValue(payload map[string]any, keys ...string) map[string]any {
	for _, key := range keys {
		if value, ok := payload[key].(map[string]any); ok {
			return value
		}
	}
	return map[string]any{}
}

func restartCommand() (string, []string) {
	if goruntime.GOOS == "windows" {
		return "shutdown", []string{"/r", "/t", "5"}
	}
	return "sudo", []string{"reboot"}
}

func (a *App) enqueueExecutionCloudSync(ctx context.Context, execution domain.InstanceExecution, result apperuntime.TriggerResult) {
	if a == nil || a.CloudSync == nil {
		return
	}
	edgeInfo, err := a.Edge.Load(ctx)
	if err != nil || edgeInfo.EdgeID == "" {
		return
	}
	payload := map[string]any{
		"message_id":       execution.ID,
		"edge_id":          edgeInfo.EdgeID,
		"execution_id":     execution.ID,
		"instance_id":      execution.InstanceID,
		"status":           execution.Status,
		"trigger_type":     execution.TriggerType,
		"occurred_at":      execution.CreatedAt.Format(time.RFC3339),
		"finished_at":      nullableTime(execution.FinishedAt),
		"duration_ms":      execution.DurationMS,
		"error_message":    execution.ErrorMessage,
		"destination_sent": execution.DestinationSent,
		"fallback_used":    execution.FallbackUsed,
		"logs":             execution.Logs,
	}
	if result.Output != nil {
		payload["output_payload"] = result.Output
	}
	if execution.OutputPayload != nil {
		payload["output_payload"] = execution.OutputPayload
	}
	_, _ = a.CloudSync.EnqueueInstanceExecution(ctx, payload)
}

func (a *App) enqueueRemoteJobCloudSync(ctx context.Context, jobID string, action string, status string, payload map[string]any, err error) error {
	if a == nil || a.CloudSync == nil {
		return nil
	}
	edgeInfo, loadErr := a.Edge.Load(ctx)
	if loadErr != nil || edgeInfo.EdgeID == "" {
		return loadErr
	}
	eventPayload := map[string]any{
		"message_id":  fmt.Sprintf("%s:%s", jobID, status),
		"edge_id":     edgeInfo.EdgeID,
		"job_id":      jobID,
		"action":      action,
		"status":      status,
		"occurred_at": time.Now().UTC().Format(time.RFC3339),
	}
	for key, value := range payload {
		if _, exists := eventPayload[key]; !exists {
			eventPayload[key] = value
		}
	}
	if commandID := firstString(payload, "command_id"); commandID != "" {
		eventPayload["command_id"] = commandID
	}
	if err != nil {
		eventPayload["error"] = err.Error()
	}
	_, enqueueErr := a.CloudSync.EnqueueRemoteJob(ctx, eventPayload)
	return enqueueErr
}

func nullableTime(value *time.Time) any {
	if value == nil || value.IsZero() {
		return nil
	}
	return value.Format(time.RFC3339)
}

func emptyStringToNil(value string) any {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return value
}

func envOrDefault(key string, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func remoteJobPayload(payload map[string]any, status string, step string, progress int) map[string]any {
	result := map[string]any{
		"status":       status,
		"current_step": step,
		"progress_pct": progress,
	}
	for key, value := range payload {
		result[key] = value
	}
	return result
}

func intValue(payload map[string]any, keys ...string) int {
	for _, key := range keys {
		switch value := payload[key].(type) {
		case int:
			return value
		case int32:
			return int(value)
		case int64:
			return int(value)
		case float64:
			return int(value)
		}
	}
	return 0
}

func boolValue(payload map[string]any, keys ...string) bool {
	for _, key := range keys {
		if value, ok := payload[key].(bool); ok {
			return value
		}
	}
	return false
}

func dependencyRefs(stepPayload map[string]any) []string {
	raw, ok := stepPayload["depends_on"]
	if !ok {
		raw = stepPayload["dependsOn"]
	}
	values, ok := raw.([]any)
	if !ok {
		return nil
	}
	result := make([]string, 0, len(values))
	for _, value := range values {
		switch typed := value.(type) {
		case string:
			trimmed := strings.TrimSpace(typed)
			if trimmed != "" {
				result = append(result, trimmed)
			}
		case int:
			result = append(result, fmt.Sprintf("%d", typed))
		case int32:
			result = append(result, fmt.Sprintf("%d", typed))
		case int64:
			result = append(result, fmt.Sprintf("%d", typed))
		case float64:
			result = append(result, fmt.Sprintf("%d", int(typed)))
		}
	}
	return result
}

func resolveWorkflowDependencies(steps []workflowStepSpec) error {
	alias := map[string]string{}
	actionCount := map[string]int{}
	for _, step := range steps {
		actionCount[step.Action]++
	}
	for _, step := range steps {
		alias[step.Key] = step.Key
		alias[step.StepName] = step.Key
		alias[fmt.Sprintf("%d", step.Index)] = step.Key
	}
	for _, step := range steps {
		if actionCount[step.Action] == 1 {
			alias[step.Action] = step.Key
		}
	}
	for index := range steps {
		resolved := make([]string, 0, len(steps[index].DependsOn))
		for _, dependency := range steps[index].DependsOn {
			key, ok := alias[dependency]
			if !ok {
				return fmt.Errorf("workflow step %d (%s) has unknown dependency %q", steps[index].Index, steps[index].Action, dependency)
			}
			if key == steps[index].Key {
				return fmt.Errorf("workflow step %d (%s) cannot depend on itself", steps[index].Index, steps[index].Action)
			}
			resolved = append(resolved, key)
		}
		steps[index].DependsOn = resolved
	}
	return nil
}

func dependencyStateError(step workflowStepSpec, states map[string]string) error {
	for _, dependency := range step.DependsOn {
		state, ok := states[dependency]
		if !ok {
			return fmt.Errorf("workflow step %d (%s) is waiting for dependency %s", step.Index, step.Action, dependency)
		}
		if state != "completed" {
			return fmt.Errorf("workflow step %d (%s) blocked by dependency %s with status %s", step.Index, step.Action, dependency, state)
		}
	}
	return nil
}

func dependencyResolutionState(step workflowStepSpec, states map[string]string) (string, error) {
	if len(step.DependsOn) == 0 {
		return "ready", nil
	}
	waiting := false
	for _, dependency := range step.DependsOn {
		state, ok := states[dependency]
		if !ok {
			waiting = true
			continue
		}
		if state != "completed" {
			return "blocked", fmt.Errorf("workflow step %d (%s) blocked by dependency %s with status %s", step.Index, step.Action, dependency, state)
		}
	}
	if waiting {
		return "waiting", nil
	}
	return "ready", nil
}

func blockedStepResult(step workflowStepSpec, err error) map[string]any {
	return map[string]any{
		"index":             step.Index,
		"key":               step.Key,
		"action":            step.Action,
		"status":            "blocked_dependency",
		"depends_on":        step.DependsOn,
		"continue_on_error": step.ContinueOnError,
		"error":             err.Error(),
	}
}

func failedContinueStepResult(step workflowStepSpec, err error) map[string]any {
	return map[string]any{
		"index":             step.Index,
		"key":               step.Key,
		"action":            step.Action,
		"status":            "failed_continue",
		"depends_on":        step.DependsOn,
		"continue_on_error": true,
		"error":             err.Error(),
	}
}
