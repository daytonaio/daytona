/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

/**
 * @deprecated The `runnerClass` concept has been superseded by `SandboxClass`.
 *
 * This enum is retained ONLY to preserve the `runnerClass` field on the
 * public `RunnerDto` for backward compatibility with existing API consumers
 * (proxies, older SDK versions). The DTO hardcodes `RunnerClass.CONTAINER`
 * regardless of the actual runner's workload class.
 *
 * DO NOT import this enum anywhere outside of `runner.dto.ts`. Use
 * `SandboxClass` from `sandbox-class.enum.ts` for any new runner /
 * sandbox class routing, gating, or scheduling decisions.
 */
export enum RunnerClass {
  CONTAINER = 'container',
}
