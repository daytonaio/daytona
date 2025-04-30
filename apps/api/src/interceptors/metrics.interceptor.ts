/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import {
  Injectable,
  NestInterceptor,
  ExecutionContext,
  CallHandler,
  OnApplicationShutdown,
  Logger,
} from '@nestjs/common'
import { Observable } from 'rxjs'
import { tap } from 'rxjs/operators'
import { PostHog } from 'posthog-node'
import { WorkspaceDto } from '../workspace/dto/workspace.dto'
import { DockerRegistryDto } from '../docker-registry/dto/docker-registry.dto'
import { CreateWorkspaceDto } from '../workspace/dto/create-workspace.dto'
import { Request } from 'express'
import { CreateImageDto } from '../workspace/dto/create-image.dto'
import { ImageDto } from '../workspace/dto/image.dto'
import { ToggleStateDto } from '../workspace/dto/toggle-state.dto'
import { ResizeDto } from '../workspace/dto/resize.dto'
import { CreateOrganizationDto } from '../organization/dto/create-organization.dto'
import { UpdateOrganizationQuotaDto } from '../organization/dto/update-organization-quota.dto'
import { OrganizationDto } from '../organization/dto/organization.dto'
import { UpdateOrganizationMemberRoleDto } from '../organization/dto/update-organization-member-role.dto'
import { UpdateAssignedOrganizationRolesDto } from '../organization/dto/update-assigned-organization-roles.dto'
import { CreateOrganizationRoleDto } from '../organization/dto/create-organization-role.dto'
import { UpdateOrganizationRoleDto } from '../organization/dto/update-organization-role.dto'
import { CreateOrganizationInvitationDto } from '../organization/dto/create-organization-invitation.dto'
import { UpdateOrganizationInvitationDto } from '../organization/dto/update-organization-invitation.dto'
import { CustomHeaders } from '../common/constants/header.constants'
import { BuildImageDto } from '../workspace/dto/build-image.dto'
import { CreateVolumeDto } from '../workspace/dto/create-volume.dto'
import { VolumeDto } from '../workspace/dto/volume.dto'

type RequestWithUser = Request & { user?: { userId: string; organizationId: string } }
type CommonCaptureProps = {
  organizationId?: string
  distinctId: string
  durationMs: number
  statusCode: number
  userAgent: string
  error?: string
  source: string
}

@Injectable()
export class MetricsInterceptor implements NestInterceptor, OnApplicationShutdown {
  private readonly posthog?: PostHog
  private readonly logger = new Logger(MetricsInterceptor.name)

  constructor() {
    if (!process.env.POSTHOG_API_KEY) {
      this.logger.warn('POSTHOG_API_KEY is not set, metrics will not be recorded')
      return
    }

    if (!process.env.POSTHOG_HOST) {
      this.logger.warn('POSTHOG_HOST is not set, metrics will not be recorded')
      return
    }

    // Initialize PostHog client
    // Make sure to set POSTHOG_API_KEY in your environment variables
    this.posthog = new PostHog(process.env.POSTHOG_API_KEY, {
      host: process.env.POSTHOG_HOST,
    })
  }

  intercept(context: ExecutionContext, next: CallHandler): Observable<any> {
    if (!this.posthog) {
      return next.handle()
    }

    const request = context.switchToHttp().getRequest()
    const startTime = Date.now()

    return next.handle().pipe(
      tap({
        next: (response) => {
          // For DELETE requests or empty responses, pass an empty object with statusCode
          const responseObj = response || { statusCode: 204 }
          this.recordMetrics(request, responseObj, startTime).catch((err) => this.logger.error(err))
        },
        error: (error) => {
          this.recordMetrics(
            request,
            { statusCode: error.status || 500 },
            startTime,
            error.message || JSON.stringify(error),
          ).catch((err) => this.logger.error(err))
        },
      }),
    )
  }

  private async recordMetrics(request: RequestWithUser, response: any, startTime: number, error?: string) {
    const durationMs = Date.now() - startTime
    const statusCode = error ? response.statusCode : response.statusCode || (request.method === 'DELETE' ? 204 : 200) // Default to 204 for DELETE requests
    const distinctId = request.user?.userId || 'anonymous'
    const userAgent = request.get('user-agent')
    const source = request.get(CustomHeaders.SOURCE.name)

    const props: CommonCaptureProps = {
      distinctId,
      organizationId: request.user?.organizationId,
      durationMs,
      statusCode,
      userAgent,
      error,
      source: Array.isArray(source) ? source[0] : source,
    }

    switch (request.method) {
      case 'POST':
        switch (request.route.path) {
          case '/api/api-keys':
            this.captureCreateApiKey(props)
            break
          case '/api/images':
            this.captureCreateImage(props, request.body, response)
            break
          case '/api/images/build':
            this.captureBuildImage(props, request.body, response)
            break
          case '/api/docker-registry':
            this.captureCreateDockerRegistry(props, response)
            break
          case '/api/workspace':
            this.captureCreateWorkspace(props, request.body, response)
            break
          case '/api/workspace/:workspaceId/start':
            this.captureStartWorkspace(props, request.params.workspaceId)
            break
          case '/api/workspace/:workspaceId/stop':
            this.captureStopWorkspace(props, request.params.workspaceId)
            break
          case '/api/workspace/:workspaceId/snapshot':
            this.captureCreateSnapshot(props, request.params.workspaceId)
            break
          case '/api/workspace/:workspaceId/public/:isPublic':
            this.captureUpdatePublicStatus(props, request.params.workspaceId, request.params.isPublic === 'true')
            break
          case '/api/workspace/:workspaceId/autostop/:interval':
            this.captureSetAutostopInterval(props, request.params.workspaceId, parseInt(request.params.interval))
            break
          case '/api/organizations/invitations/:invitationId/accept':
            this.captureAcceptInvitation(props, request.params.invitationId)
            break
          case '/api/organizations/invitations/:invitationId/decline':
            this.captureDeclineInvitation(props, request.params.invitationId)
            break
          case '/api/organizations':
            this.captureCreateOrganization(props, request.body, response)
            break
          case '/api/organizations/:organizationId/leave':
            this.captureLeaveOrganization(props, request.params.organizationId)
            break
          case '/api/organizations/:organizationId/users/:userId/role':
            this.captureUpdateOrganizationUserRole(
              props,
              request.params.organizationId,
              request.params.userId,
              request.body,
            )
            break
          case '/api/organizations/:organizationId/users/:userId/assigned-roles':
            this.captureUpdateOrganizationUserAssignedRoles(
              props,
              request.params.organizationId,
              request.params.userId,
              request.body,
            )
            break
          case '/api/organizations/:organizationId/roles':
            this.captureCreateOrganizationRole(props, request.params.organizationId, request.body)
            break
          case '/api/organizations/:organizationId/invitations':
            this.captureCreateOrganizationInvitation(props, request.params.organizationId, request.body)
            break
          case '/api/organizations/:organizationId/invitations/:invitationId/cancel':
            this.captureCancelOrganizationInvitation(props, request.params.organizationId, request.params.invitationId)
            break
          case '/api/volumes':
            this.captureCreateVolume(props, request.body, response)
            break
        }
        break
      case 'DELETE':
        switch (request.route.path) {
          case '/api/workspace/:workspaceId':
            this.captureDeleteWorkspace(props, request.params.workspaceId)
            break
          case '/api/images/:imageId':
            this.captureDeleteImage(props, request.params.imageId)
            break
          case '/api/organizations/:organizationId':
            this.captureDeleteOrganization(props, request.params.organizationId)
            break
          case '/api/organizations/:organizationId/users/:userId':
            this.captureDeleteOrganizationUser(props, request.params.organizationId, request.params.userId)
            break
          case '/api/organizations/:organizationId/roles/:roleId':
            this.captureDeleteOrganizationRole(props, request.params.organizationId, request.params.roleId)
            break
          case '/api/volumes/:volumeId':
            this.captureDeleteVolume(props, request.params.volumeId)
            break
        }
        break
      case 'PUT':
        switch (request.route.path) {
          case '/api/workspace/:workspaceId/labels':
            this.captureUpdateWorkspaceLabels(props, request.params.workspaceId)
            break
          case '/api/organizations/:organizationId/quota':
            this.captureUpdateOrganizationUserQuota(props, request.params.organizationId, request.body)
            break
          case '/api/organizations/:organizationId/roles/:roleId':
            this.captureUpdateOrganizationRole(
              props,
              request.params.organizationId,
              request.params.roleId,
              request.body,
            )
            break
          case '/api/organizations/:organizationId/invitations/:invitationId':
            this.captureUpdateOrganizationInvitation(
              props,
              request.params.organizationId,
              request.params.invitationId,
              request.body,
            )
        }
        break
      case 'PATCH':
        switch (request.route.path) {
          case '/api/images/:imageId/toggle':
            this.captureToggleImageState(props, request.params.imageId, request.body)
            break
        }
        break
    }

    if (!request.route.path.startsWith('/api/toolbox/:workspaceId/toolbox')) {
      return
    }

    const path = request.route.path.replace('/api/toolbox/:workspaceId/toolbox', '')

    switch (path) {
      case '/project-dir':
        this.captureToolboxCommand(props, request.params.workspaceId, 'project-dir_get')
        break
      case '/files':
        switch (request.method) {
          case 'GET':
            this.captureToolboxCommand(props, request.params.workspaceId, 'files_list')
            break
          case 'DELETE':
            this.captureToolboxCommand(props, request.params.workspaceId, 'files_delete')
            break
        }
        break
      case '/files/download':
        this.captureToolboxCommand(props, request.params.workspaceId, 'files_download')
        break
      case '/files/find':
        this.captureToolboxCommand(props, request.params.workspaceId, 'files_find')
        break
      case '/files/folder':
        this.captureToolboxCommand(props, request.params.workspaceId, 'files_folder_create')
        break
      case '/files/info':
        this.captureToolboxCommand(props, request.params.workspaceId, 'files_info')
        break
      case '/files/move':
        this.captureToolboxCommand(props, request.params.workspaceId, 'files_move')
        break
      case '/files/permissions':
        this.captureToolboxCommand(props, request.params.workspaceId, 'files_permissions')
        break
      case '/files/replace':
        this.captureToolboxCommand(props, request.params.workspaceId, 'files_replace')
        break
      case '/files/search':
        this.captureToolboxCommand(props, request.params.workspaceId, 'files_search')
        break
      case '/files/upload':
        this.captureToolboxCommand(props, request.params.workspaceId, 'files_upload')
        break
      case '/git/add':
        this.captureToolboxCommand(props, request.params.workspaceId, 'git_add')
        break
      case '/git/branches':
        switch (request.method) {
          case 'GET':
            this.captureToolboxCommand(props, request.params.workspaceId, 'git_branches_list')
            break
          case 'POST':
            this.captureToolboxCommand(props, request.params.workspaceId, 'git_branches_create')
            break
        }
        break
      case '/git/clone':
        this.captureToolboxCommand(props, request.params.workspaceId, 'git_clone')
        break
      case '/git/commit':
        this.captureToolboxCommand(props, request.params.workspaceId, 'git_commit')
        break
      case '/git/history':
        this.captureToolboxCommand(props, request.params.workspaceId, 'git_history')
        break
      case '/git/pull':
        this.captureToolboxCommand(props, request.params.workspaceId, 'git_pull')
        break
      case '/git/push':
        this.captureToolboxCommand(props, request.params.workspaceId, 'git_push')
        break
      case '/git/status':
        this.captureToolboxCommand(props, request.params.workspaceId, 'git_status')
        break
      case '/process/execute':
        this.captureToolboxCommand(props, request.params.workspaceId, 'process_execute', {
          command: request.body.command,
          cwd: request.body.cwd,
          exit_code: response.exitCode,
          timeout_sec: request.body.timeout,
        })
        break
      case '/process/session':
        switch (request.method) {
          case 'GET':
            this.captureToolboxCommand(props, request.params.workspaceId, 'process_session_list')
            break
          case 'POST':
            this.captureToolboxCommand(props, request.params.workspaceId, 'process_session_create', {
              session_id: request.body.sessionId,
            })
            break
        }
        break
      case '/process/session/:sessionId':
        switch (request.method) {
          case 'GET':
            this.captureToolboxCommand(props, request.params.workspaceId, 'process_session_get', {
              session_id: request.params.sessionId,
            })
            break
          case 'DELETE':
            this.captureToolboxCommand(props, request.params.workspaceId, 'process_session_delete', {
              session_id: request.params.sessionId,
            })
            break
        }
        break
      case '/process/session/:sessionId/exec':
        this.captureToolboxCommand(props, request.params.workspaceId, 'process_session_execute', {
          session_id: request.params.sessionId,
          command: request.body.command,
        })
        break
      case '/process/session/:sessionId/command/:commandId':
        this.captureToolboxCommand(props, request.params.workspaceId, 'process_session_command_get', {
          session_id: request.params.sessionId,
          command_id: request.params.commandId,
        })
        break
      case '/process/session/:sessionId/command/:commandId/logs':
        this.captureToolboxCommand(props, request.params.workspaceId, 'process_session_command_logs', {
          session_id: request.params.sessionId,
          command_id: request.params.commandId,
        })
        break
      case '/lsp/completions':
        this.captureToolboxCommand(props, request.params.workspaceId, 'lsp_completions')
        break
      case '/lsp/did-close':
        this.captureToolboxCommand(props, request.params.workspaceId, 'lsp_did_close')
        break
      case '/lsp/did-open':
        this.captureToolboxCommand(props, request.params.workspaceId, 'lsp_did_open')
        break
      case '/lsp/document-symbols':
        this.captureToolboxCommand(props, request.params.workspaceId, 'lsp_document_symbols')
        break
      case '/lsp/start':
        this.captureToolboxCommand(props, request.params.workspaceId, 'lsp_start', {
          language_id: request.body.languageId,
        })
        break
      case '/lsp/stop':
        this.captureToolboxCommand(props, request.params.workspaceId, 'lsp_stop', {
          language_id: request.body.languageId,
        })
        break
      case '/lsp/workspace-symbols':
        this.captureToolboxCommand(props, request.params.workspaceId, 'lsp_workspace_symbols', {
          language_id: request.query.languageId,
          path_to_project: request.query.pathToProject,
          query: request.query.query,
        })
        break
    }
  }

  private captureCreateApiKey(props: CommonCaptureProps) {
    this.capture('api_api_key_created', props, 'api_api_key_creation_failed')
  }

  private captureCreateDockerRegistry(props: CommonCaptureProps, response: DockerRegistryDto) {
    this.capture('api_docker_registry_created', props, 'api_docker_registry_creation_failed', {
      registry_name: response.name,
      registry_url: response.url,
    })
  }

  private captureCreateImage(props: CommonCaptureProps, request: CreateImageDto, response: ImageDto) {
    this.capture('api_image_created', props, 'api_image_creation_failed', {
      image_id: response.id,
      image_name: request.name,
      image_entrypoint: request.entrypoint,
    })
  }

  private captureBuildImage(props: CommonCaptureProps, request: BuildImageDto, response: ImageDto) {
    this.capture('api_image_built', props, 'api_image_build_failed', {
      image_id: response.id,
      image_name: request.name,
      image_build_info_context_hashes_length: request.buildInfo.contextHashes?.length,
    })
  }

  private captureDeleteImage(props: CommonCaptureProps, imageId: string) {
    this.capture('api_image_deleted', props, 'api_image_deletion_failed', {
      image_id: imageId,
    })
  }

  private captureToggleImageState(props: CommonCaptureProps, imageId: string, request: ToggleStateDto) {
    this.capture('api_image_state_toggled', props, 'api_image_state_toggle_failed', {
      image_id: imageId,
      image_enabled: request.enabled,
    })
  }

  private captureCreateWorkspace(props: CommonCaptureProps, request: CreateWorkspaceDto, response: WorkspaceDto) {
    const envVarsLength = request.env ? Object.keys(request.env).length : 0

    const records = {
      sandbox_id: response.id,
      sandbox_image_request: request.image,
      sandbox_image: response.image,
      sandbox_user_request: request.user,
      sandbox_user: response.user,
      sandbox_cpu_request: request.cpu,
      sandbox_cpu: response.cpu,
      sandbox_gpu_request: request.gpu,
      sandbox_gpu: response.gpu,
      sandbox_memory_mb_request: request.memory * 1024,
      sandbox_memory_mb: response.memory * 1024,
      sandbox_disk_gb_request: request.disk,
      sandbox_disk_gb: response.disk,
      sandbox_target_request: request.target,
      sandbox_target: response.target,
      sandbox_auto_stop_interval_min_request: request.autoStopInterval,
      sandbox_auto_stop_interval_min: response.autoStopInterval,
      sandbox_public_request: request.public,
      sandbox_public: response.public,
      sandbox_labels_request: request.labels,
      sandbox_labels: response.labels,
      sandbox_env_vars_length_request: envVarsLength,
      sandbox_volumes_length_request: request.volumes?.length,
    }

    if (request.buildInfo) {
      records['sandbox_is_dynamic_build'] = true
      records['sandbox_build_info_context_hashes_length'] = request.buildInfo.contextHashes?.length
    }

    this.capture('api_sandbox_created', props, 'api_sandbox_creation_failed', records)
  }

  private captureDeleteWorkspace(props: CommonCaptureProps, sandboxId: string) {
    this.capture('api_sandbox_deleted', props, 'api_sandbox_deletion_failed', {
      sandbox_id: sandboxId,
    })
  }

  private captureStartWorkspace(props: CommonCaptureProps, sandboxId: string) {
    this.capture('api_sandbox_started', props, 'api_sandbox_start_failed', {
      sandbox_id: sandboxId,
    })
  }

  private captureStopWorkspace(props: CommonCaptureProps, sandboxId: string) {
    this.capture('api_sandbox_stopped', props, 'api_sandbox_stop_failed', {
      sandbox_id: sandboxId,
    })
  }

  private captureCreateSnapshot(props: CommonCaptureProps, sandboxId: string) {
    this.capture('api_sandbox_snapshot_created', props, 'api_sandbox_snapshot_creation_failed', {
      sandbox_id: sandboxId,
    })
  }

  private captureUpdatePublicStatus(props: CommonCaptureProps, sandboxId: string, isPublic: boolean) {
    this.capture('api_sandbox_public_status_updated', props, 'api_sandbox_public_status_update_failed', {
      sandbox_id: sandboxId,
      sandbox_public: isPublic,
    })
  }

  private captureSetAutostopInterval(props: CommonCaptureProps, sandboxId: string, interval: number) {
    this.capture('api_sandbox_autostop_interval_updated', props, 'api_sandbox_autostop_interval_update_failed', {
      sandbox_id: sandboxId,
      sandbox_autostop_interval: interval,
    })
  }

  private captureUpdateWorkspaceLabels(props: CommonCaptureProps, sandboxId: string) {
    this.capture('api_sandbox_labels_update', props, 'api_sandbox_labels_update_failed', {
      sandbox_id: sandboxId,
    })
  }

  private captureAcceptInvitation(props: CommonCaptureProps, invitationId: string) {
    this.capture('api_invitation_accepted', props, 'api_invitation_accept_failed', {
      invitation_id: invitationId,
    })
  }

  private captureDeclineInvitation(props: CommonCaptureProps, invitationId: string) {
    this.capture('api_invitation_declined', props, 'api_invitation_decline_failed', {
      invitation_id: invitationId,
    })
  }

  private captureCreateOrganization(
    props: CommonCaptureProps,
    request: CreateOrganizationDto,
    response: OrganizationDto,
  ) {
    if (!this.posthog) {
      return
    }

    this.posthog.groupIdentify({
      groupType: 'organization',
      groupKey: response.id,
      properties: {
        name: request.name,
        created_at: response.createdAt,
        created_by: response.createdBy,
        personal: response.personal,
      },
    })

    this.capture('api_organization_created', props, 'api_organization_creation_failed', {
      organization_id: response.id,
      organization_name: request.name,
    })
  }

  private captureLeaveOrganization(props: CommonCaptureProps, organizationId: string) {
    this.capture('api_organization_left', props, 'api_organization_leave_failed', {
      organization_id: organizationId,
    })
  }

  private captureUpdateOrganizationUserQuota(
    props: CommonCaptureProps,
    organizationId: string,
    request: UpdateOrganizationQuotaDto,
  ) {
    this.capture('api_organization_user_quota_updated', props, 'api_organization_user_quota_update_failed', {
      organization_id: organizationId,
      organization_user_image_quota: request.imageQuota,
      organization_user_total_cpu_quota: request.totalCpuQuota,
      organization_user_total_memory_quota_mb: request.totalMemoryQuota * 1024,
      organization_user_total_disk_quota_gb: request.totalDiskQuota,
      organization_user_max_concurrent_workspaces: request.maxConcurrentWorkspaces,
      organization_user_max_cpu_per_workspace: request.maxCpuPerWorkspace,
      organization_user_max_memory_per_workspace_mb: request.maxMemoryPerWorkspace * 1024,
      organization_user_max_disk_per_workspace_gb: request.maxDiskPerWorkspace,
      organization_user_max_image_size_mb: request.maxImageSize * 1024,
    })
  }

  private captureDeleteOrganization(props: CommonCaptureProps, organizationId: string) {
    this.capture('api_organization_deleted', props, 'api_organization_deletion_failed', {
      organization_id: organizationId,
    })
  }

  private captureUpdateOrganizationUserRole(
    props: CommonCaptureProps,
    organizationId: string,
    userId: string,
    request: UpdateOrganizationMemberRoleDto,
  ) {
    this.capture('api_organization_user_role_updated', props, 'api_organization_user_role_update_failed', {
      organization_id: organizationId,
      organization_user_id: userId,
      organization_user_role: request.role,
    })
  }

  private captureUpdateOrganizationUserAssignedRoles(
    props: CommonCaptureProps,
    organizationId: string,
    userId: string,
    request: UpdateAssignedOrganizationRolesDto,
  ) {
    this.capture(
      'api_organization_user_assigned_roles_updated',
      props,
      'api_organization_user_assigned_roles_update_failed',
      {
        organization_id: organizationId,
        organization_user_id: userId,
        organization_user_assigned_roles: request.roleIds,
      },
    )
  }

  private captureDeleteOrganizationUser(props: CommonCaptureProps, organizationId: string, userId: string) {
    this.capture('api_organization_user_deleted', props, 'api_organization_user_deletion_failed', {
      organization_id: organizationId,
      organization_user_id: userId,
    })
  }

  private captureCreateOrganizationRole(
    props: CommonCaptureProps,
    organizationId: string,
    request: CreateOrganizationRoleDto,
  ) {
    this.capture('api_organization_role_created', props, 'api_organization_role_creation_failed', {
      organization_id: organizationId,
      organization_role_name: request.name,
      organization_role_description: request.description,
      organization_role_permissions: request.permissions,
    })
  }

  private captureDeleteOrganizationRole(props: CommonCaptureProps, organizationId: string, roleId: string) {
    this.capture('api_organization_role_deleted', props, 'api_organization_role_deletion_failed', {
      organization_id: organizationId,
      organization_role_id: roleId,
    })
  }

  private captureUpdateOrganizationRole(
    props: CommonCaptureProps,
    organizationId: string,
    roleId: string,
    request: UpdateOrganizationRoleDto,
  ) {
    this.capture('api_organization_role_updated', props, 'api_organization_role_update_failed', {
      organization_id: organizationId,
      organization_role_id: roleId,
      organization_role_name: request.name,
      organization_role_description: request.description,
      organization_role_permissions: request.permissions,
    })
  }

  private captureCreateOrganizationInvitation(
    props: CommonCaptureProps,
    organizationId: string,
    request: CreateOrganizationInvitationDto,
  ) {
    this.capture('api_organization_invitation_created', props, 'api_organization_invitation_creation_failed', {
      organization_id: organizationId,
      organization_invitation_email: request.email,
      organization_invitation_role: request.role,
      organization_invitation_assigned_role_ids: request.assignedRoleIds,
      organization_invitation_expires_at: request.expiresAt,
    })
  }

  private captureUpdateOrganizationInvitation(
    props: CommonCaptureProps,
    organizationId: string,
    invitationId: string,
    request: UpdateOrganizationInvitationDto,
  ) {
    this.capture('api_organization_invitation_updated', props, 'api_organization_invitation_update_failed', {
      organization_id: organizationId,
      organization_invitation_id: invitationId,
      organization_invitation_expires_at: request.expiresAt,
      organization_invitation_role: request.role,
      organization_invitation_assigned_role_ids: request.assignedRoleIds,
    })
  }

  private captureCancelOrganizationInvitation(props: CommonCaptureProps, organizationId: string, invitationId: string) {
    this.capture('api_organization_invitation_canceled', props, 'api_organization_invitation_cancel_failed', {
      organization_id: organizationId,
      organization_invitation_id: invitationId,
    })
  }

  private captureCreateVolume(props: CommonCaptureProps, request: CreateVolumeDto, response: VolumeDto) {
    this.capture('api_volume_created', props, 'api_volume_creation_failed', {
      volume_id: response.id,
      volume_name_request_set: !!request.name,
    })
  }

  private captureDeleteVolume(props: CommonCaptureProps, volumeId: string) {
    this.capture('api_volume_deleted', props, 'api_volume_deletion_failed', {
      volume_id: volumeId,
    })
  }

  private captureToolboxCommand(
    props: CommonCaptureProps,
    sandboxId: string,
    command: string,
    extraProps?: Record<string, any>,
  ) {
    this.capture('api_toolbox_command', props, 'api_toolbox_command_failed', {
      sandbox_id: sandboxId,
      toolbox_command: command,
      ...extraProps,
    })
  }

  private capture(event: string, props: CommonCaptureProps, errorEvent?: string, extraProps?: Record<string, any>) {
    if (!this.posthog) {
      return
    }

    this.posthog.capture({
      distinctId: props.distinctId,
      event: props.error ? errorEvent || event : event,
      groups: this.captureCommonGroups(props),
      properties: { ...this.captureCommonProperties(props), ...extraProps },
    })
  }

  private captureCommonProperties(props: CommonCaptureProps) {
    return {
      duration_ms: props.durationMs,
      status_code: props.statusCode,
      user_agent: props.userAgent,
      error: props.error,
      source: props.source,
    }
  }

  private captureCommonGroups(props: CommonCaptureProps) {
    return props.organizationId ? { organization: props.organizationId } : undefined
  }

  onApplicationShutdown(/*signal?: string*/) {
    if (this.posthog) {
      this.posthog.shutdown()
    }
  }
}
