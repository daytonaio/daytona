/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable, NotFoundException } from '@nestjs/common'
import { InjectRepository } from '@nestjs/typeorm'
import { Repository } from 'typeorm'
import { SessionTemplate } from '../entities/session-template.entity'
import { SessionTemplateDto } from '../dto/session-template.dto'

/**
 * SessionTemplateService resolves user-facing template names to backing rows. Templates
 * scoped to an org take precedence over general (org-null + general=true) templates of the
 * same name.
 */
@Injectable()
export class SessionTemplateService {
  constructor(
    @InjectRepository(SessionTemplate)
    private readonly repo: Repository<SessionTemplate>,
  ) {}

  async resolve(orgId: string, name: string): Promise<SessionTemplate> {
    const candidates = await this.repo
      .createQueryBuilder('t')
      .where('t.name = :name AND ((t.general = true AND t.organizationId IS NULL) OR t.organizationId = :orgId)', {
        name,
        orgId,
      })
      .getMany()

    if (candidates.length === 0) {
      throw new NotFoundException(`Session template "${name}" not found.`)
    }
    // Prefer org-scoped over general.
    const orgScoped = candidates.find((t) => t.organizationId === orgId)
    return orgScoped ?? candidates[0]
  }

  async list(orgId: string): Promise<SessionTemplateDto[]> {
    const rows = await this.repo
      .createQueryBuilder('t')
      .where('(t.general = true AND t.organizationId IS NULL) OR t.organizationId = :orgId', { orgId })
      .orderBy('t.name', 'ASC')
      .getMany()

    return rows.map((r) => ({
      name: r.name,
      description: r.description,
      languages: r.languages,
      packages: r.packages,
    }))
  }
}
