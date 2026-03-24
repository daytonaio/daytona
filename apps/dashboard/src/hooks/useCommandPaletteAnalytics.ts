/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { usePostHog } from 'posthog-js/react'
import { useCallback, useMemo } from 'react'

function isTrackingEnabled(): boolean {
  return import.meta.env.PROD
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
  const posthogClient = usePostHog()

  const trackOpened = useCallback(
    (trigger: string) => {
      if (!isTrackingEnabled()) return
      posthogClient?.capture('command_palette_opened', { trigger })
    },
    [posthogClient],
  )

  const trackCommandExecuted = useCallback(
    (props: { commandId: string; groupId: string; pageId: string }) => {
      if (!isTrackingEnabled()) return
      posthogClient?.capture('command_palette_command_executed', {
        ...props,
        commandId: normalizeCommandId(props.commandId),
      })
    },
    [posthogClient],
  )

  const trackPageNavigated = useCallback(
    (props: { fromPage: string; toPage: string; commandId: string; groupId: string }) => {
      if (!isTrackingEnabled()) return
      posthogClient?.capture('command_palette_page_navigated', {
        ...props,
        commandId: normalizeCommandId(props.commandId),
      })
    },
    [posthogClient],
  )

  const trackSearched = useCallback(
    (props: { pageId: string; queryLength: number; resultCount: number }) => {
      if (!isTrackingEnabled()) return
      posthogClient?.capture('command_palette_searched', props)
    },
    [posthogClient],
  )

  return useMemo(
    () => ({ trackOpened, trackCommandExecuted, trackPageNavigated, trackSearched }),
    [trackOpened, trackCommandExecuted, trackPageNavigated, trackSearched],
  )
}
