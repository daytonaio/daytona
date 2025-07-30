/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable, NotFoundException } from '@nestjs/common'
import { InjectRepository } from '@nestjs/typeorm'
import { User, UserSSHKeyPair } from './user.entity'
import { DataSource, ILike, In, Repository } from 'typeorm'
import { CreateUserDto } from './dto/create-user.dto'
import * as crypto from 'crypto'
import * as forge from 'node-forge'
import { EventEmitter2 } from '@nestjs/event-emitter'
import { UserEvents } from './constants/user-events.constant'
import { UpdateUserDto } from './dto/update-user.dto'
import { UserCreatedEvent } from './events/user-created.event'
import { UserDeletedEvent } from './events/user-deleted.event'
import { UserEmailVerifiedEvent } from './events/user-email-verified.event'

@Injectable()
export class UserService {
  constructor(
    @InjectRepository(User)
    private readonly userRepository: Repository<User>,
    private readonly eventEmitter: EventEmitter2,
    private readonly dataSource: DataSource,
  ) {}

  async create(createUserDto: CreateUserDto): Promise<User> {
    let user = new User()
    user.id = createUserDto.id
    user.name = createUserDto.name
    const keyPair = await this.generatePrivateKey()
    user.keyPair = keyPair
    user.publicKeys = []
    user.emailVerified = createUserDto.emailVerified

    if (createUserDto.email) {
      user.email = createUserDto.email
    }

    if (createUserDto.role) {
      user.role = createUserDto.role
    }

    await this.dataSource.transaction(async (em) => {
      user = await em.save(user)
      await this.eventEmitter.emitAsync(
        UserEvents.CREATED,
        new UserCreatedEvent(em, user, createUserDto.personalOrganizationQuota),
      )
    })

    return user
  }

  async findAll(): Promise<User[]> {
    return this.userRepository.find()
  }

  async findByIds(ids: string[]): Promise<User[]> {
    if (ids.length === 0) {
      return []
    }

    return this.userRepository.find({
      where: {
        id: In(ids),
      },
    })
  }

  async findOne(id: string): Promise<User | null> {
    return this.userRepository.findOne({ where: { id } })
  }

  async findOneOrFail(id: string): Promise<User> {
    return this.userRepository.findOneOrFail({ where: { id } })
  }

  async findOneByEmail(email: string, ignoreCase = false): Promise<User | null> {
    return this.userRepository.findOne({
      where: {
        email: ignoreCase ? ILike(email) : email,
      },
    })
  }

  async remove(id: string): Promise<void> {
    await this.dataSource.transaction(async (em) => {
      await em.delete(User, id)
      await this.eventEmitter.emitAsync(UserEvents.DELETED, new UserDeletedEvent(em, id))
    })
  }

  async regenerateKeyPair(id: string): Promise<User> {
    const user = await this.userRepository.findOneBy({ id: id })
    const keyPair = await this.generatePrivateKey()
    user.keyPair = keyPair
    return this.userRepository.save(user)
  }

  private generatePrivateKey(): Promise<UserSSHKeyPair> {
    const comment = 'daytona'

    return new Promise((resolve, reject) => {
      crypto.generateKeyPair(
        'rsa',
        {
          modulusLength: 4096,
          publicKeyEncoding: {
            type: 'pkcs1',
            format: 'pem',
          },
          privateKeyEncoding: {
            type: 'pkcs1',
            format: 'pem',
          },
        },
        (error, publicKey, privateKey) => {
          if (error) {
            reject(error)
          } else {
            const publicKeySShEncoded = forge.ssh.publicKeyToOpenSSH(forge.pki.publicKeyFromPem(publicKey), comment)

            const privateKeySShEncoded = forge.ssh.privateKeyToOpenSSH(forge.pki.privateKeyFromPem(privateKey))

            resolve({
              publicKey: publicKeySShEncoded,
              privateKey: privateKeySShEncoded,
            })
          }
        },
      )
    })
  }

  // TODO: discuss if we need separate methods for updating specific fields
  async update(userId: string, updateUserDto: UpdateUserDto): Promise<User> {
    const user = await this.userRepository.findOne({
      where: {
        id: userId,
      },
    })

    if (!user) {
      throw new NotFoundException(`User with ID ${userId} not found.`)
    }

    if (updateUserDto.name) {
      user.name = updateUserDto.name
    }

    if (updateUserDto.email) {
      user.email = updateUserDto.email
    }

    if (updateUserDto.role) {
      user.role = updateUserDto.role
    }

    if (updateUserDto.emailVerified) {
      user.emailVerified = updateUserDto.emailVerified
      await this.dataSource.transaction(async (em) => {
        await em.save(user)
        await this.eventEmitter.emitAsync(UserEvents.EMAIL_VERIFIED, new UserEmailVerifiedEvent(em, user.id))
      })
    }

    return this.userRepository.save(user)
  }
}
