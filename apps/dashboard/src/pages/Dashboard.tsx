/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import React, { useEffect, useMemo, useState } from 'react'
import { Outlet } from 'react-router-dom'

import { AnnouncementBanner } from '@/components/AnnouncementBanner'
import { Sidebar } from '@/components/Sidebar'
import { SidebarInset, SidebarProvider } from '@/components/ui/sidebar'
import { Toaster } from '@/components/ui/sonner'
import { VerifyEmailDialog } from '@/components/VerifyEmailDialog'
import { LocalStorageKey } from '@/enums/LocalStorageKey'
import { RoutePath } from '@/enums/RoutePath'
import { useOwnerWalletQuery } from '@/hooks/queries/billingQueries'
import { useConfig } from '@/hooks/useConfig'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import { useSuspensionBanner } from '@/hooks/useSuspensionBanner'
import { cn } from '@/lib/utils'
import { useNavigate } from 'react-router-dom'

const Dashboard: React.FC = () => {
  const { selectedOrganization } = useSelectedOrganization()
  const [showVerifyEmailDialog, setShowVerifyEmailDialog] = useState(false)
  const config = useConfig()
  useOwnerWalletQuery() // prefetch wallet

  const navigate = useNavigate()

  useSuspensionBanner(selectedOrganization)

  useEffect(() => {
    if (
      selectedOrganization?.suspended &&
      selectedOrganization.suspensionReason === 'Please verify your email address'
    ) {
      setShowVerifyEmailDialog(true)
    }
  }, [selectedOrganization])

  useEffect(() => {
    if (!config.billingApiUrl) {
      return
    }

    if (!selectedOrganization) {
      return
    }

    if (!selectedOrganization.defaultRegionId) {
      navigate(RoutePath.SETTINGS)
      return
    }
  }, [config.billingApiUrl, selectedOrganization]) // Do not depend on navigate to avoid infinite loops

  const [bannerText, bannerLearnMoreUrl] = useMemo(() => {
    if (!config.announcements || Object.entries(config.announcements).length === 0) {
      return [null, null]
    }

    return [Object.values(config.announcements)[0].text, Object.values(config.announcements)[0].learnMoreUrl]
  }, [config.announcements])
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
        <Sidebar isBannerVisible={isBannerVisible} billingEnabled={!!config.billingApiUrl} version={config.version} />
        <SidebarInset className="overflow-y-auto">
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
