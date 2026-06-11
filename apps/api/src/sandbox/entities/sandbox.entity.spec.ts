/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Sandbox } from './sandbox.entity'
import { SandboxState } from '../enums/sandbox-state.enum'

describe('Sandbox.enforceInvariants', () => {
  describe('degradedReason', () => {
    it('clears degradedReason when state is not STARTED', () => {
      const sandbox = new Sandbox()
      sandbox.state = SandboxState.STOPPED
      sandbox.degradedReason = 'fd-exhaustion: too many open files'

      const changes = sandbox.enforceInvariants()

      expect(changes.degradedReason).toBeNull()
      expect(sandbox.degradedReason).toBeNull()
    })

    it('keeps degradedReason while state is STARTED', () => {
      const sandbox = new Sandbox()
      sandbox.state = SandboxState.STARTED
      sandbox.degradedReason = 'fd-exhaustion: too many open files'

      const changes = sandbox.enforceInvariants()

      expect(changes).not.toHaveProperty('degradedReason')
      expect(sandbox.degradedReason).toBe('fd-exhaustion: too many open files')
    })

    it('emits no change when degradedReason is already unset on a non-STARTED sandbox', () => {
      const sandbox = new Sandbox()
      sandbox.state = SandboxState.STOPPED

      const changes = sandbox.enforceInvariants()

      expect(changes).not.toHaveProperty('degradedReason')
    })
  })
})
