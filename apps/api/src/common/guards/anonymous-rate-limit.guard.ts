/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable, ExecutionContext } from '@nestjs/common'
import { Reflector } from '@nestjs/core'
import { ThrottlerGuard, ThrottlerModuleOptions, ThrottlerRequest, ThrottlerStorage } from '@nestjs/throttler'
import { Request } from 'express'
import { sha256 } from '@nestjs/throttler/dist/hash'

@Injectable()
export class AnonymousRateLimitGuard extends ThrottlerGuard {
  constructor(options: ThrottlerModuleOptions, storageService: ThrottlerStorage, reflector: Reflector) {
    super(options, storageService, reflector)
  }

  protected async getTracker(req: Request): Promise<string> {
    // For anonymous requests, use IP address as tracker
    const ip = req.ips.length ? req.ips[0] : req.ip
    return `anonymous:${ip}`
  }

  protected generateKey(context: ExecutionContext, suffix: string, name: string): string {
    // Override to make rate limiting per-rate-limit-type, not per-route
    // This ensures all routes share the same counter for anonymous rate limiting
    return sha256(`${name}-${suffix}`)
  }

  async handleRequest(requestProps: ThrottlerRequest): Promise<boolean> {
    const { throttler } = requestProps

    // Apply anonymous throttler to ALL requests (with or without Bearer tokens)
    // This ensures we catch invalid/malicious tokens before they reach authentication
    if (throttler.name === 'anonymous') {
      return super.handleRequest(requestProps)
    }

    // Skip other throttlers in this guard
    return true
  }
}
