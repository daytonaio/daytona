/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

const configuration = {
  production: process.env.NODE_ENV === 'production',
  environment: process.env.ENVIRONMENT,
  port: parseInt(process.env.PORT, 10),
  appUrl: process.env.APP_URL,
  database: {
    host: process.env.DB_HOST,
    port: parseInt(process.env.DB_PORT || '5432', 10),
    username: process.env.DB_USERNAME,
    password: process.env.DB_PASSWORD,
    database: process.env.DB_DATABASE,
  },
  redis: {
    host: process.env.REDIS_HOST,
    port: parseInt(process.env.REDIS_PORT || '6379', 10),
    tls: process.env.REDIS_TLS === 'true' ? {} : undefined,
  },
  posthog: {
    apiKey: process.env.POSTHOG_API_KEY,
    host: process.env.POSTHOG_HOST,
  },
  oidc: {
    clientId: process.env.OIDC_CLIENT_ID || process.env.OID_CLIENT_ID,
    issuer: process.env.OIDC_ISSUER_BASE_URL || process.env.OID_ISSUER_BASE_URL,
    audience: process.env.OIDC_AUDIENCE || process.env.OID_AUDIENCE,
  },
  smtp: {
    host: process.env.SMTP_HOST,
    port: parseInt(process.env.SMTP_PORT || '587', 10),
    user: process.env.SMTP_USER,
    password: process.env.SMTP_PASSWORD,
    secure: process.env.SMTP_SECURE === 'true',
    from: process.env.SMTP_EMAIL_FROM || 'noreply@mail.daytona.io',
  },
  defaultImage: process.env.DEFAULT_IMAGE,
  dashboardUrl: process.env.DASHBOARD_URL,
  transientRegistry: {
    url: process.env.TRANSIENT_REGISTRY_URL,
    admin: process.env.TRANSIENT_REGISTRY_ADMIN,
    password: process.env.TRANSIENT_REGISTRY_PASSWORD,
    projectId: process.env.TRANSIENT_REGISTRY_PROJECT_ID,
  },
  internalRegistry: {
    url: process.env.INTERNAL_REGISTRY_URL,
    admin: process.env.INTERNAL_REGISTRY_ADMIN,
    password: process.env.INTERNAL_REGISTRY_PASSWORD,
    projectId: process.env.INTERNAL_REGISTRY_PROJECT_ID,
  },
  s3: {
    endpoint: process.env.S3_ENDPOINT,
    region: process.env.S3_REGION,
    accessKey: process.env.S3_ACCESS_KEY,
    secretKey: process.env.S3_SECRET_KEY,
    defaultBucket: process.env.S3_DEFAULT_BUCKET,
  },
  skipConnections: process.env.SKIP_CONNECTIONS === 'true',
}

export { configuration }
