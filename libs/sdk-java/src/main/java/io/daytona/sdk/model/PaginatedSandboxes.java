// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk.model;

import com.fasterxml.jackson.annotation.JsonIgnoreProperties;
import com.fasterxml.jackson.annotation.JsonProperty;

import java.util.ArrayList;
import java.util.List;
import java.util.Map;

@JsonIgnoreProperties(ignoreUnknown = true)
/**
 * Paginated list response for Sandboxes.
 */
public class PaginatedSandboxes {
    @JsonProperty("items")
    private List<Map<String, Object>> items;
    @JsonProperty("total")
    private Integer total;
    @JsonProperty("page")
    private Integer page;
    @JsonProperty("totalPages")
    private Integer totalPages;

    /**
     * Returns Sandbox items in the current page.
     *
     * @return page items
     */
    public List<Map<String, Object>> getItems() { return items == null ? new ArrayList<Map<String, Object>>() : items; }

    /**
     * Sets Sandbox items in the current page.
     *
     * @param items page items
     */
    public void setItems(List<Map<String, Object>> items) { this.items = items; }

    /**
     * Returns total Sandbox count.
     *
     * @return total number of Sandboxes
     */
    public int getTotal() { return total == null ? 0 : total; }

    /**
     * Sets total Sandbox count.
     *
     * @param total total number of Sandboxes
     */
    public void setTotal(Integer total) { this.total = total; }

    /**
     * Returns current page number.
     *
     * @return current page index
     */
    public int getPage() { return page == null ? 0 : page; }

    /**
     * Sets current page number.
     *
     * @param page current page index
     */
    public void setPage(Integer page) { this.page = page; }

    /**
     * Returns total page count.
     *
     * @return total pages
     */
    public int getTotalPages() { return totalPages == null ? 0 : totalPages; }

    /**
     * Sets total page count.
     *
     * @param totalPages total pages
     */
    public void setTotalPages(Integer totalPages) { this.totalPages = totalPages; }
}
