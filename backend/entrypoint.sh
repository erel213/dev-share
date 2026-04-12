#!/bin/sh
set -e

# If secrets are not already provided via environment (.env / Docker Compose),
# attempt to fetch them from the cloud provider's secret manager.
if [ -z "$JWT_SECRET" ]; then

  if [ -n "$AWS_SECRET_ID" ]; then
    echo "Fetching secrets from AWS Secrets Manager..."
    SECRET_JSON=$(aws secretsmanager get-secret-value \
      --secret-id "$AWS_SECRET_ID" \
      --query SecretString \
      --output text)

  elif [ -n "$AZURE_KEYVAULT_NAME" ]; then
    echo "Fetching secrets from Azure Key Vault..."
    SECRET_JSON=$(jq -n \
      --arg jwt "$(az keyvault secret show --vault-name "$AZURE_KEYVAULT_NAME" --name jwt-secret --query value -o tsv)" \
      --arg enc "$(az keyvault secret show --vault-name "$AZURE_KEYVAULT_NAME" --name encryption-key --query value -o tsv)" \
      --arg admin "$(az keyvault secret show --vault-name "$AZURE_KEYVAULT_NAME" --name admin-init-token --query value -o tsv 2>/dev/null || echo '')" \
      '{JWT_SECRET: $jwt, ENCRYPTION_KEY: $enc, ADMIN_INIT_TOKEN: $admin}')

  elif [ -n "$GCP_SECRET_NAME" ]; then
    echo "Fetching secrets from GCP Secret Manager..."
    SECRET_JSON=$(gcloud secrets versions access latest \
      --secret="$GCP_SECRET_NAME" \
      --format="get(payload.data)" | base64 -d)
  fi

  if [ -n "$SECRET_JSON" ]; then
    export JWT_SECRET=$(echo "$SECRET_JSON" | jq -r '.JWT_SECRET')
    export ENCRYPTION_KEY=$(echo "$SECRET_JSON" | jq -r '.ENCRYPTION_KEY')
    export ADMIN_INIT_TOKEN=$(echo "$SECRET_JSON" | jq -r '.ADMIN_INIT_TOKEN // empty')
  fi
fi

echo "Running database migrations..."
./migrate

echo "Starting server..."
exec ./main
