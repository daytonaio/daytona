/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { SandboxApi, DefaultApi, SnapshotsApi, Configuration } from '@daytonaio/runner-api-client'
import { Runner } from '../entities/runner.entity'
import { Injectable } from '@nestjs/common'
import axios from 'axios'
import axiosDebug from 'axios-debug-log'

const isDebugEnabled = process.env.DEBUG === 'true'

if (isDebugEnabled) {
  axiosDebug({
    request: function (debug, config) {
      debug('Request with ' + JSON.stringify(config))
      return config
    },
    response: function (debug, response) {
      debug('Response with ' + response)
      return response
    },
    error: function (debug, error) {
      debug('Error with ' + error)
      return Promise.reject(error)
    },
  })
}

@Injectable()
export class RunnerApiFactory {
  createRunnerApi(runner: Runner): DefaultApi {
    const axiosInstance = axios.create({
      baseURL: runner.apiUrl,
      headers: {
        Authorization: `Bearer ${runner.apiKey}`,
      },
      timeout: 1 * 60 * 60 * 1000, // 1 hour
    })

    axiosInstance.interceptors.response.use(
      (response) => {
        return response
      },
      (error) => {
        const errorMessage = error.response?.data?.message || error.response?.data || error.message || String(error)

        throw new Error(String(errorMessage))
      },
    )

    if (isDebugEnabled) {
      axiosDebug.addLogger(axiosInstance)
    }

    return new DefaultApi(new Configuration(), '', axiosInstance)
  }

  createSnapshotApi(runner: Runner): SnapshotsApi {
    const axiosInstance = axios.create({
      baseURL: runner.apiUrl,
      headers: {
        Authorization: `Bearer ${runner.apiKey}`,
      },
      timeout: 1 * 60 * 60 * 1000, // 1 hour
    })

    if (isDebugEnabled) {
      axiosDebug.addLogger(axiosInstance)
    }

    return new SnapshotsApi(new Configuration(), '', axiosInstance)
  }

  createSandboxApi(runner: Runner): SandboxApi {
    const axiosInstance = axios.create({
      baseURL: runner.apiUrl,
      headers: {
        Authorization: `Bearer ${runner.apiKey}`,
      },
      timeout: 1 * 60 * 60 * 1000, // 1 hour
    })

    if (isDebugEnabled) {
      axiosDebug.addLogger(axiosInstance)
    }

    return new SandboxApi(new Configuration(), '', axiosInstance)
  }
}
