/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { useTheme } from '@/contexts/ThemeContext'
import { cn } from '@/lib/utils'
import { Highlight, themes, type PrismTheme, type Token } from 'prism-react-renderer'
import { CopyButton } from './CopyButton'

interface CodeBlockProps {
  code: string
  language: string
  showCopy?: boolean
  codeAreaClassName?: string
}

interface HighlightProps {
  style: React.CSSProperties
  tokens: Token[][]
  getLineProps: (props: { line: Token[]; key: number }) => React.HTMLAttributes<HTMLDivElement>
  getTokenProps: (props: { token: Token; key: number }) => React.HTMLAttributes<HTMLSpanElement>
}

const oneDark = {
  ...themes.oneDark,
  plain: {
    ...themes.oneDark.plain,
    background: 'hsl(var(--code-background))',
  },
}

const CodeBlock: React.FC<CodeBlockProps> = ({ code, language, showCopy = true, codeAreaClassName }) => {
  const { theme } = useTheme()

  return (
    <div className="relative rounded-lg">
      <Highlight
        theme={(theme === 'dark' ? oneDark : themes.oneLight) as PrismTheme}
        code={code.trim()}
        language={language}
      >
        {({ style, tokens, getLineProps, getTokenProps }: HighlightProps) => (
          <pre className={cn('p-4 rounded-lg overflow-x-auto', codeAreaClassName)} style={style}>
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
        <CopyButton
          value={code.trim()}
          variant="ghost"
          className="absolute text-muted-foreground right-2 top-2.5 p-2"
        />
      )}
    </div>
  )
}

export default CodeBlock
