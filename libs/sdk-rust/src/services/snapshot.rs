// Copyright 2019-2025 Daytona Labs Inc.
// SPDX-License-Identifier: Apache-2.0

use crate::error::DaytonaError;
use crate::services::ServiceClient;
use crate::types::{CreateSnapshotParams, PaginatedSnapshots, Snapshot};
use crate::utils::with_auth;
use futures::StreamExt;
use serde_json::json;
use tokio::sync::mpsc;

#[derive(Clone)]
pub struct SnapshotService {
    client: ServiceClient,
}

impl SnapshotService {
    pub fn new(client: ServiceClient) -> Self {
        Self { client }
    }

    pub async fn list(
        &self,
        page: Option<i32>,
        limit: Option<i32>,
    ) -> Result<PaginatedSnapshots, DaytonaError> {
        self.client
            .post_json("/snapshot/list", &json!({ "page": page, "limit": limit }))
            .await
    }

    pub async fn get(&self, name_or_id: &str) -> Result<Snapshot, DaytonaError> {
        self.client
            .get_json(&format!("/snapshot/{name_or_id}"))
            .await
    }

    pub async fn create(
        &self,
        params: &CreateSnapshotParams,
    ) -> Result<(Snapshot, mpsc::Receiver<String>), DaytonaError> {
        // 1. Create the snapshot
        let snapshot: Snapshot = self.client.post_json("/snapshot", params).await?;

        // 2. Set up log streaming channel
        let (log_tx, log_rx) = mpsc::channel::<String>(100);

        // 3. If state is PENDING_BUILD, start streaming logs in background
        let state = snapshot.state.as_deref().unwrap_or("");
        if state == "pending_build" || state == "pending" {
            let svc = self.clone();
            let snapshot_id = snapshot.id.clone();
            tokio::spawn(async move {
                let _ = svc.stream_build_logs(&snapshot_id, log_tx).await;
            });
        } else {
            // Not a build state — close the channel immediately
            drop(log_tx);
        }

        Ok((snapshot, log_rx))
    }

    async fn stream_build_logs(
        &self,
        snapshot_id: &str,
        log_tx: mpsc::Sender<String>,
    ) -> Result<(), DaytonaError> {
        // 1. Poll /snapshot/{id} until state is no longer pending/pending_build
        loop {
            let current: Snapshot = self
                .client
                .get_json(&format!("/snapshot/{snapshot_id}"))
                .await?;

            let state = current.state.as_deref().unwrap_or("");
            match state {
                "pending" | "pending_build" => {
                    tokio::time::sleep(std::time::Duration::from_secs(1)).await;
                }
                _ => break,
            }
        }

        // 2. Stream from GET /snapshots/{id}/build-logs?follow=true
        let url = format!(
            "{}/snapshots/{}/build-logs?follow=true",
            self.client.base_url, snapshot_id
        );

        let req = with_auth(self.client.http.get(&url), &self.client.config)?;
        let resp = req
            .send()
            .await
            .map_err(|e| DaytonaError::Network(e.to_string()))?;

        if !resp.status().is_success() {
            let status = resp.status();
            let text = resp.text().await.unwrap_or_default();
            return Err(DaytonaError::from_status(status, text));
        }

        // 3. Use reqwest Response::bytes_stream() for streaming
        let mut stream = resp.bytes_stream();
        let mut buffer = String::new();

        while let Some(chunk) = stream.next().await {
            let chunk = chunk.map_err(|e| DaytonaError::Network(e.to_string()))?;
            let text = String::from_utf8_lossy(&chunk);
            buffer.push_str(&text);

            // 4. Parse lines and send to log_tx
            while let Some(pos) = buffer.find('\n') {
                let line = buffer[..pos].trim_end_matches('\r').to_string();
                buffer = buffer[pos + 1..].to_string();
                if !line.is_empty() && log_tx.send(line).await.is_err() {
                    // Receiver dropped, stop streaming
                    return Ok(());
                }
            }
        }

        // Send any remaining content in buffer
        let remaining = buffer.trim().to_string();
        if !remaining.is_empty() {
            let _ = log_tx.send(remaining).await;
        }

        Ok(())
    }

    pub async fn delete(&self, name_or_id: &str) -> Result<(), DaytonaError> {
        self.client
            .delete_empty(&format!("/snapshot/{name_or_id}"))
            .await
    }
}
