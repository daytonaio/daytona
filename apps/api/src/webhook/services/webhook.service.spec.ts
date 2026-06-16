/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ServiceUnavailableException } from '@nestjs/common'
import { TypedConfigService } from '../../config/typed-config.service'
import { WebhookInitialization } from '../entities/webhook-initialization.entity'
import { WebhookService } from './webhook.service'
import { Repository } from 'typeorm'

describe('WebhookService', () => {
  const organizationId = 'org_123'
  const appPortalAccess = {
    token: 'appsk_test',
    url: 'https://app.svix.com/consumer/app_123',
  }

  const createService = (publicServerUrl?: string) => {
    const configService = {
      get: jest.fn((key: string) => (key === 'webhook.publicServerUrl' ? publicServerUrl : undefined)),
    } as unknown as TypedConfigService

    const service = new WebhookService(configService, {} as Repository<WebhookInitialization>)
    const appPortalAccessMock = jest.fn().mockResolvedValue(appPortalAccess)
    ;(service as unknown as { svix: unknown }).svix = {
      authentication: {
        appPortalAccess: appPortalAccessMock,
      },
    }

    return { service, appPortalAccessMock }
  }

  describe('getAppPortalAccess', () => {
    it('returns token and url from Svix', async () => {
      const { service, appPortalAccessMock } = createService()

      const result = await service.getAppPortalAccess(organizationId)

      expect(appPortalAccessMock).toHaveBeenCalledWith(organizationId, {})
      expect(result.token).toBe(appPortalAccess.token)
      expect(result.url).toBe(appPortalAccess.url)
    })

    it('includes serverUrl when browser-facing Svix URL is configured', async () => {
      const { service } = createService('https://svix.example.com')

      await expect(service.getAppPortalAccess(organizationId)).resolves.toEqual({
        ...appPortalAccess,
        serverUrl: 'https://svix.example.com',
      })
    })

    it('omits serverUrl when no Svix URL is configured', async () => {
      const { service } = createService()

      await expect(service.getAppPortalAccess(organizationId)).resolves.toEqual(appPortalAccess)
    })

    it('throws when Svix is not configured', async () => {
      const service = new WebhookService(
        { get: jest.fn() } as unknown as TypedConfigService,
        {} as Repository<WebhookInitialization>,
      )

      await expect(service.getAppPortalAccess(organizationId)).rejects.toBeInstanceOf(ServiceUnavailableException)
    })
  })
})
