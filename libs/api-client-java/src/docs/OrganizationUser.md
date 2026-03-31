

# OrganizationUser


## Properties

| Name | Type | Description | Notes |
|------------ | ------------- | ------------- | -------------|
|**userId** | **String** | User ID |  |
|**organizationId** | **String** | Organization ID |  |
|**name** | **String** | User name |  |
|**email** | **String** | User email |  |
|**role** | [**RoleEnum**](#RoleEnum) | Member role |  |
|**assignedRoles** | [**List&lt;OrganizationRole&gt;**](OrganizationRole.md) | Roles assigned to the user |  |
|**createdAt** | **OffsetDateTime** | Creation timestamp |  |
|**updatedAt** | **OffsetDateTime** | Last update timestamp |  |



## Enum: RoleEnum

| Name | Value |
|---- | -----|
| OWNER | &quot;owner&quot; |
| MEMBER | &quot;member&quot; |



