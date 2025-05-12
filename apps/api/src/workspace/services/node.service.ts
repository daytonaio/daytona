/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable, Logger } from '@nestjs/common'
import { InjectRepository } from '@nestjs/typeorm'
import { Cron } from '@nestjs/schedule'
import { In, Not, Repository } from 'typeorm'
import { Node } from '../entities/node.entity'
import { CreateNodeDto } from '../dto/create-node.dto'
import { WorkspaceClass } from '../enums/workspace-class.enum'
import { NodeRegion } from '../enums/node-region.enum'
import { NodeApiFactory } from '../runner-api/runnerApi'
import { NodeState } from '../enums/node-state.enum'
import { BadRequestError } from '../../exceptions/bad-request.exception'
import { WorkspaceEvents } from './../../workspace/constants/workspace-events.constants'
import { OnEvent } from '@nestjs/event-emitter'
import { WorkspaceStateUpdatedEvent } from './../../workspace/events/workspace-state-updated.event'
import { WorkspaceState } from './../../workspace/enums/workspace-state.enum'
import { Workspace } from './../../workspace/entities/workspace.entity'
import { ImageNode } from './../../workspace/entities/image-node.entity'
import { ImageNodeState } from './../../workspace/enums/image-node-state.enum'
import { ImageManager } from '../managers/image.manager'

@Injectable()
export class NodeService {
  private readonly logger = new Logger(NodeService.name)
  private checkingNodes = false

  constructor(
    @InjectRepository(Node)
    private readonly nodeRepository: Repository<Node>,
    private readonly nodeApiFactory: NodeApiFactory,
    @InjectRepository(Workspace)
    private readonly workspaceRepository: Repository<Workspace>,
    @InjectRepository(ImageNode)
    private readonly imageNodeRepository: Repository<ImageNode>,
    private readonly imageStateManager: ImageManager,
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

  async findAvailableNodes(region: NodeRegion, workspaceClass: WorkspaceClass, imageRef?: string): Promise<Node[]> {
    const whereCondition: any = {
      state: ImageNodeState.READY,
    }

    if (imageRef !== undefined) {
      whereCondition.imageRef = imageRef
    }

    const imageNodes = await this.imageNodeRepository.find({
      where: whereCondition,
    })

    const nodes = this.nodeRepository.find({
      where: {
        id: In(imageNodes.map((imageNode) => imageNode.nodeId)),
        state: NodeState.READY,
        region,
        class: workspaceClass,
        unschedulable: Not(true),
      },
    })
    return (await nodes).filter((node) => node.used < node.capacity)
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
        // Do something with the node
        const nodeApi = this.nodeApiFactory.createNodeApi(node)
        await nodeApi.healthCheck()
        await this.nodeRepository.update(node.id, {
          state: NodeState.READY,
          lastChecked: new Date(),
        })

        await this.recalculateNodeUsage(node.id)
      } catch (e) {
        if (e.code === 'ECONNREFUSED') {
          this.logger.error('Node not reachable')
        } else {
          this.logger.error(`Error checking node ${node.id}: ${e.message}`)
          this.logger.error(e)
        }

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

  async getRandomAvailableNode(region: NodeRegion, workspaceClass: WorkspaceClass, imageRef?: string): Promise<string> {
    const availableNodes = await this.findAvailableNodes(region, workspaceClass, imageRef)

    //  TODO: implement a better algorithm to get a random available node based on the node's usage

    if (availableNodes.length === 0) {
      throw new BadRequestError('No available nodes')
    }

    availableNodes.sort((a, b) => a.capacity - a.used - (b.capacity - b.used))
    //  use the first 10 nodes
    const optimalNodes = availableNodes.slice(0, 10)

    // Get random node from available nodes using inclusive bounds
    const randomIntFromInterval = (min: number, max: number) => Math.floor(Math.random() * (max - min + 1) + min)

    return optimalNodes[randomIntFromInterval(0, optimalNodes.length - 1)].id
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
    if (state != ImageNodeState.ERROR) {
      this.imageStateManager.syncNodeImageState(imageNode)
    }
  }
}
