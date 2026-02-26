import type { Meta, StoryObj } from '@storybook/react'
import { Separator } from '../separator'

const meta: Meta<typeof Separator> = {
  title: 'UI/Separator',
  component: Separator,
}

export default meta
type Story = StoryObj<typeof Separator>

export const Horizontal: Story = {
  decorators: [
    (Story) => (
      <div className="w-64">
        <div className="text-sm">Above</div>
        <Story />
        <div className="text-sm">Below</div>
      </div>
    ),
  ],
}

export const Vertical: Story = {
  args: { orientation: 'vertical' },
  decorators: [
    (Story) => (
      <div className="flex h-8 items-center gap-4">
        <span className="text-sm">Left</span>
        <Story />
        <span className="text-sm">Right</span>
      </div>
    ),
  ],
}
