/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import type { Meta, StoryObj } from '@storybook/react'
import { useState } from 'react'

import { Button } from '../button'
import {
  FileUpload,
  FileUploadDropzone,
  FileUploadDropzoneContent,
  FileUploadItem,
  FileUploadItemDelete,
  FileUploadItemMetadata,
  FileUploadItemPreview,
  FileUploadList,
  FileUploadTrigger,
  type RejectedFile,
} from '../file-upload'

function createMockFile(name: string, type: string, contents: string) {
  return new File([contents], name, { type })
}

function UploadDemo({
  accept,
  defaultFiles = [],
  maxFiles,
  maxSize,
  multiple = true,
}: {
  accept?: string
  defaultFiles?: File[]
  maxFiles?: number
  maxSize?: number
  multiple?: boolean
}) {
  const [files, setFiles] = useState<File[]>(defaultFiles)
  const [rejectedFiles, setRejectedFiles] = useState<RejectedFile[]>([])

  return (
    <div className="max-w-xl">
      <FileUpload
        accept={accept}
        maxFiles={maxFiles}
        maxSize={maxSize}
        multiple={multiple}
        value={files}
        onChange={setFiles}
        onReject={setRejectedFiles}
        className="space-y-4"
      >
        <FileUploadDropzone className="min-h-56 justify-center border border-dashed border-border bg-muted/10">
          <FileUploadDropzoneContent
            title="Click to upload or drop files"
            description="Drag and drop, paste, or browse from your device."
          />
          <div className="mt-2">
            <FileUploadTrigger asChild>
              <Button variant="outline" size="sm">
                Browse files
              </Button>
            </FileUploadTrigger>
          </div>
        </FileUploadDropzone>

        <FileUploadList>
          {files.map((file) => (
            <FileUploadItem key={`${file.name}-${file.size}-${file.lastModified}`} file={file}>
              <FileUploadItemPreview />
              <FileUploadItemMetadata unit="KB" />
              <FileUploadItemDelete />
            </FileUploadItem>
          ))}
        </FileUploadList>

        {rejectedFiles.length > 0 ? (
          <div className="rounded-xl border border-destructive/30 bg-destructive/5 p-3 text-sm">
            <p className="font-medium text-foreground">Rejected files</p>
            <ul className="mt-2 space-y-1 text-muted-foreground">
              {rejectedFiles.map((file) => (
                <li key={`${file.name}-${file.size}-${file.lastModified}`}>
                  {file.name}: {file.cause.message}
                </li>
              ))}
            </ul>
          </div>
        ) : null}
      </FileUpload>
    </div>
  )
}

const meta: Meta<typeof FileUpload> = {
  title: 'UI/FileUpload',
  component: FileUpload,
}

export default meta
type Story = StoryObj<typeof FileUpload>

export const Default: Story = {
  render: () => (
    <UploadDemo
      defaultFiles={[
        createMockFile('notes.md', 'text/markdown', '# Notes\n\nA short markdown file.'),
        createMockFile('avatar.png', 'image/png', 'mock-image'),
      ]}
    />
  ),
}

export const ImagesOnly: Story = {
  render: () => <UploadDemo accept="image/*" maxFiles={2} maxSize={1024 * 1024} />,
}
