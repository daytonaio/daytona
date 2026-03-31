# Daytona.ApiClient.Model.CreateWorkspace

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Image** | **string** | The image used for the workspace | [optional] 
**User** | **string** | The user associated with the project | [optional] 
**Env** | **Dictionary&lt;string, string&gt;** | Environment variables for the workspace | [optional] 
**Labels** | **Dictionary&lt;string, string&gt;** | Labels for the workspace | [optional] 
**Public** | **bool** | Whether the workspace http preview is publicly accessible | [optional] 
**Class** | **string** | The workspace class type | [optional] 
**Target** | **string** | The target (region) where the workspace will be created | [optional] 
**Cpu** | **int** | CPU cores allocated to the workspace | [optional] 
**Gpu** | **int** | GPU units allocated to the workspace | [optional] 
**Memory** | **int** | Memory allocated to the workspace in GB | [optional] 
**Disk** | **int** | Disk space allocated to the workspace in GB | [optional] 
**AutoStopInterval** | **int** | Auto-stop interval in minutes (0 means disabled) | [optional] 
**AutoArchiveInterval** | **int** | Auto-archive interval in minutes (0 means the maximum interval will be used) | [optional] 
**Volumes** | [**List&lt;SandboxVolume&gt;**](SandboxVolume.md) | Array of volumes to attach to the workspace | [optional] 
**BuildInfo** | [**CreateBuildInfo**](CreateBuildInfo.md) | Build information for the workspace | [optional] 

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)

