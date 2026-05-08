// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk;

import java.util.OptionalLong;

/**
 * Progress information for a streaming download.
 */
public final class DownloadProgress {
    private final long bytesReceived;
    private final OptionalLong totalBytes;

    public DownloadProgress(long bytesReceived, OptionalLong totalBytes) {
        this.bytesReceived = bytesReceived;
        this.totalBytes = totalBytes;
    }

    /** Cumulative bytes received so far. */
    public long getBytesReceived() {
        return bytesReceived;
    }

    /** Total bytes expected, if known. */
    public OptionalLong getTotalBytes() {
        return totalBytes;
    }
}
