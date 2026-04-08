// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk;

import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;
import io.daytona.sdk.exception.DaytonaException;
import io.daytona.sdk.exception.DaytonaTimeoutException;
import io.daytona.toolbox.client.api.InterpreterApi;
import io.daytona.toolbox.client.model.CreateContextRequest;
import io.daytona.toolbox.client.model.InterpreterContext;
import io.daytona.toolbox.client.model.ListContextsResponse;
import okhttp3.Request;
import okhttp3.Response;
import okhttp3.WebSocket;
import okhttp3.WebSocketListener;

import java.util.ArrayList;
import java.util.Collections;
import java.util.List;
import java.util.concurrent.CountDownLatch;
import java.util.concurrent.atomic.AtomicBoolean;
import java.util.concurrent.atomic.AtomicReference;

/**
 * Stateful code interpretation interface for a Sandbox.
 *
 * <p>Provides Python code execution in interpreter contexts that preserve state between runs.
 */
public class CodeInterpreter {
    private static final ObjectMapper OBJECT_MAPPER = new ObjectMapper();
    private static final int WEBSOCKET_TIMEOUT_CODE = 4008;

    private final InterpreterApi interpreterApi;
    private final Sandbox sandbox;

    CodeInterpreter(InterpreterApi interpreterApi, Sandbox sandbox) {
        this.interpreterApi = interpreterApi;
        this.sandbox = sandbox;
    }

    /**
     * Executes Python code in the default interpreter context.
     *
     * @param code Python code to execute
     * @return aggregated execution result
     * @throws DaytonaException if code is empty, connection fails, or execution fails
     */
    public ExecutionResult runCode(String code) {
        return runCode(code, null);
    }

    /**
     * Executes Python code with streaming callbacks and options.
     *
     * @param code Python code to execute
     * @param options execution options including callbacks and timeout; may be {@code null}
     * @return aggregated execution result
     * @throws DaytonaException if code is empty, connection fails, or execution fails
     */
    public ExecutionResult runCode(String code, RunCodeOptions options) {
        if (code == null || code.trim().isEmpty()) {
            throw new DaytonaException("Code is required for execution");
        }
        final RunCodeOptions opts = options == null ? new RunCodeOptions() : options;

        final String wsUrl = buildInterpreterWebSocketUrl(sandbox.getToolboxApiClient().getBasePath());
        final Request wsRequest = new Request.Builder()
                .url(wsUrl)
                .addHeader("Authorization", "Bearer " + sandbox.getApiKey())
                .build();

        final StringBuilder stdout = new StringBuilder();
        final StringBuilder stderr = new StringBuilder();
        final AtomicReference<ExecutionError> executionError = new AtomicReference<ExecutionError>(null);
        final AtomicReference<RuntimeException> failure = new AtomicReference<RuntimeException>(null);
        final CountDownLatch doneLatch = new CountDownLatch(1);
        final AtomicBoolean finished = new AtomicBoolean(false);

        sandbox.getToolboxApiClient().getHttpClient().newWebSocket(wsRequest, new WebSocketListener() {
            @Override
            public void onOpen(WebSocket webSocket, Response response) {
                try {
                    com.fasterxml.jackson.databind.node.ObjectNode payload = OBJECT_MAPPER.createObjectNode();
                    payload.put("code", code);
                    if (opts.getTimeout() != null) {
                        payload.put("timeout", opts.getTimeout());
                    }
                    webSocket.send(OBJECT_MAPPER.writeValueAsString(payload));
                } catch (Exception e) {
                    failure.compareAndSet(null, new DaytonaException("Failed to send execute request", e));
                    closeAndComplete(webSocket, 1011, "send failure");
                }
            }

            @Override
            public void onMessage(WebSocket webSocket, String text) {
                if (text == null || text.isEmpty()) {
                    return;
                }
                try {
                    JsonNode chunk = OBJECT_MAPPER.readTree(text);
                    String chunkType = chunk.path("type").asText("");
                    if ("stdout".equals(chunkType)) {
                        String content = chunk.path("text").asText("");
                        stdout.append(content);
                        if (opts.getOnStdout() != null) {
                            opts.getOnStdout().accept(content);
                        }
                        return;
                    }
                    if ("stderr".equals(chunkType)) {
                        String content = chunk.path("text").asText("");
                        stderr.append(content);
                        if (opts.getOnStderr() != null) {
                            opts.getOnStderr().accept(content);
                        }
                        return;
                    }
                    if ("error".equals(chunkType)) {
                        ExecutionError err = new ExecutionError(
                                chunk.path("name").asText(""),
                                chunk.path("value").asText(""),
                                chunk.path("traceback").asText("")
                        );
                        executionError.set(err);
                        if (opts.getOnError() != null) {
                            opts.getOnError().accept(err);
                        }
                        return;
                    }
                    if ("control".equals(chunkType)) {
                        String status = chunk.path("text").asText("");
                        if ("completed".equals(status) || "interrupted".equals(status)) {
                            closeAndComplete(webSocket, 1000, status);
                        }
                    }
                } catch (Exception ignored) {
                }
            }

            @Override
            public void onClosing(WebSocket webSocket, int code, String reason) {
                handleClose(code, reason);
                webSocket.close(code, reason);
            }

            @Override
            public void onClosed(WebSocket webSocket, int code, String reason) {
                handleClose(code, reason);
            }

            @Override
            public void onFailure(WebSocket webSocket, Throwable t, Response response) {
                String message = t == null ? "Interpreter WebSocket failure" : t.getMessage();
                failure.compareAndSet(null, new DaytonaException("Failed to execute code: " + message));
                complete();
            }

            private void handleClose(int code, String reason) {
                if (code == WEBSOCKET_TIMEOUT_CODE) {
                    failure.compareAndSet(null, new DaytonaTimeoutException(
                            "Execution timed out: operation exceeded the configured timeout."));
                } else if (code != 1000 && code != 1001) {
                    String suffix = (reason == null || reason.isEmpty()) ? "" : (": " + reason);
                    failure.compareAndSet(null,
                            new DaytonaException("Code execution failed: WebSocket closed with code " + code + suffix));
                }
                complete();
            }

            private void closeAndComplete(WebSocket webSocket, int code, String reason) {
                try {
                    webSocket.close(code, reason);
                } catch (Exception ignored) {
                }
                complete();
            }

            private void complete() {
                if (finished.compareAndSet(false, true)) {
                    doneLatch.countDown();
                }
            }
        });

        try {
            doneLatch.await();
        } catch (InterruptedException e) {
            Thread.currentThread().interrupt();
            throw new DaytonaException("Interrupted while waiting for interpreter execution", e);
        }

        RuntimeException exception = failure.get();
        if (exception != null) {
            throw exception;
        }

        return new ExecutionResult(stdout.toString(), stderr.toString(), executionError.get());
    }

    /**
     * Creates a new interpreter context using sandbox defaults.
     *
     * @return created interpreter context metadata
     * @throws DaytonaException if context creation fails
     */
    public InterpreterContext createContext() {
        return createContext(null);
    }

    /**
     * Creates a new interpreter context.
     *
     * @param cwd working directory for the new context; {@code null} uses sandbox default
     * @return created interpreter context metadata
     * @throws DaytonaException if context creation fails
     */
    public InterpreterContext createContext(String cwd) {
        CreateContextRequest request = new CreateContextRequest().cwd(cwd);
        return ExceptionMapper.callToolbox(() -> interpreterApi.createInterpreterContext(request));
    }

    /**
     * Lists all user-created interpreter contexts.
     *
     * @return list of interpreter contexts; empty list when no contexts exist
     * @throws DaytonaException if listing contexts fails
     */
    public List<InterpreterContext> listContexts() {
        ListContextsResponse response = ExceptionMapper.callToolbox(interpreterApi::listInterpreterContexts);
        if (response == null || response.getContexts() == null) {
            return Collections.emptyList();
        }
        return new ArrayList<InterpreterContext>(response.getContexts());
    }

    /**
     * Deletes an interpreter context.
     *
     * @param contextId context identifier to delete
     * @throws DaytonaException if deletion fails
     */
    public void deleteContext(String contextId) {
        ExceptionMapper.runToolbox(() -> interpreterApi.deleteInterpreterContext(contextId));
    }

    private String buildInterpreterWebSocketUrl(String toolboxBaseUrl) {
        if (toolboxBaseUrl == null || toolboxBaseUrl.isEmpty()) {
            throw new DaytonaException("Toolbox base URL is not available");
        }
        String wsBase = toolboxBaseUrl
                .replaceFirst("^https://", "wss://")
                .replaceFirst("^http://", "ws://");
        return wsBase + "/process/interpreter/execute";
    }

    /**
     * Structured execution error returned by the interpreter.
     */
    public static class ExecutionError {
        private final String name;
        private final String value;
        private final String traceback;

        /**
         * Creates an interpreter execution error object.
         *
         * @param name error type name
         * @param value error message value
         * @param traceback Python traceback text
         */
        public ExecutionError(String name, String value, String traceback) {
            this.name = name;
            this.value = value;
            this.traceback = traceback;
        }

        /**
         * Returns the interpreter error type.
         *
         * @return error type name
         */
        public String getName() {
            return name;
        }

        /**
         * Returns the interpreter error message.
         *
         * @return error value
         */
        public String getValue() {
            return value;
        }

        /**
         * Returns the interpreter traceback text.
         *
         * @return traceback string
         */
        public String getTraceback() {
            return traceback;
        }
    }

    /**
     * Aggregated result of interpreter execution.
     */
    public static class ExecutionResult {
        private final String stdout;
        private final String stderr;
        private final ExecutionError error;

        /**
         * Creates an execution result object.
         *
         * @param stdout aggregated standard output
         * @param stderr aggregated standard error
         * @param error structured execution error, or {@code null} when execution succeeded
         */
        public ExecutionResult(String stdout, String stderr, ExecutionError error) {
            this.stdout = stdout;
            this.stderr = stderr;
            this.error = error;
        }

        /**
         * Returns aggregated standard output.
         *
         * @return stdout text
         */
        public String getStdout() {
            return stdout;
        }

        /**
         * Returns aggregated standard error.
         *
         * @return stderr text
         */
        public String getStderr() {
            return stderr;
        }

        /**
         * Returns structured execution error details.
         *
         * @return execution error or {@code null}
         */
        public ExecutionError getError() {
            return error;
        }
    }
}
