/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { cn } from '@/lib/utils'

function Skeleton({ className, ...props }: React.HTMLAttributes<HTMLDivElement>) {
  return (
    <div
      className={cn(
        'rounded-md dark:bg-muted/60 bg-muted relative overflow-hidden isolate',
        'after:absolute after:inset-y-0 after:left-0 after:w-2/3 after:-translate-x-full after:[animation:skeleton-shimmer_1.8s_ease-in-out_infinite] after:bg-[linear-gradient(90deg,transparent_0%,color-mix(in_oklch,var(--muted-foreground)_0%,transparent)_10%,color-mix(in_oklch,var(--muted-foreground)_4%,transparent)_30%,color-mix(in_oklch,var(--muted-foreground)_9%,transparent)_50%,color-mix(in_oklch,var(--muted-foreground)_4%,transparent)_70%,color-mix(in_oklch,var(--muted-foreground)_0%,transparent)_90%,transparent_100%)] after:will-change-transform motion-reduce:after:hidden',
        className,
      )}
      {...props}
    />
  )
}

export { Skeleton }
