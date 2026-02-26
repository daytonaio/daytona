import type { Meta, StoryObj } from '@storybook/react'
import { Spinner } from '../spinner'

const meta: Meta<typeof Spinner> = {
  title: 'UI/Spinner',
  component: Spinner,
}

export default meta
type Story = StoryObj<typeof Spinner>

export const Default: Story = {}
export const Large: Story = { args: { className: 'size-8' } }
