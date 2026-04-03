package application

import (
	"context"
	"log/slog"
	"time"

	"backend/internal/domain"
	"backend/internal/domain/repository"
	"backend/internal/domain/storage"
	"backend/internal/infra/terraform"
)

type EnvironmentReaper struct {
	queueRepo        repository.TeardownQueueRepository
	envRepo          repository.EnvironmentRepository
	executionStorage storage.ExecutionStorage
	tfExecutor       *terraform.Executor
	envVarService    EnvironmentVariableValueService
	interval         time.Duration
}

func NewEnvironmentReaper(
	queueRepo repository.TeardownQueueRepository,
	envRepo repository.EnvironmentRepository,
	executionStorage storage.ExecutionStorage,
	tfExecutor *terraform.Executor,
	envVarService EnvironmentVariableValueService,
	interval time.Duration,
) *EnvironmentReaper {
	return &EnvironmentReaper{
		queueRepo:        queueRepo,
		envRepo:          envRepo,
		executionStorage: executionStorage,
		tfExecutor:       tfExecutor,
		envVarService:    envVarService,
		interval:         interval,
	}
}

func (r *EnvironmentReaper) Start(ctx context.Context) {
	r.recoverFromCrash(ctx)

	ticker := time.NewTicker(r.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			slog.Info("reaper: shutting down")
			return
		case <-ticker.C:
			r.processNext(ctx)
		}
	}
}

func (r *EnvironmentReaper) recoverFromCrash(ctx context.Context) {
	if err := r.queueRepo.ResetProcessing(ctx); err != nil {
		slog.Error("reaper: failed to reset processing entries", "error", err)
	}
}

func (r *EnvironmentReaper) processNext(ctx context.Context) {
	entry, err := r.queueRepo.FindDue(ctx, time.Now().UTC())
	if err != nil {
		slog.Error("reaper: failed to find due entry", "error", err)
		return
	}
	if entry == nil {
		return
	}

	env, err := r.envRepo.GetByID(ctx, entry.EnvironmentID)
	if err != nil {
		slog.Warn("reaper: environment not found, marking completed", "env_id", entry.EnvironmentID)
		r.queueRepo.UpdateStatus(ctx, entry.EnvironmentID, domain.TeardownStatusCompleted)
		return
	}

	if env.Status == domain.EnvironmentStatusDestroyed {
		r.queueRepo.UpdateStatus(ctx, entry.EnvironmentID, domain.TeardownStatusCompleted)
		return
	}

	if err := env.CanStartOperation(); err != nil {
		slog.Info("reaper: environment busy, will retry", "env_id", env.ID, "status", env.Status)
		return
	}

	if err := r.queueRepo.UpdateStatus(ctx, entry.EnvironmentID, domain.TeardownStatusProcessing); err != nil {
		slog.Error("reaper: failed to mark processing", "env_id", env.ID, "error", err)
		return
	}

	env, acquireErr := r.envRepo.AcquireOperation(ctx, env.ID, domain.EnvironmentStatusDestroying)
	if acquireErr != nil {
		slog.Info("reaper: failed to acquire operation, resetting to pending", "env_id", entry.EnvironmentID)
		r.queueRepo.UpdateStatus(ctx, entry.EnvironmentID, domain.TeardownStatusPending)
		return
	}

	go r.executeDestroy(env)
}

func (r *EnvironmentReaper) executeDestroy(env *domain.Environment) {
	ctx := context.Background()

	if err := r.writeVarsFile(ctx, env); err != nil {
		env.Status = domain.EnvironmentStatusError
		env.LastError = err.Error()
		slog.Error("reaper: failed to write tfvars", "env_id", env.ID, "error", err)
		r.envRepo.Update(ctx, env)
		r.queueRepo.UpdateStatus(ctx, env.ID, domain.TeardownStatusPending)
		return
	}

	_, tfErr := r.tfExecutor.Destroy(ctx, env.ExecutionPath())

	if tfErr != nil {
		env.Status = domain.EnvironmentStatusError
		env.LastError = tfErr.Error()
		env.LastOperation = "destroy"
		slog.Error("reaper: terraform destroy failed", "env_id", env.ID, "error", tfErr)
		r.envRepo.Update(ctx, env)
		r.queueRepo.UpdateStatus(ctx, env.ID, domain.TeardownStatusPending)
	} else {
		env.Status = domain.EnvironmentStatusDestroyed
		env.LastError = ""
		env.LastOperation = "destroy"
		slog.Info("reaper: environment destroyed", "env_id", env.ID)
		r.envRepo.Update(ctx, env)
		r.queueRepo.UpdateStatus(ctx, env.ID, domain.TeardownStatusCompleted)
	}
}

func (r *EnvironmentReaper) writeVarsFile(ctx context.Context, env *domain.Environment) error {
	nonsensitive, sensitive, err := r.envVarService.GetDecryptedValues(ctx, env.ID)
	if err != nil {
		return err
	}

	merged := make(map[string]string, len(nonsensitive)+len(sensitive))
	for k, v := range nonsensitive {
		merged[k] = v
	}
	for k, v := range sensitive {
		merged[k] = v
	}

	if len(merged) == 0 {
		return nil
	}

	content := storage.FormatTFVars(merged)
	return r.executionStorage.WriteVarsFile(env.ExecutionPath(), content)
}
