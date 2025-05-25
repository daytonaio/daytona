/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable } from '@nestjs/common'
import { InjectRepository } from '@nestjs/typeorm'
import { Sandbox } from '../entities/sandbox.entity'
import { Repository } from 'typeorm'

@Injectable()
export class SandboxProvisioner {
  constructor(
    @InjectRepository(Sandbox)
    private readonly sandboxRepository: Repository<Sandbox>,
  ) {}
}
