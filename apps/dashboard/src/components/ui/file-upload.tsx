/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

'use client'

import { Slot } from '@radix-ui/react-slot'
import { File as FileIcon, Trash2, Upload } from 'lucide-react'
import {
  type ComponentProps,
  type ReactNode,
  createContext,
  useCallback,
  useContext,
  useEffect,
  useId,
  useMemo,
  useRef,
  useState,
} from 'react'

import { cn } from '@/lib/utils'

class FileUploadError extends Error {
  constructor(message: string) {
    super(message)
    this.name = 'FileUploadError'
  }
}

class FileUploadSizeError extends FileUploadError {
  constructor(message: string) {
    super(message)
    this.name = 'FileUploadSizeError'
  }
}

class FileUploadMaxFilesError extends FileUploadError {
  constructor(message: string) {
    super(message)
    this.name = 'FileUploadMaxFilesError'
  }
}

class FileUploadAcceptError extends FileUploadError {
  constructor(message: string) {
    super(message)
    this.name = 'FileUploadAcceptError'
  }
}

type RejectedFile = File & { cause: FileUploadError }

function isImage(file: File) {
  return file.type.startsWith('image/')
}

function isAcceptedType(file: File, accept?: string) {
  if (!accept) {
    return true
  }

  const acceptTypes = accept.split(',').map((type) => type.trim())
  const fileType = file.type
  const parts = file.name.split('.')
  const extension = parts.length > 1 ? `.${parts.pop()?.toLowerCase()}` : ''

  return acceptTypes.some((type) => {
    if (type.endsWith('/*')) {
      return fileType.startsWith(type.slice(0, -1))
    }

    return type === fileType || (!!extension && type.toLowerCase() === extension)
  })
}

function toFileList(files: File[]) {
  const dataTransfer = new DataTransfer()

  for (const file of files) {
    dataTransfer.items.add(file)
  }

  return dataTransfer.files
}

function getInputElement(inputId: string) {
  const element = document.getElementById(inputId)
  return element instanceof HTMLInputElement ? element : null
}

type FileUploadContextValue = {
  accept?: string
  disabled: boolean
  inputId: string
  isDragOver: boolean
  maxFiles?: number
  maxSize?: number
  multiple: boolean
  onFileRemove: (file: File) => void
  openFileDialog: () => void
  required: boolean
  setIsDragOver: (isDragOver: boolean) => void
  value: File[]
}

const FileUploadContext = createContext<FileUploadContextValue | null>(null)

function useFileUploadContext() {
  const context = useContext(FileUploadContext)

  if (!context) {
    throw new Error('useFileUploadContext must be used within a FileUpload.')
  }

  return context
}

type FileUploadProps = Omit<ComponentProps<'input'>, 'children' | 'defaultValue' | 'onChange' | 'type' | 'value'> & {
  children?: ReactNode
  defaultValue?: File[]
  value?: File[]
  onChange?: (files: File[]) => void
  onFilesSelected?: (files: File[]) => void
  onReject?: (files: RejectedFile[]) => void
  maxFiles?: number
  maxSize?: number
}

function FileUpload({
  accept,
  children,
  className,
  defaultValue,
  disabled = false,
  id,
  maxFiles,
  maxSize,
  multiple = true,
  name,
  onChange,
  onFilesSelected,
  onReject,
  required = false,
  value: controlledValue,
  ...props
}: FileUploadProps) {
  const generatedId = useId()
  const inputId = id ?? generatedId
  const inputRef = useRef<HTMLInputElement>(null)
  const [internalValue, setInternalValue] = useState(defaultValue ?? [])
  const [isDragOver, setIsDragOver] = useState(false)
  const value = controlledValue ?? internalValue
  const isControlled = controlledValue !== undefined

  const updateValue = useCallback(
    (nextValue: File[]) => {
      if (!isControlled) {
        setInternalValue(nextValue)
      }

      onChange?.(nextValue)
    },
    [isControlled, onChange],
  )

  const onFileRemove = useCallback(
    (file: File) => {
      updateValue(value.filter((candidate) => candidate !== file))
    },
    [updateValue, value],
  )

  const processFiles = useCallback(
    (inputFiles: File[]) => {
      if (disabled) {
        return { acceptedFiles: [], rejectedFiles: [] as RejectedFile[] }
      }

      const acceptedFiles = new Set<File>()
      const rejectedFiles = new Set<RejectedFile>()
      const hasMaxFiles = typeof maxFiles === 'number' && maxFiles > 0
      const isSingleFileMode = !multiple || maxFiles === 1
      const effectiveMaxFiles = hasMaxFiles ? maxFiles : multiple ? Number.POSITIVE_INFINITY : 1

      let totalFiles = isSingleFileMode ? 0 : value.length

      for (const file of inputFiles) {
        if (totalFiles >= effectiveMaxFiles) {
          rejectedFiles.add(
            Object.assign(file, {
              cause: new FileUploadMaxFilesError('Max files reached.'),
            }),
          )
          continue
        }

        if (maxSize && file.size > maxSize) {
          rejectedFiles.add(
            Object.assign(file, {
              cause: new FileUploadSizeError('File size exceeds the maximum allowed size.'),
            }),
          )
          continue
        }

        if (accept && !isAcceptedType(file, accept)) {
          rejectedFiles.add(
            Object.assign(file, {
              cause: new FileUploadAcceptError('File type not accepted.'),
            }),
          )
          continue
        }

        acceptedFiles.add(file)
        totalFiles += 1
      }

      return {
        acceptedFiles: Array.from(acceptedFiles),
        rejectedFiles: Array.from(rejectedFiles),
      }
    },
    [accept, disabled, maxFiles, maxSize, multiple, value.length],
  )

  const commitFiles = useCallback(
    (files: File[]) => {
      const { acceptedFiles, rejectedFiles } = processFiles(files)

      if (acceptedFiles.length > 0) {
        const isSingleFileMode = !multiple || maxFiles === 1
        const nextValue = isSingleFileMode ? acceptedFiles.slice(0, 1) : [...value, ...acceptedFiles]
        updateValue(nextValue)
        onFilesSelected?.(acceptedFiles)
      }

      if (rejectedFiles.length > 0) {
        onReject?.(rejectedFiles)
      }
    },
    [maxFiles, multiple, onFilesSelected, onReject, processFiles, updateValue, value],
  )

  useEffect(() => {
    const input = inputRef.current
    if (!input) {
      return
    }

    input.files = toFileList(value)
  }, [value])

  const openFileDialog = useCallback(() => {
    if (disabled) {
      return
    }

    inputRef.current?.click()
  }, [disabled])

  const contextValue = useMemo<FileUploadContextValue>(
    () => ({
      accept,
      disabled,
      inputId,
      isDragOver,
      maxFiles,
      maxSize,
      multiple,
      onFileRemove,
      openFileDialog,
      required,
      setIsDragOver,
      value,
    }),
    [accept, disabled, inputId, isDragOver, maxFiles, maxSize, multiple, onFileRemove, openFileDialog, required, value],
  )

  return (
    <FileUploadContext.Provider value={contextValue}>
      <div className={className}>
        <input
          {...props}
          ref={inputRef}
          id={inputId}
          type="file"
          accept={accept}
          name={name}
          multiple={multiple}
          disabled={disabled}
          required={required && value.length === 0}
          className="sr-only"
          onChange={(event) => {
            const files = Array.from(event.target.files ?? [])
            commitFiles(files)
            event.target.value = ''
          }}
        />
        {children}
      </div>
    </FileUploadContext.Provider>
  )
}

type FileUploadDropzoneOutlineProps = ComponentProps<'svg'> & {
  isAnimating?: boolean
  strokeWidth?: number
  dashLength?: number
  gapLength?: number
  radius?: number
}

function FileUploadDropzoneOutline({
  className,
  dashLength = 5,
  gapLength = 4,
  isAnimating = false,
  radius = 8,
  strokeWidth = 1.5,
  ...props
}: FileUploadDropzoneOutlineProps) {
  return (
    <svg
      aria-hidden="true"
      data-slot="file-upload-dropzone-outline"
      className={cn('pointer-events-none absolute inset-0 size-full', className)}
      fill="none"
      {...props}
    >
      <rect
        x={strokeWidth / 2}
        y={strokeWidth / 2}
        width={`calc(100% - ${strokeWidth}px)`}
        height={`calc(100% - ${strokeWidth}px)`}
        rx={radius}
        ry={radius}
        strokeWidth={strokeWidth}
        strokeLinecap="round"
        strokeDasharray={`${dashLength} ${gapLength}`}
        className={cn('stroke-current', isAnimating && 'animate-pulse')}
      />
    </svg>
  )
}

type FileUploadDropzoneProps = ComponentProps<'div'> & {
  asChild?: boolean
}

function FileUploadDropzone({
  asChild = false,
  children,
  className,
  onClick,
  onDragEnter,
  onDragLeave,
  onDragOver,
  onDrop,
  onKeyDown,
  onPaste,
  ...props
}: FileUploadDropzoneProps) {
  const { disabled, inputId, isDragOver, openFileDialog, setIsDragOver } = useFileUploadContext()
  const Comp = asChild ? Slot : 'div'

  return (
    <Comp
      data-slot="file-upload-dropzone"
      data-state={isDragOver ? 'over' : 'idle'}
      role="button"
      tabIndex={disabled ? -1 : 0}
      aria-disabled={disabled}
      className={cn(
        'relative flex flex-col items-center gap-2 rounded-2xl px-4 py-6 text-center transition-colors select-none',
        'hover:bg-muted/40 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring/40',
        'data-[state=over]:bg-muted/60',
        disabled && 'cursor-not-allowed opacity-60 hover:bg-transparent',
        className,
      )}
      onClick={(event) => {
        onClick?.(event)
        if (!event.defaultPrevented) {
          openFileDialog()
        }
      }}
      onKeyDown={(event) => {
        onKeyDown?.(event)
        if (event.defaultPrevented || disabled) {
          return
        }

        if (event.key === 'Enter' || event.key === ' ') {
          event.preventDefault()
          openFileDialog()
        }
      }}
      onDragEnter={(event) => {
        onDragEnter?.(event)
        if (disabled) {
          return
        }
        event.preventDefault()
        setIsDragOver(true)
      }}
      onDragOver={(event) => {
        onDragOver?.(event)
        if (disabled) {
          return
        }
        event.preventDefault()
        setIsDragOver(true)
      }}
      onDragLeave={(event) => {
        onDragLeave?.(event)
        if (disabled) {
          return
        }
        event.preventDefault()
        setIsDragOver(false)
      }}
      onDrop={(event) => {
        onDrop?.(event)
        if (disabled) {
          return
        }
        event.preventDefault()
        setIsDragOver(false)

        const files = Array.from(event.dataTransfer.files ?? [])
        if (files.length === 0) {
          return
        }

        const input = getInputElement(inputId)
        if (!input) {
          return
        }

        input.files = toFileList(files)
        input.dispatchEvent(new Event('change', { bubbles: true }))
      }}
      onPaste={(event) => {
        onPaste?.(event)
        if (disabled) {
          return
        }

        const files: File[] = []
        for (const item of Array.from(event.clipboardData.items)) {
          if (item.kind === 'file') {
            const file = item.getAsFile()
            if (file) {
              files.push(file)
            }
          }
        }

        if (files.length === 0) {
          return
        }

        event.preventDefault()
        const input = getInputElement(inputId)
        if (!input) {
          return
        }

        input.files = toFileList(files)
        input.dispatchEvent(new Event('change', { bubbles: true }))
        setIsDragOver(false)
      }}
      {...props}
    >
      <FileUploadDropzoneOutline isAnimating={isDragOver} className="text-border" />
      {children}
    </Comp>
  )
}

type FileUploadTriggerProps = ComponentProps<'button'> & {
  asChild?: boolean
}

function FileUploadTrigger({ asChild = false, children, onClick, ...props }: FileUploadTriggerProps) {
  const { disabled, openFileDialog } = useFileUploadContext()
  const Comp = asChild ? Slot : 'button'

  return (
    <Comp
      data-slot="file-upload-trigger"
      type="button"
      aria-disabled={disabled}
      onClick={(event) => {
        onClick?.(event)
        if (!event.defaultPrevented) {
          openFileDialog()
        }
      }}
      {...props}
    >
      {children}
    </Comp>
  )
}

type FileUploadListProps = ComponentProps<'ul'> & {
  asChild?: boolean
}

function FileUploadList({ asChild = false, children, className, ...props }: FileUploadListProps) {
  const Comp = asChild ? Slot : 'ul'

  return (
    <Comp data-slot="file-upload-list" className={cn('flex flex-col gap-2 empty:hidden', className)} {...props}>
      {children}
    </Comp>
  )
}

type FileUploadItemContextValue = {
  file: File
}

const FileUploadItemContext = createContext<FileUploadItemContextValue | null>(null)

function useFileUploadItemContext() {
  const context = useContext(FileUploadItemContext)

  if (!context) {
    throw new Error('useFileUploadItemContext must be used within a FileUploadItem.')
  }

  return context
}

type FileUploadItemProps = ComponentProps<'li'> & {
  asChild?: boolean
  file: File
}

function FileUploadItem({ asChild = false, children, className, file, ...props }: FileUploadItemProps) {
  const Comp = asChild ? Slot : 'li'
  const value = useMemo(() => ({ file }), [file])

  return (
    <FileUploadItemContext.Provider value={value}>
      <Comp
        data-slot="file-upload-item"
        className={cn('flex items-center gap-3 rounded-2xl border border-border bg-background p-3', className)}
        {...props}
      >
        {children}
      </Comp>
    </FileUploadItemContext.Provider>
  )
}

type FileUploadItemPreviewProps = ComponentProps<'div'> & {
  asChild?: boolean
}

function FileUploadItemPreview({ asChild = false, children, className, ...props }: FileUploadItemPreviewProps) {
  const { file } = useFileUploadItemContext()
  const Comp = asChild ? Slot : 'div'
  const [objectUrl, setObjectUrl] = useState<string | null>(null)

  useEffect(() => {
    let url: string | null = null

    if (isImage(file)) {
      url = URL.createObjectURL(file)
      setObjectUrl(url)
    } else {
      setObjectUrl(null)
    }

    return () => {
      if (url) {
        URL.revokeObjectURL(url)
      }
    }
  }, [file])

  return (
    <Comp
      data-slot="file-upload-item-preview"
      className={cn(
        'relative flex size-12 items-center justify-center overflow-hidden rounded-xl border border-border bg-muted',
        className,
      )}
      {...props}
    >
      {objectUrl ? (
        <img src={objectUrl} alt="" className="size-full object-cover" />
      ) : (
        <FileIcon className="size-5 text-muted-foreground" />
      )}
      {children}
    </Comp>
  )
}

type FileUploadItemMetadataProps = ComponentProps<'div'> & {
  asChild?: boolean
  unit?: 'KB' | 'MB'
}

function FileUploadItemMetadata({
  asChild = false,
  children,
  className,
  unit = 'MB',
  ...props
}: FileUploadItemMetadataProps) {
  const { file } = useFileUploadItemContext()
  const Comp = asChild ? Slot : 'div'
  const unitSize = Math.pow(1024, unit === 'KB' ? 1 : 2)
  const fileSizeLabel = `${(file.size / unitSize).toFixed(2)} ${unit}`

  return (
    <Comp data-slot="file-upload-item-metadata" className={cn('mr-auto flex min-w-0 flex-col', className)} {...props}>
      {children ?? (
        <>
          <span className="truncate text-sm font-medium text-foreground">{file.name}</span>
          <span className="text-xs text-muted-foreground">{fileSizeLabel}</span>
        </>
      )}
    </Comp>
  )
}

type FileUploadItemDeleteProps = ComponentProps<'button'> & {
  asChild?: boolean
}

function FileUploadItemDelete({ asChild = false, children, className, onClick, ...props }: FileUploadItemDeleteProps) {
  const { onFileRemove } = useFileUploadContext()
  const { file } = useFileUploadItemContext()
  const Comp = asChild ? Slot : 'button'

  return (
    <Comp
      data-slot="file-upload-item-delete"
      type="button"
      className={cn(
        'inline-flex size-9 items-center justify-center rounded-full text-muted-foreground transition-colors hover:bg-muted hover:text-foreground',
        className,
      )}
      onClick={(event) => {
        onClick?.(event)
        if (event.defaultPrevented) {
          return
        }

        event.preventDefault()
        event.stopPropagation()
        onFileRemove(file)
      }}
      {...props}
    >
      {children ?? <Trash2 className="size-4" />}
    </Comp>
  )
}

function FileUploadDropzoneContent({
  className,
  description = 'Drag and drop, paste, or choose files.',
  title = 'Drop files here',
  ...props
}: ComponentProps<'div'> & {
  description?: string
  title?: string
}) {
  return (
    <div className={cn('flex flex-col items-center gap-2 px-4', className)} {...props}>
      <div className="flex size-12 items-center justify-center rounded-full bg-muted text-foreground">
        <Upload className="size-6" />
      </div>
      <div className="flex flex-col gap-1">
        <p className="text-sm font-medium text-foreground">{title}</p>
        <p className="text-xs text-muted-foreground">{description}</p>
      </div>
    </div>
  )
}

export {
  FileUpload,
  FileUploadAcceptError,
  FileUploadDropzone,
  FileUploadDropzoneContent,
  FileUploadDropzoneOutline,
  FileUploadError,
  FileUploadItem,
  FileUploadItemDelete,
  FileUploadItemMetadata,
  FileUploadItemPreview,
  FileUploadList,
  FileUploadMaxFilesError,
  FileUploadSizeError,
  FileUploadTrigger,
  isAcceptedType,
  isImage,
  useFileUploadContext,
  useFileUploadItemContext,
  type RejectedFile,
}
