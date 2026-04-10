# Dev-Share

## Quick Start

### Prerequisites
- Go 1.24+
- Node.js 22+ and pnpm

### Installation

```sh
git clone <repo-url>
cd dev-share
./setup.sh
```

The script will:
1. Check prerequisites (Go, pnpm, curl)
2. Install backend and frontend dependencies
3. Build the frontend
4. Create `.env` with a generated JWT secret (if not present)
5. Run database migrations
6. Start backend and frontend servers
7. Open the setup wizard in your browser

Once running, complete the onboarding flow to create your admin account and first workspace.

### Scripts

| Script | Description |
|---|---|
| `./setup.sh` | First-time setup — installs deps, builds, migrates, and starts everything |
| `./restart.sh [target]` | Restart servers (`backend`, `frontend`, or `all` — default: `all`) |
| `./clean_db.sh` | Delete all SQLite database files (`.db`, `.db-wal`, `.db-shm`) |
| `./fast-deploy.sh` | Provision an EC2 instance and deploy the app (requires Terraform + AWS credentials) |

### Configuration (.env)

| Variable               | Default                | Description                                |
|------------------------|------------------------|--------------------------------------------|
| JWT_SECRET             | (auto-generated)       | JWT signing secret                         |
| ENCRYPTION_KEY         | (auto-generated)       | AES encryption key for sensitive data      |
| PORT                   | 8080                   | Backend API port                           |
| DB_FILE_PATH           | ./backend/devshare.db  | SQLite database file location              |
| ADMIN_INIT_TOKEN       | (empty)                | Optional token to protect /admin/init      |
| TEMPLATE_STORAGE_PATH  | ./template_storage     | Directory for uploaded template files      |
| ENV_EXECUTION_PATH     | ./env_executions       | Terraform working directory for executions |
| TF_PLUGIN_CACHE_DIR    | (empty)                | Terraform plugin cache directory           |
| VITE_API_BASE_URL      | http://localhost:8080   | Frontend API base URL (local dev only)    |
| APP_PORT               | 3000                   | Frontend port in Docker                    |

### Development

```sh
# Backend only
./restart.sh backend

# Frontend only
./restart.sh frontend

# Both
./restart.sh
```

Or run manually:

```sh
# Backend
cd backend && make run

# Frontend (separate terminal)
cd frontend && pnpm dev
```

### Reset Database

To wipe the database and start fresh:

```sh
./clean_db.sh
./setup.sh
```

### Docker Deployment

To run the full stack in Docker:

```sh
docker compose up --build
```

This starts the backend (Go + SQLite) and frontend (Nginx) containers. The frontend is available at `http://localhost:3000` (or `APP_PORT`). See `docker-compose.override.example.yml` for cloud credential configuration.

### Cloud Provider Authentication

dev-share does **not** manage cloud credentials directly. It delegates authentication entirely to the underlying IaC platform (e.g., Terraform), which uses the cloud SDK's built-in credential chain.

Since the backend runs inside a Docker container, cloud credentials must be explicitly passed in. The recommended approach is to create a `docker-compose.override.yml` (auto-merged by Docker Compose, gitignored).

```sh
cp docker-compose.override.example.yml docker-compose.override.yml
# Edit the file — uncomment the sections for your provider and method
```

#### Quick Reference

| Scenario | Method | What to configure |
|---|---|---|
| Local dev with `aws configure` / `gcloud auth` / `az login` | Volume mount | Mount `~/.aws`, `~/.config/gcloud`, or `~/.azure` read-only |
| Explicit access keys or service principal | Env vars | Set in `.env`, pass through in override |
| CI/CD pipeline | Env vars | Pipeline injects vars, listed in override |
| EC2 / GCE / Azure VM with instance role | Metadata | `network_mode: host` in override |
| GCP service account JSON | Volume mount + env var | Mount JSON file, set `GOOGLE_APPLICATION_CREDENTIALS` |

#### Option A: Credential File Mounts

Mount host credential directories read-only into the container. This works if you authenticate via CLI tools (`aws configure`, `gcloud auth application-default login`, `az login`).

```yaml
# docker-compose.override.yml
services:
  backend:
    volumes:
      # AWS
      - ${HOME}/.aws:/root/.aws:ro
      # GCP
      - ${HOME}/.config/gcloud:/root/.config/gcloud:ro
      # Azure
      - ${HOME}/.azure:/root/.azure:ro
```

> All mounts use `:ro` (read-only) — the container cannot modify host credentials.

#### Option B: Environment Variables

Pass cloud credentials as environment variables. Set them on the host or in `.env`, then list them (without `=value`) in the override so Docker Compose passes the host values through.

```yaml
# docker-compose.override.yml
services:
  backend:
    environment:
      # AWS
      - AWS_ACCESS_KEY_ID
      - AWS_SECRET_ACCESS_KEY
      - AWS_SESSION_TOKEN
      - AWS_DEFAULT_REGION
      # GCP
      - GOOGLE_APPLICATION_CREDENTIALS
      - GOOGLE_PROJECT
      # Azure (Terraform AzureRM provider uses ARM_ prefix)
      - ARM_CLIENT_ID
      - ARM_CLIENT_SECRET
      - ARM_TENANT_ID
      - ARM_SUBSCRIPTION_ID
```

#### Option C: Instance Identity (Cloud-Hosted)

On EC2 (IAM role), GCE (attached service account), or Azure VM (managed identity), Terraform gets credentials automatically from the metadata endpoint. Enable host networking so the container can reach it:

```yaml
# docker-compose.override.yml
services:
  backend:
    network_mode: host
```

> **Warning:** `network_mode: host` removes container network isolation. Only use this in trusted environments.

#### Security Notes

- **Never bake credentials into the Docker image.** Use mounts or env vars at runtime.
- **`docker-compose.override.yml` is gitignored** to prevent accidental credential commits.
- **Prefer short-lived credentials** — AWS STS / SSO sessions, GCP OIDC workload identity, Azure federated credentials — over long-lived access keys.
- **Principle of least privilege** — grant only the IAM permissions Terraform needs, not admin access.

### Stopping

Press `Ctrl+C` in the terminal running `setup.sh` or `restart.sh`, or:

```sh
kill $(lsof -t -i:8080)
```
