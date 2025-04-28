/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Test, TestingModule } from '@nestjs/testing'
import { WorkspaceService } from './workspace.service'
import { getRepositoryToken } from '@nestjs/typeorm'
import { Workspace } from '../entities/workspace.entity'
import { Repository } from 'typeorm'
import { UserService } from '../../user/user.service'
import { User } from '../../user/user.entity'
import { WorkspaceDesiredState } from '../enums/workspace-desired-state.enum'
import { WorkspaceState } from '../enums/workspace-state.enum'
import { WorkspaceClass } from '../enums/workspace-class.enum'
import { NodeRegion } from '../enums/node-region.enum'

const workspaceArray: Workspace[] = [
  {
    id: '1',
    name: 'workspace #1',
    userId: 'user-1',
    desiredState: WorkspaceDesiredState.STARTED,
    state: WorkspaceState.UNKNOWN,
    class: WorkspaceClass.SMALL,
    region: NodeRegion.US,
  },
  {
    id: '2',
    name: 'workspace #2',
    userId: 'user-2',
    desiredState: WorkspaceDesiredState.STARTED,
    state: WorkspaceState.UNKNOWN,
    class: WorkspaceClass.SMALL,
    region: NodeRegion.US,
  },
]

const oneWorkspace: Workspace = {
  id: '1',
  name: 'workspace #1',
  userId: 'user-1',
  desiredState: WorkspaceDesiredState.STARTED,
  state: WorkspaceState.UNKNOWN,
  class: WorkspaceClass.SMALL,
  region: NodeRegion.US,
}

describe('WorkspaceService', () => {
  let service: WorkspaceService
  let repository: Repository<Workspace>

  beforeEach(async () => {
    const module: TestingModule = await Test.createTestingModule({
      providers: [
        WorkspaceService,
        {
          provide: getRepositoryToken(Workspace),
          useValue: {
            find: jest.fn().mockResolvedValue(workspaceArray),
            findOneBy: jest.fn().mockResolvedValue(oneWorkspace),
            save: jest.fn().mockResolvedValue(oneWorkspace),
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

    service = module.get<WorkspaceService>(WorkspaceService)
    repository = module.get<Repository<Workspace>>(getRepositoryToken(Workspace))
  })

  it('should be defined', () => {
    expect(service).toBeDefined()
  })

  describe('create()', () => {
    it('should successfully insert a workspace', () => {
      const oneWorkspace: Workspace = {
        id: '1',
        name: 'workspace #1',
        userId: 'user-1',
        desiredState: WorkspaceDesiredState.STARTED,
        state: WorkspaceState.UNKNOWN,
        class: WorkspaceClass.SMALL,
        region: NodeRegion.US,
      }

      expect(
        service.create({
          name: oneWorkspace.name,
          userId: oneWorkspace.userId,
          class: WorkspaceClass.SMALL,
          region: NodeRegion.US,
        }),
      ).resolves.toEqual(oneWorkspace)
    })
  })

  describe('findAll()', () => {
    it('should return an array of workspaces', async () => {
      const workspaces = await service.findAll()
      expect(workspaces).toEqual(workspaceArray)
    })
  })

  describe('findOne()', () => {
    it('should get a single workspace', () => {
      const repoSpy = jest.spyOn(repository, 'findOneBy')
      expect(service.findOne('6d225ef9-b6e1-4061-81c6-a9cf639a8897')).resolves.toEqual(oneWorkspace)
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
