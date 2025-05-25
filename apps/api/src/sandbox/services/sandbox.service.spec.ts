/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Test, TestingModule } from '@nestjs/testing'
import { SandboxService } from './sandbox.service'
import { getRepositoryToken } from '@nestjs/typeorm'
import { Sandbox } from '../entities/sandbox.entity'
import { Repository } from 'typeorm'
import { UserService } from '../../user/user.service'
import { User } from '../../user/user.entity'
import { SandboxDesiredState } from '../enums/sandbox-desired-state.enum'
import { SandboxState } from '../enums/sandbox-state.enum'
import { SandboxClass } from '../enums/sandbox-class.enum'
import { RunnerRegion } from '../enums/runner-region.enum'

const sandboxArray: Sandbox[] = [
  {
    id: '1',
    name: 'sandbox #1',
    userId: 'user-1',
    desiredState: SandboxDesiredState.STARTED,
    state: SandboxState.UNKNOWN,
    class: SandboxClass.SMALL,
    region: RunnerRegion.US,
  },
  {
    id: '2',
    name: 'sandbox #2',
    userId: 'user-2',
    desiredState: SandboxDesiredState.STARTED,
    state: SandboxState.UNKNOWN,
    class: SandboxClass.SMALL,
    region: RunnerRegion.US,
  },
]

const oneSandbox: Sandbox = {
  id: '1',
  name: 'sandbox #1',
  userId: 'user-1',
  desiredState: SandboxDesiredState.STARTED,
  state: SandboxState.UNKNOWN,
  class: SandboxClass.SMALL,
  region: RunnerRegion.US,
}

describe('SandboxService', () => {
  let service: SandboxService
  let repository: Repository<Sandbox>

  beforeEach(async () => {
    const module: TestingModule = await Test.createTestingModule({
      providers: [
        SandboxService,
        {
          provide: getRepositoryToken(Sandbox),
          useValue: {
            find: jest.fn().mockResolvedValue(sandboxArray),
            findOneBy: jest.fn().mockResolvedValue(oneSandbox),
            save: jest.fn().mockResolvedValue(oneSandbox),
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

    service = module.get<SandboxService>(SandboxService)
    repository = module.get<Repository<Sandbox>>(getRepositoryToken(Sandbox))
  })

  it('should be defined', () => {
    expect(service).toBeDefined()
  })

  describe('create()', () => {
    it('should successfully insert a sandbox', () => {
      const oneSandbox: Sandbox = {
        id: '1',
        name: 'sandbox #1',
        userId: 'user-1',
        desiredState: SandboxDesiredState.STARTED,
        state: SandboxState.UNKNOWN,
        class: SandboxClass.SMALL,
        region: RunnerRegion.US,
      }

      expect(
        service.create({
          name: oneSandbox.name,
          userId: oneSandbox.userId,
          class: SandboxClass.SMALL,
          region: RunnerRegion.US,
        }),
      ).resolves.toEqual(oneSandbox)
    })
  })

  describe('findAll()', () => {
    it('should return an array of sandboxes', async () => {
      const sandboxes = await service.findAll()
      expect(sandboxes).toEqual(sandboxArray)
    })
  })

  describe('findOne()', () => {
    it('should get a single sandbox', () => {
      const repoSpy = jest.spyOn(repository, 'findOneBy')
      expect(service.findOne('6d225ef9-b6e1-4061-81c6-a9cf639a8897')).resolves.toEqual(oneSandbox)
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
