/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { RegionController } from './region.controller'
import { RegionService } from '../services/region.service'
import { CreateRegionDto } from '../dto/create-region.dto'
import { RegionDto } from '../dto/region.dto'

describe('RegionController', () => {
  let controller: RegionController
  let service: RegionService

  const mockRegion = {
    code: 'abc12345',
    name: 'us-east-1',
    organizationId: 'org-123',
    createdAt: new Date(),
    updatedAt: new Date(),
  }

  const mockRegionDto = {
    name: 'us-east-1',
    organizationId: 'org-123',
    createdAt: '2023-01-01T00:00:00.000Z',
    updatedAt: '2023-01-01T00:00:00.000Z',
  }

  const mockOrganization = {
    id: 'org-123',
    name: 'Test Org',
  }

  const mockAuthContext = {
    organization: mockOrganization,
    organizationId: 'org-123',
  }

  beforeEach(() => {
    service = {
      create: jest.fn(),
      findAll: jest.fn(),
      delete: jest.fn(),
    } as any

    controller = new RegionController(service)
  })

  it('should be defined', () => {
    expect(controller).toBeDefined()
  })

  describe('listRegions', () => {
    it('should return list of regions', async () => {
      jest.spyOn(service, 'findAll').mockResolvedValue([mockRegion])

      const result = await controller.listRegions(mockAuthContext as any)

      expect(result).toHaveLength(1)
      expect(result[0].name).toBe(mockRegionDto.name)
      expect(result[0].organizationId).toBe(mockRegionDto.organizationId)
      expect(result[0].createdAt).toBeDefined()
      expect(result[0].updatedAt).toBeDefined()
      expect(service.findAll).toHaveBeenCalledWith('org-123')
    })
  })

  describe('createRegion', () => {
    it('should create a new region', async () => {
      const createRegionDto: CreateRegionDto = { name: 'us-east-1' }

      jest.spyOn(service, 'create').mockResolvedValue(mockRegion)

      const result = await controller.createRegion(mockAuthContext as any, createRegionDto)

      expect(result.name).toBe(mockRegionDto.name)
      expect(result.organizationId).toBe(mockRegionDto.organizationId)
      expect(result.createdAt).toBeDefined()
      expect(result.updatedAt).toBeDefined()
      expect(service.create).toHaveBeenCalledWith(mockOrganization, createRegionDto)
    })
  })

  describe('deleteRegion', () => {
    it('should delete a region', async () => {
      jest.spyOn(service, 'delete').mockResolvedValue(undefined)

      await controller.deleteRegion('abc12345')

      expect(service.delete).toHaveBeenCalledWith('abc12345')
    })
  })
})
