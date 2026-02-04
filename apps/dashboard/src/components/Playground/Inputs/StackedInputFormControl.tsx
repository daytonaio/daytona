/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import React, { ReactNode } from 'react'
import { ParameterFormItem } from '@/enums/Playground'
import InputLabel from './Label'

type StackedInputFormControlProps = {
  formItem: ParameterFormItem
  children: ReactNode
}

const StackedInputFormControl: React.FC<StackedInputFormControlProps> = ({ formItem, children }) => {
  return (
    <div className="space-y-2">
      <InputLabel formItem={formItem} />
      {children}
    </div>
  )
}

export default StackedInputFormControl
