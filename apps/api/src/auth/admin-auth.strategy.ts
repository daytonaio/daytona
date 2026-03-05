/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable, Logger } from '@nestjs/common'
import { PassportStrategy } from '@nestjs/passport'
import { ExtractJwt, Strategy } from 'passport-jwt'
import { UserService } from '../user/user.service'
import { AuthContext } from '../common/interfaces/auth-context.interface'
import { Request } from 'express'
import { CustomHeaders } from '../common/constants/header.constants'
import { TypedConfigService } from '../config/typed-config.service'

@Injectable()
export class AdminAuthStrategy extends PassportStrategy(Strategy, 'admin-jwt') {
  private readonly logger = new Logger(AdminAuthStrategy.name)

  constructor(
    private readonly userService: UserService,
    private readonly configService: TypedConfigService,
  ) {
    super({
      jwtFromRequest: ExtractJwt.fromAuthHeaderAsBearerToken(),
      secretOrKey: configService.getOrThrow('jwtSecret'),
      algorithms: ['HS256'],
      passReqToCallback: true,
    })
    this.logger.debug('AdminAuthStrategy initialized')
  }

  async validate(request: Request, payload: any): Promise<AuthContext> {
    this.logger.debug('AdminAuthStrategy.validate called')

    const userId = payload.sub
    let user = await this.userService.findOne(userId)

    if (!user) {
      user = await this.userService.create({
        id: userId,
        name: 'Admin',
        email: '',
        emailVerified: true,
        personalOrganizationQuota: this.configService.getOrThrow('defaultOrganizationQuota'),
      })
      this.logger.debug(`Created admin user with ID: ${userId}`)
    }

    const organizationId = request.get(CustomHeaders.ORGANIZATION_ID.name)

    return {
      userId: user.id,
      role: user.role,
      email: user.email,
      organizationId,
    }
  }
}
