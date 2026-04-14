/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { JobController } from './job.controller'
import { RunnerAuthContextGuard } from '../guards/runner-auth-context.guard'
import { JobAccessGuard } from '../guards/job-access.guard'
import { AuthStrategyType } from '../../auth/enums/auth-strategy-type.enum'
import {
  getAuthContextGuards,
  getAllowedAuthStrategies,
  getResourceAccessGuards,
  expectArrayMatch,
  createCoverageTracker,
  isPublicEndpoint,
} from '../../test/helpers/controller-metadata.helper'

describe('[AUTH] JobController', () => {
  const trackMethod = createCoverageTracker(JobController)

  it('listJobs', () => {
    const methodName = trackMethod('listJobs')
    expect(isPublicEndpoint(JobController, methodName)).toBe(false)
    expectArrayMatch(getAllowedAuthStrategies(JobController, methodName), [AuthStrategyType.API_KEY])
    expectArrayMatch(getAuthContextGuards(JobController, methodName), [RunnerAuthContextGuard])
  })

  it('pollJobs', () => {
    const methodName = trackMethod('pollJobs')
    expect(isPublicEndpoint(JobController, methodName)).toBe(false)
    expectArrayMatch(getAllowedAuthStrategies(JobController, methodName), [AuthStrategyType.API_KEY])
    expectArrayMatch(getAuthContextGuards(JobController, methodName), [RunnerAuthContextGuard])
  })

  it('getJob', () => {
    const methodName = trackMethod('getJob')
    expect(isPublicEndpoint(JobController, methodName)).toBe(false)
    expectArrayMatch(getAllowedAuthStrategies(JobController, methodName), [AuthStrategyType.API_KEY])
    expectArrayMatch(getAuthContextGuards(JobController, methodName), [RunnerAuthContextGuard])
    expectArrayMatch(getResourceAccessGuards(JobController, methodName), [JobAccessGuard])
  })

  it('updateJobStatus', () => {
    const methodName = trackMethod('updateJobStatus')
    expect(isPublicEndpoint(JobController, methodName)).toBe(false)
    expectArrayMatch(getAllowedAuthStrategies(JobController, methodName), [AuthStrategyType.API_KEY])
    expectArrayMatch(getAuthContextGuards(JobController, methodName), [RunnerAuthContextGuard])
    expectArrayMatch(getResourceAccessGuards(JobController, methodName), [JobAccessGuard])
  })
})
