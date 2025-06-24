/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ForbiddenException, Injectable, NotFoundException } from '@nestjs/common'
import { InjectRepository } from '@nestjs/typeorm'
import { In, Repository } from 'typeorm'
import { CreateOrganizationRoleDto } from '../dto/create-organization-role.dto'
import { UpdateOrganizationRoleDto } from '../dto/update-organization-role.dto'
import { OrganizationRole } from '../entities/organization-role.entity'

@Injectable()
export class OrganizationRoleService {
  constructor(
    @InjectRepository(OrganizationRole)
    private readonly organizationRoleRepository: Repository<OrganizationRole>,
  ) {}

  async create(
    organizationId: string,
    createOrganizationRoleDto: CreateOrganizationRoleDto,
  ): Promise<OrganizationRole> {
    const role = new OrganizationRole()
    role.organizationId = organizationId
    role.name = createOrganizationRoleDto.name
    role.description = createOrganizationRoleDto.description
    role.permissions = createOrganizationRoleDto.permissions
    return this.organizationRoleRepository.save(role)
  }

  async findAll(organizationId: string): Promise<OrganizationRole[]> {
    return this.organizationRoleRepository.find({
      where: [{ organizationId }, { isGlobal: true }],
      order: {
        id: 'ASC',
      },
    })
  }

  async findByIds(roleIds: string[]): Promise<OrganizationRole[]> {
    if (roleIds.length === 0) {
      return []
    }

    return this.organizationRoleRepository.find({
      where: {
        id: In(roleIds),
      },
    })
  }

  async update(roleId: string, updateOrganizationRoleDto: UpdateOrganizationRoleDto): Promise<OrganizationRole> {
    const role = await this.organizationRoleRepository.findOne({
      where: { id: roleId },
    })

    if (!role) {
      throw new NotFoundException(`Organization role with ID ${roleId} not found`)
    }

    if (role.isGlobal) {
      throw new ForbiddenException('Global roles cannot be updated')
    }

    role.name = updateOrganizationRoleDto.name
    role.description = updateOrganizationRoleDto.description
    role.permissions = updateOrganizationRoleDto.permissions

    return this.organizationRoleRepository.save(role)
  }

  async delete(roleId: string): Promise<void> {
    const role = await this.organizationRoleRepository.findOne({
      where: { id: roleId },
    })

    if (!role) {
      throw new NotFoundException(`Organization role with ID ${roleId} not found`)
    }

    if (role.isGlobal) {
      throw new ForbiddenException('Global roles cannot be deleted')
    }

    await this.organizationRoleRepository.remove(role)
  }
}
