import type { Meta, StoryObj } from '@storybook/react'
import { Checkbox } from '../checkbox'
import { Label } from '../label'

const meta: Meta<typeof Checkbox> = {
  title: 'UI/Checkbox',
  component: Checkbox,
}

export default meta
type Story = StoryObj<typeof Checkbox>

export const Default: Story = {}
export const Checked: Story = { args: { defaultChecked: true } }
export const Disabled: Story = { args: { disabled: true } }

export const WithLabel: Story = {
  render: () => (
    <div className="flex items-center gap-2">
      <Checkbox id="terms" />
      <Label htmlFor="terms">Accept terms and conditions</Label>
    </div>
  ),
}

export const AllStates: Story = {
  render: () => (
    <div className="flex flex-col gap-2">
      <p className="text-sm font-medium text-muted-foreground">state</p>
      <div className="flex flex-col gap-4">
        <div className="flex items-center gap-2">
          <Checkbox id="unchecked" />
          <Label htmlFor="unchecked">Unchecked</Label>
        </div>
        <div className="flex items-center gap-2">
          <Checkbox id="checked" defaultChecked />
          <Label htmlFor="checked">Checked</Label>
        </div>
        <div className="flex items-center gap-2">
          <Checkbox id="indeterminate" checked="indeterminate" />
          <Label htmlFor="indeterminate">Indeterminate</Label>
        </div>
        <div className="flex items-center gap-2">
          <Checkbox id="disabled" disabled />
          <Label htmlFor="disabled">Disabled</Label>
        </div>
        <div className="flex items-center gap-2">
          <Checkbox id="disabled-checked" disabled defaultChecked />
          <Label htmlFor="disabled-checked">Disabled Checked</Label>
        </div>
      </div>
    </div>
  ),
}
