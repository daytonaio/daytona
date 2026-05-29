/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { CopyButton } from '@/components/CopyButton'
import { Accordion, AccordionContent, AccordionItem, AccordionTrigger } from '@/components/ui/accordion'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { FallbackProps } from 'react-error-boundary'

export function ErrorBoundaryFallback({ error, resetErrorBoundary }: Partial<FallbackProps>) {
  return (
    <Dialog open>
      <DialogContent className="max-h-[calc(100svh-4rem)] overflow-hidden [&>button]:hidden">
        <DialogHeader className="shrink-0">
          <DialogTitle>Something went wrong</DialogTitle>
          <DialogDescription>
            We're having trouble loading the dashboard. This could be due to a temporary service issue or network
            problem. Please try again or contact support if the issue persists.
          </DialogDescription>
        </DialogHeader>

        <div className="scrollbar-sm -mr-2 min-h-0 flex-1 space-y-4 overflow-y-auto pr-2">
          <Alert variant="destructive">
            <AlertTitle>Error Details:</AlertTitle>
            <AlertDescription>
              <p className="break-all">{error?.message || 'Unknown error'}</p>
            </AlertDescription>
          </Alert>

          {error?.stack && (
            <Accordion type="single" collapsible className="rounded-lg border border-border bg-muted/40">
              <AccordionItem value="stack-trace" className="border-b-0">
                <AccordionTrigger
                  className="px-4 py-3 text-sm font-semibold hover:no-underline"
                  right={
                    <div className="pr-2">
                      <CopyButton value={error.stack} size="icon-xs" tooltipText="Copy stack trace" />
                    </div>
                  }
                >
                  Stack Trace
                </AccordionTrigger>
                <AccordionContent className="px-4 pb-4 pt-0">
                  <pre className="scrollbar-sm max-h-48 overflow-auto whitespace-pre-wrap break-words font-mono text-xs text-muted-foreground">
                    {error.stack}
                  </pre>
                </AccordionContent>
              </AccordionItem>
            </Accordion>
          )}
        </div>

        <div className="flex shrink-0 gap-2 justify-end">
          <Button variant="outline" onClick={() => window.location.reload()}>
            Reload Page
          </Button>
          {resetErrorBoundary && (
            <Button variant="outline" onClick={resetErrorBoundary}>
              Try Again
            </Button>
          )}
        </div>
      </DialogContent>
    </Dialog>
  )
}
