// Copyright 2019-2025 Daytona Labs Inc.
// SPDX-License-Identifier: Apache-2.0

use crate::error::DaytonaError;
use crate::services::ServiceClient;
use crate::types::{FileDownloadResponse, FileInfo, FindMatch, ReplaceResult, SearchFilesResponse};
use serde_json::json;

#[derive(Clone)]
pub struct FileSystemService {
    client: ServiceClient,
}

impl FileSystemService {
    pub fn new(client: ServiceClient) -> Self {
        Self { client }
    }

    pub async fn create_folder(&self, path: &str, mode: &str) -> Result<(), DaytonaError> {
        let url = format!(
            "/files/folder?path={}&mode={}",
            urlencoding::encode(path),
            urlencoding::encode(mode)
        );
        self.client.post_empty_no_body(&url).await
    }

    pub async fn delete_file(
        &self,
        path: &str,
        recursive: Option<bool>,
    ) -> Result<(), DaytonaError> {
        let url = format!(
            "/files?path={}&recursive={}",
            urlencoding::encode(path),
            recursive.unwrap_or(false)
        );
        self.client.delete_empty(&url).await
    }

    pub async fn list_files(&self, path: &str) -> Result<Vec<FileInfo>, DaytonaError> {
        self.client
            .get_json(&format!("/files?path={}", urlencoding::encode(path)))
            .await
    }

    pub async fn get_file_info(&self, path: &str) -> Result<FileInfo, DaytonaError> {
        self.client
            .get_json(&format!("/files/info?path={}", urlencoding::encode(path)))
            .await
    }

    pub async fn upload_file(&self, source: &str, destination: &str) -> Result<(), DaytonaError> {
        self.client
            .post_empty(
                "/files/upload",
                &json!({ "source": source, "destination": destination }),
            )
            .await
    }

    pub async fn download_file(
        &self,
        source: &str,
        destination: Option<&str>,
    ) -> Result<FileDownloadResponse, DaytonaError> {
        let mut url = format!("/files/download?path={}", urlencoding::encode(source));
        if let Some(dest) = destination {
            url.push_str(&format!("&destination={}", urlencoding::encode(dest)));
        }
        self.client.get_json(&url).await
    }

    pub async fn search_files(
        &self,
        path: &str,
        pattern: &str,
    ) -> Result<SearchFilesResponse, DaytonaError> {
        self.client
            .post_json(
                "/files/search",
                &json!({ "path": path, "pattern": pattern }),
            )
            .await
    }

    pub async fn replace_in_files(
        &self,
        files: &[String],
        pattern: &str,
        new_value: &str,
    ) -> Result<Vec<ReplaceResult>, DaytonaError> {
        self.client
            .post_json(
                "/files/replace",
                &json!({ "files": files, "pattern": pattern, "newValue": new_value }),
            )
            .await
    }

    pub async fn move_files(&self, source: &str, destination: &str) -> Result<(), DaytonaError> {
        self.client
            .post_empty(
                "/files/move",
                &json!({ "source": source, "destination": destination }),
            )
            .await
    }

    pub async fn set_file_permissions(
        &self,
        path: &str,
        mode: Option<&str>,
        owner: Option<&str>,
        group: Option<&str>,
    ) -> Result<(), DaytonaError> {
        let mut body = json!({ "path": path });
        if let Some(m) = mode {
            body["mode"] = json!(m);
        }
        if let Some(o) = owner {
            body["owner"] = json!(o);
        }
        if let Some(g) = group {
            body["group"] = json!(g);
        }
        self.client.post_empty("/files/permissions", &body).await
    }

    pub async fn find_files(
        &self,
        path: &str,
        pattern: &str,
    ) -> Result<Vec<FindMatch>, DaytonaError> {
        self.client
            .post_json("/files/find", &json!({ "path": path, "pattern": pattern }))
            .await
    }

    #[allow(non_snake_case)]
    pub async fn createFolder(&self, path: &str, mode: &str) -> Result<(), DaytonaError> {
        self.create_folder(path, mode).await
    }

    #[allow(non_snake_case)]
    pub async fn deleteFile(
        &self,
        path: &str,
        recursive: Option<bool>,
    ) -> Result<(), DaytonaError> {
        self.delete_file(path, recursive).await
    }

    #[allow(non_snake_case)]
    pub async fn listFiles(&self, path: &str) -> Result<Vec<FileInfo>, DaytonaError> {
        self.list_files(path).await
    }

    #[allow(non_snake_case)]
    pub async fn getFileInfo(&self, path: &str) -> Result<FileInfo, DaytonaError> {
        self.get_file_info(path).await
    }

    #[allow(non_snake_case)]
    pub async fn uploadFile(&self, source: &str, destination: &str) -> Result<(), DaytonaError> {
        self.upload_file(source, destination).await
    }

    #[allow(non_snake_case)]
    pub async fn downloadFile(
        &self,
        source: &str,
        destination: Option<&str>,
    ) -> Result<FileDownloadResponse, DaytonaError> {
        self.download_file(source, destination).await
    }

    #[allow(non_snake_case)]
    pub async fn searchFiles(
        &self,
        path: &str,
        pattern: &str,
    ) -> Result<SearchFilesResponse, DaytonaError> {
        self.search_files(path, pattern).await
    }

    #[allow(non_snake_case)]
    pub async fn replaceInFiles(
        &self,
        files: &[String],
        pattern: &str,
        new_value: &str,
    ) -> Result<Vec<ReplaceResult>, DaytonaError> {
        self.replace_in_files(files, pattern, new_value).await
    }

    #[allow(non_snake_case)]
    pub async fn moveFiles(&self, source: &str, destination: &str) -> Result<(), DaytonaError> {
        self.move_files(source, destination).await
    }

    #[allow(non_snake_case)]
    pub async fn setFilePermissions(
        &self,
        path: &str,
        mode: Option<&str>,
        owner: Option<&str>,
        group: Option<&str>,
    ) -> Result<(), DaytonaError> {
        self.set_file_permissions(path, mode, owner, group).await
    }

    #[allow(non_snake_case)]
    pub async fn findFiles(
        &self,
        path: &str,
        pattern: &str,
    ) -> Result<Vec<FindMatch>, DaytonaError> {
        self.find_files(path, pattern).await
    }
}
