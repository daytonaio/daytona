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
import { Region } from '../entities/region.entity'
import { Sandbox } from '../entities/sandbox.entity'
import { SnapshotRunner } from '../entities/snapshot-runner.entity'
import { Snapshot } from '../entities/snapshot.entity'
import { RunnerAdapterFactory } from '../runner-adapter/runnerAdapter'

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
            findOneBy: jest.fn().mockResolvedValue(new Runner()),
            save: jest.fn().mockResolvedValue(new Runner()),
            delete: jest.fn().mockResolvedValue({ affected: 1 }),
          },
        },
        {
          provide: getRepositoryToken(Region),
          useValue: {
            findOne: jest.fn(),
          },
        },
        {
          provide: getRepositoryToken(Sandbox),
          useValue: {
            find: jest.fn(),
          },
        },
        {
          provide: getRepositoryToken(SnapshotRunner),
          useValue: {
            find: jest.fn(),
          },
        },
        {
          provide: getRepositoryToken(Snapshot),
          useValue: {
            find: jest.fn(),
          },
        },
        {
          provide: UserService,
          useValue: {
            findOne: jest.fn().mockResolvedValue(new User()),
          },
        },
        {
          provide: RunnerAdapterFactory,
          useValue: {
            create: jest.fn(),
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
          domain: 'test',
          apiUrl: 'http://test.com',
          proxyUrl: 'http://proxy.test.com',
          apiKey: 'test-key',
          capacity: 1,
          used: 0,
          state: 'READY' as any,
          version: '1.0',
          unschedulable: false,
          currentCpuUsagePercentage: 0,
          currentMemoryUsagePercentage: 0,
          currentDiskUsagePercentage: 0,
          currentAllocatedCpu: 0,
          currentAllocatedMemoryGiB: 0,
          currentAllocatedDiskGiB: 0,
          currentSnapshotCount: 0,
          availabilityScore: 100,
          lastChecked: new Date(),
          createdAt: new Date(),
          updatedAt: new Date(),
        },
      ]

      const mockAuthContext = {
        organizationId: 'org-123',
        userId: 'user-123',
        email: 'test@example.com',
        role: 'admin' as any,
        organization: {} as any,
      }

      jest.spyOn(runnerService, 'findAll').mockImplementation(async () => result)

      expect(await runnerController.findAll(undefined, mockAuthContext)).toBe(result)
    })
  })
})
