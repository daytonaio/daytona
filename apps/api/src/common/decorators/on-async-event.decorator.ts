/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { OnEvent, OnEventMetadata } from '@nestjs/event-emitter'

export function OnAsyncEvent({ event, options = {} }: OnEventMetadata): MethodDecorator {
  return OnEvent(event, {
    ...options,
    promisify: true,
    suppressErrors: false,
  })
}
