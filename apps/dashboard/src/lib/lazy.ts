/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { lazy, type ComponentType, type LazyExoticComponent } from 'react'

type LazyModule<T extends ComponentType<any>> = { default: T }

export type PreloadableLazyComponent<T extends ComponentType<any>> = LazyExoticComponent<T> & {
  preload: () => Promise<LazyModule<T>>
}

export function lazyWithPreload<T extends ComponentType<any>>(
  loadModule: () => Promise<LazyModule<T>>,
  { preload = false }: { preload?: boolean } = {},
): PreloadableLazyComponent<T> {
  let modulePromise: Promise<LazyModule<T>> | null = null

  const load = () => {
    modulePromise ??= loadModule().catch((error) => {
      modulePromise = null
      throw error
    })
    return modulePromise
  }

  const LazyComponent = lazy(load) as PreloadableLazyComponent<T>
  LazyComponent.preload = load

  if (preload) {
    load().catch(() => {
      // React.lazy will surface import failures when the component renders.
    })
  }

  return LazyComponent
}
