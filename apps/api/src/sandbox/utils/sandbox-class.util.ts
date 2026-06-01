/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */
import { SandboxClass } from '../enums/sandbox-class.enum'

/**
 * Temporary utility function to allow Android snapshots/sandboxes on Container runners
 */
export function getRunnerSandboxClass(sandboxClass: SandboxClass): SandboxClass {
  if (sandboxClass !== SandboxClass.ANDROID) {
    return sandboxClass
  }

  return SandboxClass.CONTAINER
}

/**
 * Returns true when snapshots of this class are stored as Docker/OCI references in a registry
 * (and therefore go through `parseDockerImage` / `findInternalRegistryBySnapshotRef` / runner Docker pulls).
 *
 * Returns false for classes whose `snapshot.ref` is NOT a registry reference — currently only
 * `WINDOWS`, where `snapshot.ref` is an S3 object key pointing at a VHD blob. Callers that
 * extract a registry from `snapshot.ref`, propagate via Docker pull, or otherwise treat the
 * ref as an OCI name MUST short-circuit for non-registry-based classes.
 */
export function isRegistryBasedSandboxClass(sandboxClass: SandboxClass): boolean {
  return sandboxClass !== SandboxClass.WINDOWS
}
