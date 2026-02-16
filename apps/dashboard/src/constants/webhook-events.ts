/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { WebhookEvent } from '@daytonaio/api-client'

export const WEBHOOK_EVENTS: Array<{ value: WebhookEvent; label: string; category: string }> = [
  { value: WebhookEvent.SANDBOX_CREATED, label: 'Sandbox Created', category: 'Sandbox' },
  { value: WebhookEvent.SANDBOX_STATE_UPDATED, label: 'Sandbox State Updated', category: 'Sandbox' },
  { value: WebhookEvent.SNAPSHOT_CREATED, label: 'Snapshot Created', category: 'Snapshot' },
  { value: WebhookEvent.SNAPSHOT_REMOVED, label: 'Snapshot Removed', category: 'Snapshot' },
  { value: WebhookEvent.SNAPSHOT_STATE_UPDATED, label: 'Snapshot State Updated', category: 'Snapshot' },
  { value: WebhookEvent.VOLUME_CREATED, label: 'Volume Created', category: 'Volume' },
  { value: WebhookEvent.VOLUME_STATE_UPDATED, label: 'Volume State Updated', category: 'Volume' },
] as const

export const WEBHOOK_EVENT_CATEGORIES = ['Sandbox', 'Snapshot', 'Volume'] as const

export type WebhookEventValue = (typeof WEBHOOK_EVENTS)[number]['value']
export type WebhookEventCategory = (typeof WEBHOOK_EVENT_CATEGORIES)[number]
