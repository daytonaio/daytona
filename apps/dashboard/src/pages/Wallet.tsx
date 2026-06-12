/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { BillingInfoCard } from '@/components/BillingInfoCard'
import { PageContent, PageHeader, PageIntro, PageLayout } from '@/components/PageLayout'
import { AutomaticTopUpCard } from '@/components/billing/AutomaticTopUpCard'
import { ChargesTable } from '@/components/billing/Charges'
import { InvoicesTable } from '@/components/billing/Invoices'
import { OneTimeTopUpCard } from '@/components/billing/OneTimeTopUpCard'
import { PaymentMethodsCard } from '@/components/billing/PaymentMethodsCard'
import { WalletOverviewCard } from '@/components/billing/WalletOverviewCard'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Empty, EmptyContent, EmptyDescription, EmptyHeader, EmptyMedia, EmptyTitle } from '@/components/ui/empty'
import { Skeleton } from '@/components/ui/skeleton'
import { Spinner } from '@/components/ui/spinner'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { RoutePath } from '@/enums/RoutePath'
import { useCreateInvoicePaymentUrlMutation } from '@/hooks/mutations/useCreateInvoicePaymentUrlMutation'
import { useOwnerInvoicesQuery, useOwnerWalletQuery } from '@/hooks/queries/billingQueries'
import { useChargesQuery } from '@/hooks/queries/useChargesQuery'
import { usePaymentMethodsQuery } from '@/hooks/queries/usePaymentMethodsQuery'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import { BillingType, type Invoice, type OrganizationWallet } from '@daytona/billing-api-client'
import { SparklesIcon, TriangleAlertIcon } from 'lucide-react'
import { type ReactNode, useCallback, useState } from 'react'
import { useAuth } from 'react-oidc-context'
import { Navigate, useLocation, useNavigate, useParams } from 'react-router'
import { toast } from 'sonner'

const DEFAULT_PAGE_SIZE = 10
const WALLET_PAGE_GENERAL = 'general'
const WALLET_PAGE_HISTORY = 'history'

type WalletPage = typeof WALLET_PAGE_GENERAL | typeof WALLET_PAGE_HISTORY

const WALLET_TABS: { value: WalletPage; label: string }[] = [
  { value: WALLET_PAGE_GENERAL, label: 'General' },
  { value: WALLET_PAGE_HISTORY, label: 'History' },
]

function isWalletPage(page: string | undefined): page is WalletPage {
  return page === WALLET_PAGE_GENERAL || page === WALLET_PAGE_HISTORY
}

const Wallet = () => {
  const { selectedOrganization } = useSelectedOrganization()
  const { user } = useAuth()
  const { page } = useParams()
  const location = useLocation()
  const navigate = useNavigate()
  const [invoicesPagination, setInvoicesPagination] = useState({
    pageIndex: 0,
    pageSize: DEFAULT_PAGE_SIZE,
  })

  const walletQuery = useOwnerWalletQuery({ refetchOnMount: 'always' })
  const invoicesQuery = useOwnerInvoicesQuery(invoicesPagination.pageIndex + 1, invoicesPagination.pageSize)
  const chargesQuery = useChargesQuery({
    organizationId: selectedOrganization?.id ?? '',
    enabled: Boolean(selectedOrganization),
  })
  const paymentMethodsQuery = usePaymentMethodsQuery({
    organizationId: selectedOrganization?.id ?? '',
    enabled: Boolean(selectedOrganization),
  })

  const wallet = walletQuery.data
  const paymentMethods = paymentMethodsQuery.data
  const hasNoPaymentMethod = (paymentMethods?.length ?? 0) === 0
  const createInvoicePaymentUrlMutation = useCreateInvoicePaymentUrlMutation()

  const handlePayInvoice = useCallback(
    async (invoice: Invoice) => {
      if (!selectedOrganization) {
        return
      }

      const newWindow = window.open('', '_blank')
      try {
        const result = await createInvoicePaymentUrlMutation.mutateAsync({
          organizationId: selectedOrganization.id,
          invoiceId: invoice.id ?? '',
        })
        if (newWindow) {
          newWindow.location.href = result.url ?? ''
        }
      } catch (error) {
        newWindow?.close()
        toast.error('Failed to open invoice', {
          description: String(error),
        })
      }
    },
    [selectedOrganization, createInvoicePaymentUrlMutation],
  )

  const handleViewInvoice = useCallback(
    async (invoice: Invoice) => {
      if (!selectedOrganization) {
        return
      }

      window.open(invoice.fileUrl ?? '', '_blank')
    },
    [selectedOrganization],
  )

  const isBillingLoading = walletQuery.isLoading
  const isPostPaid = wallet?.billingType === BillingType.BillingTypePostPaid
  const showCreditCardBonusPrompt = Boolean(
    hasNoPaymentMethod && user?.profile.email_verified && selectedOrganization?.personal,
  )
  const activePage = isWalletPage(page) ? page : null

  const handleChangePage = useCallback(
    (value: string) => {
      if (!isWalletPage(value)) {
        return
      }

      navigate(`${RoutePath.BILLING_WALLET}/${value}${location.search}`)
    },
    [location.search, navigate],
  )

  if (!activePage) {
    return <Navigate to={`${RoutePath.BILLING_WALLET}/${WALLET_PAGE_GENERAL}${location.search}`} replace />
  }

  return (
    <PageLayout>
      <PageHeader />

      <PageContent size="full" className="gap-0 p-0">
        <div className="mx-auto flex w-full max-w-5xl flex-col gap-4 p-4 pb-0">
          <PageIntro title="Wallet" className="mb-4" />
          {isBillingLoading && <WalletSkeleton />}
          {walletQuery.isError && !wallet && (
            <WalletErrorState onRetry={() => walletQuery.refetch()} retrying={walletQuery.isFetching} />
          )}
          {wallet && (
            <>
              <WalletAlerts wallet={wallet} showCreditCardBonusPrompt={showCreditCardBonusPrompt} user={user} />
              <WalletOverviewCard
                organizationId={selectedOrganization?.id}
                wallet={wallet}
                isPostPaid={isPostPaid}
                user={user}
              />
            </>
          )}
        </div>

        {wallet && (
          <WalletTabs activePage={activePage} onValueChange={handleChangePage}>
            <div className="mx-auto w-full max-w-5xl p-4">
              <TabsContent value={WALLET_PAGE_GENERAL} className="flex flex-col gap-4">
                {selectedOrganization && (
                  <>
                    <BillingInfoCard organizationId={selectedOrganization.id} />
                    <PaymentMethodsCard organizationId={selectedOrganization.id} />
                  </>
                )}

                {selectedOrganization && !isPostPaid && (
                  <AutomaticTopUpCard organizationId={selectedOrganization.id} wallet={wallet} />
                )}

                {selectedOrganization && <OneTimeTopUpCard organizationId={selectedOrganization.id} />}
              </TabsContent>

              <TabsContent value={WALLET_PAGE_HISTORY} className="flex flex-col gap-4 pb-[80px]">
                <WalletTableSection
                  title="Invoices"
                  description="View and download your billing invoices. All invoices are automatically generated and sent to your billing emails."
                >
                  <InvoicesTable
                    data={invoicesQuery.data?.items ?? []}
                    pagination={invoicesPagination}
                    pageCount={invoicesQuery.data?.totalPages ?? 0}
                    totalItems={invoicesQuery.data?.totalItems ?? 0}
                    onPaginationChange={setInvoicesPagination}
                    loading={invoicesQuery.isLoading}
                    onViewInvoice={handleViewInvoice}
                    onPayInvoice={handlePayInvoice}
                  />
                </WalletTableSection>

                {selectedOrganization && (
                  <WalletTableSection
                    title="Charges"
                    description="All payment attempts on your organization, including failed ones."
                  >
                    <ChargesTable data={chargesQuery.charges} loading={chargesQuery.isLoading} />
                  </WalletTableSection>
                )}
              </TabsContent>
            </div>
          </WalletTabs>
        )}
      </PageContent>
    </PageLayout>
  )
}

function WalletSkeleton() {
  return (
    <div className="flex flex-col gap-6">
      <Card className="flex flex-col gap-4">
        <CardContent className="flex flex-col gap-4">
          <Skeleton className="h-5 w-full max-w-sm" />
          <div className="flex items-center gap-2">
            <Skeleton className="h-10 flex-1" />
            <Skeleton className="h-10 flex-1" />
          </div>
          <Skeleton className="h-10" />
          <Skeleton className="h-10" />
        </CardContent>
      </Card>
      <Card className="flex flex-col gap-4">
        <CardContent className="flex flex-col gap-4">
          <Skeleton className="h-5 w-full max-w-sm" />
          <div className="flex items-center gap-2">
            <Skeleton className="h-10 flex-1" />
            <Skeleton className="h-10 flex-1" />
          </div>
          <Skeleton className="h-10" />
        </CardContent>
      </Card>
    </div>
  )
}

function WalletErrorState({ onRetry, retrying }: { onRetry: () => void; retrying: boolean }) {
  return (
    <Empty className="flex-none rounded-md border py-12" variant="destructive">
      <EmptyHeader>
        <EmptyMedia variant="icon">
          <TriangleAlertIcon />
        </EmptyMedia>
        <EmptyTitle>Failed to load wallet</EmptyTitle>
        <EmptyDescription>Something went wrong while fetching your wallet. Please try again.</EmptyDescription>
      </EmptyHeader>
      <EmptyContent>
        <Button variant="secondary" size="sm" onClick={onRetry} disabled={retrying}>
          {retrying && <Spinner />}
          Retry
        </Button>
      </EmptyContent>
    </Empty>
  )
}

function WalletAlerts({
  wallet,
  showCreditCardBonusPrompt,
  user,
}: {
  wallet: OrganizationWallet
  showCreditCardBonusPrompt: boolean
  user: ReturnType<typeof useAuth>['user']
}) {
  return (
    <>
      {user && (
        <>
          {!user.profile.email_verified && (
            <Alert variant="info">
              <TriangleAlertIcon />
              <AlertTitle>Verify your email</AlertTitle>
              <AlertDescription>
                {(wallet.balanceCents ?? 0) > 0 ? (
                  <>
                    Please verify your email address to complete your account setup.
                    <br />A verification email was sent to you.
                  </>
                ) : (
                  <>
                    Verify your email address to receive $100 of credits.
                    <br />A verification email was sent to you.
                  </>
                )}
              </AlertDescription>
            </Alert>
          )}
          {showCreditCardBonusPrompt && (
            <Alert variant="neutral">
              <SparklesIcon />
              <AlertDescription>Connect a credit card to receive an additional $100 of credits.</AlertDescription>
            </Alert>
          )}
        </>
      )}
      {wallet.hasFailedOrPendingInvoice && (
        <Alert variant="destructive">
          <TriangleAlertIcon />
          <AlertTitle>Outstanding invoices</AlertTitle>
          <AlertDescription>
            You have failed or pending invoices that need to be resolved before adding new funds. Please review your
            invoices below and complete or void any outstanding payments.
          </AlertDescription>
        </Alert>
      )}
      {wallet.automaticTopUp?.disabled && (
        <Alert variant="destructive">
          <TriangleAlertIcon />
          <AlertTitle>Automatic top-up disabled</AlertTitle>
          <AlertDescription>
            Your automatic top-up was disabled because of a failed payment. Please update your payment method and enable
            it again manually below.
          </AlertDescription>
        </Alert>
      )}
    </>
  )
}

function WalletTabs({
  activePage,
  onValueChange,
  children,
}: {
  activePage: WalletPage
  onValueChange: (value: string) => void
  children: ReactNode
}) {
  return (
    <Tabs value={activePage} onValueChange={onValueChange} className="mt-10 w-full gap-0">
      <div className="border-b border-border">
        <div className="mx-auto max-w-5xl px-4">
          <TabsList variant="underline" className="border-b-0">
            {WALLET_TABS.map((tab) => (
              <TabsTrigger key={tab.value} value={tab.value}>
                {tab.label}
              </TabsTrigger>
            ))}
          </TabsList>
        </div>
      </div>
      {children}
    </Tabs>
  )
}

function WalletTableSection({
  title,
  description,
  children,
}: {
  title: ReactNode
  description: ReactNode
  children: ReactNode
}) {
  return (
    <Card className="w-full">
      <CardHeader>
        <CardTitle>{title}</CardTitle>
        <CardDescription>{description}</CardDescription>
      </CardHeader>
      <CardContent>{children}</CardContent>
    </Card>
  )
}

export default Wallet
