/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Module } from '@nestjs/common'
import { PassportModule } from '@nestjs/passport'
import { JwtStrategy } from './jwt.strategy'
import { ApiKeyStrategy } from './api-key.strategy'
import { UserModule } from '../user/user.module'
import { ApiKeyModule } from '../api-key/api-key.module'
import { SandboxModule } from '../sandbox/sandbox.module'
import { TypedConfigService } from '../config/typed-config.service'
import { HttpModule, HttpService } from '@nestjs/axios'
import { OidcMetadata } from 'oidc-client-ts'
import { firstValueFrom } from 'rxjs'
import { UserService } from '../user/user.service'
import { TypedConfigModule } from '../config/typed-config.module'
import { catchError, map } from 'rxjs/operators'
@Module({
  imports: [
    PassportModule.register({
      defaultStrategy: ['jwt', 'api-key'],
      property: 'user',
      session: false,
    }),
    TypedConfigModule,
    UserModule,
    ApiKeyModule,
    SandboxModule,
    HttpModule,
  ],
  providers: [
    ApiKeyStrategy,
    {
      provide: JwtStrategy,
      useFactory: async (userService: UserService, httpService: HttpService, configService: TypedConfigService) => {
        if (configService.get('skipConnections')) {
          return
        }

        // Get the OpenID configuration from the issuer
        const discoveryUrl = `${configService.get('oidc.issuer')}/.well-known/openid-configuration`
        const metadata = await firstValueFrom(
          httpService.get(discoveryUrl).pipe(
            map((response) => response.data as OidcMetadata),
            catchError((error) => {
              throw new Error(`Failed to fetch OpenID configuration: ${error.message}`)
            }),
          ),
        )

        let jwksUri = metadata.jwks_uri

        const internalIssuer = configService.getOrThrow('oidc.issuer')
        const publicIssuer = configService.get('oidc.publicIssuer')
        if (publicIssuer) {
          // Replace localhost URLs with Docker network URLs for internal API use
          jwksUri = metadata.jwks_uri.replace(publicIssuer, internalIssuer)
        }
        return new JwtStrategy(
          {
            audience: configService.get('oidc.audience'),
            issuer: metadata.issuer,
            jwksUri: jwksUri,
          },
          userService,
          configService,
        )
      },
      inject: [UserService, HttpService, TypedConfigService],
    },
  ],
  exports: [PassportModule, JwtStrategy, ApiKeyStrategy],
})
export class AuthModule {}
