/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { OrganizationEmail } from '@/billing-api'
import { Table } from '@tanstack/react-table'

export interface OrganizationEmailsTableProps {
  data: OrganizationEmail[]
  loading: boolean
  handleDelete: (email: string) => void
  handleResendVerification: (email: string) => void
  handleAddEmail: (email: string) => void
  onRowClick?: (email: OrganizationEmail) => void
}

export interface OrganizationEmailsTableActionsProps {
  email: OrganizationEmail
  isLoading: boolean
  onDelete: (email: string) => void
  onResendVerification: (email: string) => void
}

export interface OrganizationEmailsTableHeaderProps {
  table: Table<OrganizationEmail>
  onAddEmail: (email: string) => void
}

export interface FacetedFilterOption {
  label: string
  value: string
  icon?: any
}
