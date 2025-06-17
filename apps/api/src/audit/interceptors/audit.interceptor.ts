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
import { AuditLog } from '../entities/audit-log.entity'
import { AUDIT_METADATA_KEY, AuditMetadata } from '../decorators/audit.decorator'
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

    const auditMetadata = this.reflector.get<AuditMetadata>(AUDIT_METADATA_KEY, context.getHandler())

    if (!auditMetadata) {
      this.logger.warn('Non-audited request:', request.url)
      return next.handle()
    }

    if (!request.user) {
      this.logger.warn('No user context found for audited request')
      throw new UnauthorizedException()
    }

    return new Observable((observer) => {
      this.handleAuditedRequest(auditMetadata, request, next, observer)
    })
  }

  // An audit log must be created before the request is handled
  // After the request is handled, the audit log is optimistically updated with the outcome
  private async handleAuditedRequest(
    auditMetadata: AuditMetadata,
    request: RequestWithUser,
    next: CallHandler,
    observer: Subscriber<any>,
  ): Promise<void> {
    try {
      const auditLog = await this.auditService.createLog({
        userId: request.user.userId,
        userEmail: request.user.email,
        organizationId: request.user.organizationId,
        action: auditMetadata.action,
        targetType: auditMetadata.targetType,
        targetId: this.resolveTargetId(auditMetadata, request),
        ipAddress: request.ip,
        userAgent: request.get('user-agent'),
        source: request.get(CustomHeaders.SOURCE.name),
        outcome: AuditOutcome.UNKNOWN,
      })

      try {
        const result = await firstValueFrom(next.handle())
        const targetId = this.resolveTargetId(auditMetadata, request, result)
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

  private resolveTargetId(auditMetadata: AuditMetadata, request: RequestWithUser, result?: any): string | null {
    if (auditMetadata.targetIdParam) {
      const targetId = request.params[auditMetadata.targetIdParam]
      if (targetId) {
        return targetId
      }
    }

    if (auditMetadata.targetIdResolver && result) {
      const targetId = auditMetadata.targetIdResolver(result)
      if (targetId) {
        return targetId
      }
    }

    return null
  }
}
