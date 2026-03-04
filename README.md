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

### Configuration (.env)

| Variable           | Default                | Description                           |
|--------------------|------------------------|---------------------------------------|
| JWT_SECRET         | (auto-generated)       | JWT signing secret                    |
| PORT               | 8080                   | Backend API port                      |
| DB_FILE_PATH       | ./backend/devshare.db  | SQLite database file location         |
| ADMIN_INIT_TOKEN   | (empty)                | Optional token to protect /admin/init |
| VITE_API_BASE_URL  | http://localhost:8080   | Frontend API base URL                 |

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

### Cloud Provider Authentication

dev-share does **not** manage cloud credentials directly. It delegates authentication entirely to the underlying IaC platform (e.g., Terraform), which uses the cloud SDK's built-in credential chain.

**Your responsibility** is to ensure the host running dev-share has valid credentials configured for the target cloud provider.

#### AWS

Any method supported by the [AWS credential chain](https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-files.html):

```sh
# Option 1: Environment variables
export AWS_ACCESS_KEY_ID=...
export AWS_SECRET_ACCESS_KEY=...

# Option 2: Shared credentials file (~/.aws/credentials)
aws configure

# Option 3: IAM instance profile (on EC2/ECS — automatic, no setup needed)
```

#### GCP

Any method supported by [Application Default Credentials](https://cloud.google.com/docs/authentication/application-default-credentials):

```sh
# Option 1: User credentials (local dev)
gcloud auth application-default login

# Option 2: Service account key
export GOOGLE_APPLICATION_CREDENTIALS=/path/to/key.json

# Option 3: Attached service account (on GCE/GKE — automatic, no setup needed)
```

#### Azure

Any method supported by [DefaultAzureCredential](https://learn.microsoft.com/en-us/azure/developer/go/azure-sdk-authentication):

```sh
# Option 1: Azure CLI (local dev)
az login

# Option 2: Service principal
export AZURE_CLIENT_ID=...
export AZURE_TENANT_ID=...
export AZURE_CLIENT_SECRET=...

# Option 3: Managed identity (on Azure VMs/App Service — automatic, no setup needed)
```

> **Tip:** For production and CI/CD, prefer short-lived credentials via OIDC workload identity federation or instance-attached roles over long-lived secrets.

### Stopping

Press `Ctrl+C` in the terminal running `setup.sh` or `restart.sh`, or:

```sh
kill $(lsof -t -i:8080)
```
