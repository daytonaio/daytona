/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Redis } from 'ioredis'
import { NotificationService } from './notification.service'
import { NotificationEmitter } from '../gateways/notification-emitter.abstract'
import { OrganizationUserRemovedEvent } from '../../organization/events/organization-user-removed.event'
import { RegionService } from '../../region/services/region.service'
import { SandboxService } from '../../sandbox/services/sandbox.service'

describe('NotificationService', () => {
  let service: NotificationService
  let emitter: jest.Mocked<Pick<NotificationEmitter, 'evictUserFromOrganization'>>

  beforeEach(() => {
    emitter = { evictUserFromOrganization: jest.fn() }
    // Only the emitter is exercised by the membership-removal handler; the other dependencies are
    // not touched on this path.
    service = new NotificationService(
      emitter as unknown as NotificationEmitter,
      undefined as unknown as RegionService,
      undefined as unknown as SandboxService,
      undefined as unknown as Redis,
    )
  })

  describe('handleOrganizationUserRemoved', () => {
    it('evicts the removed user from the organization room (userId keyed, org room target)', () => {
      service.handleOrganizationUserRemoved(new OrganizationUserRemovedEvent('user-1', 'org-1'))

      expect(emitter.evictUserFromOrganization).toHaveBeenCalledTimes(1)
      expect(emitter.evictUserFromOrganization).toHaveBeenCalledWith('user-1', 'org-1')
    })

    it('swallows eviction failures so they cannot fail the membership-removal request', () => {
      emitter.evictUserFromOrganization.mockImplementation(() => {
        throw new Error('redis down')
      })

      expect(() =>
        service.handleOrganizationUserRemoved(new OrganizationUserRemovedEvent('user-1', 'org-1')),
      ).not.toThrow()
    })
  })
})
