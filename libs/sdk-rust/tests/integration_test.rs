use daytona::{Client, Config, DaytonaError};

#[tokio::test]
async fn client_constructs_with_api_key() {
    let config = Config::builder().api_key("test-key").build();
    let client = Client::new(config);
    assert!(client.is_ok());
}

#[tokio::test]
async fn client_fails_without_auth() {
    let config = Config::builder().build();
    let client = Client::new(Config {
        api_key: None,
        jwt_token: None,
        organization_id: None,
        api_url: config.api_url,
        target: config.target,
        timeout: None,
        otel_enabled: false,
        experimental: None,
    });
    assert!(client.is_err());
}

#[tokio::test]
async fn client_constructs_with_jwt_token() {
    let config = Config::builder()
        .jwt_token("test-jwt-token")
        .organization_id("test-org-id")
        .build();
    let client = Client::new(config);
    assert!(client.is_ok());
}

#[tokio::test]
async fn client_fails_with_jwt_but_no_organization() {
    let config = Config::builder().jwt_token("test-jwt-token").build();
    let client = Client::new(config);
    assert!(client.is_err());
}

#[test]
fn config_builder_sets_api_url() {
    let config = Config::builder()
        .api_key("test-key")
        .api_url("https://custom.api.url/api")
        .build();
    assert_eq!(config.api_url, "https://custom.api.url/api");
}

#[test]
fn config_builder_sets_target() {
    let config = Config::builder()
        .api_key("test-key")
        .target("us-west-2")
        .build();
    assert_eq!(config.target, Some("us-west-2".to_string()));
}

#[test]
fn config_builder_sets_timeout() {
    use std::time::Duration;
    let config = Config::builder()
        .api_key("test-key")
        .timeout(Duration::from_secs(120))
        .build();
    assert_eq!(config.timeout, Some(Duration::from_secs(120)));
}

#[test]
fn error_is_not_found() {
    let error = DaytonaError::NotFound("resource not found".to_string());
    assert!(error.is_not_found());
    assert!(!error.is_rate_limit());
    assert!(!error.is_timeout());
}

#[test]
fn error_is_rate_limit() {
    let error = DaytonaError::RateLimit("rate limit exceeded".to_string());
    assert!(!error.is_not_found());
    assert!(error.is_rate_limit());
    assert!(!error.is_timeout());
}

#[test]
fn error_is_timeout() {
    let error = DaytonaError::Timeout("request timed out".to_string());
    assert!(!error.is_not_found());
    assert!(!error.is_rate_limit());
    assert!(error.is_timeout());
}

#[test]
fn error_status_code() {
    let error = DaytonaError::NotFound("not found".to_string());
    assert_eq!(error.status_code(), Some(404));

    let error = DaytonaError::RateLimit("rate limited".to_string());
    assert_eq!(error.status_code(), Some(429));

    let error = DaytonaError::Timeout("timed out".to_string());
    assert_eq!(error.status_code(), Some(408));

    let error = DaytonaError::Unauthorized("unauthorized".to_string());
    assert_eq!(error.status_code(), Some(401));

    let error = DaytonaError::InternalServerError("internal error".to_string());
    assert_eq!(error.status_code(), Some(500));

    let error = DaytonaError::Network("network error".to_string());
    assert_eq!(error.status_code(), None);
}

#[test]
fn error_message() {
    let error = DaytonaError::NotFound("resource not found".to_string());
    assert_eq!(error.message(), "resource not found");

    let error = DaytonaError::Api {
        status: 400,
        message: "bad request".to_string(),
    };
    assert_eq!(error.message(), "bad request");
}
