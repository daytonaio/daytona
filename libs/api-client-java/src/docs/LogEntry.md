

# LogEntry


## Properties

| Name | Type | Description | Notes |
|------------ | ------------- | ------------- | -------------|
|**timestamp** | **String** | Timestamp of the log entry |  |
|**body** | **String** | Log message body |  |
|**severityText** | **String** | Severity level text (e.g., INFO, WARN, ERROR) |  |
|**severityNumber** | **BigDecimal** | Severity level number |  [optional] |
|**serviceName** | **String** | Service name that generated the log |  |
|**resourceAttributes** | **Map&lt;String, String&gt;** | Resource attributes from OTEL |  |
|**logAttributes** | **Map&lt;String, String&gt;** | Log-specific attributes |  |
|**traceId** | **String** | Associated trace ID if available |  [optional] |
|**spanId** | **String** | Associated span ID if available |  [optional] |



