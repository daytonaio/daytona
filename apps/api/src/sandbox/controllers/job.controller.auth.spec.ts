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
} from '../../test/helpers/controller-metadata.helper'

describe('[AUTH] JobController', () => {
  const trackMethod = createCoverageTracker(JobController)

  it('listJobs', () => {
    const methodName = trackMethod('listJobs')
    expectArrayMatch(getAllowedAuthStrategies(JobController, methodName), [AuthStrategyType.API_KEY])
    expectArrayMatch(getAuthContextGuards(JobController, methodName), [RunnerAuthContextGuard])
  })

  it('pollJobs', () => {
    const methodName = trackMethod('pollJobs')
    expectArrayMatch(getAllowedAuthStrategies(JobController, methodName), [AuthStrategyType.API_KEY])
    expectArrayMatch(getAuthContextGuards(JobController, methodName), [RunnerAuthContextGuard])
  })

  it('getJob', () => {
    const methodName = trackMethod('getJob')
    expectArrayMatch(getAllowedAuthStrategies(JobController, methodName), [AuthStrategyType.API_KEY])
    expectArrayMatch(getAuthContextGuards(JobController, methodName), [RunnerAuthContextGuard])
    expectArrayMatch(getResourceAccessGuards(JobController, methodName), [JobAccessGuard])
  })

  it('updateJobStatus', () => {
    const methodName = trackMethod('updateJobStatus')
    expectArrayMatch(getAllowedAuthStrategies(JobController, methodName), [AuthStrategyType.API_KEY])
    expectArrayMatch(getAuthContextGuards(JobController, methodName), [RunnerAuthContextGuard])
    expectArrayMatch(getResourceAccessGuards(JobController, methodName), [JobAccessGuard])
  })
})
