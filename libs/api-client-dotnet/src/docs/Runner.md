# Daytona.ApiClient.Model.Runner

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | **string** | The ID of the runner | 
**Domain** | **string** | The domain of the runner | [optional] 
**ApiUrl** | **string** | The API URL of the runner | [optional] 
**ProxyUrl** | **string** | The proxy URL of the runner | [optional] 
**Cpu** | **decimal** | The CPU capacity of the runner | 
**Memory** | **decimal** | The memory capacity of the runner in GiB | 
**Disk** | **decimal** | The disk capacity of the runner in GiB | 
**Gpu** | **decimal** | The GPU capacity of the runner | [optional] 
**GpuType** | **string** | The type of GPU | [optional] 
**Class** | **SandboxClass** | The class of the runner | 
**CurrentCpuUsagePercentage** | **decimal** | Current CPU usage percentage | [optional] 
**CurrentMemoryUsagePercentage** | **decimal** | Current RAM usage percentage | [optional] 
**CurrentDiskUsagePercentage** | **decimal** | Current disk usage percentage | [optional] 
**CurrentAllocatedCpu** | **decimal** | Current allocated CPU | [optional] 
**CurrentAllocatedMemoryGiB** | **decimal** | Current allocated memory in GiB | [optional] 
**CurrentAllocatedDiskGiB** | **decimal** | Current allocated disk in GiB | [optional] 
**CurrentSnapshotCount** | **decimal** | Current snapshot count | [optional] 
**CurrentStartedSandboxes** | **decimal** | Current number of started sandboxes | [optional] 
**AvailabilityScore** | **decimal** | Runner availability score | [optional] 
**Region** | **string** | The region of the runner | 
**Name** | **string** | The name of the runner | 
**State** | **RunnerState** | The state of the runner | 
**LastChecked** | **string** | The last time the runner was checked | [optional] 
**Unschedulable** | **bool** | Whether the runner is unschedulable | 
**CreatedAt** | **string** | The creation timestamp of the runner | 
**UpdatedAt** | **string** | The last update timestamp of the runner | 
**VarVersion** | **string** | The version of the runner (deprecated in favor of apiVersion) | 
**ApiVersion** | **string** | The api version of the runner | 
**AppVersion** | **string** | The app version of the runner | [optional] 

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)

