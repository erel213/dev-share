---
title: "Quick start"
description: "Install and run Dev-Share with Docker. Secrets are delivered to the container via tmpfs-backed file mounts from a cloud secret manager, with a local .env fallback for development."
weight: 10
lastmod: 2026-04-23
draft: false
---

Get Dev-Share running with Docker. The recommended flow fetches application secrets from a cloud secret manager on the host and mounts them into the container as tmpfs-backed files — they never appear in the container's environment. A local `.env` path is documented at the end for quick local development.

## Prerequisites

For either path:

- [Docker](https://docs.docker.com/get-docker/) (with the Docker daemon running)
- [Docker Compose](https://docs.docker.com/compose/install/) (included with Docker Desktop)
- [curl](https://curl.se/)

Additionally, for the recommended cloud-secret flow:

- A Linux host (the tmpfs path uses `/dev/shm`)
- One of the cloud CLIs installed and authenticated: `aws`, `az`, or `gcloud`
- `jq` (required by `scripts/fetch-secrets.sh` for AWS and GCP JSON payloads)
- Host credentials with `GetSecretValue` / `Key Vault Secrets User` / `Secret Accessor` on your target secret

## Clone the repository

```bash
git clone <repo-url>
cd dev-share
```

## Recommended: tmpfs file-mounted secrets

This is the default path for any non-laptop deploy (staging, prod, shared infra). Secrets stay out of container env vars and off the host disk.

### 1. Store your secrets in a cloud secret manager

Pick one provider. Example with AWS:

```bash
aws secretsmanager create-secret \
  --name devshare/app-secrets \
  --secret-string '{
    "JWT_SECRET": "'"$(openssl rand -base64 32)"'",
    "ENCRYPTION_KEY": "'"$(openssl rand -hex 32)"'"
  }'
```

For Azure Key Vault and GCP Secret Manager, see [Manage application secrets]({{< ref "docs/guides/secrets-management" >}}).

### 2. Enable the tmpfs mount in the compose override

`docker-compose.override.yml` is gitignored, so this step is per-deployment:

```bash
cp docker-compose.override.example.yml docker-compose.override.yml
```

Open `docker-compose.override.yml` in your editor and uncomment **OPTION A** (the `environment:` + `secrets:` block under `services.backend`) and the **top-level `secrets:` block** at the bottom of the file.

### 3. Export the cloud secret ID and run setup

```bash
export AWS_SECRET_ID=devshare/app-secrets
export AWS_DEFAULT_REGION=us-east-1
./setup.sh
```

The script performs the following steps:

1. Verifies Docker, Docker Compose, and curl
2. Checks that `docker-compose.override.yml` has OPTION A uncommented (fails with guidance otherwise)
3. Runs `scripts/fetch-secrets.sh` on the host — fetches secrets via the cloud CLI and writes them to `/dev/shm/devshare-secrets/` (mode `0600`, tmpfs)
4. Builds and starts the backend and frontend containers (Compose auto-merges `docker-compose.override.yml`, bind-mounting the tmpfs files at `/run/secrets/*`)
5. Waits for the backend health check
6. Opens the setup wizard in your browser
7. On exit, shreds the tmpfs files and removes the directory

{{< callout type="info" >}}
Want to confirm secrets aren't leaking into the container env? Run `docker inspect devshare-backend | jq '.[0].Config.Env'` — you should see only `*_FILE` entries, never the raw secret values.
{{< /callout >}}

## Alternative: local `.env` (quick dev fallback)

For local experimentation on a laptop — especially on macOS, where the tmpfs path isn't available — use the generated `.env`:

```bash
./setup.sh --generate-env
```

This creates `.env` from `.env.example` and populates `JWT_SECRET` and `ENCRYPTION_KEY` with random values. Secrets live in the container environment, so treat this as a dev-only convenience — not a production pattern.

## Complete the setup wizard

Your browser opens to `http://localhost:3000/setup`. Follow the on-screen instructions to create your admin account and first workspace.

If the browser does not open automatically, navigate to `http://localhost:3000/setup` manually.

## Verify the installation

Confirm that Dev-Share is running by checking the health endpoint:

```bash
curl http://localhost:3000/health
```

A successful response indicates both the backend and frontend are operational.

## Stop and restart

Stop all containers:

```bash
docker compose down
```

Start them again (without rebuilding):

```bash
docker compose up -d
```

Rebuild after code changes:

```bash
docker compose up --build -d
```

{{< callout type="warning" >}}
When using the cloud-secret flow, always restart via `./setup.sh` rather than raw `docker compose up`. The setup script is what repopulates the tmpfs secrets directory; a bare `docker compose up` will fail if the tmpfs files are missing (for example, after a host reboot wiped `/dev/shm`).
{{< /callout >}}

## Reset the database

To wipe the database and start fresh:

```bash
./clean_db.sh
./setup.sh                      # cloud-secret flow, or
./setup.sh --generate-env       # local .env fallback
```

{{< callout type="warning" >}}
This deletes all data including users, workspaces, templates, and environments.
{{< /callout >}}

## Next steps

- [Manage application secrets]({{< ref "docs/guides/secrets-management" >}}) for Azure Key Vault, GCP Secret Manager, and IAM best practices
- [Configure cloud credentials]({{< ref "docs/getting-started/cloud-credentials" >}}) to provision infrastructure with Terraform
- [Environment variables reference]({{< ref "docs/reference/environment-variables" >}}) for all configuration options
