/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { useTheme } from '@/contexts/ThemeContext'
import { cn } from '@/lib/utils'
import { CheckCircleIcon, InfoIcon, WarningIcon, XCircleIcon } from '@phosphor-icons/react'
import { DismissableLayerBranch } from '@radix-ui/react-dismissable-layer'
import { Toaster as Sonner } from 'sonner'

type ToasterProps = React.ComponentProps<typeof Sonner>

const Toaster = ({ className, toastOptions, ...props }: ToasterProps) => {
  const { theme } = useTheme()

  return (
    // this is needed to prevent sheet and dialog from closing when a toast is open. when migrating to base-ui, we'll migrate the toast too and then we can remove this
    <DismissableLayerBranch asChild>
      <div className="contents">
        <Sonner
          theme={theme as ToasterProps['theme']}
          className={cn('toaster group pointer-events-auto', className)}
          icons={{
            success: <CheckCircleIcon weight="fill" className="size-4 text-success" />,
            error: <XCircleIcon weight="fill" className="size-4 text-destructive" />,
            warning: <WarningIcon weight="fill" className="size-4 text-warning" />,
            info: <InfoIcon weight="fill" className="size-4 text-foreground" />,
          }}
          toastOptions={{
            ...toastOptions,
            classNames: {
              ...toastOptions?.classNames,
              toast: cn(
                'group toast pointer-events-auto !items-start group-[.toaster]:border group-[.toaster]:border-border group-[.toaster]:bg-background group-[.toaster]:text-foreground group-[.toaster]:shadow-lg',
                toastOptions?.classNames?.toast,
              ),
              icon: cn('mt-0.5', toastOptions?.classNames?.icon),
              content: cn('gap-1', toastOptions?.classNames?.content),
              title: cn(
                'font-medium group-data-[type=error]:text-destructive group-data-[type=success]:text-success group-data-[type=warning]:text-warning group-data-[type=info]:text-foreground',
                toastOptions?.classNames?.title,
              ),
              description: cn('whitespace-pre-line !text-muted-foreground', toastOptions?.classNames?.description),
              actionButton: cn(
                'group-[.toast]:bg-primary group-[.toast]:text-primary-foreground',
                toastOptions?.classNames?.actionButton,
              ),
              cancelButton: cn(
                'group-[.toast]:bg-muted group-[.toast]:text-muted-foreground',
                toastOptions?.classNames?.cancelButton,
              ),
              closeButton: cn(
                'pointer-events-auto !left-auto !right-0 !border-border !bg-background !text-muted-foreground !shadow-none ![transform:translate(35%,-35%)] hover:!border-border hover:!bg-accent hover:!text-foreground',
                toastOptions?.classNames?.closeButton,
              ),
            },
          }}
          closeButton
          {...props}
        />
      </div>
    </DismissableLayerBranch>
  )
}

export { Toaster }
