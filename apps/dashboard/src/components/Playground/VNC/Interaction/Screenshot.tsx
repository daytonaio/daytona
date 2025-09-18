/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Label } from '@/components/ui/label'
import InlineInputFormControl from '../../Inputs/InlineInputFormControl'
import FormSelectInput from '../../Inputs/SelectInput'
import FormNumberInput from '../../Inputs/NumberInput'
import FormCheckboxInput from '../../Inputs/CheckboxInput'
import { usePlayground } from '@/hooks/usePlayground'
import { PlaygroundActionInvokeApi } from '@/contexts/PlaygroundContext'
import { ScreenshotRegion } from '@daytonaio/sdk'
import {
  CustomizedScreenshotOptions,
  ScreenshotActions,
  ScreenshotActionFormData,
  ParameterFormData,
} from '@/enums/Playground'
import { NumberParameterFormItem, ParameterFormItem, ScreenshotFormatOption } from '@/enums/Playground'
import PlaygroundActionForm from '../../ActionForm'
import { useState } from 'react'

const VNCScreenshootOperations: React.FC = () => {
  return <div />
}

export default VNCScreenshootOperations
