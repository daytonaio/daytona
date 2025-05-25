/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Test } from '@nestjs/testing'
import { SandboxController } from './sandbox.controller'
import { SandboxService } from '../services/sandbox.service'
import { Sandbox } from '../entities/sandbox.entity'
import { getRepositoryToken } from '@nestjs/typeorm'
import { UserService } from '../../user/user.service'
import { User } from '../../user/user.entity'
import { SandboxDesiredState } from '../enums/sandbox-desired-state.enum'
import { SandboxState } from '../enums/sandbox-state.enum'
import { SandboxClass } from '../enums/sandbox-class.enum'
import { RunnerRegion } from '../enums/runner-region.enum'

describe('SandboxController', () => {
  let sandboxController: SandboxController
  let sandboxService: SandboxService

  beforeEach(async () => {
    const moduleRef = await Test.createTestingModule({
      controllers: [SandboxController],
      providers: [
        SandboxService,
        {
          provide: getRepositoryToken(Sandbox),
          useValue: {
            find: jest.fn().mockResolvedValue([]),
            findOne: jest.fn().mockResolvedValue(new Sandbox()),
            save: jest.fn().mockResolvedValue(new Sandbox()),
            delete: jest.fn().mockResolvedValue({ affected: 1 }),
          },
        },
        {
          provide: UserService,
          useValue: {
            findOne: jest.fn().mockResolvedValue(new User()),
          },
        },
      ],
    }).compile()

    sandboxService = moduleRef.get<SandboxService>(SandboxService)
    sandboxController = moduleRef.get<SandboxController>(SandboxController)
  })

  describe('findAll', () => {
    it('should return an array of sandboxes', async () => {
      const result: Sandbox[] = [
        {
          id: 'id1',
          name: 'name1',
          userId: 'userId1',
          desiredState: SandboxDesiredState.STARTED,
          state: SandboxState.UNKNOWN,
          class: SandboxClass.SMALL,
          region: RunnerRegion.US,
        },
      ]
      jest.spyOn(sandboxService, 'findAll').mockImplementation(async () => result)

      expect(await sandboxController.findAll()).toBe(result)
    })
  })
})
