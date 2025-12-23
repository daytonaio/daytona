/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */
'use client'

import { Command as CommandPrimitive } from 'cmdk'
import { Check, SearchIcon } from 'lucide-react'
import * as React from 'react'

import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { cn } from '@/lib/utils'

function Command({ className, ...props }: React.ComponentProps<typeof CommandPrimitive>) {
  return (
    <CommandPrimitive
      data-slot="command"
      className={cn(
        'bg-popover text-popover-foreground flex h-full w-full flex-col overflow-hidden rounded-md',
        className,
      )}
      {...props}
    />
  )
}

function CommandDialog({
  title = 'Command Palette',
  description = 'Search for a command to run...',
  children,
  className,
  showCloseButton = false,
  overlay,
  ref,
  ...props
}: React.ComponentProps<typeof Dialog> & {
  title?: string
  overlay?: React.ReactNode
  description?: string
  className?: string
  showCloseButton?: boolean
  ref?: React.RefObject<HTMLDivElement>
}) {
  return (
    <Dialog {...props}>
      <DialogHeader className="sr-only">
        <DialogTitle>{title}</DialogTitle>
        <DialogDescription>{description}</DialogDescription>
      </DialogHeader>
      <DialogContent
        className={cn('overflow-hidden p-0 [&_[data-slot=dialog-close]]:hidden', className)}
        overlay={overlay}
        ref={ref}
      >
        {children}
      </DialogContent>
    </Dialog>
  )
}

function CommandInputButton({ className, ...props }: React.ComponentProps<'button'>) {
  return <button className={cn('text-sm text-muted-foreground hover:text-foreground px-2', className)} {...props} />
}

const CommandInput = React.forwardRef<
  React.ComponentRef<typeof CommandPrimitive.Input>,
  React.ComponentProps<typeof CommandPrimitive.Input> & { icon?: React.ReactNode }
>(({ className, children, icon, ...props }, ref) => {
  return (
    <div data-slot="command-input-wrapper" className="flex items-center gap-2 border-b px-3">
      {icon !== null ? icon || <SearchIcon className="size-4 shrink-0 opacity-50" /> : null}
      <CommandPrimitive.Input
        ref={ref}
        data-slot="command-input"
        className={cn(
          'placeholder:text-muted-foreground flex h-10 w-full rounded-md bg-transparent py-3 text-sm outline-none disabled:cursor-not-allowed disabled:opacity-50',
          className,
        )}
        {...props}
      />
      {children}
    </div>
  )
})
CommandInput.displayName = 'CommandInput'

function CommandList({ className, ...props }: React.ComponentProps<typeof CommandPrimitive.List>) {
  return (
    <CommandPrimitive.List
      data-slot="command-list"
      className={cn(
        'max-h-[300px] scroll-py-1 overflow-x-hidden overflow-y-auto scrollbar-thin scrollbar-thumb-border scrollbar-track-background',
        className,
      )}
      {...props}
    />
  )
}

function CommandEmpty({ className, ...props }: React.ComponentProps<typeof CommandPrimitive.Empty>) {
  return (
    <CommandPrimitive.Empty
      data-slot="command-empty"
      className={cn('py-6 text-center text-sm', className)}
      {...props}
    />
  )
}

function CommandGroup({ className, ...props }: React.ComponentProps<typeof CommandPrimitive.Group>) {
  return (
    <CommandPrimitive.Group
      data-slot="command-group"
      className={cn(
        'text-foreground [&_[cmdk-group-heading]]:text-muted-foreground overflow-hidden p-1 [&_[cmdk-group-heading]]:px-2 [&_[cmdk-group-heading]]:py-1.5 [&_[cmdk-group-heading]]:text-xs [&_[cmdk-group-heading]]:font-medium',
        className,
      )}
      {...props}
    />
  )
}

function CommandSeparator({ className, ...props }: React.ComponentProps<typeof CommandPrimitive.Separator>) {
  return (
    <CommandPrimitive.Separator
      data-slot="command-separator"
      className={cn('bg-border -mx-1 h-px', className)}
      {...props}
    />
  )
}

function CommandItem({ className, ...props }: React.ComponentProps<typeof CommandPrimitive.Item>) {
  return (
    <CommandPrimitive.Item
      data-slot="command-item"
      className={cn(
        "data-[selected=true]:bg-accent data-[selected=true]:text-accent-foreground [&_svg:not([class*='text-'])]:text-muted-foreground relative flex cursor-default items-center gap-2 rounded-sm px-2 py-1.5 text-sm outline-hidden select-none data-[disabled=true]:pointer-events-none data-[disabled=true]:opacity-50 [&_svg]:pointer-events-none [&_svg]:shrink-0 [&_svg:not([class*='size-'])]:size-4",
        className,
      )}
      {...props}
    />
  )
}

function CommandShortcut({ className, ...props }: React.ComponentProps<'span'>) {
  return (
    <span
      data-slot="command-shortcut"
      className={cn('text-muted-foreground ml-auto text-xs tracking-widest', className)}
      {...props}
    />
  )
}

function CommandCheckboxItem({
  className,
  children,
  checked,
  ...props
}: React.ComponentProps<typeof CommandPrimitive.Item> & { checked: boolean }) {
  return (
    <CommandItem {...props}>
      <div className="flex items-center">
        <div className="mr-2 flex h-4 w-4 items-center justify-center rounded-sm border border-primary">
          <Check className="h-4 w-4 " />
        </div>
      </div>
    </CommandItem>
  )
}

export {
  Command,
  CommandCheckboxItem,
  CommandDialog,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandInputButton,
  CommandItem,
  CommandList,
  CommandSeparator,
  CommandShortcut,
}
