/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { GpuType } from '../enums/gpu-type.enum'
import { normalizeGpuType } from './gpu-type-normalizer.util'

describe('normalizeGpuType', () => {
  describe('H100 variants', () => {
    it.each([
      ['NVIDIA H100 80GB HBM3'],
      ['NVIDIA H100 PCIe'],
      ['NVIDIA H100'],
      ['H100'],
      ['nvidia h100 nvl'],
      ['Tesla H100-SXM5-80GB'],
    ])('canonicalizes %p to GpuType.H100', (raw) => {
      expect(normalizeGpuType(raw)).toBe(GpuType.H100)
    })
  })

  describe('RTX PRO 6000 variants', () => {
    it.each([
      ['NVIDIA RTX PRO 6000 Blackwell Workstation Edition'],
      ['NVIDIA RTX PRO 6000'],
      ['RTX PRO 6000'],
      ['rtx pro 6000'],
      ['RTX  PRO  6000'],
      ['RTXPRO6000'],
    ])('canonicalizes %p to GpuType.RTX_PRO_6000', (raw) => {
      expect(normalizeGpuType(raw)).toBe(GpuType.RTX_PRO_6000)
    })
  })

  describe('unsupported GPU strings', () => {
    it.each([
      ['NVIDIA L4'],
      ['NVIDIA A100'],
      ['NVIDIA GeForce RTX 4090'],
      ['Tesla T4'],
      ['some random string'],
      ['RTX 6000'],
      ['RTX PRO 5000'],
    ])('returns null for unsupported GPU %p', (raw) => {
      expect(normalizeGpuType(raw)).toBeNull()
    })
  })

  describe('empty/nullish inputs', () => {
    it('returns null for null', () => {
      expect(normalizeGpuType(null)).toBeNull()
    })

    it('returns null for undefined', () => {
      expect(normalizeGpuType(undefined)).toBeNull()
    })

    it('returns null for empty string', () => {
      expect(normalizeGpuType('')).toBeNull()
    })
  })

  describe('first-match precedence (H100 declared before RTX_PRO_6000)', () => {
    it('returns H100 when both markers are present', () => {
      expect(normalizeGpuType('H100 RTX PRO 6000')).toBe(GpuType.H100)
    })
  })
})
