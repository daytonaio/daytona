/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Test, TestingModule } from '@nestjs/testing'
import { getRepositoryToken } from '@nestjs/typeorm'
import { RegionService } from './region.service'
import { Region } from '../entities/region.entity'
import { CreateRegionDto } from '../dto/create-region.dto'
import { Organization } from '../../organization/entities/organization.entity'
import { ConflictException, NotFoundException } from '@nestjs/common'

describe('RegionService', () => {
  let service: RegionService
  let mockRepository: any

  const mockRegion = {
    code: 'abc12345',
    name: 'us-east-1',
    organizationId: 'org-123',
    createdAt: new Date(),
    updatedAt: new Date(),
  }

  const mockOrganization = {
    id: 'org-123',
    name: 'Test Org',
  } as Organization

  beforeEach(async () => {
    mockRepository = {
      findOne: jest.fn(),
      save: jest.fn(),
      find: jest.fn(),
      remove: jest.fn(),
    }

    const module: TestingModule = await Test.createTestingModule({
      providers: [
        RegionService,
        {
          provide: getRepositoryToken(Region),
          useValue: mockRepository,
        },
      ],
    }).compile()

    service = module.get<RegionService>(RegionService)
  })

  it('should be defined', () => {
    expect(service).toBeDefined()
  })

  describe('create', () => {
    it('should create a new region', async () => {
      const createRegionDto: CreateRegionDto = { name: 'us-east-1' }

      mockRepository.findOne.mockResolvedValue(null)
      mockRepository.save.mockResolvedValue(mockRegion)

      const result = await service.create(mockOrganization, createRegionDto)

      expect(result).toEqual(mockRegion)
      expect(mockRepository.save).toHaveBeenCalled()
    })

    it('should throw ConflictException if region name already exists', async () => {
      const createRegionDto: CreateRegionDto = { name: 'us-east-1' }

      mockRepository.findOne.mockResolvedValue(mockRegion)

      await expect(service.create(mockOrganization, createRegionDto)).rejects.toThrow(ConflictException)
    })
  })

  describe('findOne', () => {
    it('should return a region by code', async () => {
      mockRepository.findOne.mockResolvedValue(mockRegion)

      const result = await service.findOne('abc12345')

      expect(result).toEqual(mockRegion)
    })

    it('should throw NotFoundException if region not found', async () => {
      mockRepository.findOne.mockResolvedValue(null)

      await expect(service.findOne('invalid-code')).rejects.toThrow(NotFoundException)
    })
  })

  describe('findAll', () => {
    it('should return all regions for an organization', async () => {
      mockRepository.find.mockResolvedValue([mockRegion])

      const result = await service.findAll('org-123')

      expect(result).toEqual([mockRegion])
      expect(mockRepository.find).toHaveBeenCalledWith({
        where: { organizationId: 'org-123' },
        order: { name: 'ASC' },
      })
    })
  })

  describe('delete', () => {
    it('should delete a region', async () => {
      mockRepository.findOne.mockResolvedValue(mockRegion)
      mockRepository.remove.mockResolvedValue(undefined)

      await service.delete('abc12345')

      expect(mockRepository.remove).toHaveBeenCalledWith(mockRegion)
    })
  })
})
