/* tslint:disable */

/**
 * Daytona
 * Daytona AI platform API Docs
 *
 * The version of the OpenAPI document: 1.0
 * Contact: support@daytona.com
 *
 * NOTE: This class is auto generated by OpenAPI Generator (https://openapi-generator.tech).
 * https://openapi-generator.tech
 * Do not edit the class manually.
 */

/**
 *
 * @export
 * @interface CreateAuditLog
 */
export interface CreateAuditLog {
  /**
   *
   * @type {string}
   * @memberof CreateAuditLog
   */
  actorId: string
  /**
   *
   * @type {string}
   * @memberof CreateAuditLog
   */
  actorEmail: string
  /**
   *
   * @type {string}
   * @memberof CreateAuditLog
   */
  organizationId?: string
  /**
   *
   * @type {string}
   * @memberof CreateAuditLog
   */
  action: CreateAuditLogActionEnum
  /**
   *
   * @type {string}
   * @memberof CreateAuditLog
   */
  targetType?: CreateAuditLogTargetTypeEnum
  /**
   *
   * @type {string}
   * @memberof CreateAuditLog
   */
  targetId?: string
}

export const CreateAuditLogActionEnum = {
  CREATE: 'create',
  READ: 'read',
  UPDATE: 'update',
  DELETE: 'delete',
  LOGIN: 'login',
  SET_DEFAULT: 'set_default',
  UPDATE_ROLE: 'update_role',
  UPDATE_ASSIGNED_ROLES: 'update_assigned_roles',
  UPDATE_QUOTA: 'update_quota',
  SUSPEND: 'suspend',
  UNSUSPEND: 'unsuspend',
  ACCEPT: 'accept',
  DECLINE: 'decline',
  LINK_ACCOUNT: 'link_account',
  UNLINK_ACCOUNT: 'unlink_account',
  LEAVE_ORGANIZATION: 'leave_organization',
  REGENERATE_KEY_PAIR: 'regenerate_key_pair',
  UPDATE_SCHEDULING: 'update_scheduling',
  START: 'start',
  STOP: 'stop',
  REPLACE_LABELS: 'replace_labels',
  CREATE_BACKUP: 'create_backup',
  UPDATE_PUBLIC_STATUS: 'update_public_status',
  SET_AUTO_STOP_INTERVAL: 'set_auto_stop_interval',
  SET_AUTO_ARCHIVE_INTERVAL: 'set_auto_archive_interval',
  SET_AUTO_DELETE_INTERVAL: 'set_auto_delete_interval',
  ARCHIVE: 'archive',
  GET_PORT_PREVIEW_URL: 'get_port_preview_url',
  SET_GENERAL_STATUS: 'set_general_status',
  ACTIVATE: 'activate',
  DEACTIVATE: 'deactivate',
  TOOLBOX_DELETE_FILE: 'toolbox_delete_file',
  TOOLBOX_DOWNLOAD_FILE: 'toolbox_download_file',
  TOOLBOX_CREATE_FOLDER: 'toolbox_create_folder',
  TOOLBOX_MOVE_FILE: 'toolbox_move_file',
  TOOLBOX_SET_FILE_PERMISSIONS: 'toolbox_set_file_permissions',
  TOOLBOX_REPLACE_IN_FILES: 'toolbox_replace_in_files',
  TOOLBOX_UPLOAD_FILE: 'toolbox_upload_file',
  TOOLBOX_BULK_UPLOAD_FILES: 'toolbox_bulk_upload_files',
  TOOLBOX_GIT_ADD_FILES: 'toolbox_git_add_files',
  TOOLBOX_GIT_CREATE_BRANCH: 'toolbox_git_create_branch',
  TOOLBOX_GIT_DELETE_BRANCH: 'toolbox_git_delete_branch',
  TOOLBOX_GIT_CLONE_REPOSITORY: 'toolbox_git_clone_repository',
  TOOLBOX_GIT_COMMIT_CHANGES: 'toolbox_git_commit_changes',
  TOOLBOX_GIT_PULL_CHANGES: 'toolbox_git_pull_changes',
  TOOLBOX_GIT_PUSH_CHANGES: 'toolbox_git_push_changes',
  TOOLBOX_GIT_CHECKOUT_BRANCH: 'toolbox_git_checkout_branch',
  TOOLBOX_EXECUTE_COMMAND: 'toolbox_execute_command',
  TOOLBOX_CREATE_SESSION: 'toolbox_create_session',
  TOOLBOX_SESSION_EXECUTE_COMMAND: 'toolbox_session_execute_command',
  TOOLBOX_DELETE_SESSION: 'toolbox_delete_session',
  TOOLBOX_COMPUTER_USE_START: 'toolbox_computer_use_start',
  TOOLBOX_COMPUTER_USE_STOP: 'toolbox_computer_use_stop',
  TOOLBOX_COMPUTER_USE_RESTART_PROCESS: 'toolbox_computer_use_restart_process',
} as const

export type CreateAuditLogActionEnum = (typeof CreateAuditLogActionEnum)[keyof typeof CreateAuditLogActionEnum]
export const CreateAuditLogTargetTypeEnum = {
  API_KEY: 'api_key',
  ORGANIZATION: 'organization',
  ORGANIZATION_INVITATION: 'organization_invitation',
  ORGANIZATION_ROLE: 'organization_role',
  ORGANIZATION_USER: 'organization_user',
  DOCKER_REGISTRY: 'docker_registry',
  RUNNER: 'runner',
  SANDBOX: 'sandbox',
  SNAPSHOT: 'snapshot',
  USER: 'user',
  VOLUME: 'volume',
} as const

export type CreateAuditLogTargetTypeEnum =
  (typeof CreateAuditLogTargetTypeEnum)[keyof typeof CreateAuditLogTargetTypeEnum]
