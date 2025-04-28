/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Test, TestingModule } from '@nestjs/testing'
import { NodeService } from './node.service'
import { getRepositoryToken } from '@nestjs/typeorm'
import { Node } from '../entities/node.entity'
import { Repository } from 'typeorm'
import { UserService } from '../../user/user.service'
import { User } from '../../user/user.entity'
import { WorkspaceClass } from '../enums/workspace-class.enum'
import { NodeRegion } from '../enums/node-region.enum'

const nodeArray: Node[] = [
  {
    id: '1',
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
  {
    id: '2',
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

const oneNode: Node = {
  id: '1',
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
}

describe('NodeService', () => {
  let service: NodeService
  let repository: Repository<Node>

  beforeEach(async () => {
    const module: TestingModule = await Test.createTestingModule({
      providers: [
        NodeService,
        {
          provide: getRepositoryToken(Node),
          useValue: {
            find: jest.fn().mockResolvedValue(nodeArray),
            findOneBy: jest.fn().mockResolvedValue(oneNode),
            save: jest.fn().mockResolvedValue(oneNode),
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

    service = module.get<NodeService>(NodeService)
    repository = module.get<Repository<Node>>(getRepositoryToken(Node))
  })

  it('should be defined', () => {
    expect(service).toBeDefined()
  })

  describe('create()', () => {
    it('should successfully insert a node', () => {
      const oneNode: Node = {
        id: '1',
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
      }

      expect(
        service.create({
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
        }),
      ).resolves.toEqual(oneNode)
    })
  })

  describe('findAll()', () => {
    it('should return an array of nodes', async () => {
      const nodes = await service.findAll()
      expect(nodes).toEqual(nodeArray)
    })
  })

  describe('findOne()', () => {
    it('should get a single node', () => {
      const repoSpy = jest.spyOn(repository, 'findOneBy')
      expect(service.findOne('6d225ef9-b6e1-4061-81c6-a9cf639a8897')).resolves.toEqual(oneNode)
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
