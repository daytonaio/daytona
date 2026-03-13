// Copyright 2019-2025 Daytona Labs Inc.
// SPDX-License-Identifier: Apache-2.0

use std::env;
use std::time::Duration;

pub const DEFAULT_API_URL: &str = "https://app.daytona.io/api";

#[derive(Debug, Clone)]
pub struct Config {
    pub api_key: Option<String>,
    pub jwt_token: Option<String>,
    pub organization_id: Option<String>,
    pub api_url: String,
    pub target: Option<String>,
    pub timeout: Option<Duration>,
    pub otel_enabled: bool,
    /// Experimental features configuration
    pub experimental: Option<serde_json::Value>,
}

impl Config {
    pub fn from_env() -> Self {
        Self {
            api_key: env::var("DAYTONA_API_KEY").ok(),
            jwt_token: env::var("DAYTONA_JWT_TOKEN").ok(),
            organization_id: env::var("DAYTONA_ORGANIZATION_ID").ok(),
            api_url: env::var("DAYTONA_API_URL")
                .or_else(|_| {
                    env::var("DAYTONA_SERVER_URL").inspect(|_url| {
                        eprintln!("[Deprecation Warning] Environment variable `DAYTONA_SERVER_URL` is deprecated and will be removed in future versions. Use `DAYTONA_API_URL` instead.");
                    })
                })
                .unwrap_or_else(|_| DEFAULT_API_URL.to_string()),
            target: env::var("DAYTONA_TARGET").ok(),
            timeout: env::var("DAYTONA_TIMEOUT")
                .ok()
                .and_then(|s| s.parse::<u64>().ok())
                .map(Duration::from_secs),
            otel_enabled: env::var("DAYTONA_OTEL_ENABLED")
                .ok()
                .map(|v| v == "true" || v == "1")
                .unwrap_or(false),
            experimental: None,
        }
    }

    pub fn builder() -> ConfigBuilder {
        ConfigBuilder::default()
    }

    pub fn bearer_token(&self) -> Option<&str> {
        self.api_key.as_deref().or(self.jwt_token.as_deref())
    }
}

#[derive(Debug, Default, Clone)]
pub struct ConfigBuilder {
    api_key: Option<String>,
    jwt_token: Option<String>,
    organization_id: Option<String>,
    api_url: Option<String>,
    target: Option<String>,
    timeout: Option<Duration>,
    otel_enabled: Option<bool>,
    experimental: Option<serde_json::Value>,
}

impl ConfigBuilder {
    pub fn api_key(mut self, value: impl Into<String>) -> Self {
        self.api_key = Some(value.into());
        self
    }

    pub fn jwt_token(mut self, value: impl Into<String>) -> Self {
        self.jwt_token = Some(value.into());
        self
    }

    pub fn organization_id(mut self, value: impl Into<String>) -> Self {
        self.organization_id = Some(value.into());
        self
    }

    pub fn api_url(mut self, value: impl Into<String>) -> Self {
        self.api_url = Some(value.into());
        self
    }

    pub fn target(mut self, value: impl Into<String>) -> Self {
        self.target = Some(value.into());
        self
    }

    pub fn timeout(mut self, value: Duration) -> Self {
        self.timeout = Some(value);
        self
    }

    pub fn otel_enabled(mut self, value: bool) -> Self {
        self.otel_enabled = Some(value);
        self
    }

    pub fn experimental(mut self, value: serde_json::Value) -> Self {
        self.experimental = Some(value);
        self
    }

    pub fn build(self) -> Config {
        let env_cfg = Config::from_env();
        Config {
            api_key: self.api_key.or(env_cfg.api_key),
            jwt_token: self.jwt_token.or(env_cfg.jwt_token),
            organization_id: self.organization_id.or(env_cfg.organization_id),
            api_url: self.api_url.unwrap_or(env_cfg.api_url),
            target: self.target.or(env_cfg.target),
            timeout: self.timeout.or(env_cfg.timeout),
            otel_enabled: self.otel_enabled.unwrap_or(env_cfg.otel_enabled),
            experimental: self.experimental.or(env_cfg.experimental),
        }
    }
}
