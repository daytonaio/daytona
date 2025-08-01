/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import React, { useEffect, useState } from 'react'
import { Outlet } from 'react-router-dom'

import { Sidebar } from '@/components/Sidebar'
import { SidebarInset, SidebarProvider, SidebarTrigger } from '@/components/ui/sidebar'
import { Toaster } from '@/components/ui/sonner'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import { VerifyEmailDialog } from '@/components/VerifyEmailDialog'
import { AnnouncementBanner } from '@/components/AnnouncementBanner'
import { LocalStorageKey } from '@/enums/LocalStorageKey'
import { cn } from '@/lib/utils'
import { useBilling } from '@/hooks/useBilling'
import { RoutePath } from '@/enums/RoutePath'
import { useNavigate } from 'react-router-dom'

const Dashboard: React.FC = () => {
  const { selectedOrganization } = useSelectedOrganization()
  const [showVerifyEmailDialog, setShowVerifyEmailDialog] = useState(false)
  const { wallet } = useBilling()
  const navigate = useNavigate()

  useEffect(() => {
    if (
      selectedOrganization?.suspended &&
      selectedOrganization.suspensionReason === 'Please verify your email address'
    ) {
      setShowVerifyEmailDialog(true)
    }
  }, [selectedOrganization])

  useEffect(() => {
    if (!import.meta.env.VITE_BILLING_API_URL) {
      return
    }

    if (!wallet || wallet?.ongoingBalanceCents <= 0) {
      navigate(RoutePath.BILLING_WALLET)
    }
  }, [wallet])

  const bannerText = import.meta.env.VITE_ANNOUNCEMENT_BANNER_TEXT
  const bannerLearnMoreUrl = import.meta.env.VITE_ANNOUNCEMENT_BANNER_LEARN_MORE_URL
  const [isBannerVisible, setIsBannerVisible] = useState(false)

  useEffect(() => {
    if (!bannerText) {
      setIsBannerVisible(false)
      return
    }

    // Check if this announcement has been dismissed
    const dismissedBanners = JSON.parse(localStorage.getItem(LocalStorageKey.AnnouncementBannerDismissed) || '[]')
    const isDismissed = dismissedBanners.includes(bannerText)

    setIsBannerVisible(!isDismissed)
  }, [bannerText])

  const handleDismissBanner = () => {
    // Add this announcement to the dismissed list
    const dismissedBanners = JSON.parse(localStorage.getItem(LocalStorageKey.AnnouncementBannerDismissed) || '[]')
    localStorage.setItem(LocalStorageKey.AnnouncementBannerDismissed, JSON.stringify([...dismissedBanners, bannerText]))

    setIsBannerVisible(false)
  }

  return (
    <div className="relative w-full">
      {isBannerVisible && bannerText && (
        <AnnouncementBanner text={bannerText} onDismiss={handleDismissBanner} learnMoreUrl={bannerLearnMoreUrl} />
      )}
      <SidebarProvider isBannerVisible={isBannerVisible} defaultOpen={true}>
        <Sidebar isBannerVisible={isBannerVisible} />

        <SidebarInset className="overflow-hidden">
          <div className="relative md:hidden px-6 pt-4">
            <SidebarTrigger className="[&_svg]:size-5" />
          </div>
          <div className={cn('w-full min-h-screen overscroll-none', isBannerVisible ? 'md:pt-12' : '')}>
            <Outlet />
          </div>
        </SidebarInset>

        <Toaster />
        <VerifyEmailDialog open={showVerifyEmailDialog} onOpenChange={setShowVerifyEmailDialog} />
      </SidebarProvider>
    </div>
  )
}

export default Dashboard
