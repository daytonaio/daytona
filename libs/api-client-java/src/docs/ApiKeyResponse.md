

# ApiKeyResponse


## Properties

| Name | Type | Description | Notes |
|------------ | ------------- | ------------- | -------------|
|**name** | **String** | The name of the API key |  |
|**value** | **String** | The API key value |  |
|**createdAt** | **OffsetDateTime** | When the API key was created |  |
|**permissions** | [**List&lt;PermissionsEnum&gt;**](#List&lt;PermissionsEnum&gt;) | The list of organization resource permissions assigned to the API key |  |
|**expiresAt** | **OffsetDateTime** | When the API key expires |  |



## Enum: List&lt;PermissionsEnum&gt;

| Name | Value |
|---- | -----|
| WRITE_REGISTRIES | &quot;write:registries&quot; |
| DELETE_REGISTRIES | &quot;delete:registries&quot; |
| WRITE_SNAPSHOTS | &quot;write:snapshots&quot; |
| DELETE_SNAPSHOTS | &quot;delete:snapshots&quot; |
| WRITE_SANDBOXES | &quot;write:sandboxes&quot; |
| DELETE_SANDBOXES | &quot;delete:sandboxes&quot; |
| READ_VOLUMES | &quot;read:volumes&quot; |
| WRITE_VOLUMES | &quot;write:volumes&quot; |
| DELETE_VOLUMES | &quot;delete:volumes&quot; |
| WRITE_REGIONS | &quot;write:regions&quot; |
| DELETE_REGIONS | &quot;delete:regions&quot; |
| READ_RUNNERS | &quot;read:runners&quot; |
| WRITE_RUNNERS | &quot;write:runners&quot; |
| DELETE_RUNNERS | &quot;delete:runners&quot; |
| READ_AUDIT_LOGS | &quot;read:audit_logs&quot; |



