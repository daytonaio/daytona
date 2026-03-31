

# UpdateSandboxStateDto


## Properties

| Name | Type | Description | Notes |
|------------ | ------------- | ------------- | -------------|
|**state** | [**StateEnum**](#StateEnum) | The new state for the sandbox |  |
|**errorReason** | **String** | Optional error message when reporting an error state |  [optional] |
|**recoverable** | **Boolean** | Whether the sandbox is recoverable |  [optional] |



## Enum: StateEnum

| Name | Value |
|---- | -----|
| CREATING | &quot;creating&quot; |
| RESTORING | &quot;restoring&quot; |
| DESTROYED | &quot;destroyed&quot; |
| DESTROYING | &quot;destroying&quot; |
| STARTED | &quot;started&quot; |
| STOPPED | &quot;stopped&quot; |
| STARTING | &quot;starting&quot; |
| STOPPING | &quot;stopping&quot; |
| ERROR | &quot;error&quot; |
| BUILD_FAILED | &quot;build_failed&quot; |
| PENDING_BUILD | &quot;pending_build&quot; |
| BUILDING_SNAPSHOT | &quot;building_snapshot&quot; |
| UNKNOWN | &quot;unknown&quot; |
| PULLING_SNAPSHOT | &quot;pulling_snapshot&quot; |
| ARCHIVED | &quot;archived&quot; |
| ARCHIVING | &quot;archiving&quot; |
| RESIZING | &quot;resizing&quot; |



