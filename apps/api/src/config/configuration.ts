/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { SandboxClass } from '../sandbox/enums/sandbox-class.enum'

const configuration = {
  production: process.env.NODE_ENV === 'production',
  version: process.env.VERSION || '0.0.0-dev',
  environment: process.env.ENVIRONMENT,
  runMigrations: process.env.RUN_MIGRATIONS === 'true',
  port: parseInt(process.env.PORT, 10),
  appUrl: process.env.APP_URL,
  database: {
    host: process.env.DB_HOST,
    port: parseInt(process.env.DB_PORT || '5432', 10),
    username: process.env.DB_USERNAME,
    password: process.env.DB_PASSWORD,
    database: process.env.DB_DATABASE,
    tls: {
      enabled: process.env.DB_TLS_ENABLED === 'true',
      rejectUnauthorized: process.env.DB_TLS_REJECT_UNAUTHORIZED !== 'false',
    },
  },
  redis: {
    host: process.env.REDIS_HOST,
    port: parseInt(process.env.REDIS_PORT || '6379', 10),
    tls: process.env.REDIS_TLS === 'true' ? {} : undefined,
  },
  posthog: {
    apiKey: process.env.POSTHOG_API_KEY,
    host: process.env.POSTHOG_HOST,
    environment: process.env.POSTHOG_ENVIRONMENT,
  },
  oidc: {
    clientId: process.env.OIDC_CLIENT_ID || process.env.OID_CLIENT_ID,
    issuer: process.env.OIDC_ISSUER_BASE_URL || process.env.OID_ISSUER_BASE_URL,
    publicIssuer: process.env.PUBLIC_OIDC_DOMAIN,
    audience: process.env.OIDC_AUDIENCE || process.env.OID_AUDIENCE,
    managementApi: {
      enabled: process.env.OIDC_MANAGEMENT_API_ENABLED === 'true',
      clientId: process.env.OIDC_MANAGEMENT_API_CLIENT_ID,
      clientSecret: process.env.OIDC_MANAGEMENT_API_CLIENT_SECRET,
      audience: process.env.OIDC_MANAGEMENT_API_AUDIENCE,
    },
  },
  smtp: {
    host: process.env.SMTP_HOST,
    port: parseInt(process.env.SMTP_PORT || '587', 10),
    user: process.env.SMTP_USER,
    password: process.env.SMTP_PASSWORD,
    secure: process.env.SMTP_SECURE === 'true',
    from: process.env.SMTP_EMAIL_FROM || 'noreply@mail.daytona.io',
  },
  defaultSnapshot: process.env.DEFAULT_SNAPSHOT,
  dashboardUrl: process.env.DASHBOARD_URL,
  // Default to empty string - dashboard will then hit '/api'
  dashboardBaseApiUrl: process.env.DASHBOARD_BASE_API_URL || '',
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
    stsEndpoint: process.env.S3_STS_ENDPOINT,
    region: process.env.S3_REGION,
    accessKey: process.env.S3_ACCESS_KEY,
    secretKey: process.env.S3_SECRET_KEY,
    defaultBucket: process.env.S3_DEFAULT_BUCKET,
    accountId: process.env.S3_ACCOUNT_ID,
    roleName: process.env.S3_ROLE_NAME,
  },
  skipConnections: process.env.SKIP_CONNECTIONS === 'true',
  maxAutoArchiveInterval: parseInt(process.env.MAX_AUTO_ARCHIVE_INTERVAL || '43200', 10),
  maintananceMode: process.env.MAINTENANCE_MODE === 'true',
  disableCronJobs: process.env.DISABLE_CRON_JOBS === 'true',
  appRole: process.env.APP_ROLE || 'all',
  proxy: {
    domain: process.env.PROXY_DOMAIN,
    protocol: process.env.PROXY_PROTOCOL,
    apiKey: process.env.PROXY_API_KEY,
    templateUrl: process.env.PROXY_TEMPLATE_URL,
    toolboxUrl:
      (process.env.PROXY_TOOLBOX_BASE_URL || `${process.env.PROXY_PROTOCOL}://${process.env.PROXY_DOMAIN}`) +
      '/toolbox',
  },
  audit: {
    toolboxRequestsEnabled: process.env.AUDIT_TOOLBOX_REQUESTS_ENABLED === 'true',
    retentionDays: process.env.AUDIT_LOG_RETENTION_DAYS
      ? parseInt(process.env.AUDIT_LOG_RETENTION_DAYS, 10)
      : undefined,
    consoleLogEnabled: process.env.AUDIT_CONSOLE_LOG_ENABLED === 'true',
    publish: {
      enabled: process.env.AUDIT_PUBLISH_ENABLED === 'true',
      batchSize: process.env.AUDIT_PUBLISH_BATCH_SIZE ? parseInt(process.env.AUDIT_PUBLISH_BATCH_SIZE, 10) : 1000,
      mode: (process.env.AUDIT_PUBLISH_MODE || 'direct') as 'direct' | 'kafka',
      storageAdapter: process.env.AUDIT_PUBLISH_STORAGE_ADAPTER || 'opensearch',
      opensearchIndexName: process.env.AUDIT_PUBLISH_OPENSEARCH_INDEX_NAME || 'audit-logs',
    },
  },
  kafka: {
    enabled: process.env.KAFKA_ENABLED === 'true',
    brokers: process.env.KAFKA_BROKERS || 'localhost:9092',
    clientId: process.env.KAFKA_CLIENT_ID,
    sasl: {
      mechanism: process.env.KAFKA_SASL_MECHANISM,
      username: process.env.KAFKA_SASL_USERNAME,
      password: process.env.KAFKA_SASL_PASSWORD,
    },
    tls: {
      enabled: process.env.KAFKA_TLS_ENABLED === 'true',
      rejectUnauthorized: process.env.KAFKA_TLS_REJECT_UNAUTHORIZED !== 'false',
    },
  },
  opensearch: {
    nodes: process.env.OPENSEARCH_NODES || 'https://localhost:9200',
    username: process.env.OPENSEARCH_USERNAME,
    password: process.env.OPENSEARCH_PASSWORD,
    aws: {
      roleArn: process.env.OPENSEARCH_AWS_ROLE_ARN,
      region: process.env.OPENSEARCH_AWS_REGION,
    },
    tls: {
      rejectUnauthorized: process.env.OPENSEARCH_TLS_REJECT_UNAUTHORIZED !== 'false',
    },
  },
  cronTimeZone: process.env.CRON_TIMEZONE,
  maxConcurrentBackupsPerRunner: parseInt(process.env.MAX_CONCURRENT_BACKUPS_PER_RUNNER || '6', 10),
  webhook: {
    authToken: process.env.SVIX_AUTH_TOKEN,
    serverUrl: process.env.SVIX_SERVER_URL,
  },
  sshGateway: {
    apiKey: process.env.SSH_GATEWAY_API_KEY,
    command: process.env.SSH_GATEWAY_COMMAND,
    publicKey: process.env.SSH_GATEWAY_PUBLIC_KEY,
  },
  organizationSandboxDefaultLimitedNetworkEgress:
    process.env.ORGANIZATION_SANDBOX_DEFAULT_LIMITED_NETWORK_EGRESS === 'true',
  pylonAppId: process.env.PYLON_APP_ID,
  billingApiUrl: process.env.BILLING_API_URL,
  defaultRunner: {
    domain: process.env.DEFAULT_RUNNER_DOMAIN,
    apiKey: process.env.DEFAULT_RUNNER_API_KEY,
    proxyUrl: process.env.DEFAULT_RUNNER_PROXY_URL,
    apiUrl: process.env.DEFAULT_RUNNER_API_URL,
    cpu: parseInt(process.env.DEFAULT_RUNNER_CPU || '4', 10),
    memory: parseInt(process.env.DEFAULT_RUNNER_MEMORY || '8', 10),
    disk: parseInt(process.env.DEFAULT_RUNNER_DISK || '50', 10),
    gpu: parseInt(process.env.DEFAULT_RUNNER_GPU || '0', 10),
    gpuType: process.env.DEFAULT_RUNNER_GPU_TYPE,
    class: process.env.DEFAULT_RUNNER_CLASS ? (process.env.DEFAULT_RUNNER_CLASS as SandboxClass) : undefined,
    version: process.env.DEFAULT_RUNNER_VERSION || '0',
  },
  runnerUsage: {
    declarativeBuildScoreThreshold: parseInt(process.env.RUNNER_DECLARATIVE_BUILD_SCORE_THRESHOLD || '10', 10),
    availabilityScoreThreshold: parseInt(process.env.RUNNER_AVAILABILITY_SCORE_THRESHOLD || '10', 10),
    startScoreThreshold: parseInt(process.env.RUNNER_START_SCORE_THRESHOLD || '3', 10),
    cpuUsageWeight: parseFloat(process.env.RUNNER_CPU_USAGE_WEIGHT || '0.25'),
    memoryUsageWeight: parseFloat(process.env.RUNNER_MEMORY_USAGE_WEIGHT || '0.4'),
    diskUsageWeight: parseFloat(process.env.RUNNER_DISK_USAGE_WEIGHT || '0.4'),
    allocatedCpuWeight: parseFloat(process.env.RUNNER_ALLOCATED_CPU_WEIGHT || '0.03'),
    allocatedMemoryWeight: parseFloat(process.env.RUNNER_ALLOCATED_MEMORY_WEIGHT || '0.03'),
    allocatedDiskWeight: parseFloat(process.env.RUNNER_ALLOCATED_DISK_WEIGHT || '0.03'),
    cpuPenaltyExponent: parseFloat(process.env.RUNNER_CPU_PENALTY_EXPONENT || '0.15'),
    memoryPenaltyExponent: parseFloat(process.env.RUNNER_MEMORY_PENALTY_EXPONENT || '0.15'),
    diskPenaltyExponent: parseFloat(process.env.RUNNER_DISK_PENALTY_EXPONENT || '0.15'),
    cpuPenaltyThreshold: parseInt(process.env.RUNNER_CPU_PENALTY_THRESHOLD || '90', 10),
    memoryPenaltyThreshold: parseInt(process.env.RUNNER_MEMORY_PENALTY_THRESHOLD || '75', 10),
    diskPenaltyThreshold: parseInt(process.env.RUNNER_DISK_PENALTY_THRESHOLD || '75', 10),
  },
  rateLimit: {
    anonymous: {
      ttl: process.env.RATE_LIMIT_ANONYMOUS_TTL ? parseInt(process.env.RATE_LIMIT_ANONYMOUS_TTL, 10) : undefined,
      limit: process.env.RATE_LIMIT_ANONYMOUS_LIMIT ? parseInt(process.env.RATE_LIMIT_ANONYMOUS_LIMIT, 10) : undefined,
    },
    failedAuth: {
      ttl: process.env.RATE_LIMIT_FAILED_AUTH_TTL ? parseInt(process.env.RATE_LIMIT_FAILED_AUTH_TTL, 10) : undefined,
      limit: process.env.RATE_LIMIT_FAILED_AUTH_LIMIT
        ? parseInt(process.env.RATE_LIMIT_FAILED_AUTH_LIMIT, 10)
        : undefined,
    },
    authenticated: {
      ttl: process.env.RATE_LIMIT_AUTHENTICATED_TTL
        ? parseInt(process.env.RATE_LIMIT_AUTHENTICATED_TTL, 10)
        : undefined,
      limit: process.env.RATE_LIMIT_AUTHENTICATED_LIMIT
        ? parseInt(process.env.RATE_LIMIT_AUTHENTICATED_LIMIT, 10)
        : undefined,
    },
    sandboxCreate: {
      ttl: process.env.RATE_LIMIT_SANDBOX_CREATE_TTL
        ? parseInt(process.env.RATE_LIMIT_SANDBOX_CREATE_TTL, 10)
        : undefined,
      limit: process.env.RATE_LIMIT_SANDBOX_CREATE_LIMIT
        ? parseInt(process.env.RATE_LIMIT_SANDBOX_CREATE_LIMIT, 10)
        : undefined,
    },
    sandboxLifecycle: {
      ttl: process.env.RATE_LIMIT_SANDBOX_LIFECYCLE_TTL
        ? parseInt(process.env.RATE_LIMIT_SANDBOX_LIFECYCLE_TTL, 10)
        : undefined,
      limit: process.env.RATE_LIMIT_SANDBOX_LIFECYCLE_LIMIT
        ? parseInt(process.env.RATE_LIMIT_SANDBOX_LIFECYCLE_LIMIT, 10)
        : undefined,
    },
  },
  log: {
    console: {
      disabled: process.env.LOG_CONSOLE_DISABLED === 'true',
    },
    level: process.env.LOG_LEVEL || 'info',
    requests: {
      enabled: process.env.LOG_REQUESTS_ENABLED === 'true',
    },
  },
  defaultOrganizationQuota: {
    totalCpuQuota: parseInt(process.env.DEFAULT_ORG_QUOTA_TOTAL_CPU_QUOTA || '10', 10),
    totalMemoryQuota: parseInt(process.env.DEFAULT_ORG_QUOTA_TOTAL_MEMORY_QUOTA || '10', 10),
    totalDiskQuota: parseInt(process.env.DEFAULT_ORG_QUOTA_TOTAL_DISK_QUOTA || '30', 10),
    maxCpuPerSandbox: parseInt(process.env.DEFAULT_ORG_QUOTA_MAX_CPU_PER_SANDBOX || '4', 10),
    maxMemoryPerSandbox: parseInt(process.env.DEFAULT_ORG_QUOTA_MAX_MEMORY_PER_SANDBOX || '8', 10),
    maxDiskPerSandbox: parseInt(process.env.DEFAULT_ORG_QUOTA_MAX_DISK_PER_SANDBOX || '10', 10),
    snapshotQuota: parseInt(process.env.DEFAULT_ORG_QUOTA_SNAPSHOT_QUOTA || '100', 10),
    maxSnapshotSize: parseInt(process.env.DEFAULT_ORG_QUOTA_MAX_SNAPSHOT_SIZE || '20', 10),
    volumeQuota: parseInt(process.env.DEFAULT_ORG_QUOTA_VOLUME_QUOTA || '100', 10),
  },
  defaultRegion: {
    id: process.env.DEFAULT_REGION_ID || 'us',
    name: process.env.DEFAULT_REGION_NAME || 'us',
    enforceQuotas: process.env.DEFAULT_REGION_ENFORCE_QUOTAS === 'true',
  },
  admin: {
    apiKey: process.env.ADMIN_API_KEY,
    totalCpuQuota: parseInt(process.env.ADMIN_TOTAL_CPU_QUOTA || '0', 10),
    totalMemoryQuota: parseInt(process.env.ADMIN_TOTAL_MEMORY_QUOTA || '0', 10),
    totalDiskQuota: parseInt(process.env.ADMIN_TOTAL_DISK_QUOTA || '0', 10),
    maxCpuPerSandbox: parseInt(process.env.ADMIN_MAX_CPU_PER_SANDBOX || '0', 10),
    maxMemoryPerSandbox: parseInt(process.env.ADMIN_MAX_MEMORY_PER_SANDBOX || '0', 10),
    maxDiskPerSandbox: parseInt(process.env.ADMIN_MAX_DISK_PER_SANDBOX || '0', 10),
    snapshotQuota: parseInt(process.env.ADMIN_SNAPSHOT_QUOTA || '100', 10),
    maxSnapshotSize: parseInt(process.env.ADMIN_MAX_SNAPSHOT_SIZE || '100', 10),
    volumeQuota: parseInt(process.env.ADMIN_VOLUME_QUOTA || '0', 10),
  },
  skipUserEmailVerification: process.env.SKIP_USER_EMAIL_VERIFICATION === 'true',
  apiKey: {
    validationCacheTtlSeconds: parseInt(process.env.API_KEY_VALIDATION_CACHE_TTL_SECONDS || '10', 10),
    userCacheTtlSeconds: parseInt(process.env.API_KEY_USER_CACHE_TTL_SECONDS || '60', 10),
  },
  runnerHealthTimeout: parseInt(process.env.RUNNER_HEALTH_TIMEOUT_SECONDS || '3', 10),
  sandboxOtel: {
    endpointUrl: process.env.SANDBOX_OTEL_ENDPOINT_URL,
  },
  otelCollector: {
    apiKey: process.env.OTEL_COLLECTOR_API_KEY,
  },
  encryption: {
    key: process.env.ENCRYPTION_KEY,
    salt: process.env.ENCRYPTION_SALT,
  },
}

export { configuration }
