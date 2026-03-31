# Daytona.ApiClient.Model.LogEntry

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Timestamp** | **string** | Timestamp of the log entry | 
**Body** | **string** | Log message body | 
**SeverityText** | **string** | Severity level text (e.g., INFO, WARN, ERROR) | 
**SeverityNumber** | **decimal** | Severity level number | [optional] 
**ServiceName** | **string** | Service name that generated the log | 
**ResourceAttributes** | **Dictionary&lt;string, string&gt;** | Resource attributes from OTEL | 
**LogAttributes** | **Dictionary&lt;string, string&gt;** | Log-specific attributes | 
**TraceId** | **string** | Associated trace ID if available | [optional] 
**SpanId** | **string** | Associated span ID if available | [optional] 

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)

