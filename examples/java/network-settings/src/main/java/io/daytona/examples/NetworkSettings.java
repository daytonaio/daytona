// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.examples;

import io.daytona.sdk.Daytona;
import io.daytona.sdk.Sandbox;

public class NetworkSettings {
    public static void main(String[] args) {
        try (Daytona daytona = new Daytona()) {
            System.out.println("Creating sandbox");
            Sandbox sandbox = daytona.create();
            System.out.println("Sandbox created: " + sandbox.getId());

            try {
                System.out.println("id: " + sandbox.getId());
                System.out.println("state: " + sandbox.getState());
            } finally {
                System.out.println("Deleting sandbox");
                sandbox.delete();
            }
        }
    }
}
