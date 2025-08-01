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
import { SandboxDto } from '../sandbox/dto/sandbox.dto'
import { DockerRegistryDto } from '../docker-registry/dto/docker-registry.dto'
import { CreateSandboxDto } from '../sandbox/dto/create-sandbox.dto'
import { Request } from 'express'
import { CreateSnapshotDto } from '../sandbox/dto/create-snapshot.dto'
import { SnapshotDto } from '../sandbox/dto/snapshot.dto'
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
import { CreateVolumeDto } from '../sandbox/dto/create-volume.dto'
import { VolumeDto } from '../sandbox/dto/volume.dto'
import { CreateWorkspaceDto } from '../sandbox/dto/create-workspace.deprecated.dto'
import { WorkspaceDto } from '../sandbox/dto/workspace.deprecated.dto'
import { TypedConfigService } from '../config/typed-config.service'

type RequestWithUser = Request & { user?: { userId: string; organizationId: string } }
type CommonCaptureProps = {
  organizationId?: string
  distinctId: string
  durationMs: number
  statusCode: number
  userAgent: string
  error?: string
  source: string
  isDeprecated?: boolean
  sdkVersion?: string
  environment?: string
}

@Injectable()
export class MetricsInterceptor implements NestInterceptor, OnApplicationShutdown {
  private readonly posthog?: PostHog
  private readonly logger = new Logger(MetricsInterceptor.name)

  constructor(private readonly configService: TypedConfigService) {
    if (!this.configService.get('posthog.apiKey')) {
      this.logger.warn('POSTHOG_API_KEY is not set, metrics will not be recorded')
      return
    }

    if (!this.configService.get('posthog.host')) {
      this.logger.warn('POSTHOG_HOST is not set, metrics will not be recorded')
      return
    }

    // Initialize PostHog client
    // Make sure to set POSTHOG_API_KEY in your environment variables
    this.posthog = new PostHog(this.configService.getOrThrow('posthog.apiKey'), {
      host: this.configService.getOrThrow('posthog.host'),
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
    const sdkVersion = request.get(CustomHeaders.SDK_VERSION.name)

    const props: CommonCaptureProps = {
      distinctId,
      organizationId: request.user?.organizationId,
      durationMs,
      statusCode,
      userAgent,
      error,
      source: Array.isArray(source) ? source[0] : source,
      isDeprecated: request.route.path.includes('/workspace') || request.route.path.includes('/images'),
      sdkVersion,
      environment: this.configService.get('posthog.environment'),
    }

    switch (request.method) {
      case 'POST':
        switch (request.route.path) {
          case '/api/api-keys':
            this.captureCreateApiKey(props)
            break
          case '/api/snapshots':
            this.captureCreateSnapshot(props, request.body, response)
            break
          case '/api/snapshots/:snapshotId/activate':
            this.captureActivateSnapshot(props, request.params.snapshotId)
            break
          case '/api/snapshots/:snapshotId/deactivate':
            this.captureDeactivateSnapshot(props, request.params.snapshotId)
            break
          case '/api/docker-registry':
            this.captureCreateDockerRegistry(props, response)
            break
          case '/api/sandbox':
            this.captureCreateSandbox(props, request.body, response)
            break
          case '/api/workspace':
            this.captureCreateWorkspace_deprecated(props, request.body, response)
            break
          case '/api/sandbox/:sandboxId/start':
          case '/api/workspace/:workspaceId/start':
            this.captureStartSandbox(props, request.params.sandboxId)
            break
          case '/api/sandbox/:sandboxId/stop':
          case '/api/workspace/:workspaceId/stop':
            this.captureStopSandbox(props, request.params.sandboxId)
            break
          case '/api/sandbox/:sandboxId/backup':
            this.captureCreateBackup(props, request.params.sandboxId)
            break
          case '/api/sandbox/:sandboxId/public/:isPublic':
          case '/api/workspace/:workspaceId/public/:isPublic':
            this.captureUpdatePublicStatus(props, request.params.sandboxId, request.params.isPublic === 'true')
            break
          case '/api/sandbox/:sandboxId/autostop/:interval':
          case '/api/workspace/:workspaceId/autostop/:interval':
            this.captureSetAutostopInterval(props, request.params.sandboxId, parseInt(request.params.interval))
            break
          case '/api/sandbox/:sandboxId/autoarchive/:interval':
          case '/api/workspace/:workspaceId/autoarchive/:interval':
            this.captureSetAutoArchiveInterval(props, request.params.sandboxId, parseInt(request.params.interval))
            break
          case '/api/sandbox/:sandboxId/autodelete/:interval':
            this.captureSetAutoDeleteInterval(props, request.params.sandboxId, parseInt(request.params.interval))
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
          case '/api/sandbox/:sandboxId':
          case '/api/workspace/:workspaceId':
            this.captureDeleteSandbox(props, request.params.sandboxId)
            break
          case '/api/snapshots/:snapshotId':
            this.captureDeleteSnapshot(props, request.params.snapshotId)
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
          case '/api/sandbox/:sandboxId/labels':
          case '/api/workspace/:workspaceId/labels':
            this.captureUpdateSandboxLabels(props, request.params.sandboxId)
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
          case '/api/organizations/:organizationId/quota':
            this.captureUpdateOrganizationQuota(props, request.params.organizationId, request.body)
            break
        }
        break
    }

    if (!request.route.path.startsWith('/api/toolbox/:sandboxId/toolbox')) {
      return
    }

    const path = request.route.path.replace('/api/toolbox/:sandboxId/toolbox', '')

    switch (path) {
      case '/project-dir':
        this.captureToolboxCommand(props, request.params.sandboxId, 'project-dir_get')
        break
      case '/files':
        switch (request.method) {
          case 'GET':
            this.captureToolboxCommand(props, request.params.sandboxId, 'files_list')
            break
          case 'DELETE':
            this.captureToolboxCommand(props, request.params.sandboxId, 'files_delete')
            break
        }
        break
      case '/files/download':
        this.captureToolboxCommand(props, request.params.sandboxId, 'files_download')
        break
      case '/files/find':
        this.captureToolboxCommand(props, request.params.sandboxId, 'files_find')
        break
      case '/files/folder':
        this.captureToolboxCommand(props, request.params.sandboxId, 'files_folder_create')
        break
      case '/files/info':
        this.captureToolboxCommand(props, request.params.sandboxId, 'files_info')
        break
      case '/files/move':
        this.captureToolboxCommand(props, request.params.sandboxId, 'files_move')
        break
      case '/files/permissions':
        this.captureToolboxCommand(props, request.params.sandboxId, 'files_permissions')
        break
      case '/files/replace':
        this.captureToolboxCommand(props, request.params.sandboxId, 'files_replace')
        break
      case '/files/search':
        this.captureToolboxCommand(props, request.params.sandboxId, 'files_search')
        break
      case '/files/upload':
        this.captureToolboxCommand(props, request.params.sandboxId, 'files_upload')
        break
      case '/git/add':
        this.captureToolboxCommand(props, request.params.sandboxId, 'git_add')
        break
      case '/git/branches':
        switch (request.method) {
          case 'GET':
            this.captureToolboxCommand(props, request.params.sandboxId, 'git_branches_list')
            break
          case 'POST':
            this.captureToolboxCommand(props, request.params.sandboxId, 'git_branches_create')
            break
        }
        break
      case '/git/clone':
        this.captureToolboxCommand(props, request.params.sandboxId, 'git_clone')
        break
      case '/git/commit':
        this.captureToolboxCommand(props, request.params.sandboxId, 'git_commit')
        break
      case '/git/history':
        this.captureToolboxCommand(props, request.params.sandboxId, 'git_history')
        break
      case '/git/pull':
        this.captureToolboxCommand(props, request.params.sandboxId, 'git_pull')
        break
      case '/git/push':
        this.captureToolboxCommand(props, request.params.sandboxId, 'git_push')
        break
      case '/git/status':
        this.captureToolboxCommand(props, request.params.sandboxId, 'git_status')
        break
      case '/process/execute':
        this.captureToolboxCommand(props, request.params.sandboxId, 'process_execute', {
          command: request.body.command,
          cwd: request.body.cwd,
          exit_code: response.exitCode,
          timeout_sec: request.body.timeout,
        })
        break
      case '/process/session':
        switch (request.method) {
          case 'GET':
            this.captureToolboxCommand(props, request.params.sandboxId, 'process_session_list')
            break
          case 'POST':
            this.captureToolboxCommand(props, request.params.sandboxId, 'process_session_create', {
              session_id: request.body.sessionId,
            })
            break
        }
        break
      case '/process/session/:sessionId':
        switch (request.method) {
          case 'GET':
            this.captureToolboxCommand(props, request.params.sandboxId, 'process_session_get', {
              session_id: request.params.sessionId,
            })
            break
          case 'DELETE':
            this.captureToolboxCommand(props, request.params.sandboxId, 'process_session_delete', {
              session_id: request.params.sessionId,
            })
            break
        }
        break
      case '/process/session/:sessionId/exec':
        this.captureToolboxCommand(props, request.params.sandboxId, 'process_session_execute', {
          session_id: request.params.sessionId,
          command: request.body.command,
        })
        break
      case '/process/session/:sessionId/command/:commandId':
        this.captureToolboxCommand(props, request.params.sandboxId, 'process_session_command_get', {
          session_id: request.params.sessionId,
          command_id: request.params.commandId,
        })
        break
      case '/process/session/:sessionId/command/:commandId/logs':
        this.captureToolboxCommand(props, request.params.sandboxId, 'process_session_command_logs', {
          session_id: request.params.sessionId,
          command_id: request.params.commandId,
        })
        break
      case '/lsp/completions':
        this.captureToolboxCommand(props, request.params.sandboxId, 'lsp_completions')
        break
      case '/lsp/did-close':
        this.captureToolboxCommand(props, request.params.sandboxId, 'lsp_did_close')
        break
      case '/lsp/did-open':
        this.captureToolboxCommand(props, request.params.sandboxId, 'lsp_did_open')
        break
      case '/lsp/document-symbols':
        this.captureToolboxCommand(props, request.params.sandboxId, 'lsp_document_symbols')
        break
      case '/lsp/start':
        this.captureToolboxCommand(props, request.params.sandboxId, 'lsp_start', {
          language_id: request.body.languageId,
        })
        break
      case '/lsp/stop':
        this.captureToolboxCommand(props, request.params.sandboxId, 'lsp_stop', {
          language_id: request.body.languageId,
        })
        break
      case '/lsp/sandbox-symbols':
        this.captureToolboxCommand(props, request.params.sandboxId, 'lsp_sandbox_symbols', {
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

  private captureCreateSnapshot(props: CommonCaptureProps, request: CreateSnapshotDto, response: SnapshotDto) {
    this.capture('api_snapshot_created', props, 'api_snapshot_creation_failed', {
      snapshot_id: response.id,
      snapshot_name: request.name,
      snapshot_image_name: request.imageName,
      snapshot_entrypoint: request.entrypoint,
      snapshot_cpu: request.cpu,
      snapshot_gpu: request.gpu,
      snapshot_memory: request.memory,
      snapshot_disk: request.disk,
      snapshot_is_build: request.buildInfo ? true : false,
      snapshot_build_info_context_hashes_length: request.buildInfo?.contextHashes?.length,
    })
  }

  private captureActivateSnapshot(props: CommonCaptureProps, snapshotId: string) {
    this.capture('api_snapshot_activated', props, 'api_snapshot_activation_failed', {
      snapshot_id: snapshotId,
    })
  }

  private captureDeactivateSnapshot(props: CommonCaptureProps, snapshotId: string) {
    this.capture('api_snapshot_deactivated', props, 'api_snapshot_deactivation_failed', {
      snapshot_id: snapshotId,
    })
  }

  private captureDeleteSnapshot(props: CommonCaptureProps, snapshotId: string) {
    this.capture('api_snapshot_deleted', props, 'api_snapshot_deletion_failed', {
      snapshot_id: snapshotId,
    })
  }

  private captureCreateSandbox(props: CommonCaptureProps, request: CreateSandboxDto, response: SandboxDto) {
    const envVarsLength = request.env ? Object.keys(request.env).length : 0

    const records = {
      sandbox_id: response.id,
      sandbox_snapshot_request: request.snapshot,
      sandbox_snapshot: response.snapshot,
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
      sandbox_auto_archive_interval_min_request: request.autoArchiveInterval,
      sandbox_auto_archive_interval_min: response.autoArchiveInterval,
      sandbox_auto_delete_interval_min_request: request.autoDeleteInterval,
      sandbox_auto_delete_interval_min: response.autoDeleteInterval,
      sandbox_public_request: request.public,
      sandbox_public: response.public,
      sandbox_labels_request: request.labels,
      sandbox_labels: response.labels,
      sandbox_env_vars_length_request: envVarsLength,
      sandbox_volumes_length_request: request.volumes?.length,
      sandbox_daemon_version: response.daemonVersion,
    }

    if (request.buildInfo) {
      records['sandbox_is_dynamic_build'] = true
      records['sandbox_build_info_context_hashes_length'] = request.buildInfo.contextHashes?.length
    }

    this.capture('api_sandbox_created', props, 'api_sandbox_creation_failed', records)
  }

  private captureCreateWorkspace_deprecated(
    props: CommonCaptureProps,
    request: CreateWorkspaceDto,
    response: WorkspaceDto,
  ) {
    const envVarsLength = request.env ? Object.keys(request.env).length : 0

    const records = {
      sandbox_id: response.id,
      sandbox_snapshot_request: request.image,
      sandbox_snapshot: response.snapshot,
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
      sandbox_auto_archive_interval_min_request: request.autoArchiveInterval,
      sandbox_auto_archive_interval_min: response.autoArchiveInterval,
      sandbox_public_request: request.public,
      sandbox_public: response.public,
      sandbox_labels_request: request.labels,
      sandbox_labels: response.labels,
      sandbox_env_vars_length_request: envVarsLength,
      sandbox_volumes_length_request: request.volumes?.length,
      sandbox_daemon_version: response.daemonVersion,
    }

    if (request.buildInfo) {
      records['sandbox_is_dynamic_build'] = true
      records['sandbox_build_info_context_hashes_length'] = request.buildInfo.contextHashes?.length
    }

    this.capture('api_sandbox_created', props, 'api_sandbox_creation_failed', records)
  }

  private captureDeleteSandbox(props: CommonCaptureProps, sandboxId: string) {
    this.capture('api_sandbox_deleted', props, 'api_sandbox_deletion_failed', {
      sandbox_id: sandboxId,
    })
  }

  private captureStartSandbox(props: CommonCaptureProps, sandboxId: string) {
    this.capture('api_sandbox_started', props, 'api_sandbox_start_failed', {
      sandbox_id: sandboxId,
    })
  }

  private captureStopSandbox(props: CommonCaptureProps, sandboxId: string) {
    this.capture('api_sandbox_stopped', props, 'api_sandbox_stop_failed', {
      sandbox_id: sandboxId,
    })
  }

  private captureCreateBackup(props: CommonCaptureProps, sandboxId: string) {
    this.capture('api_sandbox_backup_created', props, 'api_sandbox_backup_creation_failed', {
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

  private captureSetAutoArchiveInterval(props: CommonCaptureProps, sandboxId: string, interval: number) {
    this.capture('api_sandbox_autoarchive_interval_updated', props, 'api_sandbox_autoarchive_interval_update_failed', {
      sandbox_id: sandboxId,
      sandbox_autoarchive_interval: interval,
    })
  }

  private captureSetAutoDeleteInterval(props: CommonCaptureProps, sandboxId: string, interval: number) {
    this.capture('api_sandbox_autodelete_interval_updated', props, 'api_sandbox_autodelete_interval_update_failed', {
      sandbox_id: sandboxId,
      sandbox_autodelete_interval: interval,
    })
  }

  private captureUpdateSandboxLabels(props: CommonCaptureProps, sandboxId: string) {
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

  private captureUpdateOrganizationQuota(
    props: CommonCaptureProps,
    organizationId: string,
    request: UpdateOrganizationQuotaDto,
  ) {
    this.capture('api_organization_quota_updated', props, 'api_organization_quota_update_failed', {
      organization_id: organizationId,
      organization_total_cpu_quota: request.totalCpuQuota,
      organization_total_memory_quota_mb: request.totalMemoryQuota ? request.totalMemoryQuota * 1024 : null,
      organization_total_disk_quota_gb: request.totalDiskQuota,
      organization_max_cpu_per_sandbox: request.maxCpuPerSandbox,
      organization_max_memory_per_sandbox_mb: request.maxMemoryPerSandbox ? request.maxMemoryPerSandbox * 1024 : null,
      organization_max_disk_per_sandbox_gb: request.maxDiskPerSandbox,
      organization_snapshot_quota: request.snapshotQuota,
      organization_max_snapshot_size_mb: request.maxSnapshotSize ? request.maxSnapshotSize * 1024 : null,
      organization_volume_quota: request.volumeQuota,
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
      is_deprecated: props.isDeprecated,
      sdk_version: props.sdkVersion,
      environment: props.environment,
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
