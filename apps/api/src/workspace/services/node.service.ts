/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable, Logger } from '@nestjs/common'
import { InjectRepository } from '@nestjs/typeorm'
import { Cron } from '@nestjs/schedule'
import { FindOptionsWhere, In, Not, Raw, Repository } from 'typeorm'
import { Node } from '../entities/node.entity'
import { CreateNodeDto } from '../dto/create-node.dto'
import { WorkspaceClass } from '../enums/workspace-class.enum'
import { NodeRegion } from '../enums/node-region.enum'
import { NodeState } from '../enums/node-state.enum'
import { BadRequestError } from '../../exceptions/bad-request.exception'
import { WorkspaceEvents } from './../../workspace/constants/workspace-events.constants'
import { OnEvent } from '@nestjs/event-emitter'
import { WorkspaceStateUpdatedEvent } from './../../workspace/events/workspace-state-updated.event'
import { WorkspaceState } from './../../workspace/enums/workspace-state.enum'
import { Workspace } from './../../workspace/entities/workspace.entity'
import { ImageNode } from './../../workspace/entities/image-node.entity'
import { ImageNodeState } from './../../workspace/enums/image-node-state.enum'
import { RunnerClientFactory } from '../runner-api/runnerApi'

@Injectable()
export class NodeService {
  private readonly logger = new Logger(NodeService.name)
  private checkingNodes = false

  constructor(
    @InjectRepository(Node)
    private readonly nodeRepository: Repository<Node>,
    private readonly runnerClientFactory: RunnerClientFactory,
    @InjectRepository(Workspace)
    private readonly workspaceRepository: Repository<Workspace>,
    @InjectRepository(ImageNode)
    private readonly imageNodeRepository: Repository<ImageNode>,
  ) {}

  async create(createNodeDto: CreateNodeDto): Promise<Node> {
    // Validate region and class
    if (!this.isValidRegion(createNodeDto.region)) {
      throw new Error('Invalid region')
    }
    if (!this.isValidClass(createNodeDto.class)) {
      throw new Error('Invalid class')
    }

    const node = new Node()
    node.domain = createNodeDto.domain
    node.apiUrl = createNodeDto.apiUrl
    node.apiKey = createNodeDto.apiKey
    node.cpu = createNodeDto.cpu
    node.memory = createNodeDto.memory
    node.disk = createNodeDto.disk
    node.gpu = createNodeDto.gpu
    node.gpuType = createNodeDto.gpuType
    node.used = 0
    node.capacity = createNodeDto.capacity
    node.region = createNodeDto.region
    node.class = createNodeDto.class

    return this.nodeRepository.save(node)
  }

  async findAll(): Promise<Node[]> {
    return this.nodeRepository.find()
  }

  findOne(id: string): Promise<Node | null> {
    return this.nodeRepository.findOneBy({ id })
  }

  async findAvailableNodes(params: GetNodeParams): Promise<Node[]> {
    const nodeFilter: FindOptionsWhere<Node> = {
      state: NodeState.READY,
      unschedulable: Not(true),
      used: Raw((alias) => `${alias} < capacity`),
    }

    if (params.imageRef !== undefined) {
      const imageNodes = await this.imageNodeRepository.find({
        where: {
          state: ImageNodeState.READY,
          imageRef: params.imageRef,
        },
      })

      let nodeIds = imageNodes.map((imageNode) => imageNode.nodeId)

      if (params.excludedNodeIds?.length) {
        nodeIds = nodeIds.filter((id) => !params.excludedNodeIds.includes(id))
      }

      if (!nodeIds.length) {
        return []
      }

      nodeFilter.id = In(nodeIds)
    } else if (params.excludedNodeIds?.length) {
      nodeFilter.id = Not(In(params.excludedNodeIds))
    }

    if (params.region !== undefined) {
      nodeFilter.region = params.region
    }

    if (params.workspaceClass !== undefined) {
      nodeFilter.class = params.workspaceClass
    }

    const nodes = await this.nodeRepository.find({
      where: nodeFilter,
    })

    return nodes.sort((a, b) => a.used / a.capacity - b.used / b.capacity).slice(0, 10)
  }

  async remove(id: string): Promise<void> {
    await this.nodeRepository.delete(id)
  }

  @OnEvent(WorkspaceEvents.STATE_UPDATED)
  async handleWorkspaceStateUpdate(event: WorkspaceStateUpdatedEvent) {
    if (![WorkspaceState.DESTROYED, WorkspaceState.CREATING, WorkspaceState.ARCHIVED].includes(event.newState)) {
      return
    }

    await this.recalculateNodeUsage(event.workspace.nodeId)
  }

  @Cron('45 * * * * *')
  private async handleCheckNodes() {
    if (this.checkingNodes) {
      return
    }
    this.checkingNodes = true
    const nodes = await this.nodeRepository.find({
      where: {
        unschedulable: Not(true),
      },
    })
    for (const node of nodes) {
      this.logger.debug(`Checking node ${node.id}`)
      try {
        const runnerClient = this.runnerClientFactory.create(node)
        this.logger.debug(`Attempting health check for node ${node.id} at ${node.apiUrl}`)

        // Add timeout to the health check
        const response = (await Promise.race([
          runnerClient.healthCheck({}),
          new Promise((_, reject) => setTimeout(() => reject(new Error('Health check timeout')), 5000)),
        ])) as { status: string; version: string }

        // Log the full response for debugging
        this.logger.debug(`Health check response from node ${node.id}:`, JSON.stringify(response))

        // Check if response is valid
        if (!response) {
          throw new Error('Empty health check response received')
        }

        // Check if status field exists and is a string
        if (!response.status || typeof response.status !== 'string') {
          throw new Error('Health check response missing status field')
        }

        // Verify the health check response
        if (response.status !== 'healthy') {
          throw new Error(`Node reported unhealthy status: ${response.status}`)
        }

        await this.nodeRepository.update(node.id, {
          state: NodeState.READY,
          lastChecked: new Date(),
        })

        await this.recalculateNodeUsage(node.id)
      } catch (e) {
        this.logger.error(`Error checking node ${node.id}: ${e.message}`)
        this.logger.error(e)

        // Update node state to UNRESPONSIVE for any error
        await this.nodeRepository.update(node.id, {
          state: NodeState.UNRESPONSIVE,
          lastChecked: new Date(),
        })
      }
    }
    this.checkingNodes = false
  }

  async recalculateNodeUsage(nodeId: string) {
    const node = await this.nodeRepository.findOne({ where: { id: nodeId } })
    if (!node) {
      throw new Error('Node not found')
    }
    //  recalculate node usage
    const workspaces = await this.workspaceRepository.find({
      where: {
        nodeId: node.id,
        state: Not(WorkspaceState.DESTROYED),
      },
    })
    node.used = workspaces.length

    await this.nodeRepository.save(node)
  }

  private isValidRegion(region: NodeRegion): boolean {
    return Object.values(NodeRegion).includes(region)
  }

  private isValidClass(workspaceClass: WorkspaceClass): boolean {
    return Object.values(WorkspaceClass).includes(workspaceClass)
  }

  async updateSchedulingStatus(id: string, unschedulable: boolean): Promise<Node> {
    const node = await this.nodeRepository.findOne({ where: { id } })
    if (!node) {
      throw new Error('Node not found')
    }

    node.unschedulable = unschedulable
    return this.nodeRepository.save(node)
  }

  async getRandomAvailableNode(params: GetNodeParams): Promise<string> {
    const availableNodes = await this.findAvailableNodes(params)

    //  TODO: implement a better algorithm to get a random available node based on the node's usage

    if (availableNodes.length === 0) {
      throw new BadRequestError('No available nodes')
    }

    // Get random node from available nodes using inclusive bounds
    const randomIntFromInterval = (min: number, max: number) => Math.floor(Math.random() * (max - min + 1) + min)

    return availableNodes[randomIntFromInterval(0, availableNodes.length - 1)].id
  }

  async getImageNode(nodeId, imageRef: string): Promise<ImageNode> {
    return this.imageNodeRepository.findOne({
      where: {
        nodeId,
        imageRef,
      },
    })
  }

  async getImageNodes(imageRef: string): Promise<ImageNode[]> {
    return this.imageNodeRepository.find({
      where: {
        imageRef,
      },
      order: {
        state: 'ASC', // Sorts state BUILDING_IMAGE before ERROR
        createdAt: 'ASC', // Sorts first node to start building image on top
      },
    })
  }

  async createImageNode(nodeId: string, imageRef: string, state: ImageNodeState, errorReason?: string): Promise<void> {
    const imageNode = new ImageNode()
    imageNode.nodeId = nodeId
    imageNode.imageRef = imageRef
    imageNode.state = state
    if (errorReason) {
      imageNode.errorReason = errorReason
    }
    await this.imageNodeRepository.save(imageNode)
  }

  async getNodesWithMultipleImagesBuilding(maxImageCount = 2): Promise<string[]> {
    const nodes = await this.workspaceRepository
      .createQueryBuilder('workspace')
      .select('workspace.nodeId')
      .where('workspace.state = :state', { state: WorkspaceState.BUILDING_IMAGE })
      .andWhere('workspace.buildInfoImageRef IS NOT NULL')
      .groupBy('workspace.nodeId')
      .having('COUNT(DISTINCT workspace.buildInfoImageRef) > :maxImageCount', { maxImageCount })
      .getRawMany()

    return nodes.map((item) => item.nodeId)
  }
}

export class GetNodeParams {
  region?: NodeRegion
  workspaceClass?: WorkspaceClass
  imageRef?: string
  excludedNodeIds?: string[]
}
