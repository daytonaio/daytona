/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import {
  Body,
  Controller,
  Delete,
  Get,
  Param,
  Patch,
  Post,
  Query,
  UseGuards,
  HttpCode,
  ForbiddenException,
} from '@nestjs/common'
import { ImageService } from '../services/image.service'
import {
  ApiOAuth2,
  ApiTags,
  ApiOperation,
  ApiResponse,
  ApiParam,
  ApiQuery,
  ApiHeader,
  ApiBearerAuth,
} from '@nestjs/swagger'
import { CreateImageDto } from '../dto/create-image.dto'
import { ToggleStateDto } from '../dto/toggle-state.dto'
import { ImageDto } from '../dto/image.dto'
import { PaginatedImagesDto } from '../dto/paginated-images.dto'
import { ImageAccessGuard } from '../guards/image-access.guard'
import { CustomHeaders } from '../../common/constants/header.constants'
import { AuthContext } from '../../common/decorators/auth-context.decorator'
import { OrganizationAuthContext } from '../../common/interfaces/auth-context.interface'
import { RequiredOrganizationResourcePermissions } from '../../organization/decorators/required-organization-resource-permissions.decorator'
import { OrganizationResourcePermission } from '../../organization/enums/organization-resource-permission.enum'
import { OrganizationResourceActionGuard } from '../../organization/guards/organization-resource-action.guard'
import { CombinedAuthGuard } from '../../auth/combined-auth.guard'
import { SystemActionGuard } from '../../auth/system-action.guard'
import { RequiredSystemRole } from '../../common/decorators/required-system-role.decorator'
import { SystemRole } from '../../user/enums/system-role.enum'
import { SetImageGeneralStatusDto } from '../dto/update-image.dto'
import { BuildImageDto } from '../dto/build-image.dto'

@ApiTags('images')
@Controller('images')
@ApiHeader(CustomHeaders.ORGANIZATION_ID)
@UseGuards(CombinedAuthGuard, SystemActionGuard, OrganizationResourceActionGuard)
@ApiOAuth2(['openid', 'profile', 'email'])
@ApiBearerAuth()
export class ImageController {
  constructor(private readonly imageService: ImageService) {}

  @Post()
  @HttpCode(200)
  @ApiOperation({
    summary: 'Create a new image',
    operationId: 'createImage',
  })
  @ApiResponse({
    status: 200,
    description: 'The image has been successfully created.',
    type: ImageDto,
  })
  @ApiResponse({
    status: 400,
    description: 'Bad request - Images with tag ":latest" are not allowed',
  })
  @RequiredOrganizationResourcePermissions([OrganizationResourcePermission.WRITE_IMAGES])
  async createImage(
    @AuthContext() authContext: OrganizationAuthContext,
    @Body() createImageDto: CreateImageDto,
  ): Promise<ImageDto> {
    if (createImageDto.general && authContext.role !== SystemRole.ADMIN) {
      throw new ForbiddenException('Insufficient permissions for creating general images')
    }

    // TODO: consider - if using transient registry, prepend the image name with the username
    const image = await this.imageService.createImage(authContext.organizationId, createImageDto)
    return ImageDto.fromImage(image)
  }

  @Get(':id')
  @ApiOperation({
    summary: 'Get image by ID',
    operationId: 'getImage',
  })
  @ApiParam({
    name: 'id',
    description: 'Image ID',
  })
  @ApiResponse({
    status: 200,
    description: 'The image',
    type: ImageDto,
  })
  @ApiResponse({
    status: 404,
    description: 'Image not found',
  })
  @UseGuards(ImageAccessGuard)
  async getImage(@Param('id') imageId: string): Promise<ImageDto> {
    const image = await this.imageService.getImage(imageId)
    return ImageDto.fromImage(image)
  }

  @Post('build')
  @HttpCode(200)
  @ApiOperation({
    summary: 'Build a Docker image',
    operationId: 'buildImage',
  })
  @ApiResponse({
    status: 200,
    description: 'The image has been successfully built.',
    type: ImageDto,
  })
  @ApiResponse({
    status: 400,
    description: 'Bad request - Invalid build parameters',
  })
  @RequiredOrganizationResourcePermissions([OrganizationResourcePermission.WRITE_IMAGES])
  async buildImage(
    @AuthContext() authContext: OrganizationAuthContext,
    @Body() buildImageDto: BuildImageDto,
  ): Promise<ImageDto> {
    const image = await this.imageService.createImage(
      authContext.organizationId,
      {
        name: buildImageDto.name,
      },
      {
        dockerfileContent: buildImageDto.buildInfo.dockerfileContent,
        contextHashes: buildImageDto.buildInfo.contextHashes,
      },
    )
    return ImageDto.fromImage(image)
  }

  @Patch(':id/toggle')
  @ApiOperation({
    summary: 'Toggle image state',
    operationId: 'toggleImageState',
  })
  @ApiParam({
    name: 'id',
    description: 'Image ID',
  })
  @ApiResponse({
    status: 200,
    description: 'Image state has been toggled',
    type: ImageDto,
  })
  @RequiredOrganizationResourcePermissions([OrganizationResourcePermission.WRITE_IMAGES])
  @UseGuards(ImageAccessGuard)
  async toggleImageState(@Param('id') imageId: string, @Body() toggleDto: ToggleStateDto): Promise<ImageDto> {
    const image = await this.imageService.toggleImageState(imageId, toggleDto.enabled)
    return ImageDto.fromImage(image)
  }

  @Delete(':id')
  @ApiOperation({
    summary: 'Delete image',
    operationId: 'removeImage',
  })
  @ApiParam({
    name: 'id',
    description: 'Image ID',
  })
  @ApiResponse({
    status: 200,
    description: 'Image has been deleted',
  })
  @RequiredOrganizationResourcePermissions([OrganizationResourcePermission.DELETE_IMAGES])
  @UseGuards(ImageAccessGuard)
  async removeImage(@Param('id') imageId: string): Promise<void> {
    await this.imageService.removeImage(imageId)
  }

  @Get()
  @ApiOperation({
    summary: 'List all images',
    operationId: 'getAllImages',
  })
  @ApiQuery({
    name: 'page',
    required: false,
    type: Number,
    description: 'Page number',
  })
  @ApiQuery({
    name: 'limit',
    required: false,
    type: Number,
    description: 'Number of items per page',
  })
  @ApiResponse({
    status: 200,
    description: 'List of all images with pagination',
    type: PaginatedImagesDto,
  })
  async getAllImages(
    @AuthContext() authContext: OrganizationAuthContext,
    @Query('page') page = 1,
    @Query('limit') limit = 10,
  ): Promise<PaginatedImagesDto> {
    const result = await this.imageService.getAllImages(authContext.organizationId, page, limit)
    return {
      items: result.items.map(ImageDto.fromImage),
      total: result.total,
      page: result.page,
      totalPages: result.totalPages,
    }
  }

  @Patch(':id/general')
  @ApiOperation({
    summary: 'Set image general status',
    operationId: 'setImageGeneralStatus',
  })
  @ApiParam({
    name: 'id',
    description: 'Image ID',
  })
  @ApiResponse({
    status: 200,
    description: 'Image general status has been set',
    type: ImageDto,
  })
  @RequiredSystemRole(SystemRole.ADMIN)
  async setImageGeneralStatus(@Param('id') imageId: string, @Body() dto: SetImageGeneralStatusDto): Promise<ImageDto> {
    const image = await this.imageService.setImageGeneralStatus(imageId, dto.general)
    return ImageDto.fromImage(image)
  }
}
