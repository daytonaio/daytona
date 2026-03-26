/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { useBanner } from '@/components/Banner'
import { RoutePath } from '@/enums/RoutePath'
import { Organization } from '@daytonaio/api-client'
import { addHours, formatDistanceToNow } from 'date-fns'
import { CreditCardIcon, MailIcon } from 'lucide-react'
import { useEffect, useRef } from 'react'
import { useLocation, useNavigate } from 'react-router-dom'

const SUSPENSION_BANNER_ID = 'suspension-banner'

// todo: enumerate reasons
const PAYMENT_METHOD_REQUIRED_REASON = 'Payment method required'
const VERIFY_EMAIL_REASON = 'Please verify your email address'
const CREDITS_DEPLETED_REASON = 'Credits depleted'

function isSetupRequiredSuspension(reason: string) {
  return reason === PAYMENT_METHOD_REQUIRED_REASON || reason === VERIFY_EMAIL_REASON
}

function isCreditsDepletionSuspension(reason: string) {
  return reason === CREDITS_DEPLETED_REASON
}

type Suspension = Pick<
  Organization,
  'suspended' | 'suspensionReason' | 'suspendedAt' | 'suspensionCleanupGracePeriodHours'
>

export function useSuspensionBanner(suspension?: Suspension | null) {
  const { addBanner, removeBanner } = useBanner()
  const navigate = useNavigate()
  const location = useLocation()
  const path = location?.pathname
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

    if (isSetupRequiredSuspension(reason)) {
      if (reason === PAYMENT_METHOD_REQUIRED_REASON) {
        addBanner({
          id: SUSPENSION_BANNER_ID,
          variant: 'info',
          title: 'Setup Required',
          description: 'Add a payment method to start creating sandboxes.',
          icon: <CreditCardIcon className="h-4 w-4 flex-shrink-0 text-current" />,
          action:
            path !== RoutePath.BILLING_WALLET
              ? {
                  label: 'Go to Billing',
                  onClick: () => navigate(RoutePath.BILLING_WALLET),
                }
              : undefined,
          isDismissible: false,
        })
      } else if (reason === VERIFY_EMAIL_REASON) {
        addBanner({
          id: SUSPENSION_BANNER_ID,
          variant: 'info',
          title: 'Verification Required',
          description: 'Please verify your email address to access all features.',
          icon: <MailIcon className="h-4 w-4 flex-shrink-0 text-current" />,
          isDismissible: false,
        })
      }
      return
    }

    if (isCreditsDepletionSuspension(reason)) {
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
        title: 'Credits depleted',
        description: cleanupText,
        action:
          path !== RoutePath.BILLING_WALLET
            ? {
                label: 'Go to Billing',
                onClick: () => navigate(RoutePath.BILLING_WALLET),
              }
            : undefined,
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
  }, [suspension, addBanner, removeBanner, navigate, path])
}
