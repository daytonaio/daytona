/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Suspense, type ComponentProps } from 'react'
import { LoadingFallbackContent } from './LoadingFallbackContent'

function PageSuspenseFallback() {
  return (
    <div className="flex min-h-screen w-full items-center justify-center bg-background p-6">
      <LoadingFallbackContent />
    </div>
  )
}

type PageSuspenseProps = ComponentProps<typeof Suspense>

export function PageSuspense({ fallback = <PageSuspenseFallback />, ...props }: PageSuspenseProps) {
  return <Suspense fallback={fallback} {...props} />
}
