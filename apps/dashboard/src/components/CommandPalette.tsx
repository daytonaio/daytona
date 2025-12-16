/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

'use client'

import { useDeepCompareMemo } from '@/hooks/useDeepCompareMemo'
import { cn, pluralize } from '@/lib/utils'
import { useCommandState } from 'cmdk'
import { AlertCircle, ChevronRight, Loader2 } from 'lucide-react'
import { AnimatePresence, motion, useAnimate } from 'motion/react'
import React, { createContext, useCallback, useContext, useEffect, useMemo, useRef, type ReactNode } from 'react'
import { createStore, useStore, type StoreApi } from 'zustand'
import {
  Command,
  CommandDialog,
  CommandEmpty as CommandEmptyPrimitive,
  CommandGroup as CommandGroupPrimitive,
  CommandInput,
  CommandItem as CommandItemPrimitive,
  CommandList,
} from './ui/command'
import { DialogDescription, DialogOverlay, DialogTitle } from './ui/dialog'
import { Kbd } from './ui/kbd'
import { Skeleton } from './ui/skeleton'

export type CommandConfig = {
  id: string
  label: ReactNode
  icon?: ReactNode
  loading?: boolean
  page?: string
  keywords?: string[]
  onSelect?: () => void
  disabled?: boolean
  chainable?: boolean
  value?: string
  className?: string
}

export type RegisterCommandsOptions = {
  pageId?: string
  groupId?: string
  groupLabel?: string
  groupOrder?: number
}

export type PageConfig = {
  id: string
  label?: string
  placeholder?: string
}

type CommandGroup = {
  id: string
  label?: string
  order: number
  commands: Map<string, CommandConfig>
}

type PageData = {
  meta: PageConfig
  groups: Map<string, CommandGroup>
}

type CommandPaletteState = {
  isOpen: boolean
  activePageId: string
  pageStack: string[]
  searchByPage: Map<string, string>
  shouldFilter: boolean
  barMode: 'flash' | 'pulse'
  pages: Map<string, PageData>
}

type CommandPaletteActions = {
  setIsOpen: (open: boolean) => void
  setSearch: (value: string) => void
  setShouldFilter: (value: boolean) => void
  setBarMode: (mode: 'flash' | 'pulse') => void
  pushPage: (pageId: string) => void
  popPage: () => void
  goToPage: (pageId: string) => void
  popToRoot: () => void
  registerPage: (config: PageConfig) => void
  registerCommands: (commands: CommandConfig[], options?: RegisterCommandsOptions) => () => void
  unregisterCommands: (commandIds: string[], options?: { pageId?: string; groupId?: string }) => void
}

type CommandPaletteStore = CommandPaletteState & { actions: CommandPaletteActions }

const createCommandPaletteStore = (defaultPage = 'root') => {
  return createStore<CommandPaletteStore>((set, get) => ({
    isOpen: false,
    activePageId: defaultPage,
    pageStack: [defaultPage],
    searchByPage: new Map(),
    shouldFilter: true,
    barMode: 'flash',
    pages: new Map([
      [
        defaultPage,
        {
          meta: { id: defaultPage, label: 'Home', placeholder: 'Type a command or search...' },
          groups: new Map(),
        },
      ],
    ]),

    actions: {
      setIsOpen: (isOpen) => set({ isOpen }),
      setSearch: (value) =>
        set((state) => {
          const newSearchByPage = new Map(state.searchByPage)
          newSearchByPage.set(state.activePageId, value)

          return { searchByPage: newSearchByPage }
        }),
      setShouldFilter: (value) => set({ shouldFilter: value }),
      setBarMode: (mode) => set({ barMode: mode }),

      pushPage: (pageId) =>
        set((state) => {
          if (!state.pages.has(pageId)) {
            return state
          }

          return {
            pageStack: [...state.pageStack, pageId],
            activePageId: pageId,
          }
        }),

      popPage: () =>
        set((state) => {
          if (state.pageStack.length <= 1) return state
          const newStack = state.pageStack.slice(0, -1)

          return {
            pageStack: newStack,
            activePageId: newStack[newStack.length - 1],
          }
        }),

      goToPage: (pageId) =>
        set((state) => {
          const pageIndex = state.pageStack.indexOf(pageId)
          if (pageIndex !== -1) {
            return {
              pageStack: state.pageStack.slice(0, pageIndex + 1),
              activePageId: pageId,
            }
          }

          return state
        }),

      popToRoot: () =>
        set({
          pageStack: [defaultPage],
          activePageId: defaultPage,
          searchByPage: new Map(),
        }),

      registerPage: (config) =>
        set((state) => {
          const newPages = new Map(state.pages)
          const existing = newPages.get(config.id)

          newPages.set(config.id, {
            meta: { ...existing?.meta, ...config },
            groups: existing?.groups ?? new Map(),
          })

          return { pages: newPages }
        }),

      registerCommands: (commands, options = {}) => {
        const { pageId = defaultPage, groupId = 'default', groupLabel, groupOrder = 100 } = options

        set((state) => {
          const page = state.pages.get(pageId)
          if (!page) {
            return state
          }

          const newGroups = new Map(page.groups)
          const existingGroup = newGroups.get(groupId)

          const newCommands = new Map(existingGroup?.commands ?? new Map())

          for (const cmd of commands) {
            newCommands.set(cmd.id, cmd)
          }

          newGroups.set(groupId, {
            id: groupId,
            label: groupLabel ?? existingGroup?.label,
            order: groupOrder ?? existingGroup?.order ?? 0,
            commands: newCommands,
          })

          const newPages = new Map(state.pages)

          newPages.set(pageId, { ...page, groups: newGroups })

          return { pages: newPages }
        })

        const commandIds = commands.map((c) => c.id)
        return () => get().actions.unregisterCommands(commandIds, { pageId, groupId })
      },

      unregisterCommands: (commandIds, options = {}) => {
        const { pageId = defaultPage, groupId = 'default' } = options

        set((state) => {
          const page = state.pages.get(pageId)
          if (!page) {
            return state
          }

          const group = page.groups.get(groupId)
          if (!group) {
            return state
          }

          const newCommands = new Map(group.commands)
          for (const id of commandIds) {
            newCommands.delete(id)
          }

          const newPages = new Map(state.pages)
          const newGroups = new Map(page.groups)

          if (newCommands.size === 0) {
            newGroups.delete(groupId)
          } else {
            newGroups.set(groupId, { ...group, commands: newCommands })
          }

          newPages.set(pageId, { ...page, groups: newGroups })

          return { pages: newPages }
        })
      },
    },
  }))
}

const CommandPaletteContext = createContext<StoreApi<CommandPaletteStore> | null>(null)

export function useCommandPalette<T = CommandPaletteStore>(
  selector: (state: CommandPaletteStore) => T = (state) => state as T,
): T {
  const store = useContext(CommandPaletteContext)
  if (!store) {
    throw new Error('useCommandPalette must be used within <CommandPaletteProvider />')
  }
  return useStore(store, selector)
}

export function useCommandPaletteActions() {
  const store = useContext(CommandPaletteContext)
  if (!store) {
    throw new Error('useCommandPaletteActions must be used within <CommandPaletteProvider />')
  }

  return useStore(store, (state) => state.actions)
}

export type CommandPaletteProviderProps = {
  children: ReactNode
  defaultPage?: string
  enableGlobalShortcut?: boolean
}

export function CommandPaletteProvider({
  children,
  defaultPage = 'root',
  enableGlobalShortcut = true,
}: CommandPaletteProviderProps) {
  const storeRef = useRef<StoreApi<CommandPaletteStore> | null>(null)

  if (!storeRef.current) {
    storeRef.current = createCommandPaletteStore(defaultPage)
  }

  useEffect(() => {
    if (!enableGlobalShortcut) return

    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'k' && (e.metaKey || e.ctrlKey)) {
        e.preventDefault()
        const state = storeRef.current?.getState()
        state?.actions.setIsOpen(!state.isOpen)
      }
    }

    document.addEventListener('keydown', handleKeyDown)
    return () => document.removeEventListener('keydown', handleKeyDown)
  }, [enableGlobalShortcut])

  return <CommandPaletteContext.Provider value={storeRef.current}>{children}</CommandPaletteContext.Provider>
}

export function useRegisterCommands(commands: CommandConfig[], options?: RegisterCommandsOptions) {
  const { registerCommands } = useCommandPaletteActions()

  const optionsMemo = useDeepCompareMemo(options)

  useEffect(() => {
    const unregister = registerCommands(commands, optionsMemo)
    return () => unregister()
  }, [commands, optionsMemo, registerCommands])
}

export function useRegisterPage(config: PageConfig) {
  const { registerPage } = useCommandPaletteActions()

  const configMemo = useDeepCompareMemo(config)

  useEffect(() => {
    registerPage(configMemo)
  }, [configMemo, registerPage])
}

export type CommandPaletteProps = {
  className?: string
  overlay?: ReactNode
}

export function CommandPalette({ className, overlay }: CommandPaletteProps) {
  const pages = useCommandPalette((state) => state.pages)
  const activePageId = useCommandPalette((state) => state.activePageId)
  const isOpen = useCommandPalette((state) => state.isOpen)
  const search = useCommandPalette((state) => state.searchByPage.get(state.activePageId) ?? '')
  const shouldFilter = useCommandPalette((state) => state.shouldFilter)
  const barMode = useCommandPalette((state) => state.barMode)
  const pageStack = useCommandPalette((state) => state.pageStack)

  const { setIsOpen, setSearch, popPage, popToRoot } = useCommandPaletteActions()

  const activePage = pages.get(activePageId)
  const inputRef = useRef<HTMLInputElement>(null)

  const [scope, animate] = useAnimate()

  useEffect(() => {
    if (isOpen) {
      popToRoot()
      requestAnimationFrame(() => inputRef.current?.focus())
    }
  }, [isOpen, popToRoot])

  useEffect(() => {
    if (isOpen && scope.current) {
      animate(scope.current, { scale: [0.975, 1] }, { duration: 0.3 })
    }
  }, [activePageId, isOpen, animate, scope])

  const sortedGroups = useMemo(() => {
    if (!activePage) {
      return []
    }
    return Array.from(activePage.groups.values()).sort((a, b) => a.order - b.order)
  }, [activePage])

  return (
    <CommandDialog
      open={isOpen}
      onOpenChange={setIsOpen}
      className={cn(
        'sm:max-w-xl w-full top-[calc(50%-250px)] translate-y-0 data-[state=closed]:!slide-out-to-top-4 data-[state=open]:!slide-in-from-top-4',
        'bg-transparent border-none shadow-none p-0 overflow-visible',
        className,
      )}
      overlay={overlay ?? <DialogOverlay className="bg-black/80" />}
    >
      <motion.div
        ref={scope}
        className="flex flex-col w-full h-full border border-border/50 dark:bg-popover/70 bg-popover backdrop-blur rounded-xl overflow-hidden shadow-2xl transform-gpu"
      >
        <Command
          shouldFilter={shouldFilter}
          loop
          className="bg-transparent [&_[cmdk-group-heading]]:px-2 [&_[cmdk-group-heading]]:font-medium [&_[cmdk-group-heading]]:text-muted-foreground [&_[cmdk-group]:not([hidden])_~[cmdk-group]]:pt-0 [&_[cmdk-group]]:px-2 [&_[cmdk-input-wrapper]_svg]:h-5 [&_[cmdk-input-wrapper]_svg]:w-5 [&_[cmdk-input]]:h-12 [&_[cmdk-item]]:px-2 [&_[cmdk-item]]:py-3 [&_[cmdk-item]_svg]:h-5 [&_[cmdk-item]_svg]:w-5"
        >
          <DialogTitle className="sr-only">Command Palette</DialogTitle>
          <DialogDescription className="sr-only">
            Use the command palette to navigate the application.
          </DialogDescription>

          <Breadcrumbs />

          <CommandInput
            value={search}
            onValueChange={setSearch}
            placeholder={activePage?.meta.placeholder ?? 'Type a command or search...'}
            ref={inputRef}
            onKeyDown={(e) => {
              if (e.key === 'Backspace' && !search && pageStack.length > 1) {
                e.preventDefault()
                popPage()
              }
            }}
          />

          <PulseBar mode={barMode} isVisible={isOpen} />

          <CommandList className="max-h-[400px] scroll-mask transition-[height] duration-150 ease-in-out h-[var(--cmdk-list-height)]">
            {sortedGroups.map((group) => (
              <CommandGroupRenderer key={group.id} group={group} />
            ))}
            <CommandEmpty search={search} />
          </CommandList>

          <CommandFooter hideResultsCount={!shouldFilter} className="mt-1">
            {pageStack.length > 1 ? (
              <button
                onClick={popPage}
                className="hover:text-foreground mr-2 flex items-center gap-1 transition-colors"
              >
                <Kbd>Backspace</Kbd> to go back
              </button>
            ) : (
              <span>
                Use <Kbd>↑</Kbd> <Kbd>↓</Kbd> to navigate
              </span>
            )}
          </CommandFooter>
        </Command>
      </motion.div>
    </CommandDialog>
  )
}

function Breadcrumbs() {
  const { pageStack, pages } = useCommandPalette()
  const { goToPage } = useCommandPaletteActions()

  return (
    <div className="flex items-center gap-1 border-b-[0.5px] px-3 py-2 text-xs text-muted-foreground dark:bg-muted/20 bg-muted">
      {pageStack.map((id, i) => {
        const isLast = i === pageStack.length - 1
        const page = pages.get(id)

        return (
          <React.Fragment key={id}>
            {i > 0 && (
              <span className="opacity-50" aria-hidden="true">
                /
              </span>
            )}
            <button
              className={cn(isLast && 'font-medium', 'hover:text-foreground transition-colors')}
              onClick={() => goToPage(id)}
            >
              {page?.meta.label ?? id}
            </button>
          </React.Fragment>
        )
      })}
    </div>
  )
}

function CommandGroupRenderer({ group }: { group: CommandGroup }) {
  const sortedCommands = useMemo(() => Array.from(group.commands.values()), [group.commands])

  if (sortedCommands.length === 0) return null

  return (
    <CommandGroupPrimitive heading={group.label} className="px-2">
      {sortedCommands.map((cmd) => (
        <CommandItem key={cmd.id} config={cmd} />
      ))}
    </CommandGroupPrimitive>
  )
}

function CommandItem({ config }: { config: CommandConfig }) {
  const { pushPage, setIsOpen } = useCommandPaletteActions()

  const handleSelect = useCallback(() => {
    if (config.page) {
      pushPage(config.page)
    } else {
      config.onSelect?.()
      if (!config.chainable) {
        setIsOpen(false)
      }
    }
  }, [config, pushPage, setIsOpen])

  const value = config.value ?? (typeof config.label === 'string' ? config.label : config.id)

  return (
    <CommandItemPrimitive
      value={value}
      onSelect={handleSelect}
      disabled={config.disabled || config.loading}
      keywords={config.keywords}
      className={cn(
        'dark:data-[selected=true]:bg-accent/80 data-[selected=true]:bg-accent py-3 px-2 text-foreground ',
        config.className,
      )}
    >
      <div className="flex items-center gap-2 flex-1 line-clamp-1">
        {config.loading ? (
          <Loader2 className="h-4 w-4 animate-spin text-muted-foreground" />
        ) : (
          config.icon && <span className="text-muted-foreground">{config.icon}</span>
        )}
        <span className="overflow-hidden">{config.label}</span>
      </div>

      {config.page && <ChevronRight className="ml-2 h-4 w-4 text-muted-foreground/50" />}
    </CommandItemPrimitive>
  )
}

function CommandFooter({
  children,
  hideResultsCount = false,
  className,
}: {
  children: ReactNode
  hideResultsCount?: boolean
  className?: string
}) {
  const resultsCount = useCommandState((state) => state.filtered.count)

  return (
    <div
      className={cn(
        'border-t-[0.5px] p-2 text-xs text-muted-foreground dark:bg-muted/20 bg-muted flex justify-between',
        className,
      )}
    >
      {children}
      {!hideResultsCount && <span className="ml-auto">{pluralize(resultsCount, 'result', 'results')}</span>}
    </div>
  )
}

const CommandEmpty = function CommandEmpty({ search }: { search: string }) {
  return (
    <CommandEmptyPrimitive className="text-muted-foreground py-6 text-center text-sm">
      No results found for <span className="text-foreground">"{search}"</span>.
    </CommandEmptyPrimitive>
  )
}

const PulseBar = function PulseBar({
  mode,
  isVisible,
  className,
}: {
  mode: 'flash' | 'pulse'
  isVisible: boolean
  className?: string
}) {
  const gradientBackground = `linear-gradient(90deg, transparent, #66F0C2, #00C241, #66F0C2, transparent)`

  if (!isVisible) return null

  return (
    <div
      className={cn(
        'relative flex items-center justify-center overflow-hidden w-full h-[0.5px] transform-gpu',
        className,
      )}
    >
      <AnimatePresence>
        {mode === 'flash' && (
          <motion.div
            key="flash-mode"
            className="absolute left-1/2 top-0 bottom-0 w-1/2"
            style={{ background: gradientBackground }}
            initial={{ x: '-50%', width: 100, opacity: 0 }}
            animate={{ width: '200%', opacity: [1, 1, 1, 0] }}
            transition={{ duration: 0.3, delay: 0.1 }}
          />
        )}

        {mode === 'pulse' && (
          <motion.div
            key="pulse-mode"
            className="absolute left-0 top-0 bottom-0 w-full"
            style={{ background: gradientBackground }}
            initial={{ opacity: 0 }}
            animate={{ opacity: [0, 1, 0] }}
            transition={{ duration: 0.75, repeat: Infinity }}
          />
        )}
      </AnimatePresence>
    </div>
  )
}

export function CommandHighlight({ children }: { children: ReactNode }) {
  return <span className="text-foreground bg-card rounded-sm px-1 border border-border">{children}</span>
}

export function CommandError({
  message = 'Something went wrong',
  onRetry,
  className,
}: {
  message?: string
  onRetry?: () => void
  className?: string
}) {
  return (
    <div className={cn('flex flex-col items-center justify-center py-6 px-4 text-center', className)}>
      <AlertCircle className="h-6 w-6 text-destructive mb-2" />
      <p className="text-sm text-muted-foreground">{message}</p>
      {onRetry && (
        <button
          onClick={onRetry}
          className="mt-2 text-sm text-primary hover:underline focus:outline-none focus-visible:ring-2 focus-visible:ring-ring rounded"
        >
          Try again
        </button>
      )}
    </div>
  )
}

export function CommandLoading({ count = 3, className }: { count?: number; className?: string }) {
  return (
    <div className={cn('p-1', className)}>
      {Array.from({ length: count }).map((_, i) => (
        <div key={i} className="flex items-center gap-2 px-2 py-3">
          <Skeleton className="h-4 w-4 rounded" />
          <Skeleton className="h-4 flex-1 max-w-[180px] rounded" />
        </div>
      ))}
    </div>
  )
}

export { Kbd, KbdGroup } from './ui/kbd'
