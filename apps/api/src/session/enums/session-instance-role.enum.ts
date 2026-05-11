/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

/**
 * Role of a SessionInstance within its (organization, template) fleet.
 *  - WARM: part of the always-on minimum-warm floor; never scaled in below `minWarm`.
 *  - OVERFLOW: added by the autoscaler under load; reaped first once idle.
 */
export enum SessionInstanceRole {
  WARM = 'warm',
  OVERFLOW = 'overflow',
}
