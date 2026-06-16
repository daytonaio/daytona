/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import type { Meta, StoryObj } from '@storybook/react'
import { useState } from 'react'
import { FacetedFilter } from '../faceted-filter'

const meta: Meta<typeof FacetedFilter> = {
  title: 'UI/FacetedFilter',
  component: FacetedFilter,
}

export default meta
type Story = StoryObj<typeof FacetedFilter>

const options = [
  { label: 'Active', value: 'active' },
  { label: 'Inactive', value: 'inactive' },
  { label: 'Pending', value: 'pending' },
]

export const Default: Story = {
  render: () => {
    const [selected, setSelected] = useState<Set<string>>(new Set())
    return <FacetedFilter title="Status" options={options} values={selected} onValuesChange={setSelected} />
  },
}
