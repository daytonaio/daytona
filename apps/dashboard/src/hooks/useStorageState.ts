/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import {
  type Dispatch,
  type RefObject,
  type SetStateAction,
  useCallback,
  useMemo,
  useRef,
  useSyncExternalStore,
} from 'react'

type StorageStateStorage = Pick<Storage, 'getItem' | 'removeItem' | 'setItem'>
type StorageStateStorageProvider = () => StorageStateStorage | null | undefined
type StorageStateInitialValue<TValue> = TValue | (() => TValue)

type StorageStateError = {
  error: unknown
  key: string
  operation: 'remove' | 'subscribe' | 'write'
}

type UseStorageStateOptions<TValue> = {
  deserialize?: (storedValue: string) => TValue
  onError?: (error: StorageStateError) => void
  serialize?: (value: TValue) => string
  storage?: StorageStateStorage | StorageStateStorageProvider | null
}

type StorageStateSnapshot<TValue> = {
  hasStoredValue: boolean
  isPersistent: boolean
  value: TValue
}

type StorageStateSnapshotReader<TValue> = {
  clearMemoryValue: () => void
  getServerSnapshot: () => StorageStateSnapshot<TValue>
  getSnapshot: () => StorageStateSnapshot<TValue>
  setMemoryValue: (rawValue: string | null) => void
}

type StorageStateSnapshotSource = {
  kind: 'memory' | 'storage' | 'unavailable'
  rawValue: string | null
  storage: StorageStateStorage | null
}

type StorageStateChangeListener = () => void

type StorageStateChangeDetail = {
  isPersistent: boolean
  key: string
  rawValue: string | null
  storage: StorageStateStorage | null
}

type UseStorageStateResult<TValue> = readonly [
  TValue,
  Dispatch<SetStateAction<TValue>>,
  () => void,
  {
    hasStoredValue: boolean
    isPersistent: boolean
  },
]

const getBrowserLocalStorage: StorageStateStorageProvider = () => {
  if (typeof window === 'undefined') {
    return null
  }

  return window.localStorage
}

const STORAGE_STATE_CHANGE_EVENT = 'daytona:storage-state-change'

function useStorageState<TValue>(
  key: string,
  initialValue: StorageStateInitialValue<TValue>,
  {
    deserialize = deserializeJson,
    onError = reportStorageStateError,
    serialize = serializeJson,
    storage = getBrowserLocalStorage,
  }: UseStorageStateOptions<TValue> = {},
): UseStorageStateResult<TValue> {
  const initialValueRef = useRef(initialValue)
  initialValueRef.current = initialValue

  const snapshotReader = useMemo(
    () =>
      createStorageSnapshotReader(key, initialValueRef, {
        deserialize,
        storage,
      }),
    [deserialize, key, storage],
  )

  const subscribe = useCallback(
    (listener: StorageStateChangeListener) =>
      subscribeStorageValue(key, storage, snapshotReader, listener, { onError }),
    [key, onError, snapshotReader, storage],
  )

  const snapshot = useSyncExternalStore(subscribe, snapshotReader.getSnapshot, snapshotReader.getServerSnapshot)

  const setValue = useCallback<Dispatch<SetStateAction<TValue>>>(
    (nextValueOrUpdater) => {
      const currentSnapshot = snapshotReader.getSnapshot()
      const nextValue =
        typeof nextValueOrUpdater === 'function'
          ? (nextValueOrUpdater as (currentValue: TValue) => TValue)(currentSnapshot.value)
          : nextValueOrUpdater

      writeStorageValue(key, nextValue, snapshotReader, {
        onError,
        serialize,
        storage,
      })
    },
    [key, onError, serialize, snapshotReader, storage],
  )

  const removeValue = useCallback(() => {
    removeStorageValue(key, snapshotReader, { onError, storage })
  }, [key, onError, snapshotReader, storage])

  return [
    snapshot.value,
    setValue,
    removeValue,
    {
      hasStoredValue: snapshot.hasStoredValue,
      isPersistent: snapshot.isPersistent,
    },
  ] as const
}

function createStorageSnapshotReader<TValue>(
  key: string,
  initialValueRef: RefObject<StorageStateInitialValue<TValue>>,
  { deserialize, storage }: Required<Pick<UseStorageStateOptions<TValue>, 'deserialize' | 'storage'>>,
): StorageStateSnapshotReader<TValue> {
  let cachedServerSnapshot: StorageStateSnapshot<TValue> | null = null
  let cachedSnapshot: StorageStateSnapshot<TValue> | null = null
  let cachedSource: StorageStateSnapshotSource | null = null
  let memoryValue: string | null | undefined

  const getInitialSnapshot = (): StorageStateSnapshot<TValue> => ({
    hasStoredValue: false,
    isPersistent: false,
    value: resolveInitialValue(initialValueRef.current),
  })

  const getServerSnapshot = () => {
    if (!cachedServerSnapshot) {
      cachedServerSnapshot = getInitialSnapshot()
    }

    return cachedServerSnapshot
  }

  const getSnapshot = () => {
    const storageArea = resolveStorage(storage).storage
    let source: StorageStateSnapshotSource

    if (memoryValue !== undefined) {
      source = {
        kind: 'memory',
        rawValue: memoryValue,
        storage: storageArea,
      }
    } else if (!storageArea) {
      source = {
        kind: 'unavailable',
        rawValue: null,
        storage: null,
      }
    } else {
      try {
        source = {
          kind: 'storage',
          rawValue: storageArea.getItem(key),
          storage: storageArea,
        }
      } catch {
        source = {
          kind: 'unavailable',
          rawValue: null,
          storage: storageArea,
        }
      }
    }

    if (cachedSnapshot && isMatchingSnapshotSource(cachedSource, source)) {
      return cachedSnapshot
    }

    cachedSource = source

    if (source.rawValue === null) {
      cachedSnapshot = {
        hasStoredValue: false,
        isPersistent: source.kind === 'storage',
        value: resolveInitialValue(initialValueRef.current),
      }

      return cachedSnapshot
    }

    try {
      cachedSnapshot = {
        hasStoredValue: source.rawValue !== null,
        isPersistent: source.kind === 'storage',
        value: deserialize(source.rawValue),
      }
    } catch {
      cachedSnapshot = getInitialSnapshot()
    }

    return cachedSnapshot
  }

  return {
    clearMemoryValue: () => {
      memoryValue = undefined
    },
    getServerSnapshot,
    getSnapshot,
    setMemoryValue: (rawValue) => {
      memoryValue = rawValue
    },
  }
}

function writeStorageValue<TValue>(
  key: string,
  value: TValue,
  snapshotReader: StorageStateSnapshotReader<TValue>,
  { onError, serialize, storage }: Required<Pick<UseStorageStateOptions<TValue>, 'onError' | 'serialize' | 'storage'>>,
) {
  const storageResolution = resolveStorage(storage)
  const storageArea = storageResolution.storage
  let serializedValue: string

  try {
    serializedValue = serialize(value)
  } catch (error) {
    onError({ error, key, operation: 'write' })
    return false
  }

  if (storageResolution.error) {
    onError({ error: storageResolution.error, key, operation: 'write' })
  }

  if (!storageArea) {
    snapshotReader.setMemoryValue(serializedValue)
    emitStorageValueChange({ isPersistent: false, key, rawValue: serializedValue, storage: storageArea })
    return false
  }

  try {
    storageArea.setItem(key, serializedValue)
    snapshotReader.clearMemoryValue()
    emitStorageValueChange({ isPersistent: true, key, rawValue: serializedValue, storage: storageArea })
    return true
  } catch (error) {
    snapshotReader.setMemoryValue(serializedValue)
    emitStorageValueChange({ isPersistent: false, key, rawValue: serializedValue, storage: storageArea })
    onError({ error, key, operation: 'write' })
    return false
  }
}

function removeStorageValue<TValue>(
  key: string,
  snapshotReader: StorageStateSnapshotReader<TValue>,
  { onError, storage }: Required<Pick<UseStorageStateOptions<TValue>, 'onError' | 'storage'>>,
) {
  const storageResolution = resolveStorage(storage)
  const storageArea = storageResolution.storage

  if (storageResolution.error) {
    onError({ error: storageResolution.error, key, operation: 'remove' })
  }

  if (!storageArea) {
    snapshotReader.setMemoryValue(null)
    emitStorageValueChange({ isPersistent: false, key, rawValue: null, storage: storageArea })
    return false
  }

  try {
    storageArea.removeItem(key)
    snapshotReader.clearMemoryValue()
    emitStorageValueChange({ isPersistent: true, key, rawValue: null, storage: storageArea })
    return true
  } catch (error) {
    snapshotReader.setMemoryValue(null)
    emitStorageValueChange({ isPersistent: false, key, rawValue: null, storage: storageArea })
    onError({ error, key, operation: 'remove' })
    return false
  }
}

function subscribeStorageValue(
  key: string,
  storage: StorageStateStorage | StorageStateStorageProvider | null | undefined,
  snapshotReader: StorageStateSnapshotReader<unknown>,
  listener: StorageStateChangeListener,
  { onError }: Required<Pick<UseStorageStateOptions<unknown>, 'onError'>>,
) {
  if (typeof window === 'undefined') {
    return () => undefined
  }

  const storageResolution = resolveStorage(storage)
  const storageArea = storageResolution.storage

  if (storageResolution.error) {
    onError({ error: storageResolution.error, key, operation: 'subscribe' })
  }

  const handleStorageStateChange = (event: Event) => {
    const detail = (event as CustomEvent<StorageStateChangeDetail>).detail

    if (!isStorageStateChangeDetail(detail) || detail.key !== key || detail.storage !== storageArea) {
      return
    }

    if (detail.isPersistent) {
      snapshotReader.clearMemoryValue()
    } else {
      snapshotReader.setMemoryValue(detail.rawValue)
    }
    listener()
  }

  const handleBrowserStorageChange = (event: StorageEvent) => {
    if (!isStorageEventForValue(event, key, storageArea)) {
      return
    }

    snapshotReader.clearMemoryValue()
    listener()
  }

  window.addEventListener(STORAGE_STATE_CHANGE_EVENT, handleStorageStateChange)
  window.addEventListener('storage', handleBrowserStorageChange)

  return () => {
    window.removeEventListener(STORAGE_STATE_CHANGE_EVENT, handleStorageStateChange)
    window.removeEventListener('storage', handleBrowserStorageChange)
  }
}

function emitStorageValueChange(detail: StorageStateChangeDetail) {
  if (typeof window === 'undefined') {
    return
  }

  window.dispatchEvent(new CustomEvent<StorageStateChangeDetail>(STORAGE_STATE_CHANGE_EVENT, { detail }))
}

function isMatchingSnapshotSource(
  previousSource: StorageStateSnapshotSource | null,
  nextSource: StorageStateSnapshotSource,
) {
  return (
    previousSource?.kind === nextSource.kind &&
    previousSource.rawValue === nextSource.rawValue &&
    previousSource.storage === nextSource.storage
  )
}

function isNativeStorage(storage: StorageStateStorage): storage is Storage {
  return typeof Storage !== 'undefined' && storage instanceof Storage
}

function isStorageEventForValue(event: StorageEvent, key: string, storage: StorageStateStorage | null) {
  if (!storage || !isNativeStorage(storage)) {
    return false
  }

  if (event.key !== null && event.key !== key) {
    return false
  }

  return !event.storageArea || storage === event.storageArea
}

function isStorageStateChangeDetail(detail: unknown): detail is StorageStateChangeDetail {
  return Boolean(
    detail &&
      typeof detail === 'object' &&
      'key' in detail &&
      'isPersistent' in detail &&
      'storage' in detail &&
      'rawValue' in detail &&
      typeof detail.key === 'string' &&
      typeof detail.isPersistent === 'boolean' &&
      (typeof detail.rawValue === 'string' || detail.rawValue === null),
  )
}

function resolveStorage(storage: StorageStateStorage | StorageStateStorageProvider | null | undefined) {
  try {
    return {
      storage: (typeof storage === 'function' ? storage() : storage) ?? null,
    }
  } catch (error) {
    return {
      error,
      storage: null,
    }
  }
}

function resolveInitialValue<TValue>(initialValue: StorageStateInitialValue<TValue>) {
  return typeof initialValue === 'function' ? (initialValue as () => TValue)() : initialValue
}

function serializeJson<TValue>(value: TValue) {
  const serializedValue = JSON.stringify(value)

  if (typeof serializedValue !== 'string') {
    throw new Error('Storage state values must be JSON serializable')
  }

  return serializedValue
}

function deserializeJson<TValue>(storedValue: string) {
  return JSON.parse(storedValue) as TValue
}

function reportStorageStateError({ error, key, operation }: StorageStateError) {
  console.error(`Failed to ${operation} storage value for "${key}":`, error)
}

export { getBrowserLocalStorage, useStorageState }
export type {
  StorageStateError,
  StorageStateStorage,
  StorageStateStorageProvider,
  UseStorageStateOptions,
  UseStorageStateResult,
}
