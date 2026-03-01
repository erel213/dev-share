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

### Stopping

Press `Ctrl+C` in the terminal running `setup.sh` or `restart.sh`, or:

```sh
kill $(lsof -t -i:8080)
```
