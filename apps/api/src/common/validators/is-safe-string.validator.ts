/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import {
  registerDecorator,
  ValidationOptions,
  ValidatorConstraint,
  ValidatorConstraintInterface,
} from 'class-validator'

// Matches opening or closing HTML tags: <div>, </div>, <img />, etc.
const HTML_TAG_PATTERN = /<[a-zA-Z/][^>]*>/

// Matches URL schemes and www prefix
const URL_PATTERN =
  /https?:\/\/|ftps?:\/\/|sftp:\/\/|ssh:\/\/|wss?:\/\/|file:\/\/|ldaps?:\/\/|javascript:|data:|mailto:|tel:|www\./i

// Matches control characters except tab (\x09), line feed (\x0A), and carriage return (\x0D)
// eslint-disable-next-line no-control-regex
const CONTROL_CHAR_PATTERN = /[\x00-\x08\x0B\x0C\x0E-\x1F\x7F]/

@ValidatorConstraint({ async: false })
export class IsSafeDisplayStringConstraint implements ValidatorConstraintInterface {
  validate(value: unknown): boolean {
    if (value === undefined || value === null) {
      return true
    }

    if (typeof value !== 'string') {
      return false
    }

    if (HTML_TAG_PATTERN.test(value)) {
      return false
    }

    if (URL_PATTERN.test(value)) {
      return false
    }

    if (CONTROL_CHAR_PATTERN.test(value)) {
      return false
    }

    return true
  }

  defaultMessage(args?: import('class-validator').ValidationArguments): string {
    const field = args?.property ?? 'Value'
    return `${field} must not contain HTML tags, URLs, or control characters`
  }
}

/**
 * Rejects HTML tags, URL schemes, and control characters.
 * Intended for user-facing display fields (names, descriptions) that are
 * rendered to other users in emails, dashboard UI, and other shared contexts.
 * NOT a general-purpose string safety validator.
 */
export function IsSafeDisplayString(validationOptions?: ValidationOptions): PropertyDecorator {
  return function (object: object, propertyName: string | symbol) {
    registerDecorator({
      target: object.constructor,
      propertyName: propertyName as string,
      options: validationOptions,
      constraints: [],
      validator: IsSafeDisplayStringConstraint,
    })
  }
}
