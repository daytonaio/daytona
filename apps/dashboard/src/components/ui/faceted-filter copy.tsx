/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { PlusCircle, X } from 'lucide-react'
import { AnimatePresence, motion, useReducedMotion } from 'motion/react'
import type { ReactNode } from 'react'
import { useRef, useState } from 'react'

import { cn } from '@/lib/utils'
import { buttonVariants } from './button'
import {
  Command,
  CommandCheckboxItem,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
  CommandSeparator,
} from './command'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuRadioGroup,
  DropdownMenuRadioItem,
  DropdownMenuTrigger,
} from './dropdown-menu'
import { Popover, PopoverAnchor, PopoverContent, PopoverTrigger } from './popover'

const defaultIcon = <PlusCircle />
const segmentTransition = {
  layout: { duration: 0.16 },
  opacity: { duration: 0.12 },
  x: { duration: 0.12 },
  filter: { duration: 0.12 },
}
const surfaceMotion = {
  initial: { opacity: 0 },
  animate: { opacity: 1 },
  exit: { opacity: 0 },
}
const segmentMotion = {
  initial: { opacity: 0, x: -4, filter: 'blur(2px)' },
  animate: { opacity: 1, x: 0, filter: 'blur(0px)' },
  exit: { opacity: 0, x: -4, filter: 'blur(2px)' },
}
const clearSegmentMotion = {
  initial: { opacity: 0, x: -4, filter: 'blur(2px)' },
  animate: {
    opacity: 1,
    x: 0,
    filter: 'blur(0px)',
    transition: {
      opacity: { duration: 0.18, delay: 0.08 },
      x: { duration: 0.18, delay: 0.08 },
      filter: { duration: 0.18, delay: 0.08 },
    },
  },
  exit: {
    opacity: 0,
    x: -4,
    filter: 'blur(2px)',
    transition: {
      opacity: { duration: 0.07 },
      x: { duration: 0.07 },
      filter: { duration: 0.07 },
    },
  },
}
const reducedMotionAnimation = { initial: false } as const

function getDelayedSegmentTransition(delay: number) {
  return {
    layout: segmentTransition.layout,
    opacity: { ...segmentTransition.opacity, delay },
    x: { ...segmentTransition.x, delay },
    filter: { ...segmentTransition.filter, delay },
  }
}

export type FacetedFilterOption = {
  label: ReactNode
  value: string
  icon?: ReactNode
}

export type FacetedFilterOperator = {
  label: ReactNode
  value: string
}

const defaultOperators = [{ label: 'is', value: 'is' }] satisfies readonly FacetedFilterOperator[]

export interface FacetedFilterProps {
  title: string
  options: readonly FacetedFilterOption[]
  values: ReadonlySet<string>
  onValuesChange: (values: Set<string>) => void
  operator?: string
  operators?: readonly FacetedFilterOperator[]
  onOperatorChange?: (operator: string) => void
  facets?: ReadonlyMap<string, number>
  maxValues?: number
  className?: string
  contentClassName?: string
  icon?: ReactNode
}

function pluralizeFilterTitle(title: string | undefined, count: number) {
  const label = title?.trim().toLowerCase() || 'value'

  if (count === 1) {
    return label
  }

  if (label.endsWith('status')) {
    return `${label}es`
  }

  if (label.endsWith('class')) {
    return `${label}es`
  }

  if (label.endsWith('y')) {
    return `${label.slice(0, -1)}ies`
  }

  if (label.endsWith('s')) {
    return label
  }

  return `${label}s`
}

function getSelectedValueItems(options: readonly FacetedFilterOption[], values: ReadonlySet<string>) {
  return Array.from(values).map((value) => ({
    value,
    label: options.find((option) => option.value === value)?.label ?? value,
  }))
}

function IconSlot({ children }: { children: ReactNode }) {
  return (
    <span className="flex size-4 shrink-0 items-center justify-center text-muted-foreground [&_svg:not([class*='size-'])]:size-4 [&_svg]:shrink-0">
      {children}
    </span>
  )
}

function shouldRenderIcon(icon: ReactNode) {
  return icon !== null && icon !== undefined && typeof icon !== 'boolean'
}

function FacetedFilterOperatorSegment({
  operator,
  operators,
  onOperatorChange,
  layout,
  transition,
  animation,
}: {
  operator: string
  operators: readonly FacetedFilterOperator[]
  onOperatorChange?: (operator: string) => void
  layout: boolean | 'position'
  transition: typeof segmentTransition
  animation: typeof segmentMotion | { initial: false }
}) {
  const selectedOperator = operators.find((option) => option.value === operator) ?? operators[0] ?? defaultOperators[0]
  const canChangeOperator = operators.length > 1 && !!onOperatorChange

  if (!canChangeOperator) {
    return (
      <motion.span
        key="operator"
        layout={layout}
        transition={transition}
        {...animation}
        className="-ml-px inline-flex cursor-default items-center border border-input bg-transparent px-3 text-muted-foreground dark:bg-input/30"
      >
        {selectedOperator.label}
      </motion.span>
    )
  }

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <motion.button
          key="operator"
          layout={layout}
          transition={transition}
          type="button"
          className={cn(
            '-ml-px inline-flex cursor-pointer items-center border border-input bg-transparent px-3 text-muted-foreground transition-colors outline-hidden hover:text-foreground focus-visible:z-10 focus-visible:border-ring focus-visible:ring-ring/50 focus-visible:ring-[3px] dark:bg-input/30',
            'hover:bg-accent/60 dark:hover:bg-accent/40',
          )}
          aria-label="Change filter operator"
          {...animation}
        >
          {selectedOperator.label}
        </motion.button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="start" className="min-w-24">
        <DropdownMenuRadioGroup value={selectedOperator.value} onValueChange={onOperatorChange}>
          {operators.map((option) => (
            <DropdownMenuRadioItem key={option.value} value={option.value}>
              {option.label}
            </DropdownMenuRadioItem>
          ))}
        </DropdownMenuRadioGroup>
      </DropdownMenuContent>
    </DropdownMenu>
  )
}

export function FacetedFilter({
  title,
  options,
  values,
  onValuesChange,
  operator,
  operators = defaultOperators,
  onOperatorChange,
  facets,
  maxValues = 2,
  className,
  contentClassName,
  icon = defaultIcon,
}: FacetedFilterProps) {
  const [open, setOpen] = useState(false)
  const labelButtonRef = useRef<HTMLButtonElement | null>(null)
  const labelPointerStartedOpenRef = useRef<boolean | null>(null)
  const shouldReduceMotion = useReducedMotion()
  const layout = shouldReduceMotion ? false : true
  const contentLayout = shouldReduceMotion ? false : 'position'
  const surfaceAnimation = shouldReduceMotion ? reducedMotionAnimation : surfaceMotion
  const segmentAnimation = shouldReduceMotion ? reducedMotionAnimation : segmentMotion
  const clearSegmentAnimation = shouldReduceMotion ? reducedMotionAnimation : clearSegmentMotion
  const valueSegmentTransition = shouldReduceMotion ? segmentTransition : getDelayedSegmentTransition(0.02)
  const clearSegmentTransition = shouldReduceMotion ? segmentTransition : { layout: { duration: 0.1 } }
  const selectedCount = values.size
  const selectedValueItems = getSelectedValueItems(options, values)
  const shouldShowSummary = selectedCount > maxValues
  const hasSelectedValues = selectedCount > 0
  const hasIcon = shouldRenderIcon(icon)
  const selectedOperator = operator ?? operators[0]?.value ?? defaultOperators[0].value

  const handleClear = () => {
    onValuesChange(new Set())
  }

  const handleLabelButtonPointerDown = () => {
    labelPointerStartedOpenRef.current = open
  }

  const handleLabelButtonClick = () => {
    const wasOpenOnPointerDown = labelPointerStartedOpenRef.current ?? open
    labelPointerStartedOpenRef.current = null
    setOpen(!wasOpenOnPointerDown)
  }

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverAnchor asChild>
        <div className={cn('inline-flex h-8 min-w-0 items-stretch rounded-md text-sm', className)}>
          <motion.button
            ref={labelButtonRef}
            layout={layout}
            transition={segmentTransition}
            type="button"
            disabled={hasSelectedValues}
            className={cn(
              'inline-flex cursor-pointer items-center gap-1.5 border border-input bg-transparent font-medium text-foreground transition-colors outline-hidden focus-visible:z-10 focus-visible:border-ring focus-visible:ring-ring/50 focus-visible:ring-[3px] disabled:pointer-events-none disabled:cursor-default dark:bg-input/30',
              'hover:bg-accent/60 dark:hover:bg-accent/40',
              !hasSelectedValues && buttonVariants({ variant: 'outline', size: 'sm' }),
              'h-full',
              {
                'rounded-l-md px-3': hasSelectedValues,
                'rounded-md! border-dashed': !hasSelectedValues,
              },
            )}
            onPointerDown={handleLabelButtonPointerDown}
            onClick={handleLabelButtonClick}
            aria-label={`Filter by ${title}`}
          >
            {hasIcon && <IconSlot>{icon}</IconSlot>}
            {title}
          </motion.button>
          <AnimatePresence initial={false} mode="popLayout">
            {hasSelectedValues && (
              <FacetedFilterOperatorSegment
                operator={selectedOperator}
                operators={operators}
                onOperatorChange={onOperatorChange}
                layout={contentLayout}
                transition={segmentTransition}
                animation={segmentAnimation}
              />
            )}
          </AnimatePresence>
          <PopoverTrigger asChild className={cn({ 'rounded-none!': hasSelectedValues })}>
            <motion.button
              layout={layout}
              transition={hasSelectedValues ? valueSegmentTransition : segmentTransition}
              type="button"
              tabIndex={hasSelectedValues ? 0 : -1}
              aria-hidden={!hasSelectedValues}
              className={cn(
                'inline-flex h-full min-w-0 items-stretch overflow-hidden font-medium text-foreground transition-colors outline-hidden hover:text-accent-foreground focus-visible:z-10 focus-visible:border-ring focus-visible:ring-ring/50 focus-visible:ring-[3px]',
                'hover:bg-accent/60 dark:hover:bg-accent/40',
                {
                  '-ml-px max-w-72 cursor-pointer border border-input bg-transparent dark:bg-input/30':
                    hasSelectedValues,
                  'w-0 max-w-0 border-0 p-0 opacity-0 pointer-events-none': !hasSelectedValues,
                  'px-1': hasSelectedValues && !shouldShowSummary,
                  'px-2': hasSelectedValues && shouldShowSummary,
                },
              )}
              aria-label={`Edit ${title} filter`}
              {...(hasSelectedValues ? segmentAnimation : reducedMotionAnimation)}
            >
              <AnimatePresence initial={false} mode="popLayout">
                {hasSelectedValues ? (
                  shouldShowSummary ? (
                    <motion.span
                      key="summary"
                      layout={contentLayout}
                      transition={segmentTransition}
                      className="inline-flex items-center truncate"
                    >
                      {selectedCount} {pluralizeFilterTitle(title, selectedCount)}
                    </motion.span>
                  ) : (
                    <motion.span
                      key="values"
                      layout={contentLayout}
                      transition={segmentTransition}
                      className="flex min-w-0 items-stretch overflow-hidden"
                    >
                      <AnimatePresence initial={false} mode="popLayout">
                        {selectedValueItems.map(({ value, label }, index) => (
                          <motion.span
                            layout={contentLayout}
                            key={value}
                            transition={segmentTransition}
                            initial={shouldReduceMotion ? false : { opacity: 0, x: 4, filter: 'blur(2px)' }}
                            animate={shouldReduceMotion ? undefined : { opacity: 1, x: 0, filter: 'blur(0px)' }}
                            exit={shouldReduceMotion ? undefined : { opacity: 0, x: 4, filter: 'blur(2px)' }}
                            className={cn('inline-flex min-w-0 items-center text-foreground', {
                              'border-l border-input': index > 0,
                            })}
                          >
                            <span className="inline-flex min-w-0 max-w-40 items-center overflow-hidden whitespace-nowrap px-2">
                              {label}
                            </span>
                          </motion.span>
                        ))}
                      </AnimatePresence>
                    </motion.span>
                  )
                ) : null}
              </AnimatePresence>
            </motion.button>
          </PopoverTrigger>
          <AnimatePresence initial={false} mode="popLayout">
            {hasSelectedValues && (
              <motion.button
                key="clear"
                layout={contentLayout}
                transition={clearSegmentTransition}
                type="button"
                className={cn(
                  '-ml-px inline-flex w-8 cursor-pointer items-center justify-center rounded-r-md border border-input bg-transparent text-muted-foreground transition-colors outline-hidden hover:text-foreground focus-visible:z-10 focus-visible:border-ring focus-visible:ring-ring/50 focus-visible:ring-[3px] dark:bg-input/30',
                  'hover:bg-accent/60 dark:hover:bg-accent/40',
                )}
                onClick={handleClear}
                aria-label={`Clear ${title} filter`}
                {...clearSegmentAnimation}
              >
                <X className="size-3.5" />
              </motion.button>
            )}
          </AnimatePresence>
        </div>
      </PopoverAnchor>
      <PopoverContent
        className={cn('w-[200px] p-0', contentClassName)}
        align="start"
        onCloseAutoFocus={(event) => {
          if (hasSelectedValues) {
            return
          }

          event.preventDefault()
          labelButtonRef.current?.focus()
        }}
      >
        <Command>
          <CommandInput placeholder={title} />
          <CommandList>
            <CommandEmpty>No results found.</CommandEmpty>
            <CommandGroup>
              {options.map((option) => {
                const isSelected = values.has(option.value)
                const facetCount = facets?.get(option.value)

                return (
                  <CommandCheckboxItem
                    checked={isSelected}
                    key={option.value}
                    onSelect={() => {
                      const newValue = new Set(values)

                      if (isSelected) {
                        newValue.delete(option.value)
                      } else {
                        newValue.add(option.value)
                      }

                      onValuesChange(newValue)
                    }}
                  >
                    {shouldRenderIcon(option.icon) && (
                      <span className="mr-2">
                        <IconSlot>{option.icon}</IconSlot>
                      </span>
                    )}
                    {option.label}
                    {facetCount !== undefined && (
                      <span className="ml-auto flex h-4 shrink-0 items-center justify-end pl-2 font-mono text-xs">
                        {facetCount}
                      </span>
                    )}
                  </CommandCheckboxItem>
                )
              })}
            </CommandGroup>
            {hasSelectedValues && (
              <>
                <CommandSeparator />
                <CommandGroup>
                  <CommandItem onSelect={handleClear} className="justify-center text-center">
                    Clear filters
                  </CommandItem>
                </CommandGroup>
              </>
            )}
          </CommandList>
        </Command>
      </PopoverContent>
    </Popover>
  )
}
