/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Action, toast } from 'sonner'
import { DaytonaError } from '@/api/errors'

interface HandleApiErrorOptions {
  action?: React.ReactNode | Action
  toastId?: string
}

export function handleApiError(error: unknown, message: string, options?: HandleApiErrorOptions) {
  const isDaytonaError = error instanceof DaytonaError

  toast.error(message, {
    ...(options?.toastId ? { id: options.toastId } : {}),
    description: isDaytonaError ? error.message : 'Please try again or check the console for more details',
    action: options?.action,
  })

  if (!isDaytonaError) {
    console.error(message, error)
  }
}
