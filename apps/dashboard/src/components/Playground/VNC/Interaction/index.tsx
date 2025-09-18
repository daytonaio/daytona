/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Accordion, AccordionItem, AccordionTrigger, AccordionContent } from '@/components/ui/accordion'
import { VNCInteractionOptionsSections, VNCInteractionOptionsSectionsData } from '@/enums/Playground'
import VNCDisplayOperations from './Display'
import VNCKeyboardOperations from './Keyboard'
import VNCMouseOperations from './Mouse'
import VNCScreenshootOperations from './Screenshot'
import { Plus, Minus } from 'lucide-react'
import { useState } from 'react'

const VNCInteractionOptions: React.FC = () => {
  const [openedInteractionOptionsSections, setOpenedInteractionOptionsSections] = useState<
    VNCInteractionOptionsSections[]
  >([])

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
                    {section.value === VNCInteractionOptionsSections.DISPLAY && <VNCDisplayOperations />}
                    {section.value === VNCInteractionOptionsSections.KEYBOARD && <VNCKeyboardOperations />}
                    {section.value === VNCInteractionOptionsSections.MOUSE && <VNCMouseOperations />}
                    {section.value === VNCInteractionOptionsSections.SCREENSHOT && <VNCScreenshootOperations />}
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
