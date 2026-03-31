# Daytona.ApiClient.Model.Organization

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | **string** | Organization ID | 
**Name** | **string** | Organization name | 
**CreatedBy** | **string** | User ID of the organization creator | 
**Personal** | **bool** | Personal organization flag | 
**CreatedAt** | **DateTime** | Creation timestamp | 
**UpdatedAt** | **DateTime** | Last update timestamp | 
**Suspended** | **bool** | Suspended flag | 
**SuspendedAt** | **DateTime** | Suspended at | 
**SuspensionReason** | **string** | Suspended reason | 
**SuspendedUntil** | **DateTime** | Suspended until | 
**SuspensionCleanupGracePeriodHours** | **decimal** | Suspension cleanup grace period hours | 
**MaxCpuPerSandbox** | **decimal** | Max CPU per sandbox | 
**MaxMemoryPerSandbox** | **decimal** | Max memory per sandbox | 
**MaxDiskPerSandbox** | **decimal** | Max disk per sandbox | 
**SnapshotDeactivationTimeoutMinutes** | **decimal** | Time in minutes before an unused snapshot is deactivated | [default to 20160M]
**SandboxLimitedNetworkEgress** | **bool** | Sandbox default network block all | 
**DefaultRegionId** | **string** | Default region ID | [optional] 
**AuthenticatedRateLimit** | **decimal?** | Authenticated rate limit per minute | 
**SandboxCreateRateLimit** | **decimal?** | Sandbox create rate limit per minute | 
**SandboxLifecycleRateLimit** | **decimal?** | Sandbox lifecycle rate limit per minute | 
**ExperimentalConfig** | **Object** | Experimental configuration | 
**AuthenticatedRateLimitTtlSeconds** | **decimal?** | Authenticated rate limit TTL in seconds | 
**SandboxCreateRateLimitTtlSeconds** | **decimal?** | Sandbox create rate limit TTL in seconds | 
**SandboxLifecycleRateLimitTtlSeconds** | **decimal?** | Sandbox lifecycle rate limit TTL in seconds | 

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)

