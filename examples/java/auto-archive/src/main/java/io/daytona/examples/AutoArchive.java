// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.examples;

import io.daytona.sdk.Daytona;
import io.daytona.sdk.Sandbox;

public class AutoArchive {
    public static void main(String[] args) {
        try (Daytona daytona = new Daytona()) {
            Sandbox sandbox = daytona.create();
            try {
                System.out.println("autoArchiveInterval: " + sandbox.getAutoArchiveInterval());

                sandbox.setAutoArchiveInterval(60);
                System.out.println("autoArchiveInterval: " + sandbox.getAutoArchiveInterval());
            } finally {
                System.out.println("Deleting sandbox");
                sandbox.delete();
            }
        }
    }
}
