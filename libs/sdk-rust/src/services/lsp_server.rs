// Copyright 2019-2025 Daytona Labs Inc.
// SPDX-License-Identifier: Apache-2.0

use crate::error::DaytonaError;
use crate::services::ServiceClient;
use crate::types::Position;
use serde_json::{json, Value};

#[derive(Clone)]
pub struct LspServerService {
    client: ServiceClient,
}

impl LspServerService {
    pub fn new(client: ServiceClient) -> Self {
        Self { client }
    }

    pub async fn start(
        &self,
        language_id: &str,
        path_to_project: &str,
    ) -> Result<(), DaytonaError> {
        self.client
            .post_empty(
                "/lsp/start",
                &json!({ "languageId": language_id, "pathToProject": path_to_project }),
            )
            .await
    }

    pub async fn did_open(
        &self,
        language_id: &str,
        path_to_project: &str,
        uri: &str,
    ) -> Result<(), DaytonaError> {
        self.client
            .post_empty(
                "/lsp/did-open",
                &json!({ "languageId": language_id, "pathToProject": path_to_project, "uri": uri }),
            )
            .await
    }

    pub async fn did_close(
        &self,
        language_id: &str,
        path_to_project: &str,
        uri: &str,
    ) -> Result<(), DaytonaError> {
        self.client
            .post_empty(
                "/lsp/did-close",
                &json!({ "languageId": language_id, "pathToProject": path_to_project, "uri": uri }),
            )
            .await
    }

    pub async fn document_symbols(
        &self,
        language_id: &str,
        path_to_project: &str,
        uri: &str,
    ) -> Result<Value, DaytonaError> {
        self.client
            .post_json(
                "/lsp/document-symbols",
                &json!({ "languageId": language_id, "pathToProject": path_to_project, "uri": uri }),
            )
            .await
    }

    pub async fn completions(
        &self,
        language_id: &str,
        path_to_project: &str,
        uri: &str,
        position: Position,
    ) -> Result<Value, DaytonaError> {
        self.client
            .post_json(
                "/lsp/completions",
                &json!({ "languageId": language_id, "pathToProject": path_to_project, "uri": uri, "position": position }),
            )
            .await
    }

    pub async fn definition(
        &self,
        language_id: &str,
        path_to_project: &str,
        uri: &str,
        position: Position,
    ) -> Result<Value, DaytonaError> {
        self.client
            .post_json(
                "/lsp/definition",
                &json!({ "languageId": language_id, "pathToProject": path_to_project, "uri": uri, "position": position }),
            )
            .await
    }

    pub async fn references(
        &self,
        language_id: &str,
        path_to_project: &str,
        uri: &str,
        position: Position,
    ) -> Result<Value, DaytonaError> {
        self.client
            .post_json(
                "/lsp/references",
                &json!({ "languageId": language_id, "pathToProject": path_to_project, "uri": uri, "position": position }),
            )
            .await
    }

    #[allow(non_snake_case)]
    pub async fn didOpen(
        &self,
        language_id: &str,
        path_to_project: &str,
        uri: &str,
    ) -> Result<(), DaytonaError> {
        self.did_open(language_id, path_to_project, uri).await
    }

    #[allow(non_snake_case)]
    pub async fn didClose(
        &self,
        language_id: &str,
        path_to_project: &str,
        uri: &str,
    ) -> Result<(), DaytonaError> {
        self.did_close(language_id, path_to_project, uri).await
    }

    #[allow(non_snake_case)]
    pub async fn documentSymbols(
        &self,
        language_id: &str,
        path_to_project: &str,
        uri: &str,
    ) -> Result<Value, DaytonaError> {
        self.document_symbols(language_id, path_to_project, uri)
            .await
    }

    #[allow(non_snake_case)]
    pub async fn Completions(
        &self,
        language_id: &str,
        path_to_project: &str,
        uri: &str,
        position: Position,
    ) -> Result<Value, DaytonaError> {
        self.completions(language_id, path_to_project, uri, position)
            .await
    }

    #[allow(non_snake_case)]
    pub async fn Definition(
        &self,
        language_id: &str,
        path_to_project: &str,
        uri: &str,
        position: Position,
    ) -> Result<Value, DaytonaError> {
        self.definition(language_id, path_to_project, uri, position)
            .await
    }

    #[allow(non_snake_case)]
    pub async fn References(
        &self,
        language_id: &str,
        path_to_project: &str,
        uri: &str,
        position: Position,
    ) -> Result<Value, DaytonaError> {
        self.references(language_id, path_to_project, uri, position)
            .await
    }
}
