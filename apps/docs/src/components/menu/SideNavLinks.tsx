import { GTProvider, T } from 'gt-react'
import loadTranslations from 'src/i18n/loadTranslations'

import gtConfig from '../../../gt.config.json'

const SideNavLinksContent = () => {
  return (
    <T>
      <div className="nav-item call">
        <a
          href="https://app.daytona.io"
          target="_blank"
          className="nav__link"
          rel="noreferrer"
        >
          Sign in
        </a>
      </div>
    </T>
  )
}

export const SideNavLinks = ({ locale }: { locale: string }) => {
  return (
    <GTProvider
      config={gtConfig}
      loadTranslations={loadTranslations}
      locale={locale}
      projectId={import.meta.env.PUBLIC_VITE_GT_PROJECT_ID}
      devApiKey={import.meta.env.PUBLIC_VITE_GT_API_KEY}
    >
      <SideNavLinksContent />
    </GTProvider>
  )
}
