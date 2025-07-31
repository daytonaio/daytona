/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Organization } from '@daytonaio/api-client'
import { createContext } from 'react'

export interface IOrganizationsContext {
  organizations: Organization[]
  refreshOrganizations: (selectedOrganizationId?: string) => Promise<Organization[]>
}

export const OrganizationsContext = createContext<IOrganizationsContext | undefined>(undefined)
