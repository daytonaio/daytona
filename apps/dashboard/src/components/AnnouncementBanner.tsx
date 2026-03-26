/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { X, Info } from 'lucide-react'
import { Button } from './ui/button'

interface AnnouncementBannerProps {
  text: string
  learnMoreUrl?: string
  onDismiss: () => void
}

export function AnnouncementBanner({ text, learnMoreUrl, onDismiss }: AnnouncementBannerProps) {
  return (
    <div className="fixed top-0 left-0 right-0 z-50 bg-primary text-primary-foreground px-4 md:px-6 h-16 md:h-12 flex items-center">
      <div className="flex items-center justify-between w-full">
        <div className="flex items-center justify-start flex-1 gap-4 md:gap-3">
          <Info className="h-4 w-4 flex-shrink-0" />
          <div className="flex items-center gap-4 md:gap-3">
            <p className="text-sm font-medium">{text}</p>
            {learnMoreUrl && (
              <a
                href={learnMoreUrl}
                target="_blank"
                rel="noopener noreferrer"
                className="inline-flex items-center gap-1 text-sm font-medium underline whitespace-nowrap hover:text-primary-foreground/80"
              >
                Learn More
              </a>
            )}
          </div>
        </div>
        <Button
          variant="ghost"
          size="sm"
          className="h-auto p-1 hover:bg-primary-foreground/10 text-primary-foreground hover:text-primary-foreground"
          onClick={onDismiss}
          aria-label="Dismiss announcement"
        >
          <X className="h-4 w-4" />
        </Button>
      </div>
    </div>
  )
}
