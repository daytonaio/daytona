// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk.model;

import com.fasterxml.jackson.annotation.JsonIgnoreProperties;
import com.fasterxml.jackson.annotation.JsonProperty;

@JsonIgnoreProperties(ignoreUnknown = true)
/**
 * Snapshot metadata returned by Daytona APIs.
 */
public class Snapshot {
    @JsonProperty("id")
    private String id;
    @JsonProperty("name")
    private String name;
    @JsonProperty("imageName")
    private String imageName;
    @JsonProperty("state")
    private String state;

    /**
     * Returns snapshot identifier.
     *
     * @return snapshot ID
     */
    public String getId() { return id; }

    /**
     * Sets snapshot identifier.
     *
     * @param id snapshot ID
     */
    public void setId(String id) { this.id = id; }

    /**
     * Returns snapshot name.
     *
     * @return snapshot name
     */
    public String getName() { return name; }

    /**
     * Sets snapshot name.
     *
     * @param name snapshot name
     */
    public void setName(String name) { this.name = name; }

    /**
     * Returns snapshot image name.
     *
     * @return backing image name
     */
    public String getImageName() { return imageName; }

    /**
     * Sets snapshot image name.
     *
     * @param imageName backing image name
     */
    public void setImageName(String imageName) { this.imageName = imageName; }

    /**
     * Returns snapshot state.
     *
     * @return lifecycle state
     */
    public String getState() { return state; }

    /**
     * Sets snapshot state.
     *
     * @param state lifecycle state
     */
    public void setState(String state) { this.state = state; }
}
