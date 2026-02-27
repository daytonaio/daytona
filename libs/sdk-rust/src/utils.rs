// Copyright 2019-2025 Daytona Labs Inc.
// SPDX-License-Identifier: Apache-2.0

use crate::config::Config;
use crate::error::DaytonaError;
use reqwest::{RequestBuilder, Response};
use serde::de::DeserializeOwned;
use serde_json::Value;

pub fn with_auth(builder: RequestBuilder, config: &Config) -> Result<RequestBuilder, DaytonaError> {
    let token = config
        .bearer_token()
        .ok_or_else(|| DaytonaError::Unauthorized("Authentication required. Please set DAYTONA_API_KEY or DAYTONA_JWT_TOKEN environment variable, or provide api_key or jwt_token in Config".to_string()))?;

    let mut req = builder
        .bearer_auth(token)
        .header("X-Daytona-Source", "rust-sdk")
        .header("X-Daytona-SDK-Version", env!("CARGO_PKG_VERSION"));

    // Add target header if configured
    if let Some(target) = &config.target {
        req = req.header("X-Daytona-Target", target);
    }

    if config.jwt_token.is_some() {
        if let Some(org) = &config.organization_id {
            req = req.header("X-Daytona-Organization-ID", org);
        }
    }

    Ok(req)
}

pub async fn decode_json<T: DeserializeOwned>(resp: Response) -> Result<T, DaytonaError> {
    let status = resp.status();
    if status.is_success() {
        return resp.json::<T>().await.map_err(|e| DaytonaError::Api {
            status: 0,
            message: format!("failed to parse response body: {e}"),
        });
    }

    let text = resp.text().await.unwrap_or_default();
    Err(DaytonaError::from_status(status, text))
}

pub async fn decode_empty(resp: Response) -> Result<(), DaytonaError> {
    let status = resp.status();
    if status.is_success() {
        return Ok(());
    }

    let text = resp.text().await.unwrap_or_default();
    Err(DaytonaError::from_status(status, text))
}

pub fn join_url(base: &str, suffix: &str) -> String {
    format!(
        "{}/{}",
        base.trim_end_matches('/'),
        suffix.trim_start_matches('/')
    )
}

pub fn extract_toolbox_proxy_url(payload: &Value) -> Option<String> {
    payload
        .get("url")
        .and_then(Value::as_str)
        .map(ToString::to_string)
}
