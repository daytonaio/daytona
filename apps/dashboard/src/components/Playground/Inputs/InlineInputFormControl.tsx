/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import React, { ReactNode } from 'react'
import { ParameterFormItem } from '@/contexts/PlaygroundContext'
import InputLabel from './Label'

type InlineInputFormControlProps = {
  formItem: ParameterFormItem
  children: ReactNode
}

const InlineInputFormControl: React.FC<InlineInputFormControlProps> = ({ formItem, children }) => {
  return (
    <div className="flex items-center gap-4">
      <InputLabel formItem={formItem} />
      {children}
    </div>
  )
}

export default InlineInputFormControl
