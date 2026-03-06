/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { LifeBuoyIcon } from 'lucide-react'
import { useMemo } from 'react'
import { useRegisterCommands } from '../../components/CommandPalette'
import { usePylon } from './usePylon'

export function usePylonCommands() {
  const { toggle, isEnabled } = usePylon()

  const commands = useMemo(
    () => [
      {
        id: 'pylon-support',
        label: 'Help & Support',
        icon: <LifeBuoyIcon className="w-4 h-4" />,
        keywords: ['help', 'support', 'chat', 'pylon', 'assist'],
        onSelect: toggle,
      },
    ],
    [toggle],
  )

  useRegisterCommands(isEnabled ? commands : [], { groupId: 'support', groupLabel: 'Support', groupOrder: 10 })
}
