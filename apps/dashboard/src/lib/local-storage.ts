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

/**
 * Remove localStorage items by key prefix
 * Useful for cleaning up related items
 */
export const removeLocalStorageItemsByPrefix = (prefix: string): void => {
  try {
    const keysToRemove: string[] = []
    for (let i = 0; i < localStorage.length; i++) {
      const key = localStorage.key(i)
      if (key && key.startsWith(prefix)) {
        keysToRemove.push(key)
      }
    }
    keysToRemove.forEach((key) => localStorage.removeItem(key))
  } catch (error) {
    console.error(`Failed to remove localStorage items with prefix ${prefix}:`, error)
  }
}
