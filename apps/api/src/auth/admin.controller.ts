/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Body, Controller, HttpCode, HttpStatus, Logger, Post, Req, UnauthorizedException } from '@nestjs/common'
import { ApiOperation, ApiProperty, ApiResponse, ApiTags } from '@nestjs/swagger'
import { IsString, MinLength } from 'class-validator'
import { JwtService } from '@nestjs/jwt'
import { Request } from 'express'
import { TypedConfigService } from '../config/typed-config.service'

export class AdminLoginDto {
    @ApiProperty({ description: 'Admin password', example: 'changeme' })
    @IsString()
    @MinLength(1)
    password: string
}

export class AdminLoginResponseDto {
    @ApiProperty({ description: 'JWT access token' })
    token: string
}

@ApiTags('admin')
@Controller('admin')
export class AdminController {
    private readonly logger = new Logger(AdminController.name)

    constructor(
        private readonly jwtService: JwtService,
        private readonly configService: TypedConfigService,
    ) { }

    @Post('login')
    @HttpCode(HttpStatus.OK)
    @ApiOperation({ summary: 'Admin login with password' })
    @ApiResponse({ status: 200, description: 'Login successful', type: AdminLoginResponseDto })
    @ApiResponse({ status: 401, description: 'Invalid password' })
    async login(@Body() dto: AdminLoginDto, @Req() req: Request): Promise<AdminLoginResponseDto> {
        const ip = (req.ips && req.ips.length ? req.ips[0] : req.ip) || 'unknown'
        const adminPassword = this.configService.get('adminPassword')
        if (!adminPassword) {
            this.logger.warn('ADMIN_PASSWORD is not configured')
            throw new UnauthorizedException('Admin authentication is not configured')
        }

        if (dto.password !== adminPassword) {
            this.logger.warn(`Failed admin login attempt from IP: ${ip}`)
            throw new UnauthorizedException('Invalid password')
        }

        this.logger.log(`Admin login successful from IP: ${ip}`)

        const token = await this.jwtService.signAsync({
            sub: 'admin',
            name: 'Admin',
        })

        return { token }
    }

    @Post('logout')
    @HttpCode(HttpStatus.OK)
    @ApiOperation({ summary: 'Admin logout (client-side token removal)' })
    @ApiResponse({ status: 200, description: 'Logout successful' })
    logout(): { message: string } {
        // Stateless JWT - client is responsible for removing the token
        return { message: 'Logged out successfully' }
    }
}
