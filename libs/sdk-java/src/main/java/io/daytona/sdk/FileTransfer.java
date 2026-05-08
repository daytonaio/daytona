// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk;

import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;
import io.daytona.toolbox.client.ApiClient;
import io.daytona.toolbox.client.api.FileSystemApi;
import io.daytona.toolbox.client.model.FilesDownloadRequest;
import okhttp3.Call;
import okhttp3.MediaType;
import okhttp3.MultipartBody;
import okhttp3.OkHttpClient;
import okhttp3.Request;
import okhttp3.RequestBody;
import okhttp3.Response;
import okhttp3.ResponseBody;
import okio.BufferedSink;

import java.io.BufferedInputStream;
import java.io.ByteArrayOutputStream;
import java.io.FilterInputStream;
import java.io.IOException;
import java.io.InputStream;
import java.nio.charset.StandardCharsets;
import java.util.Collections;
import java.util.HashMap;
import java.util.Locale;
import java.util.Map;
import java.util.OptionalLong;
import java.util.concurrent.TimeUnit;
import java.util.function.Consumer;
import java.util.regex.Matcher;
import java.util.regex.Pattern;

final class FileTransfer {
    static final int DEFAULT_DOWNLOAD_STREAM_TIMEOUT_SECONDS = 30 * 60;
    private static final ObjectMapper OBJECT_MAPPER = new ObjectMapper();
    private static final Pattern BOUNDARY_PATTERN = Pattern.compile("boundary=\"?([^\";]+)\"?");
    private static final Pattern PART_NAME_PATTERN = Pattern.compile("name=\"([^\"]+)\"");

    private FileTransfer() {
    }

    static InputStream streamDownload(FileSystemApi fileSystemApi, String remotePath, DownloadStreamOptions options) throws io.daytona.sdk.exception.DaytonaException {
        int timeoutSeconds = options.getTimeoutSeconds();
        if (timeoutSeconds < 0) {
            throw new io.daytona.sdk.exception.DaytonaException("Timeout must be non-negative");
        }

        ApiClient apiClient = fileSystemApi.getApiClient();
        if (apiClient == null || apiClient.getBasePath() == null || apiClient.getBasePath().isEmpty()) {
            throw new io.daytona.sdk.exception.DaytonaException("Toolbox client is not configured");
        }

        OkHttpClient httpClient = apiClient.getHttpClient();
        if (httpClient == null) {
            throw new io.daytona.sdk.exception.DaytonaException("Toolbox client is not configured");
        }

        OkHttpClient streamingClient = httpClient.newBuilder()
                .readTimeout(timeoutSeconds, TimeUnit.SECONDS)
                .build();

        Request request = buildDownloadFileStreamRequest(apiClient, remotePath);
        Call call = streamingClient.newCall(request);
        CancellationToken cancellationToken = options.getCancellationToken();
        Runnable unregister = cancellationToken != null
                ? cancellationToken.onCancel(call::cancel)
                : () -> {};

        Response response = null;
        boolean handedOff = false;
        try {
            response = call.execute();
            InputStream stream = extractDownloadFileStream(response, options, call);
            Runnable finalUnregister = unregister;
            InputStream wrapped = new java.io.FilterInputStream(stream) {
                private boolean closed = false;

                @Override
                public void close() throws IOException {
                    if (closed) return;
                    closed = true;
                    try {
                        super.close();
                    } finally {
                        finalUnregister.run();
                    }
                }
            };
            handedOff = true;
            return wrapped;
        } catch (IOException e) {
            if (call.isCanceled()) {
                throw new io.daytona.sdk.exception.DaytonaException("Download cancelled");
            }
            throw new io.daytona.sdk.exception.DaytonaException("Failed to download file stream", e);
        } finally {
            // On any failure path (IOException *or* RuntimeException / DaytonaException
            // thrown by extractDownloadFileStream) we must deregister the cancel handler
            // and close the response. On the success path handedOff=true skips this block
            // because the FilterInputStream owns the cleanup.
            if (!handedOff) {
                unregister.run();
                if (response != null) {
                    response.close();
                }
            }
        }
    }

    private static Request buildDownloadFileStreamRequest(ApiClient apiClient, String remotePath) {
        try {
            RequestBody requestBody = apiClient.serialize(
                    new FilesDownloadRequest().paths(Collections.singletonList(remotePath)),
                    "application/json"
            );

            Map<String, String> headerParams = new HashMap<String, String>();
            headerParams.put("Accept", "multipart/form-data");
            headerParams.put("Content-Type", "application/json");

            Request.Builder requestBuilder = new Request.Builder()
                    .url(apiClient.getBasePath().replaceAll("/+$", "") + "/files/bulk-download")
                    .post(requestBody);

            apiClient.processHeaderParams(headerParams, requestBuilder);
            apiClient.processCookieParams(new HashMap<String, String>(), requestBuilder);
            return requestBuilder.build();
        } catch (io.daytona.toolbox.client.ApiException e) {
            throw ExceptionMapper.map(e.getCode(), e.getResponseBody());
        }
    }

    private static InputStream extractDownloadFileStream(Response response, DownloadStreamOptions options, Call call) throws IOException {
        ResponseBody responseBody = response.body();
        if (responseBody == null) {
            response.close();
            throw new io.daytona.sdk.exception.DaytonaException("Download response body is empty");
        }

        if (!response.isSuccessful()) {
            byte[] responseBytes = responseBody.bytes();
            response.close();
            throw parseDownloadError(responseBytes, response.code());
        }

        String boundary = extractBoundary(response.header("Content-Type"));
        if (boundary == null || boundary.isEmpty()) {
            response.close();
            throw new io.daytona.sdk.exception.DaytonaException("Missing multipart boundary in download response");
        }

        BufferedInputStream bufferedStream = new BufferedInputStream(responseBody.byteStream());
        try {
            moveToFirstPart(bufferedStream, boundary);
            Map<String, String> partHeaders = readPartHeaders(bufferedStream);
            String partName = extractPartName(partHeaders.get("content-disposition"));
            long totalBytes = parsePartContentLength(partHeaders.get("content-length"));
            if ("file".equals(partName)) {
                InputStream result = new MultipartPartInputStream(bufferedStream, response, boundary, call, "Download cancelled");
                return withProgress(result, options, totalBytes);
            }

            if ("error".equals(partName)) {
                try (InputStream errorStream = new MultipartPartInputStream(bufferedStream, response, boundary, call, "Download cancelled")) {
                    throw parseDownloadError(errorStream.readAllBytes(), response.code());
                }
            }

            response.close();
            throw new io.daytona.sdk.exception.DaytonaException("File stream not found in download response");
        } catch (IOException e) {
            response.close();
            if (call.isCanceled()) {
                throw new io.daytona.sdk.exception.DaytonaException("Download cancelled");
            }
            throw e;
        } catch (RuntimeException e) {
            response.close();
            throw e;
        }
    }

    private static void moveToFirstPart(BufferedInputStream inputStream, String boundary) throws IOException {
        String expectedBoundary = "--" + boundary;
        String closingBoundary = expectedBoundary + "--";
        while (true) {
            String line = readLine(inputStream);
            if (line == null) {
                throw new io.daytona.sdk.exception.DaytonaException("File stream not found in download response");
            }
            if (expectedBoundary.equals(line)) {
                return;
            }
            if (closingBoundary.equals(line)) {
                throw new io.daytona.sdk.exception.DaytonaException("File stream not found in download response");
            }
        }
    }

    private static Map<String, String> readPartHeaders(BufferedInputStream inputStream) throws IOException {
        Map<String, String> headers = new HashMap<String, String>();
        while (true) {
            String line = readLine(inputStream);
            if (line == null) {
                throw new io.daytona.sdk.exception.DaytonaException("Unexpected end of multipart response");
            }
            if (line.isEmpty()) {
                return headers;
            }
            int separatorIndex = line.indexOf(':');
            if (separatorIndex > 0) {
                headers.put(
                        line.substring(0, separatorIndex).trim().toLowerCase(Locale.ROOT),
                        line.substring(separatorIndex + 1).trim()
                );
            }
        }
    }

    private static String readLine(InputStream inputStream) throws IOException {
        ByteArrayOutputStream buffer = new ByteArrayOutputStream();
        boolean sawCarriageReturn = false;

        while (true) {
            int nextByte = inputStream.read();
            if (nextByte == -1) {
                if (sawCarriageReturn) {
                    buffer.write('\r');
                }
                return buffer.size() == 0 ? null : buffer.toString(StandardCharsets.ISO_8859_1.name());
            }
            if (sawCarriageReturn) {
                if (nextByte == '\n') {
                    return buffer.toString(StandardCharsets.ISO_8859_1.name());
                }
                buffer.write('\r');
                sawCarriageReturn = false;
            }
            if (nextByte == '\r') {
                sawCarriageReturn = true;
                continue;
            }
            if (nextByte == '\n') {
                return buffer.toString(StandardCharsets.ISO_8859_1.name());
            }
            buffer.write(nextByte);
        }
    }

    private static String extractBoundary(String contentType) {
        if (contentType == null) {
            return null;
        }
        Matcher matcher = BOUNDARY_PATTERN.matcher(contentType);
        return matcher.find() ? matcher.group(1) : null;
    }

    private static String extractPartName(String contentDisposition) {
        if (contentDisposition == null) {
            return null;
        }
        Matcher matcher = PART_NAME_PATTERN.matcher(contentDisposition);
        return matcher.find() ? matcher.group(1) : null;
    }

    private static io.daytona.sdk.exception.DaytonaException parseDownloadError(byte[] body, int fallbackStatusCode) {
        String responseBody = body == null ? "" : new String(body, StandardCharsets.UTF_8).trim();
        int statusCode = fallbackStatusCode;

        if (!responseBody.isEmpty()) {
            try {
                JsonNode root = OBJECT_MAPPER.readTree(responseBody);
                JsonNode statusCodeNode = root.get("statusCode");
                if (statusCodeNode != null && statusCodeNode.canConvertToInt()) {
                    statusCode = statusCodeNode.asInt();
                }
            } catch (IOException ignored) {
            }
        }

        if (responseBody.isEmpty()) {
            responseBody = "{\"message\":\"Download failed\"}";
        }

        return ExceptionMapper.map(statusCode, responseBody);
    }

    private static long parsePartContentLength(String contentLengthHeader) {
        if (contentLengthHeader == null) {
            return -1;
        }
        try {
            long parsed = Long.parseLong(contentLengthHeader.trim());
            return parsed >= 0 ? parsed : -1;
        } catch (NumberFormatException e) {
            return -1;
        }
    }

    private static InputStream withProgress(InputStream inputStream, DownloadStreamOptions options, long totalBytes) {
        if (options.getOnProgress() == null) {
            return inputStream;
        }
        return new ProgressInputStream(inputStream, options.getOnProgress(), totalBytes);
    }

    private static final class ProgressInputStream extends FilterInputStream {
        // Coarse cadence for the byte-at-a-time read() overload: emit at most
        // once per 8 KiB. Bulk reads still emit per call (chunks already coalesce
        // bytes), so typical consumers see one event per network chunk.
        private static final long SINGLE_BYTE_REPORT_INTERVAL = 8192;

        private final Consumer<DownloadProgress> onProgress;
        private final OptionalLong totalBytes;
        private long total;
        private long lastReported;

        private ProgressInputStream(InputStream in, Consumer<DownloadProgress> onProgress, long totalBytesValue) {
            super(in);
            this.onProgress = onProgress;
            this.totalBytes = totalBytesValue >= 0 ? OptionalLong.of(totalBytesValue) : OptionalLong.empty();
        }

        @Override
        public int read() throws IOException {
            int b = super.read();
            if (b != -1) {
                total++;
                if (total - lastReported >= SINGLE_BYTE_REPORT_INTERVAL) {
                    emit();
                }
            } else if (total > lastReported) {
                emit();
            }
            return b;
        }

        @Override
        public int read(byte[] b, int off, int len) throws IOException {
            int n = super.read(b, off, len);
            if (n > 0) {
                total += n;
                emit();
            } else if (n < 0 && total > lastReported) {
                emit();
            }
            return n;
        }

        private void emit() {
            onProgress.accept(new DownloadProgress(total, totalBytes));
            lastReported = total;
        }
    }

    private static final class MultipartPartInputStream extends InputStream {
        private static final int BUF_SIZE = 8192;

        private final InputStream source;
        private final Response response;
        private final byte[] delimiter;
        private final Call call;
        private final String cancelledMessage;

        private byte[] buf = new byte[BUF_SIZE];
        private int pos;
        private int limit;
        private int delimiterAt = -1;
        private boolean sourceEnded;
        private boolean finished;
        private boolean closed;

        private MultipartPartInputStream(InputStream source, Response response, String boundary, Call call, String cancelledMessage) {
            this.source = source;
            this.response = response;
            this.delimiter = ("\r\n--" + boundary).getBytes(StandardCharsets.ISO_8859_1);
            this.call = call;
            this.cancelledMessage = cancelledMessage;
        }

        @Override
        public int read() throws IOException {
            byte[] single = new byte[1];
            int n = read(single, 0, 1);
            return n == -1 ? -1 : single[0] & 0xFF;
        }

        @Override
        public int read(byte[] b, int off, int len) throws IOException {
            if (finished || closed) return -1;
            if (b == null) throw new NullPointerException();
            if (off < 0 || len < 0 || len > b.length - off) throw new IndexOutOfBoundsException();
            if (len == 0) return 0;

            while (true) {
                int safe = safeBytes();
                if (safe > 0) {
                    int n = Math.min(len, safe);
                    System.arraycopy(buf, pos, b, off, n);
                    pos += n;
                    return n;
                }

                if (delimiterAt == pos) {
                    finished = true;
                    close();
                    return -1;
                }

                if (sourceEnded && pos >= limit) {
                    finished = true;
                    close();
                    return -1;
                }

                if (!fill()) {
                    int remaining = limit - pos;
                    if (remaining > 0) {
                        int n = Math.min(len, remaining);
                        System.arraycopy(buf, pos, b, off, n);
                        pos += n;
                        return n;
                    }
                    finished = true;
                    close();
                    return -1;
                }
            }
        }

        @Override
        public void close() {
            if (closed) return;
            closed = true;
            response.close();
        }

        private int safeBytes() {
            int available = limit - pos;
            if (available <= 0) return 0;

            if (delimiterAt >= 0) return delimiterAt - pos;

            if (sourceEnded) return available;

            return Math.max(0, available - (delimiter.length - 1));
        }

        private boolean fill() throws IOException {
            compact();
            if (sourceEnded) return limit > 0;

            int read;
            try {
                read = source.read(buf, limit, buf.length - limit);
            } catch (IOException e) {
                if (call.isCanceled()) {
                    throw new io.daytona.sdk.exception.DaytonaException(cancelledMessage);
                }
                throw e;
            }
            if (read == -1) {
                sourceEnded = true;
                return limit > 0;
            }
            limit += read;
            scanForDelimiter();
            return true;
        }

        private void compact() {
            if (pos == 0) return;
            int remaining = limit - pos;
            if (remaining > 0) {
                System.arraycopy(buf, pos, buf, 0, remaining);
            }
            if (delimiterAt >= 0) {
                delimiterAt -= pos;
            }
            limit = remaining;
            pos = 0;
        }

        private void scanForDelimiter() {
            if (delimiterAt >= 0) return;
            int end = limit - delimiter.length + 1;
            for (int i = pos; i < end; i++) {
                if (matchesAt(i)) {
                    delimiterAt = i;
                    return;
                }
            }
        }

        private boolean matchesAt(int offset) {
            for (int j = 0; j < delimiter.length; j++) {
                if (buf[offset + j] != delimiter[j]) return false;
            }
            return true;
        }
    }

    /**
     * Streams an upload to {@code /files/bulk-upload}. The body is constructed as a
     * multipart envelope whose file part is a {@link StreamingRequestBody} — OkHttp
     * pulls bytes from the user's {@link InputStream} via {@code BufferedSink} and
     * writes them straight to the wire, with no intermediate buffering. The wrapper
     * meters bytes as they flow and invokes the optional progress callback.
     *
     * <p>Cancellation flows through OkHttp's call timeout. The daemon owns atomicity
     * (writes to a sibling temp file then renames), so a client-side abort just
     * leaves no destination file at all — partial uploads are never visible.
     */
    static void streamUpload(FileSystemApi fileSystemApi, InputStream source, String remotePath, UploadStreamOptions options) {
        int timeoutSeconds = options.getTimeoutSeconds();
        if (timeoutSeconds < 0) {
            throw new io.daytona.sdk.exception.DaytonaException("Timeout must be non-negative");
        }
        if (source == null) {
            throw new io.daytona.sdk.exception.DaytonaException("Upload source must not be null");
        }
        if (remotePath == null || remotePath.isEmpty()) {
            throw new io.daytona.sdk.exception.DaytonaException("remotePath must not be empty");
        }

        ApiClient apiClient = fileSystemApi.getApiClient();
        if (apiClient == null || apiClient.getBasePath() == null || apiClient.getBasePath().isEmpty()) {
            throw new io.daytona.sdk.exception.DaytonaException("Toolbox client is not configured");
        }

        OkHttpClient httpClient = apiClient.getHttpClient();
        if (httpClient == null) {
            throw new io.daytona.sdk.exception.DaytonaException("Toolbox client is not configured");
        }

        OkHttpClient streamingClient = httpClient.newBuilder()
                .writeTimeout(timeoutSeconds, TimeUnit.SECONDS)
                .callTimeout(timeoutSeconds, TimeUnit.SECONDS)
                .build();

        StreamingRequestBody fileBody = new StreamingRequestBody(source,
                options.getOnProgress(), MediaType.parse("application/octet-stream"));

        MultipartBody multipart = new MultipartBody.Builder()
                .setType(MultipartBody.FORM)
                .addFormDataPart("files[0].path", remotePath)
                .addFormDataPart("files[0].file", remotePath, fileBody)
                .build();

        Request.Builder requestBuilder = new Request.Builder()
                .url(apiClient.getBasePath().replaceAll("/+$", "") + "/files/bulk-upload")
                .post(multipart);
        Map<String, String> headerParams = new HashMap<String, String>();
        headerParams.put("Accept", "*/*");
        apiClient.processHeaderParams(headerParams, requestBuilder);
        apiClient.processCookieParams(new HashMap<String, String>(), requestBuilder);

        Request request = requestBuilder.build();
        Call call = streamingClient.newCall(request);
        CancellationToken cancellationToken = options.getCancellationToken();
        Runnable unregister = cancellationToken != null
                ? cancellationToken.onCancel(call::cancel)
                : () -> {};

        try (Response response = call.execute()) {
            if (!response.isSuccessful()) {
                ResponseBody body = response.body();
                byte[] bodyBytes = body == null ? new byte[0] : body.bytes();
                throw parseDownloadError(bodyBytes, response.code());
            }
        } catch (IOException e) {
            if (call.isCanceled()) {
                throw new io.daytona.sdk.exception.DaytonaException("Upload cancelled");
            }
            throw new io.daytona.sdk.exception.DaytonaException("Failed to upload file stream", e);
        } finally {
            unregister.run();
        }
    }

    /**
     * RequestBody that streams from a caller-provided InputStream straight into OkHttp's
     * BufferedSink. Bytes are forwarded a chunk at a time so heap usage is bounded by
     * the buffer size, not the upload size. When supplied, the progress callback fires
     * once per chunk written.
     *
     * <p>{@link #contentLength()} returns {@code -1} (unknown), which makes OkHttp use
     * chunked transfer encoding. The source's natural EOF terminates the upload — no
     * advance byte count is needed.
     */
    private static final class StreamingRequestBody extends RequestBody {
        private static final int CHUNK_SIZE = 64 * 1024;

        private final InputStream source;
        private final Consumer<UploadProgress> onProgress;
        private final MediaType mediaType;

        private StreamingRequestBody(InputStream source,
                                     Consumer<UploadProgress> onProgress,
                                     MediaType mediaType) {
            this.source = source;
            this.onProgress = onProgress;
            this.mediaType = mediaType;
        }

        @Override
        public MediaType contentType() {
            return mediaType;
        }

        @Override
        public long contentLength() {
            return -1L;
        }

        @Override
        public void writeTo(BufferedSink sink) throws IOException {
            byte[] buf = new byte[CHUNK_SIZE];
            long sent = 0;
            int n;
            while ((n = source.read(buf)) != -1) {
                sink.write(buf, 0, n);
                sent += n;
                if (onProgress != null) {
                    onProgress.accept(new UploadProgress(sent));
                }
            }
        }
    }
}
