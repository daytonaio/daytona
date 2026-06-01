/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Action, toast } from 'sonner'

interface HandleApiErrorOptions {
  action?: React.ReactNode | Action
  toastId?: string
}

const fallbackErrorDescription = 'Please try again or check the console for more details'

function toMessage(value: unknown): string | undefined {
  if (typeof value === 'string') {
    return value
  }

  if (Array.isArray(value)) {
    const messages = value.map(toMessage).filter(Boolean)
    return messages.length > 0 ? messages.join('\n') : undefined
  }

  return undefined
}

function getResponseData(error: unknown): unknown {
  if (typeof error !== 'object' || error === null || !('response' in error)) {
    return undefined
  }

  return (error as { response?: { data?: unknown } }).response?.data
}

function getErrorDescription(error: unknown): string {
  const responseData = getResponseData(error)

  if (typeof responseData === 'object' && responseData !== null && 'message' in responseData) {
    const message = toMessage(responseData.message)
    if (message) {
      return message
    }
  }

  const responseMessage = toMessage(responseData)
  if (responseMessage) {
    return responseMessage
  }

  return error instanceof Error && error.message ? error.message : fallbackErrorDescription
}

export function handleApiError(error: unknown, message: string, options?: HandleApiErrorOptions) {
  const description = getErrorDescription(error)

  toast.error(message, {
    ...(options?.toastId ? { id: options.toastId } : {}),
    description,
    action: options?.action,
  })

  if (!(error instanceof Error)) {
    console.error(message, error)
  }
}
