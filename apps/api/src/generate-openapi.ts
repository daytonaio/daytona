#!/usr/bin/env ts-node
import * as fs from 'fs'
import * as path from 'path'
import { NestFactory } from '@nestjs/core'
import { AppModule } from './app.module'
import { SwaggerModule } from '@nestjs/swagger'
import type { OpenAPIObject } from '@nestjs/swagger'
import { getOpenApiConfig } from './openapi.config'
import { addWebhookDocumentation } from './openapi-webhooks'
import {
  SandboxCreatedWebhookDto,
  SandboxStateUpdatedWebhookDto,
  SnapshotCreatedWebhookDto,
  SnapshotStateUpdatedWebhookDto,
  SnapshotRemovedWebhookDto,
  VolumeCreatedWebhookDto,
  VolumeStateUpdatedWebhookDto,
} from './webhook/dto/webhook-event-payloads.dto'

async function generateOpenAPI() {
  try {
    const app = await NestFactory.create(AppModule, {
      logger: ['error'], // Reduce logging noise
    })

    const config = getOpenApiConfig('http://localhost:3000')

    const document = {
      ...SwaggerModule.createDocument(app, config),
    }
    addEnumVarnames(document)
    const openapiPath = './dist/apps/api/openapi.json'
    fs.mkdirSync(path.dirname(openapiPath), { recursive: true })
    fs.writeFileSync(openapiPath, JSON.stringify(document, null, 2))

    // Generate 3.1.0 version of the OpenAPI specification
    // Needed for the webhook documentation
    const document_3_1_0 = {
      ...SwaggerModule.createDocument(app, config, {
        extraModels: [
          SandboxCreatedWebhookDto,
          SandboxStateUpdatedWebhookDto,
          SnapshotCreatedWebhookDto,
          SnapshotStateUpdatedWebhookDto,
          SnapshotRemovedWebhookDto,
          VolumeCreatedWebhookDto,
          VolumeStateUpdatedWebhookDto,
        ],
      }),
      openapi: '3.1.0',
    }
    addEnumVarnames(document_3_1_0)
    const documentWithWebhooks = addWebhookDocumentation(document_3_1_0)
    const openapi310Path = './dist/apps/api/openapi.3.1.0.json'
    fs.mkdirSync(path.dirname(openapi310Path), { recursive: true })
    fs.writeFileSync(openapi310Path, JSON.stringify(documentWithWebhooks, null, 2))

    await app.close()
    console.log('OpenAPI specification generated successfully!')
    clearTimeout(timeout)
    process.exit(0)
  } catch (error) {
    console.error('Failed to generate OpenAPI specification:', error)
    clearTimeout(timeout)
    process.exit(1)
  }
}

// Add timeout to prevent hanging
const timeout = setTimeout(() => {
  console.error('Generation timed out after 30 seconds')
  process.exit(1)
}, 30000)

// Clear timeout if process exits normally
process.on('exit', () => {
  clearTimeout(timeout)
})

// https://openapi-generator.tech/docs/templating/#enum
// Adds x-enum-varnames to string enums whose values contain characters that
// OpenAPI Generator would otherwise mangle (e.g. "linux-vm" -> LINUX_MINUS_VM in Python).
function addEnumVarnames(document: OpenAPIObject): void {
  const visited = new WeakSet<object>()

  const toVarname = (value: string, index: number): string => {
    const cleaned = value
      .replace(/[^A-Za-z0-9]+/g, '_')
      .replace(/^_+|_+$/g, '')
      .toUpperCase()
    if (!cleaned) return `VALUE_${index}`
    return /^\d/.test(cleaned) ? `_${cleaned}` : cleaned
  }

  const visit = (node: unknown): void => {
    if (!node || typeof node !== 'object' || visited.has(node)) return
    visited.add(node)

    const schema = node as Record<string, unknown>
    const values = schema.enum
    if (
      schema.type === 'string' &&
      Array.isArray(values) &&
      !schema['x-enum-varnames'] &&
      values.every((v): v is string => typeof v === 'string') &&
      values.some((v) => /[^A-Za-z0-9_]/.test(v))
    ) {
      const varnames = values.map(toVarname)
      if (new Set(varnames).size === varnames.length) {
        schema['x-enum-varnames'] = varnames
      }
    }

    for (const value of Object.values(schema)) visit(value)
  }

  visit(document)
}

generateOpenAPI()
