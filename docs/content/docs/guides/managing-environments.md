---
title: "Managing environments"
description: "Create cloud environments from Terraform templates, run plan and apply, view outputs, set a TTL, and destroy resources from the Dev-Share UI."
weight: 40
draft: false
lastmod: 2026-04-25
---

An environment is a deployed instance of a Terraform template with a specific set of variable values. You own the environments you create and can apply, destroy, and delete them at any time.

For a deeper explanation of what environments are and how they're provisioned, see [Architecture and deployment model]({{< ref "docs/concepts/architecture" >}}).

## Prerequisites

- You are logged in (any role)
- At least one template exists in your workspace (an Admin or Editor must create it first)
- Cloud credentials are configured in the backend (contact your Admin if you're unsure)

## View your environments

Navigate to **Environments** in the left sidebar.

![Environments page](/images/guides/environments.png)

The table shows your environments with their name, template, current status, and when they were created.

**Admin users** also see an **All Environments** dashboard on the home page, showing every environment across all users with search and filter controls.

## Create an environment

1. Click **New Environment**.

   ![Create environment dialog](/images/guides/create-environment-dialog.png)

2. Fill in:

   | Field | Description |
   |---|---|
   | **Template** | Select the Terraform template to deploy. Only templates your groups allow are shown. |
   | **Name** | A short name for this environment (3–255 characters). Must be unique within your account. |
   | **Description** | Optional. A note about what this environment is for. |
   | **Time to live** | How long before the environment auto-destructs. Options: 1 hour, 4 hours, 24 hours, or None (persists until manually destroyed). |

3. Click **Create**.

The environment is created with status `pending`. The next step is to fill in variable values, then plan and apply.

## Set variable values

If the template has input variables, you need to supply values before applying. Click the environment row to expand it and find the variable form, or navigate to the environment detail view.

For each variable:

- Required variables (marked with `*`) must have a value before you can apply.
- Sensitive variables show a masked input. The value is encrypted and never displayed again after saving.
- Variables with a default value are pre-filled — you can leave them as-is or override.

Click **Save variables** when done.

## Plan

Running a plan shows you what Terraform intends to create, modify, or destroy — without making any changes.

Click the **Plan** button on the environment row. The status changes to `planning` while Terraform runs. When complete, the status returns to `initialized` and a plan summary is available.

Plan is optional. You can skip directly to Apply if you trust the template.

## Apply

Applying provisions the actual cloud infrastructure.

Click the **Apply** button. The status changes to `applying`. Terraform runs `apply -auto-approve` with your variable values.

When apply succeeds, the status changes to `ready`. Any [Terraform outputs](https://developer.hashicorp.com/terraform/language/values/outputs) defined in the template are now available.

## View outputs

After a successful apply, click the **expand arrow** on the environment row to reveal the Terraform outputs section. Outputs are key-value pairs defined in the template — typically things like endpoint URLs, resource IDs, or connection strings.

## Destroy

Destroying an environment tears down the cloud resources Terraform created. The environment record remains in the table with status `destroyed` so you have a record of it.

Click the **Destroy** button on the environment row and confirm. The status changes to `destroying` while Terraform runs `destroy -auto-approve`. When complete, the status is `destroyed`.

{{< callout type="info" >}}
Always destroy before deleting. Deleting the environment record without destroying first leaves cloud resources running and accumulating costs.
{{< /callout >}}

## Delete the environment record

After destroying, you can remove the environment from the table entirely by clicking the **trash icon** and confirming.

## TTL and auto-cleanup

If you set a TTL when creating the environment, Dev-Share automatically destroys it when the TTL expires. A background process checks continuously and triggers a destroy when `created_at + ttl_seconds` is reached.

If the auto-destroy fails (e.g., cloud API error), the environment enters `error` status and you will need to destroy it manually.

## Error states

If an apply or destroy fails, the environment enters `error` status. Common causes:

- Missing or incorrect variable values
- Cloud credentials not configured or expired
- Terraform provider errors (quota limits, invalid resource names)

To recover:

1. Check the error message by expanding the environment row.
2. Fix the underlying issue (update variables, check credentials with your Admin).
3. Re-run Apply (or Destroy if you want to clean up).
