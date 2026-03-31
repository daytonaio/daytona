

# UpdateOrganizationInvitation


## Properties

| Name | Type | Description | Notes |
|------------ | ------------- | ------------- | -------------|
|**role** | [**RoleEnum**](#RoleEnum) | Organization member role |  |
|**assignedRoleIds** | **List&lt;String&gt;** | Array of role IDs |  |
|**expiresAt** | **OffsetDateTime** | Expiration date of the invitation |  [optional] |



## Enum: RoleEnum

| Name | Value |
|---- | -----|
| OWNER | &quot;owner&quot; |
| MEMBER | &quot;member&quot; |



