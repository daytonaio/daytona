// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.examples;

import io.daytona.sdk.Daytona;
import io.daytona.sdk.Sandbox;
import io.daytona.sdk.model.ListSandboxesQuery;
import io.daytona.sdk.model.SandboxListSortDirection;
import io.daytona.sdk.model.SandboxListSortField;
import io.daytona.sdk.model.SandboxState;

import java.util.List;
import java.util.Map;

public class Pagination {
    public static void main(String[] args) {
        try (Daytona daytona = new Daytona()) {
            ListSandboxesQuery query = new ListSandboxesQuery();
            query.setLimit(10);
            query.setLabels(Map.of("env", "dev"));
            query.setStates(List.of(SandboxState.STARTED));
            query.setSort(SandboxListSortField.CREATED_AT);
            query.setOrder(SandboxListSortDirection.DESC);

            for (Sandbox sandbox : daytona.list(query)) {
                System.out.println(sandbox.getId());
            }
        }
    }
}
