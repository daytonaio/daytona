/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { measureNaturalWidth, prepareWithSegments } from '@chenglou/pretext'
import { useEffect, useMemo, useRef, useState, type ReactNode } from 'react'

import { cn } from '@/lib/utils'

import { getCanvasFont } from './utils'

function fitSubstring(text: string, font: string, maxWidth: number, suffix = '') {
  if (!text || maxWidth <= 0) {
    return ''
  }

  let low = 0
  let high = text.length
  let best = ''

  while (low <= high) {
    const middle = Math.floor((low + high) / 2)
    const candidate = text.slice(0, middle) + suffix

    if (measureNaturalWidth(prepareWithSegments(candidate, font)) <= maxWidth) {
      best = candidate
      low = middle + 1
    } else {
      high = middle - 1
    }
  }

  return best
}

function fitSubstringEnd(text: string, font: string, maxWidth: number, prefix = '') {
  if (!text || maxWidth <= 0) {
    return ''
  }

  let low = 0
  let high = text.length
  let best = ''

  while (low <= high) {
    const middle = Math.floor((low + high) / 2)
    const candidate = prefix + text.slice(text.length - middle)

    if (measureNaturalWidth(prepareWithSegments(candidate, font)) <= maxWidth) {
      best = candidate
      low = middle + 1
    } else {
      high = middle - 1
    }
  }

  return best
}

export function buildPathHeaderSnippet(text: string, font: string, maxWidth: number) {
  if (!text || !font || maxWidth <= 0) {
    return text
  }

  if (measureNaturalWidth(prepareWithSegments(text, font)) <= maxWidth) {
    return text
  }

  const ellipsis = '…'
  const lastSlashIndex = text.lastIndexOf('/')

  if (lastSlashIndex <= 0) {
    return fitSubstring(text, font, maxWidth, ellipsis) || text
  }

  const prefix = text.slice(0, lastSlashIndex + 1)
  const basename = text.slice(lastSlashIndex + 1)
  const basenameWithEllipsis = fitSubstringEnd(basename, font, Math.max(maxWidth / 2, 80), ellipsis)
  const guaranteedTail = basenameWithEllipsis || fitSubstringEnd(basename, font, maxWidth, ellipsis) || basename
  const tailWidth = measureNaturalWidth(prepareWithSegments(guaranteedTail, font))
  const prefixBudget = Math.max(0, maxWidth - tailWidth)
  const visiblePrefix = fitSubstring(prefix, font, prefixBudget)

  if (!visiblePrefix) {
    return guaranteedTail
  }

  return `${visiblePrefix}${guaranteedTail}`
}

export function buildSearchLabelSnippet(text: string, query: string, font: string, maxWidth: number) {
  if (!text || !query || !font || maxWidth <= 0) {
    return text
  }

  if (measureNaturalWidth(prepareWithSegments(text, font)) <= maxWidth) {
    return text
  }

  const normalizedQuery = query.trim()
  const lowercaseText = text.toLowerCase()
  const lowercaseQuery = normalizedQuery.toLowerCase()
  const matchIndex = lowercaseText.indexOf(lowercaseQuery)

  if (matchIndex === -1) {
    return fitSubstring(text, font, maxWidth, '…') || text
  }

  const ellipsis = '…'
  const matchTail = text.slice(matchIndex)
  const fullPrefix = text.slice(0, matchIndex)
  const queryEndIndex = matchIndex + normalizedQuery.length
  const tailWithTruncation = fitSubstring(
    matchTail,
    font,
    maxWidth,
    matchTail.length > normalizedQuery.length ? ellipsis : '',
  )
  const fallbackTail =
    measureNaturalWidth(prepareWithSegments(matchTail, font)) <= maxWidth
      ? matchTail
      : fitSubstring(text.slice(matchIndex, queryEndIndex), font, maxWidth)

  const tail = tailWithTruncation || fallbackTail || text.slice(matchIndex, queryEndIndex)
  const tailWidth = measureNaturalWidth(prepareWithSegments(tail, font))
  const prefixBudget = Math.max(0, maxWidth - tailWidth - measureNaturalWidth(prepareWithSegments(ellipsis, font)))
  const prefix = prefixBudget > 0 && fullPrefix.length > 0 ? fitSubstring(fullPrefix, font, prefixBudget) : ''

  if (!prefix) {
    return measureNaturalWidth(prepareWithSegments(tail, font)) <= maxWidth ? tail : fitSubstring(tail, font, maxWidth)
  }

  return `${prefix}${ellipsis}${tail}`
}

function HighlightedMatch({ query, text }: { query: string; text: string }) {
  const normalizedQuery = query.trim()

  if (!normalizedQuery) {
    return text
  }

  const lowercaseText = text.toLowerCase()
  const lowercaseQuery = normalizedQuery.toLowerCase()
  const parts: ReactNode[] = []
  let searchStart = 0
  let matchIndex = lowercaseText.indexOf(lowercaseQuery, searchStart)

  while (matchIndex !== -1) {
    if (matchIndex > searchStart) {
      parts.push(text.slice(searchStart, matchIndex))
    }

    const matchEnd = matchIndex + normalizedQuery.length
    parts.push(
      <mark
        key={`${text}-${matchIndex}`}
        className="rounded bg-[#2fcc712b] px-[2px] text-[#058157] not-italic dark:text-[#2fcc71]"
      >
        {text.slice(matchIndex, matchEnd)}
      </mark>,
    )

    searchStart = matchEnd
    matchIndex = lowercaseText.indexOf(lowercaseQuery, searchStart)
  }

  if (searchStart < text.length) {
    parts.push(text.slice(searchStart))
  }

  return parts.length > 0 ? parts : text
}

export function SearchResultLabel({
  availableWidth,
  className,
  font,
  query,
  text,
}: {
  availableWidth: number
  className?: string
  font: string
  query: string
  text: string
}) {
  const displayText = useMemo(() => {
    if (!font || availableWidth <= 0) {
      return text
    }

    return buildSearchLabelSnippet(text, query, font, availableWidth)
  }, [availableWidth, font, query, text])

  return (
    <span className={cn('block min-w-0 overflow-hidden whitespace-nowrap', className)}>
      <HighlightedMatch text={displayText} query={query} />
    </span>
  )
}

export function PathHeaderLabel({ className, text }: { className?: string; text: string }) {
  const containerRef = useRef<HTMLSpanElement>(null)
  const [availableWidth, setAvailableWidth] = useState(0)
  const [font, setFont] = useState('')

  useEffect(() => {
    const element = containerRef.current
    if (!element) {
      return
    }

    const updateMeasurements = () => {
      setAvailableWidth(element.getBoundingClientRect().width)
      setFont(getCanvasFont(element))
    }

    updateMeasurements()

    const resizeObserver = new ResizeObserver(() => {
      updateMeasurements()
    })

    resizeObserver.observe(element)
    return () => resizeObserver.disconnect()
  }, [])

  const displayText = useMemo(() => {
    if (!font || availableWidth <= 0) {
      return text
    }

    return buildPathHeaderSnippet(text, font, availableWidth)
  }, [availableWidth, font, text])

  return (
    <span ref={containerRef} className={cn('block min-w-0 overflow-hidden whitespace-nowrap', className)}>
      {displayText}
    </span>
  )
}
