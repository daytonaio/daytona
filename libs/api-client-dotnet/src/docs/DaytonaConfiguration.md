# Daytona.ApiClient.Model.DaytonaConfiguration

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**VarVersion** | **string** | Daytona version | 
**Posthog** | [**PosthogConfig**](PosthogConfig.md) | PostHog configuration | [optional] 
**Oidc** | [**OidcConfig**](OidcConfig.md) | OIDC configuration | 
**LinkedAccountsEnabled** | **bool** | Whether linked accounts are enabled | 
**Announcements** | [**Dictionary&lt;string, Announcement&gt;**](Announcement.md) | System announcements | 
**PylonAppId** | **string** | Pylon application ID | [optional] 
**ProxyTemplateUrl** | **string** | Proxy template URL | 
**ProxyToolboxUrl** | **string** | Toolbox template URL | 
**DefaultSnapshot** | **string** | Default snapshot for sandboxes | 
**DashboardUrl** | **string** | Dashboard URL | 
**MaxAutoArchiveInterval** | **decimal** | Maximum auto-archive interval in minutes | 
**MaintananceMode** | **bool** | Whether maintenance mode is enabled | 
**VarEnvironment** | **string** | Current environment | 
**BillingApiUrl** | **string** | Billing API URL | [optional] 
**AnalyticsApiUrl** | **string** | Analytics API URL | [optional] 
**SshGatewayCommand** | **string** | SSH Gateway command | [optional] 
**SshGatewayPublicKey** | **string** | Base64 encoded SSH Gateway public key | [optional] 
**RateLimit** | [**RateLimitConfig**](RateLimitConfig.md) | Rate limit configuration | [optional] 

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)

