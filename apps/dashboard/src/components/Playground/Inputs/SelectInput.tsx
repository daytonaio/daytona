/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { ParameterFormItem } from '@/contexts/PlaygroundContext'
import { Loader2 } from 'lucide-react'

type SelectOption = {
  value: string
  label: string
}

type FormSelectInputProps = {
  selectOptions: SelectOption[]
  selectValue: string | undefined
  formItem: ParameterFormItem
  onChangeHandler: (value: string) => void
  loading?: boolean
}

const FormSelectInput: React.FC<FormSelectInputProps> = ({
  selectOptions,
  selectValue,
  formItem,
  onChangeHandler,
  loading,
}) => {
  return (
    <Select value={selectValue} onValueChange={onChangeHandler}>
      <SelectTrigger className="w-full box-border" size="sm" aria-label={formItem.label}>
        {loading ? (
          <div className="w-full flex items-center justify-center gap-2">
            <Loader2 className="h-4 w-4 animate-spin text-muted-foreground" />
            <span className="text-muted-foreground">Loading...</span>
          </div>
        ) : (
          <SelectValue id={formItem.key} placeholder={formItem.placeholder} />
        )}
      </SelectTrigger>
      <SelectContent>
        {selectOptions.map((option) => (
          <SelectItem key={option.value} value={option.value}>
            {option.label}
          </SelectItem>
        ))}
      </SelectContent>
    </Select>
  )
}

export default FormSelectInput
