/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable, Logger } from '@nestjs/common'
import { ClientGrpc, Transport, ClientProxyFactory } from '@nestjs/microservices'
import { Node } from '../entities/node.entity'
import { join } from 'path'
import { RunnerClient } from '@daytonaio/runner-grpc-client'
import * as grpc from '@grpc/grpc-js'

@Injectable()
export class RunnerClientFactory {
  private readonly logger = new Logger(RunnerClientFactory.name)

  create(node: Node): RunnerClient {
    // Ensure URL is properly formatted for gRPC
    const url =
      node.apiUrl.startsWith('http://') || node.apiUrl.startsWith('https://')
        ? node.apiUrl.replace(/^https?:\/\//, '')
        : node.apiUrl

    this.logger.debug(`Creating gRPC client for runner with id: ${node.id} at ${url}`)

    try {
      const client = ClientProxyFactory.create({
        transport: Transport.GRPC,
        options: {
          package: 'runner',
          protoPath: join(__dirname, 'assets/runner.proto'),
          url: url,
          credentials: grpc.credentials.createInsecure(),
        } as any,
      }) as ClientGrpc

      const service = client.getService('Runner')
      if (!service) {
        this.logger.error(`Failed to get Runner with id: ${node.id}`)
        throw new Error('Runner not found')
      }

      // Convert Observable methods to Promise-based methods
      // Authorization is attached in the proxy for each method call
      return new Proxy(service, {
        get: (target: any, prop: string) => {
          if (typeof target[prop] === 'function') {
            return async (...args: any[]) => {
              return new Promise((resolve, reject) => {
                const metadata = new grpc.Metadata()
                metadata.add('authorization', `Bearer ${node.apiKey}`)

                // The metadata must be passed as the last argument to the gRPC call
                // Check if the last argument is already metadata
                const lastArg = args[args.length - 1]
                if (lastArg && lastArg instanceof grpc.Metadata) {
                  // If metadata already exists, add our authorization to it
                  lastArg.add('authorization', `Bearer ${node.apiKey}`)
                } else {
                  // If no metadata exists, add our metadata as the last argument
                  args.push(metadata)
                }

                const observable = target[prop](...args)
                observable.subscribe({
                  next: (value: any) => resolve(value),
                  error: (error: any) => reject(error),
                  complete: () => {
                    resolve(undefined)
                  },
                })
              })
            }
          }
          return target[prop]
        },
      }) as RunnerClient
    } catch (error) {
      this.logger.error(`Failed to create gRPC client for runner with id: ${node.id}: ${error.message}`)
      throw error
    }
  }
}
