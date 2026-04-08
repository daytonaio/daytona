// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.examples;

import io.daytona.sdk.Daytona;
import io.daytona.sdk.Sandbox;
import io.daytona.sdk.model.PaginatedSandboxes;

import java.util.HashMap;
import java.util.Map;

public class Lifecycle {
    public static void main(String[] args) {
        try (Daytona daytona = new Daytona()) {
            System.out.println("Creating sandbox");
            Sandbox sandbox = daytona.create();
            System.out.println("Sandbox created: " + sandbox.getId() + " (state: " + sandbox.getState() + ")");

            try {
                Map<String, String> labels = new HashMap<>();
                labels.put("test", "lifecycle");
                sandbox.setLabels(labels);
                System.out.println("Labels set: test=lifecycle");

                System.out.println("Stopping sandbox");
                sandbox.stop();
                System.out.println("Sandbox stopped");

                System.out.println("Starting sandbox");
                sandbox.start();
                System.out.println("Sandbox started");

                System.out.println("Getting existing sandbox");
                Sandbox fetched = daytona.get(sandbox.getId());
                System.out.println("Got sandbox: " + fetched.getId() + " (state: " + fetched.getState() + ")");

                PaginatedSandboxes sandboxes = daytona.list();
                System.out.println("Total sandboxes: " + sandboxes.getTotal());
            } finally {
                System.out.println("Deleting sandbox");
                sandbox.delete();
                System.out.println("Sandbox deleted");
            }
        }
    }
}
