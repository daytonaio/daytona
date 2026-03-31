

# Runner


## Properties

| Name | Type | Description | Notes |
|------------ | ------------- | ------------- | -------------|
|**id** | **String** | The ID of the runner |  |
|**domain** | **String** | The domain of the runner |  [optional] |
|**apiUrl** | **String** | The API URL of the runner |  [optional] |
|**proxyUrl** | **String** | The proxy URL of the runner |  [optional] |
|**cpu** | **BigDecimal** | The CPU capacity of the runner |  |
|**memory** | **BigDecimal** | The memory capacity of the runner in GiB |  |
|**disk** | **BigDecimal** | The disk capacity of the runner in GiB |  |
|**gpu** | **BigDecimal** | The GPU capacity of the runner |  [optional] |
|**gpuType** | **String** | The type of GPU |  [optional] |
|**propertyClass** | **SandboxClass** | The class of the runner |  |
|**currentCpuUsagePercentage** | **BigDecimal** | Current CPU usage percentage |  [optional] |
|**currentMemoryUsagePercentage** | **BigDecimal** | Current RAM usage percentage |  [optional] |
|**currentDiskUsagePercentage** | **BigDecimal** | Current disk usage percentage |  [optional] |
|**currentAllocatedCpu** | **BigDecimal** | Current allocated CPU |  [optional] |
|**currentAllocatedMemoryGiB** | **BigDecimal** | Current allocated memory in GiB |  [optional] |
|**currentAllocatedDiskGiB** | **BigDecimal** | Current allocated disk in GiB |  [optional] |
|**currentSnapshotCount** | **BigDecimal** | Current snapshot count |  [optional] |
|**currentStartedSandboxes** | **BigDecimal** | Current number of started sandboxes |  [optional] |
|**availabilityScore** | **BigDecimal** | Runner availability score |  [optional] |
|**region** | **String** | The region of the runner |  |
|**name** | **String** | The name of the runner |  |
|**state** | **RunnerState** | The state of the runner |  |
|**lastChecked** | **String** | The last time the runner was checked |  [optional] |
|**unschedulable** | **Boolean** | Whether the runner is unschedulable |  |
|**createdAt** | **String** | The creation timestamp of the runner |  |
|**updatedAt** | **String** | The last update timestamp of the runner |  |
|**version** | **String** | The version of the runner (deprecated in favor of apiVersion) |  |
|**apiVersion** | **String** | The api version of the runner |  |
|**appVersion** | **String** | The app version of the runner |  [optional] |



