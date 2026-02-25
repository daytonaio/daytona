// Copyright 2019-2025 Daytona Labs Inc.
// SPDX-License-Identifier: Apache-2.0

use crate::error::DaytonaError;
use crate::services::ServiceClient;
use crate::types::Volume;
use serde_json::json;

#[derive(Clone)]
pub struct VolumeService {
    client: ServiceClient,
}

impl VolumeService {
    pub fn new(client: ServiceClient) -> Self {
        Self { client }
    }

    pub async fn list(&self) -> Result<Vec<Volume>, DaytonaError> {
        self.client.get_json("/volume").await
    }

    pub async fn get(&self, name_or_id: &str) -> Result<Volume, DaytonaError> {
        self.client.get_json(&format!("/volume/{name_or_id}")).await
    }

    pub async fn create(&self, name: &str) -> Result<Volume, DaytonaError> {
        self.client
            .post_json("/volume", &json!({ "name": name }))
            .await
    }

    pub async fn delete(&self, name_or_id: &str) -> Result<(), DaytonaError> {
        self.client
            .delete_empty(&format!("/volume/{name_or_id}"))
            .await
    }

    // CamelCase aliases
    #[allow(non_snake_case)]
    pub async fn List(&self) -> Result<Vec<Volume>, DaytonaError> {
        self.list().await
    }

    #[allow(non_snake_case)]
    pub async fn Get(&self, name_or_id: &str) -> Result<Volume, DaytonaError> {
        self.get(name_or_id).await
    }

    #[allow(non_snake_case)]
    pub async fn Create(&self, name: &str) -> Result<Volume, DaytonaError> {
        self.create(name).await
    }

    #[allow(non_snake_case)]
    pub async fn Delete(&self, name_or_id: &str) -> Result<(), DaytonaError> {
        self.delete(name_or_id).await
    }
}
