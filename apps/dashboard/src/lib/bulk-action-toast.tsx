/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ExternalToast, toast } from 'sonner'

type ToastId = string | number

type BulkActionToastOptions = Omit<ExternalToast, 'id'>

interface BulkActionResultOptions {
  successTitle: string
  errorTitle: string
  warningTitle: string
  canceledTitle: string
}

interface BulkActionToast {
  id: ToastId
  loading: (message: string, options?: BulkActionToastOptions) => void
  success: (message: string, options?: BulkActionToastOptions) => void
  error: (message: string, options?: BulkActionToastOptions) => void
  warning: (message: string, options?: BulkActionToastOptions) => void
  info: (message: string, options?: BulkActionToastOptions) => void
  result: (result: { successCount: number; failureCount: number }, options: BulkActionResultOptions) => void
  dismiss: () => void
}

export function createBulkActionToast(initialMessage: string, options?: BulkActionToastOptions): BulkActionToast {
  const id = toast.loading(initialMessage, {
    ...options,
  })

  return {
    id,

    loading(message: string, opts?: BulkActionToastOptions) {
      toast.loading(message, { ...opts, id })
    },

    success(message: string, opts?: BulkActionToastOptions) {
      toast.success(message, {
        action: null,
        ...opts,
        id,
      })
    },

    error(message: string, opts?: BulkActionToastOptions) {
      toast.error(message, {
        action: null,
        ...opts,
        id,
      })
    },

    warning(message: string, opts?: BulkActionToastOptions) {
      toast.warning(message, {
        action: null,
        ...opts,
        id,
      })
    },

    info(message: string, opts?: BulkActionToastOptions) {
      toast.message(message, {
        action: null,
        ...opts,
        id,
      })
    },

    result(
      { successCount, failureCount }: { successCount: number; failureCount: number },
      opts: BulkActionResultOptions,
    ) {
      const processedCount = successCount + failureCount
      const allSucceeded = processedCount > 0 && failureCount === 0
      const allFailed = processedCount > 0 && successCount === 0

      if (allSucceeded) {
        this.success(opts.successTitle)
      } else if (allFailed) {
        this.error(opts.errorTitle)
      } else if (processedCount > 0) {
        this.warning(opts.warningTitle, {
          description: `${successCount} succeeded. ${failureCount} failed.`,
        })
      } else {
        this.info(opts.canceledTitle)
      }
    },

    dismiss() {
      toast.dismiss(id)
    },
  }
}
