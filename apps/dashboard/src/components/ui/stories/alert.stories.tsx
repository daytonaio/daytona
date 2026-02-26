import type { Meta, StoryObj } from '@storybook/react'
import { Alert, AlertDescription, AlertTitle } from '../alert'
import { InfoIcon, AlertTriangleIcon, CheckCircleIcon, XCircleIcon } from 'lucide-react'

const meta: Meta<typeof Alert> = {
  title: 'UI/Alert',
  component: Alert,
}

export default meta
type Story = StoryObj<typeof Alert>

export const Default: Story = {
  render: () => (
    <Alert>
      <InfoIcon className="size-4" />
      <AlertTitle>Default Alert</AlertTitle>
      <AlertDescription>This is a default alert message.</AlertDescription>
    </Alert>
  ),
}

export const Destructive: Story = {
  render: () => (
    <Alert variant="destructive">
      <XCircleIcon className="size-4" />
      <AlertTitle>Error</AlertTitle>
      <AlertDescription>Something went wrong.</AlertDescription>
    </Alert>
  ),
}

export const Info: Story = {
  render: () => (
    <Alert variant="info">
      <InfoIcon className="size-4" />
      <AlertTitle>Info</AlertTitle>
      <AlertDescription>Here is some useful information.</AlertDescription>
    </Alert>
  ),
}

export const Warning: Story = {
  render: () => (
    <Alert variant="warning">
      <AlertTriangleIcon className="size-4" />
      <AlertTitle>Warning</AlertTitle>
      <AlertDescription>Please review before continuing.</AlertDescription>
    </Alert>
  ),
}

export const Success: Story = {
  render: () => (
    <Alert variant="success">
      <CheckCircleIcon className="size-4" />
      <AlertTitle>Success</AlertTitle>
      <AlertDescription>Operation completed successfully.</AlertDescription>
    </Alert>
  ),
}

export const AllVariants: Story = {
  render: () => (
    <div className="flex flex-col gap-2 w-[500px]">
      <p className="text-sm font-medium text-muted-foreground">variant</p>
      <div className="flex flex-col gap-4">
        <Alert>
          <InfoIcon className="size-4" />
          <AlertTitle>Default</AlertTitle>
          <AlertDescription>Default alert variant.</AlertDescription>
        </Alert>
        <Alert variant="info">
          <InfoIcon className="size-4" />
          <AlertTitle>Info</AlertTitle>
          <AlertDescription>Informational alert variant.</AlertDescription>
        </Alert>
        <Alert variant="warning">
          <AlertTriangleIcon className="size-4" />
          <AlertTitle>Warning</AlertTitle>
          <AlertDescription>Warning alert variant.</AlertDescription>
        </Alert>
        <Alert variant="destructive">
          <XCircleIcon className="size-4" />
          <AlertTitle>Destructive</AlertTitle>
          <AlertDescription>Destructive alert variant.</AlertDescription>
        </Alert>
        <Alert variant="success">
          <CheckCircleIcon className="size-4" />
          <AlertTitle>Success</AlertTitle>
          <AlertDescription>Success alert variant.</AlertDescription>
        </Alert>
        <Alert variant="neutral">
          <InfoIcon className="size-4" />
          <AlertTitle>Neutral</AlertTitle>
          <AlertDescription>Neutral alert variant.</AlertDescription>
        </Alert>
      </div>
    </div>
  ),
}
