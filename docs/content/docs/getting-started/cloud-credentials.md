---
title: "Configure cloud credentials"
description: "Provide AWS, GCP, or Azure credentials so Dev-Share can provision cloud infrastructure with Terraform inside Docker."
weight: 30
draft: false
---

Dev-Share uses Terraform to provision infrastructure. It does not manage cloud credentials directly — it delegates authentication to Terraform, which uses each cloud provider's standard credential chain.

When running in Docker, cloud credentials must be explicitly passed into the container. The recommended approach is to create a `docker-compose.override.yml` file:

```bash
cp docker-compose.override.example.yml docker-compose.override.yml
```

This file is gitignored to prevent accidental credential commits. Uncomment the sections relevant to your cloud provider and preferred method below.

## Option A: Environment variables

Set credentials as environment variables on your host (or in `.env`), then list them in the override file so Docker Compose passes the host values through.

### AWS

```yaml
# docker-compose.override.yml
services:
  backend:
    environment:
      - AWS_ACCESS_KEY_ID
      - AWS_SECRET_ACCESS_KEY
      - AWS_SESSION_TOKEN          # For temporary credentials (STS, SSO)
      - AWS_DEFAULT_REGION
```

### GCP

```yaml
# docker-compose.override.yml
services:
  backend:
    environment:
      - GOOGLE_APPLICATION_CREDENTIALS
      - GOOGLE_PROJECT
```

### Azure

```yaml
# docker-compose.override.yml
services:
  backend:
    environment:
      - ARM_CLIENT_ID
      - ARM_CLIENT_SECRET
      - ARM_TENANT_ID
      - ARM_SUBSCRIPTION_ID
```

## Option B: Credential file mounts

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

All mounts use `:ro` (read-only) — the container cannot modify your host credentials.

For a GCP service account key file, mount the individual file:

```yaml
volumes:
  - /path/to/service-account.json:/root/.config/gcloud/sa-key.json:ro
```

## Option C: Instance identity

On EC2 (IAM role), GCE (attached service account), or Azure VM (managed identity), Terraform obtains credentials automatically from the cloud metadata endpoint. Enable host networking so the container can reach it:

```yaml
# docker-compose.override.yml
services:
  backend:
    network_mode: host
```

{{< callout type="warning" >}}
`network_mode: host` removes container network isolation. Only use this in trusted environments such as private cloud VMs.
{{< /callout >}}

## Application secrets vs. cloud credentials

This page covers **cloud provider credentials** used by Terraform to provision infrastructure. If you want to manage Dev-Share's own application secrets (`JWT_SECRET`, `ENCRYPTION_KEY`) through a cloud secret manager instead of a local `.env` file, see [Manage application secrets]({{< ref "docs/guides/secrets-management" >}}).

## Security best practices

- **Never bake credentials into the Docker image.** Use mounts or environment variables at runtime.
- **`docker-compose.override.yml` is gitignored** to prevent accidental credential commits.
- **Prefer short-lived credentials** — AWS STS sessions, GCP OIDC workload identity, Azure federated credentials — over long-lived access keys.
- **Principle of least privilege** — grant only the IAM permissions your Terraform templates require, not full admin access.
