// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk;

import java.util.function.Consumer;

/**
 * Options used when creating a PTY session in a Sandbox.
 */
public class PtyCreateOptions {
    private String id;
    private int cols = 120;
    private int rows = 30;
    private Consumer<byte[]> onData;

    /**
     * Creates PTY options with default dimensions ({@code 120x30}).
     */
    public PtyCreateOptions() {
    }

    /**
     * Creates PTY options with explicit values.
     *
     * @param id custom PTY session identifier; if {@code null}, the server generates one
     * @param cols terminal width in columns
     * @param rows terminal height in rows
     * @param onData callback invoked for each PTY output chunk
     */
    public PtyCreateOptions(String id, int cols, int rows, Consumer<byte[]> onData) {
        this.id = id;
        this.cols = cols;
        this.rows = rows;
        this.onData = onData;
    }

    /**
     * Returns the PTY session identifier to request.
     *
     * @return requested PTY session identifier, or {@code null} to auto-generate
     */
    public String getId() {
        return id;
    }

    /**
     * Sets the PTY session identifier.
     *
     * @param id PTY session identifier
     * @return this options instance
     */
    public PtyCreateOptions setId(String id) {
        this.id = id;
        return this;
    }

    /**
     * Returns terminal width in columns.
     *
     * @return terminal width
     */
    public int getCols() {
        return cols;
    }

    /**
     * Sets terminal width in columns.
     *
     * @param cols terminal width
     * @return this options instance
     */
    public PtyCreateOptions setCols(int cols) {
        this.cols = cols;
        return this;
    }

    /**
     * Returns terminal height in rows.
     *
     * @return terminal height
     */
    public int getRows() {
        return rows;
    }

    /**
     * Sets terminal height in rows.
     *
     * @param rows terminal height
     * @return this options instance
     */
    public PtyCreateOptions setRows(int rows) {
        this.rows = rows;
        return this;
    }

    /**
     * Returns callback used for streaming PTY output.
     *
     * @return PTY output callback, or {@code null} when not configured
     */
    public Consumer<byte[]> getOnData() {
        return onData;
    }

    /**
     * Sets callback invoked for each PTY output chunk.
     *
     * @param onData callback receiving raw PTY bytes
     * @return this options instance
     */
    public PtyCreateOptions setOnData(Consumer<byte[]> onData) {
        this.onData = onData;
        return this;
    }
}
