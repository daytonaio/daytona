// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk;

import io.daytona.sdk.model.*;
import io.daytona.sdk.exception.DaytonaException;
import io.daytona.toolbox.client.api.ProcessApi;
import io.daytona.toolbox.client.model.CreateSessionRequest;
import io.daytona.toolbox.client.model.ExecuteRequest;
import io.daytona.toolbox.client.model.PtyCreateRequest;
import io.daytona.toolbox.client.model.PtyCreateResponse;
import io.daytona.toolbox.client.model.PtyResizeRequest;
import okhttp3.Request;

import java.util.ArrayList;
import java.util.List;
import java.util.Map;

public class SandboxProcess {
    private final ProcessApi processApi;
    private final Sandbox sandbox;

    SandboxProcess(ProcessApi processApi, Sandbox sandbox) {
        this.processApi = processApi;
        this.sandbox = sandbox;
    }

    public ExecuteResponse executeCommand(String command) {
        return executeCommand(command, null, null, null);
    }

    public ExecuteResponse executeCommand(String command, String cwd, Map<String, String> env, Integer timeout) {
        ExecuteRequest request = new ExecuteRequest().command(command);
        if (cwd != null) {
            request.cwd(cwd);
        }
        if (timeout != null) {
            request.timeout(timeout);
        }
        io.daytona.toolbox.client.model.ExecuteResponse response = ExceptionMapper.callToolbox(() -> processApi.executeCommand(request));
        return toExecuteResponse(response);
    }

    public ExecuteResponse codeRun(String code) {
        String lang = "python";
        if (sandbox.getLabels() != null && sandbox.getLabels().containsKey("code-toolbox-language")) {
            lang = sandbox.getLabels().get("code-toolbox-language");
        }

        String scriptPath;
        String runCmd;
        if ("typescript".equalsIgnoreCase(lang)) {
            scriptPath = "/tmp/daytona_code.ts";
            runCmd = "npx ts-node " + scriptPath;
        } else if ("javascript".equalsIgnoreCase(lang)) {
            scriptPath = "/tmp/daytona_code.js";
            runCmd = "node " + scriptPath;
        } else {
            scriptPath = "/tmp/daytona_code.py";
            runCmd = "python3 " + scriptPath;
        }

        String command = "cat <<'DAYTONA_EOF' > " + scriptPath + "\n"
                + code + "\nDAYTONA_EOF\n"
                + runCmd;

        return executeCommand(command);
    }

    public void createSession(String sessionId) {
        ExceptionMapper.runToolbox(() -> processApi.createSession(new CreateSessionRequest().sessionId(sessionId)));
    }

    public Session getSession(String sessionId) {
        io.daytona.toolbox.client.model.Session session = ExceptionMapper.callToolbox(() -> processApi.getSession(sessionId));
        return toSession(session);
    }

    public SessionExecuteResponse executeSessionCommand(String sessionId, SessionExecuteRequest req) {
        io.daytona.toolbox.client.model.SessionExecuteRequest request = new io.daytona.toolbox.client.model.SessionExecuteRequest()
                .command(req.getCommand())
                .runAsync(req.getRunAsync());
        io.daytona.toolbox.client.model.SessionExecuteResponse response = ExceptionMapper.callToolbox(() -> processApi.sessionExecuteCommand(sessionId, request));
        return toSessionExecuteResponse(response);
    }

    public Command getSessionCommand(String sessionId, String commandId) {
        io.daytona.toolbox.client.model.Command command = ExceptionMapper.callToolbox(() -> processApi.getSessionCommand(sessionId, commandId));
        return toCommand(command);
    }

    public String getSessionCommandLogs(String sessionId, String commandId) {
        return ExceptionMapper.callToolbox(() -> processApi.getSessionCommandLogs(sessionId, commandId, null));
    }

    public void deleteSession(String sessionId) {
        ExceptionMapper.runToolbox(() -> processApi.deleteSession(sessionId));
    }

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

        String sessionId = response.getSessionId();
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
                createOptions.getOnData()
        );
    }

    public void resizePtySession(String sessionId, int cols, int rows) {
        ExceptionMapper.callToolbox(() -> processApi.resizePtySession(
                sessionId,
                new PtyResizeRequest().cols(cols).rows(rows)
        ));
    }

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

    private ExecuteResponse toExecuteResponse(io.daytona.toolbox.client.model.ExecuteResponse source) {
        ExecuteResponse response = new ExecuteResponse();
        if (source != null) {
            response.setExitCode(source.getExitCode());
            response.setResult(source.getResult());
        }
        return response;
    }

    private Session toSession(io.daytona.toolbox.client.model.Session source) {
        Session session = new Session();
        if (source == null) {
            return session;
        }
        session.setSessionId(source.getSessionId());
        List<Command> commands = new ArrayList<Command>();
        if (source.getCommands() != null) {
            for (io.daytona.toolbox.client.model.Command command : source.getCommands()) {
                commands.add(toCommand(command));
            }
        }
        session.setCommands(commands);
        return session;
    }

    private Command toCommand(io.daytona.toolbox.client.model.Command source) {
        Command command = new Command();
        if (source != null) {
            command.setId(source.getId());
            command.setCommand(source.getCommand());
            command.setExitCode(source.getExitCode());
        }
        return command;
    }

    private SessionExecuteResponse toSessionExecuteResponse(io.daytona.toolbox.client.model.SessionExecuteResponse source) {
        SessionExecuteResponse response = new SessionExecuteResponse();
        if (source != null) {
            response.setCmdId(source.getCmdId());
            response.setOutput(source.getOutput());
            response.setExitCode(source.getExitCode());
        }
        return response;
    }
}
