/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { useTheme } from '@/contexts/ThemeContext'
import { CheckIcon, ClipboardIcon } from 'lucide-react'
import { useState } from 'react'
import { Highlight, themes, type PrismTheme, type Token } from 'prism-react-renderer'

interface CodeBlockProps {
  code: string
  language: string
  showCopy?: boolean
}

interface HighlightProps {
  style: React.CSSProperties
  tokens: Token[][]
  getLineProps: (props: { line: Token[]; key: number }) => React.HTMLAttributes<HTMLDivElement>
  getTokenProps: (props: { token: Token; key: number }) => React.HTMLAttributes<HTMLSpanElement>
}

const CodeBlock: React.FC<CodeBlockProps> = ({ code, language, showCopy = true }) => {
  const [copied, setCopied] = useState(false)
  const { theme } = useTheme()

  const copyToClipboard = async () => {
    await navigator.clipboard.writeText(code)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  return (
    <div className="relative rounded-lg">
      <Highlight
        theme={(theme === 'dark' ? themes.oneDark : themes.oneLight) as PrismTheme}
        code={code.trim()}
        language={language}
      >
        {({ style, tokens, getLineProps, getTokenProps }: HighlightProps) => (
          <pre className="p-4 rounded-lg overflow-x-auto" style={style}>
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
