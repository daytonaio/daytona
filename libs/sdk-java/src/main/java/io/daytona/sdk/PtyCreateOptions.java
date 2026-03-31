// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk;

import java.util.function.Consumer;

public class PtyCreateOptions {
    private String id;
    private int cols = 120;
    private int rows = 30;
    private Consumer<byte[]> onData;

    public PtyCreateOptions() {
    }

    public PtyCreateOptions(String id, int cols, int rows, Consumer<byte[]> onData) {
        this.id = id;
        this.cols = cols;
        this.rows = rows;
        this.onData = onData;
    }

    public String getId() {
        return id;
    }

    public PtyCreateOptions setId(String id) {
        this.id = id;
        return this;
    }

    public int getCols() {
        return cols;
    }

    public PtyCreateOptions setCols(int cols) {
        this.cols = cols;
        return this;
    }

    public int getRows() {
        return rows;
    }

    public PtyCreateOptions setRows(int rows) {
        this.rows = rows;
        return this;
    }

    public Consumer<byte[]> getOnData() {
        return onData;
    }

    public PtyCreateOptions setOnData(Consumer<byte[]> onData) {
        this.onData = onData;
        return this;
    }
}
