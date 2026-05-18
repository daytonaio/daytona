/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { RoutePath } from '@/enums/RoutePath'
import { Navigate, useLocation, useParams } from 'react-router-dom'

export default function SandboxDetails() {
  const { sandboxId } = useParams<{ sandboxId: string }>()
  const { hash, search } = useLocation()

  const searchParams = new URLSearchParams(search)

  if (sandboxId) {
    searchParams.set('sandboxId', sandboxId)
  }

  const queryString = searchParams.toString()

  return <Navigate to={`${RoutePath.SANDBOXES}${queryString ? `?${queryString}` : ''}${hash}`} replace />
}
