---
title: "Architecture and deployment model"
description: "Dev-Share components, environment provisioning flow, lifecycle state machine, and cloud credential delegation to Terraform."
weight: 20
draft: false
lastmod: 2026-04-25
---

## System components

Dev-Share is a two-service application:

| Component | Technology | Role |
|---|---|---|
| **Backend** | Go + Fiber, SQLite | REST API, Terraform executor, background reaper |
| **Frontend** | React + Vite, Tailwind CSS | Web UI served via Nginx in Docker |

Both services run in the same Docker Compose project. The backend is the only service that touches Terraform or cloud APIs.

**Database**: SQLite (single file, no separate database server). The file path is configurable via `DB_FILE_PATH`.

**Terraform**: The backend spawns `terraform init`, `terraform plan`, and `terraform apply` as child processes in isolated working directories under `ENV_EXECUTION_PATH`. Terraform plugins are cached at `TF_PLUGIN_CACHE_DIR` to avoid repeated downloads.

## How environment provisioning works

When a user applies an environment, the backend:

1. Writes the environment's variable values to a `terraform.tfvars` file in the working directory.
2. Copies the template's Terraform files into the working directory.
3. Runs `terraform init` (uses plugin cache if configured).
4. Sets environment status to `applying` and runs `terraform apply -auto-approve`.
5. Streams logs to its internal log store.
6. On success: sets status to `ready`, reads and stores `terraform output -json`.
7. On failure: sets status to `error` and stores the error message.

Plan and apply are separate operations. A user can run **Plan** first to preview changes, then **Apply** to execute them.

Destroy follows the same pattern: sets status to `destroying`, runs `terraform destroy -auto-approve`, then sets status to `destroyed`.

Only one operation can run on an environment at a time. Attempting to start a second operation while one is in progress returns an error.

## Environment lifecycle

```
pending
  └─→ initialized
        └─→ planning ──→ initialized (plan complete)
              └─→ applying ──→ ready
                                └─→ destroying ──→ destroyed
```

Any state can transition to `error` if Terraform exits non-zero. From `error`, the environment can be re-applied or destroyed.

| Status | Meaning |
|---|---|
| `pending` | Created, no Terraform operations run yet |
| `initialized` | `terraform init` complete |
| `planning` | `terraform plan` in progress |
| `applying` | `terraform apply` in progress |
| `ready` | Apply succeeded; outputs are available |
| `destroying` | `terraform destroy` in progress |
| `destroyed` | Resources have been torn down |
| `error` | Last operation failed |

## TTL and auto-cleanup

Every environment has an optional TTL (time-to-live) in seconds. A background reaper runs continuously and destroys environments whose `created_at + ttl_seconds` has passed.

The UI exposes preset TTL values (1 hour, 4 hours, 24 hours). Admins can also set no TTL for persistent environments.

## Deployment modes

**Docker Compose (recommended for production and shared deployments):**

```
docker compose up --build
```

The `docker-compose.yml` defines `backend` and `frontend` services. A `docker-compose.override.yml` (gitignored) is auto-merged by Docker Compose and is the correct place for cloud credentials, secret mounts, and environment-specific overrides.

**Local development:**

Run the backend with `cd backend && make run` and the frontend with `cd frontend && pnpm dev`. The frontend proxies API calls to `localhost:8080` via `VITE_API_BASE_URL`. See [Local development setup]({{< ref "docs/getting-started/local-development" >}}).

## Cloud credential delegation

Dev-Share does not store or manage cloud credentials. Terraform reads credentials from its standard credential chain (environment variables, credential files, instance metadata).

Because the backend runs inside a Docker container, credentials must be explicitly provided at container start time. The supported methods are:

- **Volume mounts** — mount `~/.aws`, `~/.config/gcloud`, or `~/.azure` read-only into the container.
- **Environment variables** — pass `AWS_ACCESS_KEY_ID`, `ARM_CLIENT_ID`, etc. through the compose override.
- **Instance identity** — use `network_mode: host` to reach the cloud provider metadata endpoint (EC2, GCE, Azure VM).

See [Configure cloud credentials]({{< ref "docs/getting-started/cloud-credentials" >}}) for the full setup.
