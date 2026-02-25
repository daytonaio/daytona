// Copyright 2019-2025 Daytona Labs Inc.
// SPDX-License-Identifier: Apache-2.0

use crate::error::DaytonaError;
use crate::services::ServiceClient;
use serde_json::Value;

#[derive(Clone)]
pub struct ObjectStorageService {
    client: ServiceClient,
}

impl ObjectStorageService {
    pub fn new(client: ServiceClient) -> Self {
        Self { client }
    }

    pub async fn get_push_access(&self) -> Result<Value, DaytonaError> {
        self.client.get_json("/object-storage/push-access").await
    }
}
