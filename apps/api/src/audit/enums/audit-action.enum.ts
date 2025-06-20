/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

export enum AuditAction {
  // generic
  CREATE = 'create',
  READ = 'read',
  UPDATE = 'update',
  DELETE = 'delete',
  LOGIN = 'login',

  // docker registry
  DOCKER_REGISTRY_SET_DEFAULT = 'set_default',

  // organization user
  ORGANIZATION_USER_UPDATE_ROLE = 'update_role',
  ORGANIZATION_USER_UPDATE_ASSIGNED_ROLES = 'update_assigned_roles',

  // organization
  ORGANIZATION_UPDATE_QUOTA = 'update_quota',
  ORGANIZATION_SUSPEND = 'suspend',
  ORGANIZATION_UNSUSPEND = 'unsuspend',

  // organization invitation
  ORGANIZATION_INVITATION_ACCEPT = 'accept',
  ORGANIZATION_INVITATION_DECLINE = 'decline',

  // user
  USER_LINK_ACCOUNT = 'link_account',
  USER_UNLINK_ACCOUNT = 'unlink_account',
  USER_LEAVE_ORGANIZATION = 'leave_organization',
  USER_REGENERATE_KEY_PAIR = 'regenerate_key_pair',

  // runner
  RUNNER_UPDATE_SCHEDULING = 'update_scheduling',

  // sandbox
  SANDBOX_START = 'start',
  SANDBOX_STOP = 'stop',
  SANDBOX_REPLACE_LABELS = 'replace_labels',
  SANDBOX_CREATE_BACKUP = 'create_backup',
  SANDBOX_UPDATE_PUBLIC_STATUS = 'update_public_status',
  SANDBOX_SET_AUTO_STOP_INTERVAL = 'set_auto_stop_interval',
  SANDBOX_SET_AUTO_ARCHIVE_INTERVAL = 'set_auto_archive_interval',
  SANDBOX_ARCHIVE = 'archive',

  // snapshot
  SNAPSHOT_TOGGLE_STATE = 'toggle_state',
  SNAPSHOT_SET_GENERAL_STATUS = 'set_general_status',

  // toolbox
  TOOLBOX_DELETE_FILE = 'delete_file',
  TOOLBOX_DOWNLOAD_FILE = 'download_file',
  TOOLBOX_CREATE_FOLDER = 'create_folder',
  TOOLBOX_MOVE_FILE = 'move_file',
  TOOLBOX_SET_FILE_PERMISSIONS = 'set_file_permissions',
  TOOLBOX_REPLACE_IN_FILES = 'replace_in_files',
  TOOLBOX_UPLOAD_FILE = 'upload_file',
  TOOLBOX_BULK_UPLOAD_FILES = 'bulk_upload_files',
  TOOLBOX_GIT_ADD_FILES = 'git_add_files',
  TOOLBOX_GIT_CREATE_BRANCH = 'git_create_branch',
  TOOLBOX_GIT_DELETE_BRANCH = 'git_delete_branch',
  TOOLBOX_GIT_CLONE_REPOSITORY = 'git_clone_repository',
  TOOLBOX_GIT_COMMIT_CHANGES = 'git_commit_changes',
  TOOLBOX_GIT_PULL_CHANGES = 'git_pull_changes',
  TOOLBOX_GIT_PUSH_CHANGES = 'git_push_changes',
  TOOLBOX_GIT_CHECKOUT_BRANCH = 'git_checkout_branch',
  TOOLBOX_EXECUTE_COMMAND = 'execute_command',
  TOOLBOX_CREATE_SESSION = 'create_session',
  TOOLBOX_SESSION_EXECUTE_COMMAND = 'session_execute_command',
  TOOLBOX_DELETE_SESSION = 'delete_session',
}
