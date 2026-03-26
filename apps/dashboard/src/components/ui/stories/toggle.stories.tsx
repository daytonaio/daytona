/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import type { Meta, StoryObj } from '@storybook/react'
import { Toggle } from '../toggle'
import { BoldIcon } from 'lucide-react'

const meta: Meta<typeof Toggle> = {
  title: 'UI/Toggle',
  component: Toggle,
  args: {
    children: <BoldIcon className="size-4" />,
    'aria-label': 'Toggle bold',
  },
}

export default meta
type Story = StoryObj<typeof Toggle>

export const Default: Story = {}
export const Outline: Story = { args: { variant: 'outline' } }
export const Small: Story = { args: { size: 'sm' } }
export const Large: Story = { args: { size: 'lg' } }
export const WithText: Story = { args: { children: 'Toggle me' } }

export const AllVariants: Story = {
  render: () => (
    <div className="flex flex-col gap-6">
      <div className="flex flex-col gap-2">
        <p className="text-sm font-medium text-muted-foreground">variant</p>
        <div className="flex items-center gap-2">
          <Toggle variant="default" aria-label="Default">
            <BoldIcon className="size-4" />
          </Toggle>
          <Toggle variant="outline" aria-label="Outline">
            <BoldIcon className="size-4" />
          </Toggle>
        </div>
      </div>
      <div className="flex flex-col gap-2">
        <p className="text-sm font-medium text-muted-foreground">size</p>
        <div className="flex items-center gap-2">
          <Toggle size="sm" aria-label="Small">
            <BoldIcon className="size-4" />
          </Toggle>
          <Toggle size="default" aria-label="Default">
            <BoldIcon className="size-4" />
          </Toggle>
          <Toggle size="lg" aria-label="Large">
            <BoldIcon className="size-4" />
          </Toggle>
        </div>
      </div>
    </div>
  ),
}
