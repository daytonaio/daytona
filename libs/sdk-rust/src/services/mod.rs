// Copyright 2019-2025 Daytona Labs Inc.
// SPDX-License-Identifier: Apache-2.0

pub mod code_interpreter;
pub mod computer_use;
pub mod filesystem;
pub mod git;
pub mod lsp_server;
pub mod object_storage;
pub mod process;
pub mod snapshot;
pub mod volume;

use crate::config::Config;
use crate::error::DaytonaError;
use crate::utils::{decode_empty, decode_json, with_auth};
use reqwest::Client as HttpClient;
use serde::Serialize;
use serde_json::Value;

#[derive(Clone)]
pub struct ServiceClient {
    pub base_url: String,
    pub config: Config,
    pub http: HttpClient,
}

impl ServiceClient {
    pub fn new(base_url: String, config: Config, http: HttpClient) -> Self {
        Self {
            base_url,
            config,
            http,
        }
    }

    pub async fn get_json<T: serde::de::DeserializeOwned>(
        &self,
        path: &str,
    ) -> Result<T, DaytonaError> {
        let req = with_auth(
            self.http.get(format!("{}{}", self.base_url, path)),
            &self.config,
        )?;
        decode_json(
            req.send()
                .await
                .map_err(|e| DaytonaError::Network(e.to_string()))?,
        )
        .await
    }

    pub async fn post_json<B: Serialize, T: serde::de::DeserializeOwned>(
        &self,
        path: &str,
        body: &B,
    ) -> Result<T, DaytonaError> {
        let req = with_auth(
            self.http.post(format!("{}{}", self.base_url, path)),
            &self.config,
        )?;
        decode_json(
            req.json(body)
                .send()
                .await
                .map_err(|e| DaytonaError::Network(e.to_string()))?,
        )
        .await
    }

    pub async fn post_json_value<B: Serialize>(
        &self,
        path: &str,
        body: &B,
    ) -> Result<Value, DaytonaError> {
        self.post_json(path, body).await
    }

    pub async fn post_empty<B: Serialize>(&self, path: &str, body: &B) -> Result<(), DaytonaError> {
        let req = with_auth(
            self.http.post(format!("{}{}", self.base_url, path)),
            &self.config,
        )?;
        decode_empty(
            req.json(body)
                .send()
                .await
                .map_err(|e| DaytonaError::Network(e.to_string()))?,
        )
        .await
    }

    pub async fn post_empty_no_body(&self, path: &str) -> Result<(), DaytonaError> {
        let req = with_auth(
            self.http.post(format!("{}{}", self.base_url, path)),
            &self.config,
        )?;
        decode_empty(
            req.send()
                .await
                .map_err(|e| DaytonaError::Network(e.to_string()))?,
        )
        .await
    }

    pub async fn delete_empty(&self, path: &str) -> Result<(), DaytonaError> {
        let req = with_auth(
            self.http.delete(format!("{}{}", self.base_url, path)),
            &self.config,
        )?;
        decode_empty(
            req.send()
                .await
                .map_err(|e| DaytonaError::Network(e.to_string()))?,
        )
        .await
    }

    pub async fn put_json<B: Serialize, T: serde::de::DeserializeOwned>(
        &self,
        path: &str,
        body: &B,
    ) -> Result<T, DaytonaError> {
        let req = with_auth(
            self.http.put(format!("{}{}", self.base_url, path)),
            &self.config,
        )?;
        decode_json(
            req.json(body)
                .send()
                .await
                .map_err(|e| DaytonaError::Network(e.to_string()))?,
        )
        .await
    }

    pub async fn put_empty<B: Serialize>(&self, path: &str, body: &B) -> Result<(), DaytonaError> {
        let req = with_auth(
            self.http.put(format!("{}{}", self.base_url, path)),
            &self.config,
        )?;
        decode_empty(
            req.json(body)
                .send()
                .await
                .map_err(|e| DaytonaError::Network(e.to_string()))?,
        )
        .await
    }
}
