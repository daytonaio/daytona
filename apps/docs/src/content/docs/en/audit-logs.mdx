---
title: Audit Logs
description: View and monitor all user actions across your Daytona organization.
---

Daytona audit logs provide a detailed record of user and system activity across your organization. Use this feature to track sandbox lifecycle events, user access, system changes, and more.

- **Security audits**: monitor for unauthorized access or sandbox misuse
- **Debugging**: understand sandbox lifecycle issues (e.g. failed starts)
- **Compliance Export**: export logs for internal or external audits (coming soon)

Audit logs are available to [admins](/docs/en/organizations#organization-roles) with full access and [members](/docs/en/organizations#organization-roles) with audit log permissions. Contact your organization administrator if you cannot access audit logs.

## Access from Dashboard

Access the audit logs page directly from [Daytona Dashboard ↗](https://app.daytona.io/dashboard/audit-logs). The audit logs page displays a list of all audit logs for your organization, including the following columns:

- **Time**: the timestamp of the action
- **User**: the user who performed the action
- [Actions](#actions): the action performed
- [Targets](#targets): the resource affected by the action
- [Outcomes](#outcomes): the result of the action

To filter audit logs by time, use the date range picker in the top-left corner of the page.

## Real-time updates

Daytona provides real-time updates of audit logs. Enable the **Auto Refresh** toggle in the top-right corner of the [Daytona Audit Logs ↗](https://app.daytona.io/dashboard/audit-logs) page to automatically refresh logs as new events occur.

## Programmatic management

Daytona provides API endpoints for programmatic access to audit logs.

### Get all audit logs

To get all audit logs, use the following API endpoint:

```bash
curl https://app.daytona.io/api/audit \
  --header 'Authorization: Bearer YOUR_API_KEY'
```

For more information, see the [API](/docs/en/tools/api/#daytona/tag/audit) reference:

> [**Get all audit logs (API)**](/docs/en/tools/api/#daytona/tag/audit/GET/audit-logs)

### Get audit logs for organization

To get audit logs for a specific organization, use the following API endpoint:

```bash
curl https://app.daytona.io/api/audit/organizations/{organizationId} \
  --header 'Authorization: Bearer YOUR_API_KEY'
```

For more information, see the [API](/docs/en/tools/api/#daytona/tag/audit) reference:

> [**Get audit logs for organization (API)**](/docs/en/tools/api/#daytona/tag/audit/GET/audit/organizations/{organizationId})

## Log Structure

Each audit log entry contains the following fields:

| Field                | Type   | Description                                          |
| -------------------- | ------ | ---------------------------------------------------- |
| **`id`**             | string | Unique log entry identifier                          |
| **`actorId`**        | string | ID of the user who performed the action              |
| **`actorEmail`**     | string | Email of the user who performed the action           |
| **`organizationId`** | string | Organization ID                                      |
| **`action`**         | string | Operation executed (e.g., `create`, `start`, `stop`) |
| **`targetType`**     | string | Resource type affected (e.g., `sandbox`, `snapshot`) |
| **`targetId`**       | string | ID of the affected resource                          |
| **`statusCode`**     | number | HTTP status code of the result                       |
| **`errorMessage`**   | string | Error message if the action failed                   |
| **`ipAddress`**      | string | IP address of the request origin                     |
| **`userAgent`**      | string | User agent of the request origin                     |
| **`source`**         | string | Source of the action                                 |
| **`metadata`**       | object | Additional context about the action                  |
| **`createdAt`**      | string | ISO 8601 timestamp of when the action occurred       |

## Actions

Below is the complete list of actions logged by Daytona:

```text
create, read, update, delete, login,
set_default, update_access, update_quota, update_region_quota,
suspend, unsuspend, accept, decline,
link_account, unlink_account, leave_organization,
regenerate_key_pair, update_scheduling,
start, stop, replace_labels, create_backup,
update_public_status, set_auto_stop_interval,
set_auto_archive_interval, set_auto_delete_interval, archive,
get_port_preview_url, set_general_status, activate, deactivate,
update_network_settings, get_webhook_app_portal_access,
send_webhook_message, initialize_webhooks,
update_sandbox_default_limited_network_egress,
create_ssh_access, revoke_ssh_access,
regenerate_proxy_api_key,regenerate_ssh_gateway_api_key,regenerate_snapshot_manager_credentials,
toolbox_delete_file, toolbox_download_file, toolbox_create_folder,
toolbox_move_file, toolbox_set_file_permissions, toolbox_replace_in_files,
toolbox_upload_file, toolbox_bulk_upload_files,
toolbox_git_add_files, toolbox_git_create_branch, toolbox_git_delete_branch,
toolbox_git_clone_repository, toolbox_git_commit_changes,
toolbox_git_pull_changes, toolbox_git_push_changes,
toolbox_git_checkout_branch, toolbox_execute_command,
toolbox_create_session, toolbox_session_execute_command,
toolbox_delete_session, toolbox_computer_use_start,
toolbox_computer_use_stop, toolbox_computer_use_restart_process
```

## Targets

Each action targets a specific resource type. Possible targets include:

```text
api_key, organization, organization_invitation,
organization_role, organization_user, docker_registry,
runner, sandbox, snapshot, user, volume
```

## Outcomes

The outcome field indicates the result of the action. Statuses follow standard HTTP semantics:

| **Outcome** | **Description**               |
| ----------- | ----------------------------- |
| Info        | Informational (1xx codes)     |
| Success     | Action succeeded (2xx codes)  |
| Redirect    | Redirects (3xx codes)         |
| Error       | Client/server error (4xx/5xx) |
