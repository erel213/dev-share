---
title: "Manage application secrets"
description: "Configure Dev-Share to fetch application secrets from AWS Secrets Manager, Azure Key Vault, or GCP Secret Manager instead of a local .env file."
weight: 10
draft: false
---

Dev-Share requires two application secrets to operate: `JWT_SECRET` (signs authentication tokens) and `ENCRYPTION_KEY` (encrypts sensitive data at rest). You can provide these through a local `.env` file or fetch them automatically from a cloud secret manager at container startup.

## How secret resolution works

When the backend container starts, the entrypoint script resolves secrets in this order:

1. **Environment variables** — if `JWT_SECRET` is already set (from `.env` or Docker Compose), all cloud fetching is skipped.
2. **Cloud secret manager** — if `JWT_SECRET` is not set and a cloud provider variable is present (`AWS_SECRET_ID`, `AZURE_KEYVAULT_NAME`, or `GCP_SECRET_NAME`), secrets are fetched from that provider's secret manager.
3. **Fail** — if neither source provides secrets, the application exits with a configuration error.

This means existing `.env`-based setups continue to work with no changes.

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
Running `./setup.sh` without `--generate-env` skips `.env` creation entirely. This is the expected behavior for cloud-managed deployments.
{{< /callout >}}

## Option B: AWS Secrets Manager

Store your secrets in AWS Secrets Manager as a JSON object and let the entrypoint fetch them at startup.

### Prerequisites

- An AWS account with Secrets Manager access
- The AWS CLI installed on the host (or in the container)
- IAM credentials that allow `secretsmanager:GetSecretValue` on the target secret

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

Set these environment variables before running `docker compose up` (in `.env` or your shell):

```bash
AWS_SECRET_ID=devshare/app-secrets
AWS_DEFAULT_REGION=us-east-1
```

Do not set `JWT_SECRET` or `ENCRYPTION_KEY` — the entrypoint fetches them from Secrets Manager.

### IAM permissions

The host or container needs an IAM policy with at minimum:

```json
{
  "Effect": "Allow",
  "Action": ["secretsmanager:GetSecretValue"],
  "Resource": "arn:aws:secretsmanager:<region>:<account>:secret:devshare/*"
}
```

On EC2, attach an IAM Instance Role so no static credentials are needed on the host.

## Option C: Azure Key Vault

Store each secret as a separate Key Vault secret and let the entrypoint fetch them at startup.

### Prerequisites

- An Azure Key Vault instance
- The Azure CLI installed on the host (or in the container)
- A principal with the `Key Vault Secrets User` role on the vault

### Create the secrets

```bash
az keyvault secret set --vault-name my-devshare-vault --name jwt-secret --value "YOUR_JWT_SECRET_VALUE"
az keyvault secret set --vault-name my-devshare-vault --name encryption-key --value "YOUR_64_HEX_CHAR_ENCRYPTION_KEY"
az keyvault secret set --vault-name my-devshare-vault --name admin-init-token --value "optional-setup-token"
```

### Configure Dev-Share

```bash
AZURE_KEYVAULT_NAME=my-devshare-vault
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

Store your secrets in GCP Secret Manager as a JSON payload and let the entrypoint fetch them at startup.

### Prerequisites

- A GCP project with the Secret Manager API enabled
- The `gcloud` CLI installed on the host (or in the container)
- A service account with `roles/secretmanager.secretAccessor`

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
GCP_SECRET_NAME=devshare-secrets
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

## Security best practices

- **Prefer instance identity over static credentials.** Use IAM Instance Roles (AWS), Managed Identity (Azure), or attached service accounts (GCP) so no long-lived credentials exist on the host.
- **Scope access narrowly.** Grant only `GetSecretValue` / `Secret Accessor` on the specific secret resource, not broad secret manager access.
- **Enable audit logging.** AWS CloudTrail, Azure Monitor, and GCP Cloud Audit Logs all record secret access events.
- **Rotate secrets periodically.** AWS Secrets Manager supports automatic rotation via Lambda. For Azure and GCP, implement rotation via Event Grid or Pub/Sub notifications.

## Next steps

- [Environment variables reference]({{< ref "docs/reference/environment-variables" >}}) for the full list of configuration options
- [Configure cloud credentials]({{< ref "docs/getting-started/cloud-credentials" >}}) to set up Terraform provider authentication
