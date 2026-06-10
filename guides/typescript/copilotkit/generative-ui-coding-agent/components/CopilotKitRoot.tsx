'use client'

import { CopilotKit } from '@copilotkit/react-core/v2'
import type { ReactNode } from 'react'
import '@copilotkit/react-core/v2/styles.css'

export function CopilotKitRoot({ children }: { children: ReactNode }) {
  return <CopilotKit runtimeUrl="/api/copilotkit">{children}</CopilotKit>
}
