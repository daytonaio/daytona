/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { useEffect } from 'react'
import { useAuth } from 'react-oidc-context'
import { create } from 'zustand'

export type ConsentPreferences = {
  necessary: boolean
  analytics: boolean
  preferences: boolean
  marketing: boolean
}

const CONSENT_KEY_PREFIX = 'privacy_consent_'

const DEFAULT_PREFERENCES: ConsentPreferences = {
  necessary: true,
  analytics: false,
  preferences: false,
  marketing: false,
}

interface PrivacyConsentState {
  currentUserSub: string | null
  hasConsented: boolean
  preferences: ConsentPreferences
  init: (userSub: string) => void
  saveConsent: (newPreferences: ConsentPreferences) => void
}

const usePrivacyConsentStore = create<PrivacyConsentState>()((set, get) => ({
  currentUserSub: null,
  hasConsented: false,
  preferences: DEFAULT_PREFERENCES,

  init: (userSub: string) => {
    if (get().currentUserSub === userSub) return
    set({ currentUserSub: userSub })

    const consentKey = `${CONSENT_KEY_PREFIX}${userSub}`
    try {
      const saved = localStorage.getItem(consentKey)
      if (saved) {
        const parsedConsent: ConsentPreferences = JSON.parse(saved)
        set({ preferences: parsedConsent, hasConsented: true })
      } else {
        set({ preferences: DEFAULT_PREFERENCES, hasConsented: false })
      }
    } catch {
      console.error('Error parsing privacy consent')
      set({ preferences: DEFAULT_PREFERENCES, hasConsented: false })
    }
  },

  saveConsent: (newPreferences: ConsentPreferences) => {
    const { currentUserSub } = get()
    if (!currentUserSub) return

    const consentKey = `${CONSENT_KEY_PREFIX}${currentUserSub}`
    localStorage.setItem(consentKey, JSON.stringify(newPreferences))
    set({ preferences: newPreferences, hasConsented: true })
  },
}))

export { usePrivacyConsentStore }

export function usePrivacyConsent() {
  const { user } = useAuth()
  const init = usePrivacyConsentStore((state) => state.init)
  const hasConsented = usePrivacyConsentStore((state) => state.hasConsented)
  const preferences = usePrivacyConsentStore((state) => state.preferences)
  const saveConsent = usePrivacyConsentStore((state) => state.saveConsent)

  useEffect(() => {
    if (user?.profile?.sub) {
      init(user.profile.sub)
    }
  }, [user?.profile?.sub, init])

  return { hasConsented, preferences, saveConsent }
}
