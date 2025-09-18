/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import ResponseCard from '../ResponseCard'
import { usePlayground } from '@/hooks/usePlayground'

const VNCDesktopWindowResponse: React.FC = () => {
  const { VNCInteractionOptionsParamsState } = usePlayground()

  return (
    <>
      <Card className="w-full">
        <CardHeader>
          <CardTitle>Desktop window</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="w-full aspect-[4/3] md:aspect-[16/9] bg-white"></div>
        </CardContent>
      </Card>
      <ResponseCard responseText={VNCInteractionOptionsParamsState.responseText || ''} />
    </>
  )
}

export default VNCDesktopWindowResponse
