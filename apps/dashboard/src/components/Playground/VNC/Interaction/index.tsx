/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Accordion, AccordionContent, AccordionItem, AccordionTrigger } from '@/components/ui/accordion'
import { VNCInteractionOptionsSectionComponentProps, WrapVNCInvokeApiType } from '@/contexts/PlaygroundContext'
import { VNCInteractionOptionsSections } from '@/enums/Playground'
import { usePlayground } from '@/hooks/usePlayground'
import { usePlaygroundSandbox } from '@/hooks/usePlaygroundSandbox'
import { createErrorMessageOutput } from '@/lib/playground'

import { CameraIcon, KeyboardIcon, MonitorIcon, MousePointer2Icon } from 'lucide-react'
import { useCallback, useState } from 'react'
import VNCDisplayOperations from './Display'
import VNCKeyboardOperations from './Keyboard'
import VNCMouseOperations from './Mouse'
import VNCScreenshotOperations from './Screenshot'

const VNCInteractionOptionsSectionsData = [
  { value: VNCInteractionOptionsSections.DISPLAY, label: 'Display' },
  { value: VNCInteractionOptionsSections.KEYBOARD, label: 'Keyboard' },
  { value: VNCInteractionOptionsSections.MOUSE, label: 'Mouse' },
  { value: VNCInteractionOptionsSections.SCREENSHOT, label: 'Screenshot' },
]

const sectionIcons = {
  [VNCInteractionOptionsSections.DISPLAY]: <MonitorIcon strokeWidth={1.5} />,
  [VNCInteractionOptionsSections.KEYBOARD]: <KeyboardIcon strokeWidth={1.5} />,
  [VNCInteractionOptionsSections.MOUSE]: <MousePointer2Icon strokeWidth={1.5} />,
  [VNCInteractionOptionsSections.SCREENSHOT]: <CameraIcon strokeWidth={1.5} />,
}

const VNCInteractionOptions: React.FC = () => {
  const [openedInteractionOptionsSections, setOpenedInteractionOptionsSections] = useState<
    VNCInteractionOptionsSections[]
  >([VNCInteractionOptionsSections.DISPLAY])
  const { setVNCInteractionOptionsParamValue } = usePlayground()

  const { sandbox, vnc } = usePlaygroundSandbox()

  const ComputerUseClient = vnc.url && sandbox.instance ? sandbox.instance.computerUse : null

  // Standardize VNC invokeAPI call flow with this method
  const wrapVNCInvokeApi = useCallback<WrapVNCInvokeApiType>(
    (invokeApi) => {
      return async (actionFormData) => {
        // Set running message
        setVNCInteractionOptionsParamValue('responseContent', `Running ${actionFormData.methodName}...`)

        try {
          // Call the action API method
          await invokeApi(actionFormData)
        } catch (error) {
          setVNCInteractionOptionsParamValue('responseContent', createErrorMessageOutput(error))
        }
      }
    },
    [setVNCInteractionOptionsParamValue],
  )

  // Disable actions run if there was an error during sandbox creation or during VNC ComputerUse initialization
  const VNCActionsDisabled = !!sandbox.error || !!vnc.error || !ComputerUseClient

  const interactionOptionsSectionComponentProps: VNCInteractionOptionsSectionComponentProps = {
    disableActions: VNCActionsDisabled,
    ComputerUseClient,
    wrapVNCInvokeApi,
  }

  return (
    <div className="flex flex-col gap-6">
      <div>
        <h2>Computer Use</h2>
        <p className="text-sm text-muted-foreground mt-1">
          Automate GUI interactions or manually control the desktop environment.
        </p>
      </div>
      <Accordion
        type="multiple"
        value={openedInteractionOptionsSections}
        onValueChange={(interactionOptionsSections) =>
          setOpenedInteractionOptionsSections(interactionOptionsSections as VNCInteractionOptionsSections[])
        }
      >
        {VNCInteractionOptionsSectionsData.map((section) => {
          const isCollapsed = !openedInteractionOptionsSections.includes(section.value as VNCInteractionOptionsSections)
          return (
            <AccordionItem
              key={section.value}
              value={section.value}
              className="border px-2 last:border-b first:rounded-t-lg last:rounded-b-lg border-t-0 first:border-t"
            >
              <AccordionTrigger className="font-semibold text-muted-foreground hover:no-underline dark:bg-muted/50 bg-muted/80 hover:text-primary py-3 border-b border-b-transparent data-[state=open]:border-b-border -mx-2 px-3">
                <div className="flex items-center gap-2 [&_svg]:size-4 text-sm font-medium">
                  {sectionIcons[section.value]} {section.label}
                </div>
              </AccordionTrigger>
              <AccordionContent className="py-3 px-1">
                {!isCollapsed && (
                  <div className="space-y-4">
                    {section.value === VNCInteractionOptionsSections.DISPLAY && (
                      <VNCDisplayOperations {...interactionOptionsSectionComponentProps} />
                    )}
                    {section.value === VNCInteractionOptionsSections.KEYBOARD && (
                      <VNCKeyboardOperations {...interactionOptionsSectionComponentProps} />
                    )}
                    {section.value === VNCInteractionOptionsSections.MOUSE && (
                      <VNCMouseOperations {...interactionOptionsSectionComponentProps} />
                    )}
                    {section.value === VNCInteractionOptionsSections.SCREENSHOT && (
                      <VNCScreenshotOperations {...interactionOptionsSectionComponentProps} />
                    )}
                  </div>
                )}
              </AccordionContent>
            </AccordionItem>
          )
        })}
      </Accordion>
    </div>
  )
}

export default VNCInteractionOptions
