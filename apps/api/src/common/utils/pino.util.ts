/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import pino from 'pino'

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
