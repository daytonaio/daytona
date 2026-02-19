/**
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

/**
 * Global toast notification singleton
 * Queues toasts to prevent showing multiple at the same time
 */

type ToastVariant = 'success' | 'error' | 'warning' | 'info'

interface ToastOptions {
  title: string
  message: string
  variant?: ToastVariant
}

interface TuiShowToast {
  showToast: (options: { body: { title: string; message: string; variant: ToastVariant } }) => void
}

class ToastManager {
  private tui: TuiShowToast | null = null
  private queue: ToastOptions[] = []
  private isShowing = false

  /**
   * Initialize the toast manager with the TUI instance
   */
  initialize(tui: TuiShowToast | null | undefined): void {
    this.tui = tui || null
  }

  /**
   * Show a toast notification
   * If a toast is currently showing, this will be queued
   */
  show(options: ToastOptions): void {
    const toast: ToastOptions = {
      variant: 'info',
      ...options,
    }

    this.queue.push(toast)
    this.processQueue()
  }

  /**
   * Process the toast queue, showing one toast at a time
   */
  private processQueue(): void {
    if (this.isShowing || this.queue.length === 0) {
      return
    }

    if (!this.tui) {
      // If TUI is not available, clear the queue
      this.queue = []
      return
    }

    this.isShowing = true
    const toast = this.queue.shift()!

    try {
      this.tui.showToast({
        body: {
          title: toast.title,
          message: toast.message,
          variant: toast.variant || 'info',
        },
      })
    } catch (err) {
      // If showing fails, continue with next toast
      console.error('Failed to show toast:', err)
    }

    // Wait a bit before showing the next toast to avoid overlap
    // Most toasts are visible for 2-3 seconds, so we wait 2.5 seconds
    setTimeout(() => {
      this.isShowing = false
      this.processQueue()
    }, 2500)
  }

  /**
   * Clear all pending toasts
   */
  clear(): void {
    this.queue = []
  }
}

// Export singleton instance
export const toast = new ToastManager()
