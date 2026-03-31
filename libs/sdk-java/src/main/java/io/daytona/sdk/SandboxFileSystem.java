// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk;

import io.daytona.sdk.model.FileInfo;
import io.daytona.toolbox.client.api.FileSystemApi;
import io.daytona.toolbox.client.model.ReplaceRequest;

import java.io.File;
import java.io.IOException;
import java.nio.file.Files;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.ArrayList;

public class SandboxFileSystem {
    private final FileSystemApi fileSystemApi;

    SandboxFileSystem(FileSystemApi fileSystemApi) {
        this.fileSystemApi = fileSystemApi;
    }

    public void createFolder(String path, String mode) {
        ExceptionMapper.runToolbox(() -> fileSystemApi.createFolder(path, mode == null ? "755" : mode));
    }

    public void deleteFile(String path) {
        ExceptionMapper.runToolbox(() -> fileSystemApi.deleteFile(path, null));
    }

    public byte[] downloadFile(String remotePath) {
        File file = ExceptionMapper.callToolbox(() -> fileSystemApi.downloadFile(remotePath));
        if (file == null) {
            return new byte[0];
        }
        try {
            byte[] bytes = Files.readAllBytes(file.toPath());
            Files.deleteIfExists(file.toPath());
            return bytes;
        } catch (IOException e) {
            throw new io.daytona.sdk.exception.DaytonaException("Failed to read downloaded file", e);
        }
    }

    public void uploadFile(byte[] content, String remotePath) {
        try {
            File tempFile = File.createTempFile("daytona-upload-", ".tmp");
            Files.write(tempFile.toPath(), content == null ? new byte[0] : content);
            ExceptionMapper.callToolbox(() -> fileSystemApi.uploadFile(remotePath, tempFile));
            Files.deleteIfExists(tempFile.toPath());
        } catch (IOException e) {
            throw new io.daytona.sdk.exception.DaytonaException("Failed to upload file", e);
        }
    }

    public List<FileInfo> listFiles(String path) {
        List<io.daytona.toolbox.client.model.FileInfo> files = ExceptionMapper.callToolbox(() -> fileSystemApi.listFiles(path));
        List<FileInfo> result = new ArrayList<FileInfo>();
        if (files != null) {
            for (io.daytona.toolbox.client.model.FileInfo file : files) {
                result.add(toFileInfo(file));
            }
        }
        return result;
    }

    public FileInfo getFileDetails(String path) {
        io.daytona.toolbox.client.model.FileInfo fileInfo = ExceptionMapper.callToolbox(() -> fileSystemApi.getFileInfo(path));
        return toFileInfo(fileInfo);
    }

    public List<Map<String, Object>> findFiles(String path, String pattern) {
        List<io.daytona.toolbox.client.model.Match> matches = ExceptionMapper.callToolbox(() -> fileSystemApi.findInFiles(path, pattern));
        List<Map<String, Object>> result = new ArrayList<Map<String, Object>>();
        if (matches != null) {
            for (io.daytona.toolbox.client.model.Match match : matches) {
                Map<String, Object> item = new HashMap<String, Object>();
                item.put("content", match.getContent());
                item.put("file", match.getFile());
                item.put("line", match.getLine());
                result.add(item);
            }
        }
        return result;
    }

    public Map<String, Object> searchFiles(String path, String pattern) {
        io.daytona.toolbox.client.model.SearchFilesResponse response = ExceptionMapper.callToolbox(() -> fileSystemApi.searchFiles(path, pattern));
        Map<String, Object> result = new HashMap<String, Object>();
        result.put("files", response == null ? new ArrayList<String>() : response.getFiles());
        return result;
    }

    public void replaceInFiles(List<String> files, String pattern, String newValue) {
        ReplaceRequest request = new ReplaceRequest().files(files).pattern(pattern).newValue(newValue);
        ExceptionMapper.callToolbox(() -> fileSystemApi.replaceInFiles(request));
    }

    public void moveFiles(String source, String destination) {
        ExceptionMapper.runToolbox(() -> fileSystemApi.moveFile(source, destination));
    }

    private FileInfo toFileInfo(io.daytona.toolbox.client.model.FileInfo source) {
        FileInfo fileInfo = new FileInfo();
        if (source != null) {
            fileInfo.setName(source.getName());
            fileInfo.setSize(source.getSize() == null ? null : source.getSize().longValue());
            fileInfo.setMode(source.getMode());
            fileInfo.setModTime(source.getModTime());
            fileInfo.setDir(source.getIsDir());
        }
        return fileInfo;
    }
}