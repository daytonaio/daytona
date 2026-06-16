import type { Metadata } from 'next'
import type { ReactNode } from 'react'
import { CopilotKitRoot } from '@/components/CopilotKitRoot'
import './globals.css'

export const metadata: Metadata = {
  title: 'CopilotKit + Daytona Coding Agent',
  description: 'A generative-UI coding agent that builds Vite apps live in a Daytona sandbox.',
}

export default function RootLayout({ children }: { children: ReactNode }) {
  return (
    <html lang="en">
      <body>
        <CopilotKitRoot>{children}</CopilotKitRoot>
      </body>
    </html>
  )
}
