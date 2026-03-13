// Copyright 2019-2025 Daytona Labs Inc.
// SPDX-License-Identifier: Apache-2.0

use crate::error::DaytonaError;
use crate::services::ServiceClient;
use crate::types::{ExecutionError, ExecutionResult, InterpreterContext};
use futures::{SinkExt, StreamExt};
use reqwest::Url;
use serde_json::{json, Value};
use tokio_tungstenite::{
    connect_async_with_config,
    tungstenite::{
        handshake::client::{generate_key, Request},
        Message,
    },
};

#[derive(Clone)]
pub struct CodeInterpreterService {
    client: ServiceClient,
}

impl CodeInterpreterService {
    pub fn new(client: ServiceClient) -> Self {
        Self { client }
    }

    pub async fn run_code(
        &self,
        code: &str,
        context_id: Option<&str>,
        timeout: Option<i32>,
    ) -> Result<ExecutionResult, DaytonaError> {
        let ws_url = format!(
            "{}/process/interpreter/execute",
            self.client
                .base_url
                .replace("https://", "wss://")
                .replace("http://", "ws://")
                .trim_end_matches('/')
        );

        // Build request with auth headers
        let host = Url::parse(&ws_url)
            .map_err(|e| DaytonaError::Network(e.to_string()))?
            .host_str()
            .unwrap_or("")
            .to_string();

        let request = Request::builder()
            .uri(&ws_url)
            .header("Host", host)
            .header("Connection", "Upgrade")
            .header("Upgrade", "websocket")
            .header("Sec-WebSocket-Version", "13")
            .header("Sec-WebSocket-Key", generate_key());

        // Add auth headers
        let request = if let Some(token) = self.client.config.bearer_token() {
            request.header("Authorization", format!("Bearer {}", token))
        } else {
            request
        };

        let request = request
            .header("X-Daytona-Source", "rust-sdk")
            .header("X-Daytona-SDK-Version", env!("CARGO_PKG_VERSION"));

        let request = request
            .body(())
            .map_err(|e| DaytonaError::Network(e.to_string()))?;

        let (mut ws, _) = connect_async_with_config(request, None, false)
            .await
            .map_err(|e| DaytonaError::Network(e.to_string()))?;

        ws.send(Message::Text(
            json!({
                "code": code,
                "contextId": context_id,
                "timeout": timeout
            })
            .to_string(),
        ))
        .await
        .map_err(|e| DaytonaError::Network(e.to_string()))?;

        let mut result = ExecutionResult::default();

        while let Some(msg) = ws.next().await {
            let msg = msg.map_err(|e| DaytonaError::Network(e.to_string()))?;
            match msg {
                Message::Text(text) => {
                    if let Ok(v) = serde_json::from_str::<Value>(&text) {
                        match v.get("type").and_then(Value::as_str).unwrap_or_default() {
                            "stdout" => result.stdout.push_str(
                                v.get("text").and_then(Value::as_str).unwrap_or_default(),
                            ),
                            "stderr" => result.stderr.push_str(
                                v.get("text").and_then(Value::as_str).unwrap_or_default(),
                            ),
                            "error" => {
                                result.error = Some(ExecutionError {
                                    name: v
                                        .get("name")
                                        .and_then(Value::as_str)
                                        .unwrap_or_default()
                                        .to_string(),
                                    value: v
                                        .get("value")
                                        .and_then(Value::as_str)
                                        .unwrap_or_default()
                                        .to_string(),
                                    traceback: v
                                        .get("traceback")
                                        .and_then(Value::as_str)
                                        .map(ToString::to_string),
                                })
                            }
                            "control" => {
                                let t = v.get("text").and_then(Value::as_str).unwrap_or_default();
                                if t == "completed" || t == "interrupted" {
                                    break;
                                }
                            }
                            _ => {}
                        }
                    }
                }
                Message::Close(_) => break,
                _ => {}
            }
        }

        Ok(result)
    }
    pub async fn create_context(
        &self,
        cwd: Option<&str>,
    ) -> Result<InterpreterContext, DaytonaError> {
        self.client
            .post_json("/process/interpreter/contexts", &json!({ "cwd": cwd }))
            .await
    }

    pub async fn delete_context(&self, id: &str) -> Result<(), DaytonaError> {
        self.client
            .delete_empty(&format!("/process/interpreter/contexts/{id}"))
            .await
    }

    pub async fn list_contexts(&self) -> Result<Vec<InterpreterContext>, DaytonaError> {
        let value: Value = self
            .client
            .get_json("/process/interpreter/contexts")
            .await?;
        Ok(value
            .get("contexts")
            .and_then(Value::as_array)
            .cloned()
            .unwrap_or_default()
            .into_iter()
            .filter_map(|v| serde_json::from_value(v).ok())
            .collect())
    }

    #[allow(non_snake_case)]
    pub async fn runCode(
        &self,
        code: &str,
        context_id: Option<&str>,
        timeout: Option<i32>,
    ) -> Result<ExecutionResult, DaytonaError> {
        self.run_code(code, context_id, timeout).await
    }

    #[allow(non_snake_case)]
    pub async fn createContext(
        &self,
        cwd: Option<&str>,
    ) -> Result<InterpreterContext, DaytonaError> {
        self.create_context(cwd).await
    }

    #[allow(non_snake_case)]
    pub async fn deleteContext(&self, id: &str) -> Result<(), DaytonaError> {
        self.delete_context(id).await
    }

    #[allow(non_snake_case)]
    pub async fn listContexts(&self) -> Result<Vec<InterpreterContext>, DaytonaError> {
        self.list_contexts().await
    }
}
