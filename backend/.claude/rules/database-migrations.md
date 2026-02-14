---
paths:
  - "internal/domain/**/*.go"
  - "internal/infra/migrations/**/*.sql"
---

# Database Migration Rules

## When to Create Migrations

**CRITICAL**: Whenever you add, modify, or remove domain entities in `internal/domain/`, you MUST create corresponding database migration files.

### Triggers for New Migrations

- Adding a new domain entity (struct)
- Adding or removing fields from existing domain entities
- Changing field types in domain entities
- Adding or removing relationships between entities

## Migration File Requirements

### Location and Naming

- **Location**: All migrations MUST be in `internal/infra/migrations/`
- **Naming Convention**: `{version}_{description}.{up|down}.sql`
  - Version: 6-digit sequential number (e.g., `000001`, `000002`)
  - Description: Snake_case description (e.g., `create_users_table`)
  - Always create BOTH `.up.sql` and `.down.sql` files

### Migration Content Standards

#### Required Elements

Every migration MUST include:

1. **Idempotency Clauses**
   - Use `IF NOT EXISTS` for CREATE statements
   - Use `IF EXISTS` for DROP statements

2. **Standard Columns**
   - `id UUID PRIMARY KEY DEFAULT gen_random_uuid()`
   - `created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP`
   - `updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP`

3. **Proper Data Types**
   - Use `UUID` for all ID fields (primary and foreign keys)
   - Use `VARCHAR(255)` for standard string fields
   - Use `TEXT` for longer content (descriptions, comments)
   - Use `TIMESTAMP WITH TIME ZONE` for all datetime fields

#### Foreign Keys

- Define foreign key constraints with explicit names: `CONSTRAINT fk_{table}_{column}`
- Always specify `ON DELETE` behavior:
  - `CASCADE` for dependent entities (e.g., delete user's environments when user is deleted)
  - `RESTRICT` or `NO ACTION` for independent entities
  - `SET NULL` for optional relationships

Example:
```sql
CONSTRAINT fk_workspace FOREIGN KEY (workspace_id)
  REFERENCES workspaces(id) ON DELETE CASCADE
```

#### Indexes

Create indexes for:
- All foreign key columns: `CREATE INDEX idx_{table}_{column} ON {table}({column});`
- Frequently queried columns (email, username, etc.)
- Composite unique constraints

**DO NOT** create redundant indexes:
- Primary keys are automatically indexed
- Unique constraints automatically create indexes

#### Unique Constraints

Add unique constraints where business logic requires uniqueness:
- Use descriptive constraint names: `CONSTRAINT unique_{description}`
- For composite uniqueness, specify all columns in the constraint

Example:
```sql
CONSTRAINT unique_oauth_user UNIQUE (oauth_provider, oauth_id)
```

### Down Migrations

- Must cleanly reverse the up migration
- Use `IF EXISTS` clauses for safety
- Drop dependent objects before parent objects (reverse order of creation)
- Consider data loss implications (document if destructive)

Example:
```sql
DROP TABLE IF EXISTS environments;
```

