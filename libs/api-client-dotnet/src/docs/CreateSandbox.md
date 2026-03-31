# Daytona.ApiClient.Model.CreateSandbox

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | **string** | The name of the sandbox. If not provided, the sandbox ID will be used as the name | [optional] 
**Snapshot** | **string** | The ID or name of the snapshot used for the sandbox | [optional] 
**User** | **string** | The user associated with the project | [optional] 
**Env** | **Dictionary&lt;string, string&gt;** | Environment variables for the sandbox | [optional] 
**Labels** | **Dictionary&lt;string, string&gt;** | Labels for the sandbox | [optional] 
**Public** | **bool** | Whether the sandbox http preview is publicly accessible | [optional] 
**NetworkBlockAll** | **bool** | Whether to block all network access for the sandbox | [optional] 
**NetworkAllowList** | **string** | Comma-separated list of allowed CIDR network addresses for the sandbox | [optional] 
**Class** | **string** | The sandbox class type | [optional] 
**Target** | **string** | The target (region) where the sandbox will be created | [optional] 
**Cpu** | **int** | CPU cores allocated to the sandbox | [optional] 
**Gpu** | **int** | GPU units allocated to the sandbox | [optional] 
**Memory** | **int** | Memory allocated to the sandbox in GB | [optional] 
**Disk** | **int** | Disk space allocated to the sandbox in GB | [optional] 
**AutoStopInterval** | **int** | Auto-stop interval in minutes (0 means disabled) | [optional] 
**AutoArchiveInterval** | **int** | Auto-archive interval in minutes (0 means the maximum interval will be used) | [optional] 
**AutoDeleteInterval** | **int** | Auto-delete interval in minutes (negative value means disabled, 0 means delete immediately upon stopping) | [optional] 
**Volumes** | [**List&lt;SandboxVolume&gt;**](SandboxVolume.md) | Array of volumes to attach to the sandbox | [optional] 
**BuildInfo** | [**CreateBuildInfo**](CreateBuildInfo.md) | Build information for the sandbox | [optional] 

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)

