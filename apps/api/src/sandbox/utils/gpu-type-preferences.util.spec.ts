/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { BadRequestError } from '../../exceptions/bad-request.exception'
import { GpuType } from '../enums/gpu-type.enum'
import { resolveGpuTypePreferences } from './gpu-type-preferences.util'

describe('resolveGpuTypePreferences', () => {
  describe('non-GPU requests', () => {
    it.each([
      [0, undefined, null],
      [0, [GpuType.H100], null],
      [0, undefined, [GpuType.H100]],
      [0, [GpuType.H100], []],
    ])('returns undefined when gpu=%p regardless of inputs', (gpu, prefs, allowed) => {
      expect(resolveGpuTypePreferences(gpu, prefs, allowed)).toBeUndefined()
    })
  })

  describe('null allowlist (no restriction)', () => {
    it('returns undefined when no preferences are given', () => {
      expect(resolveGpuTypePreferences(1, undefined, null)).toBeUndefined()
    })

    it('treats empty preferences as undefined', () => {
      expect(resolveGpuTypePreferences(1, [], null)).toBeUndefined()
    })

    it('returns preferences verbatim when provided', () => {
      expect(resolveGpuTypePreferences(1, [GpuType.H100], null)).toEqual([GpuType.H100])
    })

    it('treats undefined allowlist as null', () => {
      expect(resolveGpuTypePreferences(1, [GpuType.H100], undefined)).toEqual([GpuType.H100])
    })
  })

  describe('empty allowlist (block all)', () => {
    it('throws when no preferences are given', () => {
      expect(() => resolveGpuTypePreferences(1, undefined, [])).toThrow(BadRequestError)
      expect(() => resolveGpuTypePreferences(1, undefined, [])).toThrow('No GPU types are allowed in this region.')
    })

    it('throws when any specific type is requested', () => {
      expect(() => resolveGpuTypePreferences(1, [GpuType.H100], [])).toThrow('No GPU types are allowed in this region.')
    })
  })

  describe('non-empty allowlist with no preferences (untyped request)', () => {
    it('narrows to the allowlist', () => {
      expect(resolveGpuTypePreferences(1, undefined, [GpuType.H100])).toEqual([GpuType.H100])
    })

    it('returns full allowlist when it has multiple types', () => {
      expect(resolveGpuTypePreferences(1, undefined, [GpuType.H100, GpuType.RTX_PRO_6000])).toEqual([
        GpuType.H100,
        GpuType.RTX_PRO_6000,
      ])
    })
  })

  describe('non-empty allowlist with preferences (intersection)', () => {
    it('returns intersection when overlapping', () => {
      expect(resolveGpuTypePreferences(1, [GpuType.H100, GpuType.RTX_PRO_6000], [GpuType.H100])).toEqual([GpuType.H100])
    })

    it('preserves preference order in the intersection', () => {
      expect(
        resolveGpuTypePreferences(1, [GpuType.RTX_PRO_6000, GpuType.H100], [GpuType.H100, GpuType.RTX_PRO_6000]),
      ).toEqual([GpuType.RTX_PRO_6000, GpuType.H100])
    })

    it('throws when intersection is empty', () => {
      expect(() => resolveGpuTypePreferences(1, [GpuType.RTX_PRO_6000], [GpuType.H100])).toThrow(BadRequestError)
      expect(() => resolveGpuTypePreferences(1, [GpuType.RTX_PRO_6000], [GpuType.H100])).toThrow(
        'Requested GPU type(s) RTX-PRO-6000 not permitted in this region. Allowed: H100.',
      )
    })
  })
})
