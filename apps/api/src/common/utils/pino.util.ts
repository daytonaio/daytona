/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import pino, { TransportSingleOptions } from 'pino'
import { TypedConfigService } from '../../config/typed-config.service'

/*
 * This is a workaround to swap the message and object in the arguments array.
 * It is needed because the logger in nestjs-pino is not compatible with nestjs console logger.
 * ref: https://github.com/iamolegga/nestjs-pino/issues/2004
 *
 */
export function swapMessageAndObject(
  this: pino.Logger,
  args: Parameters<pino.LogFn>,
  method: pino.LogFn,
  level: number,
): void {
  // Type guard helper
  const isPlainObject = (val: unknown): val is Record<string, unknown> => {
    return typeof val === 'object' && val !== null && !Array.isArray(val)
  }

  // NestJS Logger adds context as first arg, so check args[1] and args[2]
  if (args.length >= 3 && isPlainObject(args[0])) {
    const contextObj = args[0]
    const firstArg: unknown = args[1]
    const secondArg: unknown = args[2]

    // Case 1: message + Error
    if (typeof firstArg === 'string' && secondArg instanceof Error) {
      method.apply(this, [{ ...contextObj, err: secondArg }, firstArg, ...args.slice(3)])
      return
    }

    // Case 2: message + additional context object
    if (typeof firstArg === 'string' && isPlainObject(secondArg)) {
      method.apply(this, [{ ...contextObj, ...secondArg }, firstArg, ...args.slice(3)])
      return
    }

    // Case 3: message + stack trace string
    if (
      typeof firstArg === 'string' &&
      typeof secondArg === 'string' &&
      secondArg.includes('\n') &&
      secondArg.includes('at ')
    ) {
      method.apply(this, [{ ...contextObj, stack: secondArg }, firstArg, ...args.slice(3)])
      return
    }
  }

  // Handle case without context (direct Pino usage)
  if (args.length >= 2) {
    const firstArg: unknown = args[0]
    const secondArg: unknown = args[1]

    // Case 1: message + Error
    if (typeof firstArg === 'string' && secondArg instanceof Error) {
      method.apply(this, [{ err: secondArg }, firstArg, ...args.slice(2)])
      return
    }

    // Case 2: message + additional context object
    if (typeof firstArg === 'string' && isPlainObject(secondArg)) {
      method.apply(this, [secondArg, firstArg, ...args.slice(2)])
      return
    }

    // Case 3: message + stack trace string
    if (
      typeof firstArg === 'string' &&
      typeof secondArg === 'string' &&
      secondArg.includes('\n') &&
      secondArg.includes('at ')
    ) {
      method.apply(this, [{ stack: secondArg }, firstArg, ...args.slice(2)])
      return
    }
  }

  // Default behavior for other cases
  method.apply(this, args)
}

type LogConfig = ReturnType<typeof TypedConfigService.prototype.get<'log'>>

/*
 * Get the pino transport based on the configuration
 * @param isProduction - whether the application is in production mode
 * @param logConfig - the log configuration
 * @returns the pino transport
 */
export function getPinoTransport(
  isProduction: boolean,
  logConfig: LogConfig,
): TransportSingleOptions<Record<string, any>> {
  switch (true) {
    // if console disabled, set destination to /dev/null
    case logConfig.console.disabled:
      return {
        target: 'pino/file',
        options: {
          destination: '/dev/null',
        },
      }
    // if production mode, no transport => raw NDJSON
    case isProduction:
      return undefined
    // if non-production use pino-pretty
    default:
      return {
        target: 'pino-pretty',
        options: {
          colorize: true,
          singleLine: true,
          ignore: 'pid,hostname',
        },
      }
  }
}
