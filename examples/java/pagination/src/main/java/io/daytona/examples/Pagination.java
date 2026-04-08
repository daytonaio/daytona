// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.examples;

import io.daytona.sdk.Daytona;
import io.daytona.sdk.model.PaginatedSandboxes;
import io.daytona.sdk.model.PaginatedSnapshots;
import io.daytona.sdk.model.Snapshot;

import java.util.Map;

public class Pagination {
    public static void main(String[] args) {
        try (Daytona daytona = new Daytona()) {
            PaginatedSandboxes sandboxes = daytona.list(null, 1, 5);
            System.out.println("Found " + sandboxes.getTotal() + " sandboxes");
            for (Map<String, Object> sb : sandboxes.getItems()) {
                System.out.println("  " + sb.get("id") + ": " + sb.get("state"));
            }

            try {
                PaginatedSnapshots snapshots = daytona.snapshot().list(1, 5);
                System.out.println("Found " + snapshots.getTotal() + " snapshots");
                for (Snapshot snap : snapshots.getItems()) {
                    System.out.println("  " + snap.getName() + " (" + snap.getImageName() + ")");
                }
            } catch (Exception e) {
                System.out.println("Snapshot listing: " + e.getMessage());
            }
        }
    }
}
