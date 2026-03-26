/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty, ApiPropertyOptional, ApiSchema } from '@nestjs/swagger'
import { IsString, IsNumber, IsOptional, ValidateNested, IsArray, IsBoolean } from 'class-validator'
import { Type } from 'class-transformer'

@ApiSchema({ name: 'LspServerRequest' })
export class LspServerRequestDto {
  @ApiProperty({ description: 'Language identifier' })
  @IsString()
  languageId: string

  @ApiProperty({ description: 'Path to the project' })
  @IsString()
  pathToProject: string
}

@ApiSchema({ name: 'LspDocumentRequest' })
export class LspDocumentRequestDto extends LspServerRequestDto {
  @ApiProperty({ description: 'Document URI' })
  @IsString()
  uri: string
}

@ApiSchema({ name: 'Position' })
export class PositionDto {
  @ApiProperty()
  @IsNumber()
  line: number

  @ApiProperty()
  @IsNumber()
  character: number
}

@ApiSchema({ name: 'CompletionContext' })
export class CompletionContextDto {
  @ApiProperty()
  @IsNumber()
  triggerKind: number

  @ApiPropertyOptional()
  @IsOptional()
  @IsString()
  triggerCharacter?: string
}

@ApiSchema({ name: 'LspCompletionParams' })
export class LspCompletionParamsDto extends LspDocumentRequestDto {
  @ApiProperty()
  @ValidateNested()
  @Type(() => PositionDto)
  position: PositionDto

  @ApiPropertyOptional()
  @IsOptional()
  @ValidateNested()
  @Type(() => CompletionContextDto)
  context?: CompletionContextDto
}

@ApiSchema({ name: 'Range' })
export class RangeDto {
  @ApiProperty()
  @ValidateNested()
  @Type(() => PositionDto)
  start: PositionDto

  @ApiProperty()
  @ValidateNested()
  @Type(() => PositionDto)
  end: PositionDto
}

@ApiSchema({ name: 'CompletionItem' })
export class CompletionItemDto {
  @ApiProperty()
  @IsString()
  label: string

  @ApiPropertyOptional()
  @IsOptional()
  @IsNumber()
  kind?: number

  @ApiPropertyOptional()
  @IsOptional()
  @IsString()
  detail?: string

  @ApiPropertyOptional()
  @IsOptional()
  documentation?: any

  @ApiPropertyOptional()
  @IsOptional()
  @IsString()
  sortText?: string

  @ApiPropertyOptional()
  @IsOptional()
  @IsString()
  filterText?: string

  @ApiPropertyOptional()
  @IsOptional()
  @IsString()
  insertText?: string
}

@ApiSchema({ name: 'CompletionList' })
export class CompletionListDto {
  @ApiProperty()
  @IsBoolean()
  isIncomplete: boolean

  @ApiProperty({ type: [CompletionItemDto] })
  @IsArray()
  @ValidateNested({ each: true })
  @Type(() => CompletionItemDto)
  items: CompletionItemDto[]
}

@ApiSchema({ name: 'LspLocation' })
export class LspLocationDto {
  @ApiProperty()
  @ValidateNested()
  @Type(() => RangeDto)
  range: RangeDto

  @ApiProperty()
  @IsString()
  uri: string
}

@ApiSchema({ name: 'LspSymbol' })
export class LspSymbolDto {
  @ApiProperty()
  @IsNumber()
  kind: number

  @ApiProperty()
  @ValidateNested()
  @Type(() => LspLocationDto)
  location: LspLocationDto

  @ApiProperty()
  @IsString()
  name: string
}

@ApiSchema({ name: 'WorkspaceSymbolParams' })
export class WorkspaceSymbolParamsDto {
  @ApiProperty()
  @IsString()
  query: string
}
