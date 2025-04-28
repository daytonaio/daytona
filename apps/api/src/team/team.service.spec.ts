/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Test, TestingModule } from '@nestjs/testing'
import { TeamService } from './team.service'
import { getRepositoryToken } from '@nestjs/typeorm'
import { Team } from './team.entity'
import { Repository } from 'typeorm'

const teamArray: Team[] = [
  {
    id: '1',
    name: 'team #1',
  },
  {
    id: '2',
    name: 'team #2',
  },
]

const oneTeam: Team = {
  id: '1',
  name: 'team #1',
}

describe('TeamService', () => {
  let service: TeamService
  let repository: Repository<Team>

  beforeEach(async () => {
    const module: TestingModule = await Test.createTestingModule({
      providers: [
        TeamService,
        {
          provide: getRepositoryToken(Team),
          useValue: {
            find: jest.fn().mockResolvedValue(teamArray),
            findOneBy: jest.fn().mockResolvedValue(oneTeam),
            save: jest.fn().mockResolvedValue(oneTeam),
            remove: jest.fn(),
            delete: jest.fn(),
          },
        },
      ],
    }).compile()

    service = module.get<TeamService>(TeamService)
    repository = module.get<Repository<Team>>(getRepositoryToken(Team))
  })

  it('should be defined', () => {
    expect(service).toBeDefined()
  })

  describe('create()', () => {
    it('should successfully insert a team', () => {
      const oneTeam: Team = {
        id: '1',
        name: 'team #1',
      }

      expect(
        service.create({
          name: oneTeam.name,
        }),
      ).resolves.toEqual(oneTeam)
    })
  })

  describe('findAll()', () => {
    it('should return an array of teams', async () => {
      const teams = await service.findAll()
      expect(teams).toEqual(teamArray)
    })
  })

  describe('findOne()', () => {
    it('should get a single team', () => {
      const repoSpy = jest.spyOn(repository, 'findOneBy')
      expect(service.findOne('6d225ef9-b6e1-4061-81c6-a9cf639a8897')).resolves.toEqual(oneTeam)
      expect(repoSpy).toHaveBeenCalledWith({ id: '6d225ef9-b6e1-4061-81c6-a9cf639a8897' })
    })
  })

  describe('remove()', () => {
    it('should call remove with the passed value', async () => {
      const removeSpy = jest.spyOn(repository, 'delete')
      const retVal = await service.remove('2')
      expect(removeSpy).toHaveBeenCalledWith('2')
      expect(retVal).toBeUndefined()
    })
  })
})
