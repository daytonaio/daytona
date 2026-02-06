/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import type { EvaluationContext, Provider, ResolutionDetails, Hook, JsonValue, Logger } from '@openfeature/server-sdk'
import { TypeMismatchError, StandardResolutionReasons, ErrorCode } from '@openfeature/server-sdk'
import { PostHog } from 'posthog-node'
import type { PostHogOptions } from 'posthog-node'

export interface OpenFeaturePostHogProviderConfig {
  /** PostHog project API key (starts with phc_) - if not provided, all flags return default values */
  apiKey?: string
  /** Optional PostHog client options */
  clientOptions?: PostHogOptions
  /** Whether to evaluate flags locally (default: false) */
  evaluateLocally?: boolean
}

export class OpenFeaturePostHogProvider implements Provider {
  readonly metadata = {
    name: 'simple-posthog-provider',
  } as const

  private readonly client?: PostHog
  private readonly evaluateLocally: boolean
  private readonly isConfigured: boolean

  constructor(config: OpenFeaturePostHogProviderConfig = {}) {
    this.evaluateLocally = config.evaluateLocally ?? false
    this.isConfigured = !!config.apiKey

    if (config.apiKey) {
      try {
        this.client = new PostHog(config.apiKey, config.clientOptions)
      } catch (error) {
        console.error('Failed to initialize PostHog client. Feature flags will use default values.', error)
      }
    } else {
      console.log('PostHog not configured. Feature flags will use default values.')
    }
  }

  async resolveBooleanEvaluation(
    flagKey: string,
    defaultValue: boolean,
    context: EvaluationContext,
    logger: Logger,
  ): Promise<ResolutionDetails<boolean>> {
    logger.debug(`Evaluating flag ${flagKey} with context and default value:`, context, defaultValue)
    const result = await this.evaluateFlag(flagKey, defaultValue, context, logger)

    if (typeof result.value === 'boolean') {
      return result as ResolutionDetails<boolean>
    }

    throw new TypeMismatchError(`Flag ${flagKey} expected boolean, got ${typeof result.value}`)
  }

  async resolveStringEvaluation(
    flagKey: string,
    defaultValue: string,
    context: EvaluationContext,
    logger: Logger,
  ): Promise<ResolutionDetails<string>> {
    const result = await this.evaluateFlag(flagKey, defaultValue, context, logger)

    if (typeof result.value === 'string') {
      return result as ResolutionDetails<string>
    }

    throw new TypeMismatchError(`Flag ${flagKey} expected string, got ${typeof result.value}`)
  }

  async resolveNumberEvaluation(
    flagKey: string,
    defaultValue: number,
    context: EvaluationContext,
    logger: Logger,
  ): Promise<ResolutionDetails<number>> {
    const result = await this.evaluateFlag(flagKey, defaultValue, context, logger)

    if (typeof result.value === 'number') {
      return result as ResolutionDetails<number>
    }

    throw new TypeMismatchError(`Flag ${flagKey} expected number, got ${typeof result.value}`)
  }

  async resolveObjectEvaluation<T extends JsonValue>(
    flagKey: string,
    defaultValue: T,
    context: EvaluationContext,
    logger: Logger,
  ): Promise<ResolutionDetails<T>> {
    // If PostHog is not configured, return default value
    if (!this.isConfigured || !this.client) {
      logger.debug(`PostHog not configured, returning default value for flag ${flagKey}`)
      return {
        value: defaultValue,
        reason: StandardResolutionReasons.DEFAULT,
      }
    }

    const targetingKey = this.getTargetingKey(context)

    if (!targetingKey) {
      return {
        value: defaultValue,
        reason: StandardResolutionReasons.ERROR,
        errorCode: ErrorCode.GENERAL,
      }
    }

    try {
      const flagContext = this.buildFlagContext(context)
      const payload = await this.client.getFeatureFlagPayload(flagKey, targetingKey, undefined, {
        onlyEvaluateLocally: this.evaluateLocally,
        sendFeatureFlagEvents: true,
        ...flagContext,
      })

      if (payload === undefined) {
        return {
          value: defaultValue,
          reason: StandardResolutionReasons.DEFAULT,
          errorCode: ErrorCode.FLAG_NOT_FOUND,
        }
      }

      return {
        value: payload as T,
        reason: StandardResolutionReasons.TARGETING_MATCH,
      }
    } catch (error) {
      logger.error(`Error evaluating flag ${flagKey}:`, error)
      return {
        value: defaultValue,
        reason: StandardResolutionReasons.ERROR,
        errorCode: ErrorCode.GENERAL,
      }
    }
  }

  private async evaluateFlag(
    flagKey: string,
    defaultValue: any,
    context: EvaluationContext,
    logger: Logger,
  ): Promise<ResolutionDetails<any>> {
    // If PostHog is not configured, return default value
    if (!this.isConfigured || !this.client) {
      logger.debug(`PostHog not configured, returning default value for flag ${flagKey}`)
      return {
        value: defaultValue,
        reason: StandardResolutionReasons.DEFAULT,
      }
    }

    const targetingKey = this.getTargetingKey(context)

    if (!targetingKey) {
      logger.warn('No targetingKey provided in context')
      return {
        value: defaultValue,
        reason: StandardResolutionReasons.ERROR,
        errorCode: ErrorCode.GENERAL,
      }
    }

    try {
      const flagContext = this.buildFlagContext(context)
      const flagValue = await this.client.getFeatureFlag(flagKey, targetingKey, {
        onlyEvaluateLocally: this.evaluateLocally,
        sendFeatureFlagEvents: true,
        ...flagContext,
      })

      if (flagValue === undefined) {
        return {
          value: defaultValue,
          reason: StandardResolutionReasons.DEFAULT,
          errorCode: ErrorCode.FLAG_NOT_FOUND,
        }
      }

      return {
        value: flagValue,
        reason: StandardResolutionReasons.TARGETING_MATCH,
      }
    } catch (error) {
      logger.error(`Error evaluating flag ${flagKey}:`, error)
      return {
        value: defaultValue,
        reason: StandardResolutionReasons.ERROR,
        errorCode: ErrorCode.GENERAL,
      }
    }
  }

  private getTargetingKey(context: EvaluationContext): string | undefined {
    return context.targetingKey
  }

  private buildFlagContext(context: EvaluationContext) {
    const flagContext: {
      groups?: Record<string, string>
      groupProperties?: Record<string, Record<string, string>>
      personProperties?: Record<string, string>
    } = {}

    // Extract groups from context
    if (context.groups) {
      flagContext.groups = context.groups as Record<string, string>
    }

    // Extract custom properties
    if (context.personProperties) {
      flagContext.personProperties = context.personProperties as Record<string, string>
    }

    if (context.groupProperties) {
      flagContext.groupProperties = context.groupProperties as Record<string, Record<string, string>>
    }

    // Use organizationId from context attributes
    if (context.organizationId && !flagContext.groups?.organization) {
      flagContext.groups = {
        ...flagContext.groups,
        organization: context.organizationId as string,
      }
    }

    return flagContext
  }

  get hooks(): Hook[] {
    return []
  }

  async onClose(): Promise<void> {
    if (this.client) {
      await this.client.shutdown()
    }
  }
}
