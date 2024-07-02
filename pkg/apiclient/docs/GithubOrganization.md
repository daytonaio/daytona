# GithubOrganization

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**AvatarUrl** | Pointer to **string** |  | [optional] 
**BillingEmail** | Pointer to **string** |  | [optional] 
**Blog** | Pointer to **string** |  | [optional] 
**Collaborators** | Pointer to **int32** |  | [optional] 
**Company** | Pointer to **string** |  | [optional] 
**CreatedAt** | Pointer to **string** |  | [optional] 
**Description** | Pointer to **string** |  | [optional] 
**DiskUsage** | Pointer to **int32** |  | [optional] 
**Email** | Pointer to **string** |  | [optional] 
**EventsUrl** | Pointer to **string** |  | [optional] 
**Followers** | Pointer to **int32** |  | [optional] 
**Following** | Pointer to **int32** |  | [optional] 
**HooksUrl** | Pointer to **string** |  | [optional] 
**HtmlUrl** | Pointer to **string** |  | [optional] 
**Id** | Pointer to **int32** |  | [optional] 
**IssuesUrl** | Pointer to **string** |  | [optional] 
**Location** | Pointer to **string** |  | [optional] 
**Login** | Pointer to **string** |  | [optional] 
**MembersUrl** | Pointer to **string** |  | [optional] 
**Name** | Pointer to **string** |  | [optional] 
**NodeId** | Pointer to **string** |  | [optional] 
**OwnedPrivateRepos** | Pointer to **int32** |  | [optional] 
**Plan** | Pointer to [**GithubPlan**](GithubPlan.md) |  | [optional] 
**PrivateGists** | Pointer to **int32** |  | [optional] 
**PublicGists** | Pointer to **int32** |  | [optional] 
**PublicMembersUrl** | Pointer to **string** |  | [optional] 
**PublicRepos** | Pointer to **int32** |  | [optional] 
**ReposUrl** | Pointer to **string** |  | [optional] 
**TotalPrivateRepos** | Pointer to **int32** |  | [optional] 
**Type** | Pointer to **string** |  | [optional] 
**UpdatedAt** | Pointer to **string** |  | [optional] 
**Url** | Pointer to **string** | API URLs | [optional] 

## Methods

### NewGithubOrganization

`func NewGithubOrganization() *GithubOrganization`

NewGithubOrganization instantiates a new GithubOrganization object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewGithubOrganizationWithDefaults

`func NewGithubOrganizationWithDefaults() *GithubOrganization`

NewGithubOrganizationWithDefaults instantiates a new GithubOrganization object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetAvatarUrl

`func (o *GithubOrganization) GetAvatarUrl() string`

GetAvatarUrl returns the AvatarUrl field if non-nil, zero value otherwise.

### GetAvatarUrlOk

`func (o *GithubOrganization) GetAvatarUrlOk() (*string, bool)`

GetAvatarUrlOk returns a tuple with the AvatarUrl field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAvatarUrl

`func (o *GithubOrganization) SetAvatarUrl(v string)`

SetAvatarUrl sets AvatarUrl field to given value.

### HasAvatarUrl

`func (o *GithubOrganization) HasAvatarUrl() bool`

HasAvatarUrl returns a boolean if a field has been set.

### GetBillingEmail

`func (o *GithubOrganization) GetBillingEmail() string`

GetBillingEmail returns the BillingEmail field if non-nil, zero value otherwise.

### GetBillingEmailOk

`func (o *GithubOrganization) GetBillingEmailOk() (*string, bool)`

GetBillingEmailOk returns a tuple with the BillingEmail field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetBillingEmail

`func (o *GithubOrganization) SetBillingEmail(v string)`

SetBillingEmail sets BillingEmail field to given value.

### HasBillingEmail

`func (o *GithubOrganization) HasBillingEmail() bool`

HasBillingEmail returns a boolean if a field has been set.

### GetBlog

`func (o *GithubOrganization) GetBlog() string`

GetBlog returns the Blog field if non-nil, zero value otherwise.

### GetBlogOk

`func (o *GithubOrganization) GetBlogOk() (*string, bool)`

GetBlogOk returns a tuple with the Blog field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetBlog

`func (o *GithubOrganization) SetBlog(v string)`

SetBlog sets Blog field to given value.

### HasBlog

`func (o *GithubOrganization) HasBlog() bool`

HasBlog returns a boolean if a field has been set.

### GetCollaborators

`func (o *GithubOrganization) GetCollaborators() int32`

GetCollaborators returns the Collaborators field if non-nil, zero value otherwise.

### GetCollaboratorsOk

`func (o *GithubOrganization) GetCollaboratorsOk() (*int32, bool)`

GetCollaboratorsOk returns a tuple with the Collaborators field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCollaborators

`func (o *GithubOrganization) SetCollaborators(v int32)`

SetCollaborators sets Collaborators field to given value.

### HasCollaborators

`func (o *GithubOrganization) HasCollaborators() bool`

HasCollaborators returns a boolean if a field has been set.

### GetCompany

`func (o *GithubOrganization) GetCompany() string`

GetCompany returns the Company field if non-nil, zero value otherwise.

### GetCompanyOk

`func (o *GithubOrganization) GetCompanyOk() (*string, bool)`

GetCompanyOk returns a tuple with the Company field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCompany

`func (o *GithubOrganization) SetCompany(v string)`

SetCompany sets Company field to given value.

### HasCompany

`func (o *GithubOrganization) HasCompany() bool`

HasCompany returns a boolean if a field has been set.

### GetCreatedAt

`func (o *GithubOrganization) GetCreatedAt() string`

GetCreatedAt returns the CreatedAt field if non-nil, zero value otherwise.

### GetCreatedAtOk

`func (o *GithubOrganization) GetCreatedAtOk() (*string, bool)`

GetCreatedAtOk returns a tuple with the CreatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedAt

`func (o *GithubOrganization) SetCreatedAt(v string)`

SetCreatedAt sets CreatedAt field to given value.

### HasCreatedAt

`func (o *GithubOrganization) HasCreatedAt() bool`

HasCreatedAt returns a boolean if a field has been set.

### GetDescription

`func (o *GithubOrganization) GetDescription() string`

GetDescription returns the Description field if non-nil, zero value otherwise.

### GetDescriptionOk

`func (o *GithubOrganization) GetDescriptionOk() (*string, bool)`

GetDescriptionOk returns a tuple with the Description field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDescription

`func (o *GithubOrganization) SetDescription(v string)`

SetDescription sets Description field to given value.

### HasDescription

`func (o *GithubOrganization) HasDescription() bool`

HasDescription returns a boolean if a field has been set.

### GetDiskUsage

`func (o *GithubOrganization) GetDiskUsage() int32`

GetDiskUsage returns the DiskUsage field if non-nil, zero value otherwise.

### GetDiskUsageOk

`func (o *GithubOrganization) GetDiskUsageOk() (*int32, bool)`

GetDiskUsageOk returns a tuple with the DiskUsage field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDiskUsage

`func (o *GithubOrganization) SetDiskUsage(v int32)`

SetDiskUsage sets DiskUsage field to given value.

### HasDiskUsage

`func (o *GithubOrganization) HasDiskUsage() bool`

HasDiskUsage returns a boolean if a field has been set.

### GetEmail

`func (o *GithubOrganization) GetEmail() string`

GetEmail returns the Email field if non-nil, zero value otherwise.

### GetEmailOk

`func (o *GithubOrganization) GetEmailOk() (*string, bool)`

GetEmailOk returns a tuple with the Email field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEmail

`func (o *GithubOrganization) SetEmail(v string)`

SetEmail sets Email field to given value.

### HasEmail

`func (o *GithubOrganization) HasEmail() bool`

HasEmail returns a boolean if a field has been set.

### GetEventsUrl

`func (o *GithubOrganization) GetEventsUrl() string`

GetEventsUrl returns the EventsUrl field if non-nil, zero value otherwise.

### GetEventsUrlOk

`func (o *GithubOrganization) GetEventsUrlOk() (*string, bool)`

GetEventsUrlOk returns a tuple with the EventsUrl field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEventsUrl

`func (o *GithubOrganization) SetEventsUrl(v string)`

SetEventsUrl sets EventsUrl field to given value.

### HasEventsUrl

`func (o *GithubOrganization) HasEventsUrl() bool`

HasEventsUrl returns a boolean if a field has been set.

### GetFollowers

`func (o *GithubOrganization) GetFollowers() int32`

GetFollowers returns the Followers field if non-nil, zero value otherwise.

### GetFollowersOk

`func (o *GithubOrganization) GetFollowersOk() (*int32, bool)`

GetFollowersOk returns a tuple with the Followers field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFollowers

`func (o *GithubOrganization) SetFollowers(v int32)`

SetFollowers sets Followers field to given value.

### HasFollowers

`func (o *GithubOrganization) HasFollowers() bool`

HasFollowers returns a boolean if a field has been set.

### GetFollowing

`func (o *GithubOrganization) GetFollowing() int32`

GetFollowing returns the Following field if non-nil, zero value otherwise.

### GetFollowingOk

`func (o *GithubOrganization) GetFollowingOk() (*int32, bool)`

GetFollowingOk returns a tuple with the Following field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFollowing

`func (o *GithubOrganization) SetFollowing(v int32)`

SetFollowing sets Following field to given value.

### HasFollowing

`func (o *GithubOrganization) HasFollowing() bool`

HasFollowing returns a boolean if a field has been set.

### GetHooksUrl

`func (o *GithubOrganization) GetHooksUrl() string`

GetHooksUrl returns the HooksUrl field if non-nil, zero value otherwise.

### GetHooksUrlOk

`func (o *GithubOrganization) GetHooksUrlOk() (*string, bool)`

GetHooksUrlOk returns a tuple with the HooksUrl field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetHooksUrl

`func (o *GithubOrganization) SetHooksUrl(v string)`

SetHooksUrl sets HooksUrl field to given value.

### HasHooksUrl

`func (o *GithubOrganization) HasHooksUrl() bool`

HasHooksUrl returns a boolean if a field has been set.

### GetHtmlUrl

`func (o *GithubOrganization) GetHtmlUrl() string`

GetHtmlUrl returns the HtmlUrl field if non-nil, zero value otherwise.

### GetHtmlUrlOk

`func (o *GithubOrganization) GetHtmlUrlOk() (*string, bool)`

GetHtmlUrlOk returns a tuple with the HtmlUrl field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetHtmlUrl

`func (o *GithubOrganization) SetHtmlUrl(v string)`

SetHtmlUrl sets HtmlUrl field to given value.

### HasHtmlUrl

`func (o *GithubOrganization) HasHtmlUrl() bool`

HasHtmlUrl returns a boolean if a field has been set.

### GetId

`func (o *GithubOrganization) GetId() int32`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *GithubOrganization) GetIdOk() (*int32, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *GithubOrganization) SetId(v int32)`

SetId sets Id field to given value.

### HasId

`func (o *GithubOrganization) HasId() bool`

HasId returns a boolean if a field has been set.

### GetIssuesUrl

`func (o *GithubOrganization) GetIssuesUrl() string`

GetIssuesUrl returns the IssuesUrl field if non-nil, zero value otherwise.

### GetIssuesUrlOk

`func (o *GithubOrganization) GetIssuesUrlOk() (*string, bool)`

GetIssuesUrlOk returns a tuple with the IssuesUrl field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIssuesUrl

`func (o *GithubOrganization) SetIssuesUrl(v string)`

SetIssuesUrl sets IssuesUrl field to given value.

### HasIssuesUrl

`func (o *GithubOrganization) HasIssuesUrl() bool`

HasIssuesUrl returns a boolean if a field has been set.

### GetLocation

`func (o *GithubOrganization) GetLocation() string`

GetLocation returns the Location field if non-nil, zero value otherwise.

### GetLocationOk

`func (o *GithubOrganization) GetLocationOk() (*string, bool)`

GetLocationOk returns a tuple with the Location field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLocation

`func (o *GithubOrganization) SetLocation(v string)`

SetLocation sets Location field to given value.

### HasLocation

`func (o *GithubOrganization) HasLocation() bool`

HasLocation returns a boolean if a field has been set.

### GetLogin

`func (o *GithubOrganization) GetLogin() string`

GetLogin returns the Login field if non-nil, zero value otherwise.

### GetLoginOk

`func (o *GithubOrganization) GetLoginOk() (*string, bool)`

GetLoginOk returns a tuple with the Login field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLogin

`func (o *GithubOrganization) SetLogin(v string)`

SetLogin sets Login field to given value.

### HasLogin

`func (o *GithubOrganization) HasLogin() bool`

HasLogin returns a boolean if a field has been set.

### GetMembersUrl

`func (o *GithubOrganization) GetMembersUrl() string`

GetMembersUrl returns the MembersUrl field if non-nil, zero value otherwise.

### GetMembersUrlOk

`func (o *GithubOrganization) GetMembersUrlOk() (*string, bool)`

GetMembersUrlOk returns a tuple with the MembersUrl field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMembersUrl

`func (o *GithubOrganization) SetMembersUrl(v string)`

SetMembersUrl sets MembersUrl field to given value.

### HasMembersUrl

`func (o *GithubOrganization) HasMembersUrl() bool`

HasMembersUrl returns a boolean if a field has been set.

### GetName

`func (o *GithubOrganization) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *GithubOrganization) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *GithubOrganization) SetName(v string)`

SetName sets Name field to given value.

### HasName

`func (o *GithubOrganization) HasName() bool`

HasName returns a boolean if a field has been set.

### GetNodeId

`func (o *GithubOrganization) GetNodeId() string`

GetNodeId returns the NodeId field if non-nil, zero value otherwise.

### GetNodeIdOk

`func (o *GithubOrganization) GetNodeIdOk() (*string, bool)`

GetNodeIdOk returns a tuple with the NodeId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodeId

`func (o *GithubOrganization) SetNodeId(v string)`

SetNodeId sets NodeId field to given value.

### HasNodeId

`func (o *GithubOrganization) HasNodeId() bool`

HasNodeId returns a boolean if a field has been set.

### GetOwnedPrivateRepos

`func (o *GithubOrganization) GetOwnedPrivateRepos() int32`

GetOwnedPrivateRepos returns the OwnedPrivateRepos field if non-nil, zero value otherwise.

### GetOwnedPrivateReposOk

`func (o *GithubOrganization) GetOwnedPrivateReposOk() (*int32, bool)`

GetOwnedPrivateReposOk returns a tuple with the OwnedPrivateRepos field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOwnedPrivateRepos

`func (o *GithubOrganization) SetOwnedPrivateRepos(v int32)`

SetOwnedPrivateRepos sets OwnedPrivateRepos field to given value.

### HasOwnedPrivateRepos

`func (o *GithubOrganization) HasOwnedPrivateRepos() bool`

HasOwnedPrivateRepos returns a boolean if a field has been set.

### GetPlan

`func (o *GithubOrganization) GetPlan() GithubPlan`

GetPlan returns the Plan field if non-nil, zero value otherwise.

### GetPlanOk

`func (o *GithubOrganization) GetPlanOk() (*GithubPlan, bool)`

GetPlanOk returns a tuple with the Plan field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPlan

`func (o *GithubOrganization) SetPlan(v GithubPlan)`

SetPlan sets Plan field to given value.

### HasPlan

`func (o *GithubOrganization) HasPlan() bool`

HasPlan returns a boolean if a field has been set.

### GetPrivateGists

`func (o *GithubOrganization) GetPrivateGists() int32`

GetPrivateGists returns the PrivateGists field if non-nil, zero value otherwise.

### GetPrivateGistsOk

`func (o *GithubOrganization) GetPrivateGistsOk() (*int32, bool)`

GetPrivateGistsOk returns a tuple with the PrivateGists field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPrivateGists

`func (o *GithubOrganization) SetPrivateGists(v int32)`

SetPrivateGists sets PrivateGists field to given value.

### HasPrivateGists

`func (o *GithubOrganization) HasPrivateGists() bool`

HasPrivateGists returns a boolean if a field has been set.

### GetPublicGists

`func (o *GithubOrganization) GetPublicGists() int32`

GetPublicGists returns the PublicGists field if non-nil, zero value otherwise.

### GetPublicGistsOk

`func (o *GithubOrganization) GetPublicGistsOk() (*int32, bool)`

GetPublicGistsOk returns a tuple with the PublicGists field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPublicGists

`func (o *GithubOrganization) SetPublicGists(v int32)`

SetPublicGists sets PublicGists field to given value.

### HasPublicGists

`func (o *GithubOrganization) HasPublicGists() bool`

HasPublicGists returns a boolean if a field has been set.

### GetPublicMembersUrl

`func (o *GithubOrganization) GetPublicMembersUrl() string`

GetPublicMembersUrl returns the PublicMembersUrl field if non-nil, zero value otherwise.

### GetPublicMembersUrlOk

`func (o *GithubOrganization) GetPublicMembersUrlOk() (*string, bool)`

GetPublicMembersUrlOk returns a tuple with the PublicMembersUrl field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPublicMembersUrl

`func (o *GithubOrganization) SetPublicMembersUrl(v string)`

SetPublicMembersUrl sets PublicMembersUrl field to given value.

### HasPublicMembersUrl

`func (o *GithubOrganization) HasPublicMembersUrl() bool`

HasPublicMembersUrl returns a boolean if a field has been set.

### GetPublicRepos

`func (o *GithubOrganization) GetPublicRepos() int32`

GetPublicRepos returns the PublicRepos field if non-nil, zero value otherwise.

### GetPublicReposOk

`func (o *GithubOrganization) GetPublicReposOk() (*int32, bool)`

GetPublicReposOk returns a tuple with the PublicRepos field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPublicRepos

`func (o *GithubOrganization) SetPublicRepos(v int32)`

SetPublicRepos sets PublicRepos field to given value.

### HasPublicRepos

`func (o *GithubOrganization) HasPublicRepos() bool`

HasPublicRepos returns a boolean if a field has been set.

### GetReposUrl

`func (o *GithubOrganization) GetReposUrl() string`

GetReposUrl returns the ReposUrl field if non-nil, zero value otherwise.

### GetReposUrlOk

`func (o *GithubOrganization) GetReposUrlOk() (*string, bool)`

GetReposUrlOk returns a tuple with the ReposUrl field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetReposUrl

`func (o *GithubOrganization) SetReposUrl(v string)`

SetReposUrl sets ReposUrl field to given value.

### HasReposUrl

`func (o *GithubOrganization) HasReposUrl() bool`

HasReposUrl returns a boolean if a field has been set.

### GetTotalPrivateRepos

`func (o *GithubOrganization) GetTotalPrivateRepos() int32`

GetTotalPrivateRepos returns the TotalPrivateRepos field if non-nil, zero value otherwise.

### GetTotalPrivateReposOk

`func (o *GithubOrganization) GetTotalPrivateReposOk() (*int32, bool)`

GetTotalPrivateReposOk returns a tuple with the TotalPrivateRepos field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTotalPrivateRepos

`func (o *GithubOrganization) SetTotalPrivateRepos(v int32)`

SetTotalPrivateRepos sets TotalPrivateRepos field to given value.

### HasTotalPrivateRepos

`func (o *GithubOrganization) HasTotalPrivateRepos() bool`

HasTotalPrivateRepos returns a boolean if a field has been set.

### GetType

`func (o *GithubOrganization) GetType() string`

GetType returns the Type field if non-nil, zero value otherwise.

### GetTypeOk

`func (o *GithubOrganization) GetTypeOk() (*string, bool)`

GetTypeOk returns a tuple with the Type field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetType

`func (o *GithubOrganization) SetType(v string)`

SetType sets Type field to given value.

### HasType

`func (o *GithubOrganization) HasType() bool`

HasType returns a boolean if a field has been set.

### GetUpdatedAt

`func (o *GithubOrganization) GetUpdatedAt() string`

GetUpdatedAt returns the UpdatedAt field if non-nil, zero value otherwise.

### GetUpdatedAtOk

`func (o *GithubOrganization) GetUpdatedAtOk() (*string, bool)`

GetUpdatedAtOk returns a tuple with the UpdatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdatedAt

`func (o *GithubOrganization) SetUpdatedAt(v string)`

SetUpdatedAt sets UpdatedAt field to given value.

### HasUpdatedAt

`func (o *GithubOrganization) HasUpdatedAt() bool`

HasUpdatedAt returns a boolean if a field has been set.

### GetUrl

`func (o *GithubOrganization) GetUrl() string`

GetUrl returns the Url field if non-nil, zero value otherwise.

### GetUrlOk

`func (o *GithubOrganization) GetUrlOk() (*string, bool)`

GetUrlOk returns a tuple with the Url field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUrl

`func (o *GithubOrganization) SetUrl(v string)`

SetUrl sets Url field to given value.

### HasUrl

`func (o *GithubOrganization) HasUrl() bool`

HasUrl returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


