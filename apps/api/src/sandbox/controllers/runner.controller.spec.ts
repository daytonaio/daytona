/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Test } from '@nestjs/testing'
import { RunnerController } from './runner.controller'
import { RunnerService } from '../services/runner.service'
import { Runner } from '../entities/runner.entity'
import { getRepositoryToken } from '@nestjs/typeorm'
import { UserService } from '../../user/user.service'
import { User } from '../../user/user.entity'
import { SandboxClass } from '../enums/sandbox-class.enum'

describe('RunnerController', () => {
  let runnerController: RunnerController
  let runnerService: RunnerService

  beforeEach(async () => {
    const moduleRef = await Test.createTestingModule({
      controllers: [RunnerController],
      providers: [
        RunnerService,
        {
          provide: getRepositoryToken(Runner),
          useValue: {
            find: jest.fn().mockResolvedValue([]),
            findOne: jest.fn().mockResolvedValue(new Runner()),
            save: jest.fn().mockResolvedValue(new Runner()),
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

    runnerService = moduleRef.get<RunnerService>(RunnerService)
    runnerController = moduleRef.get<RunnerController>(RunnerController)
  })

  describe('findAll', () => {
    it('should return an array of runners', async () => {
      const result: Runner[] = [
        {
          id: 'id1',
          class: SandboxClass.SMALL,
          region: 'us',
          cpu: 1,
          diskGiB: 1,
          memoryGiB: 1,
          gpu: 1,
          gpuType: 'test',
          key: 'test',
          domain: 'test',
          limit: 1,
        },
      ]
      jest.spyOn(runnerService, 'findAll').mockImplementation(async () => result)

      expect(await runnerController.findAll()).toBe(result)
    })
  })
})
