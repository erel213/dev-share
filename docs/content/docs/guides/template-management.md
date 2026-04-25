---
title: "Template management"
description: "Upload Terraform templates, define and auto-parse input variables, browse template files, and manage the full template lifecycle."
weight: 20
draft: false
lastmod: 2026-04-25
---

Templates are the Terraform definitions that users deploy as environments. An admin (or Editor) creates a template once; users then create any number of environment instances from it.

For background on what templates are, see [What is Dev-Share]({{< ref "docs/concepts/what-is-devshare" >}}).

## Prerequisites

- You are logged in as an **Admin** or **Editor**
- You have a set of Terraform files (`.tf`, `.hcl`) ready to upload

## View all templates

Navigate to **Templates** in the left sidebar.

![Templates list](/images/guides/template-list.png)

The table shows each template's name and timestamps. Click a template name to open the template browser.

## Create a template

1. Click **Create Template** in the top-right corner.

   ![Create template dialog](/images/guides/create-template-dialog.png)

2. Enter a **Template name**.

3. Drag and drop your Terraform files onto the upload area, or click to browse. You can upload multiple files at once. Supported formats: `.tf`, `.hcl`, and any supporting files your Terraform module references (e.g., `.json` variable definition files).

4. Click **Create**.

Dev-Share stores the files on disk and records the template in the database. The template is immediately available for users to deploy.

## Define template variables

Variables are the input parameters users fill in when creating an environment (e.g., `region`, `instance_type`, `db_password`). You can define them manually or let Dev-Share auto-parse them from your HCL.

### Auto-parse variables

After creating a template:

1. Open the template from the Templates list.
2. Go to the **Variables** tab.
3. Click **Parse variables from files**.

Dev-Share scans your `.tf` files for `variable` blocks and creates a variable entry for each one, pre-filling the key, type, description, and default value from the HCL.

### Add a variable manually

1. On the **Variables** tab, click **Add variable**.
2. Fill in the fields:

   | Field | Description |
   |---|---|
   | **Key** | The variable name, matching the HCL `variable` block name |
   | **Type** | `string`, `number`, `bool`, `list`, or `map` |
   | **Description** | Shown to users as help text when filling in the form |
   | **Default value** | Pre-fills the user's form; leave empty if the variable is required |
   | **Required** | If checked, users must supply a value before applying |
   | **Sensitive** | If checked, the value is encrypted at rest and never displayed in the UI after saving |
   | **Validation regex** | Optional regex pattern the value must match |

3. Click **Save**.

### Best practices

- Mark database passwords, API keys, and any secret as **Sensitive**. Dev-Share encrypts these values and redacts them from the UI.
- Use **Default value** for optional variables like `instance_type` to give users a sensible starting point without forcing them to look up valid values.
- Use **Validation regex** to catch bad input early (e.g., `^[a-z][a-z0-9-]*$` for resource name fields).

## Browse a template

Click any template name to open the template browser.

![Template browser](/images/guides/template-browser.png)

The browser has two tabs:

**Files** — a file tree on the left and a syntax-highlighted content viewer on the right. Click any file in the tree to view its contents.

**Variables** — lists all defined variables with their type, description, required flag, and default value.

Users also see this view when selecting a template to deploy. See [Browsing templates]({{< ref "docs/guides/browsing-templates" >}}).

## Edit a template

1. In the Templates list, click the **pencil icon** on the template row.
2. Update the template name.
3. Click **Save**.

{{< callout type="info" >}}
Editing a template's files directly through the UI is not yet supported. To update the Terraform files, delete the template and re-create it with the new files. Environments created from the old template are unaffected — they have their own copy of the files in the execution directory.
{{< /callout >}}

## Delete a template

1. In the Templates list, click the **trash icon** on the template row.
2. Confirm the deletion.

Deleting a template does not destroy environments that were already created from it. Those environments continue to exist and can still be managed (applied, destroyed, deleted) independently.
