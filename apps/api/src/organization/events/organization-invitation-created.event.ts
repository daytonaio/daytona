/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

export class OrganizationInvitationCreatedEvent {
  constructor(
    public readonly organizationName: string,
    public readonly invitedBy: string,
    public readonly inviteeEmail: string,
    public readonly invitationId: string,
    public readonly expiresAt: Date,
  ) {}
}
