package application

import (
	"context"

	apperrors "backend/internal/application/errors"
	"backend/internal/domain"
	"backend/internal/domain/repository"
	"backend/pkg/errors"

	"github.com/google/uuid"
)

// GetAccessibleTemplates returns the templates a user can access in a workspace
// based on their group memberships. Admins bypass group checks and get all templates.
func GetAccessibleTemplates(
	ctx context.Context,
	groupRepo repository.GroupRepository,
	templateRepo repository.TemplateRepository,
	userID uuid.UUID,
	workspaceID uuid.UUID,
	isAdmin bool,
) ([]*domain.Template, *errors.Error) {
	if isAdmin {
		return templateRepo.GetByWorkspaceID(ctx, workspaceID)
	}

	accessibleIDs, hasAccessAll, err := groupRepo.GetAccessibleTemplateIDs(ctx, userID, workspaceID)
	if err != nil {
		return nil, apperrors.ReturnInternalError("failed to check template access")
	}

	if hasAccessAll {
		return templateRepo.GetByWorkspaceID(ctx, workspaceID)
	}

	if len(accessibleIDs) == 0 {
		return []*domain.Template{}, nil
	}

	allTemplates, repoErr := templateRepo.GetByWorkspaceID(ctx, workspaceID)
	if repoErr != nil {
		return nil, repoErr
	}

	accessSet := make(map[uuid.UUID]struct{}, len(accessibleIDs))
	for _, id := range accessibleIDs {
		accessSet[id] = struct{}{}
	}

	var filtered []*domain.Template
	for _, t := range allTemplates {
		if _, ok := accessSet[t.ID]; ok {
			filtered = append(filtered, t)
		}
	}

	return filtered, nil
}

// CanAccessTemplate checks whether a user can access a specific template
// based on their group memberships. Admins always have access.
func CanAccessTemplate(
	ctx context.Context,
	groupRepo repository.GroupRepository,
	userID uuid.UUID,
	workspaceID uuid.UUID,
	templateID uuid.UUID,
	isAdmin bool,
) (bool, *errors.Error) {
	if isAdmin {
		return true, nil
	}

	accessibleIDs, hasAccessAll, err := groupRepo.GetAccessibleTemplateIDs(ctx, userID, workspaceID)
	if err != nil {
		return false, apperrors.ReturnInternalError("failed to check template access")
	}

	if hasAccessAll {
		return true, nil
	}

	for _, tid := range accessibleIDs {
		if tid == templateID {
			return true, nil
		}
	}

	return false, nil
}
