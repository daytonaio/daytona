/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

export const getLocalStorageItem = (key: string): string | null => {
  try {
    return localStorage.getItem(key)
  } catch (error) {
    console.error('Failed to read from localStorage:', error)
    return null
  }
}

export const setLocalStorageItem = (key: string, value: string): void => {
  try {
    localStorage.setItem(key, value)
  } catch (error) {
    console.error('Failed to write to localStorage:', error)
  }
}
