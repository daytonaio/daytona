/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { SandboxClass } from '../sandbox/enums/sandbox-class.enum'
import { getConfigValue, loadConfigMap } from './config-map.loader'

// Load config map from file specified in CONFIG_MAP_PATH environment variable
const configMap = loadConfigMap(process.env.CONFIG_MAP_PATH)

// Helper function to get config value with env precedence
const getConfig = (envVar: string, configPath: string, defaultValue?: string): string | undefined => {
  return getConfigValue(envVar, configPath, configMap) || defaultValue
}

// Helper function to parse int with config map support
const getIntConfig = (envVar: string, configPath: string, defaultValue?: string): number | undefined => {
  const value = getConfig(envVar, configPath, defaultValue)
  return value !== undefined ? parseInt(value, 10) : undefined
}

// Helper function to parse float with config map support
const getFloatConfig = (envVar: string, configPath: string, defaultValue?: string): number => {
  const value = getConfig(envVar, configPath, defaultValue)
  return parseFloat(value || '0')
}

// Helper function to parse boolean with config map support
const getBoolConfig = (envVar: string, configPath: string): boolean => {
  const value = getConfig(envVar, configPath)
  return value === 'true'
}

const configuration = {
  production: process.env.NODE_ENV === 'production',
  version: getConfig('VERSION', 'version', '0.0.0-dev'),
  environment: getConfig('ENVIRONMENT', 'environment'),
  runMigrations: getBoolConfig('RUN_MIGRATIONS', 'runMigrations'),
  port: getIntConfig('PORT', 'port'),
  appUrl: getConfig('APP_URL', 'appUrl'),
  database: {
    host: getConfig('DB_HOST', 'database.host'),
    port: getIntConfig('DB_PORT', 'database.port', '5432'),
    username: getConfig('DB_USERNAME', 'database.username'),
    password: getConfig('DB_PASSWORD', 'database.password'),
    database: getConfig('DB_DATABASE', 'database.database'),
    tls: {
      enabled: getBoolConfig('DB_TLS_ENABLED', 'database.tls.enabled'),
      rejectUnauthorized: getConfig('DB_TLS_REJECT_UNAUTHORIZED', 'database.tls.rejectUnauthorized') !== 'false',
    },
  },
  redis: {
    host: getConfig('REDIS_HOST', 'redis.host'),
    port: getIntConfig('REDIS_PORT', 'redis.port', '6379'),
    tls: getBoolConfig('REDIS_TLS', 'redis.tls') ? {} : undefined,
  },
  posthog: {
    apiKey: getConfig('POSTHOG_API_KEY', 'posthog.apiKey'),
    host: getConfig('POSTHOG_HOST', 'posthog.host'),
    environment: getConfig('POSTHOG_ENVIRONMENT', 'posthog.environment'),
  },
  oidc: {
    clientId: getConfig('OIDC_CLIENT_ID', 'oidc.clientId') || getConfig('OID_CLIENT_ID', 'oidc.clientId'),
    issuer: getConfig('OIDC_ISSUER_BASE_URL', 'oidc.issuer') || getConfig('OID_ISSUER_BASE_URL', 'oidc.issuer'),
    publicIssuer: getConfig('PUBLIC_OIDC_DOMAIN', 'oidc.publicIssuer'),
    audience: getConfig('OIDC_AUDIENCE', 'oidc.audience') || getConfig('OID_AUDIENCE', 'oidc.audience'),
    managementApi: {
      enabled: getBoolConfig('OIDC_MANAGEMENT_API_ENABLED', 'oidc.managementApi.enabled'),
      clientId: getConfig('OIDC_MANAGEMENT_API_CLIENT_ID', 'oidc.managementApi.clientId'),
      clientSecret: getConfig('OIDC_MANAGEMENT_API_CLIENT_SECRET', 'oidc.managementApi.clientSecret'),
      audience: getConfig('OIDC_MANAGEMENT_API_AUDIENCE', 'oidc.managementApi.audience'),
    },
  },
  smtp: {
    host: getConfig('SMTP_HOST', 'smtp.host'),
    port: getIntConfig('SMTP_PORT', 'smtp.port', '587'),
    user: getConfig('SMTP_USER', 'smtp.user'),
    password: getConfig('SMTP_PASSWORD', 'smtp.password'),
    secure: getBoolConfig('SMTP_SECURE', 'smtp.secure'),
    from: getConfig('SMTP_EMAIL_FROM', 'smtp.from', 'noreply@mail.daytona.io'),
  },
  defaultSnapshot: getConfig('DEFAULT_SNAPSHOT', 'defaultSnapshot'),
  dashboardUrl: getConfig('DASHBOARD_URL', 'dashboardUrl'),
  // Default to empty string - dashboard will then hit '/api'
  dashboardBaseApiUrl: getConfig('DASHBOARD_BASE_API_URL', 'dashboardBaseApiUrl', ''),
  transientRegistry: {
    url: getConfig('TRANSIENT_REGISTRY_URL', 'transientRegistry.url'),
    admin: getConfig('TRANSIENT_REGISTRY_ADMIN', 'transientRegistry.admin'),
    password: getConfig('TRANSIENT_REGISTRY_PASSWORD', 'transientRegistry.password'),
    projectId: getConfig('TRANSIENT_REGISTRY_PROJECT_ID', 'transientRegistry.projectId'),
  },
  internalRegistry: {
    url: getConfig('INTERNAL_REGISTRY_URL', 'internalRegistry.url'),
    admin: getConfig('INTERNAL_REGISTRY_ADMIN', 'internalRegistry.admin'),
    password: getConfig('INTERNAL_REGISTRY_PASSWORD', 'internalRegistry.password'),
    projectId: getConfig('INTERNAL_REGISTRY_PROJECT_ID', 'internalRegistry.projectId'),
  },
  s3: {
    endpoint: getConfig('S3_ENDPOINT', 's3.endpoint'),
    stsEndpoint: getConfig('S3_STS_ENDPOINT', 's3.stsEndpoint'),
    region: getConfig('S3_REGION', 's3.region'),
    accessKey: getConfig('S3_ACCESS_KEY', 's3.accessKey'),
    secretKey: getConfig('S3_SECRET_KEY', 's3.secretKey'),
    defaultBucket: getConfig('S3_DEFAULT_BUCKET', 's3.defaultBucket'),
    accountId: getConfig('S3_ACCOUNT_ID', 's3.accountId'),
    roleName: getConfig('S3_ROLE_NAME', 's3.roleName'),
  },
  skipConnections: getBoolConfig('SKIP_CONNECTIONS', 'skipConnections'),
  maxAutoArchiveInterval: getIntConfig('MAX_AUTO_ARCHIVE_INTERVAL', 'maxAutoArchiveInterval', '43200'),
  maintenanceMode: getBoolConfig('MAINTENANCE_MODE', 'maintenanceMode'),
  disableCronJobs: getBoolConfig('DISABLE_CRON_JOBS', 'disableCronJobs'),
  appRole: getConfig('APP_ROLE', 'appRole', 'all'),
  proxy: {
    domain: getConfig('PROXY_DOMAIN', 'proxy.domain'),
    protocol: getConfig('PROXY_PROTOCOL', 'proxy.protocol'),
    apiKey: getConfig('PROXY_API_KEY', 'proxy.apiKey'),
    templateUrl: getConfig('PROXY_TEMPLATE_URL', 'proxy.templateUrl'),
    toolboxUrl:
      (getConfig('PROXY_TOOLBOX_BASE_URL', 'proxy.toolboxUrl') ||
        `${getConfig('PROXY_PROTOCOL', 'proxy.protocol')}://${getConfig('PROXY_DOMAIN', 'proxy.domain')}`) + '/toolbox',
  },
  audit: {
    toolboxRequestsEnabled: getBoolConfig('AUDIT_TOOLBOX_REQUESTS_ENABLED', 'audit.toolboxRequestsEnabled'),
    retentionDays: getIntConfig('AUDIT_LOG_RETENTION_DAYS', 'audit.retentionDays'),
    consoleLogEnabled: getBoolConfig('AUDIT_CONSOLE_LOG_ENABLED', 'audit.consoleLogEnabled'),
    publish: {
      enabled: getBoolConfig('AUDIT_PUBLISH_ENABLED', 'audit.publish.enabled'),
      batchSize: getIntConfig('AUDIT_PUBLISH_BATCH_SIZE', 'audit.publish.batchSize', '1000'),
      mode: getConfig('AUDIT_PUBLISH_MODE', 'audit.publish.mode', 'direct') as 'direct' | 'kafka',
      storageAdapter: getConfig('AUDIT_PUBLISH_STORAGE_ADAPTER', 'audit.publish.storageAdapter', 'opensearch'),
      opensearchIndexName: getConfig(
        'AUDIT_PUBLISH_OPENSEARCH_INDEX_NAME',
        'audit.publish.opensearchIndexName',
        'audit-logs',
      ),
    },
  },
  kafka: {
    enabled: getBoolConfig('KAFKA_ENABLED', 'kafka.enabled'),
    brokers: getConfig('KAFKA_BROKERS', 'kafka.brokers', 'localhost:9092'),
    clientId: getConfig('KAFKA_CLIENT_ID', 'kafka.clientId'),
    sasl: {
      mechanism: getConfig('KAFKA_SASL_MECHANISM', 'kafka.sasl.mechanism'),
      username: getConfig('KAFKA_SASL_USERNAME', 'kafka.sasl.username'),
      password: getConfig('KAFKA_SASL_PASSWORD', 'kafka.sasl.password'),
    },
    tls: {
      enabled: getBoolConfig('KAFKA_TLS_ENABLED', 'kafka.tls.enabled'),
      rejectUnauthorized: getConfig('KAFKA_TLS_REJECT_UNAUTHORIZED', 'kafka.tls.rejectUnauthorized') !== 'false',
    },
  },
  opensearch: {
    nodes: getConfig('OPENSEARCH_NODES', 'opensearch.nodes', 'https://localhost:9200'),
    username: getConfig('OPENSEARCH_USERNAME', 'opensearch.username'),
    password: getConfig('OPENSEARCH_PASSWORD', 'opensearch.password'),
    aws: {
      roleArn: getConfig('OPENSEARCH_AWS_ROLE_ARN', 'opensearch.aws.roleArn'),
      region: getConfig('OPENSEARCH_AWS_REGION', 'opensearch.aws.region'),
    },
    tls: {
      rejectUnauthorized:
        getConfig('OPENSEARCH_TLS_REJECT_UNAUTHORIZED', 'opensearch.tls.rejectUnauthorized') !== 'false',
    },
  },
  cronTimeZone: getConfig('CRON_TIMEZONE', 'cronTimeZone'),
  maxConcurrentBackupsPerRunner: getIntConfig(
    'MAX_CONCURRENT_BACKUPS_PER_RUNNER',
    'maxConcurrentBackupsPerRunner',
    '6',
  ),
  webhook: {
    authToken: getConfig('SVIX_AUTH_TOKEN', 'webhook.authToken'),
    serverUrl: getConfig('SVIX_SERVER_URL', 'webhook.serverUrl'),
  },
  sshGateway: {
    apiKey: getConfig('SSH_GATEWAY_API_KEY', 'sshGateway.apiKey'),
    command: getConfig('SSH_GATEWAY_COMMAND', 'sshGateway.command'),
    publicKey: getConfig('SSH_GATEWAY_PUBLIC_KEY', 'sshGateway.publicKey'),
  },
  organizationSandboxDefaultLimitedNetworkEgress: getBoolConfig(
    'ORGANIZATION_SANDBOX_DEFAULT_LIMITED_NETWORK_EGRESS',
    'organizationSandboxDefaultLimitedNetworkEgress',
  ),
  pylonAppId: getConfig('PYLON_APP_ID', 'pylonAppId'),
  billingApiUrl: getConfig('BILLING_API_URL', 'billingApiUrl'),
  defaultRunner: {
    domain: getConfig('DEFAULT_RUNNER_DOMAIN', 'defaultRunner.domain'),
    apiKey: getConfig('DEFAULT_RUNNER_API_KEY', 'defaultRunner.apiKey'),
    proxyUrl: getConfig('DEFAULT_RUNNER_PROXY_URL', 'defaultRunner.proxyUrl'),
    apiUrl: getConfig('DEFAULT_RUNNER_API_URL', 'defaultRunner.apiUrl'),
    cpu: getIntConfig('DEFAULT_RUNNER_CPU', 'defaultRunner.cpu', '4'),
    memory: getIntConfig('DEFAULT_RUNNER_MEMORY', 'defaultRunner.memory', '8'),
    disk: getIntConfig('DEFAULT_RUNNER_DISK', 'defaultRunner.disk', '50'),
    gpu: getIntConfig('DEFAULT_RUNNER_GPU', 'defaultRunner.gpu', '0'),
    gpuType: getConfig('DEFAULT_RUNNER_GPU_TYPE', 'defaultRunner.gpuType'),
    class: getConfig('DEFAULT_RUNNER_CLASS', 'defaultRunner.class')
      ? (getConfig('DEFAULT_RUNNER_CLASS', 'defaultRunner.class') as SandboxClass)
      : undefined,
    version: getConfig('DEFAULT_RUNNER_VERSION', 'defaultRunner.version', '0'),
  },
  runnerUsage: {
    declarativeBuildScoreThreshold: getIntConfig(
      'RUNNER_DECLARATIVE_BUILD_SCORE_THRESHOLD',
      'runnerUsage.declarativeBuildScoreThreshold',
      '60',
    ),
    availabilityScoreThreshold: getIntConfig(
      'RUNNER_AVAILABILITY_SCORE_THRESHOLD',
      'runnerUsage.availabilityScoreThreshold',
      '60',
    ),
    cpuUsageWeight: getFloatConfig('RUNNER_CPU_USAGE_WEIGHT', 'runnerUsage.cpuUsageWeight', '0.25'),
    memoryUsageWeight: getFloatConfig('RUNNER_MEMORY_USAGE_WEIGHT', 'runnerUsage.memoryUsageWeight', '0.4'),
    diskUsageWeight: getFloatConfig('RUNNER_DISK_USAGE_WEIGHT', 'runnerUsage.diskUsageWeight', '0.4'),
    allocatedCpuWeight: getFloatConfig('RUNNER_ALLOCATED_CPU_WEIGHT', 'runnerUsage.allocatedCpuWeight', '0.03'),
    allocatedMemoryWeight: getFloatConfig(
      'RUNNER_ALLOCATED_MEMORY_WEIGHT',
      'runnerUsage.allocatedMemoryWeight',
      '0.03',
    ),
    allocatedDiskWeight: getFloatConfig('RUNNER_ALLOCATED_DISK_WEIGHT', 'runnerUsage.allocatedDiskWeight', '0.03'),
    cpuPenaltyExponent: getFloatConfig('RUNNER_CPU_PENALTY_EXPONENT', 'runnerUsage.cpuPenaltyExponent', '0.15'),
    memoryPenaltyExponent: getFloatConfig(
      'RUNNER_MEMORY_PENALTY_EXPONENT',
      'runnerUsage.memoryPenaltyExponent',
      '0.15',
    ),
    diskPenaltyExponent: getFloatConfig('RUNNER_DISK_PENALTY_EXPONENT', 'runnerUsage.diskPenaltyExponent', '0.15'),
    cpuPenaltyThreshold: getIntConfig('RUNNER_CPU_PENALTY_THRESHOLD', 'runnerUsage.cpuPenaltyThreshold', '90'),
    memoryPenaltyThreshold: getIntConfig('RUNNER_MEMORY_PENALTY_THRESHOLD', 'runnerUsage.memoryPenaltyThreshold', '75'),
    diskPenaltyThreshold: getIntConfig('RUNNER_DISK_PENALTY_THRESHOLD', 'runnerUsage.diskPenaltyThreshold', '75'),
  },
  rateLimit: {
    anonymous: {
      ttl: getIntConfig('RATE_LIMIT_ANONYMOUS_TTL', 'rateLimit.anonymous.ttl'),
      limit: getIntConfig('RATE_LIMIT_ANONYMOUS_LIMIT', 'rateLimit.anonymous.limit'),
    },
    authenticated: {
      ttl: getIntConfig('RATE_LIMIT_AUTHENTICATED_TTL', 'rateLimit.authenticated.ttl'),
      limit: getIntConfig('RATE_LIMIT_AUTHENTICATED_LIMIT', 'rateLimit.authenticated.limit'),
    },
    sandboxCreate: {
      ttl: getIntConfig('RATE_LIMIT_SANDBOX_CREATE_TTL', 'rateLimit.sandboxCreate.ttl'),
      limit: getIntConfig('RATE_LIMIT_SANDBOX_CREATE_LIMIT', 'rateLimit.sandboxCreate.limit'),
    },
    sandboxLifecycle: {
      ttl: getIntConfig('RATE_LIMIT_SANDBOX_LIFECYCLE_TTL', 'rateLimit.sandboxLifecycle.ttl'),
      limit: getIntConfig('RATE_LIMIT_SANDBOX_LIFECYCLE_LIMIT', 'rateLimit.sandboxLifecycle.limit'),
    },
  },
  log: {
    console: {
      disabled: getBoolConfig('LOG_CONSOLE_DISABLED', 'log.console.disabled'),
    },
    level: getConfig('LOG_LEVEL', 'log.level', 'info'),
    requests: {
      enabled: getBoolConfig('LOG_REQUESTS_ENABLED', 'log.requests.enabled'),
    },
  },
  defaultOrganizationQuota: {
    totalCpuQuota: getIntConfig('DEFAULT_ORG_QUOTA_TOTAL_CPU_QUOTA', 'defaultOrganizationQuota.totalCpuQuota', '10'),
    totalMemoryQuota: getIntConfig(
      'DEFAULT_ORG_QUOTA_TOTAL_MEMORY_QUOTA',
      'defaultOrganizationQuota.totalMemoryQuota',
      '10',
    ),
    totalDiskQuota: getIntConfig('DEFAULT_ORG_QUOTA_TOTAL_DISK_QUOTA', 'defaultOrganizationQuota.totalDiskQuota', '30'),
    maxCpuPerSandbox: getIntConfig(
      'DEFAULT_ORG_QUOTA_MAX_CPU_PER_SANDBOX',
      'defaultOrganizationQuota.maxCpuPerSandbox',
      '4',
    ),
    maxMemoryPerSandbox: getIntConfig(
      'DEFAULT_ORG_QUOTA_MAX_MEMORY_PER_SANDBOX',
      'defaultOrganizationQuota.maxMemoryPerSandbox',
      '8',
    ),
    maxDiskPerSandbox: getIntConfig(
      'DEFAULT_ORG_QUOTA_MAX_DISK_PER_SANDBOX',
      'defaultOrganizationQuota.maxDiskPerSandbox',
      '10',
    ),
    snapshotQuota: getIntConfig('DEFAULT_ORG_QUOTA_SNAPSHOT_QUOTA', 'defaultOrganizationQuota.snapshotQuota', '100'),
    maxSnapshotSize: getIntConfig(
      'DEFAULT_ORG_QUOTA_MAX_SNAPSHOT_SIZE',
      'defaultOrganizationQuota.maxSnapshotSize',
      '20',
    ),
    volumeQuota: getIntConfig('DEFAULT_ORG_QUOTA_VOLUME_QUOTA', 'defaultOrganizationQuota.volumeQuota', '100'),
  },
  defaultRegion: {
    id: getConfig('DEFAULT_REGION_ID', 'defaultRegion.id', 'us'),
    name: getConfig('DEFAULT_REGION_NAME', 'defaultRegion.name', 'us'),
    enforceQuotas: getBoolConfig('DEFAULT_REGION_ENFORCE_QUOTAS', 'defaultRegion.enforceQuotas'),
  },
  admin: {
    apiKey: getConfig('ADMIN_API_KEY', 'admin.apiKey'),
    totalCpuQuota: getIntConfig('ADMIN_TOTAL_CPU_QUOTA', 'admin.totalCpuQuota', '0'),
    totalMemoryQuota: getIntConfig('ADMIN_TOTAL_MEMORY_QUOTA', 'admin.totalMemoryQuota', '0'),
    totalDiskQuota: getIntConfig('ADMIN_TOTAL_DISK_QUOTA', 'admin.totalDiskQuota', '0'),
    maxCpuPerSandbox: getIntConfig('ADMIN_MAX_CPU_PER_SANDBOX', 'admin.maxCpuPerSandbox', '0'),
    maxMemoryPerSandbox: getIntConfig('ADMIN_MAX_MEMORY_PER_SANDBOX', 'admin.maxMemoryPerSandbox', '0'),
    maxDiskPerSandbox: getIntConfig('ADMIN_MAX_DISK_PER_SANDBOX', 'admin.maxDiskPerSandbox', '0'),
    snapshotQuota: getIntConfig('ADMIN_SNAPSHOT_QUOTA', 'admin.snapshotQuota', '100'),
    maxSnapshotSize: getIntConfig('ADMIN_MAX_SNAPSHOT_SIZE', 'admin.maxSnapshotSize', '100'),
    volumeQuota: getIntConfig('ADMIN_VOLUME_QUOTA', 'admin.volumeQuota', '0'),
  },
  skipUserEmailVerification: getBoolConfig('SKIP_USER_EMAIL_VERIFICATION', 'skipUserEmailVerification'),
  apiKey: {
    validationCacheTtlSeconds: getIntConfig(
      'API_KEY_VALIDATION_CACHE_TTL_SECONDS',
      'apiKey.validationCacheTtlSeconds',
      '10',
    ),
    userCacheTtlSeconds: getIntConfig('API_KEY_USER_CACHE_TTL_SECONDS', 'apiKey.userCacheTtlSeconds', '60'),
  },
  runnerHealthTimeout: getIntConfig('RUNNER_HEALTH_TIMEOUT_SECONDS', 'runnerHealthTimeout', '3'),
}

export { configuration }
