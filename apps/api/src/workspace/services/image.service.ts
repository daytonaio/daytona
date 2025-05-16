/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import {
  Injectable,
  NotFoundException,
  ConflictException,
  ForbiddenException,
  BadRequestException,
} from '@nestjs/common'
import { InjectRepository } from '@nestjs/typeorm'
import { Repository } from 'typeorm'
import { Image } from '../entities/image.entity'
import { ImageState } from '../enums/image-state.enum'
import { OrganizationService } from '../../organization/services/organization.service'
import { CreateImageDto } from '../../workspace/dto/create-image.dto'
import { BuildInfo } from '../entities/build-info.entity'
import { CreateBuildInfoDto } from '../dto/create-build-info.dto'
import { generateBuildInfoHash as generateBuildImageRef } from '../entities/build-info.entity'

@Injectable()
export class ImageService {
  constructor(
    @InjectRepository(Image)
    private readonly imageRepository: Repository<Image>,
    @InjectRepository(BuildInfo)
    private readonly buildInfoRepository: Repository<BuildInfo>,
    private readonly organizationService: OrganizationService,
  ) {}

  private validateImageName(name: string): string | null {
    if (!name.includes(':') || name.endsWith(':') || /:\s*$/.test(name)) {
      return 'Image name must include a tag (e.g., ubuntu:22.04)'
    }

    if (name.endsWith(':latest')) {
      return 'Images with tag ":latest" are not allowed'
    }

    // Basic format check
    const imageNameRegex =
      /^[a-z0-9]+(?:[._-][a-z0-9]+)*(?:\/[a-z0-9]+(?:[._-][a-z0-9]+)*)*:[a-z0-9]+(?:[._-][a-z0-9]+)*$/

    if (!imageNameRegex.test(name)) {
      return 'Invalid image name format. Must be lowercase, may contain digits, dots, dashes, and single slashes between components'
    }

    return null
  }

  async createImage(
    organizationId: string,
    createImageDto: CreateImageDto,
    buildInfo?: CreateBuildInfoDto,
    general = false,
  ) {
    const validationError = this.validateImageName(createImageDto.name)
    if (validationError) {
      throw new BadRequestException(validationError)
    }

    // get the org
    const organization = await this.organizationService.findOne(organizationId)
    if (!organization) {
      throw new NotFoundException(`Organization with ID ${organizationId} not found`)
    }

    // check if the organization has reached the image quota
    const images = await this.imageRepository.find({
      where: { organizationId },
    })

    if (images.length >= organization.imageQuota) {
      throw new ForbiddenException('Reached the maximum number of images in the organization')
    }

    try {
      const image = this.imageRepository.create({
        organizationId,
        ...createImageDto,
        state: buildInfo ? ImageState.BUILD_PENDING : ImageState.PENDING,
        general,
      })

      if (buildInfo) {
        const buildImageRef = generateBuildImageRef(buildInfo.dockerfileContent, buildInfo.contextHashes)

        // Check if buildInfo with the same imageRef already exists
        const existingBuildInfo = await this.buildInfoRepository.findOne({
          where: { imageRef: buildImageRef },
        })

        if (existingBuildInfo) {
          image.buildInfo = existingBuildInfo
        } else {
          const buildInfoEntity = this.buildInfoRepository.create({
            ...buildInfo,
          })
          await this.buildInfoRepository.save(buildInfoEntity)
          image.buildInfo = buildInfoEntity
        }
      }

      return await this.imageRepository.save(image)
    } catch (error) {
      if (error.code === '23505') {
        // PostgreSQL unique violation error code
        throw new ConflictException(`Image with name "${createImageDto.name}" already exists for this organization`)
      }
      throw error
    }
  }

  async toggleImageState(imageId: string, enabled: boolean) {
    const image = await this.imageRepository.findOne({
      where: { id: imageId },
    })

    if (!image) {
      throw new NotFoundException(`Image with ID ${imageId} not found`)
    }

    image.enabled = enabled
    return await this.imageRepository.save(image)
  }

  async removeImage(imageId: string) {
    const image = await this.imageRepository.findOne({
      where: { id: imageId },
    })

    if (!image) {
      throw new NotFoundException(`Image with ID ${imageId} not found`)
    }
    if (image.general) {
      throw new ForbiddenException('You cannot delete a general image')
    }
    image.state = ImageState.REMOVING
    await this.imageRepository.save(image)
  }

  async getAllImages(organizationId: string, page = 1, limit = 10) {
    const pageNum = Number(page)
    const limitNum = Number(limit)

    const [items, total] = await this.imageRepository.findAndCount({
      where: { organizationId },
      order: {
        createdAt: 'DESC',
      },
      skip: (pageNum - 1) * limitNum,
      take: limitNum,
    })

    return {
      items,
      total,
      page: pageNum,
      totalPages: Math.ceil(total / limit),
    }
  }

  async getImage(imageId: string): Promise<Image> {
    const image = await this.imageRepository.findOne({
      where: { id: imageId },
    })

    if (!image) {
      throw new NotFoundException(`Image with ID ${imageId} not found`)
    }

    return image
  }

  async getImageByName(imageName: string, organizationId: string): Promise<Image> {
    const image = await this.imageRepository.findOne({
      where: { name: imageName, organizationId },
    })

    if (!image) {
      //  check if the image is general
      const generalImage = await this.imageRepository.findOne({
        where: { name: imageName, general: true },
      })
      if (generalImage) {
        return generalImage
      }

      throw new NotFoundException(`Image with name ${imageName} not found`)
    }

    return image
  }

  async setImageGeneralStatus(imageId: string, general: boolean) {
    const image = await this.imageRepository.findOne({
      where: { id: imageId },
    })

    if (!image) {
      throw new NotFoundException(`Image with ID ${imageId} not found`)
    }

    image.general = general
    return await this.imageRepository.save(image)
  }
}
