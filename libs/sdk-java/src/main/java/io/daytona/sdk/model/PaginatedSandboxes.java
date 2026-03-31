// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk.model;

import com.fasterxml.jackson.annotation.JsonIgnoreProperties;
import com.fasterxml.jackson.annotation.JsonProperty;

import java.util.ArrayList;
import java.util.List;
import java.util.Map;

@JsonIgnoreProperties(ignoreUnknown = true)
public class PaginatedSandboxes {
    @JsonProperty("items")
    private List<Map<String, Object>> items;
    @JsonProperty("total")
    private Integer total;
    @JsonProperty("page")
    private Integer page;
    @JsonProperty("totalPages")
    private Integer totalPages;

    public List<Map<String, Object>> getItems() { return items == null ? new ArrayList<Map<String, Object>>() : items; }
    public void setItems(List<Map<String, Object>> items) { this.items = items; }
    public int getTotal() { return total == null ? 0 : total; }
    public void setTotal(Integer total) { this.total = total; }
    public int getPage() { return page == null ? 0 : page; }
    public void setPage(Integer page) { this.page = page; }
    public int getTotalPages() { return totalPages == null ? 0 : totalPages; }
    public void setTotalPages(Integer totalPages) { this.totalPages = totalPages; }
}