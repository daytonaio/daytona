// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk;

import io.daytona.sdk.model.ExecuteResponse;
import io.daytona.sdk.model.Session;
import io.daytona.sdk.model.SessionCommandLogsResponse;
import io.daytona.sdk.model.SessionExecuteRequest;
import io.daytona.sdk.model.SessionExecuteResponse;
import io.daytona.sdk.exception.DaytonaBadRequestException;
import io.daytona.sdk.exception.DaytonaConflictException;
import io.daytona.sdk.exception.DaytonaForbiddenException;
import io.daytona.sdk.exception.DaytonaNotFoundException;
import io.daytona.sdk.exception.DaytonaRateLimitException;
import io.daytona.sdk.exception.DaytonaServerException;
import io.daytona.toolbox.client.api.ProcessApi;
import io.daytona.toolbox.client.model.CodeRunArtifacts;
import io.daytona.toolbox.client.model.CodeRunResponse;
import io.daytona.toolbox.client.model.Command;
import io.daytona.toolbox.client.model.ExecuteRequest;
import io.daytona.toolbox.client.model.PtyCreateRequest;
import io.daytona.toolbox.client.model.PtySessionInfo;
import io.daytona.toolbox.client.model.PtyCreateResponse;
import okhttp3.WebSocket;
import okhttp3.WebSocketListener;
import okhttp3.mockwebserver.MockResponse;
import okhttp3.mockwebserver.MockWebServer;
import okio.ByteString;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.junit.jupiter.params.ParameterizedTest;
import org.junit.jupiter.params.provider.Arguments;
import org.junit.jupiter.params.provider.MethodSource;
import org.mockito.ArgumentCaptor;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import java.util.Arrays;
import java.util.Collections;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.ArrayList;
import java.util.stream.Stream;

import static org.assertj.core.api.Assertions.assertThat;
import static org.assertj.core.api.Assertions.assertThatThrownBy;
import static org.mockito.ArgumentMatchers.any;
import static org.mockito.ArgumentMatchers.eq;
import static org.mockito.Mockito.verify;
import static org.mockito.Mockito.when;

@ExtendWith(MockitoExtension.class)
class ProcessTest {

    @Mock
    private ProcessApi processApi;

    private Sandbox sandbox;
    private Process process;

    @BeforeEach
    void setUp() {
        sandbox = TestSupport.mockSandbox("http://127.0.0.1:1/toolbox", "python", "test-key", new okhttp3.OkHttpClient());
        process = new Process(processApi, sandbox);
    }

    @Test
    void executeCommandUsesMinimalRequest() {
        io.daytona.toolbox.client.model.ExecuteResponse response = new io.daytona.toolbox.client.model.ExecuteResponse();
        response.setExitCode(0);
        response.setResult("ok");
        when(processApi.executeCommand(any())).thenReturn(response);

        ExecuteResponse result = process.executeCommand("echo hello");

        assertThat(result.getExitCode()).isEqualTo(0);
        assertThat(result.getResult()).isEqualTo("ok");
        ArgumentCaptor<ExecuteRequest> captor = ArgumentCaptor.forClass(ExecuteRequest.class);
        verify(processApi).executeCommand(captor.capture());
        assertThat(captor.getValue().getCommand()).isEqualTo("echo hello");
        assertThat(captor.getValue().getCwd()).isNull();
    }

    @Test
    void executeCommandPassesCwdEnvAndTimeout() {
        Map<String, String> env = new HashMap<String, String>();
        env.put("A", "1");
        when(processApi.executeCommand(any())).thenReturn(new io.daytona.toolbox.client.model.ExecuteResponse().exitCode(0).result("ok"));

        process.executeCommand("pwd", "/workspace", env, 15);

        ArgumentCaptor<ExecuteRequest> captor = ArgumentCaptor.forClass(ExecuteRequest.class);
        verify(processApi).executeCommand(captor.capture());
        assertThat(captor.getValue().getCwd()).isEqualTo("/workspace");
        assertThat(captor.getValue().getEnvs()).containsEntry("A", "1");
        assertThat(captor.getValue().getTimeout()).isEqualTo(15);
    }

    @Test
    void codeRunUsesSandboxLanguageAndArtifacts() {
        CodeRunArtifacts artifacts = new CodeRunArtifacts();
        CodeRunResponse response = new CodeRunResponse();
        response.setExitCode(1);
        response.setResult("boom");
        response.setArtifacts(artifacts);
        when(processApi.codeRun(any())).thenReturn(response);

        ExecuteResponse result = process.codeRun("print('hi')", Arrays.asList("-u"), Collections.singletonMap("DEBUG", "1"), 30);

        assertThat(result.getExitCode()).isEqualTo(1);
        assertThat(result.getArtifacts()).isSameAs(artifacts);
        ArgumentCaptor<io.daytona.toolbox.client.model.CodeRunRequest> captor = ArgumentCaptor.forClass(io.daytona.toolbox.client.model.CodeRunRequest.class);
        verify(processApi).codeRun(captor.capture());
        assertThat(captor.getValue().getLanguage()).isEqualTo("python");
        assertThat(captor.getValue().getArgv()).containsExactly("-u");
        assertThat(captor.getValue().getEnvs()).containsEntry("DEBUG", "1");
        assertThat(captor.getValue().getTimeout()).isEqualTo(30);
    }

    @Test
    void codeRunDefaultsNullCodeToEmptyStringAndSkipsOptionalFields() {
        when(processApi.codeRun(any())).thenReturn(new CodeRunResponse().exitCode(0).result("ok"));

        process.codeRun(null, Collections.<String>emptyList(), null, null);

        ArgumentCaptor<io.daytona.toolbox.client.model.CodeRunRequest> captor = ArgumentCaptor.forClass(io.daytona.toolbox.client.model.CodeRunRequest.class);
        verify(processApi).codeRun(captor.capture());
        assertThat(captor.getValue().getCode()).isEmpty();
        assertThat(captor.getValue().getArgv()).isEmpty();
        assertThat(captor.getValue().getTimeout()).isNull();
        assertThat(captor.getValue().getEnvs()).isEmpty();
    }

    @Test
    void sessionLifecycleMethodsDelegate() {
        process.createSession("session-1");
        process.sendSessionCommandInput("session-1", "cmd-1", "input");
        process.deleteSession("session-1");

        verify(processApi).createSession(argThat(request -> "session-1".equals(request.getSessionId())));
        verify(processApi).sendInput(eq("session-1"), eq("cmd-1"), argThat(request -> "input".equals(request.getData())));
        verify(processApi).deleteSession("session-1");
    }

    @Test
    void getSessionAndEntrypointMapResponses() {
        io.daytona.toolbox.client.model.Session session = new io.daytona.toolbox.client.model.Session();
        session.setSessionId("session-1");
        session.setCommands(Collections.<Command>emptyList());
        when(processApi.getSession("session-1")).thenReturn(session);
        when(processApi.getEntrypointSession()).thenReturn(session);

        Session explicit = process.getSession("session-1");
        Session entrypoint = process.getEntrypointSession();

        assertThat(explicit.getSessionId()).isEqualTo("session-1");
        assertThat(entrypoint.getSessionId()).isEqualTo("session-1");
    }

    @Test
    void executeSessionCommandMapsResponse() {
        io.daytona.toolbox.client.model.SessionExecuteResponse response = new io.daytona.toolbox.client.model.SessionExecuteResponse();
        response.setCmdId("cmd-1");
        response.setOutput("combined");
        response.setStdout("out");
        response.setStderr("err");
        response.setExitCode(0);
        when(processApi.sessionExecuteCommand(eq("session-1"), any())).thenReturn(response);

        SessionExecuteResponse result = process.executeSessionCommand("session-1", new SessionExecuteRequest("pwd", true));

        assertThat(result.getCmdId()).isEqualTo("cmd-1");
        assertThat(result.getOutput()).isEqualTo("combined");
    }

    @Test
    void getSessionCommandAndLogsMapResponses() {
        Command command = new Command();
        command.setId("cmd-1");
        command.setCommand("ls");
        command.setExitCode(0);
        io.daytona.toolbox.client.model.SessionCommandLogsResponse logs = new io.daytona.toolbox.client.model.SessionCommandLogsResponse();
        logs.setOutput("combined");
        logs.setStdout("stdout");
        logs.setStderr("stderr");
        when(processApi.getSessionCommand("session-1", "cmd-1")).thenReturn(command);
        when(processApi.getSessionCommandLogs("session-1", "cmd-1", null)).thenReturn(logs);
        when(processApi.getEntrypointLogs(false)).thenReturn(logs);

        io.daytona.sdk.model.Command mappedCommand = process.getSessionCommand("session-1", "cmd-1");
        SessionCommandLogsResponse mappedLogs = process.getSessionCommandLogs("session-1", "cmd-1");
        SessionCommandLogsResponse entrypointLogs = process.getEntrypointLogs();

        assertThat(mappedCommand.getCommand()).isEqualTo("ls");
        assertThat(mappedLogs.getStdout()).isEqualTo("stdout");
        assertThat(entrypointLogs.getStderr()).isEqualTo("stderr");
    }

    @Test
    void websocketLogStreamingDemultiplexesStdoutAndStderr() throws Exception {
        try (MockWebServer server = new MockWebServer()) {
            byte[] payload = new byte[] {1, 1, 1, 'o', 'u', 't', 2, 2, 2, 'e', 'r', 'r'};
            server.enqueue(new MockResponse().withWebSocketUpgrade(new WebSocketListener() {
                @Override
                public void onOpen(WebSocket webSocket, okhttp3.Response response) {
                    webSocket.send(ByteString.of(payload));
                    webSocket.close(1000, "done");
                }
            }));

            Process websocketProcess = new Process(processApi, TestSupport.mockSandbox(server.url("/sandbox").toString()));
            List<String> stdout = new ArrayList<String>();
            List<String> stderr = new ArrayList<String>();

            websocketProcess.getSessionCommandLogs("session-1", "cmd-1", stdout::add, stderr::add);

            assertThat(stdout).containsExactly("out");
            assertThat(stderr).containsExactly("err");
        }
    }

    @Test
    void websocketEntrypointLogStreamingPropagatesFailures() {
        Process brokenProcess = new Process(processApi, TestSupport.mockSandbox("http://127.0.0.1:1/toolbox"));

        assertThatThrownBy(() -> brokenProcess.getEntrypointLogs(value -> { }, value -> { }))
                .hasMessageContaining("Log streaming failed");
    }

    @Test
    void listSessionsHandlesValuesAndNull() {
        io.daytona.toolbox.client.model.Session first = new io.daytona.toolbox.client.model.Session();
        first.setSessionId("s1");
        first.setCommands(Collections.<Command>emptyList());
        when(processApi.listSessions()).thenReturn(Collections.singletonList(first));

        List<Session> sessions = process.listSessions();

        assertThat(sessions).extracting(Session::getSessionId).containsExactly("s1");
        when(processApi.listSessions()).thenReturn(null);
        assertThat(process.listSessions()).isEmpty();
    }

    @Test
    void listPtySessionsAndGetPtySessionInfoHandleNulls() {
        PtySessionInfo info = new PtySessionInfo();
        info.setId("pty-1");
        when(processApi.listPtySessions()).thenReturn(null);
        when(processApi.getPtySession("pty-1")).thenReturn(info);

        assertThat(process.listPtySessions()).isEmpty();
        assertThat(process.getPtySessionInfo("pty-1").getId()).isEqualTo("pty-1");
    }

    @Test
    void createPtyBuildsRequestAndRequiresSessionId() {
        PtyCreateOptions options = new PtyCreateOptions().setId("pty-1").setCols(80).setRows(24);
        when(processApi.createPtySession(any())).thenReturn(new PtyCreateResponse().sessionId("pty-1"));

        PtyHandle handle = process.createPty(options);

        assertThat(handle.getSessionId()).isEqualTo("pty-1");
        ArgumentCaptor<PtyCreateRequest> captor = ArgumentCaptor.forClass(PtyCreateRequest.class);
        verify(processApi).createPtySession(captor.capture());
        assertThat(captor.getValue().getId()).isEqualTo("pty-1");
        assertThat(captor.getValue().getCols()).isEqualTo(80);
        assertThat(captor.getValue().getRows()).isEqualTo(24);
        handle.disconnect();
    }

    @Test
    void createPtyRejectsMissingSessionId() {
        when(processApi.createPtySession(any())).thenReturn(new PtyCreateResponse());

        assertThatThrownBy(() -> process.createPty(new PtyCreateOptions()))
                .hasMessageContaining("Failed to create PTY session");
    }

    @Test
    void connectPtyRejectsMissingToolboxBaseUrl() {
        Process brokenProcess = new Process(processApi, TestSupport.mockSandbox(""));

        assertThatThrownBy(() -> brokenProcess.connectPty("pty-1"))
                .hasMessageContaining("Toolbox base URL is not available");
    }

    @Test
    void ptyMetadataAndActionsDelegate() {
        when(processApi.listPtySessions()).thenReturn(new io.daytona.toolbox.client.model.PtyListResponse().sessions(Collections.<io.daytona.toolbox.client.model.PtySessionInfo>emptyList()));

        process.listPtySessions();
        process.resizePtySession("pty-1", 100, 40);
        process.killPtySession("pty-1");

        verify(processApi).resizePtySession(eq("pty-1"), argThat(request -> request.getCols() == 100 && request.getRows() == 40));
        verify(processApi).deletePtySession("pty-1");
    }

    @ParameterizedTest
    @MethodSource("mappedToolboxExceptions")
    void executeCommandMapsToolboxErrors(int status, Class<? extends RuntimeException> type) {
        when(processApi.executeCommand(any()))
                .thenThrow(new io.daytona.toolbox.client.ApiException(status, "boom", null, "{\"message\":\"mapped\"}"));

        assertThatThrownBy(() -> process.executeCommand("pwd"))
                .isInstanceOf(type)
                .hasMessage("mapped");
    }

    private static Stream<Arguments> mappedToolboxExceptions() {
        return Stream.of(
                Arguments.of(400, DaytonaBadRequestException.class),
                Arguments.of(403, DaytonaForbiddenException.class),
                Arguments.of(404, DaytonaNotFoundException.class),
                Arguments.of(409, DaytonaConflictException.class),
                Arguments.of(429, DaytonaRateLimitException.class),
                Arguments.of(500, DaytonaServerException.class)
        );
    }

    private static <T> T argThat(org.mockito.ArgumentMatcher<T> matcher) {
        return org.mockito.ArgumentMatchers.argThat(matcher);
    }
}
