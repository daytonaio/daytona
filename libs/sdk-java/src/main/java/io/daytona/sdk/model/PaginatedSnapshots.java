// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk.model;

import com.fasterxml.jackson.annotation.JsonIgnoreProperties;
import com.fasterxml.jackson.annotation.JsonProperty;

import java.util.ArrayList;
import java.util.List;

@JsonIgnoreProperties(ignoreUnknown = true)
/**
 * Paginated list response for Snapshots.
 */
public class PaginatedSnapshots {
    @JsonProperty("items")
    private List<Snapshot> items;
    @JsonProperty("total")
    private Integer total;
    @JsonProperty("page")
    private Integer page;
    @JsonProperty("totalPages")
    private Integer totalPages;

    /**
     * Returns Snapshot items in the current page.
     *
     * @return page items
     */
    public List<Snapshot> getItems() { return items == null ? new ArrayList<Snapshot>() : items; }

    /**
     * Sets Snapshot items in the current page.
     *
     * @param items page items
     */
    public void setItems(List<Snapshot> items) { this.items = items; }

    /**
     * Returns total Snapshot count.
     *
     * @return total snapshots
     */
    public int getTotal() { return total == null ? 0 : total; }

    /**
     * Sets total Snapshot count.
     *
     * @param total total snapshots
     */
    public void setTotal(Integer total) { this.total = total; }

    /**
     * Returns current page number.
     *
     * @return current page
     */
    public int getPage() { return page == null ? 0 : page; }

    /**
     * Sets current page number.
     *
     * @param page current page
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
