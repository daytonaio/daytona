// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk;

import io.daytona.sdk.exception.DaytonaException;
import io.daytona.sdk.exception.DaytonaBadRequestException;
import io.daytona.sdk.exception.DaytonaConflictException;
import io.daytona.sdk.exception.DaytonaForbiddenException;
import io.daytona.sdk.exception.DaytonaNotFoundException;
import io.daytona.sdk.exception.DaytonaRateLimitException;
import io.daytona.sdk.exception.DaytonaServerException;
import io.daytona.sdk.model.FileInfo;
import io.daytona.toolbox.client.api.FileSystemApi;
import io.daytona.toolbox.client.model.Match;
import io.daytona.toolbox.client.model.ReplaceRequest;
import io.daytona.toolbox.client.model.SearchFilesResponse;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.junit.jupiter.params.ParameterizedTest;
import org.junit.jupiter.params.provider.Arguments;
import org.junit.jupiter.params.provider.MethodSource;
import org.mockito.ArgumentCaptor;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import java.io.File;
import java.io.IOException;
import java.nio.file.Files;
import java.util.Arrays;
import java.util.Collections;
import java.util.List;
import java.util.Map;
import java.util.stream.Stream;

import static org.assertj.core.api.Assertions.assertThat;
import static org.assertj.core.api.Assertions.assertThatThrownBy;
import static org.mockito.ArgumentMatchers.any;
import static org.mockito.ArgumentMatchers.eq;
import static org.mockito.Mockito.doAnswer;
import static org.mockito.Mockito.verify;
import static org.mockito.Mockito.when;

@ExtendWith(MockitoExtension.class)
class FileSystemTest {

    @Mock
    private FileSystemApi fileSystemApi;

    private FileSystem fileSystem;

    @BeforeEach
    void setUp() {
        fileSystem = new FileSystem(fileSystemApi);
    }

    @Test
    void createFolderUsesDefaultModeWhenMissing() {
        fileSystem.createFolder("/workspace", null);

        verify(fileSystemApi).createFolder("/workspace", "755");
    }

    @Test
    void createFolderUsesExplicitModeWhenProvided() {
        fileSystem.createFolder("/workspace", "700");

        verify(fileSystemApi).createFolder("/workspace", "700");
    }

    @Test
    void uploadFileWritesBytesAndCleansUpTempFile() throws IOException {
        ArgumentCaptor<File> fileCaptor = ArgumentCaptor.forClass(File.class);
        final byte[][] uploadedBytes = new byte[1][];
        doAnswer(invocation -> {
            File file = invocation.getArgument(1);
            uploadedBytes[0] = Files.readAllBytes(file.toPath());
            return null;
        }).when(fileSystemApi).uploadFile(eq("/remote.txt"), any(File.class));

        fileSystem.uploadFile("hello".getBytes(), "/remote.txt");

        verify(fileSystemApi).uploadFile(eq("/remote.txt"), fileCaptor.capture());
        File uploaded = fileCaptor.getValue();
        assertThat(uploadedBytes[0]).isEqualTo("hello".getBytes());
        assertThat(uploaded.exists()).isFalse();
    }

    @Test
    void uploadFileTreatsNullContentAsEmptyFile() throws IOException {
        final byte[][] uploadedBytes = new byte[1][];
        doAnswer(invocation -> {
            File file = invocation.getArgument(1);
            uploadedBytes[0] = Files.readAllBytes(file.toPath());
            return null;
        }).when(fileSystemApi).uploadFile(eq("/empty.txt"), any(File.class));

        fileSystem.uploadFile(null, "/empty.txt");

        assertThat(uploadedBytes[0]).isEmpty();
    }

    @Test
    void downloadFileReadsAndDeletesTempFile() throws IOException {
        File tempFile = File.createTempFile("daytona-test-download", ".txt");
        Files.write(tempFile.toPath(), "payload".getBytes());
        when(fileSystemApi.downloadFile("/remote.txt")).thenReturn(tempFile);

        byte[] bytes = fileSystem.downloadFile("/remote.txt");

        assertThat(bytes).isEqualTo("payload".getBytes());
        assertThat(tempFile.exists()).isFalse();
    }

    @Test
    void downloadFileWrapsIoErrors() throws IOException {
        File directory = Files.createTempDirectory("daytona-test-download-dir").toFile();
        when(fileSystemApi.downloadFile("/remote.txt")).thenReturn(directory);

        assertThatThrownBy(() -> fileSystem.downloadFile("/remote.txt"))
                .isInstanceOf(DaytonaException.class)
                .hasMessageContaining("Failed to read downloaded file");
    }

    @Test
    void downloadFileReturnsEmptyArrayWhenApiReturnsNull() {
        when(fileSystemApi.downloadFile("/missing.txt")).thenReturn(null);

        assertThat(fileSystem.downloadFile("/missing.txt")).isEmpty();
    }

    @Test
    void listAndDetailsMapToolboxModels() {
        io.daytona.toolbox.client.model.FileInfo source = new io.daytona.toolbox.client.model.FileInfo();
        source.setName("a.txt");
        source.setSize(12);
        source.setMode("644");
        source.setModTime("now");
        source.setIsDir(false);
        when(fileSystemApi.listFiles("/workspace")).thenReturn(Collections.singletonList(source));
        when(fileSystemApi.getFileInfo("/workspace/a.txt")).thenReturn(source);

        List<FileInfo> files = fileSystem.listFiles("/workspace");
        FileInfo details = fileSystem.getFileDetails("/workspace/a.txt");

        assertThat(files).singleElement().extracting(FileInfo::getName).isEqualTo("a.txt");
        assertThat(details.getSize()).isEqualTo(12);
    }

    @Test
    void listAndDetailsHandleNullResponses() {
        when(fileSystemApi.listFiles("/workspace")).thenReturn(null);
        when(fileSystemApi.getFileInfo("/workspace/missing.txt")).thenReturn(null);

        assertThat(fileSystem.listFiles("/workspace")).isEmpty();
        assertThat(fileSystem.getFileDetails("/workspace/missing.txt").getName()).isNull();
    }

    @Test
    void findAndSearchMapResponses() {
        Match match = new Match();
        match.setContent("needle");
        match.setFile("/workspace/a.txt");
        match.setLine(7);
        when(fileSystemApi.findInFiles("/workspace", "needle")).thenReturn(Collections.singletonList(match));
        when(fileSystemApi.searchFiles("/workspace", "*.java")).thenReturn(new SearchFilesResponse().files(Arrays.asList("A.java", "B.java")));

        List<Map<String, Object>> matches = fileSystem.findFiles("/workspace", "needle");
        Map<String, Object> search = fileSystem.searchFiles("/workspace", "*.java");

        assertThat(matches).singleElement().satisfies(item -> {
            assertThat(item).containsEntry("content", "needle");
            assertThat(item).containsEntry("file", "/workspace/a.txt");
            assertThat(item).containsEntry("line", 7);
        });
        assertThat(search).containsEntry("files", Arrays.asList("A.java", "B.java"));
    }

    @Test
    void findAndSearchReturnEmptyCollectionsForNullResponses() {
        when(fileSystemApi.findInFiles("/workspace", "needle")).thenReturn(null);
        when(fileSystemApi.searchFiles("/workspace", "*.java")).thenReturn(null);

        assertThat(fileSystem.findFiles("/workspace", "needle")).isEmpty();
        assertThat(fileSystem.searchFiles("/workspace", "*.java")).containsEntry("files", Collections.emptyList());
    }

    @Test
    void replaceDeleteAndMoveDelegate() {
        fileSystem.replaceInFiles(Arrays.asList("A.java", "B.java"), "old", "new");
        fileSystem.deleteFile("/workspace/a.txt");
        fileSystem.moveFiles("/workspace/a.txt", "/workspace/b.txt");

        verify(fileSystemApi).replaceInFiles(new ReplaceRequest().files(Arrays.asList("A.java", "B.java")).pattern("old").newValue("new"));
        verify(fileSystemApi).deleteFile("/workspace/a.txt", null);
        verify(fileSystemApi).moveFile("/workspace/a.txt", "/workspace/b.txt");
    }

    @ParameterizedTest
    @MethodSource("mappedToolboxExceptions")
    void createFolderMapsToolboxErrors(int status, Class<? extends RuntimeException> type) {
        org.mockito.Mockito.doThrow(new io.daytona.toolbox.client.ApiException(status, "boom", null, "{\"message\":\"mapped\"}"))
                .when(fileSystemApi).createFolder("/workspace", "755");

        assertThatThrownBy(() -> fileSystem.createFolder("/workspace", null))
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
}
