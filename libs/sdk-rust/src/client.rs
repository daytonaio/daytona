// Copyright 2019-2025 Daytona Labs Inc.
// SPDX-License-Identifier: Apache-2.0

use crate::config::Config;
use crate::error::DaytonaError;
use crate::image::DockerImage;
#[cfg(feature = "otel")]
use crate::otel::OtelState;
use crate::sandbox::Sandbox;
use crate::services::object_storage::ObjectStorageService;
use crate::services::snapshot::SnapshotService;
use crate::services::volume::VolumeService;
use crate::services::ServiceClient;
use crate::types::{
    BuildInfo, CreateSandboxParams, PaginatedSandboxes, PreviewLink, ResizeSandboxRequest,
    SandboxDto, SignedPreviewUrl, SshAccessDto, SshAccessValidationDto,
};
use crate::utils::{decode_empty, decode_json, join_url, with_auth};
use reqwest::Client as HttpClient;
use serde_json::{json, Value};
use std::collections::HashMap;
use std::sync::Arc;
use std::time::Duration;
use tokio::sync::RwLock;

#[derive(Clone)]
pub struct Client {
    pub(crate) inner: Arc<ClientInner>,
    volume: VolumeService,
    snapshot: SnapshotService,
    object_storage: ObjectStorageService,
    #[cfg(feature = "otel")]
    otel: Option<Arc<OtelState>>,
}

pub(crate) struct ClientInner {
    pub config: Config,
    pub http: HttpClient,
    toolbox_proxy_cache: RwLock<HashMap<String, String>>,
}

impl Client {
    pub fn new(config: Config) -> Result<Self, DaytonaError> {
        if config.bearer_token().is_none() {
            return Err(DaytonaError::Unauthorized(
                "Authentication required. Please set DAYTONA_API_KEY or DAYTONA_JWT_TOKEN environment variable, or provide api_key or jwt_token in Config".to_string(),
            ));
        }

        if config.jwt_token.is_some() && config.organization_id.is_none() {
            return Err(DaytonaError::Unauthorized(
                "Organization ID is required when using JWT token".to_string(),
            ));
        }

        let http = HttpClient::builder()
            .timeout(config.timeout.unwrap_or(Duration::from_secs(300)))
            .build()
            .map_err(|e| DaytonaError::Network(e.to_string()))?;
        let inner = Arc::new(ClientInner {
            config: config.clone(),
            http: http.clone(),
            toolbox_proxy_cache: RwLock::new(HashMap::new()),
        });

        #[cfg(feature = "otel")]
        let otel = if config.otel_enabled {
            Some(Arc::new(OtelState::new()?))
        } else {
            None
        };

        Ok(Self {
            inner,
            volume: VolumeService::new(ServiceClient::new(
                config.api_url.clone(),
                config.clone(),
                http.clone(),
            )),
            snapshot: SnapshotService::new(ServiceClient::new(
                config.api_url.clone(),
                config.clone(),
                http.clone(),
            )),
            object_storage: ObjectStorageService::new(ServiceClient::new(
                config.api_url.clone(),
                config,
                http,
            )),
            #[cfg(feature = "otel")]
            otel,
        })
    }

    /// Close the client and shutdown OpenTelemetry if enabled
    pub async fn close(&self) {
        #[cfg(feature = "otel")]
        if let Some(ref otel) = self.otel {
            let _ = otel.shutdown().await;
        }
    }

    pub fn volumes(&self) -> &VolumeService {
        &self.volume
    }

    pub fn snapshots(&self) -> &SnapshotService {
        &self.snapshot
    }

    pub fn object_storage(&self) -> &ObjectStorageService {
        &self.object_storage
    }

    pub async fn create(&self, params: CreateSandboxParams) -> Result<Sandbox, DaytonaError> {
        // Handle ephemeral flag
        let mut params = params;
        if params.ephemeral.unwrap_or(false) {
            if let Some(interval) = params.auto_delete_interval {
                if interval != 0 {
                    eprintln!("[Warning] 'ephemeral' and 'autoDeleteInterval' cannot be used together. If ephemeral is true, autoDeleteInterval will be ignored and set to 0.");
                }
            }
            params.auto_delete_interval = Some(0);
        }

        // Set target from config if not already set
        if params.target.is_none() {
            params.target = self.inner.config.target.clone();
        }

        // Convert image string to build_info
        if let Some(image) = &params.image {
            if params.build_info.is_none() {
                params.build_info = Some(BuildInfo {
                    dockerfile_content: format!("FROM {}\n", image),
                    context_hashes: None,
                });
            }
        }

        let req = with_auth(
            self.inner
                .http
                .post(join_url(&self.inner.config.api_url, "/sandbox")),
            &self.inner.config,
        )?;
        let dto: SandboxDto = decode_json(
            req.json(&params)
                .send()
                .await
                .map_err(|e| DaytonaError::Network(e.to_string()))?,
        )
        .await?;
        Ok(Sandbox::from_dto(self.inner.clone(), dto))
    }

    pub async fn get(&self, sandbox_id_or_name: &str) -> Result<Sandbox, DaytonaError> {
        let req = with_auth(
            self.inner.http.get(join_url(
                &self.inner.config.api_url,
                &format!("/sandbox/{sandbox_id_or_name}"),
            )),
            &self.inner.config,
        )?;
        let dto: SandboxDto = decode_json(
            req.send()
                .await
                .map_err(|e| DaytonaError::Network(e.to_string()))?,
        )
        .await?;
        Ok(Sandbox::from_dto(self.inner.clone(), dto))
    }

    pub async fn find_one(
        &self,
        sandbox_id_or_name: Option<&str>,
        labels: Option<HashMap<String, String>>,
    ) -> Result<Sandbox, DaytonaError> {
        if let Some(id_or_name) = sandbox_id_or_name {
            if !id_or_name.is_empty() {
                return self.get(id_or_name).await;
            }
        }

        let result = self.list(labels, Some(1), Some(1)).await?;
        result
            .items
            .into_iter()
            .next()
            .map(|dto| Sandbox::from_dto(self.inner.clone(), dto))
            .ok_or_else(|| {
                DaytonaError::NotFound("No sandbox found for provided filter".to_string())
            })
    }

    pub async fn list(
        &self,
        labels: Option<HashMap<String, String>>,
        page: Option<i32>,
        limit: Option<i32>,
    ) -> Result<PaginatedSandboxes, DaytonaError> {
        // Validate page and limit parameters
        if page.is_some_and(|p| p < 1) {
            return Err(DaytonaError::Api {
                status: 0,
                message: "Page must be >= 1".to_string(),
            });
        }
        if limit.is_some_and(|l| l < 1) {
            return Err(DaytonaError::Api {
                status: 0,
                message: "Limit must be >= 1".to_string(),
            });
        }

        // Build query string
        let mut query = Vec::new();
        if let Some(p) = page {
            query.push(format!("page={}", p));
        }
        if let Some(l) = limit {
            query.push(format!("limit={}", l));
        }
        if let Some(ref lbls) = labels {
            let labels_json = serde_json::to_string(lbls).unwrap_or_default();
            query.push(format!("labels={}", urlencoding::encode(&labels_json)));
        }

        let query_str = if query.is_empty() {
            String::new()
        } else {
            format!("?{}", query.join("&"))
        };

        let req = with_auth(
            self.inner.http.get(join_url(
                &self.inner.config.api_url,
                &format!("/sandbox{}", query_str),
            )),
            &self.inner.config,
        )?;
        decode_json(
            req.send()
                .await
                .map_err(|e| DaytonaError::Network(e.to_string()))?,
        )
        .await
    }

    pub async fn delete(&self, sandbox_id_or_name: &str) -> Result<(), DaytonaError> {
        let req = with_auth(
            self.inner.http.delete(join_url(
                &self.inner.config.api_url,
                &format!("/sandbox/{sandbox_id_or_name}"),
            )),
            &self.inner.config,
        )?;
        decode_empty(
            req.send()
                .await
                .map_err(|e| DaytonaError::Network(e.to_string()))?,
        )
        .await
    }

    pub async fn start(&self, sandbox_id_or_name: &str) -> Result<(), DaytonaError> {
        let req = with_auth(
            self.inner.http.post(join_url(
                &self.inner.config.api_url,
                &format!("/sandbox/{sandbox_id_or_name}/start"),
            )),
            &self.inner.config,
        )?;
        decode_empty(
            req.send()
                .await
                .map_err(|e| DaytonaError::Network(e.to_string()))?,
        )
        .await
    }

    pub async fn stop(&self, sandbox_id_or_name: &str) -> Result<(), DaytonaError> {
        let req = with_auth(
            self.inner.http.post(join_url(
                &self.inner.config.api_url,
                &format!("/sandbox/{sandbox_id_or_name}/stop"),
            )),
            &self.inner.config,
        )?;
        decode_empty(
            req.send()
                .await
                .map_err(|e| DaytonaError::Network(e.to_string()))?,
        )
        .await
    }

    #[allow(non_snake_case)]
    pub async fn findOne(
        &self,
        sandbox_id_or_name: Option<&str>,
        labels: Option<HashMap<String, String>>,
    ) -> Result<Sandbox, DaytonaError> {
        self.find_one(sandbox_id_or_name, labels).await
    }

    #[cfg(feature = "image-builder")]
    #[allow(dead_code)]
    pub(crate) async fn process_image_context(
        &self,
        image: &DockerImage,
    ) -> Result<Vec<String>, DaytonaError> {
        use s3::creds::Credentials;
        use s3::region::Region;
        use s3::Bucket;

        let contexts = image.contexts();
        if contexts.is_empty() {
            return Ok(Vec::new());
        }

        // 1. Get push access credentials from object storage API
        let creds: Value = self.object_storage.get_push_access().await?;

        let storage_url = creds["storageUrl"]
            .as_str()
            .ok_or_else(|| DaytonaError::Api {
                status: 0,
                message: "missing storageUrl in push access response".to_string(),
            })?;
        let access_key = creds["accessKey"]
            .as_str()
            .ok_or_else(|| DaytonaError::Api {
                status: 0,
                message: "missing accessKey in push access response".to_string(),
            })?;
        let secret = creds["secret"].as_str().ok_or_else(|| DaytonaError::Api {
            status: 0,
            message: "missing secret in push access response".to_string(),
        })?;
        let bucket_name = creds["bucket"].as_str().ok_or_else(|| DaytonaError::Api {
            status: 0,
            message: "missing bucket in push access response".to_string(),
        })?;
        let organization_id =
            creds["organizationId"]
                .as_str()
                .ok_or_else(|| DaytonaError::Api {
                    status: 0,
                    message: "missing organizationId in push access response".to_string(),
                })?;

        // 2. Create S3 bucket instance with credentials
        let credentials = Credentials::new(
            Some(access_key),
            Some(secret),
            None, // session_token
            None, // expiration
            None, // provider_name
        )
        .map_err(|e| DaytonaError::Api {
            status: 0,
            message: format!("failed to create S3 credentials: {}", e),
        })?;

        // Parse region from storage URL or default to us-east-1
        let region = if storage_url.contains("amazonaws.com") {
            // Extract region from AWS URL (e.g., s3.us-west-2.amazonaws.com)
            let parts: Vec<&str> = storage_url.split('.').collect();
            if let Some(pos) = parts.iter().position(|&p| p == "s3") {
                if pos + 1 < parts.len() {
                    Region::Custom {
                        region: parts[pos + 1].to_string(),
                        endpoint: storage_url.to_string(),
                    }
                } else {
                    Region::Custom {
                        region: "us-east-1".to_string(),
                        endpoint: storage_url.to_string(),
                    }
                }
            } else {
                Region::Custom {
                    region: "us-east-1".to_string(),
                    endpoint: storage_url.to_string(),
                }
            }
        } else {
            Region::Custom {
                region: "us-east-1".to_string(),
                endpoint: storage_url.to_string(),
            }
        };

        let bucket = Bucket::new(bucket_name, region, credentials)
            .map_err(|e| DaytonaError::Api {
                status: 0,
                message: format!("failed to create S3 bucket: {}", e),
            })?
            .with_path_style();

        // 3. Upload each context to object storage and 4. Calculate SHA256 hash
        let mut context_hashes = Vec::with_capacity(contexts.len());

        for ctx in contexts {
            // Read the file content
            let content = tokio::fs::read(&ctx.source_path).await.map_err(|e| {
                DaytonaError::Network(format!(
                    "failed to read context file {}: {}",
                    ctx.source_path, e
                ))
            })?;

            // Calculate SHA256 hash
            let hash = {
                use sha2::{Digest, Sha256};
                let mut hasher = Sha256::new();
                hasher.update(&content);
                format!("{:x}", hasher.finalize())
            };

            // Upload to object storage using S3 client
            let s3_key = format!("{}/{}/context.tar", organization_id, hash);

            // Check if object already exists
            match bucket.head_object(&s3_key).await {
                Ok((_, 200)) => {
                    // Object already exists, skip upload
                    context_hashes.push(hash);
                    continue;
                }
                _ => {
                    // Object doesn't exist or error, proceed with upload
                }
            }

            let content_type = "application/x-tar";
            let response = bucket
                .put_object_with_content_type(&s3_key, &content, content_type)
                .await
                .map_err(|e| DaytonaError::Api {
                    status: 0,
                    message: format!("failed to upload context to S3: {}", e),
                })?;

            if response.status_code() != 200 {
                return Err(DaytonaError::Api {
                    status: response.status_code(),
                    message: format!(
                        "failed to upload context to object storage: {}",
                        response.status_code()
                    ),
                });
            }

            // 4. Return context hashes
            context_hashes.push(hash);
        }

        Ok(context_hashes)
    }

    #[allow(dead_code)]
    #[cfg(not(feature = "image-builder"))]
    pub(crate) async fn process_image_context(
        &self,
        _image: &DockerImage,
    ) -> Result<Vec<String>, DaytonaError> {
        Err(DaytonaError::Api {
            status: 0,
            message: "image-builder feature is required to process image contexts".to_string(),
        })
    }
}

impl ClientInner {
    pub async fn start_sandbox(&self, sandbox_id_or_name: &str) -> Result<(), DaytonaError> {
        let req = with_auth(
            self.http.post(join_url(
                &self.config.api_url,
                &format!("/sandbox/{sandbox_id_or_name}/start"),
            )),
            &self.config,
        )?;
        decode_empty(
            req.send()
                .await
                .map_err(|e| DaytonaError::Network(e.to_string()))?,
        )
        .await
    }

    pub async fn stop_sandbox(&self, sandbox_id_or_name: &str) -> Result<(), DaytonaError> {
        let req = with_auth(
            self.http.post(join_url(
                &self.config.api_url,
                &format!("/sandbox/{sandbox_id_or_name}/stop"),
            )),
            &self.config,
        )?;
        decode_empty(
            req.send()
                .await
                .map_err(|e| DaytonaError::Network(e.to_string()))?,
        )
        .await
    }

    pub async fn delete_sandbox(&self, sandbox_id_or_name: &str) -> Result<(), DaytonaError> {
        let req = with_auth(
            self.http.delete(join_url(
                &self.config.api_url,
                &format!("/sandbox/{sandbox_id_or_name}"),
            )),
            &self.config,
        )?;
        decode_empty(
            req.send()
                .await
                .map_err(|e| DaytonaError::Network(e.to_string()))?,
        )
        .await
    }

    pub async fn toolbox_base_url(
        &self,
        sandbox_id: &str,
        region: &str,
    ) -> Result<String, DaytonaError> {
        if let Some(found) = self.toolbox_proxy_cache.read().await.get(region).cloned() {
            return Ok(format!("{}/{}", found.trim_end_matches('/'), sandbox_id));
        }

        let req = with_auth(
            self.http.get(join_url(
                &self.config.api_url,
                &format!("/sandbox/{sandbox_id}/toolbox-proxy-url"),
            )),
            &self.config,
        )?;

        let payload: Value = decode_json(
            req.send()
                .await
                .map_err(|e| DaytonaError::Network(e.to_string()))?,
        )
        .await?;
        let proxy = payload
            .get("url")
            .and_then(Value::as_str)
            .ok_or_else(|| DaytonaError::Api {
                status: 0,
                message: "missing toolbox proxy URL".to_string(),
            })?
            .to_string();

        self.toolbox_proxy_cache
            .write()
            .await
            .insert(region.to_string(), proxy.clone());

        Ok(format!("{}/{}", proxy.trim_end_matches('/'), sandbox_id))
    }

    pub async fn get_sandbox(&self, sandbox_id: &str) -> Result<SandboxDto, DaytonaError> {
        let req = with_auth(
            self.http.get(join_url(
                &self.config.api_url,
                &format!("/sandbox/{}", sandbox_id),
            )),
            &self.config,
        )?;
        decode_json(
            req.send()
                .await
                .map_err(|e| DaytonaError::Network(e.to_string()))?,
        )
        .await
    }

    pub async fn archive_sandbox(&self, sandbox_id: &str) -> Result<(), DaytonaError> {
        let req = with_auth(
            self.http.post(join_url(
                &self.config.api_url,
                &format!("/sandbox/{}/archive", sandbox_id),
            )),
            &self.config,
        )?;
        decode_empty(
            req.send()
                .await
                .map_err(|e| DaytonaError::Network(e.to_string()))?,
        )
        .await
    }

    pub async fn resize_sandbox(
        &self,
        sandbox_id: &str,
        resources: &ResizeSandboxRequest,
    ) -> Result<SandboxDto, DaytonaError> {
        let req = with_auth(
            self.http.post(join_url(
                &self.config.api_url,
                &format!("/sandbox/{}/resize", sandbox_id),
            )),
            &self.config,
        )?;
        decode_json(
            req.json(resources)
                .send()
                .await
                .map_err(|e| DaytonaError::Network(e.to_string()))?,
        )
        .await
    }

    pub async fn set_labels(
        &self,
        sandbox_id: &str,
        labels: HashMap<String, String>,
    ) -> Result<(), DaytonaError> {
        let req = with_auth(
            self.http.put(join_url(
                &self.config.api_url,
                &format!("/sandbox/{}/labels", sandbox_id),
            )),
            &self.config,
        )?;
        decode_empty(
            req.json(&json!({ "labels": labels }))
                .send()
                .await
                .map_err(|e| DaytonaError::Network(e.to_string()))?,
        )
        .await
    }

    pub async fn set_auto_stop_interval(
        &self,
        sandbox_id: &str,
        interval: i32,
    ) -> Result<(), DaytonaError> {
        let req = with_auth(
            self.http.post(join_url(
                &self.config.api_url,
                &format!("/sandbox/{}/auto-stop-interval/{}", sandbox_id, interval),
            )),
            &self.config,
        )?;
        decode_empty(
            req.send()
                .await
                .map_err(|e| DaytonaError::Network(e.to_string()))?,
        )
        .await
    }

    pub async fn recover(&self, sandbox_id: &str) -> Result<(), DaytonaError> {
        let req = with_auth(
            self.http.post(join_url(
                &self.config.api_url,
                &format!("/sandbox/{}/recover", sandbox_id),
            )),
            &self.config,
        )?;
        decode_empty(
            req.send()
                .await
                .map_err(|e| DaytonaError::Network(e.to_string()))?,
        )
        .await
    }

    pub async fn refresh_activity(&self, sandbox_id: &str) -> Result<(), DaytonaError> {
        let req = with_auth(
            self.http.post(join_url(
                &self.config.api_url,
                &format!("/sandbox/{}/activity", sandbox_id),
            )),
            &self.config,
        )?;
        decode_empty(
            req.send()
                .await
                .map_err(|e| DaytonaError::Network(e.to_string()))?,
        )
        .await
    }

    pub async fn get_signed_preview_url(
        &self,
        sandbox_id: &str,
        port: i32,
        expires_in_seconds: Option<i32>,
    ) -> Result<SignedPreviewUrl, DaytonaError> {
        let url = match expires_in_seconds {
            Some(expires) => format!(
                "/sandbox/{}/preview/{}/signed?expiresIn={}",
                sandbox_id, port, expires
            ),
            None => format!("/sandbox/{}/preview/{}/signed", sandbox_id, port),
        };
        let req = with_auth(
            self.http.get(join_url(&self.config.api_url, &url)),
            &self.config,
        )?;
        decode_json(
            req.send()
                .await
                .map_err(|e| DaytonaError::Network(e.to_string()))?,
        )
        .await
    }

    pub async fn expire_signed_preview_url(
        &self,
        sandbox_id: &str,
        port: i32,
        token: &str,
    ) -> Result<(), DaytonaError> {
        let req = with_auth(
            self.http.delete(join_url(
                &self.config.api_url,
                &format!("/sandbox/{}/preview/{}/signed/{}", sandbox_id, port, token),
            )),
            &self.config,
        )?;
        decode_empty(
            req.send()
                .await
                .map_err(|e| DaytonaError::Network(e.to_string()))?,
        )
        .await
    }

    pub async fn set_auto_archive_interval(
        &self,
        sandbox_id: &str,
        interval: i32,
    ) -> Result<(), DaytonaError> {
        let req = with_auth(
            self.http.post(join_url(
                &self.config.api_url,
                &format!("/sandbox/{}/auto-archive-interval/{}", sandbox_id, interval),
            )),
            &self.config,
        )?;
        decode_empty(
            req.send()
                .await
                .map_err(|e| DaytonaError::Network(e.to_string()))?,
        )
        .await
    }

    pub async fn set_auto_delete_interval(
        &self,
        sandbox_id: &str,
        interval: i32,
    ) -> Result<(), DaytonaError> {
        let req = with_auth(
            self.http.post(join_url(
                &self.config.api_url,
                &format!("/sandbox/{}/auto-delete-interval/{}", sandbox_id, interval),
            )),
            &self.config,
        )?;
        decode_empty(
            req.send()
                .await
                .map_err(|e| DaytonaError::Network(e.to_string()))?,
        )
        .await
    }

    pub async fn create_ssh_access(
        &self,
        sandbox_id: &str,
        expires_in_minutes: Option<i32>,
    ) -> Result<SshAccessDto, DaytonaError> {
        let url = match expires_in_minutes {
            Some(minutes) => format!("/sandbox/{}/ssh-access?expiresIn={}", sandbox_id, minutes),
            None => format!("/sandbox/{}/ssh-access", sandbox_id),
        };
        let req = with_auth(
            self.http.post(join_url(&self.config.api_url, &url)),
            &self.config,
        )?;
        decode_json(
            req.send()
                .await
                .map_err(|e| DaytonaError::Network(e.to_string()))?,
        )
        .await
    }

    pub async fn revoke_ssh_access(
        &self,
        sandbox_id: &str,
        token: &str,
    ) -> Result<(), DaytonaError> {
        let req = with_auth(
            self.http.delete(join_url(
                &self.config.api_url,
                &format!("/sandbox/{}/ssh-access/{}", sandbox_id, token),
            )),
            &self.config,
        )?;
        decode_empty(
            req.send()
                .await
                .map_err(|e| DaytonaError::Network(e.to_string()))?,
        )
        .await
    }

    pub async fn validate_ssh_access(
        &self,
        token: &str,
    ) -> Result<SshAccessValidationDto, DaytonaError> {
        let req = with_auth(
            self.http.get(join_url(
                &self.config.api_url,
                &format!("/ssh-access/validate/{}", token),
            )),
            &self.config,
        )?;
        decode_json(
            req.send()
                .await
                .map_err(|e| DaytonaError::Network(e.to_string()))?,
        )
        .await
    }

    pub async fn get_preview_link(
        &self,
        sandbox_id: &str,
        port: i32,
    ) -> Result<PreviewLink, DaytonaError> {
        let req = with_auth(
            self.http.get(join_url(
                &self.config.api_url,
                &format!("/sandbox/{}/preview/{}", sandbox_id, port),
            )),
            &self.config,
        )?;
        decode_json(
            req.send()
                .await
                .map_err(|e| DaytonaError::Network(e.to_string()))?,
        )
        .await
    }
}
