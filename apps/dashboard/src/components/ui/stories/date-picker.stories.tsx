import type { Meta, StoryObj } from '@storybook/react'
import { useState } from 'react'
import { DatePicker } from '../date-picker'

const meta: Meta<typeof DatePicker> = {
  title: 'UI/DatePicker',
  component: DatePicker,
}

export default meta
type Story = StoryObj<typeof DatePicker>

export const Default: Story = {
  render: () => {
    const [date, setDate] = useState<Date | undefined>()
    return <DatePicker value={date} onChange={setDate} />
  },
}

export const WithValue: Story = {
  render: () => {
    const [date, setDate] = useState<Date | undefined>(new Date())
    return <DatePicker value={date} onChange={setDate} />
  },
}
