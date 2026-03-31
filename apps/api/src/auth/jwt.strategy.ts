/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable, Logger } from '@nestjs/common'
import { PassportStrategy } from '@nestjs/passport'
import { ExtractJwt, Strategy } from 'passport-jwt'
import { passportJwtSecret } from 'jwks-rsa'
import { createRemoteJWKSet, JWTPayload, jwtVerify } from 'jose'
import { UserService } from '../user/user.service'
import { UserAuthContext } from '../common/interfaces/user-auth-context.interface'
import { AuthStrategyType } from './enums/auth-strategy-type.enum'
import { RequestWithAuthMetadata } from './interfaces/request-with-auth-metadata.interface'
import { CustomHeaders } from '../common/constants/header.constants'
import { TypedConfigService } from '../config/typed-config.service'

interface JwtStrategyConfig {
  jwksUri: string
  audience: string
  issuer: string
}

@Injectable()
export class JwtStrategy extends PassportStrategy(Strategy) {
  private readonly logger = new Logger(JwtStrategy.name)
  private JWKS: ReturnType<typeof createRemoteJWKSet>

  constructor(
    private readonly options: JwtStrategyConfig,
    private readonly userService: UserService,
    private readonly configService: TypedConfigService,
  ) {
    super({
      secretOrKeyProvider: passportJwtSecret({
        cache: true,
        rateLimit: true,
        jwksRequestsPerMinute: 5,
        jwksUri: options.jwksUri,
      }),
      jwtFromRequest: ExtractJwt.fromAuthHeaderAsBearerToken(),
      audience: options.audience,
      issuer: options.issuer,
      algorithms: ['RS256'],
      passReqToCallback: true,
    })
    this.JWKS = createRemoteJWKSet(new URL(options.jwksUri))
  }

  async validate(request: RequestWithAuthMetadata, payload: any): Promise<UserAuthContext | null> {
    if (!request.authMetadata?.isStrategyAllowed(AuthStrategyType.JWT)) {
      return null
    }

    const organizationId = request.get(CustomHeaders.ORGANIZATION_ID.name)

    let userId = payload.sub
    let email = payload.email

    /**
     * OKTA does not return the userId in access_token sub claim
     * real userId is in the uid claim and email is in the sub claim
     */
    if (payload.cid && payload.uid) {
      userId = payload.uid
      email = payload.sub
    }

    try {
      let existingUser = await this.userService.findOne(userId)

      if (!existingUser) {
        const newUser = await this.userService.create({
          id: userId,
          name: payload.name || payload.username || 'Unknown',
          email: email || '',
          emailVerified: payload.email_verified || false,
          personalOrganizationQuota: this.configService.getOrThrow('defaultOrganizationQuota'),
        })

        return {
          userId: newUser.id,
          role: newUser.role,
          email: newUser.email,
          organizationId,
        } satisfies UserAuthContext
      }

      if (!existingUser.emailVerified && payload.email_verified) {
        existingUser = await this.userService.update(existingUser.id, {
          emailVerified: payload.email_verified,
        })
      }

      if (existingUser.name === 'Unknown' || !existingUser.email) {
        existingUser = await this.userService.update(existingUser.id, {
          name: payload.name || payload.username || 'Unknown',
          email: email || '',
        })
      } else if (existingUser.email !== email) {
        existingUser = await this.userService.update(existingUser.id, {
          email: email || '',
        })
      }

      return {
        userId: existingUser.id,
        role: existingUser.role,
        email: existingUser.email,
        organizationId,
      } satisfies UserAuthContext
    } catch (error) {
      this.logger.error(`JWT validation failed for user ${userId}:`, error)
    }

    return null
  }

  async verifyToken(token: string): Promise<JWTPayload> {
    const { payload } = await jwtVerify(token, this.JWKS, {
      audience: this.options.audience,
      issuer: this.options.issuer,
      algorithms: ['RS256'],
    })
    return payload
  }
}
