// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk;

import java.util.function.Consumer;

public class UploadStreamOptions {
    private int timeoutSeconds = 30 * 60;
    private Consumer<UploadProgress> onProgress;
    private CancellationToken cancellationToken;

    public int getTimeoutSeconds() { return timeoutSeconds; }
    public Consumer<UploadProgress> getOnProgress() { return onProgress; }
    public CancellationToken getCancellationToken() { return cancellationToken; }

    public UploadStreamOptions setTimeout(int timeoutSeconds) {
        this.timeoutSeconds = timeoutSeconds;
        return this;
    }

    public UploadStreamOptions setOnProgress(Consumer<UploadProgress> onProgress) {
        this.onProgress = onProgress;
        return this;
    }

    public UploadStreamOptions setCancellationToken(CancellationToken token) {
        this.cancellationToken = token;
        return this;
    }
}
