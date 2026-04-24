---
title: "Manage application secrets"
description: "Configure Dev-Share to fetch application secrets from AWS Secrets Manager, Azure Key Vault, or GCP Secret Manager on the host and deliver them to the container via tmpfs-backed files."
weight: 10
lastmod: 2026-04-23
draft: false
---

Dev-Share requires two application secrets to operate: `JWT_SECRET` (signs authentication tokens) and `ENCRYPTION_KEY` (encrypts sensitive data at rest). You can provide these through a local `.env` file or fetch them from a cloud secret manager on the host and deliver them to the backend container as tmpfs-backed files.

## How secret resolution works

For each secret, the Go config loader checks two sources, in order:

1. **`${NAME}_FILE` env var** — if set, reads the file at that path (whitespace-trimmed). This is how the cloud-secret flow delivers values.
2. **`${NAME}` env var** — falls back to the plain env var, typically populated from `.env`.

If neither produces a usable value, the application exits with a configuration error (minimum length / hex validation is enforced).

## Why files and not env vars?

Secrets passed as container environment variables are visible to anyone who can run `docker inspect <container>` or read `/proc/<pid>/environ` on the host, and they commonly leak into logs and crash reports. The file-based flow keeps secrets in memory (tmpfs) on the host and mounts them read-only into the container at `/run/secrets/*` — they never appear in the container's env list.

## Option A: Local .env file

Generate a `.env` file with random secrets using the setup script:

```bash
./setup.sh --generate-env
```

This creates `.env` from `.env.example` and populates `JWT_SECRET` and `ENCRYPTION_KEY` with cryptographically random values. You can also create the file manually:

```bash
cp .env.example .env

# Generate and paste into .env
openssl rand -base64 32    # JWT_SECRET
openssl rand -hex 32       # ENCRYPTION_KEY
```

{{< callout type="info" >}}
On macOS and other non-Linux hosts, use Option A. The tmpfs-backed cloud flow is Linux-only because it relies on `/dev/shm`.
{{< /callout >}}

## How the cloud-secret flow works

Before the first cloud-secret deploy, create `docker-compose.override.yml` from the shipped example and uncomment **OPTION A** (plus the top-level `secrets:` block at the bottom of the file):

```bash
cp docker-compose.override.example.yml docker-compose.override.yml
# edit docker-compose.override.yml — uncomment OPTION A and the top-level secrets: block
```

`docker-compose.override.yml` is gitignored, so the uncomment state is per-deployment.

After that, export one of `AWS_SECRET_ID` / `AZURE_KEYVAULT_NAME` / `GCP_SECRET_NAME` and run `./setup.sh`:

1. `setup.sh` verifies `docker-compose.override.yml` has `JWT_SECRET_FILE` uncommented (fails with guidance if not).
2. `setup.sh` invokes `scripts/fetch-secrets.sh` on the host.
3. The script uses the host's cloud CLI (`aws` / `az` / `gcloud`) to fetch the secrets.
4. Values are written to `/dev/shm/devshare-secrets/` (tmpfs, mode `0700` directory / `0600` files).
5. Compose auto-merges `docker-compose.override.yml`, bind-mounting the tmpfs files at `/run/secrets/*`.
6. The backend reads them via `JWT_SECRET_FILE`, `ENCRYPTION_KEY_FILE`, `ADMIN_INIT_TOKEN_FILE`.
7. On `./setup.sh` exit, the tmpfs files are `shred`-ed and the directory removed.

## Option B: AWS Secrets Manager

Store your secrets in AWS Secrets Manager as a JSON object and let the host script fetch them.

### Prerequisites

- An AWS account with Secrets Manager access
- **The AWS CLI installed on the host** (`aws --version`)
- **`jq` installed on the host** (used by `fetch-secrets.sh` to parse the JSON payload)
- Host credentials that allow `secretsmanager:GetSecretValue` on the target secret (instance role preferred over static keys)

### Create the secret

```bash
aws secretsmanager create-secret \
  --name devshare/app-secrets \
  --secret-string '{
    "JWT_SECRET": "YOUR_JWT_SECRET_VALUE",
    "ENCRYPTION_KEY": "YOUR_64_HEX_CHAR_ENCRYPTION_KEY",
    "ADMIN_INIT_TOKEN": "optional-setup-token"
  }'
```

### Configure Dev-Share

Export before running `./setup.sh` (do **not** put these in `.env` — they are host-side):

```bash
export AWS_SECRET_ID=devshare/app-secrets
export AWS_DEFAULT_REGION=us-east-1
./setup.sh
```

### IAM permissions

The host needs an IAM policy with at minimum:

```json
{
  "Effect": "Allow",
  "Action": ["secretsmanager:GetSecretValue"],
  "Resource": "arn:aws:secretsmanager:<region>:<account>:secret:devshare/*"
}
```

On EC2, attach an IAM Instance Role so no static credentials are needed on the host.

## Option C: Azure Key Vault

Store each secret as a separate Key Vault secret and let the host script fetch them.

### Prerequisites

- An Azure Key Vault instance
- **The Azure CLI installed on the host** (`az --version`)
- A host principal with the `Key Vault Secrets User` role on the vault

### Create the secrets

```bash
az keyvault secret set --vault-name my-devshare-vault --name jwt-secret       --value "YOUR_JWT_SECRET_VALUE"
az keyvault secret set --vault-name my-devshare-vault --name encryption-key   --value "YOUR_64_HEX_CHAR_ENCRYPTION_KEY"
az keyvault secret set --vault-name my-devshare-vault --name admin-init-token --value "optional-setup-token"
```

### Configure Dev-Share

```bash
export AZURE_KEYVAULT_NAME=my-devshare-vault
./setup.sh
```

### Authentication

On Azure VMs, enable System-Assigned Managed Identity and assign the `Key Vault Secrets User` role:

```bash
az vm identity assign --name <vm-name> --resource-group <rg>
az role assignment create \
  --role "Key Vault Secrets User" \
  --assignee <managed-identity-principal-id> \
  --scope /subscriptions/<sub>/resourceGroups/<rg>/providers/Microsoft.KeyVault/vaults/<vault-name>
```

## Option D: GCP Secret Manager

Store your secrets in GCP Secret Manager as a JSON payload and let the host script fetch them.

### Prerequisites

- A GCP project with the Secret Manager API enabled
- **The `gcloud` CLI installed on the host** (`gcloud --version`)
- **`jq` installed on the host**
- A host service account with `roles/secretmanager.secretAccessor`

### Create the secret

```bash
echo '{
  "JWT_SECRET": "YOUR_JWT_SECRET_VALUE",
  "ENCRYPTION_KEY": "YOUR_64_HEX_CHAR_ENCRYPTION_KEY",
  "ADMIN_INIT_TOKEN": "optional-setup-token"
}' | gcloud secrets create devshare-secrets --data-file=-
```

### Configure Dev-Share

```bash
export GCP_SECRET_NAME=devshare-secrets
./setup.sh
```

### Authentication

On GCE, attach a service account with the `Secret Accessor` role:

```bash
gcloud secrets add-iam-policy-binding devshare-secrets \
  --member="serviceAccount:<sa-email>" \
  --role="roles/secretmanager.secretAccessor"
```

## Secret format reference

AWS and GCP store secrets as a single JSON object. Azure stores each value as a separate Key Vault secret.

### AWS and GCP JSON format

```json
{
  "JWT_SECRET": "<base64-encoded, minimum 32 characters>",
  "ENCRYPTION_KEY": "<64 hex characters (32 bytes)>",
  "ADMIN_INIT_TOKEN": "<optional>"
}
```

### Azure Key Vault secret names

| Key Vault secret name | Maps to |
|----------------------|---------|
| `jwt-secret` | `JWT_SECRET` |
| `encryption-key` | `ENCRYPTION_KEY` |
| `admin-init-token` | `ADMIN_INIT_TOKEN` |

## Running Compose manually

If you invoke Compose directly (without `setup.sh`), the flow is:

```bash
# One-time: copy + uncomment OPTION A in the override
cp docker-compose.override.example.yml docker-compose.override.yml
$EDITOR docker-compose.override.yml   # uncomment OPTION A + the top-level secrets: block

# Each deploy:
eval "$(./scripts/fetch-secrets.sh)"   # prints SECRETS_DIR=...
export SECRETS_DIR
docker compose up -d                   # override.yml is auto-merged
```

## Security best practices

- **Prefer instance identity over static credentials.** Use IAM Instance Roles (AWS), Managed Identity (Azure), or attached service accounts (GCP) so no long-lived credentials exist on the host.
- **Scope access narrowly.** Grant only `GetSecretValue` / `Secret Accessor` on the specific secret resource, not broad secret manager access.
- **Enable audit logging.** AWS CloudTrail, Azure Monitor, and GCP Cloud Audit Logs all record secret access events.
- **Rotate secrets periodically.** AWS Secrets Manager supports automatic rotation via Lambda. For Azure and GCP, implement rotation via Event Grid or Pub/Sub notifications.

## Next steps

- [Environment variables reference]({{< ref "docs/reference/environment-variables" >}}) for the full list of configuration options
- [Configure cloud credentials]({{< ref "docs/getting-started/cloud-credentials" >}}) to set up Terraform provider authentication
