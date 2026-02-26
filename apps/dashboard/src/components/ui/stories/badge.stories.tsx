import type { Meta, StoryObj } from '@storybook/react'
import { Badge } from '../badge'

const meta: Meta<typeof Badge> = {
  title: 'UI/Badge',
  component: Badge,
  args: {
    children: 'Badge',
  },
}

export default meta
type Story = StoryObj<typeof Badge>

export const Default: Story = {}
export const Secondary: Story = { args: { variant: 'secondary' } }
export const Destructive: Story = { args: { variant: 'destructive' } }
export const Outline: Story = { args: { variant: 'outline' } }
export const Info: Story = { args: { variant: 'info' } }
export const Warning: Story = { args: { variant: 'warning' } }
export const Success: Story = { args: { variant: 'success' } }

export const AllVariants: Story = {
  render: () => (
    <div className="flex flex-col gap-2">
      <p className="text-sm font-medium text-muted-foreground">variant</p>
      <div className="flex flex-wrap items-center gap-2">
        <Badge variant="default">Default</Badge>
        <Badge variant="secondary">Secondary</Badge>
        <Badge variant="destructive">Destructive</Badge>
        <Badge variant="outline">Outline</Badge>
        <Badge variant="info">Info</Badge>
        <Badge variant="warning">Warning</Badge>
        <Badge variant="success">Success</Badge>
      </div>
    </div>
  ),
}
