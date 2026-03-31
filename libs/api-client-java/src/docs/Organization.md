

# Organization


## Properties

| Name | Type | Description | Notes |
|------------ | ------------- | ------------- | -------------|
|**id** | **String** | Organization ID |  |
|**name** | **String** | Organization name |  |
|**createdBy** | **String** | User ID of the organization creator |  |
|**personal** | **Boolean** | Personal organization flag |  |
|**createdAt** | **OffsetDateTime** | Creation timestamp |  |
|**updatedAt** | **OffsetDateTime** | Last update timestamp |  |
|**suspended** | **Boolean** | Suspended flag |  |
|**suspendedAt** | **OffsetDateTime** | Suspended at |  |
|**suspensionReason** | **String** | Suspended reason |  |
|**suspendedUntil** | **OffsetDateTime** | Suspended until |  |
|**suspensionCleanupGracePeriodHours** | **BigDecimal** | Suspension cleanup grace period hours |  |
|**maxCpuPerSandbox** | **BigDecimal** | Max CPU per sandbox |  |
|**maxMemoryPerSandbox** | **BigDecimal** | Max memory per sandbox |  |
|**maxDiskPerSandbox** | **BigDecimal** | Max disk per sandbox |  |
|**snapshotDeactivationTimeoutMinutes** | **BigDecimal** | Time in minutes before an unused snapshot is deactivated |  |
|**sandboxLimitedNetworkEgress** | **Boolean** | Sandbox default network block all |  |
|**defaultRegionId** | **String** | Default region ID |  [optional] |
|**authenticatedRateLimit** | **BigDecimal** | Authenticated rate limit per minute |  |
|**sandboxCreateRateLimit** | **BigDecimal** | Sandbox create rate limit per minute |  |
|**sandboxLifecycleRateLimit** | **BigDecimal** | Sandbox lifecycle rate limit per minute |  |
|**experimentalConfig** | **Object** | Experimental configuration |  |
|**authenticatedRateLimitTtlSeconds** | **BigDecimal** | Authenticated rate limit TTL in seconds |  |
|**sandboxCreateRateLimitTtlSeconds** | **BigDecimal** | Sandbox create rate limit TTL in seconds |  |
|**sandboxLifecycleRateLimitTtlSeconds** | **BigDecimal** | Sandbox lifecycle rate limit TTL in seconds |  |



