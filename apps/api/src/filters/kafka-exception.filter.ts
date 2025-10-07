/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Catch, ArgumentsHost, Logger } from '@nestjs/common'
import { KafkaContext } from '@nestjs/microservices'
import NodeCache from 'node-cache'

interface KafkaMaxRetryOptions {
  retries?: number
  sendToDlq?: boolean
  commitOffset?: boolean
}

@Catch()
export class KafkaMaxRetryExceptionFilter {
  private readonly logger = new Logger(KafkaMaxRetryExceptionFilter.name)
  private readonly maxRetries: number
  private readonly sendToDlq: boolean
  private readonly dlqTopicSuffix = '.dlq'
  private readonly commitOffset: boolean
  private readonly retryTracker: NodeCache

  constructor(options: KafkaMaxRetryOptions = {}) {
    this.maxRetries = options.retries ?? 3
    this.sendToDlq = options.sendToDlq ?? false
    this.commitOffset = options.commitOffset ?? true

    // Initialize retry tracker with 5 minutes TTL
    this.retryTracker = new NodeCache({
      stdTTL: 300,
      checkperiod: 60,
      useClones: false,
    })
  }

  async catch(exception: unknown, host: ArgumentsHost) {
    try {
      const kafkaContext = host.switchToRpc().getContext<KafkaContext>()
      const message = kafkaContext.getMessage()
      const messageKey = this.createMessageKey(kafkaContext)

      this.logger.debug('Processing message', { messageKey, offset: message.offset })

      const currentRetryCount = (this.retryTracker.get(messageKey) as number) || 0

      if (currentRetryCount >= this.maxRetries) {
        await this.handleMaxRetriesExceeded(kafkaContext, message, messageKey, currentRetryCount, exception)
        return
      }

      // Allow retry
      this.retryTracker.set(messageKey, currentRetryCount + 1, 300) // 5 minutes TTL
      this.logger.debug(`Allowing retry ${currentRetryCount + 1}/${this.maxRetries} for message ${messageKey}`)
      throw exception
    } catch (filterError) {
      this.logger.error('Error in filter:', filterError)
      throw exception
    }
  }

  private createMessageKey(context: KafkaContext): string {
    return `${context.getTopic()}-${context.getPartition()}-${context.getMessage().offset}`
  }

  private async handleMaxRetriesExceeded(
    context: KafkaContext,
    message: any,
    messageKey: string,
    retryCount: number,
    exception: unknown,
  ): Promise<void> {
    this.logger.warn(`Max retries (${this.maxRetries}) exceeded for message ${messageKey}`)

    // Clean up retry tracker
    this.retryTracker.del(messageKey)

    if (this.sendToDlq) {
      await this.sendToDLQ(context, message, retryCount, exception)
    }

    if (this.commitOffset) {
      await this.commitMessageOffset(context, messageKey)
    }
  }

  private async sendToDLQ(context: KafkaContext, message: any, retryCount: number, exception: unknown): Promise<void> {
    try {
      const producer = context.getProducer()
      if (!producer) {
        this.logger.warn('Producer not available, cannot send to DLQ')
        return
      }

      const dlqTopic = `${context.getTopic()}${this.dlqTopicSuffix}`
      const dlqMessage = this.createDLQMessage(message, retryCount, context, exception)

      await producer.send({
        topic: dlqTopic,
        messages: [dlqMessage],
      })

      this.logger.log(`Message sent to DLQ: ${dlqTopic}`)
    } catch (error) {
      this.logger.error('Failed to send message to DLQ:', error)
    }
  }

  private createDLQMessage(message: any, retryCount: number, context: KafkaContext, exception: unknown) {
    return {
      value: JSON.stringify(message.value),
      headers: {
        ...message.headers,
        'retry-count': retryCount.toString(),
        'original-topic': context.getTopic(),
        'original-offset': String(message.offset),
        'failed-at': new Date().toISOString(),
        'error-type': exception instanceof Error ? exception.constructor.name : typeof exception,
        'error-message': exception instanceof Error ? exception.message : String(exception),
        'error-stack': exception instanceof Error ? exception.stack : undefined,
      },
    }
  }

  private async commitMessageOffset(context: KafkaContext, messageKey: string): Promise<void> {
    try {
      const consumer = context.getConsumer()
      if (!consumer) {
        this.logger.warn('Consumer not available, cannot commit offset')
        return
      }

      await consumer.commitOffsets([
        {
          topic: context.getTopic(),
          partition: context.getPartition(),
          offset: String(Number(context.getMessage().offset) + 1),
        },
      ])

      this.logger.log(`Offset committed for message ${messageKey}`)
    } catch (error) {
      this.logger.error(`Failed to commit offset for message ${messageKey}:`, error)
    }
  }
}
