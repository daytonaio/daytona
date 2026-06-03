/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Button } from '@/components/ui/button'
import { cn } from '@/lib/utils'
import { Plus } from 'lucide-react'
import type { ComponentProps, ReactNode } from 'react'

type CreateResourceButtonProps = Omit<ComponentProps<typeof Button>, 'asChild'> & {
  resource: ReactNode
  label?: ReactNode
}

export function CreateResourceButton({
  resource,
  children,
  className,
  label = 'Create',
  ...props
}: CreateResourceButtonProps) {
  return (
    <Button
      variant="default"
      size="sm"
      className={cn('w-8 gap-0 px-0 xs:w-auto xs:gap-1.5 xs:px-3', className)}
      {...props}
    >
      <Plus className="size-4" />
      <span className="sr-only xs:not-sr-only">{label}</span>
      <span className="hidden sm:inline">{children ?? resource}</span>
    </Button>
  )
}
