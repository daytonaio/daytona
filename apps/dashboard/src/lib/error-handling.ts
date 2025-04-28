/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Action, toast } from 'sonner'
import { DaytonaError } from '@/api/errors'

export function handleApiError(error: unknown, message: string, toastAction?: React.ReactNode | Action) {
  const isDaytonaError = error instanceof DaytonaError

  toast.error(message, {
    description: isDaytonaError ? error.message : 'Please try again or check the console for more details',
    action: toastAction,
  })

  if (!isDaytonaError) {
    console.error(message, error)
  }
}
