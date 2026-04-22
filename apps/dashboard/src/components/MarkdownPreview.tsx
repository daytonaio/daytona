/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'

import CodeBlock from '@/components/CodeBlock'
import { cn } from '@/lib/utils'

export function MarkdownPreview({ content, isLoading }: { content: string; isLoading?: boolean }) {
  return (
    <div
      className={cn(
        'scrollbar-sm h-full min-h-0 overflow-auto rounded-md border border-border bg-muted/20 p-4 text-[14px] leading-6 transition-opacity',
        {
          'opacity-50': isLoading,
        },
      )}
    >
      <div
        className={cn(
          '[&_a]:text-foreground [&_a]:underline [&_a]:underline-offset-4',
          '[&_blockquote]:my-4 [&_blockquote]:border-l-2 [&_blockquote]:border-border [&_blockquote]:pl-4',
          '[&_code]:text-foreground',
          '[&_h1]:mb-3 [&_h1]:mt-6 [&_h1]:text-2xl [&_h1]:font-semibold',
          '[&_h2]:mb-3 [&_h2]:mt-6 [&_h2]:text-xl [&_h2]:font-semibold',
          '[&_h3]:mb-3 [&_h3]:mt-5 [&_h3]:text-lg [&_h3]:font-semibold',
          '[&_hr]:my-6 [&_hr]:border-border',
          '[&_li]:text-foreground',
          '[&_ol]:my-3 [&_ol]:list-decimal [&_ol]:pl-6',
          '[&_p]:my-3 [&_p]:text-foreground',
          '[&_strong]:text-foreground',
          '[&_table]:my-4 [&_table]:w-full [&_table]:border-collapse',
          '[&_td]:border [&_td]:border-border [&_td]:px-3 [&_td]:py-2',
          '[&_th]:border [&_th]:border-border [&_th]:px-3 [&_th]:py-2 [&_th]:text-left',
          '[&_ul]:my-3 [&_ul]:list-disc [&_ul]:pl-6',
        )}
      >
        <ReactMarkdown
          remarkPlugins={[remarkGfm]}
          components={{
            a: ({ children, href }) => (
              <a href={href} target="_blank" rel="noopener noreferrer" className="underline underline-offset-4">
                {children}
              </a>
            ),
            code: ({ children, className }) => {
              const languageMatch = /language-([\w-]+)/.exec(className || '')
              const language = languageMatch?.[1]
              const value = String(children).replace(/\n$/, '')

              if (language) {
                return (
                  <CodeBlock
                    code={value}
                    language={language}
                    showCopy={false}
                    className="my-4 border border-border"
                    codeAreaClassName="scrollbar-sm overflow-auto rounded-lg text-[14px] leading-6"
                  />
                )
              }

              return <code className="rounded bg-muted px-1 py-1 font-mono text-[13px]">{children}</code>
            },
          }}
        >
          {content}
        </ReactMarkdown>
      </div>
    </div>
  )
}
