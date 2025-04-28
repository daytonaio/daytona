#!/usr/bin/env ts-node
import * as fs from 'fs'
import { NestFactory } from '@nestjs/core'
import { AppModule } from './app.module'
import { SwaggerModule } from '@nestjs/swagger'
import { getOpenApiConfig } from './openapi.config'

async function generateOpenAPI() {
  const app = await NestFactory.create(AppModule, {
    logger: ['error'], // Reduce logging noise
  })

  const config = getOpenApiConfig('http://localhost:3000')
  const document = SwaggerModule.createDocument(app, config)
  fs.writeFileSync('./dist/apps/api/openapi.json', JSON.stringify(document, null, 2))

  await app.close()
  console.log('OpenAPI specification generated successfully!')
  process.exit(0)
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

generateOpenAPI()
