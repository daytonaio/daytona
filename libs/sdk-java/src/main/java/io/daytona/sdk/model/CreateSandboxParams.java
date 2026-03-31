// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk.model;

import com.fasterxml.jackson.annotation.JsonIgnoreProperties;

import java.util.List;
import java.util.Map;

@JsonIgnoreProperties(ignoreUnknown = true)
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

    public String getName() { return name; }
    public void setName(String name) { this.name = name; }
    public String getUser() { return user; }
    public void setUser(String user) { this.user = user; }
    public String getLanguage() { return language; }
    public void setLanguage(String language) { this.language = language; }
    public Map<String, String> getEnvVars() { return envVars; }
    public void setEnvVars(Map<String, String> envVars) { this.envVars = envVars; }
    public Map<String, String> getLabels() { return labels; }
    public void setLabels(Map<String, String> labels) { this.labels = labels; }
    public Boolean getPublic() { return isPublic; }
    public void setPublic(Boolean aPublic) { isPublic = aPublic; }
    public Integer getAutoStopInterval() { return autoStopInterval; }
    public void setAutoStopInterval(Integer autoStopInterval) { this.autoStopInterval = autoStopInterval; }
    public Integer getAutoArchiveInterval() { return autoArchiveInterval; }
    public void setAutoArchiveInterval(Integer autoArchiveInterval) { this.autoArchiveInterval = autoArchiveInterval; }
    public Integer getAutoDeleteInterval() { return autoDeleteInterval; }
    public void setAutoDeleteInterval(Integer autoDeleteInterval) { this.autoDeleteInterval = autoDeleteInterval; }
    public List<VolumeMount> getVolumes() { return volumes; }
    public void setVolumes(List<VolumeMount> volumes) { this.volumes = volumes; }
    public Boolean getNetworkBlockAll() { return networkBlockAll; }
    public void setNetworkBlockAll(Boolean networkBlockAll) { this.networkBlockAll = networkBlockAll; }
}