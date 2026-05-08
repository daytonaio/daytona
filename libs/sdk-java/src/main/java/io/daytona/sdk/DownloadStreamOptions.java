// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk;

import java.util.function.Consumer;

public class DownloadStreamOptions {
    private int timeoutSeconds = 30 * 60;
    private Consumer<DownloadProgress> onProgress;
    private CancellationToken cancellationToken;

    public int getTimeoutSeconds() { return timeoutSeconds; }
    public Consumer<DownloadProgress> getOnProgress() { return onProgress; }
    public CancellationToken getCancellationToken() { return cancellationToken; }

    public DownloadStreamOptions setTimeout(int timeoutSeconds) {
        this.timeoutSeconds = timeoutSeconds;
        return this;
    }

    public DownloadStreamOptions setOnProgress(Consumer<DownloadProgress> onProgress) {
        this.onProgress = onProgress;
        return this;
    }

    public DownloadStreamOptions setCancellationToken(CancellationToken token) {
        this.cancellationToken = token;
        return this;
    }
}
