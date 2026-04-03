/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import type {
  EvaluationContext,
  JsonValue,
  Logger,
  Provider,
  ResolutionDetails,
  ProviderEventEmitter,
  AnyProviderEvent,
} from '@openfeature/react-sdk'
import { OpenFeatureEventEmitter, ProviderEvents, StandardResolutionReasons } from '@openfeature/react-sdk'
import type { PostHog } from 'posthog-js'

export class PostHogWebProvider implements Provider {
  public readonly runsOn = 'client' as const

  readonly metadata = { name: 'posthog-web' } as const

  public readonly events: ProviderEventEmitter<AnyProviderEvent> = new OpenFeatureEventEmitter()

  constructor(private readonly posthog: PostHog) {}

  async initialize(): Promise<void> {
    await new Promise<void>((resolve) => {
      this.posthog.onFeatureFlags(() => resolve())
    })

    this.posthog.onFeatureFlags(() => {
      this.events.emit(ProviderEvents.ConfigurationChanged)
    })
  }

  resolveBooleanEvaluation(
    flagKey: string,
    defaultValue: boolean,
    _context: EvaluationContext,
    _logger: Logger,
  ): ResolutionDetails<boolean> {
    const value = this.posthog.isFeatureEnabled(flagKey)
    if (value === undefined) {
      return { value: defaultValue, reason: StandardResolutionReasons.DEFAULT }
    }
    return { value, reason: StandardResolutionReasons.TARGETING_MATCH }
  }

  resolveStringEvaluation(
    flagKey: string,
    defaultValue: string,
    _context: EvaluationContext,
    _logger: Logger,
  ): ResolutionDetails<string> {
    const variant = this.posthog.getFeatureFlag(flagKey)
    if (typeof variant !== 'string') {
      return { value: defaultValue, reason: StandardResolutionReasons.DEFAULT }
    }
    return { value: variant, reason: StandardResolutionReasons.TARGETING_MATCH }
  }

  resolveNumberEvaluation(
    flagKey: string,
    defaultValue: number,
    _context: EvaluationContext,
    _logger: Logger,
  ): ResolutionDetails<number> {
    const payload = this.posthog.getFeatureFlagPayload(flagKey)
    if (typeof payload !== 'number') {
      return { value: defaultValue, reason: StandardResolutionReasons.DEFAULT }
    }
    return { value: payload, reason: StandardResolutionReasons.TARGETING_MATCH }
  }

  resolveObjectEvaluation<T extends JsonValue>(
    flagKey: string,
    defaultValue: T,
    _context: EvaluationContext,
    _logger: Logger,
  ): ResolutionDetails<T> {
    const payload = this.posthog.getFeatureFlagPayload(flagKey)
    if (payload === undefined || payload === null) {
      return { value: defaultValue, reason: StandardResolutionReasons.DEFAULT }
    }
    return { value: payload as T, reason: StandardResolutionReasons.TARGETING_MATCH }
  }

  async onClose(): Promise<void> {
    // PostHog lifecycle managed by PostHogProviderWrapper — nothing to do here
  }
}
