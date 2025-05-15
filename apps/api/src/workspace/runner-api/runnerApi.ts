/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { SandboxApi, DefaultApi, ImagesApi, Configuration } from '@daytonaio/runner-api-client'
import { Node } from '../entities/node.entity'
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
export class NodeApiFactory {
  createNodeApi(node: Node): DefaultApi {
    const axiosInstance = axios.create({
      baseURL: node.apiUrl,
      headers: {
        Authorization: `Bearer ${node.apiKey}`,
      },
      timeout: 1 * 60 * 60 * 1000, // 1 hour
    })

    if (isDebugEnabled) {
      axiosDebug.addLogger(axiosInstance)
    }

    return new DefaultApi(new Configuration(), '', axiosInstance)
  }

  createImageApi(node: Node): ImagesApi {
    const axiosInstance = axios.create({
      baseURL: node.apiUrl,
      headers: {
        Authorization: `Bearer ${node.apiKey}`,
      },
      timeout: 1 * 60 * 60 * 1000, // 1 hour
    })

    if (isDebugEnabled) {
      axiosDebug.addLogger(axiosInstance)
    }

    return new ImagesApi(new Configuration(), '', axiosInstance)
  }

  createWorkspaceApi(node: Node): SandboxApi {
    const axiosInstance = axios.create({
      baseURL: node.apiUrl,
      headers: {
        Authorization: `Bearer ${node.apiKey}`,
      },
      timeout: 1 * 60 * 60 * 1000, // 1 hour
    })

    if (isDebugEnabled) {
      axiosDebug.addLogger(axiosInstance)
    }

    return new SandboxApi(new Configuration(), '', axiosInstance)
  }
}
