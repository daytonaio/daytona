

# DaytonaConfiguration


## Properties

| Name | Type | Description | Notes |
|------------ | ------------- | ------------- | -------------|
|**version** | **String** | Daytona version |  |
|**posthog** | [**PosthogConfig**](PosthogConfig.md) | PostHog configuration |  [optional] |
|**oidc** | [**OidcConfig**](OidcConfig.md) | OIDC configuration |  |
|**linkedAccountsEnabled** | **Boolean** | Whether linked accounts are enabled |  |
|**announcements** | [**Map&lt;String, Announcement&gt;**](Announcement.md) | System announcements |  |
|**pylonAppId** | **String** | Pylon application ID |  [optional] |
|**proxyTemplateUrl** | **String** | Proxy template URL |  |
|**proxyToolboxUrl** | **String** | Toolbox template URL |  |
|**defaultSnapshot** | **String** | Default snapshot for sandboxes |  |
|**dashboardUrl** | **String** | Dashboard URL |  |
|**maxAutoArchiveInterval** | **BigDecimal** | Maximum auto-archive interval in minutes |  |
|**maintananceMode** | **Boolean** | Whether maintenance mode is enabled |  |
|**environment** | **String** | Current environment |  |
|**billingApiUrl** | **String** | Billing API URL |  [optional] |
|**analyticsApiUrl** | **String** | Analytics API URL |  [optional] |
|**sshGatewayCommand** | **String** | SSH Gateway command |  [optional] |
|**sshGatewayPublicKey** | **String** | Base64 encoded SSH Gateway public key |  [optional] |
|**rateLimit** | [**RateLimitConfig**](RateLimitConfig.md) | Rate limit configuration |  [optional] |



