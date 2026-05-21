/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { PageContent, PageHeader, PageIntro, PageLayout } from '@/components/PageLayout'
import { Button } from '@/components/ui/button'
import { Empty, EmptyContent, EmptyDescription, EmptyHeader, EmptyMedia, EmptyTitle } from '@/components/ui/empty'
import { Spinner } from '@/components/ui/spinner'
import { useInitializeWebhooksMutation } from '@/hooks/mutations/useInitializeWebhooksMutation'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import { handleApiError } from '@/lib/error-handling'
import { Webhook } from 'lucide-react'
import { toast } from 'sonner'

export function WebhooksGetStarted() {
  const { selectedOrganization } = useSelectedOrganization()
  const initializeMutation = useInitializeWebhooksMutation()

  const handleEnable = async () => {
    if (!selectedOrganization?.id) {
      return
    }
    try {
      await initializeMutation.mutateAsync(selectedOrganization.id)
      toast.success('Webhooks enabled')
    } catch (error) {
      handleApiError(error, 'Failed to enable webhooks')
    }
  }

  return (
    <PageLayout contained>
      <PageHeader />
      <PageContent size="full">
        <PageIntro title="Webhooks" />
        <Empty className="border flex-none py-16">
          <EmptyHeader>
            <EmptyMedia variant="icon">
              <Webhook />
            </EmptyMedia>
            <EmptyTitle>Webhooks aren't enabled yet</EmptyTitle>
            <EmptyDescription>
              Receive real-time notifications when sandboxes, snapshots, and volumes change state. Enable webhooks for
              this organization to start creating endpoints and delivering events.
            </EmptyDescription>
          </EmptyHeader>
          <EmptyContent>
            <Button onClick={handleEnable} disabled={initializeMutation.isPending}>
              {initializeMutation.isPending && <Spinner />}
              Enable webhooks
            </Button>
          </EmptyContent>
        </Empty>
      </PageContent>
    </PageLayout>
  )
}
