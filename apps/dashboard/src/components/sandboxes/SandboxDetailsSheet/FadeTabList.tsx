/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Button } from '@/components/ui/button'
import { ScrollArea } from '@/components/ui/scroll-area'
import { TabsList } from '@/components/ui/tabs'
import { cn } from '@/lib/utils'
import { ChevronLeft, ChevronRight } from 'lucide-react'
import { AnimatePresence, motion } from 'motion/react'
import React, { useCallback, useEffect, useState } from 'react'

function FadeTabListButton({
  direction,
  visible,
  onClick,
}: {
  direction: 'left' | 'right'
  visible: boolean
  onClick: () => void
}) {
  const isLeft = direction === 'left'

  return (
    <AnimatePresence initial={false}>
      {visible && (
        <motion.div
          key={`scroll-tabs-${direction}`}
          initial={{ opacity: 0, x: isLeft ? -10 : 10 }}
          animate={{ opacity: 1, x: 0 }}
          exit={{ opacity: 0, x: isLeft ? -10 : 10 }}
          transition={{ duration: 0.16, ease: 'easeOut' }}
          className={cn(
            "pointer-events-none absolute inset-y-0 z-20 flex w-20 items-center after:absolute after:inset-0 after:z-0 after:bg-background/90 after:content-['']",
            {
              'left-0 justify-start pl-1 after:[mask-image:linear-gradient(to_right,black_50%,transparent_100%)]':
                isLeft,
              'right-0 justify-end pr-1 after:[mask-image:linear-gradient(to_left,black_50%,transparent_100%)]':
                !isLeft,
            },
          )}
        >
          <Button
            type="button"
            variant="ghost"
            size="icon-sm"
            className="pointer-events-auto relative z-10 size-8 text-muted-foreground hover:text-foreground"
            onClick={onClick}
            aria-label={isLeft ? 'Scroll tabs left' : 'Scroll tabs right'}
          >
            {isLeft ? <ChevronLeft className="size-4" /> : <ChevronRight className="size-4" />}
          </Button>
        </motion.div>
      )}
    </AnimatePresence>
  )
}

export function FadeTabList({
  children,
  leadingContent,
}: {
  children: React.ReactNode
  leadingContent?: React.ReactNode
}) {
  const rootRef = React.useRef<HTMLDivElement | null>(null)
  const tabViewportRef = React.useRef<HTMLDivElement | null>(null)
  const [canScrollLeft, setCanScrollLeft] = useState(false)
  const [canScrollRight, setCanScrollRight] = useState(false)
  const [isTabStripHovered, setIsTabStripHovered] = useState(false)

  const updateScrollState = useCallback(() => {
    const viewport = tabViewportRef.current
    if (!viewport) {
      setCanScrollLeft(false)
      setCanScrollRight(false)
      return
    }

    const remainingScroll = viewport.scrollWidth - viewport.clientWidth - viewport.scrollLeft
    setCanScrollLeft(viewport.scrollLeft > 1)
    setCanScrollRight(remainingScroll > 1)
  }, [])

  const scrollTabs = useCallback((direction: 'left' | 'right') => {
    const viewport = tabViewportRef.current
    if (!viewport) {
      return
    }

    viewport.scrollBy({
      left: (direction === 'left' ? -1 : 1) * Math.max(180, viewport.clientWidth * 0.7),
      behavior: 'smooth',
    })
  }, [])

  useEffect(() => {
    const root = rootRef.current
    const viewport = tabViewportRef.current
    if (!root || !viewport) {
      return
    }

    updateScrollState()
    viewport.addEventListener('scroll', updateScrollState, { passive: true })
    const resizeObserver = new ResizeObserver(updateScrollState)
    resizeObserver.observe(root)

    return () => {
      viewport.removeEventListener('scroll', updateScrollState)
      resizeObserver.disconnect()
    }
  }, [updateScrollState])

  useEffect(() => {
    updateScrollState()
  }, [children, leadingContent, updateScrollState])

  return (
    <div
      ref={rootRef}
      className="relative flex h-[42px] shrink-0 border-b border-border"
      onMouseEnter={() => setIsTabStripHovered(true)}
      onMouseLeave={() => setIsTabStripHovered(false)}
      onFocusCapture={() => setIsTabStripHovered(true)}
      onBlurCapture={(event) => {
        if (!event.currentTarget.contains(event.relatedTarget)) {
          setIsTabStripHovered(false)
        }
      }}
    >
      {leadingContent ? <div className="flex h-full shrink-0 items-center">{leadingContent}</div> : null}
      <div className="relative h-full min-w-0 flex-1">
        <ScrollArea
          fade="mask"
          horizontal
          vertical={false}
          fadeOffset={36}
          viewportRef={tabViewportRef}
          className="h-full [&_[data-slot=scroll-area-scrollbar]]:hidden [&_[data-slot=scroll-area-viewport]]:pb-px"
        >
          <TabsList variant="underline" className="h-[41px] w-max min-w-full border-b-0">
            {children}
          </TabsList>
        </ScrollArea>
        <FadeTabListButton
          direction="left"
          visible={canScrollLeft && isTabStripHovered}
          onClick={() => scrollTabs('left')}
        />
        <FadeTabListButton
          direction="right"
          visible={canScrollRight && isTabStripHovered}
          onClick={() => scrollTabs('right')}
        />
      </div>
    </div>
  )
}
