/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { cn } from '@/lib/utils'
import { cva, type VariantProps } from 'class-variance-authority'
import {
  AlertCircleIcon,
  AlertTriangleIcon,
  CheckCircle2Icon,
  ChevronRight,
  InfoIcon,
  MegaphoneIcon,
  XIcon,
} from 'lucide-react'
import { AnimatePresence, motion } from 'motion/react'
import { createContext, useCallback, useContext, useEffect, useMemo, useState } from 'react'
import { v4 as uuidv4 } from 'uuid'

type BannerVariant = 'info' | 'success' | 'warning' | 'error' | 'neutral'

const variantIcons = {
  info: InfoIcon,
  success: CheckCircle2Icon,
  warning: AlertTriangleIcon,
  error: AlertCircleIcon,
  neutral: MegaphoneIcon,
}

const priorityMap: Record<BannerVariant, number> = {
  error: 0,
  warning: 1,
  success: 2,
  info: 3,
  neutral: 4,
}

const bannerVariants = cva('relative overflow-hidden backdrop-blur-xl border-y w-full', {
  variants: {
    variant: {
      info: 'bg-info-background text-info-foreground border-info-separator',
      success: 'bg-success-background text-success-foreground border-success-separator',
      warning: 'bg-warning-background text-warning-foreground border-warning-separator',
      error: 'bg-destructive-background text-destructive-foreground border-destructive-separator',
      neutral: 'bg-muted/40 border-border',
    },
  },
  defaultVariants: {
    variant: 'info',
  },
})

interface BannerAction {
  label: string
  onClick: () => void
}

interface BannerNotification {
  id?: string
  variant?: BannerVariant
  title: string
  description?: string
  action?: BannerAction
  icon?: React.ReactNode
  onDismiss?: () => void
  isDismissible?: boolean
}

interface BannerContextValue {
  notifications: BannerNotification[]
  addBanner: (notification: BannerNotification) => string
  removeBanner: (id: string) => void
  clearBanners: () => void
}

const BannerContext = createContext<BannerContextValue | null>(null)

export const useBanner = () => {
  const context = useContext(BannerContext)
  if (!context) {
    throw new Error('useBanner must be used within a BannerProvider')
  }
  return context
}

interface BannerProviderProps {
  children: React.ReactNode
  defaultNotifications?: BannerNotification[]
}

export const BannerProvider = ({ children, defaultNotifications = [] }: BannerProviderProps) => {
  const [notifications, setNotifications] = useState<BannerNotification[]>(defaultNotifications)

  const addBanner = useCallback((notification: BannerNotification) => {
    const id = notification.id || uuidv4()
    setNotifications((prev) => {
      const existingIndex = prev.findIndex((n) => n.id === id)
      if (existingIndex >= 0) {
        const updated = [...prev]
        updated[existingIndex] = { ...notification, id }
        return updated
      }
      return [{ ...notification, id }, ...prev]
    })
    return id
  }, [])

  const removeBanner = useCallback((id: string) => {
    setNotifications((prev) => prev.filter((n) => n.id !== id))
  }, [])

  const clearBanners = useCallback(() => {
    setNotifications([])
  }, [])

  const sortedNotifications = useMemo(() => {
    return [...notifications].sort((a, b) => {
      const variantA = (a.variant || 'info') as BannerVariant
      const variantB = (b.variant || 'info') as BannerVariant
      return priorityMap[variantA] - priorityMap[variantB]
    })
  }, [notifications])

  const contextValue = useMemo(
    () => ({
      notifications: sortedNotifications,
      addBanner,
      removeBanner,
      clearBanners,
    }),
    [sortedNotifications, addBanner, removeBanner, clearBanners],
  )

  return <BannerContext.Provider value={contextValue}>{children}</BannerContext.Provider>
}

interface BannerProps extends VariantProps<typeof bannerVariants> {
  title: string
  description?: string
  action?: BannerAction
  onDismiss?: () => void
  total?: number
  currentIndex?: number
  onNext?: () => void
  isDismissible?: boolean
  className?: string
  icon?: React.ReactNode
}

export const Banner = ({
  variant = 'info',
  title,
  description,
  action,
  onDismiss,
  total = 0,
  currentIndex = 0,
  onNext,
  isDismissible = true,
  className,
  icon,
  ...props
}: BannerProps & React.ComponentProps<typeof motion.div>) => {
  const IconComponent = variantIcons[variant ?? 'info']
  const role = variant === 'error' || variant === 'warning' ? 'alert' : 'status'

  return (
    <motion.div layout className={cn('w-full relative z-30 origin-top', className)} {...props}>
      <div className={cn(bannerVariants({ variant }))} role={role}>
        <div className="grid sm:grid-cols-[auto_1fr_auto_auto] grid-cols-[auto_1fr_auto_auto] grid-rows-[auto_auto] sm:grid-rows-1 items-center gap-x-2 px-4 sm:px-5 py-2 max-w-5xl mx-auto">
          {icon || <IconComponent className="h-4 w-4 flex-shrink-0 text-current" />}

          <div className="flex items-center gap-3 overflow-hidden">
            <span className="text-sm font-semibold shrink-0">{title}</span>
            {description && (
              <>
                <span className="hidden md:flex opacity-20 border-l border-current h-6" />
                <span className="text-sm opacity-90 line-clamp-1 max-w-2xl">{description}</span>
              </>
            )}
          </div>

          {action && (
            <button
              type="button"
              onClick={action.onClick}
              className="text-sm font-medium underline-offset-4 underline row-[2] sm:row-[1] col-[2] sm:col-[3] justify-self-start"
            >
              {action.label}
            </button>
          )}

          {total > 1 && (
            <div className="flex items-center gap-2">
              <span className="opacity-20 border-l border-current h-6" />
              <div className="flex items-center gap-1">
                <span className="text-xs font-medium tabular-nums">
                  {currentIndex + 1}/{total}
                </span>
                <BannerButton onClick={() => onNext?.()} aria-label="Next Notification">
                  <ChevronRight className="w-4 h-4" />
                </BannerButton>
              </div>
            </div>
          )}

          <div className="flex items-center justify-center min-w-6 col-[-1]">
            {isDismissible && (
              <BannerButton onClick={() => onDismiss?.()} aria-label="Dismiss">
                <XIcon className="w-4 h-4" />
              </BannerButton>
            )}
          </div>
        </div>
      </div>
    </motion.div>
  )
}

function BannerButton({ className, ...props }: React.ComponentProps<'button'>) {
  return (
    <button
      type="button"
      className={cn(
        'p-1 rounded transition-colors hover:bg-black/10 dark:hover:bg-white/10 focus-visible:border-ring focus-visible:ring-ring/50 focus-visible:ring-[3px]',
        className,
      )}
      {...props}
    />
  )
}

export const BannerStack = () => {
  const { notifications, removeBanner } = useBanner()
  const [activeIndex, setActiveIndex] = useState(0)

  const next = () => setActiveIndex((prev) => (prev + 1) % notifications.length)

  useEffect(() => {
    if (notifications.length > 0 && activeIndex >= notifications.length) {
      setActiveIndex(Math.max(0, notifications.length - 1))
    }
  }, [notifications.length, activeIndex])

  const activeItem = notifications.length > 0 ? notifications[activeIndex] : null

  if (!activeItem) {
    return null
  }

  return (
    <motion.div
      layout
      className="relative w-full overflow-hidden"
      initial={false}
      animate={{ height: activeItem ? 'auto' : 0 }}
      transition={{ duration: 0.2 }}
    >
      <AnimatePresence mode="popLayout" initial={false}>
        {activeItem && (
          <Banner
            key={activeItem.id}
            {...activeItem}
            total={notifications.length}
            currentIndex={activeIndex}
            onNext={next}
            onDismiss={() => {
              activeItem.onDismiss?.()
              if (activeItem.id) {
                removeBanner(activeItem.id)
              }
            }}
            initial={{ opacity: 0, y: -20, filter: 'blur(2px)' }}
            animate={{ opacity: 1, y: 0, filter: 'blur(0px)' }}
            exit={{ opacity: 0, y: 20, filter: 'blur(2px)' }}
            transition={{
              duration: 0.2,
            }}
          />
        )}
      </AnimatePresence>
    </motion.div>
  )
}
