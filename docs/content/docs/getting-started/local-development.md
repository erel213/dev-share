---
title: "Local development setup"
description: "Run the Dev-Share backend and frontend directly on your machine for development with hot reloading and fast iteration."
weight: 20
draft: false
---

Run Dev-Share outside of Docker for a faster development workflow with hot reloading.

## Prerequisites

- [Go](https://go.dev/dl/) 1.24 or later
- [Node.js](https://nodejs.org/) 22 or later
- [pnpm](https://pnpm.io/installation) package manager

## Clone and configure

```bash
git clone <repo-url>
cd dev-share
```

Create your environment file:

```bash
cp .env.example .env
```

Generate the required secrets and update `.env`:

```bash
# Generate JWT_SECRET
openssl rand -base64 32

# Generate ENCRYPTION_KEY
openssl rand -hex 32
```

Copy the output of each command into the corresponding field in `.env`.

## Start the backend

```bash
cd backend
make run
```

The backend starts on `http://localhost:8080`. Database migrations run automatically on startup.

Verify it is running:

```bash
curl http://localhost:8080/health
```

## Start the frontend

In a separate terminal:

```bash
cd frontend
pnpm install
pnpm dev
```

The frontend starts on `http://localhost:5173` with hot module replacement enabled. It connects to the backend using the `VITE_API_BASE_URL` value in `.env` (defaults to `http://localhost:8080`).

## Start both at once

The `restart.sh` script starts both services in the background:

```bash
./restart.sh            # Start both backend and frontend
./restart.sh backend    # Backend only
./restart.sh frontend   # Frontend only
```

## Useful scripts

| Script | Description |
|--------|-------------|
| `./restart.sh [target]` | Restart servers (`backend`, `frontend`, or `all` — default: `all`) |
| `./clean_db.sh` | Delete all SQLite database files and start fresh |

## Next steps

- [Configure cloud credentials]({{< ref "docs/getting-started/cloud-credentials" >}}) to provision infrastructure with Terraform
- [Environment variables reference]({{< ref "docs/reference/environment-variables" >}}) for all configuration options
