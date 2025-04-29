/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import React from 'react'

interface QuotaLineProps {
  current: number
  total: number
}

const QuotaLine: React.FC<QuotaLineProps> = ({ current, total }) => {
  const percentage = Math.min((current / total) * 100, 100)

  return (
    <div className="w-full h-1 bg-muted-foreground rounded-full overflow-hidden">
      <div className="h-full flex">
        {/* Green section (0-60%) */}
        <div className="h-full bg-green-500" style={{ width: `${Math.min(percentage, 60)}%` }} />
        {/* Yellow section (60-90%) */}
        {percentage > 60 && (
          <div className="h-full bg-yellow-500" style={{ width: `${Math.min(percentage - 60, 30)}%` }} />
        )}
        {/* Red section (90-100%) */}
        {percentage > 90 && (
          <div className="h-full bg-red-500" style={{ width: `${Math.min(percentage - 90, 10)}%` }} />
        )}
      </div>
    </div>
  )
}

export default QuotaLine
