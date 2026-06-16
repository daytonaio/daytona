/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { RoutePath } from '@/enums/RoutePath'
import LinkifyIt from 'linkify-it'
import { type ReactNode } from 'react'
import { Link } from 'react-router'
import { Action, toast } from 'sonner'

interface HandleApiErrorOptions {
  action?: ReactNode | Action
  toastId?: string
}

const fallbackErrorDescription = 'Please try again or check the console for more details'
const linkify = new LinkifyIt({ fuzzyLink: false, fuzzyEmail: false })

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

function getInternalLinkTo(url: string): string | undefined {
  if (typeof window === 'undefined') {
    return undefined
  }

  const parsedUrl = new URL(url)
  const isDashboardPath = `${parsedUrl.pathname}/`.startsWith(`${RoutePath.DASHBOARD}/`)

  if (parsedUrl.origin !== window.location.origin || !isDashboardPath) {
    return undefined
  }

  return `${parsedUrl.pathname}${parsedUrl.search}${parsedUrl.hash}`
}

function linkifyMessage(message: string): ReactNode {
  const matches = linkify.match(message)?.filter((match) => match.schema === 'http:' || match.schema === 'https:')

  if (!matches?.length) {
    return message
  }

  const parts: ReactNode[] = []
  let lastIndex = 0

  for (const match of matches) {
    const internalLinkTo = getInternalLinkTo(match.url)

    if (match.index > lastIndex) {
      parts.push(message.slice(lastIndex, match.index))
    }

    parts.push(
      internalLinkTo ? (
        <Link
          key={match.index}
          to={internalLinkTo}
          className="font-medium underline underline-offset-2 hover:text-foreground"
        >
          {match.text}
        </Link>
      ) : (
        <a
          key={match.index}
          href={match.url}
          target="_blank"
          rel="noreferrer"
          className="font-medium underline underline-offset-2 hover:text-foreground"
        >
          {match.text}
        </a>
      ),
    )

    lastIndex = match.lastIndex
  }

  if (lastIndex < message.length) {
    parts.push(message.slice(lastIndex))
  }

  return parts
}

export function handleApiError(error: unknown, message: string, options?: HandleApiErrorOptions) {
  const description = getErrorDescription(error)

  toast.error(message, {
    ...(options?.toastId ? { id: options.toastId } : {}),
    description: linkifyMessage(description),
    action: options?.action,
  })

  if (!(error instanceof Error)) {
    console.error(message, error)
  }
}
