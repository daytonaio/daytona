/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Test } from '@nestjs/testing'
import { NodeController } from './node.controller'
import { NodeService } from '../services/node.service'
import { Node } from '../entities/node.entity'
import { getRepositoryToken } from '@nestjs/typeorm'
import { UserService } from '../../user/user.service'
import { User } from '../../user/user.entity'
import { WorkspaceClass } from '../enums/workspace-class.enum'
import { NodeRegion } from '../enums/node-region.enum'

describe('NodeController', () => {
  let nodeController: NodeController
  let nodeService: NodeService

  beforeEach(async () => {
    const moduleRef = await Test.createTestingModule({
      controllers: [NodeController],
      providers: [
        NodeService,
        {
          provide: getRepositoryToken(Node),
          useValue: {
            find: jest.fn().mockResolvedValue([]),
            findOne: jest.fn().mockResolvedValue(new Node()),
            save: jest.fn().mockResolvedValue(new Node()),
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

    nodeService = moduleRef.get<NodeService>(NodeService)
    nodeController = moduleRef.get<NodeController>(NodeController)
  })

  describe('findAll', () => {
    it('should return an array of nodes', async () => {
      const result: Node[] = [
        {
          id: 'id1',
          class: WorkspaceClass.SMALL,
          region: NodeRegion.US,
          cpu: 1,
          disk: 1,
          memory: 1,
          gpu: 1,
          gpuType: 'test',
          key: 'test',
          domain: 'test',
          limit: 1,
        },
      ]
      jest.spyOn(nodeService, 'findAll').mockImplementation(async () => result)

      expect(await nodeController.findAll()).toBe(result)
    })
  })
})
