// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk.model;

import com.fasterxml.jackson.annotation.JsonIgnoreProperties;

@JsonIgnoreProperties(ignoreUnknown = true)
/**
 * Parameters for creating a Sandbox from an image definition.
 */
public class CreateSandboxFromImageParams extends CreateSandboxParams {
    private Object image;
    private Resources resources;

    /**
     * Returns image source used for Sandbox creation.
     *
     * @return image reference as {@link String} or {@link io.daytona.sdk.Image}
     */
    public Object getImage() { return image; }

    /**
     * Sets image source used for Sandbox creation.
     *
     * @param image image reference as string or {@link io.daytona.sdk.Image}
     */
    public void setImage(Object image) { this.image = image; }

    /**
     * Returns resource overrides for the Sandbox.
     *
     * @return resource configuration
     */
    public Resources getResources() { return resources; }

    /**
     * Sets resource overrides for the Sandbox.
     *
     * @param resources CPU, memory, disk, and GPU configuration
     */
    public void setResources(Resources resources) { this.resources = resources; }
}
