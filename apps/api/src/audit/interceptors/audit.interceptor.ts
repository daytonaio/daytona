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
  UnauthorizedException,
  InternalServerErrorException,
} from '@nestjs/common'
import { Reflector } from '@nestjs/core'
import { Request } from 'express'
import { Observable, Subscriber, firstValueFrom } from 'rxjs'
import { AuditLog, AuditLogMetadata } from '../entities/audit-log.entity'
import { AUDIT_CONTEXT_KEY, AuditContext } from '../decorators/audit.decorator'
import { AuthContext } from '../../common/interfaces/auth-context.interface'
import { AuditService } from '../services/audit.service'
import { CustomHeaders } from '../../common/constants/header.constants'
import { AuditOutcome } from '../enums/audit-outcome-enum'

type RequestWithUser = Request & {
  user?: AuthContext
}

@Injectable()
export class AuditInterceptor implements NestInterceptor {
  private readonly logger = new Logger(AuditInterceptor.name)

  constructor(
    private readonly reflector: Reflector,
    private readonly auditService: AuditService,
  ) {}

  intercept(context: ExecutionContext, next: CallHandler): Observable<any> {
    const request = context.switchToHttp().getRequest<RequestWithUser>()

    const auditContext = this.reflector.get<AuditContext>(AUDIT_CONTEXT_KEY, context.getHandler())

    if (!auditContext) {
      this.logger.warn('Non-audited request:', request.url)
      return next.handle()
    }

    if (!request.user) {
      this.logger.warn('No user context found for audited request')
      throw new UnauthorizedException()
    }

    return new Observable((observer) => {
      this.handleAuditedRequest(auditContext, request, next, observer)
    })
  }

  // An audit log must be created before the request is handled
  // After the request is handled, the audit log is optimistically updated with the outcome
  private async handleAuditedRequest(
    auditContext: AuditContext,
    request: RequestWithUser,
    next: CallHandler,
    observer: Subscriber<any>,
  ): Promise<void> {
    try {
      const auditLog = await this.auditService.createLog({
        userId: request.user.userId,
        userEmail: request.user.email,
        organizationId: request.user.organizationId,
        action: auditContext.action,
        targetType: auditContext.targetType,
        targetId: this.resolveTargetId(auditContext, request),
        ipAddress: request.ip,
        userAgent: request.get('user-agent'),
        source: request.get(CustomHeaders.SOURCE.name),
        outcome: AuditOutcome.UNKNOWN,
        metadata: this.resolveRequestMetadata(auditContext, request),
      })

      try {
        const result = await firstValueFrom(next.handle())
        const targetId = this.resolveTargetId(auditContext, request, result)
        await this.recordSuccessOutcome(auditLog, targetId)
        observer.next(result)
        observer.complete()
      } catch (handlerError) {
        const errorMessage = handlerError.message || 'Unknown error'
        await this.recordErrorOutcome(auditLog, errorMessage)
        observer.error(handlerError)
      }
    } catch (createLogError) {
      this.logger.error('Failed to create audit log:', createLogError)
      observer.error(new InternalServerErrorException())
    }
  }

  private async recordSuccessOutcome(auditLog: AuditLog, targetId: string | null): Promise<void> {
    try {
      await this.auditService.updateLog(auditLog.id, {
        outcome: AuditOutcome.SUCCESS,
        targetId,
      })
    } catch (error) {
      this.logger.error('Failed to set "success" outcome for audit log:', error)
    }
  }

  private async recordErrorOutcome(auditLog: AuditLog, errorMessage: string): Promise<void> {
    try {
      await this.auditService.updateLog(auditLog.id, {
        outcome: AuditOutcome.ERROR,
        errorMessage,
      })
    } catch (error) {
      this.logger.error('Failed to set "error" outcome for audit log:', error)
    }
  }

  private resolveTargetId(auditContext: AuditContext, request: RequestWithUser, result?: any): string | null {
    if (auditContext.targetIdFromRequest) {
      const targetId = auditContext.targetIdFromRequest(request)
      if (targetId) {
        return targetId
      }
    }

    if (auditContext.targetIdFromResult && result) {
      const targetId = auditContext.targetIdFromResult(result)
      if (targetId) {
        return targetId
      }
    }

    return null
  }

  private resolveRequestMetadata(auditContext: AuditContext, request: RequestWithUser): AuditLogMetadata | null {
    if (!auditContext.requestMetadata) {
      return null
    }

    const resolvedMetadata: AuditLogMetadata = {}

    for (const [key, valueOrResolver] of Object.entries(auditContext.requestMetadata)) {
      try {
        if (typeof valueOrResolver === 'function') {
          resolvedMetadata[key] = valueOrResolver(request)
        } else {
          resolvedMetadata[key] = valueOrResolver
        }
      } catch (error) {
        this.logger.warn(`Failed to resolve audit log metadata key "${key}":`, error)
        resolvedMetadata[key] = null
      }
    }

    return Object.keys(resolvedMetadata).length > 0 ? resolvedMetadata : null
  }
}
