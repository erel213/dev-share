package application

import (
	"context"
	"regexp"

	apperrors "backend/internal/application/errors"
	"backend/internal/domain"
	"backend/internal/domain/repository"
	"backend/pkg/contracts"
	"backend/pkg/crypto"
	"backend/pkg/errors"
	"backend/pkg/jwt"
	"backend/pkg/validation"

	"github.com/google/uuid"
)

const sensitiveValueMask = "********"

type EnvironmentVariableValueService struct {
	envVarRepo      repository.EnvironmentVariableValueRepository
	templateVarRepo repository.TemplateVariableRepository
	environmentRepo repository.EnvironmentRepository
	encryptor       crypto.Encryptor
	validator       *validation.Service
}

func NewEnvironmentVariableValueService(
	envVarRepo repository.EnvironmentVariableValueRepository,
	templateVarRepo repository.TemplateVariableRepository,
	environmentRepo repository.EnvironmentRepository,
	encryptor crypto.Encryptor,
	validator *validation.Service,
) EnvironmentVariableValueService {
	return EnvironmentVariableValueService{
		envVarRepo:      envVarRepo,
		templateVarRepo: templateVarRepo,
		environmentRepo: environmentRepo,
		encryptor:       encryptor,
		validator:       validator,
	}
}

func (s EnvironmentVariableValueService) verifyEnvironmentOwnership(ctx context.Context, environmentID uuid.UUID) (*domain.Environment, *errors.Error) {
	claims, ok := jwt.ClaimsFromContext(ctx)
	if !ok {
		return nil, apperrors.ReturnUnauthorized("missing JWT claims in context")
	}

	env, stdErr := s.environmentRepo.GetByID(ctx, environmentID)
	if stdErr != nil {
		return nil, apperrors.ReturnNotFound("environment not found")
	}

	if env.WorkspaceID.String() != claims.WorkspaceID {
		return nil, apperrors.ReturnForbidden("environment does not belong to your workspace")
	}

	return env, nil
}

func (s EnvironmentVariableValueService) SetVariableValues(ctx context.Context, request contracts.SetEnvironmentVariableValues) *errors.Error {
	if err := s.validator.Validate(request); err != nil {
		return err
	}

	env, err := s.verifyEnvironmentOwnership(ctx, request.EnvironmentID)
	if err != nil {
		return err
	}

	// Get template variables for this environment's template
	templateVars, repoErr := s.templateVarRepo.GetByTemplateID(ctx, env.TemplateID)
	if repoErr != nil {
		return repoErr
	}

	// Build lookup
	templateVarByID := make(map[uuid.UUID]*domain.TemplateVariable)
	for _, tv := range templateVars {
		templateVarByID[tv.ID] = tv
	}

	// Build map of provided values
	providedValues := make(map[uuid.UUID]string)
	for _, entry := range request.Values {
		providedValues[entry.TemplateVariableID] = entry.Value
	}

	// Validate required variables and regex
	for _, tv := range templateVars {
		value, provided := providedValues[tv.ID]

		if tv.IsRequired && tv.DefaultValue == "" && !provided {
			return apperrors.ReturnBadRequest("required variable missing: " + tv.Key)
		}

		if provided && tv.ValidationRegex != "" {
			matched, regexErr := regexp.MatchString(tv.ValidationRegex, value)
			if regexErr != nil {
				return apperrors.ReturnBadRequest("invalid validation regex for variable: " + tv.Key)
			}
			if !matched {
				return apperrors.ReturnBadRequest("value for variable " + tv.Key + " does not match validation pattern")
			}
		}
	}

	// Build values to upsert
	var values []domain.EnvironmentVariableValue
	for _, entry := range request.Values {
		tv, exists := templateVarByID[entry.TemplateVariableID]
		if !exists {
			return apperrors.ReturnBadRequest("template variable not found: " + entry.TemplateVariableID.String())
		}

		finalValue := entry.Value
		if tv.IsSensitive {
			encrypted, encErr := s.encryptor.Encrypt(finalValue)
			if encErr != nil {
				return apperrors.ReturnInternalError("failed to encrypt sensitive variable: " + tv.Key)
			}
			finalValue = encrypted
		}

		values = append(values, *domain.NewEnvironmentVariableValue(
			request.EnvironmentID,
			entry.TemplateVariableID,
			finalValue,
		))
	}

	return s.envVarRepo.UpsertBatch(ctx, values)
}

type EnvironmentVariableValueResponse struct {
	TemplateVariableID uuid.UUID `json:"template_variable_id"`
	Key                string    `json:"key"`
	Value              string    `json:"value"`
	IsSensitive        bool      `json:"is_sensitive"`
}

func (s EnvironmentVariableValueService) GetVariableValues(ctx context.Context, request contracts.GetEnvironmentVariableValues) ([]EnvironmentVariableValueResponse, *errors.Error) {
	if err := s.validator.Validate(request); err != nil {
		return nil, err
	}

	env, err := s.verifyEnvironmentOwnership(ctx, request.EnvironmentID)
	if err != nil {
		return nil, err
	}

	// Get template variables
	templateVars, repoErr := s.templateVarRepo.GetByTemplateID(ctx, env.TemplateID)
	if repoErr != nil {
		return nil, repoErr
	}
	templateVarByID := make(map[uuid.UUID]*domain.TemplateVariable)
	for _, tv := range templateVars {
		templateVarByID[tv.ID] = tv
	}

	// Get stored values
	storedValues, repoErr := s.envVarRepo.GetByEnvironmentID(ctx, request.EnvironmentID)
	if repoErr != nil {
		return nil, repoErr
	}

	var response []EnvironmentVariableValueResponse
	for _, sv := range storedValues {
		tv, exists := templateVarByID[sv.TemplateVariableID]
		if !exists {
			continue
		}

		displayValue := sv.Value
		if tv.IsSensitive {
			displayValue = sensitiveValueMask
		}

		response = append(response, EnvironmentVariableValueResponse{
			TemplateVariableID: sv.TemplateVariableID,
			Key:                tv.Key,
			Value:              displayValue,
			IsSensitive:        tv.IsSensitive,
		})
	}

	return response, nil
}

// GetDecryptedValues returns decrypted variable values for deployment flow.
// Returns separate maps for non-sensitive and sensitive values.
func (s EnvironmentVariableValueService) GetDecryptedValues(ctx context.Context, environmentID uuid.UUID) (nonsensitive map[string]string, sensitive map[string]string, retErr *errors.Error) {
	env, stdErr := s.environmentRepo.GetByID(ctx, environmentID)
	if stdErr != nil {
		return nil, nil, apperrors.ReturnNotFound("environment not found")
	}

	templateVars, err := s.templateVarRepo.GetByTemplateID(ctx, env.TemplateID)
	if err != nil {
		return nil, nil, err
	}
	templateVarByID := make(map[uuid.UUID]*domain.TemplateVariable)
	for _, tv := range templateVars {
		templateVarByID[tv.ID] = tv
	}

	storedValues, err := s.envVarRepo.GetByEnvironmentID(ctx, environmentID)
	if err != nil {
		return nil, nil, err
	}

	nonsensitive = make(map[string]string)
	sensitive = make(map[string]string)

	for _, sv := range storedValues {
		tv, exists := templateVarByID[sv.TemplateVariableID]
		if !exists {
			continue
		}

		if tv.IsSensitive {
			decrypted, decErr := s.encryptor.Decrypt(sv.Value)
			if decErr != nil {
				return nil, nil, apperrors.ReturnInternalError("failed to decrypt variable: " + tv.Key)
			}
			sensitive[tv.Key] = decrypted
		} else {
			nonsensitive[tv.Key] = sv.Value
		}
	}

	return nonsensitive, sensitive, nil
}
