# GithubRepository

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**AllowMergeCommit** | Pointer to **bool** |  | [optional] 
**AllowRebaseMerge** | Pointer to **bool** |  | [optional] 
**AllowSquashMerge** | Pointer to **bool** |  | [optional] 
**ArchiveUrl** | Pointer to **string** |  | [optional] 
**Archived** | Pointer to **bool** |  | [optional] 
**AssigneesUrl** | Pointer to **string** |  | [optional] 
**AutoInit** | Pointer to **bool** |  | [optional] 
**BlobsUrl** | Pointer to **string** |  | [optional] 
**BranchesUrl** | Pointer to **string** |  | [optional] 
**CloneUrl** | Pointer to **string** |  | [optional] 
**CodeOfConduct** | Pointer to [**GithubCodeOfConduct**](GithubCodeOfConduct.md) |  | [optional] 
**CollaboratorsUrl** | Pointer to **string** |  | [optional] 
**CommentsUrl** | Pointer to **string** |  | [optional] 
**CommitsUrl** | Pointer to **string** |  | [optional] 
**CompareUrl** | Pointer to **string** |  | [optional] 
**ContentsUrl** | Pointer to **string** |  | [optional] 
**ContributorsUrl** | Pointer to **string** |  | [optional] 
**CreatedAt** | Pointer to [**GithubTimestamp**](GithubTimestamp.md) |  | [optional] 
**DefaultBranch** | Pointer to **string** |  | [optional] 
**DeploymentsUrl** | Pointer to **string** |  | [optional] 
**Description** | Pointer to **string** |  | [optional] 
**DownloadsUrl** | Pointer to **string** |  | [optional] 
**EventsUrl** | Pointer to **string** |  | [optional] 
**Fork** | Pointer to **bool** |  | [optional] 
**ForksCount** | Pointer to **int32** |  | [optional] 
**ForksUrl** | Pointer to **string** |  | [optional] 
**FullName** | Pointer to **string** |  | [optional] 
**GitCommitsUrl** | Pointer to **string** |  | [optional] 
**GitRefsUrl** | Pointer to **string** |  | [optional] 
**GitTagsUrl** | Pointer to **string** |  | [optional] 
**GitUrl** | Pointer to **string** |  | [optional] 
**GitignoreTemplate** | Pointer to **string** |  | [optional] 
**HasDownloads** | Pointer to **bool** |  | [optional] 
**HasIssues** | Pointer to **bool** |  | [optional] 
**HasPages** | Pointer to **bool** |  | [optional] 
**HasProjects** | Pointer to **bool** |  | [optional] 
**HasWiki** | Pointer to **bool** |  | [optional] 
**Homepage** | Pointer to **string** |  | [optional] 
**HooksUrl** | Pointer to **string** |  | [optional] 
**HtmlUrl** | Pointer to **string** |  | [optional] 
**Id** | Pointer to **int32** |  | [optional] 
**IssueCommentUrl** | Pointer to **string** |  | [optional] 
**IssueEventsUrl** | Pointer to **string** |  | [optional] 
**IssuesUrl** | Pointer to **string** |  | [optional] 
**KeysUrl** | Pointer to **string** |  | [optional] 
**LabelsUrl** | Pointer to **string** |  | [optional] 
**Language** | Pointer to **string** |  | [optional] 
**LanguagesUrl** | Pointer to **string** |  | [optional] 
**License** | Pointer to [**GithubLicense**](GithubLicense.md) | Only provided when using RepositoriesService.Get while in preview | [optional] 
**LicenseTemplate** | Pointer to **string** |  | [optional] 
**MasterBranch** | Pointer to **string** |  | [optional] 
**MergesUrl** | Pointer to **string** |  | [optional] 
**MilestonesUrl** | Pointer to **string** |  | [optional] 
**MirrorUrl** | Pointer to **string** |  | [optional] 
**Name** | Pointer to **string** |  | [optional] 
**NetworkCount** | Pointer to **int32** |  | [optional] 
**NodeId** | Pointer to **string** |  | [optional] 
**NotificationsUrl** | Pointer to **string** |  | [optional] 
**OpenIssuesCount** | Pointer to **int32** |  | [optional] 
**Organization** | Pointer to [**GithubOrganization**](GithubOrganization.md) |  | [optional] 
**Owner** | Pointer to [**GithubUser**](GithubUser.md) |  | [optional] 
**Parent** | Pointer to [**GithubRepository**](GithubRepository.md) |  | [optional] 
**Permissions** | Pointer to **map[string]bool** |  | [optional] 
**Private** | Pointer to **bool** | Additional mutable fields when creating and editing a repository | [optional] 
**PullsUrl** | Pointer to **string** |  | [optional] 
**PushedAt** | Pointer to [**GithubTimestamp**](GithubTimestamp.md) |  | [optional] 
**ReleasesUrl** | Pointer to **string** |  | [optional] 
**Size** | Pointer to **int32** |  | [optional] 
**Source** | Pointer to [**GithubRepository**](GithubRepository.md) |  | [optional] 
**SshUrl** | Pointer to **string** |  | [optional] 
**StargazersCount** | Pointer to **int32** |  | [optional] 
**StargazersUrl** | Pointer to **string** |  | [optional] 
**StatusesUrl** | Pointer to **string** |  | [optional] 
**SubscribersCount** | Pointer to **int32** |  | [optional] 
**SubscribersUrl** | Pointer to **string** |  | [optional] 
**SubscriptionUrl** | Pointer to **string** |  | [optional] 
**SvnUrl** | Pointer to **string** |  | [optional] 
**TagsUrl** | Pointer to **string** |  | [optional] 
**TeamId** | Pointer to **int32** | Creating an organization repository. Required for non-owners. | [optional] 
**TeamsUrl** | Pointer to **string** |  | [optional] 
**TextMatches** | Pointer to [**[]GithubTextMatch**](GithubTextMatch.md) | TextMatches is only populated from search results that request text matches See: search.go and https://developer.github.com/v3/search/#text-match-metadata | [optional] 
**Topics** | Pointer to **[]string** |  | [optional] 
**TreesUrl** | Pointer to **string** |  | [optional] 
**UpdatedAt** | Pointer to [**GithubTimestamp**](GithubTimestamp.md) |  | [optional] 
**Url** | Pointer to **string** | API URLs | [optional] 
**WatchersCount** | Pointer to **int32** |  | [optional] 

## Methods

### NewGithubRepository

`func NewGithubRepository() *GithubRepository`

NewGithubRepository instantiates a new GithubRepository object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewGithubRepositoryWithDefaults

`func NewGithubRepositoryWithDefaults() *GithubRepository`

NewGithubRepositoryWithDefaults instantiates a new GithubRepository object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetAllowMergeCommit

`func (o *GithubRepository) GetAllowMergeCommit() bool`

GetAllowMergeCommit returns the AllowMergeCommit field if non-nil, zero value otherwise.

### GetAllowMergeCommitOk

`func (o *GithubRepository) GetAllowMergeCommitOk() (*bool, bool)`

GetAllowMergeCommitOk returns a tuple with the AllowMergeCommit field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAllowMergeCommit

`func (o *GithubRepository) SetAllowMergeCommit(v bool)`

SetAllowMergeCommit sets AllowMergeCommit field to given value.

### HasAllowMergeCommit

`func (o *GithubRepository) HasAllowMergeCommit() bool`

HasAllowMergeCommit returns a boolean if a field has been set.

### GetAllowRebaseMerge

`func (o *GithubRepository) GetAllowRebaseMerge() bool`

GetAllowRebaseMerge returns the AllowRebaseMerge field if non-nil, zero value otherwise.

### GetAllowRebaseMergeOk

`func (o *GithubRepository) GetAllowRebaseMergeOk() (*bool, bool)`

GetAllowRebaseMergeOk returns a tuple with the AllowRebaseMerge field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAllowRebaseMerge

`func (o *GithubRepository) SetAllowRebaseMerge(v bool)`

SetAllowRebaseMerge sets AllowRebaseMerge field to given value.

### HasAllowRebaseMerge

`func (o *GithubRepository) HasAllowRebaseMerge() bool`

HasAllowRebaseMerge returns a boolean if a field has been set.

### GetAllowSquashMerge

`func (o *GithubRepository) GetAllowSquashMerge() bool`

GetAllowSquashMerge returns the AllowSquashMerge field if non-nil, zero value otherwise.

### GetAllowSquashMergeOk

`func (o *GithubRepository) GetAllowSquashMergeOk() (*bool, bool)`

GetAllowSquashMergeOk returns a tuple with the AllowSquashMerge field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAllowSquashMerge

`func (o *GithubRepository) SetAllowSquashMerge(v bool)`

SetAllowSquashMerge sets AllowSquashMerge field to given value.

### HasAllowSquashMerge

`func (o *GithubRepository) HasAllowSquashMerge() bool`

HasAllowSquashMerge returns a boolean if a field has been set.

### GetArchiveUrl

`func (o *GithubRepository) GetArchiveUrl() string`

GetArchiveUrl returns the ArchiveUrl field if non-nil, zero value otherwise.

### GetArchiveUrlOk

`func (o *GithubRepository) GetArchiveUrlOk() (*string, bool)`

GetArchiveUrlOk returns a tuple with the ArchiveUrl field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetArchiveUrl

`func (o *GithubRepository) SetArchiveUrl(v string)`

SetArchiveUrl sets ArchiveUrl field to given value.

### HasArchiveUrl

`func (o *GithubRepository) HasArchiveUrl() bool`

HasArchiveUrl returns a boolean if a field has been set.

### GetArchived

`func (o *GithubRepository) GetArchived() bool`

GetArchived returns the Archived field if non-nil, zero value otherwise.

### GetArchivedOk

`func (o *GithubRepository) GetArchivedOk() (*bool, bool)`

GetArchivedOk returns a tuple with the Archived field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetArchived

`func (o *GithubRepository) SetArchived(v bool)`

SetArchived sets Archived field to given value.

### HasArchived

`func (o *GithubRepository) HasArchived() bool`

HasArchived returns a boolean if a field has been set.

### GetAssigneesUrl

`func (o *GithubRepository) GetAssigneesUrl() string`

GetAssigneesUrl returns the AssigneesUrl field if non-nil, zero value otherwise.

### GetAssigneesUrlOk

`func (o *GithubRepository) GetAssigneesUrlOk() (*string, bool)`

GetAssigneesUrlOk returns a tuple with the AssigneesUrl field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAssigneesUrl

`func (o *GithubRepository) SetAssigneesUrl(v string)`

SetAssigneesUrl sets AssigneesUrl field to given value.

### HasAssigneesUrl

`func (o *GithubRepository) HasAssigneesUrl() bool`

HasAssigneesUrl returns a boolean if a field has been set.

### GetAutoInit

`func (o *GithubRepository) GetAutoInit() bool`

GetAutoInit returns the AutoInit field if non-nil, zero value otherwise.

### GetAutoInitOk

`func (o *GithubRepository) GetAutoInitOk() (*bool, bool)`

GetAutoInitOk returns a tuple with the AutoInit field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAutoInit

`func (o *GithubRepository) SetAutoInit(v bool)`

SetAutoInit sets AutoInit field to given value.

### HasAutoInit

`func (o *GithubRepository) HasAutoInit() bool`

HasAutoInit returns a boolean if a field has been set.

### GetBlobsUrl

`func (o *GithubRepository) GetBlobsUrl() string`

GetBlobsUrl returns the BlobsUrl field if non-nil, zero value otherwise.

### GetBlobsUrlOk

`func (o *GithubRepository) GetBlobsUrlOk() (*string, bool)`

GetBlobsUrlOk returns a tuple with the BlobsUrl field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetBlobsUrl

`func (o *GithubRepository) SetBlobsUrl(v string)`

SetBlobsUrl sets BlobsUrl field to given value.

### HasBlobsUrl

`func (o *GithubRepository) HasBlobsUrl() bool`

HasBlobsUrl returns a boolean if a field has been set.

### GetBranchesUrl

`func (o *GithubRepository) GetBranchesUrl() string`

GetBranchesUrl returns the BranchesUrl field if non-nil, zero value otherwise.

### GetBranchesUrlOk

`func (o *GithubRepository) GetBranchesUrlOk() (*string, bool)`

GetBranchesUrlOk returns a tuple with the BranchesUrl field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetBranchesUrl

`func (o *GithubRepository) SetBranchesUrl(v string)`

SetBranchesUrl sets BranchesUrl field to given value.

### HasBranchesUrl

`func (o *GithubRepository) HasBranchesUrl() bool`

HasBranchesUrl returns a boolean if a field has been set.

### GetCloneUrl

`func (o *GithubRepository) GetCloneUrl() string`

GetCloneUrl returns the CloneUrl field if non-nil, zero value otherwise.

### GetCloneUrlOk

`func (o *GithubRepository) GetCloneUrlOk() (*string, bool)`

GetCloneUrlOk returns a tuple with the CloneUrl field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCloneUrl

`func (o *GithubRepository) SetCloneUrl(v string)`

SetCloneUrl sets CloneUrl field to given value.

### HasCloneUrl

`func (o *GithubRepository) HasCloneUrl() bool`

HasCloneUrl returns a boolean if a field has been set.

### GetCodeOfConduct

`func (o *GithubRepository) GetCodeOfConduct() GithubCodeOfConduct`

GetCodeOfConduct returns the CodeOfConduct field if non-nil, zero value otherwise.

### GetCodeOfConductOk

`func (o *GithubRepository) GetCodeOfConductOk() (*GithubCodeOfConduct, bool)`

GetCodeOfConductOk returns a tuple with the CodeOfConduct field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCodeOfConduct

`func (o *GithubRepository) SetCodeOfConduct(v GithubCodeOfConduct)`

SetCodeOfConduct sets CodeOfConduct field to given value.

### HasCodeOfConduct

`func (o *GithubRepository) HasCodeOfConduct() bool`

HasCodeOfConduct returns a boolean if a field has been set.

### GetCollaboratorsUrl

`func (o *GithubRepository) GetCollaboratorsUrl() string`

GetCollaboratorsUrl returns the CollaboratorsUrl field if non-nil, zero value otherwise.

### GetCollaboratorsUrlOk

`func (o *GithubRepository) GetCollaboratorsUrlOk() (*string, bool)`

GetCollaboratorsUrlOk returns a tuple with the CollaboratorsUrl field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCollaboratorsUrl

`func (o *GithubRepository) SetCollaboratorsUrl(v string)`

SetCollaboratorsUrl sets CollaboratorsUrl field to given value.

### HasCollaboratorsUrl

`func (o *GithubRepository) HasCollaboratorsUrl() bool`

HasCollaboratorsUrl returns a boolean if a field has been set.

### GetCommentsUrl

`func (o *GithubRepository) GetCommentsUrl() string`

GetCommentsUrl returns the CommentsUrl field if non-nil, zero value otherwise.

### GetCommentsUrlOk

`func (o *GithubRepository) GetCommentsUrlOk() (*string, bool)`

GetCommentsUrlOk returns a tuple with the CommentsUrl field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCommentsUrl

`func (o *GithubRepository) SetCommentsUrl(v string)`

SetCommentsUrl sets CommentsUrl field to given value.

### HasCommentsUrl

`func (o *GithubRepository) HasCommentsUrl() bool`

HasCommentsUrl returns a boolean if a field has been set.

### GetCommitsUrl

`func (o *GithubRepository) GetCommitsUrl() string`

GetCommitsUrl returns the CommitsUrl field if non-nil, zero value otherwise.

### GetCommitsUrlOk

`func (o *GithubRepository) GetCommitsUrlOk() (*string, bool)`

GetCommitsUrlOk returns a tuple with the CommitsUrl field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCommitsUrl

`func (o *GithubRepository) SetCommitsUrl(v string)`

SetCommitsUrl sets CommitsUrl field to given value.

### HasCommitsUrl

`func (o *GithubRepository) HasCommitsUrl() bool`

HasCommitsUrl returns a boolean if a field has been set.

### GetCompareUrl

`func (o *GithubRepository) GetCompareUrl() string`

GetCompareUrl returns the CompareUrl field if non-nil, zero value otherwise.

### GetCompareUrlOk

`func (o *GithubRepository) GetCompareUrlOk() (*string, bool)`

GetCompareUrlOk returns a tuple with the CompareUrl field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCompareUrl

`func (o *GithubRepository) SetCompareUrl(v string)`

SetCompareUrl sets CompareUrl field to given value.

### HasCompareUrl

`func (o *GithubRepository) HasCompareUrl() bool`

HasCompareUrl returns a boolean if a field has been set.

### GetContentsUrl

`func (o *GithubRepository) GetContentsUrl() string`

GetContentsUrl returns the ContentsUrl field if non-nil, zero value otherwise.

### GetContentsUrlOk

`func (o *GithubRepository) GetContentsUrlOk() (*string, bool)`

GetContentsUrlOk returns a tuple with the ContentsUrl field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetContentsUrl

`func (o *GithubRepository) SetContentsUrl(v string)`

SetContentsUrl sets ContentsUrl field to given value.

### HasContentsUrl

`func (o *GithubRepository) HasContentsUrl() bool`

HasContentsUrl returns a boolean if a field has been set.

### GetContributorsUrl

`func (o *GithubRepository) GetContributorsUrl() string`

GetContributorsUrl returns the ContributorsUrl field if non-nil, zero value otherwise.

### GetContributorsUrlOk

`func (o *GithubRepository) GetContributorsUrlOk() (*string, bool)`

GetContributorsUrlOk returns a tuple with the ContributorsUrl field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetContributorsUrl

`func (o *GithubRepository) SetContributorsUrl(v string)`

SetContributorsUrl sets ContributorsUrl field to given value.

### HasContributorsUrl

`func (o *GithubRepository) HasContributorsUrl() bool`

HasContributorsUrl returns a boolean if a field has been set.

### GetCreatedAt

`func (o *GithubRepository) GetCreatedAt() GithubTimestamp`

GetCreatedAt returns the CreatedAt field if non-nil, zero value otherwise.

### GetCreatedAtOk

`func (o *GithubRepository) GetCreatedAtOk() (*GithubTimestamp, bool)`

GetCreatedAtOk returns a tuple with the CreatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedAt

`func (o *GithubRepository) SetCreatedAt(v GithubTimestamp)`

SetCreatedAt sets CreatedAt field to given value.

### HasCreatedAt

`func (o *GithubRepository) HasCreatedAt() bool`

HasCreatedAt returns a boolean if a field has been set.

### GetDefaultBranch

`func (o *GithubRepository) GetDefaultBranch() string`

GetDefaultBranch returns the DefaultBranch field if non-nil, zero value otherwise.

### GetDefaultBranchOk

`func (o *GithubRepository) GetDefaultBranchOk() (*string, bool)`

GetDefaultBranchOk returns a tuple with the DefaultBranch field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDefaultBranch

`func (o *GithubRepository) SetDefaultBranch(v string)`

SetDefaultBranch sets DefaultBranch field to given value.

### HasDefaultBranch

`func (o *GithubRepository) HasDefaultBranch() bool`

HasDefaultBranch returns a boolean if a field has been set.

### GetDeploymentsUrl

`func (o *GithubRepository) GetDeploymentsUrl() string`

GetDeploymentsUrl returns the DeploymentsUrl field if non-nil, zero value otherwise.

### GetDeploymentsUrlOk

`func (o *GithubRepository) GetDeploymentsUrlOk() (*string, bool)`

GetDeploymentsUrlOk returns a tuple with the DeploymentsUrl field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDeploymentsUrl

`func (o *GithubRepository) SetDeploymentsUrl(v string)`

SetDeploymentsUrl sets DeploymentsUrl field to given value.

### HasDeploymentsUrl

`func (o *GithubRepository) HasDeploymentsUrl() bool`

HasDeploymentsUrl returns a boolean if a field has been set.

### GetDescription

`func (o *GithubRepository) GetDescription() string`

GetDescription returns the Description field if non-nil, zero value otherwise.

### GetDescriptionOk

`func (o *GithubRepository) GetDescriptionOk() (*string, bool)`

GetDescriptionOk returns a tuple with the Description field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDescription

`func (o *GithubRepository) SetDescription(v string)`

SetDescription sets Description field to given value.

### HasDescription

`func (o *GithubRepository) HasDescription() bool`

HasDescription returns a boolean if a field has been set.

### GetDownloadsUrl

`func (o *GithubRepository) GetDownloadsUrl() string`

GetDownloadsUrl returns the DownloadsUrl field if non-nil, zero value otherwise.

### GetDownloadsUrlOk

`func (o *GithubRepository) GetDownloadsUrlOk() (*string, bool)`

GetDownloadsUrlOk returns a tuple with the DownloadsUrl field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDownloadsUrl

`func (o *GithubRepository) SetDownloadsUrl(v string)`

SetDownloadsUrl sets DownloadsUrl field to given value.

### HasDownloadsUrl

`func (o *GithubRepository) HasDownloadsUrl() bool`

HasDownloadsUrl returns a boolean if a field has been set.

### GetEventsUrl

`func (o *GithubRepository) GetEventsUrl() string`

GetEventsUrl returns the EventsUrl field if non-nil, zero value otherwise.

### GetEventsUrlOk

`func (o *GithubRepository) GetEventsUrlOk() (*string, bool)`

GetEventsUrlOk returns a tuple with the EventsUrl field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEventsUrl

`func (o *GithubRepository) SetEventsUrl(v string)`

SetEventsUrl sets EventsUrl field to given value.

### HasEventsUrl

`func (o *GithubRepository) HasEventsUrl() bool`

HasEventsUrl returns a boolean if a field has been set.

### GetFork

`func (o *GithubRepository) GetFork() bool`

GetFork returns the Fork field if non-nil, zero value otherwise.

### GetForkOk

`func (o *GithubRepository) GetForkOk() (*bool, bool)`

GetForkOk returns a tuple with the Fork field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFork

`func (o *GithubRepository) SetFork(v bool)`

SetFork sets Fork field to given value.

### HasFork

`func (o *GithubRepository) HasFork() bool`

HasFork returns a boolean if a field has been set.

### GetForksCount

`func (o *GithubRepository) GetForksCount() int32`

GetForksCount returns the ForksCount field if non-nil, zero value otherwise.

### GetForksCountOk

`func (o *GithubRepository) GetForksCountOk() (*int32, bool)`

GetForksCountOk returns a tuple with the ForksCount field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetForksCount

`func (o *GithubRepository) SetForksCount(v int32)`

SetForksCount sets ForksCount field to given value.

### HasForksCount

`func (o *GithubRepository) HasForksCount() bool`

HasForksCount returns a boolean if a field has been set.

### GetForksUrl

`func (o *GithubRepository) GetForksUrl() string`

GetForksUrl returns the ForksUrl field if non-nil, zero value otherwise.

### GetForksUrlOk

`func (o *GithubRepository) GetForksUrlOk() (*string, bool)`

GetForksUrlOk returns a tuple with the ForksUrl field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetForksUrl

`func (o *GithubRepository) SetForksUrl(v string)`

SetForksUrl sets ForksUrl field to given value.

### HasForksUrl

`func (o *GithubRepository) HasForksUrl() bool`

HasForksUrl returns a boolean if a field has been set.

### GetFullName

`func (o *GithubRepository) GetFullName() string`

GetFullName returns the FullName field if non-nil, zero value otherwise.

### GetFullNameOk

`func (o *GithubRepository) GetFullNameOk() (*string, bool)`

GetFullNameOk returns a tuple with the FullName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFullName

`func (o *GithubRepository) SetFullName(v string)`

SetFullName sets FullName field to given value.

### HasFullName

`func (o *GithubRepository) HasFullName() bool`

HasFullName returns a boolean if a field has been set.

### GetGitCommitsUrl

`func (o *GithubRepository) GetGitCommitsUrl() string`

GetGitCommitsUrl returns the GitCommitsUrl field if non-nil, zero value otherwise.

### GetGitCommitsUrlOk

`func (o *GithubRepository) GetGitCommitsUrlOk() (*string, bool)`

GetGitCommitsUrlOk returns a tuple with the GitCommitsUrl field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGitCommitsUrl

`func (o *GithubRepository) SetGitCommitsUrl(v string)`

SetGitCommitsUrl sets GitCommitsUrl field to given value.

### HasGitCommitsUrl

`func (o *GithubRepository) HasGitCommitsUrl() bool`

HasGitCommitsUrl returns a boolean if a field has been set.

### GetGitRefsUrl

`func (o *GithubRepository) GetGitRefsUrl() string`

GetGitRefsUrl returns the GitRefsUrl field if non-nil, zero value otherwise.

### GetGitRefsUrlOk

`func (o *GithubRepository) GetGitRefsUrlOk() (*string, bool)`

GetGitRefsUrlOk returns a tuple with the GitRefsUrl field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGitRefsUrl

`func (o *GithubRepository) SetGitRefsUrl(v string)`

SetGitRefsUrl sets GitRefsUrl field to given value.

### HasGitRefsUrl

`func (o *GithubRepository) HasGitRefsUrl() bool`

HasGitRefsUrl returns a boolean if a field has been set.

### GetGitTagsUrl

`func (o *GithubRepository) GetGitTagsUrl() string`

GetGitTagsUrl returns the GitTagsUrl field if non-nil, zero value otherwise.

### GetGitTagsUrlOk

`func (o *GithubRepository) GetGitTagsUrlOk() (*string, bool)`

GetGitTagsUrlOk returns a tuple with the GitTagsUrl field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGitTagsUrl

`func (o *GithubRepository) SetGitTagsUrl(v string)`

SetGitTagsUrl sets GitTagsUrl field to given value.

### HasGitTagsUrl

`func (o *GithubRepository) HasGitTagsUrl() bool`

HasGitTagsUrl returns a boolean if a field has been set.

### GetGitUrl

`func (o *GithubRepository) GetGitUrl() string`

GetGitUrl returns the GitUrl field if non-nil, zero value otherwise.

### GetGitUrlOk

`func (o *GithubRepository) GetGitUrlOk() (*string, bool)`

GetGitUrlOk returns a tuple with the GitUrl field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGitUrl

`func (o *GithubRepository) SetGitUrl(v string)`

SetGitUrl sets GitUrl field to given value.

### HasGitUrl

`func (o *GithubRepository) HasGitUrl() bool`

HasGitUrl returns a boolean if a field has been set.

### GetGitignoreTemplate

`func (o *GithubRepository) GetGitignoreTemplate() string`

GetGitignoreTemplate returns the GitignoreTemplate field if non-nil, zero value otherwise.

### GetGitignoreTemplateOk

`func (o *GithubRepository) GetGitignoreTemplateOk() (*string, bool)`

GetGitignoreTemplateOk returns a tuple with the GitignoreTemplate field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGitignoreTemplate

`func (o *GithubRepository) SetGitignoreTemplate(v string)`

SetGitignoreTemplate sets GitignoreTemplate field to given value.

### HasGitignoreTemplate

`func (o *GithubRepository) HasGitignoreTemplate() bool`

HasGitignoreTemplate returns a boolean if a field has been set.

### GetHasDownloads

`func (o *GithubRepository) GetHasDownloads() bool`

GetHasDownloads returns the HasDownloads field if non-nil, zero value otherwise.

### GetHasDownloadsOk

`func (o *GithubRepository) GetHasDownloadsOk() (*bool, bool)`

GetHasDownloadsOk returns a tuple with the HasDownloads field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetHasDownloads

`func (o *GithubRepository) SetHasDownloads(v bool)`

SetHasDownloads sets HasDownloads field to given value.

### HasHasDownloads

`func (o *GithubRepository) HasHasDownloads() bool`

HasHasDownloads returns a boolean if a field has been set.

### GetHasIssues

`func (o *GithubRepository) GetHasIssues() bool`

GetHasIssues returns the HasIssues field if non-nil, zero value otherwise.

### GetHasIssuesOk

`func (o *GithubRepository) GetHasIssuesOk() (*bool, bool)`

GetHasIssuesOk returns a tuple with the HasIssues field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetHasIssues

`func (o *GithubRepository) SetHasIssues(v bool)`

SetHasIssues sets HasIssues field to given value.

### HasHasIssues

`func (o *GithubRepository) HasHasIssues() bool`

HasHasIssues returns a boolean if a field has been set.

### GetHasPages

`func (o *GithubRepository) GetHasPages() bool`

GetHasPages returns the HasPages field if non-nil, zero value otherwise.

### GetHasPagesOk

`func (o *GithubRepository) GetHasPagesOk() (*bool, bool)`

GetHasPagesOk returns a tuple with the HasPages field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetHasPages

`func (o *GithubRepository) SetHasPages(v bool)`

SetHasPages sets HasPages field to given value.

### HasHasPages

`func (o *GithubRepository) HasHasPages() bool`

HasHasPages returns a boolean if a field has been set.

### GetHasProjects

`func (o *GithubRepository) GetHasProjects() bool`

GetHasProjects returns the HasProjects field if non-nil, zero value otherwise.

### GetHasProjectsOk

`func (o *GithubRepository) GetHasProjectsOk() (*bool, bool)`

GetHasProjectsOk returns a tuple with the HasProjects field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetHasProjects

`func (o *GithubRepository) SetHasProjects(v bool)`

SetHasProjects sets HasProjects field to given value.

### HasHasProjects

`func (o *GithubRepository) HasHasProjects() bool`

HasHasProjects returns a boolean if a field has been set.

### GetHasWiki

`func (o *GithubRepository) GetHasWiki() bool`

GetHasWiki returns the HasWiki field if non-nil, zero value otherwise.

### GetHasWikiOk

`func (o *GithubRepository) GetHasWikiOk() (*bool, bool)`

GetHasWikiOk returns a tuple with the HasWiki field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetHasWiki

`func (o *GithubRepository) SetHasWiki(v bool)`

SetHasWiki sets HasWiki field to given value.

### HasHasWiki

`func (o *GithubRepository) HasHasWiki() bool`

HasHasWiki returns a boolean if a field has been set.

### GetHomepage

`func (o *GithubRepository) GetHomepage() string`

GetHomepage returns the Homepage field if non-nil, zero value otherwise.

### GetHomepageOk

`func (o *GithubRepository) GetHomepageOk() (*string, bool)`

GetHomepageOk returns a tuple with the Homepage field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetHomepage

`func (o *GithubRepository) SetHomepage(v string)`

SetHomepage sets Homepage field to given value.

### HasHomepage

`func (o *GithubRepository) HasHomepage() bool`

HasHomepage returns a boolean if a field has been set.

### GetHooksUrl

`func (o *GithubRepository) GetHooksUrl() string`

GetHooksUrl returns the HooksUrl field if non-nil, zero value otherwise.

### GetHooksUrlOk

`func (o *GithubRepository) GetHooksUrlOk() (*string, bool)`

GetHooksUrlOk returns a tuple with the HooksUrl field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetHooksUrl

`func (o *GithubRepository) SetHooksUrl(v string)`

SetHooksUrl sets HooksUrl field to given value.

### HasHooksUrl

`func (o *GithubRepository) HasHooksUrl() bool`

HasHooksUrl returns a boolean if a field has been set.

### GetHtmlUrl

`func (o *GithubRepository) GetHtmlUrl() string`

GetHtmlUrl returns the HtmlUrl field if non-nil, zero value otherwise.

### GetHtmlUrlOk

`func (o *GithubRepository) GetHtmlUrlOk() (*string, bool)`

GetHtmlUrlOk returns a tuple with the HtmlUrl field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetHtmlUrl

`func (o *GithubRepository) SetHtmlUrl(v string)`

SetHtmlUrl sets HtmlUrl field to given value.

### HasHtmlUrl

`func (o *GithubRepository) HasHtmlUrl() bool`

HasHtmlUrl returns a boolean if a field has been set.

### GetId

`func (o *GithubRepository) GetId() int32`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *GithubRepository) GetIdOk() (*int32, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *GithubRepository) SetId(v int32)`

SetId sets Id field to given value.

### HasId

`func (o *GithubRepository) HasId() bool`

HasId returns a boolean if a field has been set.

### GetIssueCommentUrl

`func (o *GithubRepository) GetIssueCommentUrl() string`

GetIssueCommentUrl returns the IssueCommentUrl field if non-nil, zero value otherwise.

### GetIssueCommentUrlOk

`func (o *GithubRepository) GetIssueCommentUrlOk() (*string, bool)`

GetIssueCommentUrlOk returns a tuple with the IssueCommentUrl field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIssueCommentUrl

`func (o *GithubRepository) SetIssueCommentUrl(v string)`

SetIssueCommentUrl sets IssueCommentUrl field to given value.

### HasIssueCommentUrl

`func (o *GithubRepository) HasIssueCommentUrl() bool`

HasIssueCommentUrl returns a boolean if a field has been set.

### GetIssueEventsUrl

`func (o *GithubRepository) GetIssueEventsUrl() string`

GetIssueEventsUrl returns the IssueEventsUrl field if non-nil, zero value otherwise.

### GetIssueEventsUrlOk

`func (o *GithubRepository) GetIssueEventsUrlOk() (*string, bool)`

GetIssueEventsUrlOk returns a tuple with the IssueEventsUrl field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIssueEventsUrl

`func (o *GithubRepository) SetIssueEventsUrl(v string)`

SetIssueEventsUrl sets IssueEventsUrl field to given value.

### HasIssueEventsUrl

`func (o *GithubRepository) HasIssueEventsUrl() bool`

HasIssueEventsUrl returns a boolean if a field has been set.

### GetIssuesUrl

`func (o *GithubRepository) GetIssuesUrl() string`

GetIssuesUrl returns the IssuesUrl field if non-nil, zero value otherwise.

### GetIssuesUrlOk

`func (o *GithubRepository) GetIssuesUrlOk() (*string, bool)`

GetIssuesUrlOk returns a tuple with the IssuesUrl field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIssuesUrl

`func (o *GithubRepository) SetIssuesUrl(v string)`

SetIssuesUrl sets IssuesUrl field to given value.

### HasIssuesUrl

`func (o *GithubRepository) HasIssuesUrl() bool`

HasIssuesUrl returns a boolean if a field has been set.

### GetKeysUrl

`func (o *GithubRepository) GetKeysUrl() string`

GetKeysUrl returns the KeysUrl field if non-nil, zero value otherwise.

### GetKeysUrlOk

`func (o *GithubRepository) GetKeysUrlOk() (*string, bool)`

GetKeysUrlOk returns a tuple with the KeysUrl field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetKeysUrl

`func (o *GithubRepository) SetKeysUrl(v string)`

SetKeysUrl sets KeysUrl field to given value.

### HasKeysUrl

`func (o *GithubRepository) HasKeysUrl() bool`

HasKeysUrl returns a boolean if a field has been set.

### GetLabelsUrl

`func (o *GithubRepository) GetLabelsUrl() string`

GetLabelsUrl returns the LabelsUrl field if non-nil, zero value otherwise.

### GetLabelsUrlOk

`func (o *GithubRepository) GetLabelsUrlOk() (*string, bool)`

GetLabelsUrlOk returns a tuple with the LabelsUrl field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLabelsUrl

`func (o *GithubRepository) SetLabelsUrl(v string)`

SetLabelsUrl sets LabelsUrl field to given value.

### HasLabelsUrl

`func (o *GithubRepository) HasLabelsUrl() bool`

HasLabelsUrl returns a boolean if a field has been set.

### GetLanguage

`func (o *GithubRepository) GetLanguage() string`

GetLanguage returns the Language field if non-nil, zero value otherwise.

### GetLanguageOk

`func (o *GithubRepository) GetLanguageOk() (*string, bool)`

GetLanguageOk returns a tuple with the Language field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLanguage

`func (o *GithubRepository) SetLanguage(v string)`

SetLanguage sets Language field to given value.

### HasLanguage

`func (o *GithubRepository) HasLanguage() bool`

HasLanguage returns a boolean if a field has been set.

### GetLanguagesUrl

`func (o *GithubRepository) GetLanguagesUrl() string`

GetLanguagesUrl returns the LanguagesUrl field if non-nil, zero value otherwise.

### GetLanguagesUrlOk

`func (o *GithubRepository) GetLanguagesUrlOk() (*string, bool)`

GetLanguagesUrlOk returns a tuple with the LanguagesUrl field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLanguagesUrl

`func (o *GithubRepository) SetLanguagesUrl(v string)`

SetLanguagesUrl sets LanguagesUrl field to given value.

### HasLanguagesUrl

`func (o *GithubRepository) HasLanguagesUrl() bool`

HasLanguagesUrl returns a boolean if a field has been set.

### GetLicense

`func (o *GithubRepository) GetLicense() GithubLicense`

GetLicense returns the License field if non-nil, zero value otherwise.

### GetLicenseOk

`func (o *GithubRepository) GetLicenseOk() (*GithubLicense, bool)`

GetLicenseOk returns a tuple with the License field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLicense

`func (o *GithubRepository) SetLicense(v GithubLicense)`

SetLicense sets License field to given value.

### HasLicense

`func (o *GithubRepository) HasLicense() bool`

HasLicense returns a boolean if a field has been set.

### GetLicenseTemplate

`func (o *GithubRepository) GetLicenseTemplate() string`

GetLicenseTemplate returns the LicenseTemplate field if non-nil, zero value otherwise.

### GetLicenseTemplateOk

`func (o *GithubRepository) GetLicenseTemplateOk() (*string, bool)`

GetLicenseTemplateOk returns a tuple with the LicenseTemplate field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLicenseTemplate

`func (o *GithubRepository) SetLicenseTemplate(v string)`

SetLicenseTemplate sets LicenseTemplate field to given value.

### HasLicenseTemplate

`func (o *GithubRepository) HasLicenseTemplate() bool`

HasLicenseTemplate returns a boolean if a field has been set.

### GetMasterBranch

`func (o *GithubRepository) GetMasterBranch() string`

GetMasterBranch returns the MasterBranch field if non-nil, zero value otherwise.

### GetMasterBranchOk

`func (o *GithubRepository) GetMasterBranchOk() (*string, bool)`

GetMasterBranchOk returns a tuple with the MasterBranch field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMasterBranch

`func (o *GithubRepository) SetMasterBranch(v string)`

SetMasterBranch sets MasterBranch field to given value.

### HasMasterBranch

`func (o *GithubRepository) HasMasterBranch() bool`

HasMasterBranch returns a boolean if a field has been set.

### GetMergesUrl

`func (o *GithubRepository) GetMergesUrl() string`

GetMergesUrl returns the MergesUrl field if non-nil, zero value otherwise.

### GetMergesUrlOk

`func (o *GithubRepository) GetMergesUrlOk() (*string, bool)`

GetMergesUrlOk returns a tuple with the MergesUrl field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMergesUrl

`func (o *GithubRepository) SetMergesUrl(v string)`

SetMergesUrl sets MergesUrl field to given value.

### HasMergesUrl

`func (o *GithubRepository) HasMergesUrl() bool`

HasMergesUrl returns a boolean if a field has been set.

### GetMilestonesUrl

`func (o *GithubRepository) GetMilestonesUrl() string`

GetMilestonesUrl returns the MilestonesUrl field if non-nil, zero value otherwise.

### GetMilestonesUrlOk

`func (o *GithubRepository) GetMilestonesUrlOk() (*string, bool)`

GetMilestonesUrlOk returns a tuple with the MilestonesUrl field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMilestonesUrl

`func (o *GithubRepository) SetMilestonesUrl(v string)`

SetMilestonesUrl sets MilestonesUrl field to given value.

### HasMilestonesUrl

`func (o *GithubRepository) HasMilestonesUrl() bool`

HasMilestonesUrl returns a boolean if a field has been set.

### GetMirrorUrl

`func (o *GithubRepository) GetMirrorUrl() string`

GetMirrorUrl returns the MirrorUrl field if non-nil, zero value otherwise.

### GetMirrorUrlOk

`func (o *GithubRepository) GetMirrorUrlOk() (*string, bool)`

GetMirrorUrlOk returns a tuple with the MirrorUrl field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMirrorUrl

`func (o *GithubRepository) SetMirrorUrl(v string)`

SetMirrorUrl sets MirrorUrl field to given value.

### HasMirrorUrl

`func (o *GithubRepository) HasMirrorUrl() bool`

HasMirrorUrl returns a boolean if a field has been set.

### GetName

`func (o *GithubRepository) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *GithubRepository) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *GithubRepository) SetName(v string)`

SetName sets Name field to given value.

### HasName

`func (o *GithubRepository) HasName() bool`

HasName returns a boolean if a field has been set.

### GetNetworkCount

`func (o *GithubRepository) GetNetworkCount() int32`

GetNetworkCount returns the NetworkCount field if non-nil, zero value otherwise.

### GetNetworkCountOk

`func (o *GithubRepository) GetNetworkCountOk() (*int32, bool)`

GetNetworkCountOk returns a tuple with the NetworkCount field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNetworkCount

`func (o *GithubRepository) SetNetworkCount(v int32)`

SetNetworkCount sets NetworkCount field to given value.

### HasNetworkCount

`func (o *GithubRepository) HasNetworkCount() bool`

HasNetworkCount returns a boolean if a field has been set.

### GetNodeId

`func (o *GithubRepository) GetNodeId() string`

GetNodeId returns the NodeId field if non-nil, zero value otherwise.

### GetNodeIdOk

`func (o *GithubRepository) GetNodeIdOk() (*string, bool)`

GetNodeIdOk returns a tuple with the NodeId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodeId

`func (o *GithubRepository) SetNodeId(v string)`

SetNodeId sets NodeId field to given value.

### HasNodeId

`func (o *GithubRepository) HasNodeId() bool`

HasNodeId returns a boolean if a field has been set.

### GetNotificationsUrl

`func (o *GithubRepository) GetNotificationsUrl() string`

GetNotificationsUrl returns the NotificationsUrl field if non-nil, zero value otherwise.

### GetNotificationsUrlOk

`func (o *GithubRepository) GetNotificationsUrlOk() (*string, bool)`

GetNotificationsUrlOk returns a tuple with the NotificationsUrl field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNotificationsUrl

`func (o *GithubRepository) SetNotificationsUrl(v string)`

SetNotificationsUrl sets NotificationsUrl field to given value.

### HasNotificationsUrl

`func (o *GithubRepository) HasNotificationsUrl() bool`

HasNotificationsUrl returns a boolean if a field has been set.

### GetOpenIssuesCount

`func (o *GithubRepository) GetOpenIssuesCount() int32`

GetOpenIssuesCount returns the OpenIssuesCount field if non-nil, zero value otherwise.

### GetOpenIssuesCountOk

`func (o *GithubRepository) GetOpenIssuesCountOk() (*int32, bool)`

GetOpenIssuesCountOk returns a tuple with the OpenIssuesCount field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOpenIssuesCount

`func (o *GithubRepository) SetOpenIssuesCount(v int32)`

SetOpenIssuesCount sets OpenIssuesCount field to given value.

### HasOpenIssuesCount

`func (o *GithubRepository) HasOpenIssuesCount() bool`

HasOpenIssuesCount returns a boolean if a field has been set.

### GetOrganization

`func (o *GithubRepository) GetOrganization() GithubOrganization`

GetOrganization returns the Organization field if non-nil, zero value otherwise.

### GetOrganizationOk

`func (o *GithubRepository) GetOrganizationOk() (*GithubOrganization, bool)`

GetOrganizationOk returns a tuple with the Organization field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOrganization

`func (o *GithubRepository) SetOrganization(v GithubOrganization)`

SetOrganization sets Organization field to given value.

### HasOrganization

`func (o *GithubRepository) HasOrganization() bool`

HasOrganization returns a boolean if a field has been set.

### GetOwner

`func (o *GithubRepository) GetOwner() GithubUser`

GetOwner returns the Owner field if non-nil, zero value otherwise.

### GetOwnerOk

`func (o *GithubRepository) GetOwnerOk() (*GithubUser, bool)`

GetOwnerOk returns a tuple with the Owner field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOwner

`func (o *GithubRepository) SetOwner(v GithubUser)`

SetOwner sets Owner field to given value.

### HasOwner

`func (o *GithubRepository) HasOwner() bool`

HasOwner returns a boolean if a field has been set.

### GetParent

`func (o *GithubRepository) GetParent() GithubRepository`

GetParent returns the Parent field if non-nil, zero value otherwise.

### GetParentOk

`func (o *GithubRepository) GetParentOk() (*GithubRepository, bool)`

GetParentOk returns a tuple with the Parent field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetParent

`func (o *GithubRepository) SetParent(v GithubRepository)`

SetParent sets Parent field to given value.

### HasParent

`func (o *GithubRepository) HasParent() bool`

HasParent returns a boolean if a field has been set.

### GetPermissions

`func (o *GithubRepository) GetPermissions() map[string]bool`

GetPermissions returns the Permissions field if non-nil, zero value otherwise.

### GetPermissionsOk

`func (o *GithubRepository) GetPermissionsOk() (*map[string]bool, bool)`

GetPermissionsOk returns a tuple with the Permissions field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPermissions

`func (o *GithubRepository) SetPermissions(v map[string]bool)`

SetPermissions sets Permissions field to given value.

### HasPermissions

`func (o *GithubRepository) HasPermissions() bool`

HasPermissions returns a boolean if a field has been set.

### GetPrivate

`func (o *GithubRepository) GetPrivate() bool`

GetPrivate returns the Private field if non-nil, zero value otherwise.

### GetPrivateOk

`func (o *GithubRepository) GetPrivateOk() (*bool, bool)`

GetPrivateOk returns a tuple with the Private field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPrivate

`func (o *GithubRepository) SetPrivate(v bool)`

SetPrivate sets Private field to given value.

### HasPrivate

`func (o *GithubRepository) HasPrivate() bool`

HasPrivate returns a boolean if a field has been set.

### GetPullsUrl

`func (o *GithubRepository) GetPullsUrl() string`

GetPullsUrl returns the PullsUrl field if non-nil, zero value otherwise.

### GetPullsUrlOk

`func (o *GithubRepository) GetPullsUrlOk() (*string, bool)`

GetPullsUrlOk returns a tuple with the PullsUrl field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPullsUrl

`func (o *GithubRepository) SetPullsUrl(v string)`

SetPullsUrl sets PullsUrl field to given value.

### HasPullsUrl

`func (o *GithubRepository) HasPullsUrl() bool`

HasPullsUrl returns a boolean if a field has been set.

### GetPushedAt

`func (o *GithubRepository) GetPushedAt() GithubTimestamp`

GetPushedAt returns the PushedAt field if non-nil, zero value otherwise.

### GetPushedAtOk

`func (o *GithubRepository) GetPushedAtOk() (*GithubTimestamp, bool)`

GetPushedAtOk returns a tuple with the PushedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPushedAt

`func (o *GithubRepository) SetPushedAt(v GithubTimestamp)`

SetPushedAt sets PushedAt field to given value.

### HasPushedAt

`func (o *GithubRepository) HasPushedAt() bool`

HasPushedAt returns a boolean if a field has been set.

### GetReleasesUrl

`func (o *GithubRepository) GetReleasesUrl() string`

GetReleasesUrl returns the ReleasesUrl field if non-nil, zero value otherwise.

### GetReleasesUrlOk

`func (o *GithubRepository) GetReleasesUrlOk() (*string, bool)`

GetReleasesUrlOk returns a tuple with the ReleasesUrl field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetReleasesUrl

`func (o *GithubRepository) SetReleasesUrl(v string)`

SetReleasesUrl sets ReleasesUrl field to given value.

### HasReleasesUrl

`func (o *GithubRepository) HasReleasesUrl() bool`

HasReleasesUrl returns a boolean if a field has been set.

### GetSize

`func (o *GithubRepository) GetSize() int32`

GetSize returns the Size field if non-nil, zero value otherwise.

### GetSizeOk

`func (o *GithubRepository) GetSizeOk() (*int32, bool)`

GetSizeOk returns a tuple with the Size field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSize

`func (o *GithubRepository) SetSize(v int32)`

SetSize sets Size field to given value.

### HasSize

`func (o *GithubRepository) HasSize() bool`

HasSize returns a boolean if a field has been set.

### GetSource

`func (o *GithubRepository) GetSource() GithubRepository`

GetSource returns the Source field if non-nil, zero value otherwise.

### GetSourceOk

`func (o *GithubRepository) GetSourceOk() (*GithubRepository, bool)`

GetSourceOk returns a tuple with the Source field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSource

`func (o *GithubRepository) SetSource(v GithubRepository)`

SetSource sets Source field to given value.

### HasSource

`func (o *GithubRepository) HasSource() bool`

HasSource returns a boolean if a field has been set.

### GetSshUrl

`func (o *GithubRepository) GetSshUrl() string`

GetSshUrl returns the SshUrl field if non-nil, zero value otherwise.

### GetSshUrlOk

`func (o *GithubRepository) GetSshUrlOk() (*string, bool)`

GetSshUrlOk returns a tuple with the SshUrl field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSshUrl

`func (o *GithubRepository) SetSshUrl(v string)`

SetSshUrl sets SshUrl field to given value.

### HasSshUrl

`func (o *GithubRepository) HasSshUrl() bool`

HasSshUrl returns a boolean if a field has been set.

### GetStargazersCount

`func (o *GithubRepository) GetStargazersCount() int32`

GetStargazersCount returns the StargazersCount field if non-nil, zero value otherwise.

### GetStargazersCountOk

`func (o *GithubRepository) GetStargazersCountOk() (*int32, bool)`

GetStargazersCountOk returns a tuple with the StargazersCount field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStargazersCount

`func (o *GithubRepository) SetStargazersCount(v int32)`

SetStargazersCount sets StargazersCount field to given value.

### HasStargazersCount

`func (o *GithubRepository) HasStargazersCount() bool`

HasStargazersCount returns a boolean if a field has been set.

### GetStargazersUrl

`func (o *GithubRepository) GetStargazersUrl() string`

GetStargazersUrl returns the StargazersUrl field if non-nil, zero value otherwise.

### GetStargazersUrlOk

`func (o *GithubRepository) GetStargazersUrlOk() (*string, bool)`

GetStargazersUrlOk returns a tuple with the StargazersUrl field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStargazersUrl

`func (o *GithubRepository) SetStargazersUrl(v string)`

SetStargazersUrl sets StargazersUrl field to given value.

### HasStargazersUrl

`func (o *GithubRepository) HasStargazersUrl() bool`

HasStargazersUrl returns a boolean if a field has been set.

### GetStatusesUrl

`func (o *GithubRepository) GetStatusesUrl() string`

GetStatusesUrl returns the StatusesUrl field if non-nil, zero value otherwise.

### GetStatusesUrlOk

`func (o *GithubRepository) GetStatusesUrlOk() (*string, bool)`

GetStatusesUrlOk returns a tuple with the StatusesUrl field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStatusesUrl

`func (o *GithubRepository) SetStatusesUrl(v string)`

SetStatusesUrl sets StatusesUrl field to given value.

### HasStatusesUrl

`func (o *GithubRepository) HasStatusesUrl() bool`

HasStatusesUrl returns a boolean if a field has been set.

### GetSubscribersCount

`func (o *GithubRepository) GetSubscribersCount() int32`

GetSubscribersCount returns the SubscribersCount field if non-nil, zero value otherwise.

### GetSubscribersCountOk

`func (o *GithubRepository) GetSubscribersCountOk() (*int32, bool)`

GetSubscribersCountOk returns a tuple with the SubscribersCount field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSubscribersCount

`func (o *GithubRepository) SetSubscribersCount(v int32)`

SetSubscribersCount sets SubscribersCount field to given value.

### HasSubscribersCount

`func (o *GithubRepository) HasSubscribersCount() bool`

HasSubscribersCount returns a boolean if a field has been set.

### GetSubscribersUrl

`func (o *GithubRepository) GetSubscribersUrl() string`

GetSubscribersUrl returns the SubscribersUrl field if non-nil, zero value otherwise.

### GetSubscribersUrlOk

`func (o *GithubRepository) GetSubscribersUrlOk() (*string, bool)`

GetSubscribersUrlOk returns a tuple with the SubscribersUrl field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSubscribersUrl

`func (o *GithubRepository) SetSubscribersUrl(v string)`

SetSubscribersUrl sets SubscribersUrl field to given value.

### HasSubscribersUrl

`func (o *GithubRepository) HasSubscribersUrl() bool`

HasSubscribersUrl returns a boolean if a field has been set.

### GetSubscriptionUrl

`func (o *GithubRepository) GetSubscriptionUrl() string`

GetSubscriptionUrl returns the SubscriptionUrl field if non-nil, zero value otherwise.

### GetSubscriptionUrlOk

`func (o *GithubRepository) GetSubscriptionUrlOk() (*string, bool)`

GetSubscriptionUrlOk returns a tuple with the SubscriptionUrl field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSubscriptionUrl

`func (o *GithubRepository) SetSubscriptionUrl(v string)`

SetSubscriptionUrl sets SubscriptionUrl field to given value.

### HasSubscriptionUrl

`func (o *GithubRepository) HasSubscriptionUrl() bool`

HasSubscriptionUrl returns a boolean if a field has been set.

### GetSvnUrl

`func (o *GithubRepository) GetSvnUrl() string`

GetSvnUrl returns the SvnUrl field if non-nil, zero value otherwise.

### GetSvnUrlOk

`func (o *GithubRepository) GetSvnUrlOk() (*string, bool)`

GetSvnUrlOk returns a tuple with the SvnUrl field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSvnUrl

`func (o *GithubRepository) SetSvnUrl(v string)`

SetSvnUrl sets SvnUrl field to given value.

### HasSvnUrl

`func (o *GithubRepository) HasSvnUrl() bool`

HasSvnUrl returns a boolean if a field has been set.

### GetTagsUrl

`func (o *GithubRepository) GetTagsUrl() string`

GetTagsUrl returns the TagsUrl field if non-nil, zero value otherwise.

### GetTagsUrlOk

`func (o *GithubRepository) GetTagsUrlOk() (*string, bool)`

GetTagsUrlOk returns a tuple with the TagsUrl field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTagsUrl

`func (o *GithubRepository) SetTagsUrl(v string)`

SetTagsUrl sets TagsUrl field to given value.

### HasTagsUrl

`func (o *GithubRepository) HasTagsUrl() bool`

HasTagsUrl returns a boolean if a field has been set.

### GetTeamId

`func (o *GithubRepository) GetTeamId() int32`

GetTeamId returns the TeamId field if non-nil, zero value otherwise.

### GetTeamIdOk

`func (o *GithubRepository) GetTeamIdOk() (*int32, bool)`

GetTeamIdOk returns a tuple with the TeamId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTeamId

`func (o *GithubRepository) SetTeamId(v int32)`

SetTeamId sets TeamId field to given value.

### HasTeamId

`func (o *GithubRepository) HasTeamId() bool`

HasTeamId returns a boolean if a field has been set.

### GetTeamsUrl

`func (o *GithubRepository) GetTeamsUrl() string`

GetTeamsUrl returns the TeamsUrl field if non-nil, zero value otherwise.

### GetTeamsUrlOk

`func (o *GithubRepository) GetTeamsUrlOk() (*string, bool)`

GetTeamsUrlOk returns a tuple with the TeamsUrl field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTeamsUrl

`func (o *GithubRepository) SetTeamsUrl(v string)`

SetTeamsUrl sets TeamsUrl field to given value.

### HasTeamsUrl

`func (o *GithubRepository) HasTeamsUrl() bool`

HasTeamsUrl returns a boolean if a field has been set.

### GetTextMatches

`func (o *GithubRepository) GetTextMatches() []GithubTextMatch`

GetTextMatches returns the TextMatches field if non-nil, zero value otherwise.

### GetTextMatchesOk

`func (o *GithubRepository) GetTextMatchesOk() (*[]GithubTextMatch, bool)`

GetTextMatchesOk returns a tuple with the TextMatches field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTextMatches

`func (o *GithubRepository) SetTextMatches(v []GithubTextMatch)`

SetTextMatches sets TextMatches field to given value.

### HasTextMatches

`func (o *GithubRepository) HasTextMatches() bool`

HasTextMatches returns a boolean if a field has been set.

### GetTopics

`func (o *GithubRepository) GetTopics() []string`

GetTopics returns the Topics field if non-nil, zero value otherwise.

### GetTopicsOk

`func (o *GithubRepository) GetTopicsOk() (*[]string, bool)`

GetTopicsOk returns a tuple with the Topics field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTopics

`func (o *GithubRepository) SetTopics(v []string)`

SetTopics sets Topics field to given value.

### HasTopics

`func (o *GithubRepository) HasTopics() bool`

HasTopics returns a boolean if a field has been set.

### GetTreesUrl

`func (o *GithubRepository) GetTreesUrl() string`

GetTreesUrl returns the TreesUrl field if non-nil, zero value otherwise.

### GetTreesUrlOk

`func (o *GithubRepository) GetTreesUrlOk() (*string, bool)`

GetTreesUrlOk returns a tuple with the TreesUrl field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTreesUrl

`func (o *GithubRepository) SetTreesUrl(v string)`

SetTreesUrl sets TreesUrl field to given value.

### HasTreesUrl

`func (o *GithubRepository) HasTreesUrl() bool`

HasTreesUrl returns a boolean if a field has been set.

### GetUpdatedAt

`func (o *GithubRepository) GetUpdatedAt() GithubTimestamp`

GetUpdatedAt returns the UpdatedAt field if non-nil, zero value otherwise.

### GetUpdatedAtOk

`func (o *GithubRepository) GetUpdatedAtOk() (*GithubTimestamp, bool)`

GetUpdatedAtOk returns a tuple with the UpdatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdatedAt

`func (o *GithubRepository) SetUpdatedAt(v GithubTimestamp)`

SetUpdatedAt sets UpdatedAt field to given value.

### HasUpdatedAt

`func (o *GithubRepository) HasUpdatedAt() bool`

HasUpdatedAt returns a boolean if a field has been set.

### GetUrl

`func (o *GithubRepository) GetUrl() string`

GetUrl returns the Url field if non-nil, zero value otherwise.

### GetUrlOk

`func (o *GithubRepository) GetUrlOk() (*string, bool)`

GetUrlOk returns a tuple with the Url field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUrl

`func (o *GithubRepository) SetUrl(v string)`

SetUrl sets Url field to given value.

### HasUrl

`func (o *GithubRepository) HasUrl() bool`

HasUrl returns a boolean if a field has been set.

### GetWatchersCount

`func (o *GithubRepository) GetWatchersCount() int32`

GetWatchersCount returns the WatchersCount field if non-nil, zero value otherwise.

### GetWatchersCountOk

`func (o *GithubRepository) GetWatchersCountOk() (*int32, bool)`

GetWatchersCountOk returns a tuple with the WatchersCount field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWatchersCount

`func (o *GithubRepository) SetWatchersCount(v int32)`

SetWatchersCount sets WatchersCount field to given value.

### HasWatchersCount

`func (o *GithubRepository) HasWatchersCount() bool`

HasWatchersCount returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


