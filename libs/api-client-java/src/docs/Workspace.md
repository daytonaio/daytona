

# Workspace


## Properties

| Name | Type | Description | Notes |
|------------ | ------------- | ------------- | -------------|
|**id** | **String** | The ID of the sandbox |  |
|**organizationId** | **String** | The organization ID of the sandbox |  |
|**name** | **String** | The name of the sandbox |  |
|**snapshot** | **String** | The snapshot used for the sandbox |  [optional] |
|**user** | **String** | The user associated with the project |  |
|**env** | **Map&lt;String, String&gt;** | Environment variables for the sandbox |  |
|**labels** | **Map&lt;String, String&gt;** | Labels for the sandbox |  |
|**_public** | **Boolean** | Whether the sandbox http preview is public |  |
|**networkBlockAll** | **Boolean** | Whether to block all network access for the sandbox |  |
|**networkAllowList** | **String** | Comma-separated list of allowed CIDR network addresses for the sandbox |  [optional] |
|**target** | **String** | The target environment for the sandbox |  |
|**cpu** | **BigDecimal** | The CPU quota for the sandbox |  |
|**gpu** | **BigDecimal** | The GPU quota for the sandbox |  |
|**memory** | **BigDecimal** | The memory quota for the sandbox |  |
|**disk** | **BigDecimal** | The disk quota for the sandbox |  |
|**state** | **SandboxState** | The state of the sandbox |  [optional] |
|**desiredState** | **SandboxDesiredState** | The desired state of the sandbox |  [optional] |
|**errorReason** | **String** | The error reason of the sandbox |  [optional] |
|**recoverable** | **Boolean** | Whether the sandbox error is recoverable. |  [optional] |
|**backupState** | [**BackupStateEnum**](#BackupStateEnum) | The state of the backup |  [optional] |
|**backupCreatedAt** | **String** | The creation timestamp of the last backup |  [optional] |
|**autoStopInterval** | **BigDecimal** | Auto-stop interval in minutes (0 means disabled) |  [optional] |
|**autoArchiveInterval** | **BigDecimal** | Auto-archive interval in minutes |  [optional] |
|**autoDeleteInterval** | **BigDecimal** | Auto-delete interval in minutes (negative value means disabled, 0 means delete immediately upon stopping) |  [optional] |
|**volumes** | [**List&lt;SandboxVolume&gt;**](SandboxVolume.md) | Array of volumes attached to the sandbox |  [optional] |
|**buildInfo** | [**BuildInfo**](BuildInfo.md) | Build information for the sandbox |  [optional] |
|**createdAt** | **String** | The creation timestamp of the sandbox |  [optional] |
|**updatedAt** | **String** | The last update timestamp of the sandbox |  [optional] |
|**propertyClass** | [**PropertyClassEnum**](#PropertyClassEnum) | The class of the sandbox |  [optional] |
|**daemonVersion** | **String** | The version of the daemon running in the sandbox |  [optional] |
|**runnerId** | **String** | The runner ID of the sandbox |  [optional] |
|**toolboxProxyUrl** | **String** | The toolbox proxy URL for the sandbox |  |
|**image** | **String** | The image used for the workspace |  [optional] |
|**snapshotState** | [**SnapshotStateEnum**](#SnapshotStateEnum) | The state of the snapshot |  [optional] |
|**snapshotCreatedAt** | **String** | The creation timestamp of the last snapshot |  [optional] |
|**info** | [**SandboxInfo**](SandboxInfo.md) | Additional information about the sandbox |  [optional] |



## Enum: BackupStateEnum

| Name | Value |
|---- | -----|
| NONE | &quot;None&quot; |
| PENDING | &quot;Pending&quot; |
| IN_PROGRESS | &quot;InProgress&quot; |
| COMPLETED | &quot;Completed&quot; |
| ERROR | &quot;Error&quot; |



## Enum: PropertyClassEnum

| Name | Value |
|---- | -----|
| SMALL | &quot;small&quot; |
| MEDIUM | &quot;medium&quot; |
| LARGE | &quot;large&quot; |



## Enum: SnapshotStateEnum

| Name | Value |
|---- | -----|
| NONE | &quot;None&quot; |
| PENDING | &quot;Pending&quot; |
| IN_PROGRESS | &quot;InProgress&quot; |
| COMPLETED | &quot;Completed&quot; |
| ERROR | &quot;Error&quot; |



