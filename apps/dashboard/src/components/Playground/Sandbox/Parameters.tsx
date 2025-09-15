/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Accordion, AccordionItem, AccordionTrigger, AccordionContent } from '@/components/ui/accordion'
import { SandboxParametersSections, sandboxParametersSectionsData } from '@/enums/Playground'
import { Plus, Minus } from 'lucide-react'
import { useState } from 'react'

const SandboxParameters: React.FC = () => {
  const [openedParametersSection, setOpenedParametersSection] = useState<SandboxParametersSections[]>([
    SandboxParametersSections.SANDBOX_MANAGMENT,
  ])

  return (
    <div className="flex flex-col space-y-2">
      <Accordion
        type="multiple"
        value={openedParametersSection}
        onValueChange={(parametersSections) =>
          setOpenedParametersSection(parametersSections as SandboxParametersSections[])
        }
      >
        {sandboxParametersSectionsData.map((section) => {
          const isCollapsed = !openedParametersSection.includes(section.value as SandboxParametersSections)
          return (
            <AccordionItem value={section.value}>
              <AccordionTrigger icon={isCollapsed ? <Plus /> : <Minus />}>{section.label}</AccordionTrigger>
              <AccordionContent>{!isCollapsed && section.label}</AccordionContent>
            </AccordionItem>
          )
        })}
      </Accordion>
    </div>
  )
}

export default SandboxParameters
