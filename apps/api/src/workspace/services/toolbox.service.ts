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
import { ConfigService } from '@nestjs/config'

@Injectable()
export class ToolboxService {
  private readonly logger = new Logger(ToolboxService.name)

  constructor(
    @InjectRepository(Workspace)
    private readonly workspaceRepository: Repository<Workspace>,
    @InjectRepository(Node)
    private readonly nodeRepository: Repository<Node>,
    private readonly configService: ConfigService,
  ) {}

  async forwardRequestToNode(sandboxId: string, method: string, path: string, data?: any): Promise<any> {
    const node = await this.getNode(sandboxId)

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

        const proxyUrl = await this.getProxyUrl(sandboxId)

        const requestConfig: any = {
          method,
          url: `${proxyUrl}sandbox/${sandboxId}/toolbox${path}`,
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

  public async getNode(sandboxId: string): Promise<Node> {
    try {
      const sandbox = await this.workspaceRepository.findOne({
        where: { id: sandboxId },
      })

      if (!sandbox) {
        throw new NotFoundException('Sandbox not found')
      }

      const node = await this.nodeRepository.findOne({
        where: { id: sandbox.nodeId },
      })

      if (!node) {
        throw new NotFoundException('Node not found for the sandbox')
      }

      if (sandbox.state !== WorkspaceState.STARTED) {
        throw new BadRequestException('Sandbox is not running')
      }

      return node
    } finally {
      await this.workspaceRepository.update(sandboxId, {
        lastActivityAt: new Date(),
      })
    }
  }

  public async getProxyUrl(sandboxId: string): Promise<string> {
    const node = await this.getNode(sandboxId)

    const proxyPort = this.configService.get<number>('RUNNER_PROXY_PORT')

    const url = new URL(node.apiUrl)
    url.port = proxyPort.toString()

    return url.toString()
  }
}
