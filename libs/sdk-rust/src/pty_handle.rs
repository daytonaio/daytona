// Copyright 2019-2025 Daytona Labs Inc.
// SPDX-License-Identifier: Apache-2.0

use crate::error::DaytonaError;
use futures::stream::{SplitSink, SplitStream};
use futures::{SinkExt, StreamExt};
use std::future::Future;
use std::pin::Pin;
use std::sync::Arc;
use tokio::net::TcpStream;
use tokio::sync::{mpsc, Mutex, Notify, RwLock};
use tokio_tungstenite::{tungstenite::Message, MaybeTlsStream, WebSocketStream};

/// WebSocket stream type alias.
type WsStream = WebSocketStream<MaybeTlsStream<TcpStream>>;

/// Control message from PTY WebSocket server.
#[allow(dead_code)]
#[derive(Debug, serde:: Deserialize)]
struct ControlMessage {
    #[serde(rename = "type")]
    type_: String,
    status: Option<String>,
    error: Option<String>,
}

/// Exit data parsed from WebSocket close reason.
#[allow(dead_code)]
#[derive(Debug, serde::Deserialize)]
#[serde(rename_all = "camelCase")]
struct ExitData {
    exit_code: Option<i32>,
    exit_reason: Option<String>,
    error: Option<String>,
}

/// Type alias for the async resize callback.
pub type ResizeFn = Box<
    dyn Fn(i32, i32) -> Pin<Box<dyn Future<Output = Result<(), DaytonaError>> + Send>>
        + Send
        + Sync,
>;

/// Type alias for the async kill callback.
pub type KillFn =
    Box<dyn Fn() -> Pin<Box<dyn Future<Output = Result<(), DaytonaError>> + Send>> + Send + Sync>;

/// Handle for an interactive PTY session over WebSocket.
///
/// Provides methods for sending input, receiving output via channels,
/// resizing the terminal, and managing the connection lifecycle.
///
/// Create a `PtyHandle` via `ProcessService::create_pty` or similar.
///
/// # Example
///
/// ```no_run
/// # async fn example(handle: &mut daytona::PtyHandle) -> Result<(), daytona::DaytonaError> {
/// handle.wait_for_connection().await?;
/// handle.send_input(b"ls -la\n").await?;
///
/// while let Some(data) = handle.data_chan().recv().await {
///     print!("{}", String::from_utf8_lossy(&data));
/// }
/// # Ok(())
/// # }
/// ```
pub struct PtyHandle {
    write: Arc<Mutex<SplitSink<WsStream, Message>>>,
    session_id: String,
    data_rx: mpsc::Receiver<Vec<u8>>,
    exit_code: Arc<RwLock<Option<i32>>>,
    error: Arc<RwLock<Option<String>>>,
    connected: Arc<RwLock<bool>>,
    connection_established: Arc<RwLock<bool>>,
    done: Arc<Notify>,
    handle_resize: Arc<ResizeFn>,
    handle_kill: Arc<KillFn>,
}

#[allow(dead_code)]
impl PtyHandle {
    /// Create a new PtyHandle from an established WebSocket connection.
    ///
    /// Splits the WebSocket into read/write halves and spawns a background
    /// task to read messages, parse control messages, and forward terminal
    /// output to the data channel.
    pub(crate) fn new(
        ws: WsStream,
        session_id: String,
        handle_resize: ResizeFn,
        handle_kill: KillFn,
    ) -> Self {
        let (write, read) = ws.split();
        let write = Arc::new(Mutex::new(write));
        let (data_tx, data_rx) = mpsc::channel(100);
        let exit_code: Arc<RwLock<Option<i32>>> = Arc::new(RwLock::new(None));
        let error: Arc<RwLock<Option<String>>> = Arc::new(RwLock::new(None));
        let connected = Arc::new(RwLock::new(false));
        let connection_established = Arc::new(RwLock::new(false));
        let done = Arc::new(Notify::new());

        // Spawn background message handler
        {
            let exit_code = Arc::clone(&exit_code);
            let error = Arc::clone(&error);
            let connected = Arc::clone(&connected);
            let connection_established = Arc::clone(&connection_established);
            let done = Arc::clone(&done);

            tokio::spawn(async move {
                Self::handle_messages(
                    read,
                    data_tx,
                    exit_code,
                    error,
                    connected,
                    connection_established,
                    done,
                )
                .await;
            });
        }

        Self {
            write,
            session_id,
            data_rx,
            exit_code,
            error,
            connected,
            connection_established,
            done,
            handle_resize: Arc::new(handle_resize),
            handle_kill: Arc::new(handle_kill),
        }
    }

    /// Get the session ID.
    pub fn session_id(&self) -> &str {
        &self.session_id
    }

    /// Check if connected to the PTY session.
    pub async fn is_connected(&self) -> bool {
        *self.connected.read().await
    }

    /// Wait for the WebSocket connection to be established.
    ///
    /// Blocks until the PTY session confirms it is ready (via a
    /// `{"type":"control","status":"connected"}` message), or until
    /// a 10-second timeout expires.
    pub async fn wait_for_connection(&self) -> Result<(), DaytonaError> {
        if *self.connection_established.read().await {
            return Ok(());
        }

        let start = tokio::time::Instant::now();
        let timeout = std::time::Duration::from_secs(10);

        loop {
            if *self.connection_established.read().await {
                return Ok(());
            }

            if let Some(err) = self.error.read().await.as_ref() {
                return Err(DaytonaError::Network(err.clone()));
            }

            if start.elapsed() >= timeout {
                return Err(DaytonaError::Timeout("PTY connection timeout".to_string()));
            }

            tokio::time::sleep(std::time::Duration::from_millis(100)).await;
        }
    }

    /// Get mutable reference to the data channel for reading PTY output.
    ///
    /// Use `recv().await` on the returned receiver to get output bytes.
    /// The channel is closed when the PTY session ends.
    pub fn data_chan(&mut self) -> &mut mpsc::Receiver<Vec<u8>> {
        &mut self.data_rx
    }

    /// Send input data to the PTY session.
    ///
    /// Data is sent as a binary WebSocket message and will be processed
    /// as if typed in the terminal.
    pub async fn send_input(&self, data: &[u8]) -> Result<(), DaytonaError> {
        if !*self.connected.read().await {
            return Err(DaytonaError::Network("PTY is not connected".to_string()));
        }

        let mut write = self.write.lock().await;
        write
            .send(Message::Binary(data.to_vec()))
            .await
            .map_err(|e| DaytonaError::Network(format!("Failed to send input to PTY: {}", e)))
    }

    /// Resize the PTY terminal dimensions.
    ///
    /// Notifies terminal applications about the new dimensions via SIGWINCH.
    pub async fn resize(&self, cols: i32, rows: i32) -> Result<(), DaytonaError> {
        (self.handle_resize)(cols, rows).await
    }

    /// Get exit code if the session has ended.
    pub async fn exit_code(&self) -> Option<i32> {
        *self.exit_code.read().await
    }

    /// Get error message if the session failed.
    pub async fn error(&self) -> Option<String> {
        self.error.read().await.clone()
    }

    /// Disconnect from PTY by closing the WebSocket connection.
    ///
    /// This does not terminate the underlying process — use [`kill`](Self::kill)
    /// for that.
    pub async fn disconnect(self) -> Result<(), DaytonaError> {
        let mut write = self.write.lock().await;
        let _ = write.send(Message::Close(None)).await;
        write
            .close()
            .await
            .map_err(|e| DaytonaError::Network(format!("Failed to close WebSocket: {}", e)))
    }

    /// Kill the PTY session and terminate the associated process.
    ///
    /// Calls the kill handler (typically an HTTP DELETE to the toolbox API).
    pub async fn kill(&self) -> Result<(), DaytonaError> {
        (self.handle_kill)().await
    }

    /// Wait for the PTY process to exit and return the exit code.
    ///
    /// Blocks until the PTY session ends (via WebSocket close or error).
    pub async fn wait(&self) -> Result<i32, DaytonaError> {
        // Check if already exited
        if let Some(code) = *self.exit_code.read().await {
            return Ok(code);
        }

        loop {
            tokio::select! {
                _ = self.done.notified() => {
                    if let Some(code) = *self.exit_code.read().await {
                        return Ok(code);
                    }
                    if let Some(err) = self.error.read().await.as_ref() {
                        return Err(DaytonaError::Network(err.clone()));
                    }
                }
                _ = tokio::time::sleep(std::time::Duration::from_millis(100)) => {
                    if let Some(code) = *self.exit_code.read().await {
                        return Ok(code);
                    }
                    if let Some(err) = self.error.read().await.as_ref() {
                        return Err(DaytonaError::Network(err.clone()));
                    }
                }
            }
        }
    }

    // ── Background message handler ──────────────────────────────────────

    /// Background task: read WebSocket messages, dispatch to data channel
    /// or handle control / close frames.
    ///
    /// 1. Text messages that parse as `{"type":"control",...}` update
    ///    internal state (connected, error).
    /// 2. Other text and all binary messages are forwarded to `data_tx`.
    /// 3. Close frames are parsed for exit code / error information.
    async fn handle_messages(
        mut read: SplitStream<WsStream>,
        data_tx: mpsc::Sender<Vec<u8>>,
        exit_code: Arc<RwLock<Option<i32>>>,
        error: Arc<RwLock<Option<String>>>,
        connected: Arc<RwLock<bool>>,
        connection_established: Arc<RwLock<bool>>,
        done: Arc<Notify>,
    ) {
        loop {
            match read.next().await {
                Some(Ok(Message::Text(text))) => {
                    // Try to parse as control message
                    if let Ok(ctrl) = serde_json::from_str::<ControlMessage>(&text) {
                        if ctrl.type_ == "control" {
                            match ctrl.status.as_deref() {
                                Some("connected") => {
                                    *connected.write().await = true;
                                    *connection_established.write().await = true;
                                }
                                Some("error") => {
                                    let err_msg = ctrl
                                        .error
                                        .unwrap_or_else(|| "Unknown connection error".into());
                                    *error.write().await = Some(err_msg);
                                    *connected.write().await = false;
                                }
                                _ => {}
                            }
                            continue;
                        }
                    }
                    // Regular text output — forward to data channel
                    let _ = data_tx.send(text.as_bytes().to_vec()).await;
                }
                Some(Ok(Message::Binary(data))) => {
                    let _ = data_tx.send(data.to_vec()).await;
                }
                Some(Ok(Message::Close(frame))) => {
                    *connected.write().await = false;
                    if let Some(frame) = frame {
                        let reason = frame.reason.to_string();
                        Self::parse_close_reason(&reason, &exit_code, &error).await;
                    } else {
                        *exit_code.write().await = Some(0);
                    }
                    done.notify_waiters();
                    return;
                }
                Some(Ok(_)) => {
                    // Ping / Pong — ignore
                }
                Some(Err(e)) => {
                    use tokio_tungstenite::tungstenite::Error as WsError;
                    match &e {
                        WsError::ConnectionClosed | WsError::AlreadyClosed => {
                            if exit_code.read().await.is_none() {
                                *exit_code.write().await = Some(0);
                            }
                        }
                        _ => {
                            *error.write().await = Some(e.to_string());
                        }
                    }
                    *connected.write().await = false;
                    done.notify_waiters();
                    return;
                }
                None => {
                    // Stream ended
                    *connected.write().await = false;
                    if exit_code.read().await.is_none() {
                        *exit_code.write().await = Some(0);
                    }
                    done.notify_waiters();
                    return;
                }
            }
        }
    }

    /// Parse close reason string as JSON exit data.
    async fn parse_close_reason(
        reason: &str,
        exit_code: &Arc<RwLock<Option<i32>>>,
        error: &Arc<RwLock<Option<String>>>,
    ) {
        if reason.is_empty() {
            *exit_code.write().await = Some(0);
            return;
        }

        if let Ok(exit_data) = serde_json::from_str::<ExitData>(reason) {
            if let Some(code) = exit_data.exit_code {
                *exit_code.write().await = Some(code);
            }
            if let Some(exit_reason) = exit_data.exit_reason {
                *error.write().await = Some(exit_reason);
            }
            if let Some(err) = exit_data.error {
                *error.write().await = Some(err);
            }
        } else {
            // Not JSON, default to exit code 0
            *exit_code.write().await = Some(0);
        }
    }
}
