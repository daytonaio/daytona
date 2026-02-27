// Copyright 2019-2025 Daytona Labs Inc.
// SPDX-License-Identifier: Apache-2.0

use crate::client::ClientInner;
use crate::error::DaytonaError;
use crate::services::code_interpreter::CodeInterpreterService;
use crate::services::computer_use::ComputerUseService;
use crate::services::filesystem::FileSystemService;
use crate::services::git::GitService;
use crate::services::lsp_server::LspServerService;
use crate::services::process::ProcessService;
use crate::services::ServiceClient;
use crate::types::{
    BuildInfo, DirResponse, ResizeSandboxRequest, SandboxBackupState, SandboxDto, SandboxState,
    SshAccessDto, SshAccessValidationDto, VolumeMount,
};
use std::collections::HashMap;
use std::sync::Arc;
use std::time::Duration;
use tokio::time::sleep;

#[derive(Clone)]
pub struct Sandbox {
    inner: Arc<ClientInner>,
    pub id: String,
    pub name: String,
    /// Organization ID that owns the sandbox
    pub organization_id: String,
    /// Daytona snapshot used to create the sandbox
    pub snapshot: Option<String>,
    /// OS user running in the sandbox
    pub user: String,
    /// Sandbox state
    pub state: SandboxState,
    /// Target location where the sandbox runs
    pub target: String,
    /// CPU cores allocated
    pub cpu: i32,
    /// GPU units allocated
    pub gpu: i32,
    /// Memory in GiB
    pub memory: i32,
    /// Disk in GiB
    pub disk: i32,
    /// Environment variables set in the sandbox
    pub env: Option<HashMap<String, String>>,
    /// Custom labels attached to the sandbox
    pub labels: Option<HashMap<String, String>>,
    /// Whether the sandbox is publicly accessible
    pub public: bool,
    /// Auto-stop interval in minutes
    pub auto_stop_interval: Option<i32>,
    /// Auto-archive interval in minutes
    pub auto_archive_interval: Option<i32>,
    /// Auto-delete interval in minutes
    pub auto_delete_interval: Option<i32>,
    /// Whether to block all network access
    pub network_block_all: bool,
    /// Allowed network addresses (CIDR notation)
    pub network_allow_list: Option<String>,
    /// Error reason if sandbox is in error state
    pub error_reason: Option<String>,
    /// Whether the error is recoverable
    pub recoverable: Option<bool>,
    /// Current backup state
    pub backup_state: SandboxBackupState,
    /// When the backup was created
    pub backup_created_at: Option<String>,
    /// Volumes attached to the sandbox
    pub volumes: Option<Vec<VolumeMount>>,
    /// Build information for the sandbox
    pub build_info: Option<BuildInfo>,
    /// When the sandbox was created
    pub created_at: Option<String>,
    /// When the sandbox was last updated
    pub updated_at: Option<String>,
}

impl Sandbox {
    pub(crate) fn from_dto(inner: Arc<ClientInner>, dto: SandboxDto) -> Self {
        Self {
            inner,
            id: dto.id,
            name: dto.name,
            organization_id: dto.organization_id,
            snapshot: dto.snapshot,
            user: dto.user.unwrap_or_default(),
            state: dto.state,
            target: dto.target,
            cpu: dto.cpu.unwrap_or(2),
            gpu: dto.gpu.unwrap_or(0),
            memory: dto.memory.unwrap_or(4),
            disk: dto.disk.unwrap_or(10),
            env: dto.env,
            labels: dto.labels,
            public: dto.public.unwrap_or(false),
            auto_stop_interval: dto.auto_stop_interval,
            auto_archive_interval: dto.auto_archive_interval,
            auto_delete_interval: dto.auto_delete_interval,
            network_block_all: dto.network_block_all,
            network_allow_list: dto.network_allow_list,
            error_reason: dto.error_reason,
            recoverable: dto.recoverable,
            backup_state: dto.backup_state,
            backup_created_at: dto.backup_created_at,
            volumes: dto.volumes,
            build_info: dto.build_info,
            created_at: dto.created_at,
            updated_at: dto.updated_at,
        }
    }

    async fn service_client(&self) -> Result<ServiceClient, DaytonaError> {
        let base = self.inner.toolbox_base_url(&self.id, &self.target).await?;
        Ok(ServiceClient::new(
            base,
            self.inner.config.clone(),
            self.inner.http.clone(),
        ))
    }

    pub async fn filesystem(&self) -> Result<FileSystemService, DaytonaError> {
        Ok(FileSystemService::new(self.service_client().await?))
    }

    pub async fn git(&self) -> Result<GitService, DaytonaError> {
        Ok(GitService::new(self.service_client().await?))
    }

    pub async fn process(&self) -> Result<ProcessService, DaytonaError> {
        Ok(ProcessService::new(self.service_client().await?))
    }

    pub async fn code_interpreter(&self) -> Result<CodeInterpreterService, DaytonaError> {
        Ok(CodeInterpreterService::new(self.service_client().await?))
    }

    pub async fn computer_use(&self) -> Result<ComputerUseService, DaytonaError> {
        Ok(ComputerUseService::new(self.service_client().await?))
    }

    pub async fn lsp_server(&self) -> Result<LspServerService, DaytonaError> {
        Ok(LspServerService::new(self.service_client().await?))
    }

    /// Create a new Language Server Protocol (LSP) server instance.
    ///
    /// The LSP server provides language-specific features like code completion,
    /// diagnostics, and more.
    ///
    /// # Arguments
    /// * `language_id` - The language server type (e.g., "typescript", "python")
    /// * `path_to_project` - Path to the project root directory. Relative paths are resolved based on the sandbox working directory.
    ///
    /// # Returns
    /// A new `LspServerService` instance configured for the specified language
    pub async fn create_lsp_server(
        &self,
        language_id: &str,
        path_to_project: &str,
    ) -> Result<LspServerService, DaytonaError> {
        let service = LspServerService::new(self.service_client().await?);
        service.start(language_id, path_to_project).await?;
        Ok(service)
    }

    /// CamelCase alias for create_lsp_server
    #[allow(non_snake_case)]
    pub async fn createLspServer(
        &self,
        language_id: &str,
        path_to_project: &str,
    ) -> Result<LspServerService, DaytonaError> {
        self.create_lsp_server(language_id, path_to_project).await
    }

    pub async fn start(&self) -> Result<(), DaytonaError> {
        self.inner.start_sandbox(&self.id).await
    }

    pub async fn stop(&self) -> Result<(), DaytonaError> {
        self.inner.stop_sandbox(&self.id).await
    }

    pub async fn delete(&self) -> Result<(), DaytonaError> {
        self.inner.delete_sandbox(&self.id).await
    }

    pub async fn refresh_data(&mut self) -> Result<(), DaytonaError> {
        let dto = self.inner.get_sandbox(&self.id).await?;
        self.state = dto.state;
        self.name = dto.name;
        self.organization_id = dto.organization_id;
        self.snapshot = dto.snapshot;
        self.user = dto.user.unwrap_or_default();
        self.cpu = dto.cpu.unwrap_or(2);
        self.gpu = dto.gpu.unwrap_or(0);
        self.memory = dto.memory.unwrap_or(4);
        self.disk = dto.disk.unwrap_or(10);
        self.labels = dto.labels;
        self.public = dto.public.unwrap_or(false);
        self.auto_stop_interval = dto.auto_stop_interval;
        self.auto_archive_interval = dto.auto_archive_interval;
        self.auto_delete_interval = dto.auto_delete_interval;
        self.network_block_all = dto.network_block_all;
        self.network_allow_list = dto.network_allow_list;
        self.env = dto.env;
        self.error_reason = dto.error_reason;
        self.recoverable = dto.recoverable;
        self.backup_state = dto.backup_state;
        self.backup_created_at = dto.backup_created_at;
        self.volumes = dto.volumes;
        self.build_info = dto.build_info;
        self.created_at = dto.created_at;
        self.updated_at = dto.updated_at;
        Ok(())
    }

    pub async fn wait_for_start(&mut self, timeout: Duration) -> Result<(), DaytonaError> {
        let start = std::time::Instant::now();
        let check_interval = Duration::from_secs(1);

        loop {
            self.refresh_data().await?;

            if self.state == SandboxState::Started {
                return Ok(());
            }

            if self.state == SandboxState::Error {
                return Err(DaytonaError::Api {
                    status: 0,
                    message: "Sandbox failed to start".to_string(),
                });
            }

            if timeout > Duration::ZERO && start.elapsed() > timeout {
                return Err(DaytonaError::Timeout(
                    "Sandbox did not start within timeout".to_string(),
                ));
            }

            sleep(check_interval).await;
        }
    }

    pub async fn wait_for_stop(&mut self, timeout: Duration) -> Result<(), DaytonaError> {
        let start = std::time::Instant::now();
        let check_interval = Duration::from_secs(1);

        loop {
            self.refresh_data().await?;

            if self.state == SandboxState::Stopped {
                return Ok(());
            }

            if timeout > Duration::ZERO && start.elapsed() > timeout {
                return Err(DaytonaError::Timeout(
                    "Sandbox did not stop within timeout".to_string(),
                ));
            }

            sleep(check_interval).await;
        }
    }

    pub async fn archive(&mut self) -> Result<(), DaytonaError> {
        self.inner.archive_sandbox(&self.id).await?;
        self.refresh_data().await?;
        Ok(())
    }

    pub async fn resize(
        &mut self,
        resources: &ResizeSandboxRequest,
        timeout: Duration,
    ) -> Result<(), DaytonaError> {
        let start = std::time::Instant::now();
        self.inner.resize_sandbox(&self.id, resources).await?;
        let elapsed = start.elapsed();
        let remaining = if timeout > Duration::ZERO {
            timeout.saturating_sub(elapsed)
        } else {
            Duration::ZERO
        };
        self.wait_for_resize(remaining).await?;
        Ok(())
    }

    pub async fn wait_for_resize(&mut self, timeout: Duration) -> Result<(), DaytonaError> {
        let start = std::time::Instant::now();
        let check_interval = Duration::from_secs(1);

        loop {
            self.refresh_data().await?;

            if self.state != SandboxState::Resizing {
                return Ok(());
            }

            if self.state == SandboxState::Error {
                return Err(DaytonaError::Api {
                    status: 0,
                    message: "Sandbox resize failed".to_string(),
                });
            }

            if timeout > Duration::ZERO && start.elapsed() > timeout {
                return Err(DaytonaError::Timeout(
                    "Sandbox resize did not complete within timeout".to_string(),
                ));
            }

            sleep(check_interval).await;
        }
    }

    pub async fn set_labels(
        &mut self,
        labels: HashMap<String, String>,
    ) -> Result<HashMap<String, String>, DaytonaError> {
        self.inner.set_labels(&self.id, labels).await?;
        self.refresh_data().await?;
        Ok(self.labels.clone().unwrap_or_default())
    }

    pub async fn get_user_home_dir(&self) -> Result<Option<String>, DaytonaError> {
        let client = self.service_client().await?;
        let resp: DirResponse = client.get_json("/info/user-home-dir").await?;
        Ok(resp.dir)
    }

    pub async fn get_working_dir(&self) -> Result<Option<String>, DaytonaError> {
        let client = self.service_client().await?;
        let resp: DirResponse = client.get_json("/info/work-dir").await?;
        Ok(resp.dir)
    }

    /// @deprecated Use `get_user_home_dir` instead. This method will be removed in a future version.
    pub async fn get_user_root_dir(&self) -> Result<Option<String>, DaytonaError> {
        self.get_user_home_dir().await
    }

    // CamelCase aliases
    #[allow(non_snake_case)]
    pub async fn getUserHomeDir(&self) -> Result<Option<String>, DaytonaError> {
        self.get_user_home_dir().await
    }

    #[allow(non_snake_case)]
    pub async fn getWorkingDir(&self) -> Result<Option<String>, DaytonaError> {
        self.get_working_dir().await
    }

    /// @deprecated Use `getUserHomeDir` instead.
    #[allow(non_snake_case)]
    pub async fn getUserRootDir(&self) -> Result<Option<String>, DaytonaError> {
        self.get_user_home_dir().await
    }

    pub async fn set_auto_archive_interval(&mut self, interval: i32) -> Result<(), DaytonaError> {
        self.inner
            .set_auto_archive_interval(&self.id, interval)
            .await?;
        self.auto_archive_interval = Some(interval);
        Ok(())
    }

    pub async fn set_auto_delete_interval(&mut self, interval: i32) -> Result<(), DaytonaError> {
        self.inner
            .set_auto_delete_interval(&self.id, interval)
            .await?;
        self.auto_delete_interval = Some(interval);
        Ok(())
    }

    pub async fn set_auto_stop_interval(&mut self, interval: i32) -> Result<(), DaytonaError> {
        self.inner
            .set_auto_stop_interval(&self.id, interval)
            .await?;
        self.auto_stop_interval = Some(interval);
        Ok(())
    }

    pub async fn recover(&mut self, timeout: Option<Duration>) -> Result<(), DaytonaError> {
        let timeout = timeout.unwrap_or(Duration::from_secs(60));

        self.inner.recover(&self.id).await?;

        // Wait for the sandbox to start
        self.wait_for_start(timeout).await?;

        // Refresh data to get updated state
        self.refresh_data().await?;

        Ok(())
    }

    pub async fn refresh_activity(&self) -> Result<(), DaytonaError> {
        self.inner.refresh_activity(&self.id).await
    }

    pub async fn get_signed_preview_url(
        &self,
        port: i32,
        expires_in_seconds: Option<i32>,
    ) -> Result<crate::types::SignedPreviewUrl, DaytonaError> {
        self.inner
            .get_signed_preview_url(&self.id, port, expires_in_seconds)
            .await
    }

    pub async fn expire_signed_preview_url(
        &self,
        port: i32,
        token: &str,
    ) -> Result<(), DaytonaError> {
        self.inner
            .expire_signed_preview_url(&self.id, port, token)
            .await
    }

    // CamelCase aliases
    #[allow(non_snake_case)]
    pub async fn setAutoArchiveInterval(&mut self, interval: i32) -> Result<(), DaytonaError> {
        self.set_auto_archive_interval(interval).await
    }

    #[allow(non_snake_case)]
    pub async fn setAutoDeleteInterval(&mut self, interval: i32) -> Result<(), DaytonaError> {
        self.set_auto_delete_interval(interval).await
    }

    pub async fn create_ssh_access(
        &self,
        expires_in_minutes: Option<i32>,
    ) -> Result<SshAccessDto, DaytonaError> {
        self.inner
            .create_ssh_access(&self.id, expires_in_minutes)
            .await
    }

    pub async fn revoke_ssh_access(&self, token: &str) -> Result<(), DaytonaError> {
        self.inner.revoke_ssh_access(&self.id, token).await
    }

    pub async fn validate_ssh_access(
        &self,
        token: &str,
    ) -> Result<SshAccessValidationDto, DaytonaError> {
        self.inner.validate_ssh_access(token).await
    }

    pub async fn get_preview_link(
        &self,
        port: i32,
    ) -> Result<crate::types::PreviewLink, DaytonaError> {
        self.inner.get_preview_link(&self.id, port).await
    }

    // CamelCase alias
    #[allow(non_snake_case)]
    pub async fn getPreviewLink(
        &self,
        port: i32,
    ) -> Result<crate::types::PreviewLink, DaytonaError> {
        self.get_preview_link(port).await
    }

    #[allow(non_snake_case)]
    pub async fn setAutoStopInterval(&mut self, interval: i32) -> Result<(), DaytonaError> {
        self.set_auto_stop_interval(interval).await
    }

    #[allow(non_snake_case)]
    pub async fn refreshActivity(&self) -> Result<(), DaytonaError> {
        self.refresh_activity().await
    }

    #[allow(non_snake_case)]
    pub async fn getSignedPreviewUrl(
        &self,
        port: i32,
        expires_in_seconds: Option<i32>,
    ) -> Result<crate::types::SignedPreviewUrl, DaytonaError> {
        self.get_signed_preview_url(port, expires_in_seconds).await
    }

    #[allow(non_snake_case)]
    pub async fn expireSignedPreviewUrl(&self, port: i32, token: &str) -> Result<(), DaytonaError> {
        self.expire_signed_preview_url(port, token).await
    }
}
