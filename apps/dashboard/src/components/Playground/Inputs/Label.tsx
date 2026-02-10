/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Label } from '@/components/ui/label'
import { ParameterFormItem } from '@/contexts/PlaygroundContext'

type InputLabelProps = {
  formItem: ParameterFormItem
}

const InputLabel: React.FC<InputLabelProps> = ({ formItem }) => {
  return (
    <Label htmlFor={formItem.key} className="w-32 flex-shrink-0">
      <span className="relative">
        {formItem.label}
        {formItem.required ? <span className="text-muted-foreground">* </span> : null}
      </span>
    </Label>
  )
}

export default InputLabel
