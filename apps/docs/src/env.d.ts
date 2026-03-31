// eslint-disable-next-line
/// <reference path="../.astro/types.d.ts" />

declare global {
  namespace App {
    interface Locals {
      t: import('@astrojs/starlight/utils/createTranslationSystem').I18nT
      starlightRoute: import('@astrojs/starlight/route-data').StarlightRouteData
    }
  }
}

export {}
