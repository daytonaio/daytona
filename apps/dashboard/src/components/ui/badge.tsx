/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Slot } from '@radix-ui/react-slot'
import { cva, type VariantProps } from 'class-variance-authority'
import * as React from 'react'

import { cn } from '@/lib/utils'

const badgeVariants = cva(
  'inline-flex items-center rounded-full border px-2.5 py-0.5 text-xs font-medium transition-colors focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring focus-visible:ring-offset-1',
  {
    variants: {
      variant: {
        info: 'bg-info-background text-info-foreground border-info-separator',
        warning: 'bg-warning-background text-warning-foreground border-warning-separator',
        default: 'border-transparent bg-primary text-primary-foreground hover:bg-primary/80',
        secondary: 'border bg-secondary text-secondary-foreground hover:bg-secondary/80',
        destructive:
          'border-destructive-separator bg-destructive-background text-destructive-foreground hover:bg-destructive-background/80',
        success:
          'border-success-separator bg-success-background text-success-foreground hover:bg-success-background/80',
        outline: 'text-foreground',
      },
    },
    defaultVariants: {
      variant: 'default',
    },
  },
)

export interface BadgeProps extends React.ComponentProps<'div'>, VariantProps<typeof badgeVariants> {
  asChild?: boolean
}

function Badge({ className, variant, asChild = false, ...props }: BadgeProps) {
  const Comp = asChild ? Slot : 'div'

  return <Comp data-slot="badge" className={cn(badgeVariants({ variant }), className)} {...props} />
}

export { Badge, badgeVariants }
