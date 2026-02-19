/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import React from 'react'
import { Badge } from '@/components/ui/badge'

interface SeverityBadgeProps {
  severity: string
}

export const SeverityBadge: React.FC<SeverityBadgeProps> = ({ severity }) => {
  const getSeverityVariant = (sev: string) => {
    const upperSev = sev.toUpperCase()
    switch (upperSev) {
      case 'ERROR':
      case 'FATAL':
        return 'destructive'
      case 'WARN':
      case 'WARNING':
        return 'warning'
      case 'INFO':
        return 'info'
      case 'DEBUG':
      case 'TRACE':
        return 'secondary'
      default:
        return 'outline'
    }
  }

  return <Badge variant={getSeverityVariant(severity)}>{severity}</Badge>
}
