/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { DocumentBuilder } from '@nestjs/swagger'

const getOpenApiConfig = (oidcIssuer: string) =>
  new DocumentBuilder()
    .setTitle('Daytona')
    .addServer('http://localhost:3000')
    .setDescription('Daytona AI platform API Docs')
    .setContact('Daytona Platforms Inc.', 'https://www.daytona.io', 'support@daytona.com')
    .setVersion('1.0')
    .addBearerAuth({
      type: 'http',
      scheme: 'bearer',
      description: 'API Key access',
    })
    .addOAuth2({
      type: 'openIdConnect',
      flows: undefined,
      openIdConnectUrl: `${oidcIssuer}/.well-known/openid-configuration`,
    })
    .build()

export { getOpenApiConfig }
