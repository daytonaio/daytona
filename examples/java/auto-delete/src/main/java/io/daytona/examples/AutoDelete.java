// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.examples;

import io.daytona.sdk.Daytona;
import io.daytona.sdk.Sandbox;

public class AutoDelete {
    public static void main(String[] args) {
        try (Daytona daytona = new Daytona()) {
            Sandbox sandbox = daytona.create();
            try {
                System.out.println("autoDeleteInterval: " + sandbox.getAutoDeleteInterval());

                sandbox.setAutoDeleteInterval(60);
                System.out.println("autoDeleteInterval: " + sandbox.getAutoDeleteInterval());

                sandbox.setAutoDeleteInterval(0);
                System.out.println("autoDeleteInterval: " + sandbox.getAutoDeleteInterval());

                sandbox.setAutoDeleteInterval(-1);
                System.out.println("autoDeleteInterval: " + sandbox.getAutoDeleteInterval());
            } finally {
                System.out.println("Deleting sandbox");
                sandbox.delete();
            }
        }
    }
}
