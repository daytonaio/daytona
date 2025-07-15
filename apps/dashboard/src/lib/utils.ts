/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { clsx, type ClassValue } from 'clsx'
import { twMerge } from 'tailwind-merge'

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}

export function getRelativeTimeString(
  timestamp: string | Date | undefined | null,
  fallback = '-',
): { date: Date; relativeTimeString: string } {
  if (!timestamp) {
    return { date: new Date(), relativeTimeString: fallback }
  }

  try {
    const date = new Date(timestamp)
    const now = new Date()
    const diffInMinutes = Math.floor((now.getTime() - date.getTime()) / (1000 * 60))
    const isFuture = diffInMinutes < 0
    const absDiffInMinutes = Math.abs(diffInMinutes)

    if (absDiffInMinutes < 1)
      return {
        date,
        relativeTimeString: isFuture ? 'shortly' : 'just now',
      }

    if (absDiffInMinutes < 60) {
      return {
        date,
        relativeTimeString: isFuture ? `in ${absDiffInMinutes}m` : `${absDiffInMinutes}m ago`,
      }
    }

    const hours = Math.floor(absDiffInMinutes / 60)
    if (hours < 24) {
      return {
        date,
        relativeTimeString: isFuture ? `in ${hours}h` : `${hours}h ago`,
      }
    }

    const days = Math.floor(hours / 24)
    if (days < 365) {
      return {
        date,
        relativeTimeString: isFuture ? `in ${days}d` : `${days}d ago`,
      }
    }

    const years = Math.floor(days / 365)
    return {
      date,
      relativeTimeString: isFuture ? `in ${years}y` : `${years}y ago`,
    }
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

export function formatDuration(minutes: number): string {
  minutes = Math.abs(minutes)

  if (minutes < 60) {
    return `${Math.floor(minutes)}m`
  }

  const hours = minutes / 60
  if (hours < 24) {
    return `${Math.floor(hours)}h`
  }

  const days = hours / 24
  if (days < 365) {
    return `${Math.floor(days)}d`
  }

  const years = days / 365
  return `${Math.floor(years)}y`
}
