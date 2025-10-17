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
import { RunnerState } from '../enums/runner-state.enum'
import { BadRequestError } from '../../exceptions/bad-request.exception'

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
    domain: 'test1.example.com',
    apiUrl: 'https://test1.example.com/api',
    proxyUrl: 'https://test1.example.com/proxy',
    apiKey: 'test-key-1',
    state: RunnerState.READY,
    availabilityScore: 80,
    currentCpuUsagePercentage: 0,
    currentMemoryUsagePercentage: 0,
    currentDiskUsagePercentage: 0,
    currentAllocatedCpu: 0,
    currentAllocatedMemoryGiB: 0,
    currentAllocatedDiskGiB: 0,
    currentSnapshotCount: 0,
    version: '1.0.0',
    unschedulable: false,
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
    domain: 'test2.example.com',
    apiUrl: 'https://test2.example.com/api',
    proxyUrl: 'https://test2.example.com/proxy',
    apiKey: 'test-key-2',
    state: RunnerState.READY,
    availabilityScore: 90,
    currentCpuUsagePercentage: 0,
    currentMemoryUsagePercentage: 0,
    currentDiskUsagePercentage: 0,
    currentAllocatedCpu: 0,
    currentAllocatedMemoryGiB: 0,
    currentAllocatedDiskGiB: 0,
    currentSnapshotCount: 0,
    version: '1.0.0',
    unschedulable: false,
    lastChecked: new Date(),
    createdAt: new Date(),
    updatedAt: new Date(),
  },
  {
    id: '3',
    class: SandboxClass.SMALL,
    region: 'us',
    cpu: 1,
    diskGiB: 1,
    memoryGiB: 1,
    gpu: 1,
    gpuType: 'test',
    domain: 'test3.example.com',
    apiUrl: 'https://test3.example.com/api',
    proxyUrl: 'https://test3.example.com/proxy',
    apiKey: 'test-key-3',
    state: RunnerState.READY,
    availabilityScore: 85,
    currentCpuUsagePercentage: 0,
    currentMemoryUsagePercentage: 0,
    currentDiskUsagePercentage: 0,
    currentAllocatedCpu: 0,
    currentAllocatedMemoryGiB: 0,
    currentAllocatedDiskGiB: 0,
    currentSnapshotCount: 0,
    version: '1.0.0',
    unschedulable: false,
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
  domain: 'test1.example.com',
  apiUrl: 'https://test1.example.com/api',
  proxyUrl: 'https://test1.example.com/proxy',
  apiKey: 'test-key-1',
  state: RunnerState.READY,
  availabilityScore: 80,
  currentCpuUsagePercentage: 0,
  currentMemoryUsagePercentage: 0,
  currentDiskUsagePercentage: 0,
  currentAllocatedCpu: 0,
  currentAllocatedMemoryGiB: 0,
  currentAllocatedDiskGiB: 0,
  currentSnapshotCount: 0,
  version: '1.0.0',
  unschedulable: false,
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
          provide: UserService,
          useValue: {
            findOne: jest.fn().mockResolvedValue(new User()),
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
        domain: 'test1.example.com',
        apiUrl: 'https://test1.example.com/api',
        proxyUrl: 'https://test1.example.com/proxy',
        apiKey: 'test-key-1',
        state: RunnerState.INITIALIZING,
        availabilityScore: 0,
        currentCpuUsagePercentage: 0,
        currentMemoryUsagePercentage: 0,
        currentDiskUsagePercentage: 0,
        currentAllocatedCpu: 0,
        currentAllocatedMemoryGiB: 0,
        currentAllocatedDiskGiB: 0,
        currentSnapshotCount: 0,
        version: '1.0.0',
        unschedulable: false,
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
          domain: 'test1.example.com',
          apiUrl: 'https://test1.example.com/api',
          proxyUrl: 'https://test1.example.com/proxy',
          apiKey: 'test-key-1',
          version: '1.0.0',
        }),
      ).resolves.toEqual(oneRunner)
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

  describe('getRandomAvailableRunner()', () => {
    beforeEach(() => {
      // Mock the findAvailableRunners method to return our test data
      jest.spyOn(service, 'findAvailableRunners').mockResolvedValue(runnerArray)
    })

    it('should return a random runner from available runners when no preferred runners specified', async () => {
      const params = {
        region: 'us',
        sandboxClass: SandboxClass.SMALL,
      }

      const result = await service.getRandomAvailableRunner(params)

      expect(result).toBeDefined()
      expect(runnerArray).toContain(result)
      expect(service.findAvailableRunners).toHaveBeenCalledWith(params)
    })

    it('should prioritize preferred runners when they are available', async () => {
      const params = {
        region: 'us',
        sandboxClass: SandboxClass.SMALL,
        preferredRunnerIds: ['2', '3'], // Prefer runners 2 and 3
      }

      const result = await service.getRandomAvailableRunner(params)

      expect(result).toBeDefined()
      expect(['2', '3']).toContain(result.id)
      expect(service.findAvailableRunners).toHaveBeenCalledWith(params)
    })

    it('should fallback to all available runners when preferred runners are not available', async () => {
      const params = {
        region: 'us',
        sandboxClass: SandboxClass.SMALL,
        preferredRunnerIds: ['999', '888'], // Non-existent runner IDs
      }

      const result = await service.getRandomAvailableRunner(params)

      expect(result).toBeDefined()
      expect(runnerArray).toContain(result)
      expect(service.findAvailableRunners).toHaveBeenCalledWith(params)
    })

    it('should throw error when no available runners', async () => {
      jest.spyOn(service, 'findAvailableRunners').mockResolvedValue([])

      const params = {
        region: 'us',
        sandboxClass: SandboxClass.SMALL,
      }

      await expect(service.getRandomAvailableRunner(params)).rejects.toThrow(BadRequestError)
      await expect(service.getRandomAvailableRunner(params)).rejects.toThrow('No available runners')
    })
  })
})
