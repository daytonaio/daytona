/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Accordion, AccordionItem, AccordionTrigger, AccordionContent } from '@/components/ui/accordion'
import { SandboxParametersSections, sandboxParametersSectionsData } from '@/enums/Playground'
import SandboxFileSystem from './FileSystem'
import SandboxGitOperations from './GitOperations'
import SandboxManagmentParameters from './Managment'
import SandboxProcessCodeExecution from './ProcessCodeExecution'
import { Plus, Minus } from 'lucide-react'
import { useState } from 'react'

const SandboxParameters: React.FC = () => {
  const [openedParametersSections, setOpenedParametersSections] = useState<SandboxParametersSections[]>([
    SandboxParametersSections.SANDBOX_MANAGMENT,
  ])

  return (
    <div className="flex flex-col space-y-2">
      <Accordion
        type="multiple"
        value={openedParametersSections}
        onValueChange={(parametersSections) =>
          setOpenedParametersSections(parametersSections as SandboxParametersSections[])
        }
      >
        {sandboxParametersSectionsData.map((section) => {
          const isCollapsed = !openedParametersSections.includes(section.value as SandboxParametersSections)
          return (
            <AccordionItem key={section.value} value={section.value}>
              <AccordionTrigger className="text-lg" icon={isCollapsed ? <Plus /> : <Minus />}>
                {section.label}
              </AccordionTrigger>
              <AccordionContent>
                {!isCollapsed && (
                  <div className="px-2 space-y-4">
                    {section.value === SandboxParametersSections.FILE_SYSTEM && <SandboxFileSystem />}
                    {section.value === SandboxParametersSections.GIT_OPERATIONS && <SandboxGitOperations />}
                    {section.value === SandboxParametersSections.SANDBOX_MANAGMENT && <SandboxManagmentParameters />}
                    {section.value === SandboxParametersSections.PROCESS_CODE_EXECUTION && (
                      <SandboxProcessCodeExecution />
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

export default SandboxParameters
