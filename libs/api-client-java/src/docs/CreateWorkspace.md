

# CreateWorkspace


## Properties

| Name | Type | Description | Notes |
|------------ | ------------- | ------------- | -------------|
|**image** | **String** | The image used for the workspace |  [optional] |
|**user** | **String** | The user associated with the project |  [optional] |
|**env** | **Map&lt;String, String&gt;** | Environment variables for the workspace |  [optional] |
|**labels** | **Map&lt;String, String&gt;** | Labels for the workspace |  [optional] |
|**_public** | **Boolean** | Whether the workspace http preview is publicly accessible |  [optional] |
|**propertyClass** | [**PropertyClassEnum**](#PropertyClassEnum) | The workspace class type |  [optional] |
|**target** | [**TargetEnum**](#TargetEnum) | The target (region) where the workspace will be created |  [optional] |
|**cpu** | **Integer** | CPU cores allocated to the workspace |  [optional] |
|**gpu** | **Integer** | GPU units allocated to the workspace |  [optional] |
|**memory** | **Integer** | Memory allocated to the workspace in GB |  [optional] |
|**disk** | **Integer** | Disk space allocated to the workspace in GB |  [optional] |
|**autoStopInterval** | **Integer** | Auto-stop interval in minutes (0 means disabled) |  [optional] |
|**autoArchiveInterval** | **Integer** | Auto-archive interval in minutes (0 means the maximum interval will be used) |  [optional] |
|**volumes** | [**List&lt;SandboxVolume&gt;**](SandboxVolume.md) | Array of volumes to attach to the workspace |  [optional] |
|**buildInfo** | [**CreateBuildInfo**](CreateBuildInfo.md) | Build information for the workspace |  [optional] |



## Enum: PropertyClassEnum

| Name | Value |
|---- | -----|
| SMALL | &quot;small&quot; |
| MEDIUM | &quot;medium&quot; |
| LARGE | &quot;large&quot; |



## Enum: TargetEnum

| Name | Value |
|---- | -----|
| EU | &quot;eu&quot; |
| US | &quot;us&quot; |
| ASIA | &quot;asia&quot; |



