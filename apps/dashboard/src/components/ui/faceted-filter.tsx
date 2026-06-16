/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Slot } from '@radix-ui/react-slot'
import { PlusCircle, X } from 'lucide-react'
import { AnimatePresence, motion, useReducedMotion, type MotionProps } from 'motion/react'
import {
  createContext,
  use,
  useCallback,
  useMemo,
  useRef,
  useState,
  type ComponentProps,
  type ComponentPropsWithoutRef,
  type ReactNode,
  type RefObject,
} from 'react'

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
const segmentEase = 'easeOut' as const
const segmentTransitionType = 'tween' as const
const segmentTransition = {
  layout: { type: segmentTransitionType, duration: 0.16, ease: segmentEase },
  opacity: { type: segmentTransitionType, duration: 0.12, ease: segmentEase },
  x: { type: segmentTransitionType, duration: 0.12, ease: segmentEase },
  filter: { type: segmentTransitionType, duration: 0.12, ease: segmentEase },
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
      opacity: { type: segmentTransitionType, duration: 0.18, delay: 0.08, ease: segmentEase },
      x: { type: segmentTransitionType, duration: 0.18, delay: 0.08, ease: segmentEase },
      filter: { type: segmentTransitionType, duration: 0.18, delay: 0.08, ease: segmentEase },
    },
  },
  exit: {
    opacity: 0,
    x: -4,
    filter: 'blur(2px)',
    transition: {
      opacity: { type: segmentTransitionType, duration: 0.07, ease: segmentEase },
      x: { type: segmentTransitionType, duration: 0.07, ease: segmentEase },
      filter: { type: segmentTransitionType, duration: 0.07, ease: segmentEase },
    },
  },
}

type SegmentTransition = typeof segmentTransition
type ClearSegmentTransition = SegmentTransition | { layout: typeof segmentTransition.layout }

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

export type FacetedFilterValue = {
  label: ReactNode
  value: string
}

const defaultOperators = [{ label: 'is', value: 'is' }] satisfies readonly FacetedFilterOperator[]

type FacetedFilterContextValue = {
  open: boolean
  setOpen: (open: boolean) => void
  hasValue: boolean
  title: string
  onClear?: () => void
  labelButtonRef: RefObject<HTMLButtonElement | null>
  labelPointerStartedOpenRef: RefObject<boolean | null>
  layout: boolean
  contentLayout: boolean | 'position'
  segmentAnimation: typeof segmentMotion | { initial: false }
  clearSegmentAnimation: typeof clearSegmentMotion | { initial: false }
  valueSegmentTransition: SegmentTransition
  clearSegmentTransition: ClearSegmentTransition
  shouldReduceMotion: boolean
}

const FacetedFilterContext = createContext<FacetedFilterContextValue | null>(null)

type FacetedFilterButtonProps = Omit<ComponentPropsWithoutRef<'button'>, keyof MotionProps> & {
  children?: ReactNode
}

type FacetedFilterSpanProps = Omit<ComponentPropsWithoutRef<'span'>, keyof MotionProps> & {
  children?: ReactNode
}

function useFacetedFilterContext(component: string) {
  const context = use(FacetedFilterContext)

  if (!context) {
    throw new Error(`${component} must be used inside FacetedFilterRoot`)
  }

  return context
}

function composeEventHandlers<E extends { defaultPrevented: boolean }>(
  eventHandler: ((event: E) => void) | undefined,
  ourHandler: (event: E) => void,
) {
  return (event: E) => {
    eventHandler?.(event)

    if (!event.defaultPrevented) {
      ourHandler(event)
    }
  }
}

function shouldRenderIcon(icon: ReactNode) {
  return icon !== null && icon !== undefined && typeof icon !== 'boolean'
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

function FacetedFilterIcon({ children, className, ...props }: ComponentPropsWithoutRef<'span'>) {
  return (
    <span
      data-slot="faceted-filter-icon"
      className={cn(
        "flex size-4 shrink-0 items-center justify-center text-muted-foreground [&_svg:not([class*='size-'])]:size-4 [&_svg]:shrink-0",
        className,
      )}
      {...props}
    >
      {children}
    </span>
  )
}

interface FacetedFilterRootProps extends Omit<ComponentProps<typeof Popover>, 'open' | 'defaultOpen' | 'onOpenChange'> {
  open?: boolean
  defaultOpen?: boolean
  onOpenChange?: (open: boolean) => void
  hasValue?: boolean
  title?: string
  onClear?: () => void
}

function FacetedFilterRoot({
  open: openProp,
  defaultOpen = false,
  onOpenChange,
  hasValue = false,
  title = 'Filter',
  onClear,
  children,
  ...props
}: FacetedFilterRootProps) {
  const [uncontrolledOpen, setUncontrolledOpen] = useState(defaultOpen)
  const labelButtonRef = useRef<HTMLButtonElement | null>(null)
  const labelPointerStartedOpenRef = useRef<boolean | null>(null)
  const shouldReduceMotion = useReducedMotion()
  const open = openProp ?? uncontrolledOpen

  const setOpen = useCallback(
    (nextOpen: boolean) => {
      if (openProp === undefined) {
        setUncontrolledOpen(nextOpen)
      }

      onOpenChange?.(nextOpen)
    },
    [onOpenChange, openProp],
  )

  const contextValue = useMemo<FacetedFilterContextValue>(
    () => ({
      open,
      setOpen,
      hasValue,
      title,
      onClear,
      labelButtonRef,
      labelPointerStartedOpenRef,
      layout: shouldReduceMotion ? false : true,
      contentLayout: shouldReduceMotion ? false : 'position',
      segmentAnimation: shouldReduceMotion ? { initial: false } : segmentMotion,
      clearSegmentAnimation: shouldReduceMotion ? { initial: false } : clearSegmentMotion,
      valueSegmentTransition: shouldReduceMotion ? segmentTransition : getDelayedSegmentTransition(0.02),
      clearSegmentTransition: shouldReduceMotion
        ? segmentTransition
        : { layout: { type: segmentTransitionType, duration: 0.1, ease: segmentEase } },
      shouldReduceMotion: !!shouldReduceMotion,
    }),
    [hasValue, onClear, open, setOpen, shouldReduceMotion, title],
  )

  return (
    <FacetedFilterContext.Provider value={contextValue}>
      <Popover open={open} onOpenChange={setOpen} {...props}>
        {children}
      </Popover>
    </FacetedFilterContext.Provider>
  )
}

function FacetedFilterAnchor({ className, asChild = false, ...props }: ComponentProps<'div'> & { asChild?: boolean }) {
  const Comp = asChild ? Slot : 'div'

  return (
    <PopoverAnchor asChild>
      <Comp
        data-slot="faceted-filter-anchor"
        className={cn('inline-flex h-8 min-w-0 items-stretch rounded-md text-sm', className)}
        {...props}
      />
    </PopoverAnchor>
  )
}

interface FacetedFilterLabelTriggerProps extends FacetedFilterButtonProps {
  asChild?: boolean
  icon?: ReactNode
}

function FacetedFilterLabelTrigger({
  asChild = false,
  icon = defaultIcon,
  className,
  children,
  disabled,
  onPointerDown,
  onClick,
  ...props
}: FacetedFilterLabelTriggerProps) {
  const context = useFacetedFilterContext('FacetedFilterLabelTrigger')
  const Comp = asChild ? Slot : motion.button
  const animationProps = asChild ? {} : { layout: context.layout, transition: segmentTransition }
  const hasIcon = shouldRenderIcon(icon)
  const isDisabled = disabled ?? context.hasValue

  return (
    <Comp
      {...props}
      {...animationProps}
      ref={context.labelButtonRef}
      type="button"
      disabled={isDisabled}
      data-slot="faceted-filter-label-trigger"
      className={cn(
        'inline-flex cursor-pointer items-center gap-1.5 border border-input font-medium text-foreground transition-colors outline-hidden focus-visible:z-10 focus-visible:border-ring focus-visible:ring-ring/50 focus-visible:ring-[3px] disabled:pointer-events-none disabled:cursor-default',
        !context.hasValue && buttonVariants({ variant: 'outline', size: 'sm' }),
        'bg-transparent hover:bg-accent dark:bg-input/30 dark:hover:bg-accent',
        'h-full',
        {
          'rounded-l-md px-3': context.hasValue,
          'rounded-md! border-dashed': !context.hasValue,
        },
        className,
      )}
      onPointerDown={composeEventHandlers(onPointerDown, () => {
        context.labelPointerStartedOpenRef.current = context.open
      })}
      onClick={composeEventHandlers(onClick, () => {
        const wasOpenOnPointerDown = context.labelPointerStartedOpenRef.current ?? context.open
        context.labelPointerStartedOpenRef.current = null
        context.setOpen(!wasOpenOnPointerDown)
      })}
    >
      {asChild ? (
        children
      ) : (
        <>
          {hasIcon && <FacetedFilterIcon>{icon}</FacetedFilterIcon>}
          {children}
        </>
      )}
    </Comp>
  )
}

interface FacetedFilterValueTriggerProps extends FacetedFilterButtonProps {
  asChild?: boolean
}

function FacetedFilterValueTrigger({
  asChild = false,
  className,
  children,
  tabIndex,
  ...props
}: FacetedFilterValueTriggerProps) {
  const context = useFacetedFilterContext('FacetedFilterValueTrigger')
  const Comp = asChild ? Slot : motion.button
  const animationProps = asChild
    ? {}
    : {
        layout: context.layout,
        transition: context.hasValue ? context.valueSegmentTransition : segmentTransition,
        ...(context.hasValue ? context.segmentAnimation : { initial: false }),
      }

  return (
    <PopoverTrigger asChild>
      <Comp
        {...props}
        {...animationProps}
        type="button"
        tabIndex={context.hasValue ? tabIndex : -1}
        aria-hidden={!context.hasValue}
        data-slot="faceted-filter-value-trigger"
        className={cn(
          'inline-flex h-full min-w-0 items-stretch overflow-hidden font-medium text-foreground transition-colors outline-hidden hover:text-accent-foreground focus-visible:z-10 focus-visible:border-ring focus-visible:ring-ring/50 focus-visible:ring-[3px]',
          'hover:bg-accent dark:hover:bg-accent',
          {
            '-ml-px max-w-72 cursor-pointer border border-input bg-transparent dark:bg-input/30': context.hasValue,
            'w-0 max-w-0 border-0 p-0 opacity-0 pointer-events-none': !context.hasValue,
            'rounded-none!': context.hasValue && context.onClear,
            'rounded-r-md': context.hasValue && !context.onClear,
          },
          className,
        )}
      >
        {asChild ? (
          children
        ) : (
          <AnimatePresence initial={false} mode="popLayout">
            {context.hasValue ? children : null}
          </AnimatePresence>
        )}
      </Comp>
    </PopoverTrigger>
  )
}

interface FacetedFilterSegmentProps extends FacetedFilterSpanProps {
  asChild?: boolean
}

function FacetedFilterSegment({ asChild = false, className, ...props }: FacetedFilterSegmentProps) {
  const context = useFacetedFilterContext('FacetedFilterSegment')

  if (asChild) {
    return <Slot data-slot="faceted-filter-segment" className={className} {...props} />
  }

  return (
    <motion.span
      data-slot="faceted-filter-segment"
      layout={context.contentLayout}
      transition={segmentTransition}
      {...context.segmentAnimation}
      className={className}
      {...props}
    />
  )
}

interface FacetedFilterOperatorProps extends FacetedFilterSpanProps {
  operator?: string
  operators?: readonly FacetedFilterOperator[]
  onOperatorChange?: (operator: string) => void
}

function FacetedFilterOperator({
  operator,
  operators = defaultOperators,
  onOperatorChange,
  className,
  ...props
}: FacetedFilterOperatorProps) {
  const context = useFacetedFilterContext('FacetedFilterOperator')
  const selectedOperator = operators.find((option) => option.value === operator) ?? operators[0] ?? defaultOperators[0]
  const canChangeOperator = operators.length > 1 && !!onOperatorChange

  return (
    <AnimatePresence initial={false} mode="popLayout">
      {context.hasValue &&
        (canChangeOperator ? (
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <motion.button
                key="operator"
                layout={context.contentLayout}
                transition={segmentTransition}
                type="button"
                className={cn(
                  '-ml-px inline-flex cursor-pointer items-center border border-input bg-transparent px-3 text-muted-foreground transition-colors outline-hidden hover:text-foreground focus-visible:z-10 focus-visible:border-ring focus-visible:ring-ring/50 focus-visible:ring-[3px] dark:bg-input/30',
                  'hover:bg-accent dark:hover:bg-accent',
                  className,
                )}
                aria-label="Change filter operator"
                {...context.segmentAnimation}
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
        ) : (
          <FacetedFilterSegment
            key="operator"
            className={cn(
              '-ml-px inline-flex cursor-default items-center border border-input bg-transparent px-3 text-muted-foreground dark:bg-input/30',
              className,
            )}
            {...props}
          >
            {selectedOperator.label}
          </FacetedFilterSegment>
        ))}
    </AnimatePresence>
  )
}

function FacetedFilterValueSummary({ className, ...props }: FacetedFilterSpanProps) {
  return <FacetedFilterSegment className={cn('inline-flex items-center truncate', className)} {...props} />
}

function FacetedFilterValueList({ className, children, ...props }: FacetedFilterSpanProps) {
  const context = useFacetedFilterContext('FacetedFilterValueList')

  return (
    <motion.span
      data-slot="faceted-filter-value-list"
      layout={context.contentLayout}
      transition={segmentTransition}
      className={cn('flex min-w-0 items-stretch overflow-hidden', className)}
      {...props}
    >
      <AnimatePresence initial={false} mode="popLayout">
        {children}
      </AnimatePresence>
    </motion.span>
  )
}

interface FacetedFilterValueItemProps extends FacetedFilterSpanProps {
  asChild?: boolean
}

function FacetedFilterValueItem({ asChild = false, className, children, ...props }: FacetedFilterValueItemProps) {
  const context = useFacetedFilterContext('FacetedFilterValueItem')

  if (asChild) {
    return (
      <Slot
        data-slot="faceted-filter-value-item"
        className={cn('inline-flex min-w-0 items-center text-foreground', className)}
        {...props}
      >
        {children}
      </Slot>
    )
  }

  return (
    <motion.span
      data-slot="faceted-filter-value-item"
      layout={context.contentLayout}
      transition={segmentTransition}
      initial={context.shouldReduceMotion ? false : { opacity: 0, x: 4, filter: 'blur(2px)' }}
      animate={context.shouldReduceMotion ? undefined : { opacity: 1, x: 0, filter: 'blur(0px)' }}
      exit={context.shouldReduceMotion ? undefined : { opacity: 0, x: 4, filter: 'blur(2px)' }}
      className={cn('inline-flex min-w-0 items-center text-foreground', className)}
      {...props}
    >
      {children}
    </motion.span>
  )
}

interface FacetedFilterValuesProps extends FacetedFilterSpanProps {
  items: readonly FacetedFilterValue[]
  title?: string
  maxValues?: number
}

function FacetedFilterValues({ items, title, maxValues = 2, className, ...props }: FacetedFilterValuesProps) {
  const shouldShowSummary = items.length > maxValues

  if (items.length === 0) {
    return null
  }

  if (shouldShowSummary) {
    return (
      <FacetedFilterValueSummary key="summary" className={className} {...props}>
        {items.length} {pluralizeFilterTitle(title, items.length)}
      </FacetedFilterValueSummary>
    )
  }

  return (
    <FacetedFilterValueList key="values" className={className} {...props}>
      {items.map((item, index) => (
        <FacetedFilterValueItem
          key={item.value}
          className={cn({
            'border-l border-input': index > 0,
          })}
        >
          <span className="inline-flex min-w-0 max-w-40 items-center overflow-hidden whitespace-nowrap px-2">
            {item.label}
          </span>
        </FacetedFilterValueItem>
      ))}
    </FacetedFilterValueList>
  )
}

interface FacetedFilterClearProps extends FacetedFilterButtonProps {
  asChild?: boolean
}

function FacetedFilterClear({ asChild = false, className, children, onClick, ...props }: FacetedFilterClearProps) {
  const context = useFacetedFilterContext('FacetedFilterClear')
  const Comp = asChild ? Slot : motion.button
  const animationProps = asChild
    ? {}
    : {
        layout: context.contentLayout,
        transition: context.clearSegmentTransition,
        ...context.clearSegmentAnimation,
      }
  const handleClick = composeEventHandlers(onClick, () => {
    context.onClear?.()
  })

  return (
    <AnimatePresence initial={false} mode="popLayout">
      {context.hasValue && context.onClear && (
        <Comp
          key={asChild ? undefined : 'clear'}
          {...props}
          {...animationProps}
          type="button"
          data-slot="faceted-filter-clear"
          className={cn(
            '-ml-px inline-flex w-8 cursor-pointer items-center justify-center rounded-r-md border border-input bg-transparent text-muted-foreground transition-colors outline-hidden hover:text-foreground focus-visible:z-10 focus-visible:border-ring focus-visible:ring-ring/50 focus-visible:ring-[3px] dark:bg-input/30',
            'hover:bg-accent dark:hover:bg-accent',
            className,
          )}
          onClick={handleClick}
        >
          {asChild ? children : (children ?? <X className="size-3.5" />)}
        </Comp>
      )}
    </AnimatePresence>
  )
}

function FacetedFilterContent({
  className,
  align = 'start',
  onCloseAutoFocus,
  ...props
}: ComponentProps<typeof PopoverContent>) {
  const context = useFacetedFilterContext('FacetedFilterContent')

  return (
    <PopoverContent
      data-slot="faceted-filter-content"
      className={cn('w-[200px] p-0', className)}
      align={align}
      onCloseAutoFocus={(event) => {
        onCloseAutoFocus?.(event)

        if (event.defaultPrevented || context.hasValue) {
          return
        }

        event.preventDefault()
        context.labelButtonRef.current?.focus()
      }}
      {...props}
    />
  )
}

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

function FacetedFilter({
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
  const selectedCount = values.size
  const selectedValueItems = getSelectedValueItems(options, values)
  const hasSelectedValues = selectedCount > 0

  const handleClear = () => {
    onValuesChange(new Set())
  }

  return (
    <FacetedFilterRoot title={title} hasValue={hasSelectedValues} onClear={handleClear}>
      <FacetedFilterAnchor className={className}>
        <FacetedFilterLabelTrigger icon={icon} aria-label={`Filter by ${title}`}>
          {title}
        </FacetedFilterLabelTrigger>
        <FacetedFilterOperator operator={operator} operators={operators} onOperatorChange={onOperatorChange} />
        <FacetedFilterValueTrigger
          className={cn({
            'px-1': hasSelectedValues && selectedCount <= maxValues,
            'px-2': hasSelectedValues && selectedCount > maxValues,
          })}
          aria-label={`Edit ${title} filter`}
        >
          <FacetedFilterValues title={title} items={selectedValueItems} maxValues={maxValues} />
        </FacetedFilterValueTrigger>
        <FacetedFilterClear aria-label={`Clear ${title} filter`} />
      </FacetedFilterAnchor>
      <FacetedFilterContent className={contentClassName}>
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
                        <FacetedFilterIcon>{option.icon}</FacetedFilterIcon>
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
      </FacetedFilterContent>
    </FacetedFilterRoot>
  )
}

export {
  FacetedFilter,
  FacetedFilterAnchor,
  FacetedFilterClear,
  FacetedFilterContent,
  FacetedFilterIcon,
  FacetedFilterLabelTrigger,
  FacetedFilterOperator,
  FacetedFilterRoot,
  FacetedFilterSegment,
  FacetedFilterValueItem,
  FacetedFilterValueList,
  FacetedFilterValueSummary,
  FacetedFilterValueTrigger,
  FacetedFilterValues,
}
