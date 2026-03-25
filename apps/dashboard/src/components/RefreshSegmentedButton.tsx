/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Button } from '@/components/ui/button'
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectLabel,
  SelectSeparator,
  SelectTrigger,
} from '@/components/ui/select'
import { cn } from '@/lib/utils'
import { RefreshCw } from 'lucide-react'
import { motion } from 'motion/react'
import { useEffect, useMemo, useRef, useState } from 'react'

export type RefreshIntervalValue = number | false

export interface RefreshIntervalOption {
  label: string
  value: RefreshIntervalValue
}

const DEFAULT_REFRESH_OPTIONS: RefreshIntervalOption[] = [
  { label: 'Off', value: false },
  { label: 'Every 10s', value: 10000 },
  { label: 'Every 30s', value: 30000 },
  { label: 'Every 1m', value: 60000 },
  { label: 'Every 5m', value: 300000 },
]

const MotionRefreshIcon = motion(RefreshCw)

interface RefreshSegmentedButtonProps {
  value: RefreshIntervalValue
  onChange: (value: RefreshIntervalValue) => void
  onRefresh: () => void
  options?: RefreshIntervalOption[]
  isRefreshing?: boolean
  lastUpdatedAt?: number
  disabled?: boolean
  className?: string
}

function serializeIntervalValue(value: RefreshIntervalValue) {
  return value === false ? 'off' : String(value)
}

function deserializeIntervalValue(value: string): RefreshIntervalValue {
  return value === 'off' ? false : Number(value)
}

function normalizeOptions(options: RefreshIntervalOption[]) {
  const offOption = options.find((option) => option.value === false) ?? DEFAULT_REFRESH_OPTIONS[0]
  const seenValues = new Set<RefreshIntervalValue>([false])
  const deduplicatedOptions = options.filter((option) => {
    if (seenValues.has(option.value)) {
      return false
    }

    seenValues.add(option.value)
    return true
  })

  return [offOption, ...deduplicatedOptions]
}

function getFallbackLabel(value: RefreshIntervalValue) {
  if (value === false) {
    return 'Off'
  }

  if (value < 60000) {
    return `Every ${value / 1000}s`
  }

  return `Every ${value / 60000}m`
}

function formatCountdownLabel(totalSeconds: number) {
  const minutes = Math.floor(totalSeconds / 60)
  const seconds = totalSeconds % 60

  if (minutes === 0) {
    return `${seconds}s`
  }

  if (seconds === 0) {
    return `${minutes}m`
  }

  return `${minutes}m${seconds}s`
}

function normalizeRotationValue(rotation: number) {
  const normalized = rotation % 3600
  return normalized >= 0 ? normalized : normalized + 3600
}

function distanceToTarget(current: number, target: number) {
  return (target - current + 3600) % 3600
}

export function RefreshSegmentedButton({
  value,
  onChange,
  onRefresh,
  options = DEFAULT_REFRESH_OPTIONS,
  isRefreshing = false,
  lastUpdatedAt,
  disabled = false,
  className,
}: RefreshSegmentedButtonProps) {
  const normalizedOptions = useMemo(() => normalizeOptions(options), [options])
  const selectedOptionLabel = normalizedOptions.find((option) => option.value === value)?.label
  const [countdownSeconds, setCountdownSeconds] = useState<number | null>(null)
  const [countdownStartedAt, setCountdownStartedAt] = useState<number | null>(() =>
    lastUpdatedAt ? lastUpdatedAt : null,
  )
  const [rotation, setRotation] = useState(0)
  const animationFrameRef = useRef<number | null>(null)
  const lastFrameTimeRef = useRef<number | null>(null)
  const rotationRef = useRef(0)
  const settleTargetRef = useRef<number | null>(null)
  const isRefreshingRef = useRef(isRefreshing)
  const previousValueRef = useRef<RefreshIntervalValue>(value)
  const previousUpdatedAtRef = useRef<number | undefined>(lastUpdatedAt)

  useEffect(() => {
    rotationRef.current = rotation
  }, [rotation])

  useEffect(() => {
    isRefreshingRef.current = isRefreshing
  }, [isRefreshing])

  useEffect(() => {
    if (value === false) {
      setCountdownStartedAt(null)
      previousValueRef.current = value
      return
    }

    if (previousValueRef.current !== value) {
      setCountdownStartedAt(Date.now())
      previousValueRef.current = value
    }
  }, [value])

  useEffect(() => {
    if (typeof value !== 'number' || !lastUpdatedAt) {
      previousUpdatedAtRef.current = lastUpdatedAt
      return
    }

    if (previousUpdatedAtRef.current !== lastUpdatedAt) {
      setCountdownStartedAt(lastUpdatedAt)
      previousUpdatedAtRef.current = lastUpdatedAt
    }
  }, [lastUpdatedAt, value])

  useEffect(() => {
    return () => {
      if (animationFrameRef.current !== null) {
        window.cancelAnimationFrame(animationFrameRef.current)
        animationFrameRef.current = null
      }
    }
  }, [])

  useEffect(() => {
    const degreesPerSecond = 360

    const stopAnimation = () => {
      if (animationFrameRef.current !== null) {
        window.cancelAnimationFrame(animationFrameRef.current)
        animationFrameRef.current = null
      }
      lastFrameTimeRef.current = null
    }

    const tick = (timestamp: number) => {
      const lastTimestamp = lastFrameTimeRef.current ?? timestamp
      const deltaMs = timestamp - lastTimestamp
      lastFrameTimeRef.current = timestamp

      const currentRotation = rotationRef.current
      const nextRotation = normalizeRotationValue(currentRotation + (degreesPerSecond * deltaMs) / 1000)
      const settleTarget = settleTargetRef.current

      if (!isRefreshingRef.current && settleTarget !== null) {
        const currentDistance = distanceToTarget(currentRotation, settleTarget)
        const nextDistance = distanceToTarget(nextRotation, settleTarget)

        if (nextDistance > currentDistance || nextDistance < 0.001) {
          const finalRotation = normalizeRotationValue(settleTarget)
          rotationRef.current = finalRotation
          setRotation(finalRotation)
          settleTargetRef.current = null
          stopAnimation()
          return
        }
      }

      rotationRef.current = nextRotation
      setRotation(nextRotation)
      animationFrameRef.current = window.requestAnimationFrame(tick)
    }

    if (isRefreshing) {
      settleTargetRef.current = null
      if (animationFrameRef.current === null) {
        animationFrameRef.current = window.requestAnimationFrame(tick)
      }
    } else if (animationFrameRef.current !== null) {
      const currentRotation = normalizeRotationValue(rotationRef.current)
      const nextFullTurn = Math.ceil((currentRotation + 0.0001) / 360) * 360

      if (Math.abs(nextFullTurn - currentRotation) < 0.001) {
        const finalRotation = normalizeRotationValue(nextFullTurn)
        rotationRef.current = finalRotation
        setRotation(finalRotation)
        settleTargetRef.current = null
        stopAnimation()
      } else {
        settleTargetRef.current = normalizeRotationValue(nextFullTurn)
        if (animationFrameRef.current === null) {
          animationFrameRef.current = window.requestAnimationFrame(tick)
        }
      }
    }
  }, [isRefreshing])

  useEffect(() => {
    if (typeof value !== 'number') {
      setCountdownSeconds(null)
      return
    }

    const startedAt = countdownStartedAt ?? Date.now()

    const updateCountdown = () => {
      const remaining = value - (Date.now() - startedAt)
      setCountdownSeconds(Math.max(1, Math.ceil(remaining / 1000)))
    }

    updateCountdown()
    const interval = window.setInterval(updateCountdown, 1000)

    return () => window.clearInterval(interval)
  }, [value, countdownStartedAt])

  const buttonLabel =
    value === false
      ? 'Auto-refresh off'
      : isRefreshing
        ? 'Refreshing...'
        : countdownSeconds !== null
          ? formatCountdownLabel(countdownSeconds)
          : (selectedOptionLabel ?? getFallbackLabel(value))

  return (
    <div className={cn('inline-flex items-stretch rounded-md border shadow-xs', className)}>
      <Button
        type="button"
        variant="ghost"
        size="sm"
        disabled={disabled || isRefreshing}
        onClick={onRefresh}
        aria-label="Refresh now"
        className={cn('rounded-r-none border-0 px-2.5 shadow-none hover:bg-accent/80', isRefreshing && 'opacity-70')}
      >
        <MotionRefreshIcon className="size-4" style={{ rotate: rotation }} />
      </Button>

      <Select
        value={serializeIntervalValue(value)}
        onValueChange={(nextValue) => {
          const nextInterval = deserializeIntervalValue(nextValue)
          onChange(nextInterval)
        }}
        disabled={disabled}
      >
        <SelectTrigger
          size="sm"
          className="min-w-[80px] justify-end gap-2 rounded-l-none border-0 border-l px-3 shadow-none focus-visible:ring-[3px]"
        >
          <span
            className={cn(
              'min-w-0 truncate text-right',
              countdownSeconds !== null && '[font-variant-numeric:tabular-nums]',
              isRefreshing && 'text-muted-foreground',
            )}
          >
            {buttonLabel}
          </span>
        </SelectTrigger>
        <SelectContent position="popper">
          <SelectGroup>
            <SelectLabel className="pl-2">Auto-refresh interval</SelectLabel>
            <SelectSeparator />
            {normalizedOptions.map((option) => (
              <SelectItem key={serializeIntervalValue(option.value)} value={serializeIntervalValue(option.value)}>
                {option.label}
              </SelectItem>
            ))}
          </SelectGroup>
        </SelectContent>
      </Select>
    </div>
  )
}
