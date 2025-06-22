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

  // toolbox (must be prefixed with 'toolbox_')
  TOOLBOX_DELETE_FILE = 'toolbox_delete_file',
  TOOLBOX_DOWNLOAD_FILE = 'toolbox_download_file',
  TOOLBOX_CREATE_FOLDER = 'toolbox_create_folder',
  TOOLBOX_MOVE_FILE = 'toolbox_move_file',
  TOOLBOX_SET_FILE_PERMISSIONS = 'toolbox_set_file_permissions',
  TOOLBOX_REPLACE_IN_FILES = 'toolbox_replace_in_files',
  TOOLBOX_UPLOAD_FILE = 'toolbox_upload_file',
  TOOLBOX_BULK_UPLOAD_FILES = 'toolbox_bulk_upload_files',
  TOOLBOX_GIT_ADD_FILES = 'toolbox_git_add_files',
  TOOLBOX_GIT_CREATE_BRANCH = 'toolbox_git_create_branch',
  TOOLBOX_GIT_DELETE_BRANCH = 'toolbox_git_delete_branch',
  TOOLBOX_GIT_CLONE_REPOSITORY = 'toolbox_git_clone_repository',
  TOOLBOX_GIT_COMMIT_CHANGES = 'toolbox_git_commit_changes',
  TOOLBOX_GIT_PULL_CHANGES = 'toolbox_git_pull_changes',
  TOOLBOX_GIT_PUSH_CHANGES = 'toolbox_git_push_changes',
  TOOLBOX_GIT_CHECKOUT_BRANCH = 'toolbox_git_checkout_branch',
  TOOLBOX_EXECUTE_COMMAND = 'toolbox_execute_command',
  TOOLBOX_CREATE_SESSION = 'toolbox_create_session',
  TOOLBOX_SESSION_EXECUTE_COMMAND = 'toolbox_session_execute_command',
  TOOLBOX_DELETE_SESSION = 'toolbox_delete_session',
}
