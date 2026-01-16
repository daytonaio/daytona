import { ComponentProps, ReactNode } from 'react'

import { Button } from './ui/button'
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from './ui/tooltip'

type Props = ComponentProps<typeof Button> & {
  tooltipText: string
  tooltipContent?: ReactNode
  tooltipContainer?: HTMLElement
  side?: ComponentProps<typeof TooltipContent>['side']
}

function TooltipButton({
  tooltipText,
  tooltipContent,
  side = 'top',
  tooltipContainer,
  ref,
  size = 'icon-sm',
  ...props
}: Props) {
  return (
    <TooltipProvider>
      <Tooltip delayDuration={0}>
        <TooltipTrigger asChild>
          <Button ref={ref} {...props} aria-label={tooltipText} />
        </TooltipTrigger>
        <TooltipContent side={side}>{tooltipContent || <div>{tooltipText}</div>}</TooltipContent>
      </Tooltip>
    </TooltipProvider>
  )
}

export default TooltipButton
