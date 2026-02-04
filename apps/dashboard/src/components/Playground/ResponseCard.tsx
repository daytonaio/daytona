/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ReactNode } from 'react'

type ResponseCardProps = {
  responseContent: string | ReactNode
}

const ResponseCard: React.FC<ResponseCardProps> = ({ responseContent }) => {
  return (
    <div className="rounded-lg h-full">
      <pre className="max-w-full h-full p-4 rounded-lg overflow-y-auto text-sm font-mono">
        <code>{responseContent}</code>
      </pre>
    </div>
  )
}

export default ResponseCard
