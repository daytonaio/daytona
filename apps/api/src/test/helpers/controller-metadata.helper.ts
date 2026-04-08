/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import 'reflect-metadata'
import { CanActivate, Type } from '@nestjs/common'
import { PATH_METADATA } from '@nestjs/common/constants'
import { AuthContextGuard } from '../../common/guards/auth-context.guard'
import { ResourceAccessGuard } from '../../common/guards/resource-access.guard'
import { AuthStrategy } from '../../auth/decorators/auth-strategy.decorator'
import { AuthStrategyType } from '../../auth/enums/auth-strategy-type.enum'
import { RequiredSystemRole } from '../../user/decorators/required-system-role.decorator'
import { SystemRole } from '../../user/enums/system-role.enum'
import { RequiredOrganizationMemberRole } from '../../organization/decorators/required-organization-member-role.decorator'
import { OrganizationMemberRole } from '../../organization/enums/organization-member-role.enum'
import { RequiredOrganizationResourcePermissions } from '../../organization/decorators/required-organization-resource-permissions.decorator'
import { OrganizationResourcePermission } from '../../organization/enums/organization-resource-permission.enum'

const GUARDS_METADATA_KEY = '__guards__'

type GuardType = Type<CanActivate>
type ControllerType = Type<object>

/**
 * Gets the guards defined at the class level.
 */
function getClassLevelGuards(controller: ControllerType): GuardType[] {
  return (Reflect.getMetadata(GUARDS_METADATA_KEY, controller) as GuardType[] | undefined) ?? []
}

/**
 * Gets the guards defined at the method level.
 */
function getMethodLevelGuards(controller: ControllerType, methodName: string): GuardType[] {
  const method = (controller.prototype as Record<string, unknown>)[methodName]
  if (typeof method !== 'function') return []
  return (Reflect.getMetadata(GUARDS_METADATA_KEY, method) as GuardType[] | undefined) ?? []
}

/**
 * Gets the effective guards for a controller method or class. Combines the class-level and method-level guards.
 */
function getEffectiveGuards(controller: ControllerType, methodName?: string): GuardType[] {
  const classGuards = getClassLevelGuards(controller)
  const methodGuards = methodName ? getMethodLevelGuards(controller, methodName) : []
  const guards = [...classGuards, ...methodGuards]
  return guards.flatMap((g) => ('guards' in g ? (g.guards as GuardType[]) : [g]))
}

/**
 * Asserts that a method exists on a controller.
 */
function assertMethodExists(controller: ControllerType, methodName: string): void {
  const method = (controller.prototype as Record<string, unknown>)[methodName]
  if (typeof method !== 'function') {
    throw new Error(`Method '${methodName}' does not exist on ${controller.name}.`)
  }
}

/**
 * Gets the auth context guards for a controller method or class.
 *
 * Checks method-level metadata first, then falls back to class-level.
 */
export function getAuthContextGuards(controller: ControllerType): GuardType[]
export function getAuthContextGuards<T extends object>(controller: Type<T>, methodName: keyof T & string): GuardType[]
export function getAuthContextGuards(controller: ControllerType, methodName?: string): GuardType[] {
  if (methodName) assertMethodExists(controller, methodName)
  return getEffectiveGuards(controller, methodName).filter((g) => g.prototype instanceof AuthContextGuard)
}

/**
 * Gets the resource access guards for a controller method or class.
 *
 * Checks method-level metadata first, then falls back to class-level.
 */
export function getResourceAccessGuards(controller: ControllerType): GuardType[]
export function getResourceAccessGuards<T extends object>(
  controller: Type<T>,
  methodName: keyof T & string,
): GuardType[]
export function getResourceAccessGuards(controller: ControllerType, methodName?: string): GuardType[] {
  if (methodName) assertMethodExists(controller, methodName)
  return getEffectiveGuards(controller, methodName).filter((g) => g.prototype instanceof ResourceAccessGuard)
}

/**
 * Gets the allowed auth strategies for a controller method or class.
 *
 * Checks method-level metadata first, then falls back to class-level.
 */
export function getAllowedAuthStrategies(controller: ControllerType): AuthStrategyType[]
export function getAllowedAuthStrategies<T extends object>(
  controller: Type<T>,
  methodName: keyof T & string,
): AuthStrategyType[]
export function getAllowedAuthStrategies(controller: ControllerType, methodName?: string): AuthStrategyType[] {
  if (methodName) {
    assertMethodExists(controller, methodName)
    const method = (controller.prototype as Record<string, unknown>)[methodName]
    if (typeof method === 'function') {
      const methodStrategy = Reflect.getMetadata(AuthStrategy.KEY, method) as
        | AuthStrategyType
        | AuthStrategyType[]
        | undefined
      if (methodStrategy !== undefined) {
        return Array.isArray(methodStrategy) ? methodStrategy : [methodStrategy]
      }
    }
  }

  const classStrategy = Reflect.getMetadata(AuthStrategy.KEY, controller) as
    | AuthStrategyType
    | AuthStrategyType[]
    | undefined
  if (classStrategy !== undefined) {
    return Array.isArray(classStrategy) ? classStrategy : [classStrategy]
  }

  return []
}

/**
 * Gets the required system role(s) for a controller method or class.
 *
 * Checks method-level metadata first, then falls back to class-level.
 */
export function getRequiredSystemRole(controller: ControllerType): SystemRole | SystemRole[] | undefined
export function getRequiredSystemRole<T extends object>(
  controller: Type<T>,
  methodName: keyof T & string,
): SystemRole | SystemRole[] | undefined
export function getRequiredSystemRole(
  controller: ControllerType,
  methodName?: string,
): SystemRole | SystemRole[] | undefined {
  if (methodName) {
    assertMethodExists(controller, methodName)
    const method = (controller.prototype as Record<string, unknown>)[methodName]
    if (typeof method === 'function') {
      const methodRole = Reflect.getMetadata(RequiredSystemRole.KEY, method) as SystemRole | SystemRole[] | undefined
      if (methodRole !== undefined) {
        return methodRole
      }
    }
  }

  return Reflect.getMetadata(RequiredSystemRole.KEY, controller) as SystemRole | SystemRole[] | undefined
}

/**
 * Gets the required organization member role for a controller method or class.
 *
 * Checks method-level metadata first, then falls back to class-level.
 */
export function getRequiredOrganizationMemberRole(controller: ControllerType): OrganizationMemberRole | undefined
export function getRequiredOrganizationMemberRole<T extends object>(
  controller: Type<T>,
  methodName: keyof T & string,
): OrganizationMemberRole | undefined
export function getRequiredOrganizationMemberRole(
  controller: ControllerType,
  methodName?: string,
): OrganizationMemberRole | undefined {
  if (methodName) {
    assertMethodExists(controller, methodName)
    const method = (controller.prototype as Record<string, unknown>)[methodName]
    if (typeof method === 'function') {
      const methodRole = Reflect.getMetadata(RequiredOrganizationMemberRole.KEY, method) as
        | OrganizationMemberRole
        | undefined
      if (methodRole !== undefined) {
        return methodRole
      }
    }
  }

  return Reflect.getMetadata(RequiredOrganizationMemberRole.KEY, controller) as OrganizationMemberRole | undefined
}

/**
 * Gets the required organization resource permissions for a controller method or class.
 *
 * Checks method-level metadata first, then falls back to class-level.
 */
export function getRequiredOrganizationResourcePermissions(
  controller: ControllerType,
): OrganizationResourcePermission[] | undefined
export function getRequiredOrganizationResourcePermissions<T extends object>(
  controller: Type<T>,
  methodName: keyof T & string,
): OrganizationResourcePermission[] | undefined
export function getRequiredOrganizationResourcePermissions(
  controller: ControllerType,
  methodName?: string,
): OrganizationResourcePermission[] | undefined {
  if (methodName) {
    assertMethodExists(controller, methodName)
    const method = (controller.prototype as Record<string, unknown>)[methodName]
    if (typeof method === 'function') {
      const methodPermissions = Reflect.getMetadata(RequiredOrganizationResourcePermissions.KEY, method) as
        | OrganizationResourcePermission[]
        | undefined
      if (methodPermissions !== undefined) {
        return methodPermissions
      }
    }
  }

  return Reflect.getMetadata(RequiredOrganizationResourcePermissions.KEY, controller) as
    | OrganizationResourcePermission[]
    | undefined
}

/**
 * Gets the names of the route handlers on a controller.
 */
function getRouteHandlerNames(controller: ControllerType): string[] {
  return Object.getOwnPropertyNames(controller.prototype).filter((name) => {
    if (name === 'constructor') return false
    const method = (controller.prototype as Record<string, unknown>)[name]
    return typeof method === 'function' && Reflect.getMetadata(PATH_METADATA, method) !== undefined
  })
}

/**
 * Registers an afterAll hook that fails if any route handler on the controller was not tested.
 */
export function createCoverageTracker<T extends object>(controller: Type<T>) {
  const tested = new Set<string>()

  afterAll(() => {
    const handlers = getRouteHandlerNames(controller)
    const untested = handlers.filter((h) => !tested.has(h))
    if (untested.length > 0) {
      throw new Error(`Untested route handlers on ${controller.name}: ${untested.join(', ')}.`)
    }
  })

  return function trackMethod(methodName: keyof T & string): keyof T & string {
    tested.add(methodName)
    return methodName
  }
}

/**
 * Asserts the list of actual values matches the expected values exactly (order-independent).
 *
 * Used to assert that the actual list of guards or allowed auth strategies matches the corresponding expected list.
 */
export function expectArrayMatch<T>(actual: T[], expected: T[]): void {
  expect(actual).toHaveLength(expected.length)
  expect(actual).toEqual(expect.arrayContaining(expected))
}
