/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import {
  Injectable,
  NestInterceptor,
  ExecutionContext,
  CallHandler,
  Logger,
  InternalServerErrorException,
  HttpException,
  HttpStatus,
} from '@nestjs/common'
import { Reflector } from '@nestjs/core'
import { Request, Response } from 'express'
import { Observable, Subscriber, firstValueFrom } from 'rxjs'
import { AUDIT_CONTEXT_KEY, AuditContext } from '../decorators/audit.decorator'
import { AuditLog, AuditLogMetadata } from '../entities/audit-log.entity'
import { AuditAction } from '../enums/audit-action.enum'
import { AuditService } from '../services/audit.service'
import { BaseAuthContext, isBaseAuthContext } from '../../common/interfaces/base-auth-context.interface'
import { isUserAuthContext } from '../../common/interfaces/user-auth-context.interface'
import { isOrganizationAuthContext } from '../../common/interfaces/organization-auth-context.interface'
import { getAuthContext } from '../../common/utils/get-auth-context'
import { CustomHeaders } from '../../common/constants/header.constants'
import { TypedConfigService } from '../../config/typed-config.service'
import { truncateErrorMessage } from '../../common/utils/truncate-error-message'

@Injectable()
export class AuditInterceptor implements NestInterceptor {
  private readonly logger = new Logger(AuditInterceptor.name)

  constructor(
    private readonly reflector: Reflector,
    private readonly auditService: AuditService,
    private readonly configService: TypedConfigService,
  ) {}

  intercept(context: ExecutionContext, next: CallHandler): Observable<any> {
    const request = context.switchToHttp().getRequest<Request>()
    const response = context.switchToHttp().getResponse<Response>()

    const auditContext = this.reflector.get<AuditContext>(AUDIT_CONTEXT_KEY, context.getHandler())

    // Non-audited request
    if (!auditContext) {
      return next.handle()
    }

    // Toolbox requests are not audited by default
    if (this.isToolboxAction(auditContext.action) && !this.configService.get('audit.toolboxRequestsEnabled')) {
      return next.handle()
    }

    const authContext = getAuthContext(context, isBaseAuthContext)

    return new Observable((observer) => {
      this.handleAuditedRequest(auditContext, authContext, request, response, next, observer)
    })
  }

  // An audit log must be created before the request is passed to the request handler
  // After the request handler returns, the audit log is optimistically updated with the outcome
  private async handleAuditedRequest(
    auditContext: AuditContext,
    authContext: BaseAuthContext,
    request: Request,
    response: Response,
    next: CallHandler,
    observer: Subscriber<any>,
  ): Promise<void> {
    try {
      const actorId = isUserAuthContext(authContext) ? authContext.userId : authContext.role
      const actorEmail = isUserAuthContext(authContext) ? authContext.email : undefined
      const actorApiKeyPrefix = isUserAuthContext(authContext) ? authContext.apiKey?.keyPrefix : undefined
      const actorApiKeySuffix = isUserAuthContext(authContext) ? authContext.apiKey?.keySuffix : undefined
      const organizationId = isOrganizationAuthContext(authContext) ? authContext.organizationId : undefined

      const auditLog = await this.auditService.createLog({
        actorId,
        actorEmail,
        actorApiKeyPrefix,
        actorApiKeySuffix,
        organizationId,
        action: auditContext.action,
        targetType: auditContext.targetType,
        targetId: this.resolveTargetId(auditContext, request),
        ipAddress: request.ip,
        userAgent: request.get('user-agent'),
        source: request.get(CustomHeaders.SOURCE.name),
        metadata: this.resolveRequestMetadata(auditContext, request),
      })

      try {
        const result = await firstValueFrom(next.handle())

        const resolvedOrganizationId = this.resolveOrganizationId(organizationId, result)
        const resolvedTargetId = this.resolveTargetId(auditContext, request, result)
        const statusCode = response.statusCode || HttpStatus.NO_CONTENT
        await this.recordHandlerSuccess(auditLog, resolvedOrganizationId, resolvedTargetId, statusCode)

        observer.next(result)
        observer.complete()
      } catch (handlerError) {
        const errorMessage =
          handlerError instanceof HttpException
            ? truncateErrorMessage(handlerError.message)
            : 'An unexpected error occurred.'
        const statusCode = this.resolveErrorStatusCode(handlerError)
        await this.recordHandlerError(auditLog, errorMessage, statusCode)

        observer.error(handlerError)
      }
    } catch (createLogError) {
      this.logger.error('Failed to create audit log:', createLogError)
      observer.error(new InternalServerErrorException())
    }
  }

  private resolveOrganizationId(organizationId: string | undefined, result?: any): string | null {
    return result?.organizationId || organizationId || null
  }

  /**
   * Resolves the identifier of the target resource from the initial request or the response object.
   *
   * Prioritizes resolving the ID from the response object as the request may not include a unique resource identifier (e.g. delete sandbox by name).
   */
  private resolveTargetId(auditContext: AuditContext, request: Request, result?: any): string | null {
    if (auditContext.targetIdFromResult && result) {
      const targetId = auditContext.targetIdFromResult(result)
      if (targetId) {
        return targetId
      }
    }

    if (auditContext.targetIdFromRequest) {
      const targetId = auditContext.targetIdFromRequest(request)
      if (targetId) {
        return targetId
      }
    }

    return null
  }

  private resolveRequestMetadata(auditContext: AuditContext, request: Request): AuditLogMetadata | null {
    if (!auditContext.requestMetadata) {
      return null
    }

    const resolvedMetadata: AuditLogMetadata = {}

    for (const [key, resolver] of Object.entries(auditContext.requestMetadata)) {
      try {
        resolvedMetadata[key] = resolver(request)
      } catch (error) {
        this.logger.warn(`Failed to resolve audit log metadata key "${key}":`, error)
        resolvedMetadata[key] = null
      }
    }

    return Object.keys(resolvedMetadata).length > 0 ? resolvedMetadata : null
  }

  private isToolboxAction(action: AuditAction): boolean {
    return action.startsWith('toolbox_')
  }

  private async recordHandlerSuccess(
    auditLog: AuditLog,
    organizationId: string | null,
    targetId: string | null,
    statusCode: number,
  ): Promise<void> {
    try {
      await this.auditService.updateLog(auditLog.id, {
        organizationId,
        targetId,
        statusCode,
      })
    } catch (error) {
      this.logger.error('Failed to record handler result:', error)
    }
  }

  private async recordHandlerError(auditLog: AuditLog, errorMessage: string, statusCode: number): Promise<void> {
    try {
      await this.auditService.updateLog(auditLog.id, {
        errorMessage,
        statusCode,
      })
    } catch (error) {
      this.logger.error('Failed to record handler error:', error)
    }
  }

  private resolveErrorStatusCode(error: any): number {
    if (error instanceof HttpException) {
      return error.getStatus()
    }

    return HttpStatus.INTERNAL_SERVER_ERROR
  }
}
