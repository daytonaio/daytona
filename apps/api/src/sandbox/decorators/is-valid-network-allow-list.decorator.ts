/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { registerDecorator, ValidationOptions, ValidationArguments } from 'class-validator'
import { validateNetworkAllowList } from '../utils/network-validation.util'

/**
 * Custom class-validator decorator that validates network allow list using the validateNetworkAllowList utility
 * @param validationOptions - Optional validation options from class-validator
 */
export function IsValidNetworkAllowList(validationOptions?: ValidationOptions) {
  return function (object: object, propertyName: string) {
    registerDecorator({
      name: 'isValidNetworkAllowList',
      target: object.constructor,
      propertyName: propertyName,
      options: validationOptions,
      validator: {
        validate(value: any, args: ValidationArguments) {
          // Allow empty/null/undefined values
          if (!value) {
            return true
          }

          // Ensure value is a string
          if (typeof value !== 'string') {
            return false
          }

          // Use the existing validation utility
          const validationResult = validateNetworkAllowList(value)
          return validationResult === null
        },
        defaultMessage(args: ValidationArguments) {
          const value = args.value

          if (typeof value !== 'string') {
            return 'networkAllowList must be a string'
          }

          return 'networkAllowList must contain valid /24 CIDR blocks (e.g., "192.168.1.0/24,10.0.0.0/24")'
        },
      },
    })
  }
}
