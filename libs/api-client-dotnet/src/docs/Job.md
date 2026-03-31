# Daytona.ApiClient.Model.Job

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | **string** | The ID of the job | 
**Type** | **JobType** | The type of the job | 
**Status** | **JobStatus** | The status of the job | 
**ResourceType** | **string** | The type of resource this job operates on | 
**ResourceId** | **string** | The ID of the resource this job operates on (sandboxId, snapshotRef, etc.) | 
**Payload** | **string** | Job-specific JSON-encoded payload data (operational metadata) | [optional] 
**TraceContext** | **Dictionary&lt;string, Object&gt;** | OpenTelemetry trace context for distributed tracing (W3C Trace Context format) | [optional] 
**ErrorMessage** | **string** | Error message if the job failed | [optional] 
**CreatedAt** | **string** | The creation timestamp of the job | 
**UpdatedAt** | **string** | The last update timestamp of the job | [optional] 

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)

