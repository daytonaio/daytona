/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { OrganizationUserService } from './organization-user.service'
import { OrganizationMemberRole } from '../enums/organization-member-role.enum'
import { OrganizationResourcePermission } from '../enums/organization-resource-permission.enum'
import { createMockOrganizationUser } from '../../test/helpers/entity.factory'
import { MOCK_ORGANIZATION_ID, MOCK_USER_ID, MOCK_USER_EMAIL } from '../../test/helpers/constants'

/**
 * OrganizationAuthContextGuard caches the organization-user record under
 * `organization-user:<org>:<user>` for a short TTL. Every mutation that changes a member's role,
 * assigned roles, or membership must evict that key so the change takes effect immediately rather
 * than leaving a stale authorization record in place for the duration of the TTL.
 */
describe('[AUTH] OrganizationUserService cache eviction', () => {
  const CACHE_KEY = `organization-user:${MOCK_ORGANIZATION_ID}:${MOCK_USER_ID}`

  let service: OrganizationUserService
  let mockRepo: any
  let mockEntityManager: any
  let mockRoleService: any
  let mockUserService: any
  let mockEventEmitter: any
  let mockDataSource: any
  let mockRedis: any

  beforeEach(() => {
    mockEntityManager = {
      save: jest.fn().mockImplementation(async (entity) => entity),
      remove: jest.fn().mockResolvedValue(undefined),
      count: jest.fn().mockResolvedValue(2),
    }
    mockRepo = {
      findOne: jest.fn(),
      count: jest.fn().mockResolvedValue(2),
      save: jest.fn().mockImplementation(async (entity) => entity),
      manager: mockEntityManager,
    }
    mockRoleService = { findByIds: jest.fn().mockResolvedValue([]) }
    mockUserService = { findOne: jest.fn().mockResolvedValue({ id: MOCK_USER_ID, email: MOCK_USER_EMAIL }) }
    mockEventEmitter = { emitAsync: jest.fn().mockResolvedValue(undefined) }
    mockDataSource = { transaction: jest.fn().mockImplementation(async (cb) => cb(mockEntityManager)) }
    mockRedis = { del: jest.fn().mockResolvedValue(1) }

    service = new OrganizationUserService(
      mockRepo,
      mockRoleService,
      mockUserService,
      mockEventEmitter,
      mockDataSource,
      mockRedis,
    )
  })

  it('evicts the cached authorization record when an owner is demoted', async () => {
    mockRepo.findOne.mockResolvedValue(
      createMockOrganizationUser({ role: OrganizationMemberRole.OWNER, assignedRoles: [] }),
    )

    await service.updateAccess(MOCK_ORGANIZATION_ID, MOCK_USER_ID, OrganizationMemberRole.MEMBER, [])

    // Demotion drops permissions, so it runs through the transactional revoke branch...
    expect(mockDataSource.transaction).toHaveBeenCalled()
    // ...and the cache entry the guard reads must be evicted so the new role takes effect at once.
    expect(mockRedis.del).toHaveBeenCalledWith(CACHE_KEY)
  })

  it('evicts the cached authorization record on a non-permission-reducing update', async () => {
    mockRepo.findOne.mockResolvedValue(
      createMockOrganizationUser({
        role: OrganizationMemberRole.MEMBER,
        assignedRoles: [{ permissions: [OrganizationResourcePermission.WRITE_SANDBOXES] }] as any,
      }),
    )
    mockRoleService.findByIds.mockResolvedValue([
      { id: 'role-1', name: 'writer', permissions: [OrganizationResourcePermission.WRITE_SANDBOXES] },
    ])

    await service.updateAccess(MOCK_ORGANIZATION_ID, MOCK_USER_ID, OrganizationMemberRole.MEMBER, ['role-1'])

    // No permissions revoked -> non-transactional save branch, but eviction still happens.
    expect(mockDataSource.transaction).not.toHaveBeenCalled()
    expect(mockRepo.save).toHaveBeenCalled()
    expect(mockRedis.del).toHaveBeenCalledWith(CACHE_KEY)
  })

  it('evicts the cached authorization record when a member is removed', async () => {
    mockRepo.findOne.mockResolvedValue(createMockOrganizationUser({ role: OrganizationMemberRole.MEMBER }))

    await service.delete(MOCK_ORGANIZATION_ID, MOCK_USER_ID)

    expect(mockEntityManager.remove).toHaveBeenCalled()
    expect(mockRedis.del).toHaveBeenCalledWith(CACHE_KEY)
  })
})
