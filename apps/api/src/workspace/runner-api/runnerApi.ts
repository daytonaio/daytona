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
        protoPath: '/workspaces/daytona/apps/proto/runner.proto',
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
      } as any, // Type assertion to bypass type checking temporarily
    }) as ClientGrpc

    return client.getService('Runner') as RunnerClient
  }
}
