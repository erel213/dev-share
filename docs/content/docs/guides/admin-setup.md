---
title: "Initial setup"
description: "Walk through the five-step Dev-Share setup wizard to create your admin account and first workspace before inviting your team."
weight: 10
draft: false
lastmod: 2026-04-25
---

The first time Dev-Share starts, no users or workspaces exist. A one-time setup wizard walks you through creating both. This wizard is only accessible before initialization — once complete, it is locked permanently.

## Prerequisites

Dev-Share must be running before you begin. Follow the [Quick Start]({{< ref "docs/getting-started/quick-start" >}}) to bring up the Docker containers, then return here.

## Open the setup wizard

Navigate to `http://localhost:3000/setup` in your browser.

If Dev-Share is already initialized, this URL redirects to the login page.

![Setup wizard welcome screen](/images/guides/setup-wizard.png)

## Step 1: Welcome

The welcome screen introduces the wizard. Click **Get Started** to proceed.

## Step 2: Create your admin account

Fill in:

- **Name** — your display name
- **Email** — used to log in; must be a valid email address
- **Password** — must be at least 8 characters and include an uppercase letter, a lowercase letter, a number, and a special character
- **Confirm password** — must match

Click **Next**.

## Step 3: Create your workspace

Fill in:

- **Workspace name** — required; a short, descriptive name (e.g., `Acme Infra`)
- **Description** — optional; a longer description of what this workspace covers

Click **Next**.

## Step 4: Review

The wizard shows a summary of your admin account and workspace details. Confirm everything looks correct, then click **Create**.

This sends a single `POST /admin/init` request. On success, the system is initialized.

## Step 5: Done

A success screen confirms that your account and workspace have been created. Click the button to go to the dashboard and log in with the credentials you just set.

## What was created

After setup completes:

- One **workspace** — the container for all templates, environments, and users
- One **admin user** — your account with full Admin privileges

From here, your next steps are:

- [Upload Terraform templates]({{< ref "docs/guides/template-management" >}}) so users have infrastructure to deploy
- [Invite your team]({{< ref "docs/guides/user-management" >}}) with the appropriate roles

## Locked out?

If you need to reset and run setup again, wipe the database and restart:

```bash
./clean_db.sh
./setup.sh
```

{{< callout type="warning" >}}
`clean_db.sh` deletes all data — users, templates, environments, and all history. Only do this on a fresh install or a dev machine.
{{< /callout >}}
