/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Test, TestingModule } from '@nestjs/testing'
import { RunnerService } from './runner.service'
import { getRepositoryToken } from '@nestjs/typeorm'
import { Runner } from '../entities/runner.entity'
import { Repository } from 'typeorm'
import { UserService } from '../../user/user.service'
import { User } from '../../user/user.entity'
import { SandboxClass } from '../enums/sandbox-class.enum'
import { Region } from '../entities/region.entity'
import { Sandbox } from '../entities/sandbox.entity'
import { SnapshotRunner } from '../entities/snapshot-runner.entity'
import { Snapshot } from '../entities/snapshot.entity'
import { RunnerAdapterFactory } from '../runner-adapter/runnerAdapter'

const runnerArray: Runner[] = [
  {
    id: '1',
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
  {
    id: '2',
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

const oneRunner: Runner = {
  id: '1',
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
}

describe('RunnerService', () => {
  let service: RunnerService
  let repository: Repository<Runner>

  beforeEach(async () => {
    const module: TestingModule = await Test.createTestingModule({
      providers: [
        RunnerService,
        {
          provide: getRepositoryToken(Runner),
          useValue: {
            find: jest.fn().mockResolvedValue(runnerArray),
            findOneBy: jest.fn().mockResolvedValue(oneRunner),
            save: jest.fn().mockResolvedValue(oneRunner),
            remove: jest.fn(),
            delete: jest.fn(),
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

    service = module.get<RunnerService>(RunnerService)
    repository = module.get<Repository<Runner>>(getRepositoryToken(Runner))
  })

  it('should be defined', () => {
    expect(service).toBeDefined()
  })

  describe('create()', () => {
    it('should successfully insert a runner', () => {
      const oneRunner: Runner = {
        id: '1',
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
      }

      expect(
        service.create({
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
          version: '1.0',
        }),
      ).resolves.toMatchObject({
        id: '1',
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
        state: 'READY',
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
      })
    })
  })

  describe('findAll()', () => {
    it('should return an array of runners', async () => {
      const runners = await service.findAll()
      expect(runners).toEqual(runnerArray)
    })
  })

  describe('findOne()', () => {
    it('should get a single runner', () => {
      const repoSpy = jest.spyOn(repository, 'findOneBy')
      expect(service.findOne('6d225ef9-b6e1-4061-81c6-a9cf639a8897')).resolves.toEqual(oneRunner)
      expect(repoSpy).toHaveBeenCalledWith({ id: '6d225ef9-b6e1-4061-81c6-a9cf639a8897' })
    })
  })

  describe('remove()', () => {
    it('should call remove with the passed value', async () => {
      const removeSpy = jest.spyOn(repository, 'delete')
      const retVal = await service.remove('2')
      expect(removeSpy).toHaveBeenCalledWith('2')
      expect(retVal).toBeUndefined()
    })
  })
})
