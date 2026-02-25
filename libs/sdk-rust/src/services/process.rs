// Copyright 2019-2025 Daytona Labs Inc.
// SPDX-License-Identifier: Apache-2.0

use crate::error::DaytonaError;
use crate::services::ServiceClient;
use crate::types::{
    Command, ExecuteResponse, PtyCreateResponse, PtySessionInfo, Session, SessionExecuteResponse,
};
use futures::StreamExt;
use serde_json::{json, Value};
use tokio::sync::mpsc;
use tokio_tungstenite::connect_async;
use tokio_tungstenite::tungstenite::{client::IntoClientRequest, Message};

const STDOUT_PREFIX: [u8; 3] = [0x01, 0x01, 0x01];
const STDERR_PREFIX: [u8; 3] = [0x02, 0x02, 0x02];
const MAX_PREFIX_LEN: usize = 3;

fn flush_to_channel(
    data: &[u8],
    current_type: &str,
    stdout: &mpsc::Sender<String>,
    stderr: &mpsc::Sender<String>,
) {
    if data.is_empty() {
        return;
    }
    let text = String::from_utf8_lossy(data);
    match current_type {
        "stdout" => {
            let _ = stdout.try_send(text.to_string());
        }
        "stderr" => {
            let _ = stderr.try_send(text.to_string());
        }
        _ => {}
    }
}

#[derive(Clone)]
pub struct ProcessService {
    client: ServiceClient,
}

impl ProcessService {
    pub fn new(client: ServiceClient) -> Self {
        Self { client }
    }

    pub async fn execute_command(
        &self,
        command: &str,
        cwd: Option<&str>,
        env: Option<Value>,
        timeout: Option<i32>,
    ) -> Result<ExecuteResponse, DaytonaError> {
        self.client
            .post_json(
                "/process/execute",
                &json!({ "command": command, "cwd": cwd, "env": env, "timeout": timeout }),
            )
            .await
    }

    /// Execute code using the appropriate language toolbox.
    ///
    /// The command is generated using the CodeToolbox for the specified language,
    /// then executed via `execute_command` with the provided environment variables.
    pub async fn code_run(
        &self,
        code: &str,
        language: &str,
        params: Option<&crate::code_toolbox::CodeRunParams>,
        timeout: Option<i32>,
    ) -> Result<ExecuteResponse, DaytonaError> {
        let toolbox = crate::code_toolbox::CodeLanguage::from_str(language)
            .map(|lang| lang.toolbox())
            .unwrap_or_else(|| Box::new(crate::code_toolbox::PythonCodeToolbox::new()));

        let command = toolbox.get_run_command(code, params);
        let env = params.and_then(|p| {
            p.env.as_ref().and_then(|e| {
                if e.is_empty() {
                    None
                } else {
                    Some(serde_json::to_value(e).unwrap_or(Value::Null))
                }
            })
        });

        self.execute_command(&command, None, env, timeout).await
    }

    pub async fn create_pty(
        &self,
        id: Option<&str>,
        cwd: Option<&str>,
        cols: Option<i32>,
        rows: Option<i32>,
    ) -> Result<PtyCreateResponse, DaytonaError> {
        self.client
            .post_json(
                "/process/pty/create",
                &json!({ "id": id, "cwd": cwd, "cols": cols, "rows": rows }),
            )
            .await
    }

    pub async fn list_pty(&self) -> Result<Vec<PtySessionInfo>, DaytonaError> {
        self.client.get_json("/process/pty/list").await
    }

    pub async fn resize_pty(
        &self,
        session_id: &str,
        cols: i32,
        rows: i32,
    ) -> Result<PtySessionInfo, DaytonaError> {
        self.client
            .post_json(
                &format!("/process/pty/{session_id}/resize"),
                &json!({ "cols": cols, "rows": rows }),
            )
            .await
    }

    pub async fn send_input(&self, session_id: &str, data: &str) -> Result<(), DaytonaError> {
        self.client
            .post_empty(
                &format!("/process/pty/{session_id}/input"),
                &json!({ "data": data }),
            )
            .await
    }

    pub async fn read_output(&self, session_id: &str) -> Result<String, DaytonaError> {
        let value: Value = self
            .client
            .get_json(&format!("/process/pty/{session_id}/output"))
            .await?;
        Ok(value
            .get("output")
            .and_then(Value::as_str)
            .unwrap_or_default()
            .to_string())
    }

    pub async fn create_session(&self, session_id: &str) -> Result<(), DaytonaError> {
        self.client
            .post_empty("/process/session", &json!({ "sessionId": session_id }))
            .await
    }

    pub async fn get_session(&self, session_id: &str) -> Result<Session, DaytonaError> {
        self.client
            .get_json(&format!("/process/session/{}", session_id))
            .await
    }

    pub async fn delete_session(&self, session_id: &str) -> Result<(), DaytonaError> {
        self.client
            .delete_empty(&format!("/process/session/{}", session_id))
            .await
    }

    pub async fn list_sessions(&self) -> Result<Vec<Session>, DaytonaError> {
        self.client.get_json("/process/session").await
    }

    pub async fn execute_session_command(
        &self,
        session_id: &str,
        command: &str,
        run_async: bool,
        suppress_input_echo: bool,
    ) -> Result<SessionExecuteResponse, DaytonaError> {
        self.client
            .post_json(
                &format!("/process/session/{}/command", session_id),
                &json!({
                    "command": command,
                    "runAsync": run_async,
                    "suppressInputEcho": suppress_input_echo
                }),
            )
            .await
    }

    pub async fn get_session_command(
        &self,
        session_id: &str,
        command_id: &str,
    ) -> Result<Command, DaytonaError> {
        self.client
            .get_json(&format!(
                "/process/session/{}/command/{}",
                session_id, command_id
            ))
            .await
    }

    /// Send input data to a running command in a session.
    ///
    /// This is useful for interactive commands that require user input.
    pub async fn send_session_command_input(
        &self,
        session_id: &str,
        command_id: &str,
        data: &str,
    ) -> Result<(), DaytonaError> {
        self.client
            .post_empty(
                &format!(
                    "/process/session/{}/command/{}/input",
                    session_id, command_id
                ),
                &json!({ "data": data }),
            )
            .await
    }
    pub async fn get_session_command_logs(
        &self,
        session_id: &str,
        command_id: &str,
    ) -> Result<String, DaytonaError> {
        let value: serde_json::Value = self
            .client
            .get_json(&format!(
                "/process/session/{}/command/{}/logs",
                session_id, command_id
            ))
            .await?;
        Ok(value
            .get("logs")
            .and_then(serde_json::Value::as_str)
            .unwrap_or_default()
            .to_string())
    }

    /// Stream command logs via WebSocket with stdout/stderr demux
    ///
    /// Uses 3-byte markers: \[0x01,0x01,0x01\] for stdout, \[0x02,0x02,0x02\] for stderr
    pub async fn get_session_command_logs_stream(
        &self,
        session_id: &str,
        command_id: &str,
        stdout: mpsc::Sender<String>,
        stderr: mpsc::Sender<String>,
    ) -> Result<(), DaytonaError> {
        let ws_url = format!(
            "{}/process/session/{}/command/{}/logs?follow=true",
            self.client
                .base_url
                .replace("https://", "wss://")
                .replace("http://", "ws://")
                .trim_end_matches('/'),
            session_id,
            command_id
        );

        let token = self.client.config.bearer_token().ok_or_else(|| {
            DaytonaError::Unauthorized("missing DAYTONA_API_KEY or DAYTONA_JWT_TOKEN".into())
        })?;

        let mut request = ws_url
            .into_client_request()
            .map_err(|e| DaytonaError::Network(e.to_string()))?;
        request.headers_mut().insert(
            "Authorization",
            format!("Bearer {}", token)
                .parse()
                .map_err(|_| DaytonaError::Network("invalid auth header value".into()))?,
        );
        request.headers_mut().insert(
            "X-Daytona-Source",
            "rust-sdk"
                .parse()
                .map_err(|_| DaytonaError::Network("invalid header value".into()))?,
        );

        let (mut ws, _) = connect_async(request)
            .await
            .map_err(|e| DaytonaError::Network(e.to_string()))?;

        let mut current_type = "stdout";
        let mut carry: Vec<u8> = Vec::new();

        while let Some(msg) = ws.next().await {
            let msg = msg.map_err(|e| DaytonaError::Network(e.to_string()))?;
            let bytes: Vec<u8> = match msg {
                Message::Binary(data) => data,
                Message::Text(text) => text.into_bytes(),
                Message::Close(_) => break,
                _ => continue,
            };

            let mut buffer = Vec::with_capacity(carry.len() + bytes.len());
            buffer.extend_from_slice(&carry);
            buffer.extend_from_slice(&bytes);

            let mut start = 0;
            let mut pos = 0;

            while pos + MAX_PREFIX_LEN <= buffer.len() {
                if buffer[pos..pos + MAX_PREFIX_LEN] == STDOUT_PREFIX {
                    if start < pos {
                        flush_to_channel(&buffer[start..pos], current_type, &stdout, &stderr);
                    }
                    current_type = "stdout";
                    pos += MAX_PREFIX_LEN;
                    start = pos;
                } else if buffer[pos..pos + MAX_PREFIX_LEN] == STDERR_PREFIX {
                    if start < pos {
                        flush_to_channel(&buffer[start..pos], current_type, &stdout, &stderr);
                    }
                    current_type = "stderr";
                    pos += MAX_PREFIX_LEN;
                    start = pos;
                } else {
                    pos += 1;
                }
            }

            // Flush verified data (no markers in [start..pos])
            if start < pos {
                flush_to_channel(&buffer[start..pos], current_type, &stdout, &stderr);
            }

            // Carry remaining bytes that might contain a partial marker
            carry = buffer[pos..].to_vec();
        }

        // Flush any remaining carry buffer
        if !carry.is_empty() {
            flush_to_channel(&carry, current_type, &stdout, &stderr);
        }

        Ok(())
    }

    #[allow(non_snake_case)]
    pub async fn executeCommand(
        &self,
        command: &str,
        cwd: Option<&str>,
        env: Option<Value>,
        timeout: Option<i32>,
    ) -> Result<ExecuteResponse, DaytonaError> {
        self.execute_command(command, cwd, env, timeout).await
    }

    #[allow(non_snake_case)]
    pub async fn codeRun(
        &self,
        code: &str,
        language: &str,
        params: Option<&crate::code_toolbox::CodeRunParams>,
        timeout: Option<i32>,
    ) -> Result<ExecuteResponse, DaytonaError> {
        self.code_run(code, language, params, timeout).await
    }

    #[allow(non_snake_case)]
    pub async fn createPty(
        &self,
        id: Option<&str>,
        cwd: Option<&str>,
        cols: Option<i32>,
        rows: Option<i32>,
    ) -> Result<PtyCreateResponse, DaytonaError> {
        self.create_pty(id, cwd, cols, rows).await
    }

    #[allow(non_snake_case)]
    pub async fn listPty(&self) -> Result<Vec<PtySessionInfo>, DaytonaError> {
        self.list_pty().await
    }

    #[allow(non_snake_case)]
    pub async fn resizePty(
        &self,
        session_id: &str,
        cols: i32,
        rows: i32,
    ) -> Result<PtySessionInfo, DaytonaError> {
        self.resize_pty(session_id, cols, rows).await
    }

    #[allow(non_snake_case)]
    pub async fn sendInput(&self, session_id: &str, data: &str) -> Result<(), DaytonaError> {
        self.send_input(session_id, data).await
    }

    pub async fn kill_pty_session(&self, session_id: &str) -> Result<(), DaytonaError> {
        self.client
            .delete_empty(&format!("/process/pty/{}", session_id))
            .await
    }
    #[allow(non_snake_case)]
    pub async fn createSession(&self, session_id: &str) -> Result<(), DaytonaError> {
        self.create_session(session_id).await
    }

    #[allow(non_snake_case)]
    pub async fn getSession(&self, session_id: &str) -> Result<Session, DaytonaError> {
        self.get_session(session_id).await
    }

    #[allow(non_snake_case)]
    pub async fn deleteSession(&self, session_id: &str) -> Result<(), DaytonaError> {
        self.delete_session(session_id).await
    }

    #[allow(non_snake_case)]
    pub async fn listSessions(&self) -> Result<Vec<Session>, DaytonaError> {
        self.list_sessions().await
    }

    #[allow(non_snake_case)]
    pub async fn executeSessionCommand(
        &self,
        session_id: &str,
        command: &str,
        run_async: bool,
        suppress_input_echo: bool,
    ) -> Result<SessionExecuteResponse, DaytonaError> {
        self.execute_session_command(session_id, command, run_async, suppress_input_echo)
            .await
    }

    #[allow(non_snake_case)]
    pub async fn getSessionCommand(
        &self,
        session_id: &str,
        command_id: &str,
    ) -> Result<Command, DaytonaError> {
        self.get_session_command(session_id, command_id).await
    }

    #[allow(non_snake_case)]
    pub async fn sendSessionCommandInput(
        &self,
        session_id: &str,
        command_id: &str,
        data: &str,
    ) -> Result<(), DaytonaError> {
        self.send_session_command_input(session_id, command_id, data)
            .await
    }
    #[allow(non_snake_case)]
    pub async fn getSessionCommandLogsStream(
        &self,
        session_id: &str,
        command_id: &str,
        stdout: mpsc::Sender<String>,
        stderr: mpsc::Sender<String>,
    ) -> Result<(), DaytonaError> {
        self.get_session_command_logs_stream(session_id, command_id, stdout, stderr)
            .await
    }

    pub async fn get_pty_session_info(
        &self,
        session_id: &str,
    ) -> Result<PtySessionInfo, DaytonaError> {
        self.client
            .get_json(&format!("/process/pty/{}", session_id))
            .await
    }

    #[allow(non_snake_case)]
    pub async fn killPtySession(&self, session_id: &str) -> Result<(), DaytonaError> {
        self.kill_pty_session(session_id).await
    }

    #[allow(non_snake_case)]
    pub async fn getPtySessionInfo(
        &self,
        session_id: &str,
    ) -> Result<PtySessionInfo, DaytonaError> {
        self.get_pty_session_info(session_id).await
    }
}
