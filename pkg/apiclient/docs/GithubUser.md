# GithubUser

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**AvatarUrl** | Pointer to **string** |  | [optional] 
**Bio** | Pointer to **string** |  | [optional] 
**Blog** | Pointer to **string** |  | [optional] 
**Collaborators** | Pointer to **int32** |  | [optional] 
**Company** | Pointer to **string** |  | [optional] 
**CreatedAt** | Pointer to [**GithubTimestamp**](GithubTimestamp.md) |  | [optional] 
**DiskUsage** | Pointer to **int32** |  | [optional] 
**Email** | Pointer to **string** |  | [optional] 
**EventsUrl** | Pointer to **string** |  | [optional] 
**Followers** | Pointer to **int32** |  | [optional] 
**FollowersUrl** | Pointer to **string** |  | [optional] 
**Following** | Pointer to **int32** |  | [optional] 
**FollowingUrl** | Pointer to **string** |  | [optional] 
**GistsUrl** | Pointer to **string** |  | [optional] 
**GravatarId** | Pointer to **string** |  | [optional] 
**Hireable** | Pointer to **bool** |  | [optional] 
**HtmlUrl** | Pointer to **string** |  | [optional] 
**Id** | Pointer to **int32** |  | [optional] 
**Location** | Pointer to **string** |  | [optional] 
**Login** | Pointer to **string** |  | [optional] 
**Name** | Pointer to **string** |  | [optional] 
**NodeId** | Pointer to **string** |  | [optional] 
**OrganizationsUrl** | Pointer to **string** |  | [optional] 
**OwnedPrivateRepos** | Pointer to **int32** |  | [optional] 
**Permissions** | Pointer to **map[string]bool** | Permissions identifies the permissions that a user has on a given repository. This is only populated when calling Repositories.ListCollaborators. | [optional] 
**Plan** | Pointer to [**GithubPlan**](GithubPlan.md) |  | [optional] 
**PrivateGists** | Pointer to **int32** |  | [optional] 
**PublicGists** | Pointer to **int32** |  | [optional] 
**PublicRepos** | Pointer to **int32** |  | [optional] 
**ReceivedEventsUrl** | Pointer to **string** |  | [optional] 
**ReposUrl** | Pointer to **string** |  | [optional] 
**SiteAdmin** | Pointer to **bool** |  | [optional] 
**StarredUrl** | Pointer to **string** |  | [optional] 
**SubscriptionsUrl** | Pointer to **string** |  | [optional] 
**SuspendedAt** | Pointer to [**GithubTimestamp**](GithubTimestamp.md) |  | [optional] 
**TextMatches** | Pointer to [**[]GithubTextMatch**](GithubTextMatch.md) | TextMatches is only populated from search results that request text matches See: search.go and https://developer.github.com/v3/search/#text-match-metadata | [optional] 
**TotalPrivateRepos** | Pointer to **int32** |  | [optional] 
**Type** | Pointer to **string** |  | [optional] 
**UpdatedAt** | Pointer to [**GithubTimestamp**](GithubTimestamp.md) |  | [optional] 
**Url** | Pointer to **string** | API URLs | [optional] 

## Methods

### NewGithubUser

`func NewGithubUser() *GithubUser`

NewGithubUser instantiates a new GithubUser object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewGithubUserWithDefaults

`func NewGithubUserWithDefaults() *GithubUser`

NewGithubUserWithDefaults instantiates a new GithubUser object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetAvatarUrl

`func (o *GithubUser) GetAvatarUrl() string`

GetAvatarUrl returns the AvatarUrl field if non-nil, zero value otherwise.

### GetAvatarUrlOk

`func (o *GithubUser) GetAvatarUrlOk() (*string, bool)`

GetAvatarUrlOk returns a tuple with the AvatarUrl field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAvatarUrl

`func (o *GithubUser) SetAvatarUrl(v string)`

SetAvatarUrl sets AvatarUrl field to given value.

### HasAvatarUrl

`func (o *GithubUser) HasAvatarUrl() bool`

HasAvatarUrl returns a boolean if a field has been set.

### GetBio

`func (o *GithubUser) GetBio() string`

GetBio returns the Bio field if non-nil, zero value otherwise.

### GetBioOk

`func (o *GithubUser) GetBioOk() (*string, bool)`

GetBioOk returns a tuple with the Bio field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetBio

`func (o *GithubUser) SetBio(v string)`

SetBio sets Bio field to given value.

### HasBio

`func (o *GithubUser) HasBio() bool`

HasBio returns a boolean if a field has been set.

### GetBlog

`func (o *GithubUser) GetBlog() string`

GetBlog returns the Blog field if non-nil, zero value otherwise.

### GetBlogOk

`func (o *GithubUser) GetBlogOk() (*string, bool)`

GetBlogOk returns a tuple with the Blog field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetBlog

`func (o *GithubUser) SetBlog(v string)`

SetBlog sets Blog field to given value.

### HasBlog

`func (o *GithubUser) HasBlog() bool`

HasBlog returns a boolean if a field has been set.

### GetCollaborators

`func (o *GithubUser) GetCollaborators() int32`

GetCollaborators returns the Collaborators field if non-nil, zero value otherwise.

### GetCollaboratorsOk

`func (o *GithubUser) GetCollaboratorsOk() (*int32, bool)`

GetCollaboratorsOk returns a tuple with the Collaborators field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCollaborators

`func (o *GithubUser) SetCollaborators(v int32)`

SetCollaborators sets Collaborators field to given value.

### HasCollaborators

`func (o *GithubUser) HasCollaborators() bool`

HasCollaborators returns a boolean if a field has been set.

### GetCompany

`func (o *GithubUser) GetCompany() string`

GetCompany returns the Company field if non-nil, zero value otherwise.

### GetCompanyOk

`func (o *GithubUser) GetCompanyOk() (*string, bool)`

GetCompanyOk returns a tuple with the Company field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCompany

`func (o *GithubUser) SetCompany(v string)`

SetCompany sets Company field to given value.

### HasCompany

`func (o *GithubUser) HasCompany() bool`

HasCompany returns a boolean if a field has been set.

### GetCreatedAt

`func (o *GithubUser) GetCreatedAt() GithubTimestamp`

GetCreatedAt returns the CreatedAt field if non-nil, zero value otherwise.

### GetCreatedAtOk

`func (o *GithubUser) GetCreatedAtOk() (*GithubTimestamp, bool)`

GetCreatedAtOk returns a tuple with the CreatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedAt

`func (o *GithubUser) SetCreatedAt(v GithubTimestamp)`

SetCreatedAt sets CreatedAt field to given value.

### HasCreatedAt

`func (o *GithubUser) HasCreatedAt() bool`

HasCreatedAt returns a boolean if a field has been set.

### GetDiskUsage

`func (o *GithubUser) GetDiskUsage() int32`

GetDiskUsage returns the DiskUsage field if non-nil, zero value otherwise.

### GetDiskUsageOk

`func (o *GithubUser) GetDiskUsageOk() (*int32, bool)`

GetDiskUsageOk returns a tuple with the DiskUsage field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDiskUsage

`func (o *GithubUser) SetDiskUsage(v int32)`

SetDiskUsage sets DiskUsage field to given value.

### HasDiskUsage

`func (o *GithubUser) HasDiskUsage() bool`

HasDiskUsage returns a boolean if a field has been set.

### GetEmail

`func (o *GithubUser) GetEmail() string`

GetEmail returns the Email field if non-nil, zero value otherwise.

### GetEmailOk

`func (o *GithubUser) GetEmailOk() (*string, bool)`

GetEmailOk returns a tuple with the Email field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEmail

`func (o *GithubUser) SetEmail(v string)`

SetEmail sets Email field to given value.

### HasEmail

`func (o *GithubUser) HasEmail() bool`

HasEmail returns a boolean if a field has been set.

### GetEventsUrl

`func (o *GithubUser) GetEventsUrl() string`

GetEventsUrl returns the EventsUrl field if non-nil, zero value otherwise.

### GetEventsUrlOk

`func (o *GithubUser) GetEventsUrlOk() (*string, bool)`

GetEventsUrlOk returns a tuple with the EventsUrl field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEventsUrl

`func (o *GithubUser) SetEventsUrl(v string)`

SetEventsUrl sets EventsUrl field to given value.

### HasEventsUrl

`func (o *GithubUser) HasEventsUrl() bool`

HasEventsUrl returns a boolean if a field has been set.

### GetFollowers

`func (o *GithubUser) GetFollowers() int32`

GetFollowers returns the Followers field if non-nil, zero value otherwise.

### GetFollowersOk

`func (o *GithubUser) GetFollowersOk() (*int32, bool)`

GetFollowersOk returns a tuple with the Followers field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFollowers

`func (o *GithubUser) SetFollowers(v int32)`

SetFollowers sets Followers field to given value.

### HasFollowers

`func (o *GithubUser) HasFollowers() bool`

HasFollowers returns a boolean if a field has been set.

### GetFollowersUrl

`func (o *GithubUser) GetFollowersUrl() string`

GetFollowersUrl returns the FollowersUrl field if non-nil, zero value otherwise.

### GetFollowersUrlOk

`func (o *GithubUser) GetFollowersUrlOk() (*string, bool)`

GetFollowersUrlOk returns a tuple with the FollowersUrl field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFollowersUrl

`func (o *GithubUser) SetFollowersUrl(v string)`

SetFollowersUrl sets FollowersUrl field to given value.

### HasFollowersUrl

`func (o *GithubUser) HasFollowersUrl() bool`

HasFollowersUrl returns a boolean if a field has been set.

### GetFollowing

`func (o *GithubUser) GetFollowing() int32`

GetFollowing returns the Following field if non-nil, zero value otherwise.

### GetFollowingOk

`func (o *GithubUser) GetFollowingOk() (*int32, bool)`

GetFollowingOk returns a tuple with the Following field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFollowing

`func (o *GithubUser) SetFollowing(v int32)`

SetFollowing sets Following field to given value.

### HasFollowing

`func (o *GithubUser) HasFollowing() bool`

HasFollowing returns a boolean if a field has been set.

### GetFollowingUrl

`func (o *GithubUser) GetFollowingUrl() string`

GetFollowingUrl returns the FollowingUrl field if non-nil, zero value otherwise.

### GetFollowingUrlOk

`func (o *GithubUser) GetFollowingUrlOk() (*string, bool)`

GetFollowingUrlOk returns a tuple with the FollowingUrl field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFollowingUrl

`func (o *GithubUser) SetFollowingUrl(v string)`

SetFollowingUrl sets FollowingUrl field to given value.

### HasFollowingUrl

`func (o *GithubUser) HasFollowingUrl() bool`

HasFollowingUrl returns a boolean if a field has been set.

### GetGistsUrl

`func (o *GithubUser) GetGistsUrl() string`

GetGistsUrl returns the GistsUrl field if non-nil, zero value otherwise.

### GetGistsUrlOk

`func (o *GithubUser) GetGistsUrlOk() (*string, bool)`

GetGistsUrlOk returns a tuple with the GistsUrl field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGistsUrl

`func (o *GithubUser) SetGistsUrl(v string)`

SetGistsUrl sets GistsUrl field to given value.

### HasGistsUrl

`func (o *GithubUser) HasGistsUrl() bool`

HasGistsUrl returns a boolean if a field has been set.

### GetGravatarId

`func (o *GithubUser) GetGravatarId() string`

GetGravatarId returns the GravatarId field if non-nil, zero value otherwise.

### GetGravatarIdOk

`func (o *GithubUser) GetGravatarIdOk() (*string, bool)`

GetGravatarIdOk returns a tuple with the GravatarId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGravatarId

`func (o *GithubUser) SetGravatarId(v string)`

SetGravatarId sets GravatarId field to given value.

### HasGravatarId

`func (o *GithubUser) HasGravatarId() bool`

HasGravatarId returns a boolean if a field has been set.

### GetHireable

`func (o *GithubUser) GetHireable() bool`

GetHireable returns the Hireable field if non-nil, zero value otherwise.

### GetHireableOk

`func (o *GithubUser) GetHireableOk() (*bool, bool)`

GetHireableOk returns a tuple with the Hireable field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetHireable

`func (o *GithubUser) SetHireable(v bool)`

SetHireable sets Hireable field to given value.

### HasHireable

`func (o *GithubUser) HasHireable() bool`

HasHireable returns a boolean if a field has been set.

### GetHtmlUrl

`func (o *GithubUser) GetHtmlUrl() string`

GetHtmlUrl returns the HtmlUrl field if non-nil, zero value otherwise.

### GetHtmlUrlOk

`func (o *GithubUser) GetHtmlUrlOk() (*string, bool)`

GetHtmlUrlOk returns a tuple with the HtmlUrl field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetHtmlUrl

`func (o *GithubUser) SetHtmlUrl(v string)`

SetHtmlUrl sets HtmlUrl field to given value.

### HasHtmlUrl

`func (o *GithubUser) HasHtmlUrl() bool`

HasHtmlUrl returns a boolean if a field has been set.

### GetId

`func (o *GithubUser) GetId() int32`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *GithubUser) GetIdOk() (*int32, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *GithubUser) SetId(v int32)`

SetId sets Id field to given value.

### HasId

`func (o *GithubUser) HasId() bool`

HasId returns a boolean if a field has been set.

### GetLocation

`func (o *GithubUser) GetLocation() string`

GetLocation returns the Location field if non-nil, zero value otherwise.

### GetLocationOk

`func (o *GithubUser) GetLocationOk() (*string, bool)`

GetLocationOk returns a tuple with the Location field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLocation

`func (o *GithubUser) SetLocation(v string)`

SetLocation sets Location field to given value.

### HasLocation

`func (o *GithubUser) HasLocation() bool`

HasLocation returns a boolean if a field has been set.

### GetLogin

`func (o *GithubUser) GetLogin() string`

GetLogin returns the Login field if non-nil, zero value otherwise.

### GetLoginOk

`func (o *GithubUser) GetLoginOk() (*string, bool)`

GetLoginOk returns a tuple with the Login field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLogin

`func (o *GithubUser) SetLogin(v string)`

SetLogin sets Login field to given value.

### HasLogin

`func (o *GithubUser) HasLogin() bool`

HasLogin returns a boolean if a field has been set.

### GetName

`func (o *GithubUser) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *GithubUser) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *GithubUser) SetName(v string)`

SetName sets Name field to given value.

### HasName

`func (o *GithubUser) HasName() bool`

HasName returns a boolean if a field has been set.

### GetNodeId

`func (o *GithubUser) GetNodeId() string`

GetNodeId returns the NodeId field if non-nil, zero value otherwise.

### GetNodeIdOk

`func (o *GithubUser) GetNodeIdOk() (*string, bool)`

GetNodeIdOk returns a tuple with the NodeId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodeId

`func (o *GithubUser) SetNodeId(v string)`

SetNodeId sets NodeId field to given value.

### HasNodeId

`func (o *GithubUser) HasNodeId() bool`

HasNodeId returns a boolean if a field has been set.

### GetOrganizationsUrl

`func (o *GithubUser) GetOrganizationsUrl() string`

GetOrganizationsUrl returns the OrganizationsUrl field if non-nil, zero value otherwise.

### GetOrganizationsUrlOk

`func (o *GithubUser) GetOrganizationsUrlOk() (*string, bool)`

GetOrganizationsUrlOk returns a tuple with the OrganizationsUrl field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOrganizationsUrl

`func (o *GithubUser) SetOrganizationsUrl(v string)`

SetOrganizationsUrl sets OrganizationsUrl field to given value.

### HasOrganizationsUrl

`func (o *GithubUser) HasOrganizationsUrl() bool`

HasOrganizationsUrl returns a boolean if a field has been set.

### GetOwnedPrivateRepos

`func (o *GithubUser) GetOwnedPrivateRepos() int32`

GetOwnedPrivateRepos returns the OwnedPrivateRepos field if non-nil, zero value otherwise.

### GetOwnedPrivateReposOk

`func (o *GithubUser) GetOwnedPrivateReposOk() (*int32, bool)`

GetOwnedPrivateReposOk returns a tuple with the OwnedPrivateRepos field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOwnedPrivateRepos

`func (o *GithubUser) SetOwnedPrivateRepos(v int32)`

SetOwnedPrivateRepos sets OwnedPrivateRepos field to given value.

### HasOwnedPrivateRepos

`func (o *GithubUser) HasOwnedPrivateRepos() bool`

HasOwnedPrivateRepos returns a boolean if a field has been set.

### GetPermissions

`func (o *GithubUser) GetPermissions() map[string]bool`

GetPermissions returns the Permissions field if non-nil, zero value otherwise.

### GetPermissionsOk

`func (o *GithubUser) GetPermissionsOk() (*map[string]bool, bool)`

GetPermissionsOk returns a tuple with the Permissions field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPermissions

`func (o *GithubUser) SetPermissions(v map[string]bool)`

SetPermissions sets Permissions field to given value.

### HasPermissions

`func (o *GithubUser) HasPermissions() bool`

HasPermissions returns a boolean if a field has been set.

### GetPlan

`func (o *GithubUser) GetPlan() GithubPlan`

GetPlan returns the Plan field if non-nil, zero value otherwise.

### GetPlanOk

`func (o *GithubUser) GetPlanOk() (*GithubPlan, bool)`

GetPlanOk returns a tuple with the Plan field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPlan

`func (o *GithubUser) SetPlan(v GithubPlan)`

SetPlan sets Plan field to given value.

### HasPlan

`func (o *GithubUser) HasPlan() bool`

HasPlan returns a boolean if a field has been set.

### GetPrivateGists

`func (o *GithubUser) GetPrivateGists() int32`

GetPrivateGists returns the PrivateGists field if non-nil, zero value otherwise.

### GetPrivateGistsOk

`func (o *GithubUser) GetPrivateGistsOk() (*int32, bool)`

GetPrivateGistsOk returns a tuple with the PrivateGists field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPrivateGists

`func (o *GithubUser) SetPrivateGists(v int32)`

SetPrivateGists sets PrivateGists field to given value.

### HasPrivateGists

`func (o *GithubUser) HasPrivateGists() bool`

HasPrivateGists returns a boolean if a field has been set.

### GetPublicGists

`func (o *GithubUser) GetPublicGists() int32`

GetPublicGists returns the PublicGists field if non-nil, zero value otherwise.

### GetPublicGistsOk

`func (o *GithubUser) GetPublicGistsOk() (*int32, bool)`

GetPublicGistsOk returns a tuple with the PublicGists field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPublicGists

`func (o *GithubUser) SetPublicGists(v int32)`

SetPublicGists sets PublicGists field to given value.

### HasPublicGists

`func (o *GithubUser) HasPublicGists() bool`

HasPublicGists returns a boolean if a field has been set.

### GetPublicRepos

`func (o *GithubUser) GetPublicRepos() int32`

GetPublicRepos returns the PublicRepos field if non-nil, zero value otherwise.

### GetPublicReposOk

`func (o *GithubUser) GetPublicReposOk() (*int32, bool)`

GetPublicReposOk returns a tuple with the PublicRepos field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPublicRepos

`func (o *GithubUser) SetPublicRepos(v int32)`

SetPublicRepos sets PublicRepos field to given value.

### HasPublicRepos

`func (o *GithubUser) HasPublicRepos() bool`

HasPublicRepos returns a boolean if a field has been set.

### GetReceivedEventsUrl

`func (o *GithubUser) GetReceivedEventsUrl() string`

GetReceivedEventsUrl returns the ReceivedEventsUrl field if non-nil, zero value otherwise.

### GetReceivedEventsUrlOk

`func (o *GithubUser) GetReceivedEventsUrlOk() (*string, bool)`

GetReceivedEventsUrlOk returns a tuple with the ReceivedEventsUrl field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetReceivedEventsUrl

`func (o *GithubUser) SetReceivedEventsUrl(v string)`

SetReceivedEventsUrl sets ReceivedEventsUrl field to given value.

### HasReceivedEventsUrl

`func (o *GithubUser) HasReceivedEventsUrl() bool`

HasReceivedEventsUrl returns a boolean if a field has been set.

### GetReposUrl

`func (o *GithubUser) GetReposUrl() string`

GetReposUrl returns the ReposUrl field if non-nil, zero value otherwise.

### GetReposUrlOk

`func (o *GithubUser) GetReposUrlOk() (*string, bool)`

GetReposUrlOk returns a tuple with the ReposUrl field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetReposUrl

`func (o *GithubUser) SetReposUrl(v string)`

SetReposUrl sets ReposUrl field to given value.

### HasReposUrl

`func (o *GithubUser) HasReposUrl() bool`

HasReposUrl returns a boolean if a field has been set.

### GetSiteAdmin

`func (o *GithubUser) GetSiteAdmin() bool`

GetSiteAdmin returns the SiteAdmin field if non-nil, zero value otherwise.

### GetSiteAdminOk

`func (o *GithubUser) GetSiteAdminOk() (*bool, bool)`

GetSiteAdminOk returns a tuple with the SiteAdmin field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSiteAdmin

`func (o *GithubUser) SetSiteAdmin(v bool)`

SetSiteAdmin sets SiteAdmin field to given value.

### HasSiteAdmin

`func (o *GithubUser) HasSiteAdmin() bool`

HasSiteAdmin returns a boolean if a field has been set.

### GetStarredUrl

`func (o *GithubUser) GetStarredUrl() string`

GetStarredUrl returns the StarredUrl field if non-nil, zero value otherwise.

### GetStarredUrlOk

`func (o *GithubUser) GetStarredUrlOk() (*string, bool)`

GetStarredUrlOk returns a tuple with the StarredUrl field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStarredUrl

`func (o *GithubUser) SetStarredUrl(v string)`

SetStarredUrl sets StarredUrl field to given value.

### HasStarredUrl

`func (o *GithubUser) HasStarredUrl() bool`

HasStarredUrl returns a boolean if a field has been set.

### GetSubscriptionsUrl

`func (o *GithubUser) GetSubscriptionsUrl() string`

GetSubscriptionsUrl returns the SubscriptionsUrl field if non-nil, zero value otherwise.

### GetSubscriptionsUrlOk

`func (o *GithubUser) GetSubscriptionsUrlOk() (*string, bool)`

GetSubscriptionsUrlOk returns a tuple with the SubscriptionsUrl field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSubscriptionsUrl

`func (o *GithubUser) SetSubscriptionsUrl(v string)`

SetSubscriptionsUrl sets SubscriptionsUrl field to given value.

### HasSubscriptionsUrl

`func (o *GithubUser) HasSubscriptionsUrl() bool`

HasSubscriptionsUrl returns a boolean if a field has been set.

### GetSuspendedAt

`func (o *GithubUser) GetSuspendedAt() GithubTimestamp`

GetSuspendedAt returns the SuspendedAt field if non-nil, zero value otherwise.

### GetSuspendedAtOk

`func (o *GithubUser) GetSuspendedAtOk() (*GithubTimestamp, bool)`

GetSuspendedAtOk returns a tuple with the SuspendedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSuspendedAt

`func (o *GithubUser) SetSuspendedAt(v GithubTimestamp)`

SetSuspendedAt sets SuspendedAt field to given value.

### HasSuspendedAt

`func (o *GithubUser) HasSuspendedAt() bool`

HasSuspendedAt returns a boolean if a field has been set.

### GetTextMatches

`func (o *GithubUser) GetTextMatches() []GithubTextMatch`

GetTextMatches returns the TextMatches field if non-nil, zero value otherwise.

### GetTextMatchesOk

`func (o *GithubUser) GetTextMatchesOk() (*[]GithubTextMatch, bool)`

GetTextMatchesOk returns a tuple with the TextMatches field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTextMatches

`func (o *GithubUser) SetTextMatches(v []GithubTextMatch)`

SetTextMatches sets TextMatches field to given value.

### HasTextMatches

`func (o *GithubUser) HasTextMatches() bool`

HasTextMatches returns a boolean if a field has been set.

### GetTotalPrivateRepos

`func (o *GithubUser) GetTotalPrivateRepos() int32`

GetTotalPrivateRepos returns the TotalPrivateRepos field if non-nil, zero value otherwise.

### GetTotalPrivateReposOk

`func (o *GithubUser) GetTotalPrivateReposOk() (*int32, bool)`

GetTotalPrivateReposOk returns a tuple with the TotalPrivateRepos field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTotalPrivateRepos

`func (o *GithubUser) SetTotalPrivateRepos(v int32)`

SetTotalPrivateRepos sets TotalPrivateRepos field to given value.

### HasTotalPrivateRepos

`func (o *GithubUser) HasTotalPrivateRepos() bool`

HasTotalPrivateRepos returns a boolean if a field has been set.

### GetType

`func (o *GithubUser) GetType() string`

GetType returns the Type field if non-nil, zero value otherwise.

### GetTypeOk

`func (o *GithubUser) GetTypeOk() (*string, bool)`

GetTypeOk returns a tuple with the Type field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetType

`func (o *GithubUser) SetType(v string)`

SetType sets Type field to given value.

### HasType

`func (o *GithubUser) HasType() bool`

HasType returns a boolean if a field has been set.

### GetUpdatedAt

`func (o *GithubUser) GetUpdatedAt() GithubTimestamp`

GetUpdatedAt returns the UpdatedAt field if non-nil, zero value otherwise.

### GetUpdatedAtOk

`func (o *GithubUser) GetUpdatedAtOk() (*GithubTimestamp, bool)`

GetUpdatedAtOk returns a tuple with the UpdatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdatedAt

`func (o *GithubUser) SetUpdatedAt(v GithubTimestamp)`

SetUpdatedAt sets UpdatedAt field to given value.

### HasUpdatedAt

`func (o *GithubUser) HasUpdatedAt() bool`

HasUpdatedAt returns a boolean if a field has been set.

### GetUrl

`func (o *GithubUser) GetUrl() string`

GetUrl returns the Url field if non-nil, zero value otherwise.

### GetUrlOk

`func (o *GithubUser) GetUrlOk() (*string, bool)`

GetUrlOk returns a tuple with the Url field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUrl

`func (o *GithubUser) SetUrl(v string)`

SetUrl sets Url field to given value.

### HasUrl

`func (o *GithubUser) HasUrl() bool`

HasUrl returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


