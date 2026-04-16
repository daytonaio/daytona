// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk;

import io.daytona.sdk.model.*;
import io.daytona.sdk.exception.DaytonaException;
import io.daytona.toolbox.client.api.ProcessApi;
import io.daytona.toolbox.client.model.CodeRunRequest;
import io.daytona.toolbox.client.model.CodeRunResponse;
import io.daytona.toolbox.client.model.CreateSessionRequest;
import io.daytona.toolbox.client.model.ExecuteRequest;
import io.daytona.toolbox.client.model.PtyCreateRequest;
import io.daytona.toolbox.client.model.PtyCreateResponse;
import io.daytona.toolbox.client.model.PtyListResponse;
import io.daytona.toolbox.client.model.PtyResizeRequest;
import io.daytona.toolbox.client.model.PtySessionInfo;
import io.daytona.toolbox.client.model.SessionSendInputRequest;
import okhttp3.Request;
import okhttp3.Response;
import okhttp3.WebSocket;
import okhttp3.WebSocketListener;

import java.io.ByteArrayOutputStream;
import java.nio.charset.StandardCharsets;
import java.util.ArrayList;
import java.util.concurrent.CountDownLatch;
import java.util.concurrent.atomic.AtomicReference;
import java.util.List;
import java.util.Map;
import java.util.function.Consumer;

/**
 * Process and session execution interface for a Sandbox.
 *
 * <p>Supports single-command execution, code execution, long-running sessions, and PTY terminal
 * sessions.
 */
public class Process {
    private static final int LOG_STREAM_NONE = 0;
    private static final int LOG_STREAM_STDOUT = 1;
    private static final int LOG_STREAM_STDERR = 2;
    private static final byte STDOUT_PREFIX_BYTE = 0x01;
    private static final byte STDERR_PREFIX_BYTE = 0x02;
    private static final int PREFIX_REPEAT_COUNT = 3;

    private final ProcessApi processApi;
    private final Sandbox sandbox;

    Process(ProcessApi processApi, Sandbox sandbox) {
        this.processApi = processApi;
        this.sandbox = sandbox;
    }

    /**
     * Executes a shell command with default options.
     *
     * @param command command to execute
     * @return execution result
     * @throws DaytonaException if execution fails
     */
    public ExecuteResponse executeCommand(String command) {
        return executeCommand(command, null, null, null);
    }

    /**
     * Executes a shell command.
     *
     * @param command command to execute
     * @param cwd working directory, or {@code null} to use sandbox default
     * @param env environment variables to set for the command
     * @param timeout timeout in seconds
     * @return execution result
     * @throws DaytonaException if execution fails
     */
    public ExecuteResponse executeCommand(String command, String cwd, Map<String, String> env, Integer timeout) {
        ExecuteRequest request = new ExecuteRequest().command(command);
        if (cwd != null) {
            request.cwd(cwd);
        }
        if (env != null) {
            request.envs(env);
        }
        if (timeout != null) {
            request.timeout(timeout);
        }
        io.daytona.toolbox.client.model.ExecuteResponse response = ExceptionMapper.callToolbox(() -> processApi.executeCommand(request));
        return toExecuteResponse(response);
    }

    /**
     * Executes source code using Sandbox language tooling.
     *
     * @param code source code to execute
     * @return execution result
     * @throws DaytonaException if execution fails
     */
    public ExecuteResponse codeRun(String code) {
        return codeRun(code, null, null, null);
    }

    public ExecuteResponse codeRun(String code, Map<String, String> env, Integer timeout) {
        return codeRun(code, null, env, timeout);
    }

    public ExecuteResponse codeRun(String code, List<String> argv, Map<String, String> env, Integer timeout) {
        CodeRunRequest request = new CodeRunRequest()
                .code(code == null ? "" : code)
                .language(sandbox.getLanguage());
        if (argv != null && !argv.isEmpty()) {
            request.argv(argv);
        }
        if (env != null) {
            request.envs(env);
        }
        if (timeout != null) {
            request.timeout(timeout);
        }

        CodeRunResponse response = ExceptionMapper.callToolbox(() -> processApi.codeRun(request));
        return toExecuteResponse(response);
    }

    /**
     * Creates a persistent background session.
     *
     * @param sessionId unique session identifier
     * @throws DaytonaException if session creation fails
     */
    public void createSession(String sessionId) {
        ExceptionMapper.runToolbox(() -> processApi.createSession(new CreateSessionRequest().sessionId(sessionId)));
    }

    /**
     * Returns session metadata.
     *
     * @param sessionId session identifier
     * @return session metadata
     * @throws DaytonaException if retrieval fails
     */
    public Session getSession(String sessionId) {
        io.daytona.toolbox.client.model.Session session = ExceptionMapper.callToolbox(() -> processApi.getSession(sessionId));
        return toSession(session);
    }

    /**
     * Returns entrypoint session metadata.
     *
     * @return entrypoint session metadata
     * @throws DaytonaException if retrieval fails
     */
    public Session getEntrypointSession() {
        io.daytona.toolbox.client.model.Session session = ExceptionMapper.callToolbox(processApi::getEntrypointSession);
        return toSession(session);
    }

    /**
     * Executes a command in an existing session.
     *
     * @param sessionId session identifier
     * @param req execution request
     * @return command execution response
     * @throws DaytonaException if execution fails
     */
    public SessionExecuteResponse executeSessionCommand(String sessionId, SessionExecuteRequest req) {
        io.daytona.toolbox.client.model.SessionExecuteRequest request = new io.daytona.toolbox.client.model.SessionExecuteRequest()
                .command(req.getCommand())
                .runAsync(req.getRunAsync());
        io.daytona.toolbox.client.model.SessionExecuteResponse response = ExceptionMapper.callToolbox(() -> processApi.sessionExecuteCommand(sessionId, request));
        return toSessionExecuteResponse(response);
    }

    /**
     * Returns metadata for a command executed in a session.
     *
     * @param sessionId session identifier
     * @param commandId command identifier
     * @return command metadata
     * @throws DaytonaException if retrieval fails
     */
    public Command getSessionCommand(String sessionId, String commandId) {
        io.daytona.toolbox.client.model.Command command = ExceptionMapper.callToolbox(() -> processApi.getSessionCommand(sessionId, commandId));
        return new Command(command);
    }

    /**
     * Returns logs for a command executed in a session.
     *
     * @param sessionId session identifier
     * @param commandId command identifier
     * @return command logs
     * @throws DaytonaException if retrieval fails
     */
    public SessionCommandLogsResponse getSessionCommandLogs(String sessionId, String commandId) {
        return SessionCommandLogsResponse.from(ExceptionMapper.callToolbox(() -> processApi.getSessionCommandLogs(sessionId, commandId, null)));
    }

    /**
     * Streams logs for a command executed in a session via WebSocket.
     *
     * @param sessionId session identifier
     * @param commandId command identifier
     * @param onStdout callback for stdout chunks
     * @param onStderr callback for stderr chunks
     * @throws DaytonaException if streaming fails
     */
    public void getSessionCommandLogs(
            String sessionId,
            String commandId,
            Consumer<String> onStdout,
            Consumer<String> onStderr
    ) {
        String wsUrl = buildWsUrl(sandbox.getToolboxApiClient().getBasePath(),
                "/process/session/" + sessionId + "/command/" + commandId + "/logs?follow=true");
        streamDemuxedLogsViaWebSocket(wsUrl, onStdout, onStderr);
    }

    /**
     * Returns one-shot logs for the sandbox entrypoint session.
     *
     * @return entrypoint logs
     * @throws DaytonaException if retrieval fails
     */
    public SessionCommandLogsResponse getEntrypointLogs() {
        return SessionCommandLogsResponse.from(ExceptionMapper.callToolbox(() -> processApi.getEntrypointLogs(false)));
    }

    /**
     * Streams logs for the sandbox entrypoint session via WebSocket.
     *
     * @param onStdout callback for stdout chunks
     * @param onStderr callback for stderr chunks
     * @throws DaytonaException if streaming fails
     */
    public void getEntrypointLogs(Consumer<String> onStdout, Consumer<String> onStderr) {
        String wsUrl = buildWsUrl(sandbox.getToolboxApiClient().getBasePath(),
                "/process/session/entrypoint/logs?follow=true");
        streamDemuxedLogsViaWebSocket(wsUrl, onStdout, onStderr);
    }

    /**
     * Sends input data to a command executed in a session.
     *
     * @param sessionId session identifier
     * @param commandId command identifier
     * @param data input text to send
     * @throws DaytonaException if sending input fails
     */
    public void sendSessionCommandInput(String sessionId, String commandId, String data) {
        ExceptionMapper.runToolbox(() -> processApi.sendInput(
                sessionId,
                commandId,
                new SessionSendInputRequest().data(data)
        ));
    }

    /**
     * Deletes a session.
     *
     * @param sessionId session identifier
     * @throws DaytonaException if deletion fails
     */
    public void deleteSession(String sessionId) {
        ExceptionMapper.runToolbox(() -> processApi.deleteSession(sessionId));
    }

    /**
     * Lists all sessions in the Sandbox.
     *
     * @return session list
     * @throws DaytonaException if listing fails
     */
    public List<Session> listSessions() {
        List<io.daytona.toolbox.client.model.Session> sessions = ExceptionMapper.callToolbox(processApi::listSessions);
        List<Session> output = new ArrayList<Session>();
        if (sessions != null) {
            for (io.daytona.toolbox.client.model.Session session : sessions) {
                output.add(toSession(session));
            }
        }
        return output;
    }

    /**
     * Creates a PTY terminal session.
     *
     * @param options PTY options, or {@code null} to use defaults
     * @return PTY handle for streaming I/O and lifecycle operations
     * @throws DaytonaException if PTY session creation fails
     */
    public PtyHandle createPty(PtyCreateOptions options) {
        PtyCreateOptions createOptions = options == null ? new PtyCreateOptions() : options;
        PtyCreateRequest request = new PtyCreateRequest()
                .id(createOptions.getId())
                .cols(createOptions.getCols())
                .rows(createOptions.getRows());

        PtyCreateResponse response = ExceptionMapper.callToolbox(() -> processApi.createPtySession(request));
        if (response == null || response.getSessionId() == null || response.getSessionId().isEmpty()) {
            throw new DaytonaException("Failed to create PTY session");
        }

        return connectPty(response.getSessionId(), createOptions);
    }

    /**
     * Connects to an existing PTY terminal session.
     *
     * @param sessionId PTY session identifier
     * @return PTY handle for streaming I/O and lifecycle operations
     * @throws DaytonaException if websocket connection setup fails
     */
    public PtyHandle connectPty(String sessionId) {
        return connectPty(sessionId, null);
    }

    /**
     * Connects to an existing PTY terminal session.
     *
     * @param sessionId PTY session identifier
     * @param options PTY options, used for data callback configuration
     * @return PTY handle for streaming I/O and lifecycle operations
     * @throws DaytonaException if websocket connection setup fails
     */
    public PtyHandle connectPty(String sessionId, PtyCreateOptions options) {
        PtyCreateOptions connectOptions = options == null ? new PtyCreateOptions() : options;
        String wsUrl = buildPtyWebSocketUrl(sandbox.getToolboxApiClient().getBasePath(), sessionId);
        Request wsRequest = new Request.Builder()
                .url(wsUrl)
                .addHeader("Authorization", "Bearer " + sandbox.getApiKey())
                .build();

        return new PtyHandle(
                sandbox.getToolboxApiClient().getHttpClient(),
                wsRequest,
                sessionId,
                this::resizePtySession,
                this::killPtySession,
                connectOptions.getOnData()
        );
    }

    /**
     * Lists PTY sessions in the Sandbox.
     *
     * @return PTY session information list
     * @throws DaytonaException if listing fails
     */
    public List<PtySessionInfo> listPtySessions() {
        PtyListResponse response = ExceptionMapper.callToolbox(processApi::listPtySessions);
        return response == null || response.getSessions() == null
                ? new ArrayList<PtySessionInfo>()
                : response.getSessions();
    }

    /**
     * Returns PTY session information.
     *
     * @param sessionId PTY session identifier
     * @return PTY session information
     * @throws DaytonaException if retrieval fails
     */
    public PtySessionInfo getPtySessionInfo(String sessionId) {
        return ExceptionMapper.callToolbox(() -> processApi.getPtySession(sessionId));
    }

    /**
     * Resizes an active PTY session.
     *
     * @param sessionId PTY session identifier
     * @param cols terminal width in columns
     * @param rows terminal height in rows
     * @throws DaytonaException if resize fails
     */
    public void resizePtySession(String sessionId, int cols, int rows) {
        ExceptionMapper.callToolbox(() -> processApi.resizePtySession(
                sessionId,
                new PtyResizeRequest().cols(cols).rows(rows)
        ));
    }

    /**
     * Terminates a PTY session.
     *
     * @param sessionId PTY session identifier
     * @throws DaytonaException if termination fails
     */
    public void killPtySession(String sessionId) {
        ExceptionMapper.callToolbox(() -> processApi.deletePtySession(sessionId));
    }

    private String buildPtyWebSocketUrl(String toolboxBaseUrl, String sessionId) {
        if (toolboxBaseUrl == null || toolboxBaseUrl.isEmpty()) {
            throw new DaytonaException("Toolbox base URL is not available");
        }
        String wsBase = toolboxBaseUrl
                .replaceFirst("^https://", "wss://")
                .replaceFirst("^http://", "ws://");
        return wsBase + "/process/pty/" + sessionId + "/connect";
    }

    private void streamDemuxedLogsViaWebSocket(String wsUrl, Consumer<String> onStdout, Consumer<String> onStderr) {
        Request wsRequest = new Request.Builder()
                .url(wsUrl)
                .addHeader("Authorization", "Bearer " + sandbox.getApiKey())
                .build();

        final CountDownLatch doneLatch = new CountDownLatch(1);
        final AtomicReference<RuntimeException> failure = new AtomicReference<>(null);

        sandbox.getToolboxApiClient().getHttpClient().newWebSocket(wsRequest, new WebSocketListener() {
            final ByteArrayOutputStream stdoutBuf = new ByteArrayOutputStream();
            final ByteArrayOutputStream stderrBuf = new ByteArrayOutputStream();
            final ByteArrayOutputStream markerBuf = new ByteArrayOutputStream();
            int streamState = LOG_STREAM_NONE;
            byte markerByte = 0;
            int markerCount = 0;

            @Override
            public void onMessage(WebSocket webSocket, okio.ByteString bytes) {
                demux(bytes.toByteArray());
            }

            @Override
            public void onMessage(WebSocket webSocket, String text) {
                demux(text.getBytes(StandardCharsets.UTF_8));
            }

            @Override
            public void onClosing(WebSocket webSocket, int code, String reason) {
                flush();
                webSocket.close(1000, null);
                doneLatch.countDown();
            }

            @Override
            public void onClosed(WebSocket webSocket, int code, String reason) {
                flush();
                doneLatch.countDown();
            }

            @Override
            public void onFailure(WebSocket webSocket, Throwable t, Response response) {
                flush();
                failure.compareAndSet(null, new DaytonaException("Log streaming failed: " + t.getMessage()));
                doneLatch.countDown();
            }

            private void demux(byte[] data) {
                for (byte value : data) {
                    if (value == STDOUT_PREFIX_BYTE || value == STDERR_PREFIX_BYTE) {
                        if (markerCount == 0) {
                            markerByte = value;
                            markerCount = 1;
                            markerBuf.write(value);
                        } else if (markerByte == value) {
                            markerCount++;
                            markerBuf.write(value);
                        } else {
                            drainMarker();
                            markerBuf.write(value);
                            markerByte = value;
                            markerCount = 1;
                        }
                        if (markerCount >= PREFIX_REPEAT_COUNT) {
                            emitBuffer(stdoutBuf, onStdout);
                            emitBuffer(stderrBuf, onStderr);
                            markerBuf.reset();
                            markerCount = 0;
                            streamState = markerByte == STDOUT_PREFIX_BYTE ? LOG_STREAM_STDOUT : LOG_STREAM_STDERR;
                        }
                        continue;
                    }
                    drainMarker();
                    markerCount = 0;
                    appendToStream(value);
                }
            }

            private void drainMarker() {
                if (markerBuf.size() == 0) return;
                for (byte b : markerBuf.toByteArray()) {
                    appendToStream(b);
                }
                markerBuf.reset();
            }

            private void appendToStream(byte value) {
                if (streamState == LOG_STREAM_STDOUT) stdoutBuf.write(value);
                else if (streamState == LOG_STREAM_STDERR) stderrBuf.write(value);
            }

            private void flush() {
                drainMarker();
                emitBuffer(stdoutBuf, onStdout);
                emitBuffer(stderrBuf, onStderr);
            }

            private void emitBuffer(ByteArrayOutputStream buf, Consumer<String> consumer) {
                if (consumer == null || buf.size() == 0) { buf.reset(); return; }
                consumer.accept(new String(buf.toByteArray(), StandardCharsets.UTF_8));
                buf.reset();
            }
        });

        try {
            doneLatch.await();
        } catch (InterruptedException e) {
            Thread.currentThread().interrupt();
            throw new DaytonaException("Interrupted while streaming logs", e);
        }
        RuntimeException ex = failure.get();
        if (ex != null) throw ex;
    }

    private String buildWsUrl(String basePath, String path) {
        if (basePath == null || basePath.isEmpty()) {
            throw new DaytonaException("Toolbox base URL is not available");
        }
        String wsBase = basePath
                .replaceFirst("^https://", "wss://")
                .replaceFirst("^http://", "ws://");
        return wsBase + path;
    }

    private ExecuteResponse toExecuteResponse(io.daytona.toolbox.client.model.ExecuteResponse source) {
        return new ExecuteResponse(source);
    }

    private ExecuteResponse toExecuteResponse(CodeRunResponse source) {
        return new ExecuteResponse(source);
    }

    private Session toSession(io.daytona.toolbox.client.model.Session source) {
        return new Session(source);
    }

    private SessionExecuteResponse toSessionExecuteResponse(io.daytona.toolbox.client.model.SessionExecuteResponse source) {
        return new SessionExecuteResponse(source);
    }
}
