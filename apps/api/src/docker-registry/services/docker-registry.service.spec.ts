/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Test, TestingModule } from '@nestjs/testing'
import { DockerRegistryService } from './docker-registry.service'
import { getRepositoryToken } from '@nestjs/typeorm'
import { DockerRegistry } from '../entities/docker-registry.entity'
import { Repository } from 'typeorm'
import { DOCKER_REGISTRY_PROVIDER } from '../providers/docker-registry.provider.interface'
import { RegistryType } from '../enums/registry-type.enum'
import axios from 'axios'

// Mock axios
jest.mock('axios')
const mockedAxios = jest.mocked(axios)

describe('DockerRegistryService - getImageDetails', () => {
  let service: DockerRegistryService
  let repository: Repository<DockerRegistry>

  const mockRegistry: DockerRegistry = {
    id: 'test-registry-id',
    name: 'Test Registry',
    url: 'https://harbor.test.com',
    username: 'testuser',
    password: 'testpass',
    project: 'testproject',
    registryType: RegistryType.INTERNAL,
    isDefault: false,
    organizationId: 'test-org',
    createdAt: new Date(),
    updatedAt: new Date(),
  }

  const mockManifest = {
    schemaVersion: 2,
    mediaType: 'application/vnd.docker.distribution.manifest.v2+json',
    config: {
      digest: 'sha256:configdigest123',
      size: 1234,
      mediaType: 'application/vnd.docker.container.image.v1+json',
    },
    layers: [
      { digest: 'sha256:layer1', size: 1000000 },
      { digest: 'sha256:layer2', size: 2000000 },
      { digest: 'sha256:layer3', size: 3000000 },
    ],
  }

  const mockConfig = {
    config: {
      Entrypoint: ['/bin/bash'],
      Cmd: ['--version'],
      Env: ['PATH=/usr/local/bin:/usr/bin:/bin', 'NODE_VERSION=18.0.0'],
      WorkingDir: '/app',
      User: 'appuser',
    },
  }

  beforeEach(async () => {
    const module: TestingModule = await Test.createTestingModule({
      providers: [
        DockerRegistryService,
        {
          provide: getRepositoryToken(DockerRegistry),
          useValue: {
            find: jest.fn(),
            findOne: jest.fn(),
            save: jest.fn(),
            remove: jest.fn(),
            create: jest.fn(),
            findOneOrFail: jest.fn(),
            update: jest.fn(),
          },
        },
        {
          provide: DOCKER_REGISTRY_PROVIDER,
          useValue: {
            createRobotAccount: jest.fn(),
            deleteArtifact: jest.fn(),
          },
        },
      ],
    }).compile()

    service = module.get<DockerRegistryService>(DockerRegistryService)
    repository = module.get<Repository<DockerRegistry>>(getRepositoryToken(DockerRegistry))

    // Reset all mocks before each test
    jest.clearAllMocks()
  })

  describe('getImageDetails', () => {
    it('should successfully get image details with full config', async () => {
      // Mock registry lookup
      jest.spyOn(repository, 'find').mockResolvedValueOnce([mockRegistry])

      // Mock manifest request
      mockedAxios.mockResolvedValueOnce({
        status: 200,
        headers: {
          'docker-content-digest': 'sha256:manifestdigest123',
        },
        data: mockManifest,
      })

      // Mock config blob request
      mockedAxios.mockResolvedValueOnce({
        status: 200,
        data: mockConfig,
      })

      const result = await service.getImageDetails('harbor.test.com/testproject/myimage:latest', 'test-org')

      expect(result).toEqual({
        digest: 'sha256:manifestdigest123',
        sizeGB: 6000000 / (1024 * 1024 * 1024), // ~0.0056 GB
        entrypoint: ['/bin/bash'],
        cmd: ['--version'],
        env: ['PATH=/usr/local/bin:/usr/bin:/bin', 'NODE_VERSION=18.0.0'],
        workingDir: '/app',
        user: 'appuser',
      })

      // Verify axios calls
      expect(mockedAxios).toHaveBeenCalledTimes(2)
      expect(mockedAxios).toHaveBeenNthCalledWith(1, {
        method: 'get',
        url: 'https://harbor.test.com/v2/testproject/myimage/manifests/latest',
        headers: {
          Authorization: 'Basic dGVzdHVzZXI6dGVzdHBhc3M=', // base64 of testuser:testpass
          Accept: 'application/vnd.docker.distribution.manifest.v2+json',
        },
        validateStatus: expect.any(Function),
        timeout: 30000,
      })
    })

    it('should handle image name without registry prefix', async () => {
      // Mock registry lookup for image without prefix - should fail
      jest.spyOn(repository, 'find').mockResolvedValueOnce([])

      await expect(service.getImageDetails('testproject/myimage:v1.0')).rejects.toThrow(
        'No registry found for image testproject/myimage:v1.0',
      )

      expect(mockedAxios).toHaveBeenNthCalledWith(1, {
        method: 'get',
        url: 'https://harbor.test.com/v2/testproject/myimage/manifests/v1.0',
        headers: {
          Authorization: 'Basic dGVzdHVzZXI6dGVzdHBhc3M=',
          Accept: 'application/vnd.docker.distribution.manifest.v2+json',
        },
        validateStatus: expect.any(Function),
        timeout: 30000,
      })
    })

    it('should return basic info when config blob is missing', async () => {
      // Mock registry lookup for image without prefix - should fail since testproject/myimage:latest has no registry hostname
      jest.spyOn(repository, 'find').mockResolvedValueOnce([])

      await expect(service.getImageDetails('testproject/myimage:latest')).rejects.toThrow(
        'No registry found for image testproject/myimage:latest',
      )
    })

    it('should return basic info when config blob request fails', async () => {
      // Mock registry lookup to return our test registry
      jest.spyOn(repository, 'find').mockResolvedValueOnce([mockRegistry])

      mockedAxios.mockResolvedValueOnce({
        status: 200,
        headers: { 'docker-content-digest': 'sha256:digest123' },
        data: mockManifest,
      })

      // Config blob request fails
      mockedAxios.mockResolvedValueOnce({
        status: 404,
        statusText: 'Not Found',
      })

      const result = await service.getImageDetails('harbor.test.com/testproject/myimage:latest', 'test-org')

      expect(result).toEqual({
        digest: 'sha256:digest123',
        sizeGB: 6000000 / (1024 * 1024 * 1024), // ~0.0056 GB
        entrypoint: [],
        cmd: [],
        env: [],
      })
    })

    it('should handle localhost registry URLs', async () => {
      const localRegistry = {
        ...mockRegistry,
        url: 'localhost:5000',
      }

      mockedAxios.mockResolvedValueOnce({
        status: 200,
        headers: { 'docker-content-digest': 'sha256:digest123' },
        data: mockManifest,
      })

      mockedAxios.mockResolvedValueOnce({
        status: 200,
        data: mockConfig,
      })

      // Mock registry lookup
      jest.spyOn(repository, 'find').mockResolvedValueOnce([localRegistry])

      await service.getImageDetails('localhost:5000/testproject/myimage:latest', 'test-org')

      expect(mockedAxios).toHaveBeenNthCalledWith(
        1,
        expect.objectContaining({
          url: 'http://localhost:5000/v2/testproject/myimage/manifests/latest',
        }),
      )
    })

    it('should handle development registry URLs', async () => {
      const devRegistry = {
        ...mockRegistry,
        url: 'registry:5000',
      }

      mockedAxios.mockResolvedValueOnce({
        status: 200,
        headers: { 'docker-content-digest': 'sha256:digest123' },
        data: mockManifest,
      })

      mockedAxios.mockResolvedValueOnce({
        status: 200,
        data: mockConfig,
      })

      // Mock registry lookup
      jest.spyOn(repository, 'find').mockResolvedValueOnce([devRegistry])

      await service.getImageDetails('registry:5000/testproject/myimage:latest', 'test-org')

      expect(mockedAxios).toHaveBeenNthCalledWith(
        1,
        expect.objectContaining({
          url: 'http://registry:5000/v2/testproject/myimage/manifests/latest',
        }),
      )
    })

    it('should handle missing layers gracefully', async () => {
      const manifestWithoutLayers = {
        ...mockManifest,
        layers: undefined,
      }

      mockedAxios.mockResolvedValueOnce({
        status: 200,
        headers: { 'docker-content-digest': 'sha256:digest123' },
        data: manifestWithoutLayers,
      })

      mockedAxios.mockResolvedValueOnce({
        status: 200,
        data: mockConfig,
      })

      // Mock registry lookup
      jest.spyOn(repository, 'find').mockResolvedValueOnce([mockRegistry])

      const result = await service.getImageDetails('harbor.test.com/testproject/myimage:latest', 'test-org')

      expect(result.sizeGB).toBe(0)
    })

    it('should handle partial config data', async () => {
      const partialConfig = {
        config: {
          Entrypoint: ['/entrypoint.sh'],
          // Missing Cmd, Env, etc.
        },
      }

      mockedAxios.mockResolvedValueOnce({
        status: 200,
        headers: { 'docker-content-digest': 'sha256:digest123' },
        data: mockManifest,
      })

      mockedAxios.mockResolvedValueOnce({
        status: 200,
        data: partialConfig,
      })

      // Mock registry lookup
      jest.spyOn(repository, 'find').mockResolvedValueOnce([mockRegistry])

      const result = await service.getImageDetails('harbor.test.com/testproject/myimage:latest', 'test-org')

      expect(result).toEqual({
        digest: 'sha256:digest123',
        sizeGB: 6000000 / (1024 * 1024 * 1024), // ~0.0056 GB
        entrypoint: ['/entrypoint.sh'],
        cmd: [],
        env: [],
        workingDir: undefined,
        user: undefined,
      })
    })

    it('should throw error when manifest request fails', async () => {
      mockedAxios.mockResolvedValueOnce({
        status: 404,
        statusText: 'Not Found',
      })

      // Mock registry lookup (should fail for image without registry hostname)
      jest.spyOn(repository, 'find').mockResolvedValueOnce([])

      await expect(service.getImageDetails('testproject/nonexistent:latest')).rejects.toThrow(
        'No registry found for image testproject/nonexistent:latest',
      )
    })

    it('should throw error when digest is missing from headers', async () => {
      mockedAxios.mockResolvedValueOnce({
        status: 200,
        headers: {}, // No digest header
        data: mockManifest,
      })

      // Mock registry lookup (should fail for image without registry hostname)
      jest.spyOn(repository, 'find').mockResolvedValueOnce([])

      await expect(service.getImageDetails('testproject/myimage:latest')).rejects.toThrow(
        'No registry found for image testproject/myimage:latest',
      )
    })

    it('should throw error when axios throws an exception', async () => {
      mockedAxios.mockRejectedValueOnce(new Error('Network error'))

      // Mock registry lookup (should fail for image without registry hostname)
      jest.spyOn(repository, 'find').mockResolvedValueOnce([])

      await expect(service.getImageDetails('testproject/myimage:latest')).rejects.toThrow(
        'No registry found for image testproject/myimage:latest',
      )
    })

    it('should handle complex image paths correctly', async () => {
      mockedAxios.mockResolvedValueOnce({
        status: 200,
        headers: { 'docker-content-digest': 'sha256:digest123' },
        data: mockManifest,
      })

      mockedAxios.mockResolvedValueOnce({
        status: 200,
        data: mockConfig,
      })

      // Mock registry lookup
      jest.spyOn(repository, 'find').mockResolvedValueOnce([mockRegistry])

      // Test with complex nested path
      await service.getImageDetails('harbor.test.com/namespace/subproject/deeply/nested/image:tag', 'test-org')

      expect(mockedAxios).toHaveBeenNthCalledWith(
        1,
        expect.objectContaining({
          url: 'https://harbor.test.com/v2/namespace/subproject/deeply/nested/image/manifests/tag',
        }),
      )
    })

    it('should handle different tag formats', async () => {
      mockedAxios.mockResolvedValueOnce({
        status: 200,
        headers: { 'docker-content-digest': 'sha256:digest123' },
        data: mockManifest,
      })

      mockedAxios.mockResolvedValueOnce({
        status: 200,
        data: mockConfig,
      })

      // Mock registry lookup
      jest.spyOn(repository, 'find').mockResolvedValueOnce([mockRegistry])

      // Test with semantic version tag
      await service.getImageDetails('harbor.test.com/testproject/myimage:v1.2.3-alpha.1', 'test-org')

      expect(mockedAxios).toHaveBeenNthCalledWith(
        1,
        expect.objectContaining({
          url: 'https://harbor.test.com/v2/testproject/myimage/manifests/v1.2.3-alpha.1',
        }),
      )
    })
  })
})
