

# CreateOrganizationInvitation


## Properties

| Name | Type | Description | Notes |
|------------ | ------------- | ------------- | -------------|
|**email** | **String** | Email address of the invitee |  |
|**role** | [**RoleEnum**](#RoleEnum) | Organization member role for the invitee |  |
|**assignedRoleIds** | **List&lt;String&gt;** | Array of assigned role IDs for the invitee |  |
|**expiresAt** | **OffsetDateTime** | Expiration date of the invitation |  [optional] |



## Enum: RoleEnum

| Name | Value |
|---- | -----|
| OWNER | &quot;owner&quot; |
| MEMBER | &quot;member&quot; |



