// Copyright 2019-2025 Daytona Labs Inc.
// SPDX-License-Identifier: Apache-2.0

use crate::error::DaytonaError;
use crate::services::ServiceClient;
use crate::types::{BranchesResponse, GitCommitResponse, GitStatus};
use serde_json::json;

#[derive(Clone)]
pub struct GitService {
    client: ServiceClient,
}

impl GitService {
    pub fn new(client: ServiceClient) -> Self {
        Self { client }
    }

    pub async fn clone(
        &self,
        url: &str,
        path: &str,
        branch: Option<&str>,
    ) -> Result<(), DaytonaError> {
        self.client
            .post_empty(
                "/git/clone",
                &json!({ "url": url, "path": path, "branch": branch }),
            )
            .await
    }

    pub async fn add(&self, path: &str, files: &[String]) -> Result<(), DaytonaError> {
        self.client
            .post_empty("/git/add", &json!({ "path": path, "files": files }))
            .await
    }

    pub async fn commit(
        &self,
        path: &str,
        message: &str,
        author: &str,
        email: &str,
        allow_empty: Option<bool>,
    ) -> Result<GitCommitResponse, DaytonaError> {
        self.client
            .post_json(
                "/git/commit",
                &json!({
                    "path": path,
                    "message": message,
                    "author": author,
                    "email": email,
                    "allowEmpty": allow_empty
                }),
            )
            .await
    }

    pub async fn push(
        &self,
        path: &str,
        username: Option<&str>,
        password: Option<&str>,
    ) -> Result<(), DaytonaError> {
        self.client
            .post_empty(
                "/git/push",
                &json!({ "path": path, "username": username, "password": password }),
            )
            .await
    }

    pub async fn pull(
        &self,
        path: &str,
        username: Option<&str>,
        password: Option<&str>,
    ) -> Result<(), DaytonaError> {
        self.client
            .post_empty(
                "/git/pull",
                &json!({ "path": path, "username": username, "password": password }),
            )
            .await
    }

    pub async fn status(&self, path: &str) -> Result<GitStatus, DaytonaError> {
        self.client
            .get_json(&format!("/git/status?path={}", urlencoding::encode(path)))
            .await
    }

    pub async fn branches(&self, path: &str) -> Result<BranchesResponse, DaytonaError> {
        self.client
            .get_json(&format!("/git/branches?path={}", urlencoding::encode(path)))
            .await
    }
    pub async fn checkout(&self, path: &str, branch: &str) -> Result<(), DaytonaError> {
        self.client
            .post_empty("/git/checkout", &json!({ "path": path, "branch": branch }))
            .await
    }

    pub async fn delete_branch(&self, path: &str, branch: &str) -> Result<(), DaytonaError> {
        self.client
            .post_empty(
                "/git/delete-branch",
                &json!({ "path": path, "branch": branch }),
            )
            .await
    }

    // CamelCase aliases
    // Note: 'Clone' alias is not provided because 'clone' conflicts with the Clone trait

    #[allow(non_snake_case)]
    pub async fn Add(&self, path: &str, files: &[String]) -> Result<(), DaytonaError> {
        self.add(path, files).await
    }

    #[allow(non_snake_case)]
    pub async fn Commit(
        &self,
        path: &str,
        message: &str,
        author: &str,
        email: &str,
        allow_empty: Option<bool>,
    ) -> Result<GitCommitResponse, DaytonaError> {
        self.commit(path, message, author, email, allow_empty).await
    }

    #[allow(non_snake_case)]
    pub async fn Push(
        &self,
        path: &str,
        username: Option<&str>,
        password: Option<&str>,
    ) -> Result<(), DaytonaError> {
        self.push(path, username, password).await
    }

    #[allow(non_snake_case)]
    pub async fn Pull(
        &self,
        path: &str,
        username: Option<&str>,
        password: Option<&str>,
    ) -> Result<(), DaytonaError> {
        self.pull(path, username, password).await
    }

    #[allow(non_snake_case)]
    pub async fn Status(&self, path: &str) -> Result<GitStatus, DaytonaError> {
        self.status(path).await
    }

    #[allow(non_snake_case)]
    pub async fn Branches(&self, path: &str) -> Result<BranchesResponse, DaytonaError> {
        self.branches(path).await
    }

    #[allow(non_snake_case)]
    pub async fn Checkout(&self, path: &str, branch: &str) -> Result<(), DaytonaError> {
        self.checkout(path, branch).await
    }

    #[allow(non_snake_case)]
    pub async fn DeleteBranch(&self, path: &str, branch: &str) -> Result<(), DaytonaError> {
        self.delete_branch(path, branch).await
    }
}
