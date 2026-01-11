/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Accordion, AccordionItem, AccordionTrigger, AccordionContent } from '@/components/ui/accordion'
import {
  VNCInteractionOptionsSections,
  VNCInteractionOptionsSectionsData,
  VNCInteractionOptionsSectionComponentProps,
  WrapVNCInvokeApiType,
} from '@/enums/Playground'
import VNCDisplayOperations from './Display'
import VNCKeyboardOperations from './Keyboard'
import VNCMouseOperations from './Mouse'
import VNCScreenshootOperations from './Screenshot'
import { useTemporarySandbox } from '@/hooks/useTemporarySandbox'
import { usePlayground } from '@/hooks/usePlayground'
import { useApi } from '@/hooks/useApi'
import { createErrorMessageOutput } from '@/lib/playground'
import { ComputerUse } from '@daytonaio/sdk'
import { Plus, Minus } from 'lucide-react'
import { useState, useEffect, useCallback } from 'react'

const VNCInteractionOptions: React.FC = () => {
  const [openedInteractionOptionsSections, setOpenedInteractionOptionsSections] = useState<
    VNCInteractionOptionsSections[]
  >([])
  const [disableOnSandboxError, setDisableOnSandboxError] = useState(true)
  const [ComputerUseClient, setComputerUseClient] = useState<ComputerUse | null>(null)

  const { VNCInteractionOptionsParamsState, setVNCInteractionOptionsParamValue } = usePlayground()
  const VNCUrl = VNCInteractionOptionsParamsState.VNCUrl

  const { toolboxApi } = useApi()
  // Create temporary sandbox which will be used for VNC actions
  const VNCTemporarySandboxData = useTemporarySandbox()

  useEffect(() => {
    // Sync VNCDesktopWindowResponse with temporary sandbox creation data
    setVNCInteractionOptionsParamValue('VNCSandboxData', VNCTemporarySandboxData)
    if (VNCTemporarySandboxData.error) setDisableOnSandboxError(true) // In case of sandbox creation error we disable VNC actions run
  }, [setVNCInteractionOptionsParamValue, VNCTemporarySandboxData])

  useEffect(() => {
    // Create ComputerUse client when computer use is initialized for temporary sandbox
    if (!VNCUrl) setComputerUseClient(null)
    else if (VNCTemporarySandboxData.sandbox)
      setComputerUseClient(new ComputerUse(VNCTemporarySandboxData.sandbox.id, toolboxApi))
  }, [VNCUrl, VNCTemporarySandboxData, toolboxApi])

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
  const VNCActionsDisabled = disableOnSandboxError || !ComputerUseClient

  const interactionOptionsSectionComponentProps: VNCInteractionOptionsSectionComponentProps = {
    disableActions: VNCActionsDisabled,
    ComputerUseClient,
    wrapVNCInvokeApi,
  }

  return (
    <div className="flex flex-col space-y-2">
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
            <AccordionItem key={section.value} value={section.value}>
              <AccordionTrigger className="text-lg" icon={isCollapsed ? <Plus /> : <Minus />}>
                {section.label}
              </AccordionTrigger>
              <AccordionContent>
                {!isCollapsed && (
                  <div className="px-2 space-y-4">
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
                      <VNCScreenshootOperations {...interactionOptionsSectionComponentProps} />
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
