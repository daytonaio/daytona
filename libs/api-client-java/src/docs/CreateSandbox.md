

# CreateSandbox


## Properties

| Name | Type | Description | Notes |
|------------ | ------------- | ------------- | -------------|
|**name** | **String** | The name of the sandbox. If not provided, the sandbox ID will be used as the name |  [optional] |
|**snapshot** | **String** | The ID or name of the snapshot used for the sandbox |  [optional] |
|**user** | **String** | The user associated with the project |  [optional] |
|**env** | **Map&lt;String, String&gt;** | Environment variables for the sandbox |  [optional] |
|**labels** | **Map&lt;String, String&gt;** | Labels for the sandbox |  [optional] |
|**_public** | **Boolean** | Whether the sandbox http preview is publicly accessible |  [optional] |
|**networkBlockAll** | **Boolean** | Whether to block all network access for the sandbox |  [optional] |
|**networkAllowList** | **String** | Comma-separated list of allowed CIDR network addresses for the sandbox |  [optional] |
|**propertyClass** | [**PropertyClassEnum**](#PropertyClassEnum) | The sandbox class type |  [optional] |
|**target** | **String** | The target (region) where the sandbox will be created |  [optional] |
|**cpu** | **Integer** | CPU cores allocated to the sandbox |  [optional] |
|**gpu** | **Integer** | GPU units allocated to the sandbox |  [optional] |
|**memory** | **Integer** | Memory allocated to the sandbox in GB |  [optional] |
|**disk** | **Integer** | Disk space allocated to the sandbox in GB |  [optional] |
|**autoStopInterval** | **Integer** | Auto-stop interval in minutes (0 means disabled) |  [optional] |
|**autoArchiveInterval** | **Integer** | Auto-archive interval in minutes (0 means the maximum interval will be used) |  [optional] |
|**autoDeleteInterval** | **Integer** | Auto-delete interval in minutes (negative value means disabled, 0 means delete immediately upon stopping) |  [optional] |
|**volumes** | [**List&lt;SandboxVolume&gt;**](SandboxVolume.md) | Array of volumes to attach to the sandbox |  [optional] |
|**buildInfo** | [**CreateBuildInfo**](CreateBuildInfo.md) | Build information for the sandbox |  [optional] |



## Enum: PropertyClassEnum

| Name | Value |
|---- | -----|
| SMALL | &quot;small&quot; |
| MEDIUM | &quot;medium&quot; |
| LARGE | &quot;large&quot; |



