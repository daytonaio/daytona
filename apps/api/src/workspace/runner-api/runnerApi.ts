/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable } from '@nestjs/common'
import { ClientGrpc, Transport, ClientProxyFactory } from '@nestjs/microservices'
import { Node } from '../entities/node.entity'
import { join } from 'path'
import { RunnerClient } from '@daytonaio/runner-grpc-client'
@Injectable()
export class RunnerClientFactory {
  create(node: Node): RunnerClient {
    const client = ClientProxyFactory.create({
      transport: Transport.GRPC,
      options: {
        url: node.apiUrl,
        package: 'runner',
        protoPath: join(__dirname, 'apps/api/runner-grpc/proto/runner.proto'),
        credentials: this.getCredentials(node.apiKey),
      },
    }) as ClientGrpc

    return client.getService('Runner') as RunnerClient
  }

  private getCredentials(apiKey: string) {
    return { metadata: { authorization: `Bearer ${apiKey}` } }
  }
}
