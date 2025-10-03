/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { ReactNode } from 'react'

type ResponseCardProps = {
  titleText?: string
  responseText: string | ReactNode
}

const ResponseCard: React.FC<ResponseCardProps> = ({ titleText, responseText }) => {
  return (
    <Card className="w-full">
      <CardHeader>
        <CardTitle>{titleText || 'Response'}</CardTitle>
      </CardHeader>
      <CardContent>
        <div className="rounded-lg">
          <pre className="max-w-full bg-zinc-900 text-zinc-100 h-[250px] p-4 rounded-lg overflow-x-auto overflow-y-auto text-sm font-mono">
            <code>{responseText}</code>
          </pre>
        </div>
      </CardContent>
    </Card>
  )
}

export default ResponseCard
