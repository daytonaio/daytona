/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

export const WEBHOOK_EVENTS = [
  { value: 'sandbox.created', label: 'Sandbox Created', category: 'Sandbox' },
  { value: 'sandbox.state.updated', label: 'Sandbox State Updated', category: 'Sandbox' },
  { value: 'snapshot.created', label: 'Snapshot Created', category: 'Snapshot' },
  { value: 'snapshot.removed', label: 'Snapshot Removed', category: 'Snapshot' },
  { value: 'snapshot.state.updated', label: 'Snapshot State Updated', category: 'Snapshot' },
  { value: 'volume.created', label: 'Volume Created', category: 'Volume' },
  { value: 'volume.state.updated', label: 'Volume State Updated', category: 'Volume' },
] as const

export const WEBHOOK_EVENT_CATEGORIES = ['Sandbox', 'Snapshot', 'Volume'] as const

export type WebhookEventValue = (typeof WEBHOOK_EVENTS)[number]['value']
export type WebhookEventCategory = (typeof WEBHOOK_EVENT_CATEGORIES)[number]
