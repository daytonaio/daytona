/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { useContext } from 'react'
import { SelectedOrganizationContext } from '@/contexts/SelectedOrganizationContext'

export function useSelectedOrganization() {
  const context = useContext(SelectedOrganizationContext)

  if (!context) {
    throw new Error('useSelectedOrganization must be used within a SelectedOrganizationProvider')
  }

  return context
}
