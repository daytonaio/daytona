/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Inject, Injectable, Logger } from '@nestjs/common'
import { renderFile } from 'ejs'
import { createTransport, Transporter } from 'nodemailer'
import path from 'path'
import { OnAsyncEvent } from '../../common/decorators/on-async-event.decorator'
import { OrganizationEvents } from '../../organization/constants/organization-events.constant'
import { OrganizationInvitationCreatedEvent } from '../../organization/events/organization-invitation-created.event'
import { EmailModuleOptions } from '../email.module'

@Injectable()
export class EmailService {
  private readonly transporter: Transporter | null
  private readonly logger = new Logger(EmailService.name)

  constructor(@Inject('EMAIL_MODULE_OPTIONS') private readonly options: EmailModuleOptions) {
    const { host, port, user, password, secure, from, dashboardUrl } = this.options

    if (!host || !port || !from) {
      this.logger.warn('Email configuration not found, email functionality will be disabled')
      this.transporter = null
      return
    }

    this.transporter = createTransport({
      host,
      port,
      auth: user && password ? { user, pass: password } : undefined,
      secure,
    })
  }

  @OnAsyncEvent({
    event: OrganizationEvents.INVITATION_CREATED,
  })
  async handleOrganizationInvitationCreated(payload: OrganizationInvitationCreatedEvent): Promise<void> {
    if (!this.transporter) {
      this.logger.warn('Failed to send organization invitation email, email configuration not found')
      return
    }

    try {
      await this.transporter.sendMail({
        from: this.options.from,
        to: payload.inviteeEmail,
        subject: 'Invitation to join a Daytona organization',
        html: await renderFile(path.join(__dirname, 'assets/templates/organization-invitation.template.ejs'), {
          organizationName: payload.organizationName,
          invitedBy: payload.invitedBy,
          invitationLink: `${this.options.dashboardUrl}/user/invitations?id=${payload.invitationId}`,
          expiresAt: new Date(payload.expiresAt).toLocaleDateString('en-US', {
            year: 'numeric',
            month: 'long',
            day: 'numeric',
          }),
        }),
      })
    } catch (error) {
      // TODO: resilient email sending
      this.logger.error(`Failed to send organization invitation email to ${payload.inviteeEmail}`, error)
    }
  }
}
