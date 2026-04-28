/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

import { EventDispatcher, EventHandler } from './EventDispatcher'

type ManagedSubscription = {
  unsubscribe: () => void
  timer: ReturnType<typeof setTimeout>
}

function createSubscriptionId(): string {
  // Browser-safe UUID generation.
  if (typeof globalThis.crypto?.randomUUID === 'function') {
    return globalThis.crypto.randomUUID()
  }

  const randomHex = () =>
    Math.floor(Math.random() * 0xffffffff)
      .toString(16)
      .padStart(8, '0')
  return `${randomHex()}${randomHex()}${randomHex()}${randomHex()}`
}

export class EventSubscriptionManager {
  private static readonly SUBSCRIPTION_TTL = 300
  private readonly subscriptions = new Map<string, ManagedSubscription>()
  private _closed = false

  constructor(private readonly dispatcher: EventDispatcher) {}

  subscribe(resourceId: string, handler: EventHandler, events: string[]): string {
    // Reject operations after shutdown to prevent use-after-close.
    if (this._closed) {
      return ''
    }

    const subId = createSubscriptionId()
    const unsubscribe = this.dispatcher.subscribe(resourceId, handler, events)

    try {
      if (this._closed) {
        throw new Error('EventSubscriptionManager is closed')
      }

      this.subscriptions.set(subId, {
        unsubscribe,
        timer: this.createTimer(subId),
      })

      return subId
    } catch (error) {
      // Rollback dispatcher subscription on failure.
      unsubscribe()
      if (this._closed) {
        return ''
      }
      throw error
    }
  }

  refresh(subId: string): boolean {
    // Reject operations after shutdown to prevent use-after-close.
    if (this._closed) {
      return false
    }

    const subscription = this.subscriptions.get(subId)
    if (!subscription) {
      return false
    }

    clearTimeout(subscription.timer)
    subscription.timer = this.createTimer(subId)
    return true
  }

  unsubscribe(subId: string): void {
    const subscription = this.subscriptions.get(subId)
    if (!subscription) {
      return
    }

    clearTimeout(subscription.timer)
    this.subscriptions.delete(subId)
    subscription.unsubscribe()
  }

  shutdown(): void {
    this._closed = true
    for (const [subId, subscription] of this.subscriptions) {
      clearTimeout(subscription.timer)
      this.subscriptions.delete(subId)
      subscription.unsubscribe()
    }
  }

  private createTimer(subId: string): ReturnType<typeof setTimeout> {
    const timer = setTimeout(() => {
      const subscription = this.subscriptions.get(subId)
      if (!subscription || subscription.timer !== timer) {
        return
      }

      this.subscriptions.delete(subId)
      subscription.unsubscribe()
    }, EventSubscriptionManager.SUBSCRIPTION_TTL * 1000)

    if (typeof timer.unref === 'function') {
      timer.unref()
    }

    return timer
  }
}
