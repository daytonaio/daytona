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
    key: 'test',
    domain: 'test',
    limit: 1,
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
    key: 'test',
    domain: 'test',
    limit: 1,
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
  key: 'test',
  domain: 'test',
  limit: 1,
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
        key: 'test',
        domain: 'test',
        limit: 1,
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
          key: 'test',
          domain: 'test',
          limit: 1,
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
})
