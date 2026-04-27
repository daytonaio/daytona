// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk;

import io.daytona.sdk.model.Command;
import io.daytona.sdk.model.CreateSandboxFromImageParams;
import io.daytona.sdk.model.CreateSandboxFromSnapshotParams;
import io.daytona.sdk.model.ExecuteResponse;
import io.daytona.sdk.model.FileInfo;
import io.daytona.sdk.model.GitCommitResponse;
import io.daytona.sdk.model.GitStatus;
import io.daytona.sdk.model.PaginatedSandboxes;
import io.daytona.sdk.model.PaginatedSnapshots;
import io.daytona.sdk.model.Resources;
import io.daytona.sdk.model.Session;
import io.daytona.sdk.model.SessionCommandLogsResponse;
import io.daytona.sdk.model.SessionExecuteRequest;
import io.daytona.sdk.model.SessionExecuteResponse;
import io.daytona.sdk.model.Snapshot;
import io.daytona.sdk.model.Volume;
import io.daytona.sdk.model.VolumeMount;
import io.daytona.sdk.exception.DaytonaException;
import org.junit.jupiter.api.Test;

import java.util.Arrays;
import java.util.Collections;
import java.util.HashMap;
import java.util.Map;

import static org.assertj.core.api.Assertions.assertThat;
import static org.assertj.core.api.Assertions.assertThatThrownBy;

class UtilityAndModelTest {

    @Test
    void runCodeOptionsAreFluent() {
        RunCodeOptions options = new RunCodeOptions()
                .setTimeout(5)
                .setOnStdout(value -> { })
                .setOnStderr(value -> { })
                .setOnError(value -> { });

        assertThat(options.getTimeout()).isEqualTo(5);
        assertThat(options.getOnStdout()).isNotNull();
        assertThat(options.getOnStderr()).isNotNull();
        assertThat(options.getOnError()).isNotNull();
    }

    @Test
    void ptyCreateOptionsExposeDefaultsAndMutators() {
        PtyCreateOptions options = new PtyCreateOptions()
                .setId("pty-1")
                .setCols(200)
                .setRows(50)
                .setOnData(bytes -> { });

        assertThat(options.getId()).isEqualTo("pty-1");
        assertThat(options.getCols()).isEqualTo(200);
        assertThat(options.getRows()).isEqualTo(50);
        assertThat(options.getOnData()).isNotNull();
    }

    @Test
    void ptyCreateOptionsConstructorStoresExplicitValues() {
        PtyCreateOptions options = new PtyCreateOptions("pty-2", 90, 40, bytes -> { });

        assertThat(options.getId()).isEqualTo("pty-2");
        assertThat(options.getCols()).isEqualTo(90);
        assertThat(options.getRows()).isEqualTo(40);
        assertThat(options.getOnData()).isNotNull();
    }

    @Test
    void ptyResultStoresExitCodeAndError() {
        PtyResult result = new PtyResult(7, "boom");

        assertThat(result.getExitCode()).isEqualTo(7);
        assertThat(result.getError()).isEqualTo("boom");
    }

    @Test
    void codeLanguageParsesSupportedValues() {
        assertThat(CodeLanguage.fromValue("python")).isEqualTo(CodeLanguage.PYTHON);
        assertThat(CodeLanguage.fromValue("javascript")).isEqualTo(CodeLanguage.JAVASCRIPT);
        assertThat(CodeLanguage.fromValue("typescript")).isEqualTo(CodeLanguage.TYPESCRIPT);
    }

    @Test
    void codeLanguageRejectsUnsupportedValues() {
        assertThatThrownBy(() -> CodeLanguage.fromValue("ruby"))
                .isInstanceOf(DaytonaException.class)
                .hasMessageContaining("Supported languages: python, javascript, typescript");
    }

    @Test
    void createSandboxParamsModelsStoreValues() {
        Resources resources = new Resources();
        resources.setCpu(2);
        resources.setGpu(1);
        resources.setMemory(4);
        resources.setDisk(8);

        VolumeMount volumeMount = new VolumeMount();
        volumeMount.setVolumeId("vol-1");
        volumeMount.setMountPath("/workspace");

        CreateSandboxFromSnapshotParams snapshotParams = new CreateSandboxFromSnapshotParams();
        snapshotParams.setName("sandbox");
        snapshotParams.setUser("daytona");
        snapshotParams.setLanguage("python");
        snapshotParams.setEnvVars(Collections.singletonMap("A", "1"));
        snapshotParams.setLabels(Collections.singletonMap("team", "sdk"));
        snapshotParams.setPublic(true);
        snapshotParams.setAutoStopInterval(1);
        snapshotParams.setAutoArchiveInterval(2);
        snapshotParams.setAutoDeleteInterval(3);
        snapshotParams.setVolumes(Collections.singletonList(volumeMount));
        snapshotParams.setNetworkBlockAll(true);
        snapshotParams.setSnapshot("snap-1");

        CreateSandboxFromImageParams imageParams = new CreateSandboxFromImageParams();
        imageParams.setImage("python:3.12");
        imageParams.setResources(resources);

        assertThat(snapshotParams.getSnapshot()).isEqualTo("snap-1");
        assertThat(snapshotParams.getVolumes()).containsExactly(volumeMount);
        assertThat(snapshotParams.getNetworkBlockAll()).isTrue();
        assertThat(imageParams.getImage()).isEqualTo("python:3.12");
        assertThat(imageParams.getResources().getDisk()).isEqualTo(8);
    }

    @Test
    void resourcesAndVolumeMountModelsAllowNullDefaults() {
        Resources resources = new Resources();
        VolumeMount mount = new VolumeMount();

        assertThat(resources.getCpu()).isNull();
        assertThat(resources.getGpu()).isNull();
        assertThat(resources.getMemory()).isNull();
        assertThat(resources.getDisk()).isNull();
        assertThat(mount.getVolumeId()).isNull();
        assertThat(mount.getMountPath()).isNull();
    }

    @Test
    void simpleMetadataModelsStoreValues() {
        Snapshot snapshot = new Snapshot();
        snapshot.setId("snap-1");
        snapshot.setName("snapshot");
        snapshot.setImageName("python:3.12");
        snapshot.setState("active");

        Volume volume = new Volume();
        volume.setId("vol-1");
        volume.setName("volume");
        volume.setState("ready");

        GitCommitResponse commitResponse = new GitCommitResponse();
        commitResponse.setHash("abc123");

        assertThat(snapshot.getImageName()).isEqualTo("python:3.12");
        assertThat(volume.getState()).isEqualTo("ready");
        assertThat(commitResponse.getHash()).isEqualTo("abc123");
    }

    @Test
    void paginatedModelsExposeAssignedCollections() {
        PaginatedSandboxes sandboxes = new PaginatedSandboxes();
        Map<String, Object> item = new HashMap<String, Object>();
        item.put("id", "sb-1");
        sandboxes.setItems(Collections.singletonList(item));
        sandboxes.setTotal(1);
        sandboxes.setPage(2);
        sandboxes.setTotalPages(3);

        PaginatedSnapshots snapshots = new PaginatedSnapshots();
        Snapshot snapshot = new Snapshot();
        snapshot.setId("snap-1");
        snapshots.setItems(Collections.singletonList(snapshot));
        snapshots.setTotal(1);
        snapshots.setPage(4);
        snapshots.setTotalPages(5);

        assertThat(sandboxes.getItems()).containsExactly(item);
        assertThat(sandboxes.getTotal()).isEqualTo(1);
        assertThat(sandboxes.getPage()).isEqualTo(2);
        assertThat(sandboxes.getTotalPages()).isEqualTo(3);
        assertThat(snapshots.getItems()).containsExactly(snapshot);
        assertThat(snapshots.getTotal()).isEqualTo(1);
        assertThat(snapshots.getPage()).isEqualTo(4);
        assertThat(snapshots.getTotalPages()).isEqualTo(5);
    }

    @Test
    void paginatedModelsReturnEmptyDefaults() {
        PaginatedSandboxes sandboxes = new PaginatedSandboxes();
        PaginatedSnapshots snapshots = new PaginatedSnapshots();
        GitStatus status = new GitStatus();

        assertThat(sandboxes.getItems()).isEmpty();
        assertThat(sandboxes.getTotal()).isZero();
        assertThat(snapshots.getItems()).isEmpty();
        assertThat(snapshots.getTotalPages()).isZero();
        assertThat(status.getFileStatus()).isEmpty();
        assertThat(status.getAhead()).isZero();
        assertThat(status.getBehind()).isZero();
        assertThat(status.isBranchPublished()).isFalse();
    }

    @Test
    void wrapperModelsCopyToolboxValues() {
        io.daytona.toolbox.client.model.Command sourceCommand = new io.daytona.toolbox.client.model.Command();
        sourceCommand.setId("cmd-1");
        sourceCommand.setCommand("ls");
        sourceCommand.setExitCode(0);

        io.daytona.toolbox.client.model.Session sourceSession = new io.daytona.toolbox.client.model.Session();
        sourceSession.setSessionId("session-1");
        sourceSession.setCommands(Collections.singletonList(sourceCommand));

        io.daytona.toolbox.client.model.ExecuteResponse executeResponse = new io.daytona.toolbox.client.model.ExecuteResponse();
        executeResponse.setExitCode(0);
        executeResponse.setResult("ok");

        io.daytona.toolbox.client.model.CodeRunResponse codeRunResponse = new io.daytona.toolbox.client.model.CodeRunResponse();
        codeRunResponse.setExitCode(1);
        codeRunResponse.setResult("fail");

        io.daytona.toolbox.client.model.FileInfo sourceFile = new io.daytona.toolbox.client.model.FileInfo();
        sourceFile.setName("a.txt");
        sourceFile.setSize(3);
        sourceFile.setMode("644");
        sourceFile.setModTime("now");
        sourceFile.setIsDir(false);

        io.daytona.toolbox.client.model.SessionExecuteResponse sourceExecute = new io.daytona.toolbox.client.model.SessionExecuteResponse();
        sourceExecute.setCmdId("cmd-2");
        sourceExecute.setOutput("output");
        sourceExecute.setStdout("stdout");
        sourceExecute.setStderr("stderr");
        sourceExecute.setExitCode(0);

        io.daytona.toolbox.client.model.SessionCommandLogsResponse sourceLogs = new io.daytona.toolbox.client.model.SessionCommandLogsResponse();
        sourceLogs.setOutput("combined");
        sourceLogs.setStdout("out");
        sourceLogs.setStderr("err");

        SessionExecuteRequest request = new SessionExecuteRequest("pwd", true);
        Command command = new Command(sourceCommand);
        Session session = new Session(sourceSession);
        ExecuteResponse execute = new ExecuteResponse(executeResponse);
        ExecuteResponse codeRun = new ExecuteResponse(codeRunResponse);
        FileInfo fileInfo = new FileInfo(sourceFile);
        SessionExecuteResponse sessionExecuteResponse = new SessionExecuteResponse(sourceExecute);
        SessionCommandLogsResponse logsResponse = SessionCommandLogsResponse.from(sourceLogs);

        assertThat(request.getCommand()).isEqualTo("pwd");
        assertThat(request.getRunAsync()).isTrue();
        assertThat(command.getId()).isEqualTo("cmd-1");
        assertThat(session.getSessionId()).isEqualTo("session-1");
        assertThat(execute.getResult()).isEqualTo("ok");
        assertThat(codeRun.getResult()).isEqualTo("fail");
        assertThat(fileInfo.getName()).isEqualTo("a.txt");
        assertThat(sessionExecuteResponse.getCmdId()).isEqualTo("cmd-2");
        assertThat(logsResponse.getStderr()).isEqualTo("err");
    }

    @Test
    void wrapperModelsHandleNullSources() {
        SessionExecuteRequest request = new SessionExecuteRequest((io.daytona.toolbox.client.model.SessionExecuteRequest) null);
        Command command = new Command(null);
        Session session = new Session(null);
        ExecuteResponse execute = new ExecuteResponse((io.daytona.toolbox.client.model.ExecuteResponse) null);
        ExecuteResponse codeRun = new ExecuteResponse((io.daytona.toolbox.client.model.CodeRunResponse) null);
        FileInfo fileInfo = new FileInfo(null);
        SessionExecuteResponse sessionExecuteResponse = new SessionExecuteResponse(null);
        SessionCommandLogsResponse logsResponse = SessionCommandLogsResponse.from(null);

        assertThat(request.getCommand()).isNull();
        assertThat(command.getId()).isNull();
        assertThat(session.getSessionId()).isNull();
        assertThat(execute.getResult()).isNull();
        assertThat(codeRun.getArtifacts()).isNull();
        assertThat(fileInfo.getName()).isNull();
        assertThat(sessionExecuteResponse.getCmdId()).isNull();
        assertThat(logsResponse.getOutput()).isNull();
    }

    @Test
    void gitStatusFileStatusStoresPathAndStatus() {
        GitStatus.FileStatus fileStatus = new GitStatus.FileStatus();
        fileStatus.setPath("src/Main.java");
        fileStatus.setStatus("modified/untracked");

        GitStatus status = new GitStatus();
        status.setCurrentBranch("main");
        status.setAhead(2);
        status.setBehind(1);
        status.setBranchPublished(true);
        status.setFileStatus(Arrays.asList(fileStatus));

        assertThat(status.getCurrentBranch()).isEqualTo("main");
        assertThat(status.getAhead()).isEqualTo(2);
        assertThat(status.getBehind()).isEqualTo(1);
        assertThat(status.isBranchPublished()).isTrue();
        assertThat(status.getFileStatus()).singleElement().satisfies(item -> {
            assertThat(item.getPath()).isEqualTo("src/Main.java");
            assertThat(item.getStatus()).isEqualTo("modified/untracked");
        });
    }
}
