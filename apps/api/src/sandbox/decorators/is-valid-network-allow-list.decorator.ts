/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { registerDecorator, ValidationOptions, ValidatorConstraintInterface } from 'class-validator'
import { validateNetworkAllowList } from '../utils/network-validation.util'

class IsValidNetworkAllowListConstraint implements ValidatorConstraintInterface {
  private errorMessage =
    'networkAllowList must contain valid CIDR network addresses (e.g., "192.168.1.0/16,10.0.0.0/24")'

  validate(value: any): boolean {
    try {
      validateNetworkAllowList(value)
      return true
    } catch (err: any) {
      this.errorMessage = err.message
      return false
    }
  }

  defaultMessage(): string {
    return this.errorMessage
  }
}

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
      validator: IsValidNetworkAllowListConstraint,
    })
  }
}
