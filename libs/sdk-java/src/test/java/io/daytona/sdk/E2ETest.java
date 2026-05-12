// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk;

import io.daytona.api.client.api.SandboxApi;
import io.daytona.api.client.model.PortPreviewUrl;
import io.daytona.api.client.model.SignedPortPreviewUrl;
import io.daytona.sdk.CodeInterpreter.ExecutionResult;
import io.daytona.sdk.exception.DaytonaException;
import io.daytona.sdk.model.CreateSandboxFromImageParams;
import io.daytona.sdk.model.CreateSandboxFromSnapshotParams;
import io.daytona.sdk.model.ExecuteResponse;
import io.daytona.sdk.model.GitCommitResponse;
import io.daytona.sdk.model.GitStatus;
import io.daytona.sdk.model.Session;
import io.daytona.sdk.model.SessionCommandLogsResponse;
import io.daytona.sdk.model.SessionExecuteRequest;
import io.daytona.sdk.model.SessionExecuteResponse;
import io.daytona.sdk.model.Snapshot;
import io.daytona.sdk.model.Volume;
import io.daytona.toolbox.client.api.FileSystemApi;
import io.daytona.toolbox.client.api.GitApi;

import io.daytona.toolbox.client.model.GitBranchRequest;
import io.daytona.toolbox.client.model.GitCheckoutRequest;
import io.daytona.toolbox.client.model.GitDeleteBranchRequest;

import io.daytona.toolbox.client.model.LspSymbol;
import io.daytona.toolbox.client.model.PtySessionInfo;
import io.daytona.toolbox.client.model.ReplaceRequest;
import io.daytona.toolbox.client.model.ReplaceResult;
import org.junit.jupiter.api.AfterAll;
import org.junit.jupiter.api.BeforeAll;
import org.junit.jupiter.api.MethodOrderer;
import org.junit.jupiter.api.Order;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.TestInstance;
import org.junit.jupiter.api.TestMethodOrder;

import java.io.InputStream;
import java.math.BigDecimal;
import java.nio.charset.StandardCharsets;
import java.util.ArrayList;
import java.util.Collections;
import java.util.HashMap;
import java.util.List;
import java.util.Locale;
import java.util.Map;
import java.util.OptionalLong;
import java.util.concurrent.atomic.AtomicReference;

import static org.assertj.core.api.Assertions.assertThat;
import static org.assertj.core.api.Assertions.assertThatCode;
import static org.assertj.core.api.Assertions.assertThatThrownBy;


@TestInstance(TestInstance.Lifecycle.PER_CLASS)
@TestMethodOrder(MethodOrderer.OrderAnnotation.class)
class E2ETest {

    private Daytona daytona;
    private Sandbox sandbox;
    private SandboxApi sandboxApi;
    private FileSystemApi toolboxFileSystemApi;
    private GitApi toolboxGitApi;
    private LspServer lspServer;

    private String sandboxName;
    private String fsDir;
    private String gitRepoPath;
    private String sessionId;
    private String lspProjectDir;
    private String lspFilePath;
    private String ptySessionId;
    private String createdVolumeId;
    private String createdVolumeName;

    @BeforeAll
    void setUp() {
        String apiKey = System.getenv("DAYTONA_API_KEY");
        if (apiKey == null || apiKey.isEmpty()) {
            throw new IllegalStateException("DAYTONA_API_KEY environment variable is required for E2E tests");
        }

        DaytonaConfig config = new DaytonaConfig.Builder()
                .apiKey(apiKey)
                .apiUrl(envOrDefault("DAYTONA_API_URL", "https://app.daytona.io/api"))
                .target(System.getenv("DAYTONA_TARGET"))
                .build();

        daytona = new Daytona(config);
        io.daytona.api.client.ApiClient apiClient = TestSupport.getField(daytona, "apiClient", io.daytona.api.client.ApiClient.class);
        sandboxApi = new SandboxApi(apiClient);

        sandboxName = unique("sdk-java-e2e");
        fsDir = "/tmp/e2e-java-fs";
        gitRepoPath = "/tmp/e2e-java-hello-world";
        sessionId = unique("e2e-session");
        lspProjectDir = "/tmp/e2e-java-lsp";
        lspFilePath = lspProjectDir + "/sample.py";
        createdVolumeName = unique("e2e-vol");

        CreateSandboxFromSnapshotParams params = new CreateSandboxFromSnapshotParams();
        params.setName(sandboxName);
        params.setLanguage("python");
        params.setLabels(Collections.singletonMap("purpose", "e2e-test"));
        sandbox = daytona.create(params, 60);

        toolboxFileSystemApi = new FileSystemApi(sandbox.getToolboxApiClient());
        toolboxGitApi = new GitApi(sandbox.getToolboxApiClient());
    }

    @AfterAll
    void tearDown() {
        if (sandbox == null) {
            return;
        }

        try {
            sandbox.getProcess().deleteSession(sessionId);
        } catch (Exception ignored) {
        }

        try {
            if (ptySessionId != null && !ptySessionId.isEmpty()) {
                sandbox.getProcess().killPtySession(ptySessionId);
            }
        } catch (Exception ignored) {
        }

        try {
            if (lspServer != null) {
                lspServer.stop(LspServer.LspLanguageId.PYTHON.getValue(), lspProjectDir);
            }
        } catch (Exception ignored) {
        }

        try {
            if (createdVolumeId != null) {
                daytona.volume().delete(createdVolumeId);
            }
        } catch (Exception ignored) {
        }

        try {
            sandbox.delete();
        } catch (Exception ignored) {
        }

        try {
            daytona.close();
        } catch (Exception ignored) {
        }
    }

    @Test
    @Order(1)
    void sandboxLifecycleHasIdNameAndOrganizationId() throws Exception {
        io.daytona.api.client.model.Sandbox current = fetchSandboxDto();

        assertThat(sandbox.getId()).isNotBlank();
        assertThat(sandbox.getName()).isEqualTo(sandboxName);
        assertThat(current.getOrganizationId()).isNotBlank();
    }

    @Test
    @Order(2)
    void sandboxLifecycleStateResourcesAndTimestampsAreAvailable() throws Exception {
        io.daytona.api.client.model.Sandbox current = fetchSandboxDto();

        assertThat(sandbox.getState()).isEqualTo("started");
        assertThat(sandbox.getCpu()).isGreaterThan(0);
        assertThat(sandbox.getMemory()).isGreaterThan(0);
        assertThat(sandbox.getDisk()).isGreaterThan(0);
        assertThat(current.getCreatedAt()).isNotNull();
        assertThat(current.getUpdatedAt()).isNotNull();
    }

    @Test
    @Order(3)
    void sandboxLifecycleDirectoriesAndLabelsWork() {
        Map<String, String> labels = new HashMap<String, String>();
        labels.put("test", "e2e");
        labels.put("env", "ci");

        assertThat(sandbox.getUserHomeDir()).contains("/");
        assertThat(sandbox.getWorkDir()).contains("/");
        assertThat(sandbox.setLabels(labels)).containsEntry("test", "e2e").containsEntry("env", "ci");
    }

    @Test
    @Order(4)
    void sandboxLifecycleIntervalsAndRefreshWork() throws Exception {
        sandbox.setAutostopInterval(30);
        sandbox.setAutoArchiveInterval(120);
        sandbox.setAutoDeleteInterval(60);
        assertThat(sandbox.getAutoStopInterval()).isEqualTo(30);
        assertThat(sandbox.getAutoArchiveInterval()).isEqualTo(120);
        assertThat(sandbox.getAutoDeleteInterval()).isEqualTo(60);

        sandbox.setAutoDeleteInterval(-1);
        assertThat(sandbox.getAutoDeleteInterval()).isEqualTo(-1);

        sandbox.refreshData();
        assertThat(sandbox.getId()).isNotBlank();
        assertThat(sandbox.getState()).isEqualTo("started");
    }

    @Test
    @Order(5)
    void fileSystemCreateUploadListDetailsAndDownloadWork() {
        sandbox.getFs().createFolder(fsDir, "755");
        sandbox.getFs().createFolder(fsDir + "/private", "700");
        sandbox.getFs().uploadFile("hello world".getBytes(StandardCharsets.UTF_8), fsDir + "/hello.txt");
        sandbox.getFs().uploadFile("file-a-content".getBytes(StandardCharsets.UTF_8), fsDir + "/a.txt");
        sandbox.getFs().uploadFile("file-b-content".getBytes(StandardCharsets.UTF_8), fsDir + "/b.txt");

        assertThat(sandbox.getFs().listFiles(fsDir)).extracting(io.daytona.sdk.model.FileInfo::getName)
                .contains("hello.txt", "a.txt", "b.txt", "private");
        assertThat(sandbox.getFs().getFileDetails(fsDir + "/hello.txt").getSize()).isEqualTo(11);
        assertThat(new String(sandbox.getFs().downloadFile(fsDir + "/hello.txt"), StandardCharsets.UTF_8)).isEqualTo("hello world");
    }

    @Test
    @Order(6)
    void fileSystemSearchReplacePermissionsMoveDeleteAndNestedOpsWork() {
        assertThat(sandbox.getFs().findFiles(fsDir, "hello")).isNotEmpty();
        assertThat((List<?>) sandbox.getFs().searchFiles(fsDir, "*.txt").get("files")).isNotEmpty();

        sandbox.getFs().uploadFile("foo bar baz".getBytes(StandardCharsets.UTF_8), fsDir + "/replace-me.txt");
        List<ReplaceResult> replaceResults = toolboxFileSystemApi.replaceInFiles(new ReplaceRequest()
                .files(Collections.singletonList(fsDir + "/replace-me.txt"))
                .pattern("foo")
                .newValue("replaced"));
        assertThat(replaceResults).isNotEmpty();
        assertThat(new String(sandbox.getFs().downloadFile(fsDir + "/replace-me.txt"), StandardCharsets.UTF_8))
                .isEqualTo("replaced bar baz");

        sandbox.getFs().uploadFile("script".getBytes(StandardCharsets.UTF_8), fsDir + "/perm-test.txt");
        assertThatCode(() -> toolboxFileSystemApi.setFilePermissions(fsDir + "/perm-test.txt", "daytona", "daytona", "644"))
                .doesNotThrowAnyException();

        sandbox.getFs().uploadFile("moveme".getBytes(StandardCharsets.UTF_8), fsDir + "/to-move.txt");
        sandbox.getFs().moveFiles(fsDir + "/to-move.txt", fsDir + "/moved.txt");
        assertThat(sandbox.getFs().listFiles(fsDir)).extracting(io.daytona.sdk.model.FileInfo::getName)
                .contains("moved.txt").doesNotContain("to-move.txt");

        sandbox.getFs().deleteFile(fsDir + "/moved.txt");
        assertThat(sandbox.getFs().listFiles(fsDir)).extracting(io.daytona.sdk.model.FileInfo::getName)
                .doesNotContain("moved.txt");

        sandbox.getFs().createFolder(fsDir + "/parent/child", "755");
        sandbox.getFs().uploadFile("nested-content".getBytes(StandardCharsets.UTF_8), fsDir + "/parent/child/nested.txt");
        assertThat(sandbox.getFs().listFiles(fsDir + "/parent")).extracting(io.daytona.sdk.model.FileInfo::getName).contains("child");
        assertThat(new String(sandbox.getFs().downloadFile(fsDir + "/parent/child/nested.txt"), StandardCharsets.UTF_8))
                .isEqualTo("nested-content");
    }

    @Test
    @Order(13)
    void fileSystemDownloadFileStreamWorks() throws Exception {
        sandbox.getFs().uploadFile("stream content".getBytes(StandardCharsets.UTF_8), fsDir + "/stream.txt");

        try (InputStream stream = sandbox.getFs().downloadFileStream(fsDir + "/stream.txt")) {
            assertThat(new String(stream.readAllBytes(), StandardCharsets.UTF_8)).isEqualTo("stream content");
        }
    }

    @Test
    @Order(16)
    void downloadFileStreamWithProgress() throws Exception {
        sandbox.getFs().uploadFile("progress test".getBytes(StandardCharsets.UTF_8), fsDir + "/progress.txt");
        AtomicReference<DownloadProgress> lastProgress = new AtomicReference<DownloadProgress>();

        try (InputStream stream = sandbox.getFs().downloadFileStream(
                fsDir + "/progress.txt",
                new DownloadStreamOptions().setOnProgress(lastProgress::set))) {
            byte[] content = stream.readAllBytes();
            assertThat(new String(content, StandardCharsets.UTF_8)).isEqualTo("progress test");
            assertThat(lastProgress.get().getBytesReceived()).isEqualTo(content.length);
            assertThat(lastProgress.get().getTotalBytes()).satisfies(totalBytes -> {
                if (totalBytes.isPresent()) {
                    assertThat(totalBytes.getAsLong()).isEqualTo(content.length);
                }
            });
        }
    }

    @Test
    @Order(17)
    void downloadFileStreamCancellationCancelsRequest() {
        String remotePath = fsDir + "/cancel-stream.bin";
        ExecuteResponse createFile = sandbox.getProcess().executeCommand(
                "python - <<'PY'\n"
                        + "with open('" + remotePath + "', 'wb') as f:\n"
                        + "    chunk = b'x' * (1024 * 1024)\n"
                        + "    for _ in range(16):\n"
                        + "        f.write(chunk)\n"
                        + "PY"
        );
        assertThat(createFile.getExitCode()).isEqualTo(0);

        CancellationToken token = new CancellationToken();
        Thread canceller = cancelAfter(token, 50L);

        try {
            assertThatThrownBy(() -> {
                try (InputStream stream = sandbox.getFs().downloadFileStream(
                        remotePath,
                        new DownloadStreamOptions().setCancellationToken(token))) {
                    byte[] buffer = new byte[1024];
                    while (stream.read(buffer) != -1) {
                        Thread.sleep(10L);
                    }
                }
            })
                    .isInstanceOf(DaytonaException.class)
                    .hasMessageContaining("cancel");
        } finally {
            try {
                canceller.join(1000L);
            } catch (InterruptedException e) {
                Thread.currentThread().interrupt();
            }
        }
    }

    @Test
    @Order(17)
    void uploadFileStreamWithProgress() throws Exception {
        byte[] payload = ("upload-stream-content-" + java.util.UUID.randomUUID()).repeat(512).getBytes(StandardCharsets.UTF_8);
        java.util.List<UploadProgress> updates = new java.util.ArrayList<UploadProgress>();

        sandbox.getFs().uploadFileStream(
                new java.io.ByteArrayInputStream(payload),
                fsDir + "/upload-stream.bin",
                new UploadStreamOptions().setOnProgress(updates::add));

        try (InputStream stream = sandbox.getFs().downloadFileStream(fsDir + "/upload-stream.bin")) {
            byte[] roundTripped = stream.readAllBytes();
            assertThat(roundTripped).isEqualTo(payload);
        }

        assertThat(updates).isNotEmpty();
        UploadProgress last = updates.get(updates.size() - 1);
        assertThat(last.getBytesSent()).isEqualTo(payload.length);
        assertThat(updates).extracting(UploadProgress::getBytesSent).isSorted();
    }

    @Test
    @Order(7)
    void processExecutionCoversCommandsEnvAndFailures() {
        ExecuteResponse echo = sandbox.getProcess().executeCommand("echo hello");
        ExecuteResponse pwd = sandbox.getProcess().executeCommand("pwd", "/tmp", null, null);

        Map<String, String> env = new HashMap<String, String>();
        env.put("MY_VAR", "test123");
        ExecuteResponse envResponse = sandbox.getProcess().executeCommand("echo $MY_VAR", null, env, null);

        Map<String, String> envs = new HashMap<String, String>();
        envs.put("A", "alpha");
        envs.put("B", "beta");
        ExecuteResponse envsResponse = sandbox.getProcess().executeCommand("echo $A $B", null, envs, null);
        ExecuteResponse failure = sandbox.getProcess().executeCommand("exit 42");
        ExecuteResponse stderr = sandbox.getProcess().executeCommand("echo error_msg >&2");

        assertThat(echo.getExitCode()).isEqualTo(0);
        assertThat(echo.getResult()).contains("hello");
        assertThat(pwd.getResult()).contains("/tmp");
        assertThat(envResponse.getResult()).contains("test123");
        assertThat(envsResponse.getResult()).contains("alpha").contains("beta");
        assertThat(failure.getExitCode()).isEqualTo(42);
        assertThat(stderr.getExitCode()).isEqualTo(0);
    }

    @Test
    @Order(8)
    void processExecutionCodeRunCoversHappyPathAndErrors() {
        ExecuteResponse basic = sandbox.getProcess().codeRun("print('hello from python')");
        ExecuteResponse multiline = sandbox.getProcess().codeRun("x = 5\ny = 10\nprint(x + y)");
        ExecuteResponse stderr = sandbox.getProcess().codeRun("import sys; sys.stderr.write('stderr-msg\\n'); print('ok')");
        ExecuteResponse syntaxError = sandbox.getProcess().codeRun("def foo(\nprint('broken')");

        assertThat(basic.getExitCode()).isEqualTo(0);
        assertThat(basic.getResult()).contains("hello from python");
        assertThat(multiline.getResult()).contains("15");
        assertThat(stderr.getExitCode()).isEqualTo(0);
        assertThat(stderr.getResult()).contains("ok");
        assertThat(syntaxError.getExitCode()).isNotEqualTo(0);
    }

    @Test
    @Order(9)
    void sessionManagementCreateGetExecuteAndPersistState() {
        sandbox.getProcess().createSession(sessionId);
        Session session = sandbox.getProcess().getSession(sessionId);
        SessionExecuteResponse response = sandbox.getProcess().executeSessionCommand(sessionId, new SessionExecuteRequest("echo session-hello", false));

        sandbox.getProcess().executeSessionCommand(sessionId, new SessionExecuteRequest("export SESSION_VAR=persistent", false));
        SessionExecuteResponse persisted = sandbox.getProcess().executeSessionCommand(sessionId, new SessionExecuteRequest("echo $SESSION_VAR", false));

        assertThat(session.getSessionId()).isEqualTo(sessionId);
        assertThat(response.getCmdId()).isNotBlank();
        assertThat(persisted.getStdout()).contains("persistent");
    }

    @Test
    @Order(10)
    void sessionManagementLogsListAndDeleteWork() {
        SessionExecuteResponse response = sandbox.getProcess().executeSessionCommand(sessionId, new SessionExecuteRequest("echo logs-test", false));
        SessionCommandLogsResponse logs = sandbox.getProcess().getSessionCommandLogs(sessionId, response.getCmdId());

        assertThat(logs.getStdout()).contains("logs-test");
        assertThat(sandbox.getProcess().listSessions()).extracting(Session::getSessionId).contains(sessionId);

        sandbox.getProcess().deleteSession(sessionId);
        assertThat(sandbox.getProcess().listSessions()).extracting(Session::getSessionId).doesNotContain(sessionId);
    }

    @Test
    @Order(11)
    void gitOperationsCloneStatusBranchesAndBranchLifecycleWork() throws Exception {
        sandbox.getGit().clone("https://github.com/octocat/Hello-World.git", gitRepoPath);

        GitStatus status = sandbox.getGit().status(gitRepoPath);
        assertThat(status.getCurrentBranch()).isNotBlank();
        assertThat((List<?>) sandbox.getGit().branches(gitRepoPath).get("branches")).isNotEmpty();

        String featureBranch = unique("e2e-branch");
        toolboxGitApi.createBranch(new GitBranchRequest().path(gitRepoPath).name(featureBranch));
        toolboxGitApi.checkoutBranch(new GitCheckoutRequest().path(gitRepoPath).branch(featureBranch));
        assertThat(sandbox.getGit().status(gitRepoPath).getCurrentBranch()).isEqualTo(featureBranch);

        sandbox.getFs().uploadFile("e2e git test".getBytes(StandardCharsets.UTF_8), gitRepoPath + "/e2e_file.txt");
        sandbox.getGit().add(gitRepoPath, Collections.singletonList("e2e_file.txt"));
        GitCommitResponse commit = sandbox.getGit().commit(gitRepoPath, "E2E test commit", "E2E Test", "e2e@test.com");
        assertThat(commit.getHash()).isNotBlank();

        toolboxGitApi.checkoutBranch(new GitCheckoutRequest().path(gitRepoPath).branch(status.getCurrentBranch()));
        toolboxGitApi.deleteBranch(new GitDeleteBranchRequest().path(gitRepoPath).name(featureBranch));
        assertThat((List<String>) sandbox.getGit().branches(gitRepoPath).get("branches")).doesNotContain(featureBranch);
    }

    @Test
    @Order(12)
    void codeInterpreterRunCodeWorks() {
        ExecutionResult basic = sandbox.codeInterpreter.runCode("print('interpreter hello')");
        sandbox.codeInterpreter.runCode("ci_value = 42");
        ExecutionResult persisted = sandbox.codeInterpreter.runCode("print(ci_value)");
        ExecutionResult stderr = sandbox.codeInterpreter.runCode("import sys\nsys.stderr.write('ci-stderr\\n')\nprint('ci-ok')");

        assertThat(basic.getStdout()).contains("interpreter hello");
        assertThat(persisted.getStdout()).contains("42");
        assertThat(stderr.getStdout()).contains("ci-ok");
        assertThat(stderr.getStderr()).contains("ci-stderr");
    }

    @Test
    @Order(14)
    void lspServerStartDocumentSymbolsCompletionsAndStopWork() throws Exception {
        sandbox.getFs().createFolder(lspProjectDir, "755");
        sandbox.getFs().uploadFile((
                "class Greeter:\n" +
                "    def greet(self) -> str:\n" +
                "        return 'hello'\n\n" +
                "greeter = Greeter()\n" +
                "greeter.\n"
        ).getBytes(StandardCharsets.UTF_8), lspFilePath);

        lspServer = sandbox.createLspServer(LspServer.LspLanguageId.PYTHON.getValue(), lspProjectDir);
        lspServer.start(LspServer.LspLanguageId.PYTHON.getValue(), lspProjectDir);
        lspServer.didOpen(LspServer.LspLanguageId.PYTHON.getValue(), lspProjectDir, fileUri(lspFilePath));
        Thread.sleep(5000);

        List<LspSymbol> symbols = lspServer.documentSymbols(LspServer.LspLanguageId.PYTHON.getValue(), lspProjectDir, fileUri(lspFilePath));
        List<LspSymbol> workspaceSymbols = lspServer.workspaceSymbols("Greeter", LspServer.LspLanguageId.PYTHON.getValue(), lspProjectDir);

        assertThat(symbols).extracting(LspSymbol::getName).contains("Greeter");
        assertThat(workspaceSymbols).isNotEmpty();

        lspServer.didClose(LspServer.LspLanguageId.PYTHON.getValue(), lspProjectDir, fileUri(lspFilePath));
        lspServer.stop(LspServer.LspLanguageId.PYTHON.getValue(), lspProjectDir);
        lspServer = null;
    }

    @Test
    @Order(15)
    void ptyOperationsCreateListGetResizeConnectAndCloseWork() throws Exception {
        StringBuilder output = new StringBuilder();
        PtyHandle handle = null;

        try {
            ptySessionId = unique("e2e-pty");
            handle = sandbox.getProcess().createPty(new PtyCreateOptions()
                    .setId(ptySessionId)
                    .setCols(80)
                    .setRows(24)
                    .setOnData(bytes -> output.append(new String(bytes, StandardCharsets.UTF_8))));

            assertThat(sandbox.getProcess().listPtySessions()).extracting(PtySessionInfo::getId).contains(ptySessionId);
            assertThat(sandbox.getProcess().getPtySessionInfo(ptySessionId).getId()).isEqualTo(ptySessionId);

            sandbox.getProcess().resizePtySession(ptySessionId, 100, 30);
            PtySessionInfo resized = sandbox.getProcess().getPtySessionInfo(ptySessionId);
            assertThat(resized.getCols()).isEqualTo(100);
            assertThat(resized.getRows()).isEqualTo(30);
        } finally {
            if (handle != null) {
                handle.disconnect();
            }
        }

        PtyHandle connected = sandbox.getProcess().connectPty(ptySessionId, new PtyCreateOptions()
                .setOnData(bytes -> output.append(new String(bytes, StandardCharsets.UTF_8))));
        try {
            connected.sendInput("printf 'pty-output\\n'\n");
            Thread.sleep(2000);
            connected.sendInput("exit\n");
            PtyResult result = connected.waitForExit(10);
            assertThat(result.getExitCode()).isEqualTo(0);
            assertThat(output.toString()).contains("pty-output");
        } finally {
            connected.disconnect();
        }
    }

    @Test
    @Order(17)
    void previewLinksReturnUrlsAndTokens() throws Exception {
        PortPreviewUrl preview = sandboxApi.getPortPreviewUrl(sandbox.getId(), BigDecimal.valueOf(8080), null);
        SignedPortPreviewUrl signed = sandboxApi.getSignedPortPreviewUrl(sandbox.getId(), 8080, null, 60);

        assertThat(preview.getUrl()).contains("http");
        assertThat(preview.getToken()).isNotBlank();
        assertThat(signed.getUrl()).contains("http");
        assertThat(signed.getToken()).isNotBlank();
    }

    @Test
    @Order(18)
    void volumeOperationsCreateListGetAndDeleteWork() throws Exception {
        Volume volume = daytona.volume().create(createdVolumeName);
        createdVolumeId = volume.getId();

        assertThat(volume.getName()).isEqualTo(createdVolumeName);
        assertThat(daytona.volume().list()).extracting(Volume::getId).contains(createdVolumeId);
        assertThat(daytona.volume().getByName(createdVolumeName).getId()).isEqualTo(createdVolumeId);

        waitForVolumeReady(createdVolumeName);
        daytona.volume().delete(createdVolumeId);
        createdVolumeId = null;
        for (int i = 0; i < 15; i++) {
            Thread.sleep(1000);
            List<String> names = daytona.volume().list().stream()
                    .map(Volume::getName).collect(java.util.stream.Collectors.toList());
            if (!names.contains(createdVolumeName)) {
                return;
            }
        }
        List<String> remainingNames = daytona.volume().list().stream()
                .map(Volume::getName).collect(java.util.stream.Collectors.toList());
        assertThat(remainingNames).doesNotContain(createdVolumeName);
    }

    @Test
    @Order(20)
    void additionalErrorHandlingAndProcessPathsWork() {
        ExecuteResponse missingPath = sandbox.getProcess().executeCommand("ls /definitely-missing-e2e-path");
        assertThat(missingPath.getExitCode()).isNotEqualTo(0);

        assertThatThrownBy(() -> sandbox.getFs().downloadFile(fsDir + "/does-not-exist.txt"))
                .isInstanceOf(RuntimeException.class);

        String duplicateSessionId = unique("duplicate-session");
        sandbox.getProcess().createSession(duplicateSessionId);
        try {
            assertThatThrownBy(() -> sandbox.getProcess().createSession(duplicateSessionId))
                    .isInstanceOf(RuntimeException.class);
        } finally {
            sandbox.getProcess().deleteSession(duplicateSessionId);
        }

        try {
            ExecuteResponse timeout = sandbox.getProcess().executeCommand("sleep 2", null, null, 1);
            assertThat(timeout.getExitCode()).isNotEqualTo(0);
        } catch (RuntimeException e) {
            assertThat(messageOf(e).toLowerCase(Locale.ROOT)).contains("timeout");
        }

        ExecuteResponse json = sandbox.getProcess().codeRun(
                "import json\nprint(json.dumps({\"items\": [1, 2, 3], \"meta\": {\"ok\": True}}))"
        );
        assertThat(json.getExitCode()).isEqualTo(0);
        assertThat(json.getResult().trim()).isEqualTo("{\"items\": [1, 2, 3], \"meta\": {\"ok\": true}}");

        ExecuteResponse longRunning = sandbox.getProcess().executeCommand(
                "python - <<'PY'\nimport time\ntime.sleep(1)\nprint('long-run-complete')\nPY",
                null,
                null,
                10
        );
        assertThat(longRunning.getExitCode()).isEqualTo(0);
        assertThat(longRunning.getResult()).contains("long-run-complete");
    }

    @Test
    @Order(21)
    void declarativeImageBuildCreatesSandboxWithBuildLogs() {
        String cacheKey = unique("e2e-build");
        List<String> buildLogs = new ArrayList<String>();
        Image image = Image.debianSlim("3.12")
                .pipInstall("numpy")
                .env(Collections.singletonMap("CACHE_BUSTER", cacheKey));

        CreateSandboxFromImageParams params = new CreateSandboxFromImageParams();
        params.setName(unique("sdk-java-e2e-build"));
        params.setLanguage("python");
        params.setImage(image);

        Sandbox imageSandbox = null;
        try {
            imageSandbox = daytona.create(params, 300, chunk -> {
                if (chunk != null && !chunk.trim().isEmpty()) {
                    buildLogs.add(chunk);
                }
            });

            assertThat(buildLogs).isNotEmpty();
            assertThat(imageSandbox.getState()).isEqualTo("started");

            ExecuteResponse result = imageSandbox.getProcess().executeCommand("python3 -c \"import numpy; print(numpy.__version__)\"");
            assertThat(result.getExitCode()).isEqualTo(0);
            assertThat(result.getResult().trim()).contains(".");
        } finally {
            if (imageSandbox != null) {
                imageSandbox.delete();
            }
        }
    }

    @Test
    @Order(22)
    void sandboxStopStartCycleWorks() {
        sandbox.stop(60);
        sandbox.refreshData();
        if (!"stopped".equals(sandbox.getState())) {
            sandbox.waitUntilStopped(60);
        }
        assertThat(sandbox.getState()).isEqualTo("stopped");

        sandbox.start(60);
        assertThat(sandbox.getState()).isEqualTo("started");

        ExecuteResponse response = sandbox.getProcess().executeCommand("echo restarted");
        assertThat(response.getExitCode()).isEqualTo(0);
        assertThat(response.getResult()).contains("restarted");
    }

    private io.daytona.api.client.model.Sandbox fetchSandboxDto() throws Exception {
        return sandboxApi.getSandbox(sandbox.getId(), null, null);
    }


    private String messageOf(Throwable error) {
        return error == null ? "" : String.valueOf(error.getMessage());
    }

    private Thread cancelAfter(CancellationToken token, long delayMillis) {
        Thread thread = new Thread(() -> {
            try {
                Thread.sleep(delayMillis);
                token.cancel();
            } catch (InterruptedException e) {
                Thread.currentThread().interrupt();
            }
        });
        thread.start();
        return thread;
    }

    private String fileUri(String path) {
        return "file://" + path;
    }

    private void waitForVolumeReady(String volumeName) throws Exception {
        long startedAt = System.currentTimeMillis();
        while ((System.currentTimeMillis() - startedAt) < 15000L) {
            Volume volume = daytona.volume().getByName(volumeName);
            if ("error".equals(volume.getState())) {
                throw new IllegalStateException("Volume entered error state");
            }
            if ("ready".equals(volume.getState())) {
                return;
            }
            Thread.sleep(500L);
        }
        throw new IllegalStateException("Timed out waiting for volume to become ready");
    }

    private String unique(String prefix) {
        return prefix + "-" + System.currentTimeMillis();
    }

    private String envOrDefault(String name, String fallback) {
        String value = System.getenv(name);
        return value == null || value.isEmpty() ? fallback : value;
    }

}
