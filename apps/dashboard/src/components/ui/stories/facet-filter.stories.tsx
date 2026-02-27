/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import type { Meta, StoryObj } from '@storybook/react'
import { useState } from 'react'
import { FacetFilter } from '../facet-filter'

const meta: Meta<typeof FacetFilter> = {
  title: 'UI/FacetFilter',
  component: FacetFilter,
}

export default meta
type Story = StoryObj<typeof FacetFilter>

const options = [
  { label: 'Active', value: 'active' },
  { label: 'Inactive', value: 'inactive' },
  { label: 'Pending', value: 'pending' },
]

export const Default: Story = {
  render: () => {
    const [selected, setSelected] = useState<Set<string>>(new Set())
    return <FacetFilter title="Status" options={options} selectedValues={selected} setSelectedValues={setSelected} />
  },
}
