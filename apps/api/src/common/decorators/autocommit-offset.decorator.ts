/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { KafkaContext } from '@nestjs/microservices'

/**
 * Auto commit offset decorator for Kafka messages. The offset is committed only if the method completes successfully.
 * @returns A decorator function that commits the offset of the Kafka message.
 */
export function AutoCommitOffset(): MethodDecorator {
  return function (target: any, propertyKey: string | symbol, descriptor: PropertyDescriptor): PropertyDescriptor {
    const originalMethod = descriptor.value

    descriptor.value = async function (...args: any[]) {
      const result = await originalMethod.apply(this, args)

      // Find KafkaContext in arguments
      const context = args.find((arg) => arg instanceof KafkaContext)

      if (context) {
        const message = context.getMessage()
        const partition = context.getPartition()
        const topic = context.getTopic()
        const consumer = context.getConsumer()

        await consumer.commitOffsets([
          {
            topic,
            partition,
            offset: String(Number(message.offset) + 1),
          },
        ])
      }

      return result
    }

    return descriptor
  }
}
