/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { useTheme } from '@/contexts/ThemeContext'
import { CheckIcon, ClipboardIcon } from 'lucide-react'
import { useEffect, useState } from 'react'
import { Highlight, themes, type PrismTheme, type Token } from 'prism-react-renderer'

interface CodeSnippet {
  label: string
  code: string
}

type CodeBlockProps =
  | {
      code?: string
      snippets?: never
      language: string
      showCopy?: boolean
      selectedSnippetLabel?: never
      onSelectSnippet?: never
    }
  | {
      code?: never
      snippets: CodeSnippet[]
      language: string
      showCopy?: boolean
      selectedSnippetLabel?: string
      onSelectSnippet?: (label: string) => void
    }

interface HighlightProps {
  style: React.CSSProperties
  tokens: Token[][]
  getLineProps: (props: { line: Token[]; key: number }) => React.HTMLAttributes<HTMLDivElement>
  getTokenProps: (props: { token: Token; key: number }) => React.HTMLAttributes<HTMLSpanElement>
}

const CodeBlock: React.FC<CodeBlockProps> = ({
  code,
  snippets,
  language,
  showCopy = true,
  selectedSnippetLabel,
  onSelectSnippet,
}) => {
  const [copied, setCopied] = useState(false)
  const [activeTab, setActiveTab] = useState(0)
  const { theme } = useTheme()

  useEffect(() => {
    if (snippets) {
      const index = selectedSnippetLabel ? snippets.findIndex((snippet) => snippet.label === selectedSnippetLabel) : 0
      if (index !== -1 && index !== activeTab) {
        setActiveTab(index)
      } else if (index === -1 && activeTab !== 0) {
        setActiveTab(0) // Fallback to first tab if selected label is not found
      }
    }
  }, [selectedSnippetLabel, snippets, activeTab])

  const hasMultipleSnippets = snippets && snippets.length > 1
  const currentCode = snippets ? snippets[activeTab].code : code
  const displayCode = currentCode ? currentCode.trim() : ''

  const copyToClipboard = async () => {
    if (displayCode) {
      await navigator.clipboard.writeText(displayCode)
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    }
  }

  return (
    <div className="relative rounded-lg">
      {hasMultipleSnippets && (
        <div
          role="tablist"
          className="flex border-b border-gray-300 dark:border-gray-700 bg-gray-50 dark:bg-gray-900 rounded-t-lg overflow-x-auto"
        >
          {snippets!.map((snippet, idx) => (
            <button
              key={idx}
              type="button"
              role="tab"
              aria-selected={activeTab === idx}
              aria-controls={`code-tabpanel-${idx}`}
              tabIndex={activeTab === idx ? 0 : -1}
              onClick={() => {
                setActiveTab(idx)
                if (onSelectSnippet) {
                  onSelectSnippet(snippet.label)
                }
              }}
              className={`px-4 py-2 font-medium transition-colors whitespace-nowrap ${
                activeTab === idx
                  ? 'border-b-2 border-blue-500 text-blue-600 dark:text-blue-400'
                  : 'text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-gray-200'
              }`}
            >
              {snippet.label}
            </button>
          ))}
        </div>
      )}
      <Highlight
        theme={(theme === 'dark' ? themes.oneDark : themes.oneLight) as PrismTheme}
        code={displayCode}
        language={language}
      >
        {({ style, tokens, getLineProps, getTokenProps }: HighlightProps) => (
          <pre className={`p-4 overflow-x-auto ${hasMultipleSnippets ? 'rounded-b-lg' : 'rounded-lg'}`} style={style}>
            {tokens.map((line, i) => {
              const props = getLineProps({ line, key: i })
              // @ts-expect-error Workaround for the render error. Key should not be spread into JSX
              const { key, ...rest } = props
              return (
                <div key={i} {...rest}>
                  {line.map((token, key) => {
                    const tokenProps = getTokenProps({ token, key })
                    // @ts-expect-error Workaround for the render error. Key should not be spread into JSX
                    const { key: tokenKey, ...restTokenProps } = tokenProps
                    return <span key={tokenKey} {...restTokenProps} />
                  })}
                </div>
              )
            })}
          </pre>
        )}
      </Highlight>
      {showCopy && (
        <button
          onClick={copyToClipboard}
          className="absolute right-2 top-2.5 p-2 text-gray-400 hover:text-white transition-colors"
          aria-label="Copy code"
        >
          {copied ? <CheckIcon className="h-4 w-4 text-green-500" /> : <ClipboardIcon className="h-4 w-4" />}
        </button>
      )}
    </div>
  )
}

export default CodeBlock
