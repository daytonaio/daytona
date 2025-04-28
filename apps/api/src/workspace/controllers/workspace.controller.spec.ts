/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Test } from '@nestjs/testing'
import { WorkspaceController } from './workspace.controller'
import { WorkspaceService } from '../services/workspace.service'
import { Workspace } from '../entities/workspace.entity'
import { getRepositoryToken } from '@nestjs/typeorm'
import { UserService } from '../../user/user.service'
import { User } from '../../user/user.entity'
import { WorkspaceDesiredState } from '../enums/workspace-desired-state.enum'
import { WorkspaceState } from '../enums/workspace-state.enum'
import { WorkspaceClass } from '../enums/workspace-class.enum'
import { NodeRegion } from '../enums/node-region.enum'

describe('WorkspaceController', () => {
  let workspaceController: WorkspaceController
  let workspaceService: WorkspaceService

  beforeEach(async () => {
    const moduleRef = await Test.createTestingModule({
      controllers: [WorkspaceController],
      providers: [
        WorkspaceService,
        {
          provide: getRepositoryToken(Workspace),
          useValue: {
            find: jest.fn().mockResolvedValue([]),
            findOne: jest.fn().mockResolvedValue(new Workspace()),
            save: jest.fn().mockResolvedValue(new Workspace()),
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

    workspaceService = moduleRef.get<WorkspaceService>(WorkspaceService)
    workspaceController = moduleRef.get<WorkspaceController>(WorkspaceController)
  })

  describe('findAll', () => {
    it('should return an array of workspaces', async () => {
      const result: Workspace[] = [
        {
          id: 'id1',
          name: 'name1',
          userId: 'userId1',
          desiredState: WorkspaceDesiredState.STARTED,
          state: WorkspaceState.UNKNOWN,
          class: WorkspaceClass.SMALL,
          region: NodeRegion.US,
        },
      ]
      jest.spyOn(workspaceService, 'findAll').mockImplementation(async () => result)

      expect(await workspaceController.findAll()).toBe(result)
    })
  })
})
