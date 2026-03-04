// Copyright 2019-2025 Daytona Labs Inc.
// SPDX-License-Identifier: Apache-2.0

use crate::image::DockerImage;
use serde::{Deserialize, Serialize};
use std::collections::HashMap;

#[derive(Debug, Clone, Serialize, Deserialize, PartialEq, Eq, Default)]
#[serde(rename_all = "snake_case")]
pub enum SandboxState {
    Creating,
    Starting,
    Started,
    Stopping,
    Stopped,
    PendingBuild,
    BuildFailed,
    Resizing,
    Error,
    Destroyed,
    #[default]
    Unknown,
}

/// Backup state of a sandbox
#[derive(Debug, Clone, Serialize, Deserialize, PartialEq, Eq, Default)]
pub enum SandboxBackupState {
    #[default]
    Unknown,
    BackingUp,
    Restoring,
    Archiving,
    Archived,
    Error,
}

/// Build information for a sandbox created from a Docker image or custom Dockerfile
#[derive(Debug, Clone, Serialize, Deserialize, Default)]
pub struct BuildInfo {
    /// Dockerfile content for building the sandbox image
    #[serde(rename = "dockerfileContent")]
    pub dockerfile_content: String,
    /// Hashes of context files uploaded to object storage used for building the image
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub context_hashes: Option<Vec<String>>,
}

#[derive(Debug, Clone, Serialize, Deserialize, PartialEq, Eq)]
#[serde(rename_all = "lowercase")]
pub enum CodeLanguage {
    Python,
    JavaScript,
    TypeScript,
}

#[derive(Debug, Clone, Serialize, Deserialize, Default)]
pub struct Resources {
    pub cpu: Option<i32>,
    pub gpu: Option<i32>,
    pub memory: Option<i32>,
    pub disk: Option<i32>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct VolumeMount {
    #[serde(rename = "volumeId")]
    pub volume_id: String,
    #[serde(rename = "mountPath")]
    pub mount_path: String,
    pub subpath: Option<String>,
}

#[derive(Debug, Clone, Serialize, Deserialize, Default)]
pub struct CreateSandboxParams {
    #[serde(skip_serializing_if = "Option::is_none")]
    pub name: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub user: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub language: Option<CodeLanguage>,
    #[serde(rename = "envVars", skip_serializing_if = "Option::is_none")]
    pub env_vars: Option<HashMap<String, String>>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub labels: Option<HashMap<String, String>>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub public: Option<bool>,
    #[serde(rename = "autoStopInterval", skip_serializing_if = "Option::is_none")]
    pub auto_stop_interval: Option<i32>,
    #[serde(
        rename = "autoArchiveInterval",
        skip_serializing_if = "Option::is_none"
    )]
    pub auto_archive_interval: Option<i32>,
    #[serde(rename = "autoDeleteInterval", skip_serializing_if = "Option::is_none")]
    pub auto_delete_interval: Option<i32>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub volumes: Option<Vec<VolumeMount>>,
    #[serde(rename = "networkBlockAll", skip_serializing_if = "Option::is_none")]
    pub network_block_all: Option<bool>,
    #[serde(rename = "networkAllowList", skip_serializing_if = "Option::is_none")]
    pub network_allow_list: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub ephemeral: Option<bool>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub snapshot: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub image: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub resources: Option<Resources>,
    /// The target (region) where the sandbox will be created
    #[serde(skip_serializing_if = "Option::is_none")]
    pub target: Option<String>,
    /// Build information for the sandbox
    #[serde(rename = "buildInfo", skip_serializing_if = "Option::is_none")]
    pub build_info: Option<BuildInfo>,
    /// Docker image builder for custom Dockerfile definitions.
    /// When set, the image will be built from the Dockerfile content.
    /// This field is not serialized directly; it is processed by the client
    /// to generate build info before sending to the API.
    #[serde(skip)]
    pub docker_image: Option<DockerImage>,
}

#[derive(Debug, Clone, Serialize, Deserialize, Default)]
pub struct SandboxDto {
    pub id: String,
    pub name: String,
    #[serde(default)]
    pub state: SandboxState,
    #[serde(default)]
    pub target: String,
    /// Organization ID that owns the sandbox
    #[serde(rename = "organizationId")]
    pub organization_id: String,
    /// Daytona snapshot used to create the sandbox
    pub snapshot: Option<String>,
    /// OS user running in the sandbox
    pub user: Option<String>,
    /// Environment variables set in the sandbox
    pub env: Option<HashMap<String, String>>,
    pub labels: Option<HashMap<String, String>>,
    pub public: Option<bool>,
    #[serde(rename = "autoStopInterval")]
    pub auto_stop_interval: Option<i32>,
    #[serde(rename = "autoArchiveInterval")]
    pub auto_archive_interval: Option<i32>,
    #[serde(rename = "autoDeleteInterval")]
    pub auto_delete_interval: Option<i32>,
    /// Volumes attached to the sandbox
    pub volumes: Option<Vec<VolumeMount>>,
    #[serde(rename = "networkBlockAll")]
    pub network_block_all: bool,
    #[serde(rename = "networkAllowList")]
    pub network_allow_list: Option<String>,
    /// Error reason if sandbox is in error state
    #[serde(rename = "errorReason")]
    pub error_reason: Option<String>,
    /// Whether the error is recoverable
    pub recoverable: Option<bool>,
    /// CPU cores allocated
    pub cpu: Option<i32>,
    /// GPU units allocated
    pub gpu: Option<i32>,
    pub memory: Option<i32>,
    /// Disk space in GiB
    pub disk: Option<i32>,
    /// Current backup state
    #[serde(default)]
    pub backup_state: SandboxBackupState,
    /// When the backup was created
    #[serde(rename = "backupCreatedAt")]
    pub backup_created_at: Option<String>,
    /// Build information for the sandbox
    pub build_info: Option<BuildInfo>,
    /// When the sandbox was created
    #[serde(rename = "createdAt")]
    pub created_at: Option<String>,
    /// When the sandbox was last updated
    #[serde(rename = "updatedAt")]
    pub updated_at: Option<String>,
}

#[derive(Debug, Clone, Serialize, Deserialize, Default)]
pub struct PaginatedSandboxes {
    pub items: Vec<SandboxDto>,
    pub total: i32,
    pub page: i32,
    #[serde(rename = "totalPages")]
    pub total_pages: i32,
}

#[derive(Debug, Clone, Serialize, Deserialize, Default)]
pub struct ExecuteResponse {
    #[serde(rename = "exitCode")]
    pub exit_code: i32,
    pub result: String,
    pub artifacts: Option<ExecutionArtifacts>,
}

#[derive(Debug, Clone, Serialize, Deserialize, Default)]
pub struct ExecutionArtifacts {
    pub stdout: String,
    pub charts: Vec<Chart>,
}

#[derive(Debug, Clone, Serialize, Deserialize, Default)]
pub struct ExecutionResult {
    pub stdout: String,
    pub stderr: String,
    pub charts: Vec<Chart>,
    pub error: Option<ExecutionError>,
}

#[derive(Debug, Clone, Serialize, Deserialize, Default)]
pub struct ExecutionError {
    pub name: String,
    pub value: String,
    pub traceback: Option<String>,
}

#[derive(Debug, Clone, Serialize, Deserialize, Default)]
pub struct Chart {
    #[serde(rename = "type")]
    pub chart_type: ChartType,
    pub title: Option<String>,
    pub elements: Option<serde_json::Value>,
    pub png: Option<String>,
}

#[derive(Debug, Clone, Serialize, Deserialize, Default)]
pub struct FileInfo {
    pub name: String,
    pub size: i64,
    #[serde(default)]
    pub mode: String,
    #[serde(rename = "isDir", default)]
    pub is_directory: bool,
    #[serde(rename = "modTime")]
    pub modified_time: Option<String>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct FileUpload {
    pub source: String,
    pub destination: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct FileDownloadRequest {
    pub source: String,
    pub destination: Option<String>,
}

#[derive(Debug, Clone, Serialize, Deserialize, Default)]
pub struct FileDownloadResponse {
    pub source: String,
    pub result: Option<String>,
    pub error: Option<String>,
}

#[derive(Debug, Clone, Serialize, Deserialize, Default)]
pub struct SearchFilesResponse {
    pub files: Vec<String>,
}

#[derive(Debug, Clone, Serialize, Deserialize, Default)]
pub struct ReplaceResult {
    pub file: String,
    pub replaced: i32,
}

#[derive(Debug, Clone, Serialize, Deserialize, Default)]
pub struct GitFileStatus {
    pub path: String,
    pub status: String,
}

#[derive(Debug, Clone, Serialize, Deserialize, Default)]
pub struct GitStatus {
    #[serde(rename = "currentBranch")]
    pub current_branch: String,
    pub ahead: i32,
    pub behind: i32,
    #[serde(rename = "branchPublished")]
    pub branch_published: bool,
    #[serde(rename = "fileStatus")]
    pub file_status: Vec<GitFileStatus>,
}

#[derive(Debug, Clone, Serialize, Deserialize, Default)]
pub struct GitCommitResponse {
    pub sha: String,
}

#[derive(Debug, Clone, Serialize, Deserialize, Default)]
pub struct BranchesResponse {
    pub branches: Vec<String>,
}

#[derive(Debug, Clone, Serialize, Deserialize, Default)]
pub struct PtySessionInfo {
    #[serde(rename = "sessionId")]
    pub session_id: String,
    pub active: bool,
    pub cwd: Option<String>,
    pub cols: Option<i32>,
    pub rows: Option<i32>,
}

#[derive(Debug, Clone, Serialize, Deserialize, Default)]
pub struct PtyCreateResponse {
    #[serde(rename = "sessionId")]
    pub session_id: String,
}

#[derive(Debug, Clone, Serialize, Deserialize, Default)]
pub struct InterpreterContext {
    pub id: String,
    pub language: Option<String>,
    pub cwd: Option<String>,
}

#[derive(Debug, Clone, Serialize, Deserialize, Default)]
pub struct DisplayInfo {
    #[serde(flatten)]
    pub data: serde_json::Value,
}

#[derive(Debug, Clone, Serialize, Deserialize, Default)]
pub struct WindowInfo {
    pub id: Option<String>,
    pub title: Option<String>,
}

#[derive(Debug, Clone, Serialize, Deserialize, Default)]
pub struct Recording {
    pub id: Option<String>,
    #[serde(flatten)]
    pub data: serde_json::Value,
}

#[derive(Debug, Clone, Serialize, Deserialize, Default)]
pub struct Volume {
    pub id: String,
    pub name: String,
    pub state: Option<String>,
}

#[derive(Debug, Clone, Serialize, Deserialize, Default)]
pub struct Snapshot {
    pub id: String,
    pub name: String,
    pub state: Option<String>,
}

#[derive(Debug, Clone, Serialize, Deserialize, Default)]
pub struct PaginatedSnapshots {
    pub items: Vec<Snapshot>,
    pub total: i32,
    pub page: i32,
    #[serde(rename = "totalPages")]
    pub total_pages: i32,
}

#[derive(Debug, Clone, Serialize, Deserialize, Default)]
pub struct CreateSnapshotParams {
    pub name: String,
    pub image: Option<String>,
    pub resources: Option<Resources>,
    pub entrypoint: Option<Vec<String>>,
    #[serde(rename = "skipValidation")]
    pub skip_validation: Option<bool>,
    /// Docker image builder for custom Dockerfile definitions.
    /// When set, the snapshot will be built from the Dockerfile content.
    /// This field is not serialized directly; it is processed by the client
    /// to generate build info before sending to the API.
    #[serde(skip)]
    pub docker_image: Option<DockerImage>,
}

#[derive(Debug, Clone, Serialize, Deserialize, Default)]
pub struct PreviewLink {
    pub url: String,
    pub token: String,
}

#[derive(Debug, Clone, Serialize, Deserialize, Default)]
pub struct SignedPreviewUrl {
    pub url: String,
    pub token: String,
}

#[derive(Debug, Clone, Serialize, Deserialize, Default)]
pub struct Position {
    pub line: i32,
    pub character: i32,
}

#[derive(Debug, Clone, Serialize, Deserialize, Default)]
pub struct Session {
    pub session_id: String,
    pub commands: Vec<String>,
}

#[derive(Debug, Clone, Serialize, Deserialize, Default)]
pub struct SessionExecuteRequest {
    pub command: String,
    #[serde(rename = "runAsync")]
    pub run_async: bool,
    #[serde(rename = "suppressInputEcho")]
    pub suppress_input_echo: bool,
}

#[derive(Debug, Clone, Serialize, Deserialize, Default)]
pub struct SessionExecuteResponse {
    pub id: String,
    #[serde(rename = "exitCode")]
    pub exit_code: Option<i32>,
    pub stdout: Option<String>,
    pub stderr: Option<String>,
}

#[derive(Debug, Clone, Serialize, Deserialize, Default)]
pub struct Command {
    pub id: String,
    pub command: String,
    #[serde(rename = "exitCode")]
    pub exit_code: Option<i32>,
}

#[derive(Debug, Clone, Serialize, Deserialize, Default)]
pub struct FindMatch {
    pub file: String,
    pub line: i32,
    pub content: String,
}

#[derive(Debug, Clone, Serialize, Deserialize, Default)]
pub struct DirResponse {
    pub dir: Option<String>,
}

#[derive(Debug, Clone, Serialize, Deserialize, Default)]
pub struct SshAccessDto {
    pub token: String,
    #[serde(rename = "expiresAt")]
    pub expires_at: String,
    #[serde(rename = "sshCommand")]
    pub ssh_command: String,
    #[serde(rename = "sshHost")]
    pub ssh_host: String,
    #[serde(rename = "sshPort")]
    pub ssh_port: u16,
    #[serde(rename = "sshUser")]
    pub ssh_user: String,
}

#[derive(Debug, Clone, Serialize, Deserialize, Default)]
pub struct SshAccessValidationDto {
    pub valid: bool,
    #[serde(rename = "expiresAt")]
    pub expires_at: Option<String>,
    #[serde(rename = "sandboxId")]
    pub sandbox_id: Option<String>,
}

#[derive(Debug, Clone, Serialize, Deserialize, Default)]
pub struct ResizeSandboxRequest {
    pub cpu: Option<i32>,
    pub memory: Option<i32>,
    pub disk: Option<i32>,
}

#[derive(Debug, Clone, Serialize, Deserialize, Default)]
pub struct SandboxLabels {
    pub labels: HashMap<String, String>,
}

#[derive(Debug, Clone, Serialize, Deserialize, Default)]
pub struct PtySize {
    pub rows: i32,
    pub cols: i32,
}

#[derive(Debug, Clone, Serialize, Deserialize, Default)]
pub struct ScreenshotOptions {
    pub show_cursor: Option<bool>,
    pub format: Option<String>,
    pub quality: Option<i32>,
    pub scale: Option<f64>,
}

#[derive(Debug, Clone, Serialize, Deserialize, Default)]
pub struct ScreenshotRegion {
    pub x: i32,
    pub y: i32,
    pub width: i32,
    pub height: i32,
}

#[derive(Debug, Clone, Serialize, Deserialize, Default)]
pub struct ScreenshotResponse {
    pub image: String,
    pub width: i32,
    pub height: i32,
    pub size_bytes: Option<i32>,
}

#[derive(Debug, Clone, Serialize, Deserialize, Default)]
pub struct OutputMessage {
    #[serde(rename = "type")]
    pub type_: String,
    pub text: String,
    pub name: String,
    pub value: String,
    pub traceback: String,
}

#[derive(Debug, Clone, Serialize, Deserialize, PartialEq, Eq)]
#[serde(rename_all = "lowercase")]
pub enum LspLanguageId {
    Python,
    JavaScript,
    TypeScript,
}

#[derive(Debug, Clone, Serialize, Deserialize, PartialEq, Eq, Default)]
#[serde(rename_all = "snake_case")]
pub enum ChartType {
    Line,
    Scatter,
    Bar,
    Pie,
    BoxAndWhisker,
    CompositeChart,
    #[default]
    Unknown,
}

#[derive(Debug, Clone, Serialize, Deserialize, Default)]
pub struct CodeRunParams {
    pub argv: Option<Vec<String>>,
    pub env: Option<HashMap<String, String>>,
}

#[derive(Debug, Clone, Serialize, Deserialize, Default)]
pub struct PtyResult {
    #[serde(rename = "exitCode")]
    pub exit_code: Option<i32>,
    pub error: Option<String>,
}
