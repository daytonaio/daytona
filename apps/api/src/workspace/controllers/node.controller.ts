/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Body, Controller, Get, Post, Param, Patch, UseGuards } from '@nestjs/common'
import { CreateNodeDto } from '../dto/create-node.dto'
import { Node } from '../entities/node.entity'
import { NodeService } from '../services/node.service'
import { AuthGuard } from '@nestjs/passport'
import { ApiOAuth2, ApiTags, ApiOperation, ApiBearerAuth } from '@nestjs/swagger'
import { SystemActionGuard } from '../../auth/system-action.guard'
import { RequiredSystemRole } from '../../common/decorators/required-system-role.decorator'
import { SystemRole } from '../../user/enums/system-role.enum'

@ApiTags('nodes')
@Controller('nodes')
@UseGuards(AuthGuard('jwt'), SystemActionGuard)
@RequiredSystemRole(SystemRole.ADMIN)
@ApiOAuth2(['openid', 'profile', 'email'])
@ApiBearerAuth()
export class NodeController {
  constructor(private readonly nodeService: NodeService) {}

  @Post()
  @ApiOperation({
    summary: 'Create node',
    operationId: 'createNode',
  })
  async create(@Body() createNodeDto: CreateNodeDto): Promise<Node> {
    return this.nodeService.create(createNodeDto)
  }

  @Get()
  @ApiOperation({
    summary: 'List all nodes',
    operationId: 'listNodes',
  })
  async findAll(): Promise<Node[]> {
    return this.nodeService.findAll()
  }

  @Patch(':id/scheduling')
  @ApiOperation({
    summary: 'Update node scheduling status',
    operationId: 'updateNodeScheduling',
  })
  async updateSchedulingStatus(@Param('id') id: string, @Body('unschedulable') unschedulable: boolean): Promise<Node> {
    return this.nodeService.updateSchedulingStatus(id, unschedulable)
  }
}
