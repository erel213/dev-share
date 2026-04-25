---
title: "Roles and permissions"
description: "How the User, Editor, and Admin roles work, what each role can do, and how groups restrict which templates members can deploy."
weight: 30
draft: false
lastmod: 2026-04-25
---

Dev-Share uses role-based access control with three roles arranged in a privilege hierarchy.

## Roles

| Role | Who it's for |
|---|---|
| **User** | Developers who deploy environments from existing templates |
| **Editor** | Platform or infra engineers who create and maintain templates |
| **Admin** | System administrators who manage users, groups, and the workspace |

Each role includes all permissions of the role below it. An Admin can do everything an Editor can do; an Editor can do everything a User can do.

## Permission matrix

| Action | User | Editor | Admin |
|---|:---:|:---:|:---:|
| Log in | ✓ | ✓ | ✓ |
| View templates | ✓ | ✓ | ✓ |
| Browse template files and variables | ✓ | ✓ | ✓ |
| Create environments | ✓ | ✓ | ✓ |
| Plan / apply / destroy own environments | ✓ | ✓ | ✓ |
| View environment outputs | ✓ | ✓ | ✓ |
| Create and manage templates | | ✓ | ✓ |
| Define template variables | | ✓ | ✓ |
| Create and manage workspaces | | ✓ | ✓ |
| View all environments (admin dashboard) | | | ✓ |
| Invite users | | | ✓ |
| Reset user passwords | | | ✓ |
| Delete users | | | ✓ |
| Create and manage groups | | | ✓ |
| Assign users to groups | | | ✓ |
| Control group template access | | | ✓ |

## Groups and template access

By default, a User with no group membership can see all templates in the workspace. Groups let admins restrict visibility.

When a group has `access_all_templates` disabled, members can only see templates explicitly granted to that group. If a user belongs to multiple groups, they see the union of all templates those groups allow.

When a group has `access_all_templates` enabled, members see every template in the workspace (the default behavior).

**Typical use case**: an ops team uploads 20 templates covering different services. Frontend developers should only see the 3 templates relevant to their stack. An admin creates a "Frontend" group, adds the relevant templates, and adds all frontend developers as members.

## Assigning roles

Roles are assigned when a user is invited. The Admin sets the role in the Invite User dialog. There is currently no in-place role-change UI; to change a user's role, delete and re-invite them.

See [User and group management]({{< ref "docs/guides/user-management" >}}) for the step-by-step invite flow.
