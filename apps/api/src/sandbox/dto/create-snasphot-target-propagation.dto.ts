import { ApiProperty, ApiSchema } from '@nestjs/swagger'
import { IsNumber, IsString } from 'class-validator'

@ApiSchema({ name: 'CreateSnapshotTargetPropagation' })
export class CreateSnapshotTargetPropagationDto {
  @ApiProperty({
    description: 'The target environment for the snapshot',
    example: 'local',
  })
  @IsString()
  target: string

  @ApiProperty({
    description: 'The desired concurrent sandboxes for the target',
    example: 1,
  })
  @IsNumber()
  desiredConcurrentSandboxes: number
}
