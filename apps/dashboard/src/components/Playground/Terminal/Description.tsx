/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

'use client'

import { CopyButton } from '@/components/CopyButton'
import { TerminalIcon } from 'lucide-react'

const sampleCommands = ['ls -la', 'top', 'ps aux', 'df -h']

const TerminalCommand = ({ value }: { value: string }) => {
  return (
    <div className="bg-muted/50 px-2 py-1 text-muted-foreground rounded-md font-mono flex justify-between group items-center">
      <span>
        <span className="mr-2 opacity-50">$</span>
        {value}
      </span>
      <CopyButton value={value} variant="ghost" className="size-6 [&>svg]:size-3" />
    </div>
  )
}

const TerminalDescription: React.FC = () => {
  return (
    <div className="flex flex-col gap-6">
      <div>
        <h2>Web Terminal</h2>
        <p className="text-muted-foreground text-sm mt-1">
          Run commands, view files, and debug directly in the browser.
        </p>
      </div>
      <div className="text-sm">
        <h3 className="text-muted-foreground flex items-center gap-2 font-semibold py-3">
          <TerminalIcon className="size-4" />
          Common Commands
        </h3>
        <ul className="flex flex-col gap-1.5">
          {sampleCommands.map((cmd) => (
            <li key={cmd}>
              <TerminalCommand value={cmd} />
            </li>
          ))}
        </ul>
      </div>
    </div>
  )
}

export default TerminalDescription
