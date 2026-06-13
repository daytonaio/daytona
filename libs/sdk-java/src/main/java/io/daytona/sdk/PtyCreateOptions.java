// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk;

import java.util.Map;
import java.util.function.Consumer;

/**
 * Options used when creating a PTY session in a Sandbox.
 */
public class PtyCreateOptions {
    private String id;
    private int cols = 120;
    private int rows = 30;
    private Consumer<byte[]> onData;
    private String cwd;
    private Map<String, String> envs;

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

    /**
     * Returns the working directory for the PTY session.
     *
     * @return working directory, or {@code null} to use the sandbox default
     */
    public String getCwd() {
        return cwd;
    }

    /**
     * Sets the working directory for the PTY session.
     *
     * @param cwd working directory
     * @return this options instance
     */
    public PtyCreateOptions setCwd(String cwd) {
        this.cwd = cwd;
        return this;
    }

    /**
     * Returns environment variables for the PTY session.
     *
     * @return environment variables, or {@code null} when none configured
     */
    public Map<String, String> getEnvs() {
        return envs;
    }

    /**
     * Sets environment variables for the PTY session.
     *
     * @param envs environment variables
     * @return this options instance
     */
    public PtyCreateOptions setEnvs(Map<String, String> envs) {
        this.envs = envs;
        return this;
    }
}
