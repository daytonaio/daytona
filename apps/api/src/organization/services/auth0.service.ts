/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable, Logger } from '@nestjs/common'
import { ManagementClient } from 'auth0'
import { Organization } from '../entities/organization.entity'
import { TypedConfigService } from '../../config/typed-config.service'

@Injectable()
export class Auth0Service {
  private readonly logger = new Logger(Auth0Service.name)
  private readonly auth0Client: ManagementClient | null

  constructor(private readonly configService: TypedConfigService) {
    const clientId = this.configService.get('oidc.managementApi.clientId')
    const clientSecret = this.configService.get('oidc.managementApi.clientSecret')
    const domain = this.configService.get('oidc.issuer').replace('https://', '')

    this.auth0Client = new ManagementClient({ domain, clientId, clientSecret })
  }

  async createOrganization(organization: Organization): Promise<void> {
    if (!this.auth0Client) {
      this.logger.debug('Auth0 Management API is not enabled, skipping organization creation')
      return
    }

    const connections = (await this.auth0Client.connections.list()).data
    await this.auth0Client.organizations.create({
      name: organization.id,
      display_name: organization.name,
      enabled_connections: connections.map((connection) => ({ connection_id: connection.id })),
    })
  }

  async addOrganizationMembers(organizationId: string, members: string[]): Promise<void> {
    if (!this.auth0Client) {
      this.logger.debug('Auth0 Management API is not enabled, skipping member addition')
      return
    }

    if (members.length != 0) {
      try {
        const auth0Organization = await this.auth0Client.organizations.getByName(organizationId)
        await this.auth0Client.organizations.members.create(auth0Organization.id, { members })
      } catch (error) {
        this.logger.debug(`Error adding organization members: ${error}`)
      }
    }
  }

  async removeOrganizationMembers(organizationId: string, members: string[]): Promise<void> {
    if (!this.auth0Client) {
      this.logger.debug('Auth0 Management API is not enabled, skipping member removal')
      return
    }

    if (members.length != 0) {
      try {
        const auth0Organization = await this.auth0Client.organizations.getByName(organizationId)
        await this.auth0Client.organizations.members.delete(auth0Organization.id, { members })
      } catch (error) {
        this.logger.debug(`Error removing organization members: ${error}`)
      }
    }
  }
}
