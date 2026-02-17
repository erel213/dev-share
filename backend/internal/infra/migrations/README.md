# Database Schema

```mermaid
erDiagram
    workspaces {
        UUID id PK
        VARCHAR name
        TEXT description
        UUID admin_id
        TIMESTAMPTZ created_at
        TIMESTAMPTZ updated_at
    }
    users {
        UUID id PK
        VARCHAR oauth_provider
        VARCHAR oauth_id
        VARCHAR password
        VARCHAR name
        VARCHAR email
        UUID workspace_id FK
        TIMESTAMPTZ created_at
        TIMESTAMPTZ updated_at
    }
    environments {
        UUID id PK
        VARCHAR name
        TEXT description
        UUID created_by FK
        UUID workspace_id FK
        UUID template_id FK
        TIMESTAMPTZ created_at
        TIMESTAMPTZ updated_at
    }
    templates {
        UUID id PK
        VARCHAR name
        UUID workspace_id FK
        TEXT path
        TIMESTAMPTZ created_at
        TIMESTAMPTZ updated_at
    }

    workspaces ||--o{ users : "belongs to"
    workspaces ||--o{ environments : "contains"
    workspaces ||--o{ templates : "owns"
    users ||--o{ environments : "creates"
    templates ||--o{ environments : "applied to"
```
