// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.examples;

import io.daytona.sdk.Daytona;
import io.daytona.sdk.DownloadStreamOptions;
import io.daytona.sdk.Sandbox;
import io.daytona.sdk.UploadStreamOptions;
import io.daytona.sdk.model.FileInfo;

import java.io.ByteArrayInputStream;
import java.nio.charset.StandardCharsets;
import java.util.Arrays;
import java.util.List;
import java.util.stream.Collectors;

public class FileOperations {
    public static void main(String[] args) {
        try (Daytona daytona = new Daytona()) {
            System.out.println("Creating sandbox");
            Sandbox sandbox = daytona.create();
            System.out.println("Created sandbox with ID: " + sandbox.getId());

            try {
                System.out.println("Creating folder test-dir");
                sandbox.getFs().createFolder("test-dir", "755");

                System.out.println("Uploading hello.txt");
                sandbox.getFs().uploadFile("Hello, Daytona!".getBytes(StandardCharsets.UTF_8), "test-dir/hello.txt");

                List<FileInfo> files = sandbox.getFs().listFiles("test-dir");
                System.out.println("Files in test-dir: " + files.stream().map(FileInfo::getName).collect(Collectors.joining(", ")));

                System.out.println("Downloading hello.txt");
                byte[] downloaded = sandbox.getFs().downloadFile("test-dir/hello.txt");
                System.out.println("Content: " + new String(downloaded, StandardCharsets.UTF_8));

                // Stream upload — push an InputStream to the Sandbox with live
                // progress reporting. (This example builds the payload in a byte[]
                // for brevity; real-world code should pass a FileInputStream or
                // similar to avoid loading the whole file into memory.)
                System.out.println("Streaming upload streamed.bin with progress");
                byte[] generatedPayload = ("streamed-upload-content-").repeat(2048)
                        .getBytes(StandardCharsets.UTF_8);
                sandbox.getFs().uploadFileStream(
                        new ByteArrayInputStream(generatedPayload),
                        "test-dir/streamed.bin",
                        new UploadStreamOptions().setOnProgress(p -> System.out.println(
                                "  uploaded " + p.getBytesSent() + " / " + generatedPayload.length + " bytes")));

                // Stream download — process file content as chunks arrive, with progress.
                System.out.println("Streaming download hello.txt with progress");
                try (java.io.InputStream stream = sandbox.getFs().downloadFileStream(
                        "test-dir/hello.txt",
                        new DownloadStreamOptions().setOnProgress(p -> System.out.println(
                                "  downloaded " + p.getBytesReceived() + " / "
                                        + p.getTotalBytes().stream().boxed().findFirst().map(String::valueOf).orElse("?")
                                        + " bytes")))) {
                    byte[] streamed = stream.readAllBytes();
                    System.out.println("Streamed content: " + new String(streamed, StandardCharsets.UTF_8));
                } catch (java.io.IOException e) {
                    System.err.println("Stream download failed: " + e.getMessage());
                }

                System.out.println("Uploading config.json and replacing 'true' with 'false'");
                String originalConfig = "{\"debug\": true, \"name\": \"demo\"}";
                sandbox.getFs().uploadFile(originalConfig.getBytes(StandardCharsets.UTF_8), "test-dir/config.json");
                sandbox.getFs().replaceInFiles(
                        Arrays.asList("test-dir/config.json"),
                        "\"debug\": true",
                        "\"debug\": false"
                );
                String replaced = new String(sandbox.getFs().downloadFile("test-dir/config.json"), StandardCharsets.UTF_8);
                System.out.println("Updated config: " + replaced);
            } finally {
                System.out.println("Deleting sandbox");
                sandbox.delete();
            }
        }
    }
}
