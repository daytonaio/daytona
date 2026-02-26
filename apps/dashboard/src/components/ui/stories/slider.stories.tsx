import type { Meta, StoryObj } from '@storybook/react'
import { Slider } from '../slider'

const meta: Meta<typeof Slider> = {
  title: 'UI/Slider',
  component: Slider,
  decorators: [
    (Story) => (
      <div className="w-64">
        <Story />
      </div>
    ),
  ],
}

export default meta
type Story = StoryObj<typeof Slider>

export const Default: Story = { args: { defaultValue: [50] } }
export const WithRange: Story = { args: { defaultValue: [25], max: 100, step: 1 } }
