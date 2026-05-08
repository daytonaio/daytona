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
import okhttp3.mockwebserver.MockResponse;
import okhttp3.mockwebserver.MockWebServer;
import io.daytona.toolbox.client.api.FileSystemApi;
import io.daytona.toolbox.client.model.Match;
import io.daytona.toolbox.client.model.ReplaceRequest;
import io.daytona.toolbox.client.model.SearchFilesResponse;
import okio.Buffer;
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
import java.io.InputStream;
import java.io.IOException;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.util.ArrayList;
import java.util.Arrays;
import java.util.Collections;
import java.util.List;
import java.util.Map;
import java.util.OptionalLong;
import java.util.concurrent.TimeUnit;
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
    void downloadFileStreamReturnsInputStream() throws Exception {
        try (MockWebServer server = new MockWebServer()) {
            server.enqueue(new MockResponse()
                    .setHeader("Content-Type", "multipart/form-data; boundary=DAYTONA-FILE-BOUNDARY")
                    .setBody("--DAYTONA-FILE-BOUNDARY\r\n"
                            + "Content-Disposition: form-data; name=\"file\"; filename=\"remote.txt\"\r\n"
                            + "Content-Type: application/octet-stream\r\n"
                            + "\r\n"
                            + "streamed-content\r\n"
                            + "--DAYTONA-FILE-BOUNDARY--\r\n"));

            io.daytona.toolbox.client.ApiClient apiClient = new io.daytona.toolbox.client.ApiClient();
            apiClient.setBasePath(server.url("/").toString());
            apiClient.addDefaultHeader("Authorization", "Bearer secret");
            FileSystem streamingFileSystem = new FileSystem(new FileSystemApi(apiClient));

            try (InputStream stream = streamingFileSystem.downloadFileStream("/remote.txt", 45)) {
                assertThat(new String(stream.readAllBytes(), StandardCharsets.UTF_8)).isEqualTo("streamed-content");
            }

            okhttp3.mockwebserver.RecordedRequest request = server.takeRequest();
            assertThat(request.getMethod()).isEqualTo("POST");
            assertThat(request.getPath()).isEqualTo("/files/bulk-download");
            assertThat(request.getHeader("Authorization")).isEqualTo("Bearer secret");
            assertThat(request.getHeader("Accept")).isEqualTo("multipart/form-data");
            assertThat(request.getHeader("Content-Type")).startsWith("application/json");
            assertThat(request.getBody().readUtf8()).isEqualTo("{\"paths\":[\"/remote.txt\"]}");
        }
    }

    @Test
    void downloadFileStreamWithProgressCallsCallback() throws Exception {
        byte[] content = "hello progress".getBytes(StandardCharsets.UTF_8);
        String multipart = "--DAYTONA-FILE-BOUNDARY\r\n"
                + "Content-Disposition: form-data; name=\"file\"; filename=\"remote.txt\"\r\n"
                + "Content-Type: application/octet-stream\r\n"
                + "Content-Length: " + content.length + "\r\n"
                + "\r\n"
                + new String(content, StandardCharsets.UTF_8)
                + "\r\n--DAYTONA-FILE-BOUNDARY--\r\n";

        try (MockWebServer server = new MockWebServer()) {
            server.enqueue(new MockResponse()
                    .setHeader("Content-Type", "multipart/form-data; boundary=DAYTONA-FILE-BOUNDARY")
                    .setBody(multipart));

            io.daytona.toolbox.client.ApiClient apiClient = new io.daytona.toolbox.client.ApiClient();
            apiClient.setBasePath(server.url("/").toString());
            FileSystem streamingFileSystem = new FileSystem(new FileSystemApi(apiClient));
            List<DownloadProgress> progressUpdates = new ArrayList<DownloadProgress>();

            try (InputStream stream = streamingFileSystem.downloadFileStream(
                    "/remote.txt",
                    new DownloadStreamOptions().setOnProgress(progressUpdates::add))) {
                byte[] buffer = new byte[4];
                int read;
                StringBuilder received = new StringBuilder();
                while ((read = stream.read(buffer)) != -1) {
                    received.append(new String(buffer, 0, read, StandardCharsets.UTF_8));
                }

                assertThat(received.toString()).isEqualTo(new String(content, StandardCharsets.UTF_8));
            }

            assertThat(progressUpdates)
                    .extracting(DownloadProgress::getBytesReceived)
                    .isEqualTo(Arrays.asList(4L, 8L, 12L, (long) content.length));
            assertThat(progressUpdates)
                    .extracting(DownloadProgress::getTotalBytes)
                    .containsOnly(OptionalLong.of(content.length));
        }
    }

    @Test
    void downloadFileStreamSingleByteReadThrottlesProgressEvents() throws Exception {
        // 9216 bytes -> with the 8 KiB throttle on single-byte read(), the
        // callback should fire once when total crosses 8192 and once on EOF.
        byte[] content = new byte[9216];
        java.util.Arrays.fill(content, (byte) 'a');
        String multipart = "--DAYTONA-FILE-BOUNDARY\r\n"
                + "Content-Disposition: form-data; name=\"file\"; filename=\"remote.txt\"\r\n"
                + "Content-Type: application/octet-stream\r\n"
                + "Content-Length: " + content.length + "\r\n"
                + "\r\n"
                + new String(content, StandardCharsets.UTF_8)
                + "\r\n--DAYTONA-FILE-BOUNDARY--\r\n";

        try (MockWebServer server = new MockWebServer()) {
            server.enqueue(new MockResponse()
                    .setHeader("Content-Type", "multipart/form-data; boundary=DAYTONA-FILE-BOUNDARY")
                    .setBody(multipart));

            io.daytona.toolbox.client.ApiClient apiClient = new io.daytona.toolbox.client.ApiClient();
            apiClient.setBasePath(server.url("/").toString());
            FileSystem streamingFileSystem = new FileSystem(new FileSystemApi(apiClient));
            List<DownloadProgress> progressUpdates = new ArrayList<DownloadProgress>();

            try (InputStream stream = streamingFileSystem.downloadFileStream(
                    "/remote.txt",
                    new DownloadStreamOptions().setOnProgress(progressUpdates::add))) {
                int totalRead = 0;
                while (stream.read() != -1) {
                    totalRead++;
                }
                assertThat(totalRead).isEqualTo(content.length);
            }

            assertThat(progressUpdates)
                    .extracting(DownloadProgress::getBytesReceived)
                    .isEqualTo(Arrays.asList(8192L, (long) content.length));
        }
    }

    @Test
    void downloadFileStreamCancellationCancelsRequest() throws Exception {
        byte[] content = new byte[256 * 1024];
        Arrays.fill(content, (byte) 'a');
        Buffer multipart = new Buffer()
                .writeUtf8("--DAYTONA-FILE-BOUNDARY\r\n")
                .writeUtf8("Content-Disposition: form-data; name=\"file\"; filename=\"remote.txt\"\r\n")
                .writeUtf8("Content-Type: application/octet-stream\r\n")
                .writeUtf8("\r\n")
                .write(content)
                .writeUtf8("\r\n--DAYTONA-FILE-BOUNDARY--\r\n");

        try (MockWebServer server = new MockWebServer()) {
            server.enqueue(new MockResponse()
                    .setHeader("Content-Type", "multipart/form-data; boundary=DAYTONA-FILE-BOUNDARY")
                    .setBody(multipart)
                    .throttleBody(1024, 25, TimeUnit.MILLISECONDS));

            io.daytona.toolbox.client.ApiClient apiClient = new io.daytona.toolbox.client.ApiClient();
            apiClient.setBasePath(server.url("/").toString());
            FileSystem streamingFileSystem = new FileSystem(new FileSystemApi(apiClient));
            CancellationToken token = new CancellationToken();
            Thread canceller = cancelAfter(token, 50L);

            try (InputStream stream = streamingFileSystem.downloadFileStream(
                    "/remote.txt",
                    new DownloadStreamOptions().setCancellationToken(token))) {
                assertThatThrownBy(stream::readAllBytes)
                        .isInstanceOf(DaytonaException.class)
                        .hasMessageContaining("cancel");
            } finally {
                canceller.join(1000L);
            }
        }
    }

    @Test
    void downloadFileStreamThrowsOnErrorPart() throws Exception {
        try (MockWebServer server = new MockWebServer()) {
            server.enqueue(new MockResponse()
                    .setHeader("Content-Type", "multipart/form-data; boundary=DAYTONA-FILE-BOUNDARY")
                    .setBody("--DAYTONA-FILE-BOUNDARY\r\n"
                            + "Content-Disposition: form-data; name=\"error\"\r\n"
                            + "Content-Type: application/json\r\n"
                            + "\r\n"
                            + "{\"message\":\"missing file\",\"statusCode\":404,\"code\":\"NotFound\"}\r\n"
                            + "--DAYTONA-FILE-BOUNDARY--\r\n"));

            io.daytona.toolbox.client.ApiClient apiClient = new io.daytona.toolbox.client.ApiClient();
            apiClient.setBasePath(server.url("/").toString());
            FileSystem streamingFileSystem = new FileSystem(new FileSystemApi(apiClient));

            assertThatThrownBy(() -> streamingFileSystem.downloadFileStream("/missing.txt"))
                    .isInstanceOf(DaytonaNotFoundException.class)
                    .hasMessage("missing file");
        }
    }

    @Test
    void uploadFileStreamSendsBytesAndFiresProgress() throws Exception {
        byte[] payload = ("upload-stream-content-").repeat(8).getBytes(StandardCharsets.UTF_8);

        try (MockWebServer server = new MockWebServer()) {
            server.enqueue(new MockResponse().setResponseCode(200));

            io.daytona.toolbox.client.ApiClient apiClient = new io.daytona.toolbox.client.ApiClient();
            apiClient.setBasePath(server.url("/").toString());
            FileSystem streamingFileSystem = new FileSystem(new FileSystemApi(apiClient));
            List<UploadProgress> updates = new ArrayList<UploadProgress>();

            streamingFileSystem.uploadFileStream(
                    new java.io.ByteArrayInputStream(payload),
                    "/remote.txt",
                    new UploadStreamOptions().setOnProgress(updates::add));

            // The recorded request body should contain both multipart parts and the
            // file bytes verbatim — proves we actually streamed content to the server.
            okhttp3.mockwebserver.RecordedRequest request = server.takeRequest();
            assertThat(request.getMethod()).isEqualTo("POST");
            assertThat(request.getPath()).endsWith("/files/bulk-upload");
            String contentType = request.getHeader("Content-Type");
            assertThat(contentType).startsWith("multipart/form-data");
            byte[] bodyBytes = request.getBody().readByteArray();
            assertThat(new String(bodyBytes, StandardCharsets.UTF_8)).contains("name=\"files[0].path\"");
            assertThat(new String(bodyBytes, StandardCharsets.UTF_8)).contains("name=\"files[0].file\"");
            assertThat(indexOfBytes(bodyBytes, payload)).isGreaterThanOrEqualTo(0);

            assertThat(updates).isNotEmpty();
            UploadProgress last = updates.get(updates.size() - 1);
            assertThat(last.getBytesSent()).isEqualTo(payload.length);
            assertThat(updates).extracting(UploadProgress::getBytesSent).isSorted();
        }
    }

    @Test
    void uploadFileStreamCancellationCancelsRequest() throws Exception {
        byte[] payload = new byte[256 * 1024];
        Arrays.fill(payload, (byte) 'b');

        try (MockWebServer server = new MockWebServer()) {
            server.enqueue(new MockResponse().setResponseCode(200));

            io.daytona.toolbox.client.ApiClient apiClient = new io.daytona.toolbox.client.ApiClient();
            apiClient.setBasePath(server.url("/").toString());
            FileSystem streamingFileSystem = new FileSystem(new FileSystemApi(apiClient));
            CancellationToken token = new CancellationToken();
            Thread canceller = cancelAfter(token, 50L);

            try {
                assertThatThrownBy(() -> streamingFileSystem.uploadFileStream(
                        new SlowByteArrayInputStream(payload, 1024, 20L),
                        "/remote.txt",
                        new UploadStreamOptions().setCancellationToken(token)))
                        .isInstanceOf(DaytonaException.class)
                        .hasMessageContaining("cancel");
            } finally {
                canceller.join(1000L);
            }
        }
    }

    private static Thread cancelAfter(CancellationToken token, long delayMillis) {
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

    private static final class SlowByteArrayInputStream extends java.io.ByteArrayInputStream {
        private final int maxChunkSize;
        private final long delayMillis;

        private SlowByteArrayInputStream(byte[] buf, int maxChunkSize, long delayMillis) {
            super(buf);
            this.maxChunkSize = maxChunkSize;
            this.delayMillis = delayMillis;
        }

        @Override
        public synchronized int read(byte[] b, int off, int len) {
            sleepQuietly();
            return super.read(b, off, Math.min(len, maxChunkSize));
        }

        @Override
        public synchronized int read() {
            sleepQuietly();
            return super.read();
        }

        private void sleepQuietly() {
            try {
                Thread.sleep(delayMillis);
            } catch (InterruptedException e) {
                Thread.currentThread().interrupt();
            }
        }
    }

    private static int indexOfBytes(byte[] haystack, byte[] needle) {
        outer:
        for (int i = 0; i <= haystack.length - needle.length; i++) {
            for (int j = 0; j < needle.length; j++) {
                if (haystack[i + j] != needle[j]) continue outer;
            }
            return i;
        }
        return -1;
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
