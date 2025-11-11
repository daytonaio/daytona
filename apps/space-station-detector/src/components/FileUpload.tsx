import { useRef } from 'react'

interface FileUploadProps {
  onFileSelect: (file: File) => void
  accept?: string
  disabled?: boolean
}

export function FileUpload({
  onFileSelect,
  accept = 'image/*',
  disabled = false,
}: FileUploadProps) {
  const inputRef = useRef<HTMLInputElement>(null)

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (file) {
      onFileSelect(file)
    }
  }

  const handleClick = () => {
    inputRef.current?.click()
  }

  return (
    <div className="w-full">
      <input
        ref={inputRef}
        type="file"
        accept={accept}
        onChange={handleFileChange}
        className="hidden"
        disabled={disabled}
      />
      <button
        onClick={handleClick}
        disabled={disabled}
        className="w-full px-6 py-4 bg-primary hover:bg-primary/90 disabled:bg-muted disabled:cursor-not-allowed text-white font-semibold rounded-lg transition-all duration-200 transform hover:scale-105 active:scale-95"
      >
        <div className="flex items-center justify-center gap-3">
          <svg
            className="w-6 h-6"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12"
            />
          </svg>
          <span>Upload Image</span>
        </div>
      </button>
    </div>
  )
}
