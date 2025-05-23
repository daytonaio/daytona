/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable } from '@nestjs/common'
import { ClientGrpc, Transport, ClientProxyFactory } from '@nestjs/microservices'
import { Node } from '../entities/node.entity'
import { join } from 'path'
import { RunnerClient } from '@daytonaio/runner-grpc-client'
import * as grpc from '@grpc/grpc-js'

@Injectable()
export class RunnerClientFactory {
  create(node: Node): RunnerClient {
    const client = ClientProxyFactory.create({
      transport: Transport.GRPC,
      options: {
        package: 'runner',
        protoPath: join(__dirname, '../../proto/runner.proto'),
        url: node.apiUrl,
        loader: {
          keepCase: true,
          longs: String,
          enums: String,
          defaults: true,
          oneofs: true,
        },
        credentials: grpc.credentials.createInsecure(),
        metadata: { authorization: `Bearer ${node.apiKey}` },
        channelOptions: {
          'grpc.keepalive_time_ms': 10000,
          'grpc.keepalive_timeout_ms': 5000,
          'grpc.http2.min_time_between_pings_ms': 10000,
          'grpc.keepalive_permit_without_calls': 1,
          'grpc.max_receive_message_length': -1,
          'grpc.max_send_message_length': -1,
        },
        retryOptions: {
          retries: 3,
          initialRetryDelay: 1000,
          maxRetryDelay: 5000,
          retryDelayMultiplier: 1.5,
        },
      } as any,
    }) as ClientGrpc

    return client.getService('Runner') as RunnerClient
  }
}
