/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable } from '@nestjs/common'
import { InjectRepository } from '@nestjs/typeorm'
import { Team } from './team.entity'
import { Repository } from 'typeorm'
import { CreateTeamDto } from './dto/create-team.dto'

@Injectable()
export class TeamService {
  constructor(
    @InjectRepository(Team)
    private readonly teamRepository: Repository<Team>,
  ) {}

  create(createUserDto: CreateTeamDto): Promise<Team> {
    const team = new Team()
    team.name = createUserDto.name

    return this.teamRepository.save(team)
  }

  async findAll(): Promise<Team[]> {
    return this.teamRepository.find()
  }

  findOne(id: string): Promise<Team> {
    return this.teamRepository.findOneBy({ id: id })
  }

  async remove(id: string): Promise<void> {
    await this.teamRepository.delete(id)
  }
}
