/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import type { Meta, StoryObj } from '@storybook/react'
import { ScrollArea } from '../scroll-area'

const meta: Meta<typeof ScrollArea> = {
  title: 'UI/ScrollArea',
  component: ScrollArea,
}

export default meta
type Story = StoryObj<typeof ScrollArea>

export const Default: Story = {
  render: () => (
    <ScrollArea className="h-48 w-64 rounded-md border p-4">
      {Array.from({ length: 50 }, (_, i) => (
        <div key={i} className="py-1 text-sm">
          Item {i + 1}
        </div>
      ))}
    </ScrollArea>
  ),
}

export const WithShadowFade: Story = {
  render: () => (
    <ScrollArea className="h-48 w-64 rounded-md border p-4" fade="shadow">
      {Array.from({ length: 50 }, (_, i) => (
        <div key={i} className="py-1 text-sm">
          Item {i + 1}
        </div>
      ))}
    </ScrollArea>
  ),
}
