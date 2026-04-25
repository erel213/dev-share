---
title: "Browsing templates"
description: "Explore a template's Terraform files and required input variables in the template browser before deploying an environment."
weight: 50
draft: false
lastmod: 2026-04-25
---

Before creating an environment, you can inspect a template's Terraform files and understand what variables you'll need to supply. This helps you prepare values in advance — especially for variables that require you to look up a resource ID, region name, or other cloud-specific value.

## Open a template

1. Navigate to **Templates** in the left sidebar.
2. Click the template name in the list.

## Files tab

The Files tab shows the full content of every file in the template.

![Template browser](/images/guides/template-browser.png)

The **left panel** is a file tree. Folders expand and collapse when clicked. Click any file to load it in the content viewer.

The **right panel** shows the selected file with syntax highlighting. This is read-only — you cannot edit files here.

Common things to look for in the files:

- `resource` blocks — what cloud resources will be created
- `variable` blocks — what input parameters exist (the Variables tab shows these more clearly)
- `output` blocks — what values will be available after a successful apply

## Variables tab

The Variables tab lists every input parameter the template accepts.

| Column | Meaning |
|---|---|
| **Key** | The variable name you'll fill in when creating an environment |
| **Type** | The expected data type (`string`, `number`, `bool`, `list`, `map`) |
| **Description** | Help text from the template author explaining what the variable controls |
| **Required** | Whether you must supply a value (no default exists) |
| **Default** | Pre-filled value if you don't override it |
| **Sensitive** | If marked, the value is masked in the UI and encrypted at rest |

Use this tab to confirm you have all required values ready before creating an environment.

## Create an environment from this template

Once you're ready, click **Create Environment** at the top of the template browser. This opens the [Create Environment dialog]({{< ref "docs/guides/managing-environments" >}}) pre-filled with this template selected.
