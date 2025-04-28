/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable, NotFoundException, HttpException, BadRequestException, Logger } from '@nestjs/common'
import { InjectRepository } from '@nestjs/typeorm'
import { Repository } from 'typeorm'
import { Workspace } from '../entities/workspace.entity'
import { Node } from '../entities/node.entity'
import axios from 'axios'
import { WorkspaceState } from '../enums/workspace-state.enum'

@Injectable()
export class ToolboxService {
  private readonly logger = new Logger(ToolboxService.name)

  constructor(
    @InjectRepository(Workspace)
    private readonly workspaceRepository: Repository<Workspace>,
    @InjectRepository(Node)
    private readonly nodeRepository: Repository<Node>,
  ) {}

  async forwardRequestToNode(workspaceId: string, method: string, path: string, data?: any): Promise<any> {
    const node = await this.getNode(workspaceId)

    const maxRetries = 5
    let attempt = 1

    while (attempt <= maxRetries) {
      try {
        const headers: any = {
          Authorization: `Bearer ${node.apiKey}`,
        }

        // Only set Content-Type for requests with body data
        if (data && typeof data === 'object' && Object.keys(data).length > 0) {
          headers['Content-Type'] = 'application/json'
        }

        const requestConfig: any = {
          method,
          // TODO: remove /main from the path after the node-agent refactor
          url: `${node.apiUrl}/workspaces/${workspaceId}/main${path}`,
          headers,
          maxBodyLength: 209715200, // 200MB in bytes
          maxContentLength: 209715200, // 200MB in bytes
          timeout: 360000, // 360 seconds
        }

        // Only add data if it's not an empty string or undefined
        if (data !== undefined && data !== '') {
          requestConfig.data = data
        }

        const response = await axios(requestConfig)
        return response.data
      } catch (error) {
        if (error.message.includes('ECONNREFUSED')) {
          if (attempt === maxRetries) {
            throw new HttpException('Failed to connect to node after multiple attempts', 500)
          }
          // Wait for attempt * 1000ms (1s, 2s, 3s)
          await new Promise((resolve) => setTimeout(resolve, attempt * 1000))
          attempt++
          continue
        }
        // If it's an axios error with a response, throw a NestJS HttpException
        if (error.response) {
          throw new HttpException(error.response.data, error.response.status)
        }

        // For other types of errors, throw a generic 500 error
        throw new HttpException(`Error forwarding request to node: ${error.message}`, 500)
      }
    }
  }

  public async getNode(workspaceId: string): Promise<Node> {
    try {
      const workspace = await this.workspaceRepository.findOne({
        where: { id: workspaceId },
      })

      if (!workspace) {
        throw new NotFoundException('Workspace not found')
      }

      const node = await this.nodeRepository.findOne({
        where: { id: workspace.nodeId },
      })

      if (!node) {
        throw new NotFoundException('Node not found for the workspace')
      }

      if (workspace.state !== WorkspaceState.STARTED) {
        throw new BadRequestException('Workspace is not running')
      }

      return node
    } finally {
      await this.workspaceRepository.update(workspaceId, {
        lastActivityAt: new Date(),
      })
    }
  }
}
