/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Test } from '@nestjs/testing'
import { UserController } from './user.controller'
import { UserService } from './user.service'
import { User } from './user.entity'
import { getRepositoryToken } from '@nestjs/typeorm'

describe('UserController', () => {
  let userController: UserController
  let userService: UserService

  beforeEach(async () => {
    const moduleRef = await Test.createTestingModule({
      controllers: [UserController],
      providers: [
        UserService,
        {
          provide: getRepositoryToken(User),
          useValue: {
            find: jest.fn().mockResolvedValue([]),
            findOne: jest.fn().mockResolvedValue(new User()),
            save: jest.fn().mockResolvedValue(new User()),
            delete: jest.fn().mockResolvedValue({ affected: 1 }),
          },
        },
      ],
    }).compile()

    userService = moduleRef.get<UserService>(UserService)
    userController = moduleRef.get<UserController>(UserController)
  })

  describe('findAll', () => {
    it('should return an array of users', async () => {
      const result: User[] = [
        {
          id: 'id1',
          authProviderId: 'providerId1',
          name: 'name1',
        },
      ]
      jest.spyOn(userService, 'findAll').mockImplementation(async () => result)

      expect(await userController.findAll()).toBe(result)
    })
  })
})
