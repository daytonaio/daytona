// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk.model;

import com.fasterxml.jackson.annotation.JsonIgnoreProperties;

import java.util.List;
import java.util.Map;

@JsonIgnoreProperties(ignoreUnknown = true)
/**
 * Base parameters used to create a Sandbox.
 *
 * <p>Includes common options such as user, language, environment variables, labels, lifecycle
 * intervals, and volume mounts.
 */
public class CreateSandboxParams {
    private String name;
    private String user;
    private String language;
    private Map<String, String> envVars;
    private Map<String, String> labels;
    private Boolean isPublic;
    private Integer autoStopInterval;
    private Integer autoArchiveInterval;
    private Integer autoDeleteInterval;
    private List<VolumeMount> volumes;
    private Boolean networkBlockAll;

    /**
     * Returns Sandbox name.
     *
     * @return custom Sandbox name, or {@code null} for server-generated name
     */
    public String getName() { return name; }

    /**
     * Sets Sandbox name.
     *
     * @param name desired Sandbox name
     */
    public void setName(String name) { this.name = name; }

    /**
     * Returns OS user used inside the Sandbox.
     *
     * @return user name, or {@code null} to use the image default
     */
    public String getUser() { return user; }

    /**
     * Sets OS user used inside the Sandbox.
     *
     * @param user Sandbox user name
     */
    public void setUser(String user) { this.user = user; }

    /**
     * Returns code execution language label.
     *
     * @return language identifier such as {@code python} or {@code typescript}
     */
    public String getLanguage() { return language; }

    /**
     * Sets code execution language label.
     *
     * @param language language identifier
     */
    public void setLanguage(String language) { this.language = language; }

    /**
     * Returns environment variables for the Sandbox.
     *
     * @return environment variable map
     */
    public Map<String, String> getEnvVars() { return envVars; }

    /**
     * Sets environment variables for the Sandbox.
     *
     * @param envVars environment variable map
     */
    public void setEnvVars(Map<String, String> envVars) { this.envVars = envVars; }

    /**
     * Returns Sandbox labels.
     *
     * @return key-value labels map
     */
    public Map<String, String> getLabels() { return labels; }

    /**
     * Sets Sandbox labels.
     *
     * @param labels key-value labels map
     */
    public void setLabels(Map<String, String> labels) { this.labels = labels; }

    /**
     * Returns whether Sandbox preview endpoints are public.
     *
     * @return {@code true} if previews are public
     */
    public Boolean getPublic() { return isPublic; }

    /**
     * Sets whether Sandbox preview endpoints are public.
     *
     * @param aPublic preview visibility flag
     */
    public void setPublic(Boolean aPublic) { isPublic = aPublic; }

    /**
     * Returns auto-stop interval in minutes.
     *
     * @return inactivity timeout before stopping
     */
    public Integer getAutoStopInterval() { return autoStopInterval; }

    /**
     * Sets auto-stop interval in minutes.
     *
     * @param autoStopInterval minutes of inactivity before stop
     */
    public void setAutoStopInterval(Integer autoStopInterval) { this.autoStopInterval = autoStopInterval; }

    /**
     * Returns auto-archive interval in minutes.
     *
     * @return time a stopped Sandbox remains before archive
     */
    public Integer getAutoArchiveInterval() { return autoArchiveInterval; }

    /**
     * Sets auto-archive interval in minutes.
     *
     * @param autoArchiveInterval archive delay in minutes
     */
    public void setAutoArchiveInterval(Integer autoArchiveInterval) { this.autoArchiveInterval = autoArchiveInterval; }

    /**
     * Returns auto-delete interval in minutes.
     *
     * @return deletion delay after stop
     */
    public Integer getAutoDeleteInterval() { return autoDeleteInterval; }

    /**
     * Sets auto-delete interval in minutes.
     *
     * @param autoDeleteInterval deletion delay in minutes
     */
    public void setAutoDeleteInterval(Integer autoDeleteInterval) { this.autoDeleteInterval = autoDeleteInterval; }

    /**
     * Returns volume mounts to attach.
     *
     * @return list of volume mounts
     */
    public List<VolumeMount> getVolumes() { return volumes; }

    /**
     * Sets volume mounts to attach.
     *
     * @param volumes volume mount definitions
     */
    public void setVolumes(List<VolumeMount> volumes) { this.volumes = volumes; }

    /**
     * Returns whether all outbound network access is blocked.
     *
     * @return {@code true} if network is fully blocked
     */
    public Boolean getNetworkBlockAll() { return networkBlockAll; }

    /**
     * Sets whether all outbound network access is blocked.
     *
     * @param networkBlockAll network block flag
     */
    public void setNetworkBlockAll(Boolean networkBlockAll) { this.networkBlockAll = networkBlockAll; }
}
