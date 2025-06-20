/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable, NotFoundException } from '@nestjs/common'
import { InjectRepository } from '@nestjs/typeorm'
import { Repository } from 'typeorm'
import { AuditLog } from '../entities/audit-log.entity'
import { CreateAuditLogInternalDto } from '../dto/create-audit-log-internal.dto'
import { UpdateAuditLogInternalDto } from '../dto/update-audit-log-internal.dto'

@Injectable()
export class AuditService {
  constructor(
    @InjectRepository(AuditLog)
    private readonly auditLogRepository: Repository<AuditLog>,
  ) {}

  async createLog(createDto: CreateAuditLogInternalDto): Promise<AuditLog> {
    const auditLog = new AuditLog()
    auditLog.actorId = createDto.actorId
    auditLog.actorEmail = createDto.actorEmail
    auditLog.organizationId = createDto.organizationId
    auditLog.action = createDto.action
    auditLog.targetType = createDto.targetType
    auditLog.targetId = createDto.targetId
    auditLog.outcome = createDto.outcome
    auditLog.errorMessage = createDto.errorMessage
    auditLog.ipAddress = createDto.ipAddress
    auditLog.userAgent = createDto.userAgent
    auditLog.source = createDto.source
    auditLog.metadata = createDto.metadata

    return await this.auditLogRepository.save(auditLog)
  }

  async updateLog(id: string, updateDto: UpdateAuditLogInternalDto): Promise<AuditLog> {
    const auditLog = await this.auditLogRepository.findOne({ where: { id } })
    if (!auditLog) {
      throw new NotFoundException(`Audit log with ID ${id} not found`)
    }

    Object.assign(auditLog, updateDto)
    return await this.auditLogRepository.save(auditLog)
  }
}
