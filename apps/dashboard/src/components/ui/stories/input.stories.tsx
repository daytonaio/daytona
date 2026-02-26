import type { Meta, StoryObj } from '@storybook/react'
import { Input } from '../input'

const meta: Meta<typeof Input> = {
  title: 'UI/Input',
  component: Input,
}

export default meta
type Story = StoryObj<typeof Input>

export const Default: Story = { args: { placeholder: 'Enter text...' } }
export const WithValue: Story = { args: { defaultValue: 'Hello World' } }
export const Disabled: Story = { args: { placeholder: 'Disabled', disabled: true } }
export const File: Story = { args: { type: 'file' } }
