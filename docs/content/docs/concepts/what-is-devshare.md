---
title: "What is Dev-Share"
description: "Understand the problem Dev-Share solves, how it works, and the core terms you'll encounter throughout the documentation."
weight: 10
draft: false
lastmod: 2026-04-25
---

Dev-Share is a self-hosted platform that lets developers deploy isolated cloud environments from a web UI, without needing to know Terraform.

## The problem

When a development team needs per-developer cloud environments — a dedicated database, a staging stack, a short-lived sandbox — the usual approaches break down:

- **Direct Terraform**: every developer needs to learn HCL, manage state files, and handle credentials. Mistakes cost money and cause outages.
- **Shared environments**: developers step on each other. No isolation, no reproducibility.
- **Manual provisioning by ops**: creates a bottleneck. Every environment request becomes a ticket.

What teams need is a way for platform engineers to define *what* can be deployed, and for developers to deploy *instances* of those definitions on demand — with guardrails, lifecycle management, and no Terraform expertise required.

## The solution

Dev-Share wraps Terraform behind a web app:

1. **Admins** upload Terraform files as *templates* and define the input variables users must fill in.
2. **Users** pick a template, fill in the variable form, and click **Apply**. Dev-Share runs `terraform apply` on the backend.
3. When the environment is no longer needed, users click **Destroy**. Optional TTLs auto-clean environments after a set time.

No one outside the admin role touches Terraform directly. Cloud credentials are injected into the container at deploy time — users never see them.

## Key terms

| Term | Meaning |
|---|---|
| **Workspace** | Top-level container. Every template, environment, and user belongs to a workspace. Created during initial setup. |
| **Template** | A set of Terraform files uploaded by an admin. Defines what infrastructure can be provisioned. |
| **Template variable** | An input parameter for a template (e.g., `region`, `instance_type`). Variables can be required, optional, sensitive, or auto-parsed from HCL. |
| **Environment** | A deployed instance of a template with specific variable values. Has a lifecycle: `pending → applying → ready → destroyed`. |
| **Group** | A named set of users. Groups control which templates members can see when creating environments. |
| **Role** | Determines what a user can do: **User** (deploy environments), **Editor** (+ manage templates), **Admin** (+ manage users and groups). |

## What Dev-Share does not do

- **Manage cloud credentials** — Dev-Share delegates entirely to Terraform's credential chain. See [Configure cloud credentials]({{< ref "docs/getting-started/cloud-credentials" >}}).
- **Write Terraform** — Dev-Share runs the code admins upload. It does not generate infrastructure definitions.
- **Replace a CI/CD pipeline** — Dev-Share is for on-demand developer environments, not automated deployment pipelines.
