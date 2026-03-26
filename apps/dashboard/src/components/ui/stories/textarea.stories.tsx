/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import type { Meta, StoryObj } from '@storybook/react'
import { Textarea } from '../textarea'

const meta: Meta<typeof Textarea> = {
  title: 'UI/Textarea',
  component: Textarea,
}

export default meta
type Story = StoryObj<typeof Textarea>

export const Default: Story = { args: { placeholder: 'Type your message...' } }
export const Disabled: Story = { args: { placeholder: 'Disabled', disabled: true } }
export const WithValue: Story = { args: { defaultValue: 'Some content here...' } }
