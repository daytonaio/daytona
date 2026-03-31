# Daytona.ApiClient.Model.Sandbox

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | **string** | The ID of the sandbox | 
**OrganizationId** | **string** | The organization ID of the sandbox | 
**Name** | **string** | The name of the sandbox | 
**Snapshot** | **string** | The snapshot used for the sandbox | [optional] 
**User** | **string** | The user associated with the project | 
**Env** | **Dictionary&lt;string, string&gt;** | Environment variables for the sandbox | 
**Labels** | **Dictionary&lt;string, string&gt;** | Labels for the sandbox | 
**Public** | **bool** | Whether the sandbox http preview is public | 
**NetworkBlockAll** | **bool** | Whether to block all network access for the sandbox | 
**NetworkAllowList** | **string** | Comma-separated list of allowed CIDR network addresses for the sandbox | [optional] 
**Target** | **string** | The target environment for the sandbox | 
**Cpu** | **decimal** | The CPU quota for the sandbox | 
**Gpu** | **decimal** | The GPU quota for the sandbox | 
**Memory** | **decimal** | The memory quota for the sandbox | 
**Disk** | **decimal** | The disk quota for the sandbox | 
**State** | **SandboxState** | The state of the sandbox | [optional] 
**DesiredState** | **SandboxDesiredState** | The desired state of the sandbox | [optional] 
**ErrorReason** | **string** | The error reason of the sandbox | [optional] 
**Recoverable** | **bool** | Whether the sandbox error is recoverable. | [optional] 
**BackupState** | **string** | The state of the backup | [optional] 
**BackupCreatedAt** | **string** | The creation timestamp of the last backup | [optional] 
**AutoStopInterval** | **decimal** | Auto-stop interval in minutes (0 means disabled) | [optional] 
**AutoArchiveInterval** | **decimal** | Auto-archive interval in minutes | [optional] 
**AutoDeleteInterval** | **decimal** | Auto-delete interval in minutes (negative value means disabled, 0 means delete immediately upon stopping) | [optional] 
**Volumes** | [**List&lt;SandboxVolume&gt;**](SandboxVolume.md) | Array of volumes attached to the sandbox | [optional] 
**BuildInfo** | [**BuildInfo**](BuildInfo.md) | Build information for the sandbox | [optional] 
**CreatedAt** | **string** | The creation timestamp of the sandbox | [optional] 
**UpdatedAt** | **string** | The last update timestamp of the sandbox | [optional] 
**Class** | **string** | The class of the sandbox | [optional] 
**DaemonVersion** | **string** | The version of the daemon running in the sandbox | [optional] 
**RunnerId** | **string** | The runner ID of the sandbox | [optional] 
**ToolboxProxyUrl** | **string** | The toolbox proxy URL for the sandbox | 

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)

