package application

import (
	"context"
	"log/slog"
	"time"

	apperrors "backend/internal/application/errors"
	"backend/internal/domain"
	"backend/internal/domain/repository"
	"backend/internal/domain/storage"
	"backend/internal/infra/terraform"
	"backend/pkg/contracts"
	"backend/pkg/errors"
	"backend/pkg/jwt"
	"backend/pkg/validation"

	"github.com/google/uuid"
)

type EnvironmentService struct {
	envRepo          repository.EnvironmentRepository
	templateRepo     repository.TemplateRepository
	userRepo         repository.UserRepository
	validator        *validation.Service
	executionStorage storage.ExecutionStorage
	tfExecutor       *terraform.Executor
	envVarService    EnvironmentVariableValueService
	teardownRepo     repository.TeardownQueueRepository
}

func NewEnvironmentService(
	envRepo repository.EnvironmentRepository,
	templateRepo repository.TemplateRepository,
	userRepo repository.UserRepository,
	validator *validation.Service,
	executionStorage storage.ExecutionStorage,
	tfExecutor *terraform.Executor,
	envVarService EnvironmentVariableValueService,
	teardownRepo repository.TeardownQueueRepository,
) EnvironmentService {
	return EnvironmentService{
		envRepo:          envRepo,
		templateRepo:     templateRepo,
		validator:        validator,
		executionStorage: executionStorage,
		tfExecutor:       tfExecutor,
		userRepo:         userRepo,
		envVarService:    envVarService,
		teardownRepo:     teardownRepo,
	}
}

func (s EnvironmentService) verifyEnvironmentOwnership(ctx context.Context, envID uuid.UUID) (*domain.Environment, *errors.Error) {
	var err *errors.Error
	claims, ok := jwt.ClaimsFromContext(ctx)
	if !ok {
		return nil, apperrors.ReturnUnauthorized("missing JWT claims in context")
	}

	env, err := s.envRepo.GetByID(ctx, envID)
	if err != nil {
		return nil, apperrors.ReturnNotFound("environment not found")
	}

	if env.WorkspaceID.String() != claims.WorkspaceID {
		return nil, apperrors.ReturnForbidden("environment does not belong to your workspace")
	}

	return env, nil
}

// CreateEnvironment creates a new environment from a template, copies the
// template files to an execution directory, and runs terraform init in the background.
func (s EnvironmentService) CreateEnvironment(ctx context.Context, request contracts.CreateEnvironment) (*domain.Environment, *errors.Error) {
	claims, ok := jwt.ClaimsFromContext(ctx)
	if !ok {
		return nil, apperrors.ReturnUnauthorized("missing JWT claims in context")
	}

	if err := s.validator.Validate(request); err != nil {
		return nil, err
	}

	workspaceID, _ := uuid.Parse(claims.WorkspaceID)

	// Verify template exists and belongs to workspace.
	template, repoErr := s.templateRepo.GetByID(ctx, request.TemplateID)
	if repoErr != nil {
		return nil, repoErr
	}
	if template.WorkspaceID != workspaceID {
		return nil, apperrors.ReturnForbidden("template does not belong to your workspace")
	}

	createdBy, _ := uuid.Parse(claims.ID)
	env := domain.NewEnvironment(request.Name, request.Description, createdBy, workspaceID, request.TemplateID, request.TTLSeconds)

	// Copy template files into execution directory.
	if err := s.executionStorage.CopyTemplateToExecution(template.Path, env.ExecutionPath()); err != nil {
		return nil, err
	}

	// Persist to DB; on failure, cleanup execution dir.
	if repoErr := s.envRepo.Create(ctx, env); repoErr != nil {
		if cleanupErr := s.executionStorage.DeleteDir(env.ExecutionPath()); cleanupErr != nil {
			slog.Error("failed to cleanup execution dir after DB error", "path", env.ExecutionPath(), "error", cleanupErr)
		}
		return nil, apperrors.ReturnInternalError("failed to create environment: " + repoErr.Error())
	}

	// Enqueue auto-teardown if TTL is set.
	if env.TTLSeconds != nil {
		entry := &domain.TeardownEntry{
			EnvironmentID: env.ID,
			TeardownAt:    env.CreatedAt.Add(time.Duration(*env.TTLSeconds) * time.Second),
			Status:        domain.TeardownStatusPending,
		}
		if enqueueErr := s.teardownRepo.Enqueue(ctx, entry); enqueueErr != nil {
			slog.Error("failed to enqueue teardown entry", "env_id", env.ID, "error", enqueueErr)
		}
	}

	// Run terraform init in the background.
	go s.runInit(env)

	return env, nil
}

// GetEnvironment retrieves an environment by ID.
func (s EnvironmentService) GetEnvironment(ctx context.Context, request contracts.GetEnvironment) (*domain.Environment, *errors.Error) {
	claims, ok := jwt.ClaimsFromContext(ctx)
	if !ok {
		return nil, apperrors.ReturnUnauthorized("missing JWT claims in context")
	}

	if err := s.validator.Validate(request); err != nil {
		return nil, err
	}

	env, repoErr := s.envRepo.GetByID(ctx, request.ID)
	if repoErr != nil {
		return nil, apperrors.ReturnNotFound("environment not found")
	}

	if env.WorkspaceID.String() != claims.WorkspaceID {
		return nil, apperrors.ReturnForbidden("environment does not belong to your workspace")
	}

	return env, nil
}

// GetEnvironmentsByWorkspace retrieves all environments for a workspace.
func (s EnvironmentService) GetEnvironmentsByWorkspace(ctx context.Context) ([]*domain.Environment, *errors.Error) {
	claims, ok := jwt.ClaimsFromContext(ctx)
	if !ok {
		return nil, apperrors.ReturnUnauthorized("missing JWT claims in context")
	}

	workspaceID, _ := uuid.Parse(claims.WorkspaceID)

	envs, repoErr := s.envRepo.GetByWorkspaceID(ctx, workspaceID)
	if repoErr != nil {
		return nil, apperrors.ReturnInternalError("failed to list environments")
	}

	return envs, nil
}

// PlanEnvironment runs terraform plan on the environment.
func (s EnvironmentService) PlanEnvironment(ctx context.Context, request contracts.PlanEnvironment) (*domain.Environment, *errors.Error) {
	if err := s.validator.Validate(request); err != nil {
		return nil, err
	}
	return s.startOperation(ctx, request.ID, domain.EnvironmentStatusPlanning)
}

// ApplyEnvironment runs terraform apply on the environment.
func (s EnvironmentService) ApplyEnvironment(ctx context.Context, request contracts.ApplyEnvironment) (*domain.Environment, *errors.Error) {
	if err := s.validator.Validate(request); err != nil {
		return nil, err
	}
	return s.startOperation(ctx, request.ID, domain.EnvironmentStatusApplying)
}

// DestroyEnvironment runs terraform destroy on the environment.
func (s EnvironmentService) DestroyEnvironment(ctx context.Context, request contracts.DestroyEnvironment) (*domain.Environment, *errors.Error) {
	if err := s.validator.Validate(request); err != nil {
		return nil, err
	}
	return s.startOperation(ctx, request.ID, domain.EnvironmentStatusDestroying)
}

// DeleteEnvironment deletes the environment from the database and cleans up
// the execution directory. Rejects if an operation is in progress.
func (s EnvironmentService) DeleteEnvironment(ctx context.Context, request contracts.DeleteEnvironment) *errors.Error {
	env, err := s.verifyEnvironmentOwnership(ctx, request.ID)
	if err != nil {
		return err
	}

	if err := env.CanStartOperation(); err != nil {
		return apperrors.ReturnConflict(err.Error())
	}

	if repoErr := s.envRepo.Delete(ctx, request.ID); repoErr != nil {
		return apperrors.ReturnInternalError("failed to delete environment")
	}

	// Best-effort cleanup of execution directory.
	if cleanupErr := s.executionStorage.DeleteDir(env.ExecutionPath()); cleanupErr != nil {
		slog.Error("failed to cleanup execution dir after delete", "path", env.ExecutionPath(), "error", cleanupErr)
	}

	return nil
}

// startOperation acquires the atomic lock and dispatches the terraform command
// in a background goroutine.
func (s EnvironmentService) startOperation(ctx context.Context, envID uuid.UUID, status domain.EnvironmentStatus) (*domain.Environment, *errors.Error) {
	env, err := s.verifyEnvironmentOwnership(ctx, envID)
	if err != nil {
		return nil, err
	}
	// Atomic status acquisition — rejects if already in a blocking state.
	env, acquireErr := s.envRepo.AcquireOperation(ctx, env.ID, status)
	if acquireErr != nil {
		return nil, apperrors.ReturnConflict(acquireErr.Error())
	}

	go s.executeTerraform(env, status)

	return env, nil
}

func (s EnvironmentService) executeTerraform(env *domain.Environment, status domain.EnvironmentStatus) {
	ctx := context.Background()

	// Write terraform.tfvars with decrypted variable values before running any command.
	if err := s.writeVarsFile(ctx, env); err != nil {
		env.Status = domain.EnvironmentStatusError
		env.LastError = err.Error()
		slog.Error("failed to write tfvars", "env_id", env.ID, "error", err)
		if updateErr := s.envRepo.Update(ctx, env); updateErr != nil {
			slog.Error("failed to update environment after tfvars error", "env_id", env.ID, "error", updateErr)
		}
		return
	}

	var tfErr error

	switch status {
	case domain.EnvironmentStatusPlanning:
		_, tfErr = s.tfExecutor.Plan(ctx, env.ExecutionPath())
	case domain.EnvironmentStatusApplying:
		_, tfErr = s.tfExecutor.Apply(ctx, env.ExecutionPath())
	case domain.EnvironmentStatusDestroying:
		_, tfErr = s.tfExecutor.Destroy(ctx, env.ExecutionPath())
	}

	s.completeOperation(env, status, tfErr)
}

func (s EnvironmentService) writeVarsFile(ctx context.Context, env *domain.Environment) *errors.Error {
	nonsensitive, sensitive, err := s.envVarService.GetDecryptedValues(ctx, env.ID)
	if err != nil {
		return err
	}

	// Merge both maps into one for tfvars output.
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
	return s.executionStorage.WriteVarsFile(env.ExecutionPath(), content)
}

func (s EnvironmentService) completeOperation(env *domain.Environment, status domain.EnvironmentStatus, tfErr error) {
	ctx := context.Background()

	if tfErr != nil {
		env.Status = domain.EnvironmentStatusError
		env.LastError = tfErr.Error()
		slog.Error("terraform operation failed",
			"env_id", env.ID,
			"operation", operationName(status),
			"error", tfErr,
		)
	} else {
		env.LastError = ""
		switch status {
		case domain.EnvironmentStatusPlanning:
			// Plan is read-only; revert to the previous stable state.
			if env.LastAppliedAt != nil {
				env.Status = domain.EnvironmentStatusReady
			} else {
				env.Status = domain.EnvironmentStatusInitialized
			}
		case domain.EnvironmentStatusApplying:
			env.Status = domain.EnvironmentStatusReady
			now := time.Now().UTC()
			env.LastAppliedAt = &now
		case domain.EnvironmentStatusDestroying:
			env.Status = domain.EnvironmentStatusDestroyed
		}
	}

	if err := s.envRepo.Update(ctx, env); err != nil {
		slog.Error("failed to update environment after terraform operation",
			"env_id", env.ID, "error", err,
		)
	}
}

func (s EnvironmentService) runInit(env *domain.Environment) {
	ctx := context.Background()
	_, err := s.tfExecutor.Init(ctx, env.ExecutionPath())

	if err != nil {
		env.Status = domain.EnvironmentStatusError
		env.LastError = err.Error()
		env.LastOperation = "init"
		slog.Error("terraform init failed", "env_id", env.ID, "error", err)
	} else {
		env.Status = domain.EnvironmentStatusInitialized
		env.LastOperation = "init"
	}

	if updateErr := s.envRepo.Update(ctx, env); updateErr != nil {
		slog.Error("failed to update environment after init", "env_id", env.ID, "error", updateErr)
	}
}

func operationName(s domain.EnvironmentStatus) string {
	switch s {
	case domain.EnvironmentStatusPlanning:
		return "plan"
	case domain.EnvironmentStatusApplying:
		return "apply"
	case domain.EnvironmentStatusDestroying:
		return "destroy"
	default:
		return string(s)
	}
}
