package sqlite

import (
	"context"
	"database/sql"

	"backend/internal/domain"
	domainerrors "backend/internal/domain/errors"
	"backend/internal/domain/repository"
	infraerrors "backend/internal/infra/errors"
	pkgerrors "backend/pkg/errors"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
)

type groupRepository struct {
	uow *UnitOfWork
}

func newGroupRepository(uow *UnitOfWork) repository.GroupRepository {
	return &groupRepository{uow: uow}
}

// --- Group CRUD ---

func (r *groupRepository) Create(ctx context.Context, group *domain.Group) *pkgerrors.Error {
	query, args, err := builder.
		Insert("groups").
		Columns("id", "name", "description", "workspace_id", "access_all_templates").
		Values(group.ID, group.Name, group.Description, group.WorkspaceID, boolToInt(group.AccessAllTemplates)).
		Suffix("RETURNING created_at, updated_at").
		ToSql()
	if err != nil {
		return infraerrors.WrapSQLiteError(err, "create_group")
	}

	var cat, uat TimestampDest
	err = r.uow.Querier().QueryRowContext(ctx, query, args...).Scan(&cat, &uat)
	if err != nil {
		return infraerrors.WrapSQLiteError(err, "create_group")
	}

	group.CreatedAt = cat.Time()
	group.UpdatedAt = uat.Time()

	return nil
}

func (r *groupRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Group, *pkgerrors.Error) {
	query, args, err := builder.
		Select("id", "name", "description", "workspace_id", "access_all_templates", "created_at", "updated_at").
		From("groups").
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return nil, infraerrors.WrapSQLiteError(err, "get_group")
	}

	group, scanErr := r.scanGroup(r.uow.Querier().QueryRowContext(ctx, query, args...))
	if scanErr != nil {
		if scanErr == sql.ErrNoRows {
			return nil, domainerrors.NotFound("Group", id.String())
		}
		return nil, infraerrors.WrapSQLiteError(scanErr, "get_group")
	}

	return group, nil
}

func (r *groupRepository) GetByWorkspaceID(ctx context.Context, workspaceID uuid.UUID) ([]*domain.Group, *pkgerrors.Error) {
	query, args, err := builder.
		Select("id", "name", "description", "workspace_id", "access_all_templates", "created_at", "updated_at").
		From("groups").
		Where(sq.Eq{"workspace_id": workspaceID}).
		OrderBy("created_at DESC").
		ToSql()
	if err != nil {
		return nil, infraerrors.WrapSQLiteError(err, "get_groups_by_workspace")
	}

	rows, err := r.uow.Querier().QueryContext(ctx, query, args...)
	if err != nil {
		return nil, infraerrors.WrapSQLiteError(err, "get_groups_by_workspace")
	}
	defer rows.Close()

	var groups []*domain.Group
	for rows.Next() {
		var g domain.Group
		var accessAll int
		var cat, uat TimestampDest
		err := rows.Scan(&g.ID, &g.Name, &g.Description, &g.WorkspaceID, &accessAll, &cat, &uat)
		if err != nil {
			return nil, infraerrors.WrapSQLiteError(err, "scan_group")
		}
		g.AccessAllTemplates = accessAll == 1
		g.CreatedAt = cat.Time()
		g.UpdatedAt = uat.Time()
		groups = append(groups, &g)
	}

	if err := rows.Err(); err != nil {
		return nil, infraerrors.WrapSQLiteError(err, "iterate_groups")
	}

	return groups, nil
}

func (r *groupRepository) Update(ctx context.Context, group *domain.Group) *pkgerrors.Error {
	query, args, err := builder.
		Update("groups").
		Set("name", group.Name).
		Set("description", group.Description).
		Set("access_all_templates", boolToInt(group.AccessAllTemplates)).
		Set("updated_at", sq.Expr("CURRENT_TIMESTAMP")).
		Where(sq.Eq{"id": group.ID}).
		Suffix("RETURNING updated_at").
		ToSql()
	if err != nil {
		return infraerrors.WrapSQLiteError(err, "update_group")
	}

	var uat TimestampDest
	err = r.uow.Querier().QueryRowContext(ctx, query, args...).Scan(&uat)
	if err != nil {
		if err == sql.ErrNoRows {
			return domainerrors.NotFound("Group", group.ID.String())
		}
		return infraerrors.WrapSQLiteError(err, "update_group")
	}

	group.UpdatedAt = uat.Time()

	return nil
}

func (r *groupRepository) Delete(ctx context.Context, id uuid.UUID) *pkgerrors.Error {
	query, args, err := builder.
		Delete("groups").
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return infraerrors.WrapSQLiteError(err, "delete_group")
	}

	result, err := r.uow.Querier().ExecContext(ctx, query, args...)
	if err != nil {
		return infraerrors.WrapSQLiteError(err, "delete_group")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return infraerrors.WrapSQLiteError(err, "get_rows_affected")
	}

	if rowsAffected == 0 {
		return domainerrors.NotFound("Group", id.String())
	}

	return nil
}

// --- Membership ---

func (r *groupRepository) AddMembers(ctx context.Context, groupID uuid.UUID, userIDs []uuid.UUID) *pkgerrors.Error {
	if len(userIDs) == 0 {
		return nil
	}

	ins := builder.Insert("group_memberships").Columns("group_id", "user_id")
	for _, uid := range userIDs {
		ins = ins.Values(groupID, uid)
	}
	// Ignore duplicates — already a member
	query, args, err := ins.Suffix("ON CONFLICT (group_id, user_id) DO NOTHING").ToSql()
	if err != nil {
		return infraerrors.WrapSQLiteError(err, "add_group_members")
	}

	_, err = r.uow.Querier().ExecContext(ctx, query, args...)
	if err != nil {
		return infraerrors.WrapSQLiteError(err, "add_group_members")
	}

	return nil
}

func (r *groupRepository) RemoveMember(ctx context.Context, groupID uuid.UUID, userID uuid.UUID) *pkgerrors.Error {
	query, args, err := builder.
		Delete("group_memberships").
		Where(sq.Eq{"group_id": groupID, "user_id": userID}).
		ToSql()
	if err != nil {
		return infraerrors.WrapSQLiteError(err, "remove_group_member")
	}

	result, err := r.uow.Querier().ExecContext(ctx, query, args...)
	if err != nil {
		return infraerrors.WrapSQLiteError(err, "remove_group_member")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return infraerrors.WrapSQLiteError(err, "get_rows_affected")
	}

	if rowsAffected == 0 {
		return domainerrors.NotFound("GroupMembership", groupID.String()+"/"+userID.String())
	}

	return nil
}

func (r *groupRepository) GetMembers(ctx context.Context, groupID uuid.UUID) ([]uuid.UUID, *pkgerrors.Error) {
	query, args, err := builder.
		Select("user_id").
		From("group_memberships").
		Where(sq.Eq{"group_id": groupID}).
		OrderBy("created_at ASC").
		ToSql()
	if err != nil {
		return nil, infraerrors.WrapSQLiteError(err, "get_group_members")
	}

	return r.queryUUIDs(ctx, query, args, "get_group_members")
}

func (r *groupRepository) GetGroupIDsForUser(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, *pkgerrors.Error) {
	query, args, err := builder.
		Select("group_id").
		From("group_memberships").
		Where(sq.Eq{"user_id": userID}).
		ToSql()
	if err != nil {
		return nil, infraerrors.WrapSQLiteError(err, "get_groups_for_user")
	}

	return r.queryUUIDs(ctx, query, args, "get_groups_for_user")
}

// --- Template access ---

func (r *groupRepository) AddTemplateAccess(ctx context.Context, groupID uuid.UUID, templateIDs []uuid.UUID) *pkgerrors.Error {
	if len(templateIDs) == 0 {
		return nil
	}

	ins := builder.Insert("group_template_access").Columns("group_id", "template_id")
	for _, tid := range templateIDs {
		ins = ins.Values(groupID, tid)
	}
	query, args, err := ins.Suffix("ON CONFLICT (group_id, template_id) DO NOTHING").ToSql()
	if err != nil {
		return infraerrors.WrapSQLiteError(err, "add_group_template_access")
	}

	_, err = r.uow.Querier().ExecContext(ctx, query, args...)
	if err != nil {
		return infraerrors.WrapSQLiteError(err, "add_group_template_access")
	}

	return nil
}

func (r *groupRepository) RemoveTemplateAccess(ctx context.Context, groupID uuid.UUID, templateID uuid.UUID) *pkgerrors.Error {
	query, args, err := builder.
		Delete("group_template_access").
		Where(sq.Eq{"group_id": groupID, "template_id": templateID}).
		ToSql()
	if err != nil {
		return infraerrors.WrapSQLiteError(err, "remove_group_template_access")
	}

	result, err := r.uow.Querier().ExecContext(ctx, query, args...)
	if err != nil {
		return infraerrors.WrapSQLiteError(err, "remove_group_template_access")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return infraerrors.WrapSQLiteError(err, "get_rows_affected")
	}

	if rowsAffected == 0 {
		return domainerrors.NotFound("GroupTemplateAccess", groupID.String()+"/"+templateID.String())
	}

	return nil
}

func (r *groupRepository) GetTemplateAccess(ctx context.Context, groupID uuid.UUID) ([]uuid.UUID, *pkgerrors.Error) {
	query, args, err := builder.
		Select("template_id").
		From("group_template_access").
		Where(sq.Eq{"group_id": groupID}).
		OrderBy("created_at ASC").
		ToSql()
	if err != nil {
		return nil, infraerrors.WrapSQLiteError(err, "get_group_template_access")
	}

	return r.queryUUIDs(ctx, query, args, "get_group_template_access")
}

// --- Access query ---

func (r *groupRepository) GetAccessibleTemplateIDs(ctx context.Context, userID uuid.UUID, workspaceID uuid.UUID) ([]uuid.UUID, bool, *pkgerrors.Error) {
	// Check if user belongs to any group with access_all_templates=1
	hasAllQuery := `
		SELECT EXISTS(
			SELECT 1 FROM groups g
			JOIN group_memberships gm ON g.id = gm.group_id
			WHERE gm.user_id = ? AND g.workspace_id = ? AND g.access_all_templates = 1
		)`

	var hasAll int
	err := r.uow.Querier().QueryRowContext(ctx, hasAllQuery, userID, workspaceID).Scan(&hasAll)
	if err != nil {
		return nil, false, infraerrors.WrapSQLiteError(err, "check_access_all")
	}

	if hasAll == 1 {
		return nil, true, nil
	}

	// Get union of template IDs from all user's groups in this workspace
	templateQuery := `
		SELECT DISTINCT gta.template_id
		FROM group_template_access gta
		JOIN groups g ON gta.group_id = g.id
		JOIN group_memberships gm ON g.id = gm.group_id
		WHERE gm.user_id = ? AND g.workspace_id = ?`

	rows, err := r.uow.Querier().QueryContext(ctx, templateQuery, userID, workspaceID)
	if err != nil {
		return nil, false, infraerrors.WrapSQLiteError(err, "get_accessible_templates")
	}
	defer rows.Close()

	var ids []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, false, infraerrors.WrapSQLiteError(err, "scan_accessible_template")
		}
		ids = append(ids, id)
	}

	if err := rows.Err(); err != nil {
		return nil, false, infraerrors.WrapSQLiteError(err, "iterate_accessible_templates")
	}

	return ids, false, nil
}

// --- Co-member query ---

func (r *groupRepository) GetCoMemberUserIDs(ctx context.Context, userID uuid.UUID, workspaceID uuid.UUID) ([]uuid.UUID, *pkgerrors.Error) {
	query := `
		SELECT DISTINCT gm2.user_id
		FROM group_memberships gm1
		JOIN groups g ON gm1.group_id = g.id
		JOIN group_memberships gm2 ON gm1.group_id = gm2.group_id
		WHERE gm1.user_id = ? AND g.workspace_id = ? AND gm2.user_id != ?`

	return r.queryUUIDs(ctx, query, []interface{}{userID, workspaceID, userID}, "get_co_member_user_ids")
}

// --- Helpers ---

func (r *groupRepository) scanGroup(row *sql.Row) (*domain.Group, error) {
	var g domain.Group
	var accessAll int
	var cat, uat TimestampDest
	err := row.Scan(&g.ID, &g.Name, &g.Description, &g.WorkspaceID, &accessAll, &cat, &uat)
	if err != nil {
		return nil, err
	}
	g.AccessAllTemplates = accessAll == 1
	g.CreatedAt = cat.Time()
	g.UpdatedAt = uat.Time()
	return &g, nil
}

func (r *groupRepository) queryUUIDs(ctx context.Context, query string, args []interface{}, op string) ([]uuid.UUID, *pkgerrors.Error) {
	rows, err := r.uow.Querier().QueryContext(ctx, query, args...)
	if err != nil {
		return nil, infraerrors.WrapSQLiteError(err, op)
	}
	defer rows.Close()

	var ids []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, infraerrors.WrapSQLiteError(err, op)
		}
		ids = append(ids, id)
	}

	if err := rows.Err(); err != nil {
		return nil, infraerrors.WrapSQLiteError(err, op)
	}

	return ids, nil
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
