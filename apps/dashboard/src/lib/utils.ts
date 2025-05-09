/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { clsx, type ClassValue } from 'clsx'
import { twMerge } from 'tailwind-merge'

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}

export function getRelativeTimeString(timestamp?: string, fallback = '-'): { date: Date; relativeTimeString: string } {
  if (!timestamp) {
    return { date: new Date(), relativeTimeString: fallback }
  }

  try {
    const date = new Date(timestamp)
    const now = new Date()
    const diffInMinutes = Math.floor((now.getTime() - date.getTime()) / (1000 * 60))

    if (diffInMinutes < 1) return { date, relativeTimeString: 'just now' }
    if (diffInMinutes === 1) return { date, relativeTimeString: '1 minute ago' }
    if (diffInMinutes < 60) return { date, relativeTimeString: `${diffInMinutes} minutes ago` }

    const hours = Math.floor(diffInMinutes / 60)
    const minutes = diffInMinutes % 60

    if (hours === 1) {
      return minutes > 0
        ? { date, relativeTimeString: `1 hour ${minutes} minutes ago` }
        : { date, relativeTimeString: '1 hour ago' }
    }

    if (hours < 24) {
      return minutes > 0
        ? { date, relativeTimeString: `${hours} hours ${minutes} minutes ago` }
        : { date, relativeTimeString: `${hours} hours ago` }
    }

    const days = Math.floor(hours / 24)
    return { date, relativeTimeString: days === 1 ? 'yesterday' : `${days} days ago` }
  } catch (e) {
    return { date: new Date(), relativeTimeString: fallback }
  }
}

export function capitalize(value: string) {
  return value.charAt(0).toUpperCase() + value.slice(1)
}

export function getMaskedApiKey(key: string) {
  return `${key.substring(0, 3)}********************${key.slice(-3)}`
}
