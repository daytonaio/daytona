/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { DataSource, Repository } from 'typeorm'
import { EventEmitter2 } from '@nestjs/event-emitter'
import { OrganizationUserService } from './organization-user.service'
import { OrganizationRoleService } from './organization-role.service'
import { UserService } from '../../user/user.service'
import { OrganizationUser } from '../entities/organization-user.entity'
import { OrganizationMemberRole } from '../enums/organization-member-role.enum'
import { OrganizationEvents } from '../constants/organization-events.constant'
import { OrganizationUserRemovedEvent } from '../events/organization-user-removed.event'

describe('OrganizationUserService', () => {
  let service: OrganizationUserService
  let findOne: jest.Mock
  let remove: jest.Mock
  let count: jest.Mock
  let emit: jest.Mock

  beforeEach(() => {
    findOne = jest.fn()
    remove = jest.fn()
    count = jest.fn().mockResolvedValue(2)
    emit = jest.fn()
    const repo = { findOne, manager: { remove, count } }
    service = new OrganizationUserService(
      repo as unknown as Repository<OrganizationUser>,
      undefined as unknown as OrganizationRoleService,
      undefined as unknown as UserService,
      { emit, emitAsync: jest.fn() } as unknown as EventEmitter2,
      undefined as unknown as DataSource,
    )
  })

  describe('delete', () => {
    it('emits USER_REMOVED(userId, organizationId) after the membership row is removed', async () => {
      findOne.mockResolvedValue({
        organizationId: 'org-1',
        userId: 'user-1',
        role: OrganizationMemberRole.OWNER,
      } as unknown as OrganizationUser)

      await service.delete('org-1', 'user-1')

      expect(remove).toHaveBeenCalledTimes(1)
      expect(emit).toHaveBeenCalledWith(
        OrganizationEvents.USER_REMOVED,
        expect.objectContaining({ userId: 'user-1', organizationId: 'org-1' }),
      )
      expect(emit.mock.calls[0][1]).toBeInstanceOf(OrganizationUserRemovedEvent)
      // ordering matters: the eviction event must fire only after the removal has committed
      expect(remove.mock.invocationCallOrder[0]).toBeLessThan(emit.mock.invocationCallOrder[0])
    })

    it('does not emit USER_REMOVED when the membership does not exist', async () => {
      findOne.mockResolvedValue(null)

      await expect(service.delete('org-1', 'user-1')).rejects.toThrow()
      expect(emit).not.toHaveBeenCalled()
    })
  })
})
