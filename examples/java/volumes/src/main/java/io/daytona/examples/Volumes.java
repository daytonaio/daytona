// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.examples;

import io.daytona.sdk.Daytona;
import io.daytona.sdk.model.Volume;

public class Volumes {
    public static void main(String[] args) {
        try (Daytona daytona = new Daytona()) {
            String volumeName = "test-vol-" + System.currentTimeMillis();
            Volume volume = daytona.volume().create(volumeName);
            try {
                System.out.println("id: " + volume.getId());
                System.out.println("name: " + volume.getName());
                System.out.println("state: " + volume.getState());

                Volume fetched = daytona.volume().getByName(volumeName);
                System.out.println("Fetched volume: " + fetched.getId());
            } finally {
                System.out.println("Deleting volume");
                try {
                    waitUntilDeletable(daytona, volumeName);
                    daytona.volume().delete(volume.getId());
                } catch (Exception e) {
                    System.out.println("Volume cleanup: " + e.getMessage());
                }
            }
        }
    }

    private static void waitUntilDeletable(Daytona daytona, String volumeName) throws InterruptedException {
        long start = System.currentTimeMillis();
        while (System.currentTimeMillis() - start < 60_000) {
            Volume v = daytona.volume().getByName(volumeName);
            if ("ready".equalsIgnoreCase(v.getState()) || "error".equalsIgnoreCase(v.getState())) {
                return;
            }
            Thread.sleep(1000);
        }
    }
}
