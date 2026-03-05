/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { useBanner } from '@/components/Banner'
import { Organization } from '@daytonaio/api-client'
import { addHours, formatDistanceToNow } from 'date-fns'
import { MailIcon } from 'lucide-react'
import { useEffect, useRef } from 'react'
import { useLocation } from 'react-router-dom'

const SUSPENSION_BANNER_ID = 'suspension-banner'

const VERIFY_EMAIL_REASON = 'Please verify your email address'

type Suspension = Pick<
  Organization,
  'suspended' | 'suspensionReason' | 'suspendedAt' | 'suspensionCleanupGracePeriodHours'
>

export function useSuspensionBanner(suspension?: Suspension | null) {
  const { addBanner, removeBanner } = useBanner()
  const location = useLocation()
  const previousSuspendedRef = useRef<boolean | undefined>(undefined)

  useEffect(() => {
    const wasSuspended = previousSuspendedRef.current
    const isSuspended = suspension?.suspended ?? false

    if (wasSuspended && !isSuspended) {
      removeBanner(SUSPENSION_BANNER_ID)
      previousSuspendedRef.current = isSuspended
      return
    }

    previousSuspendedRef.current = isSuspended

    if (!isSuspended || !suspension?.suspensionReason) {
      return
    }

    const reason = suspension.suspensionReason

    if (reason === VERIFY_EMAIL_REASON) {
      addBanner({
        id: SUSPENSION_BANNER_ID,
        variant: 'info',
        title: 'Verification Required',
        description: 'Please verify your email address to access all features.',
        icon: <MailIcon className="h-4 w-4 flex-shrink-0 text-current" />,
        isDismissible: false,
      })
      return
    }

    const suspendedAtDate = suspension.suspendedAt ? new Date(suspension.suspendedAt) : null
    const cleanupDate = suspendedAtDate
      ? addHours(suspendedAtDate, suspension.suspensionCleanupGracePeriodHours ?? 0)
      : null

    const cleanupDatePassed = cleanupDate !== null && cleanupDate <= new Date()
    const cleanupText = cleanupDate
      ? cleanupDatePassed
        ? 'Sandboxes will be stopped'
        : `Sandboxes will be stopped ${formatDistanceToNow(cleanupDate, { addSuffix: true })}`
      : 'Sandboxes will be stopped soon'

    addBanner({
      id: SUSPENSION_BANNER_ID,
      variant: 'error',
      title: 'Organization suspended',
      description: reason ? `${reason}. ${cleanupText}` : cleanupText,
      isDismissible: false,
    })
  }, [suspension, addBanner, removeBanner, location.pathname])
}
