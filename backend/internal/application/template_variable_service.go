package application

import (
	"context"
	"path/filepath"
	"sync"

	apperrors "backend/internal/application/errors"
	apphandlers "backend/internal/application/handlers"
	"backend/internal/domain"
	"backend/internal/domain/repository"
	"backend/internal/domain/storage"
	"backend/internal/infra/tfparser"
	"backend/pkg/contracts"
	"backend/pkg/errors"
	"backend/pkg/jwt"
	"backend/pkg/validation"

	"github.com/google/uuid"
)

type TemplateVariableService struct {
	templateVarRepo repository.TemplateVariableRepository
	templateRepo    repository.TemplateRepository
	workspaceRepo   repository.WorkspaceRepository
	validator       *validation.Service
	fileStorage     storage.FileStorage
	tfParser        tfparser.TFParser
	uow             apphandlers.UnitOfWork
}

func NewTemplateVariableService(
	templateVarRepo repository.TemplateVariableRepository,
	templateRepo repository.TemplateRepository,
	workspaceRepo repository.WorkspaceRepository,
	validator *validation.Service,
	fileStorage storage.FileStorage,
	tfParser tfparser.TFParser,
	uow apphandlers.UnitOfWork,
) TemplateVariableService {
	return TemplateVariableService{
		templateVarRepo: templateVarRepo,
		templateRepo:    templateRepo,
		workspaceRepo:   workspaceRepo,
		validator:       validator,
		fileStorage:     fileStorage,
		tfParser:        tfParser,
		uow:             uow,
	}
}

func (s TemplateVariableService) verifyTemplateOwnership(ctx context.Context, templateID uuid.UUID) (*domain.Template, *errors.Error) {
	claims, ok := jwt.ClaimsFromContext(ctx)
	if !ok {
		return nil, apperrors.ReturnUnauthorized("missing JWT claims in context")
	}

	template, err := s.templateRepo.GetByID(ctx, templateID)
	if err != nil {
		return nil, err
	}

	if template.WorkspaceID.String() != claims.WorkspaceID {
		return nil, apperrors.ReturnForbidden("template does not belong to your workspace")
	}

	return template, nil
}

func (s TemplateVariableService) CreateVariable(ctx context.Context, request contracts.CreateTemplateVariable) (*domain.TemplateVariable, *errors.Error) {
	if err := s.validator.Validate(request); err != nil {
		return nil, err
	}

	if _, err := s.verifyTemplateOwnership(ctx, request.TemplateID); err != nil {
		return nil, err
	}

	variable := domain.NewTemplateVariable(domain.NewTemplateVariableParams{
		TemplateID:      request.TemplateID,
		Key:             request.Key,
		Description:     request.Description,
		VarType:         request.VarType,
		DefaultValue:    request.DefaultValue,
		IsSensitive:     request.IsSensitive,
		IsRequired:      request.IsRequired,
		ValidationRegex: request.ValidationRegex,
		IsAutoParsed:    false,
	})

	if err := s.templateVarRepo.Create(ctx, *variable); err != nil {
		return nil, err
	}

	return variable, nil
}

func (s TemplateVariableService) ListVariables(ctx context.Context, request contracts.GetTemplateVariables) ([]*domain.TemplateVariable, *errors.Error) {
	if err := s.validator.Validate(request); err != nil {
		return nil, err
	}

	if _, err := s.verifyTemplateOwnership(ctx, request.TemplateID); err != nil {
		return nil, err
	}

	return s.templateVarRepo.GetByTemplateID(ctx, request.TemplateID)
}

func (s TemplateVariableService) UpdateVariable(ctx context.Context, request contracts.UpdateTemplateVariable) (*domain.TemplateVariable, *errors.Error) {
	if err := s.validator.Validate(request); err != nil {
		return nil, err
	}

	existing, err := s.templateVarRepo.GetByID(ctx, request.ID)
	if err != nil {
		return nil, err
	}

	if _, err := s.verifyTemplateOwnership(ctx, existing.TemplateID); err != nil {
		return nil, err
	}

	if request.Description != nil {
		existing.Description = *request.Description
	}
	if request.VarType != nil {
		existing.VarType = *request.VarType
	}
	if request.DefaultValue != nil {
		existing.DefaultValue = *request.DefaultValue
	}
	if request.IsSensitive != nil {
		existing.IsSensitive = *request.IsSensitive
	}
	if request.IsRequired != nil {
		existing.IsRequired = *request.IsRequired
	}
	if request.ValidationRegex != nil {
		existing.ValidationRegex = *request.ValidationRegex
	}
	if request.DisplayOrder != nil {
		existing.DisplayOrder = *request.DisplayOrder
	}

	if err := s.templateVarRepo.Update(ctx, *existing); err != nil {
		return nil, err
	}

	return existing, nil
}

func (s TemplateVariableService) DeleteVariable(ctx context.Context, request contracts.DeleteTemplateVariable) *errors.Error {
	if err := s.validator.Validate(request); err != nil {
		return err
	}

	existing, err := s.templateVarRepo.GetByID(ctx, request.ID)
	if err != nil {
		return err
	}

	if _, err := s.verifyTemplateOwnership(ctx, existing.TemplateID); err != nil {
		return err
	}

	return s.templateVarRepo.Delete(ctx, request.ID)
}

type ParseResult struct {
	Variables []*domain.TemplateVariable `json:"variables"`
	Added     int                        `json:"added"`
	Updated   int                        `json:"updated"`
	Removed   int                        `json:"removed"`
}

func (s TemplateVariableService) ParseAndReconcileVariables(ctx context.Context, request contracts.ParseTemplateVariables) (*ParseResult, *errors.Error) {
	if err := s.validator.Validate(request); err != nil {
		return nil, err
	}

	template, err := s.verifyTemplateOwnership(ctx, request.TemplateID)
	if err != nil {
		return nil, err
	}

	// Read variables.tf from the template's storage directory
	content, readErr := s.fileStorage.ReadFile(template.Path)
	if readErr != nil {
		return nil, apperrors.ReturnBadRequest("no variables.tf found in template: " + filepath.Base(template.Path))
	}

	// Parse the variables
	parsed, parseErr := s.tfParser.ParseVariables(content, "variables.tf")
	if parseErr != nil {
		return nil, apperrors.ReturnBadRequest("failed to parse variables.tf: " + parseErr.Error())
	}

	// Get existing variables from DB
	existing, err := s.templateVarRepo.GetByTemplateID(ctx, request.TemplateID)
	if err != nil {
		return nil, err
	}

	// Build lookup maps
	existingByKey := make(map[string]*domain.TemplateVariable)
	for _, v := range existing {
		existingByKey[v.Key] = v
	}

	parsedKeys := make(map[string]bool)
	for _, p := range parsed {
		parsedKeys[p.Key] = true
	}

	result := &ParseResult{}

	// Collect batch operations
	var toCreate []domain.TemplateVariable
	var toUpdate []domain.TemplateVariable
	var keysToRemove []string

	for _, p := range parsed {
		if existingVar, exists := existingByKey[p.Key]; exists {
			if existingVar.IsAutoParsed {
				existingVar.Description = p.Description
				existingVar.VarType = p.VarType
				existingVar.DefaultValue = p.Default
				existingVar.IsSensitive = p.IsSensitive
				existingVar.IsRequired = p.IsRequired
				toUpdate = append(toUpdate, *existingVar)
			}
		} else {
			variable := domain.NewTemplateVariable(domain.NewTemplateVariableParams{
				TemplateID:   request.TemplateID,
				Key:          p.Key,
				Description:  p.Description,
				VarType:      p.VarType,
				DefaultValue: p.Default,
				IsSensitive:  p.IsSensitive,
				IsRequired:   p.IsRequired,
				IsAutoParsed: true,
			})
			toCreate = append(toCreate, *variable)
		}
	}

	for key, v := range existingByKey {
		if !parsedKeys[key] && v.IsAutoParsed {
			keysToRemove = append(keysToRemove, key)
		}
	}

	// Execute all mutations in a single transaction with concurrent goroutines
	if err := s.uow.Begin(); err != nil {
		return nil, err
	}
	defer s.uow.Rollback()

	var wg sync.WaitGroup
	errs := make(chan *errors.Error, 3)

	wg.Add(3)
	go func() {
		defer wg.Done()
		err := s.templateVarRepo.CreateBatch(ctx, toCreate)
		errs <- err
	}()
	go func() {
		defer wg.Done()
		err := s.templateVarRepo.UpdateBatch(ctx, toUpdate)
		errs <- err
	}()
	go func() {
		defer wg.Done()
		err := s.templateVarRepo.DeleteByTemplateIDAndKeys(ctx, request.TemplateID, keysToRemove)
		errs <- err
	}()
	wg.Wait()

	for i := 0; i < 3; i++ {
		select {
		case err := <-errs:
			if err != nil {
				s.uow.Rollback()
				return nil, err
			}
		case <-ctx.Done():
			s.uow.Rollback()
			return nil, apperrors.ReturnInternalError("operation timed out")
		}
	}

	if err := s.uow.Commit(); err != nil {
		return nil, err
	}

	result.Added = len(toCreate)
	result.Updated = len(toUpdate)
	result.Removed = len(keysToRemove)

	// Fetch the final state
	finalVars, err := s.templateVarRepo.GetByTemplateID(ctx, request.TemplateID)
	if err != nil {
		return nil, err
	}
	result.Variables = finalVars

	return result, nil
}
