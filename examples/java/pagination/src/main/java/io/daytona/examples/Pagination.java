// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.examples;

import io.daytona.sdk.Daytona;
import io.daytona.sdk.model.ListSandboxesQuery;

import java.util.Iterator;
import java.util.List;
import java.util.Map;

public class Pagination {
    public static void main(String[] args) {
        try (Daytona daytona = new Daytona()) {
            ListSandboxesQuery query = new ListSandboxesQuery();
            query.setLimit(10);
            query.setLabels(Map.of("env", "dev"));
            query.setStates(List.of("started"));
            query.setSort("createdAt");
            query.setOrder("desc");

            Iterator<Map<String, Object>> iter = daytona.list(query);
            while (iter.hasNext()) {
                Map<String, Object> sandbox = iter.next();
                System.out.println(sandbox.get("id"));
            }
        }
    }
}
