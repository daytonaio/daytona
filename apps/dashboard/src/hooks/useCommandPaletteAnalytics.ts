/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import posthog from 'posthog-js'
import { usePostHog } from 'posthog-js/react'
import { useCallback, useMemo } from 'react'

function isTrackingEnabled(): boolean {
  return import.meta.env.PROD
}

export function trackCommandPaletteOpened(trigger: string) {
  if (!isTrackingEnabled()) return
  posthog?.capture('command_palette_opened', { trigger })
}

/**
 * Strips dynamic suffixes from command IDs to keep PostHog cardinality low.
 * e.g. "docs-abc123" → "docs", "switch-org-uuid" → "switch-org", "toggle-theme" → "toggle-theme"
 */
const DYNAMIC_ID_PREFIXES = ['docs-', 'cli-', 'sdk-', 'switch-org-']

function normalizeCommandId(commandId: string): string {
  for (const prefix of DYNAMIC_ID_PREFIXES) {
    if (commandId.startsWith(prefix)) {
      return prefix.slice(0, -1) // strip trailing dash: "docs-" → "docs"
    }
  }
  return commandId
}

export function useCommandPaletteAnalytics() {
  const posthog = usePostHog()

  const trackCommandExecuted = useCallback(
    (props: { commandId: string; groupId: string; pageId: string }) => {
      if (!isTrackingEnabled()) return
      posthog?.capture('command_palette_command_executed', {
        ...props,
        commandId: normalizeCommandId(props.commandId),
      })
    },
    [posthog],
  )

  const trackPageNavigated = useCallback(
    (props: { fromPage: string; toPage: string; commandId: string; groupId: string }) => {
      if (!isTrackingEnabled()) return
      posthog?.capture('command_palette_page_navigated', {
        ...props,
        commandId: normalizeCommandId(props.commandId),
      })
    },
    [posthog],
  )

  const trackSearched = useCallback(
    (props: { pageId: string; queryLength: number; resultCount: number }) => {
      if (!isTrackingEnabled()) return
      posthog?.capture('command_palette_searched', props)
    },
    [posthog],
  )

  return useMemo(
    () => ({ trackCommandExecuted, trackPageNavigated, trackSearched }),
    [trackCommandExecuted, trackPageNavigated, trackSearched],
  )
}
