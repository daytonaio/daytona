// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk.model;

import com.fasterxml.jackson.annotation.JsonIgnoreProperties;

@JsonIgnoreProperties(ignoreUnknown = true)
public class CreateSandboxFromImageParams extends CreateSandboxParams {
    private Object image;
    private Resources resources;

    public Object getImage() { return image; }
    public void setImage(Object image) { this.image = image; }
    public Resources getResources() { return resources; }
    public void setResources(Resources resources) { this.resources = resources; }
}