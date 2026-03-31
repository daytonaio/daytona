// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk.model;

import com.fasterxml.jackson.annotation.JsonIgnoreProperties;
import com.fasterxml.jackson.annotation.JsonProperty;

import java.util.ArrayList;
import java.util.List;

@JsonIgnoreProperties(ignoreUnknown = true)
public class PaginatedSnapshots {
    @JsonProperty("items")
    private List<Snapshot> items;
    @JsonProperty("total")
    private Integer total;
    @JsonProperty("page")
    private Integer page;
    @JsonProperty("totalPages")
    private Integer totalPages;

    public List<Snapshot> getItems() { return items == null ? new ArrayList<Snapshot>() : items; }
    public void setItems(List<Snapshot> items) { this.items = items; }
    public int getTotal() { return total == null ? 0 : total; }
    public void setTotal(Integer total) { this.total = total; }
    public int getPage() { return page == null ? 0 : page; }
    public void setPage(Integer page) { this.page = page; }
    public int getTotalPages() { return totalPages == null ? 0 : totalPages; }
    public void setTotalPages(Integer totalPages) { this.totalPages = totalPages; }
}