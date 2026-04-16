---
title: "Quick start"
description: "Install and run Dev-Share with a single command using Docker. This guide covers prerequisites, setup, and verification."
weight: 10
draft: false
---

Get Dev-Share running in under five minutes using Docker.

## Prerequisites

You need the following installed on your machine:

- [Docker](https://docs.docker.com/get-docker/) (with the Docker daemon running)
- [Docker Compose](https://docs.docker.com/compose/install/) (included with Docker Desktop)
- [curl](https://curl.se/)

## Clone the repository

```bash
git clone <repo-url>
cd dev-share
```

## Run the setup script

```bash
./setup.sh --generate-env
```

The script performs the following steps automatically:

1. Verifies that Docker, Docker Compose, and curl are installed
2. Creates a `.env` file with auto-generated secrets (`JWT_SECRET` and `ENCRYPTION_KEY`)
3. Builds and starts the backend and frontend containers
4. Waits for the backend health check to pass
5. Opens the setup wizard in your browser

{{< callout type="info" >}}
The `--generate-env` flag tells the script to create a `.env` file with random secrets. Without it, the script assumes secrets are provided by a cloud secret manager. See [Manage application secrets]({{< ref "docs/guides/secrets-management" >}}) for cloud-based options.
{{< /callout >}}

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

## Reset the database

To wipe the database and start fresh:

```bash
./clean_db.sh
./setup.sh --generate-env
```

{{< callout type="warning" >}}
This deletes all data including users, workspaces, templates, and environments.
{{< /callout >}}

## Next steps

- [Manage application secrets]({{< ref "docs/guides/secrets-management" >}}) to use a cloud secret manager instead of local `.env`
- [Configure cloud credentials]({{< ref "docs/getting-started/cloud-credentials" >}}) to provision infrastructure with Terraform
- [Environment variables reference]({{< ref "docs/reference/environment-variables" >}}) for all configuration options
