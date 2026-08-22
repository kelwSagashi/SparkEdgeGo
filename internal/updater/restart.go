package updater

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

type RestartResult struct {
	Executed       bool      `json:"executed"`
	ManualRequired bool      `json:"manual_required"`
	Command        string    `json:"command,omitempty"`
	Message        string    `json:"message"`
	UpdatedAt      time.Time `json:"updated_at"`
}

func (s *Service) Restart(_ context.Context, execute bool) (RestartResult, error) {
	state, _ := s.LoadState()
	command := strings.TrimSpace(s.config.RestartCommand)
	if command == "" {
		command = strings.TrimSpace(defaultRestartCommand(s.config.ServiceName))
	}

	result := RestartResult{
		Command:   command,
		UpdatedAt: time.Now().UTC(),
	}

	if command == "" {
		result.ManualRequired = true
		result.Message = "Nenhum comando de reinicio foi configurado. Reinicie o SparkEdge manualmente."
		_ = s.saveStateWithHistory(state, UpdateState{
			LastDownloadedPackage: state.LastDownloadedPackage,
			LastPreparedVersion:   state.LastPreparedVersion,
			LastPreparedTarget:    state.LastPreparedTarget,
			LastApplyResult:       state.LastApplyResult,
			LastDownloadResult:    state.LastDownloadResult,
			LastExecuteResult:     state.LastExecuteResult,
			LastRollbackResult:    state.LastRollbackResult,
			LastRestartResult:     &result,
		}, HistoryEntry{
			Type:      "restart",
			Status:    "manual_required",
			Version:   state.LastPreparedVersion,
			Target:    state.LastPreparedTarget,
			Message:   result.Message,
			Artifact:  command,
			CreatedAt: result.UpdatedAt,
		})
		return result, nil
	}

	if !execute {
		result.ManualRequired = true
		result.Message = "Plano de reinicio gerado. Execute o comando retornado quando quiser concluir o update."
		_ = s.saveStateWithHistory(state, UpdateState{
			LastDownloadedPackage: state.LastDownloadedPackage,
			LastPreparedVersion:   state.LastPreparedVersion,
			LastPreparedTarget:    state.LastPreparedTarget,
			LastApplyResult:       state.LastApplyResult,
			LastDownloadResult:    state.LastDownloadResult,
			LastExecuteResult:     state.LastExecuteResult,
			LastRollbackResult:    state.LastRollbackResult,
			LastRestartResult:     &result,
		}, HistoryEntry{
			Type:      "restart",
			Status:    "planned",
			Version:   state.LastPreparedVersion,
			Target:    state.LastPreparedTarget,
			Message:   result.Message,
			Artifact:  command,
			CreatedAt: result.UpdatedAt,
		})
		return result, nil
	}

	if err := startBackgroundCommand(command); err != nil {
		return RestartResult{}, err
	}
	result.Executed = true
	result.Message = "Comando de reinicio disparado com sucesso."
	_ = s.saveStateWithHistory(state, UpdateState{
		LastDownloadedPackage: state.LastDownloadedPackage,
		LastPreparedVersion:   state.LastPreparedVersion,
		LastPreparedTarget:    state.LastPreparedTarget,
		LastApplyResult:       state.LastApplyResult,
		LastDownloadResult:    state.LastDownloadResult,
		LastExecuteResult:     state.LastExecuteResult,
		LastRollbackResult:    state.LastRollbackResult,
		LastRestartResult:     &result,
	}, HistoryEntry{
		Type:      "restart",
		Status:    "executed",
		Version:   state.LastPreparedVersion,
		Target:    state.LastPreparedTarget,
		Message:   result.Message,
		Artifact:  command,
		CreatedAt: result.UpdatedAt,
	})
	return result, nil
}

func defaultRestartCommand(serviceName string) string {
	serviceName = strings.TrimSpace(serviceName)
	if serviceName == "" {
		return ""
	}
	switch runtime.GOOS {
	case "windows":
		return fmt.Sprintf("Restart-Service -Name '%s'", strings.ReplaceAll(serviceName, "'", "''"))
	case "linux":
		return fmt.Sprintf("systemctl restart %s", serviceName)
	default:
		return ""
	}
}

func startBackgroundCommand(command string) error {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", command)
	} else {
		cmd = exec.Command("sh", "-c", command)
	}
	return cmd.Start()
}
