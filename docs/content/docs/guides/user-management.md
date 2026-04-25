---
title: "User and group management"
description: "Invite users, assign roles, reset passwords, create groups, and control which templates each group can access."
weight: 30
draft: false
lastmod: 2026-04-25
---

User and group management is available to **Admin** users only. Editors and regular Users cannot access these pages.

For an overview of roles and what each one can do, see [Roles and permissions]({{< ref "docs/concepts/roles-and-permissions" >}}).

## Managing users

Navigate to **Users** in the left sidebar (visible to Admins only).

![Users page](/images/guides/user-management.png)

The table lists every user in the system with their name, email, role, and when they joined.

### Invite a user

1. Click **Invite User**.

   ![Invite user dialog](/images/guides/invite-user-dialog.png)

2. Fill in:
   - **Name** — the user's display name
   - **Email** — used to log in
   - **Role** — select `User`, `Editor`, or `Admin`

3. Click **Invite**.

Dev-Share generates a temporary password and displays it in the dialog. **Copy the password now** — it is shown only once. Send it to the user out-of-band (Slack, email, etc.). The user should change their password after first login.

{{< callout type="warning" >}}
The generated password is shown a single time. If you close the dialog before copying it, you will need to reset the password using the reset flow below.
{{< /callout >}}

### Reset a user's password

1. In the Users table, click the **key icon** on the user's row.
2. Confirm the reset.

A new temporary password is generated and displayed. Copy and send it to the user as you would for a new invite.

### Change a user's role

There is no in-place role editor. To change a user's role:

1. Delete the user (see below).
2. Re-invite them with the new role.

The user will receive a new temporary password when re-invited.

### Delete a user

1. Click the **trash icon** on the user's row.
2. Confirm the deletion.

Deleting a user does not delete their environments. Those environments remain in the system and can still be managed by an Admin.

---

## Managing groups

Groups let you control which templates a set of users can see. Navigate to **Groups** in the left sidebar.

![Groups page](/images/guides/group-management.png)

The table shows each group's name, description, template access setting, and actions.

### Create a group

1. Click **Create Group**.

   ![Create group dialog](/images/guides/create-group-dialog.png)

2. Fill in:
   - **Name** — a short, descriptive name (e.g., `Frontend Team`)
   - **Description** — optional
   - **Access all templates** — if checked, group members see every template in the workspace (same as no group restriction)

3. Click **Create**.

### Add members to a group

1. Click the **members icon** (person silhouette) on the group row.

   ![Manage members dialog](/images/guides/manage-members-dialog.png)

2. The dialog lists current members and all available users. Use the **Add** button next to a user to add them to the group.

3. To remove a member, click **Remove** next to their name.

Changes take effect immediately. Users will see their updated template list on their next page load.

### Grant a group access to specific templates

Use this when the group has **Access all templates** disabled.

1. Click the **templates icon** on the group row.
2. The dialog lists templates currently granted to the group and all available templates.
3. Use **Add** to grant access and **Remove** to revoke it.

If a user belongs to multiple groups, they see the union of all templates those groups allow.

### Edit a group

Click the **pencil icon** on the group row to update the name, description, or `access_all_templates` flag.

### Delete a group

Click the **trash icon** on the group row and confirm. Deleting a group removes the access restrictions it imposed — members revert to the workspace default (access all templates, unless they belong to another restricting group).
