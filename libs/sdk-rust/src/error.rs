// Copyright 2019-2025 Daytona Labs Inc.
// SPDX-License-Identifier: Apache-2.0

use reqwest::StatusCode;
use thiserror::Error;

#[derive(Debug, Error)]
pub enum DaytonaError {
    #[error("resource not found: {0}")]
    NotFound(String),
    #[error("rate limit exceeded: {0}")]
    RateLimit(String),
    #[error("request timed out: {0}")]
    Timeout(String),
    #[error("unauthorized: {0}")]
    Unauthorized(String),
    #[error("internal server error: {0}")]
    InternalServerError(String),
    #[error("network error: {0}")]
    Network(String),
    #[error("api error ({status}): {message}")]
    Api { status: u16, message: String },
}

impl DaytonaError {
    pub fn from_status(status: StatusCode, message: impl Into<String>) -> Self {
        let msg = message.into();
        match status {
            StatusCode::NOT_FOUND => Self::NotFound(msg),
            StatusCode::TOO_MANY_REQUESTS => Self::RateLimit(msg),
            StatusCode::REQUEST_TIMEOUT => Self::Timeout(msg),
            StatusCode::UNAUTHORIZED | StatusCode::FORBIDDEN => Self::Unauthorized(msg),
            StatusCode::INTERNAL_SERVER_ERROR
            | StatusCode::BAD_GATEWAY
            | StatusCode::SERVICE_UNAVAILABLE
            | StatusCode::GATEWAY_TIMEOUT => Self::InternalServerError(msg),
            _ => Self::Api {
                status: status.as_u16(),
                message: msg,
            },
        }
    }

    pub fn is_not_found(&self) -> bool {
        matches!(self, DaytonaError::NotFound(_))
    }

    pub fn is_rate_limit(&self) -> bool {
        matches!(self, DaytonaError::RateLimit(_))
    }

    pub fn is_timeout(&self) -> bool {
        matches!(self, DaytonaError::Timeout(_))
    }

    pub fn status_code(&self) -> Option<u16> {
        match self {
            DaytonaError::NotFound(_) => Some(404),
            DaytonaError::RateLimit(_) => Some(429),
            DaytonaError::Timeout(_) => Some(408),
            DaytonaError::Unauthorized(_) => Some(401),
            DaytonaError::InternalServerError(_) => Some(500),
            DaytonaError::Api { status, .. } => Some(*status),
            DaytonaError::Network(_) => None,
        }
    }

    pub fn message(&self) -> String {
        match self {
            DaytonaError::NotFound(msg) => msg.clone(),
            DaytonaError::RateLimit(msg) => msg.clone(),
            DaytonaError::Timeout(msg) => msg.clone(),
            DaytonaError::Unauthorized(msg) => msg.clone(),
            DaytonaError::InternalServerError(msg) => msg.clone(),
            DaytonaError::Network(msg) => msg.clone(),
            DaytonaError::Api { message, .. } => message.clone(),
        }
    }
}

impl From<reqwest::Error> for DaytonaError {
    fn from(e: reqwest::Error) -> Self {
        DaytonaError::Network(e.to_string())
    }
}

impl From<tokio_tungstenite::tungstenite::Error> for DaytonaError {
    fn from(e: tokio_tungstenite::tungstenite::Error) -> Self {
        DaytonaError::Network(e.to_string())
    }
}
