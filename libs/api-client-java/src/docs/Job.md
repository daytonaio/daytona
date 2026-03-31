

# Job


## Properties

| Name | Type | Description | Notes |
|------------ | ------------- | ------------- | -------------|
|**id** | **String** | The ID of the job |  |
|**type** | **JobType** | The type of the job |  |
|**status** | **JobStatus** | The status of the job |  |
|**resourceType** | [**ResourceTypeEnum**](#ResourceTypeEnum) | The type of resource this job operates on |  |
|**resourceId** | **String** | The ID of the resource this job operates on (sandboxId, snapshotRef, etc.) |  |
|**payload** | **String** | Job-specific JSON-encoded payload data (operational metadata) |  [optional] |
|**traceContext** | **Map&lt;String, Object&gt;** | OpenTelemetry trace context for distributed tracing (W3C Trace Context format) |  [optional] |
|**errorMessage** | **String** | Error message if the job failed |  [optional] |
|**createdAt** | **String** | The creation timestamp of the job |  |
|**updatedAt** | **String** | The last update timestamp of the job |  [optional] |



## Enum: ResourceTypeEnum

| Name | Value |
|---- | -----|
| SANDBOX | &quot;SANDBOX&quot; |
| SNAPSHOT | &quot;SNAPSHOT&quot; |
| BACKUP | &quot;BACKUP&quot; |



