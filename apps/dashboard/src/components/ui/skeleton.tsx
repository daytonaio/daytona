/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { cn } from '@/lib/utils'

function Skeleton({ className, ...props }: React.HTMLAttributes<HTMLDivElement>) {
  return (
    <div
      className={cn(
        'rounded-md dark:bg-muted/70 bg-muted relative overflow-hidden isolate',
        'after:absolute after:inset-0 after:-translate-x-full after:[animation:skeleton-shimmer_2s_infinite] after:bg-muted-foreground/10 dark:after:bg-muted-foreground/10 after:[mask-image:linear-gradient(90deg,transparent,black,transparent)]',
        className,
      )}
      {...props}
    />
  )
}

export { Skeleton }
