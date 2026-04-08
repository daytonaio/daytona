// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.examples;

import io.daytona.sdk.Daytona;
import io.daytona.sdk.Sandbox;
import io.daytona.sdk.model.FileInfo;

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
