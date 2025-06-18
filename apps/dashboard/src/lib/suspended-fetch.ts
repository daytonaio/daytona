/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

enum PromiseStatus {
  Pending = 'pending',
  Success = 'success',
  Error = 'error',
}

export function suspendedFetch<T>(url: string): () => T {
  let status: PromiseStatus = PromiseStatus.Pending
  let result: T

  const suspend = fetch(url).then(
    (res) =>
      res.json().then(
        (data) => {
          status = PromiseStatus.Success
          result = data
        },
        (err) => {
          status = PromiseStatus.Error
          result = err
        },
      ),
    (err) => {
      status = PromiseStatus.Error
      result = err
    },
  )

  return () => {
    switch (status) {
      case PromiseStatus.Pending:
        throw suspend
      case PromiseStatus.Error:
        throw result
      case PromiseStatus.Success:
        return result
    }
  }
}

export async function minDuration<T>(callback: Promise<T>, delay = 500): Promise<T> {
  const a = await Promise.all([callback, new Promise((r) => setTimeout(r, delay))])
  return a[0]
}
