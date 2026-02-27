/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import type { Meta, StoryObj } from '@storybook/react'
import { Switch } from '../switch'
import { Label } from '../label'

const meta: Meta<typeof Switch> = {
  title: 'UI/Switch',
  component: Switch,
}

export default meta
type Story = StoryObj<typeof Switch>

export const Default: Story = {}
export const Checked: Story = { args: { defaultChecked: true } }
export const Small: Story = { args: { size: 'sm' } }
export const Disabled: Story = { args: { disabled: true } }

export const WithLabel: Story = {
  render: () => (
    <div className="flex items-center gap-2">
      <Switch id="airplane" />
      <Label htmlFor="airplane">Airplane Mode</Label>
    </div>
  ),
}

export const AllVariants: Story = {
  render: () => (
    <div className="flex flex-col gap-6">
      <div className="flex flex-col gap-2">
        <p className="text-sm font-medium text-muted-foreground">size · default</p>
        <div className="flex items-center gap-4">
          <div className="flex items-center gap-2">
            <Switch />
            <Label>Default</Label>
          </div>
          <div className="flex items-center gap-2">
            <Switch defaultChecked />
            <Label>Checked</Label>
          </div>
        </div>
      </div>
      <div className="flex flex-col gap-2">
        <p className="text-sm font-medium text-muted-foreground">size · sm</p>
        <div className="flex items-center gap-4">
          <div className="flex items-center gap-2">
            <Switch size="sm" />
            <Label>Small</Label>
          </div>
          <div className="flex items-center gap-2">
            <Switch size="sm" defaultChecked />
            <Label>Small Checked</Label>
          </div>
        </div>
      </div>
      <div className="flex flex-col gap-2">
        <p className="text-sm font-medium text-muted-foreground">disabled</p>
        <div className="flex items-center gap-2">
          <Switch disabled />
          <Label>Disabled</Label>
        </div>
      </div>
    </div>
  ),
}
