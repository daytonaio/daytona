/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import 'reflect-metadata'
import { ArgumentMetadata, BadRequestException, ValidationPipe } from '@nestjs/common'
import { validationPipeOptions } from './validation-pipe.options'
import { CreateSnapshotDto } from '../sandbox/dto/create-snapshot.dto'
import { SandboxLabelsDto } from '../sandbox/dto/sandbox.dto'
import { UpdateOrganizationQuotaDto } from '../organization/dto/update-organization-quota.dto'
import { UpdateOrganizationRegionQuotaDto } from '../organization/dto/update-organization-region-quota.dto'
import { UpdateRegionDto } from '../region/dto/update-region.dto'
import { OtelConfigDto } from '../organization/dto/otel-config.dto'

const pipe = new ValidationPipe(validationPipeOptions)
const body = (metatype: ArgumentMetadata['metatype']): ArgumentMetadata => ({ type: 'body', metatype, data: '' })

describe('global ValidationPipe — mass-assignment prevention', () => {
  describe('rejects undeclared (server-authoritative) properties', () => {
    it.each([
      ['organizationId', { organizationId: 'b2cadaa7-78a5-48a0-8eb0-f4f0e2a1bfd7' }],
      ['id', { id: 'attacker-chosen-uuid' }],
      ['initialRunnerId', { initialRunnerId: 'arbitrary-runner' }],
      ['hideFromUsers', { hideFromUsers: true }],
    ])('rejects an injected %s on CreateSnapshotDto', async (_field, injected) => {
      await expect(
        pipe.transform({ name: 'x', imageName: 'ubuntu:22.04', ...injected }, body(CreateSnapshotDto)),
      ).rejects.toBeInstanceOf(BadRequestException)
    })

    it('accepts a clean CreateSnapshotDto and preserves declared fields', async () => {
      const out = await pipe.transform(
        { name: 'x', imageName: 'ubuntu:22.04', cpu: 2, disk: 10 },
        body(CreateSnapshotDto),
      )
      expect(out).toMatchObject({ name: 'x', imageName: 'ubuntu:22.04', cpu: 2, disk: 10 })
      expect(out).not.toHaveProperty('organizationId')
    })

    it('control: a pipe without whitelist would retain the injected field (documents why the config matters)', async () => {
      const legacyPipe = new ValidationPipe({ transform: true })
      const out = await legacyPipe.transform(
        { name: 'x', imageName: 'ubuntu:22.04', organizationId: 'victim-org' },
        body(CreateSnapshotDto),
      )
      expect(out).toHaveProperty('organizationId', 'victim-org')
    })
  })

  describe('does NOT strip legitimate fields on the newly-validated DTOs (regression guard)', () => {
    it('UpdateOrganizationQuotaDto keeps its numeric fields', async () => {
      const out = await pipe.transform(
        { maxCpuPerSandbox: 8, snapshotQuota: 100, sandboxCreateRateLimit: 5 },
        body(UpdateOrganizationQuotaDto),
      )
      expect(out).toMatchObject({ maxCpuPerSandbox: 8, snapshotQuota: 100, sandboxCreateRateLimit: 5 })
    })

    it('UpdateOrganizationRegionQuotaDto keeps numeric quota fields', async () => {
      const out = await pipe.transform(
        { totalCpuQuota: 64, maxMemoryPerSandbox: 16 },
        body(UpdateOrganizationRegionQuotaDto),
      )
      expect(out).toMatchObject({ totalCpuQuota: 64, maxMemoryPerSandbox: 16 })
    })

    it('UpdateRegionDto keeps its URL fields', async () => {
      const out = await pipe.transform(
        { proxyUrl: 'https://p', sshGatewayUrl: 'ssh://g', snapshotManagerUrl: 'https://s' },
        body(UpdateRegionDto),
      )
      expect(out).toMatchObject({ proxyUrl: 'https://p', sshGatewayUrl: 'ssh://g', snapshotManagerUrl: 'https://s' })
    })

    it('OtelConfigDto keeps endpoint and the free-form headers map', async () => {
      const out = await pipe.transform(
        { endpoint: 'http://otel:4318', headers: { 'x-api-key': 'k' } },
        body(OtelConfigDto),
      )
      expect(out).toMatchObject({ endpoint: 'http://otel:4318', headers: { 'x-api-key': 'k' } })
    })

    it('SandboxLabelsDto keeps the free-form labels map (no nested-key stripping)', async () => {
      const out = await pipe.transform({ labels: { environment: 'dev', team: 'backend' } }, body(SandboxLabelsDto))
      expect(out.labels).toEqual({ environment: 'dev', team: 'backend' })
    })
  })
})
