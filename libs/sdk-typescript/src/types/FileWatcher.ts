/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

/**
 * Represents the type of file system event
 */
export enum FilesystemEventType {
  CREATE = 'CREATE',
  WRITE = 'WRITE',
  DELETE = 'DELETE',
  RENAME = 'RENAME',
  CHMOD = 'CHMOD',
}

/**
 * Represents a file system change event
 */
export interface FilesystemEvent {
  /** Type of the file system event */
  type: FilesystemEventType
  /** Full path to the file or directory that changed */
  name: string
  /** Whether the target is a directory */
  isDir: boolean
  /** Unix timestamp when the event occurred */
  timestamp: number
}

/**
 * Options for configuring file watching
 */
export interface WatchOptions {
  /** Whether to watch subdirectories recursively */
  recursive?: boolean
}

/**
 * Handle for managing a file watcher
 */
export interface WatchHandle {
  /** Stop watching and clean up resources */
  close(): Promise<void>
}

/**
 * Callback function for handling file system events
 */
export type FileWatchCallback = (event: FilesystemEvent) => void | Promise<void>
