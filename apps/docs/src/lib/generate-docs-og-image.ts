import { Resvg } from '@resvg/resvg-js'
import fs from 'node:fs'
import path from 'node:path'
import React from 'react'
import satori from 'satori'

const WIDTH = 1248
const HEIGHT = 628
/** Fixed width so multi-line title height is laid out correctly under Satori/Yoga. */
const CONTENT_MAX_WIDTH = 920

function resolvePublicFile(relativePath: string): string {
  const cwd = process.cwd()
  const candidates = [
    path.join(cwd, 'public', relativePath),
    path.join(cwd, 'client', relativePath),
    path.join(cwd, 'apps', 'docs', 'public', relativePath),
    path.join(cwd, 'apps', 'docs', 'client', relativePath),
  ]
  for (const p of candidates) {
    if (fs.existsSync(p)) return p
  }
  throw new Error(`docs OG: missing file ${relativePath}`)
}

function readFontBuffer(filename: string): ArrayBuffer {
  const buf = fs.readFileSync(
    resolvePublicFile(path.join('og-fonts', filename))
  )
  return new Uint8Array(buf).buffer
}

let bgDataUrl: string | undefined

function backgroundDataUrl(): string {
  if (!bgDataUrl) {
    const buf = fs.readFileSync(resolvePublicFile('opengraph.png'))
    bgDataUrl = `data:image/png;base64,${buf.toString('base64')}`
  }
  return bgDataUrl
}

function clampText(s: string, max: number): string {
  const t = s.trim()
  if (t.length <= max) return t
  return `${t.slice(0, max - 1).trimEnd()}...`
}

export async function generateDocsOgImagePng(options: {
  title: string
}): Promise<Buffer> {
  const title = clampText(options.title, 118)

  const fonts = [
    {
      name: 'Inter',
      data: readFontBuffer('Inter-Regular.ttf'),
      weight: 400 as const,
      style: 'normal' as const,
    },
    {
      name: 'Inter',
      data: readFontBuffer('Inter-SemiBold.ttf'),
      weight: 600 as const,
      style: 'normal' as const,
    },
  ]

  const markup = React.createElement(
    'div',
    {
      style: {
        width: WIDTH,
        height: HEIGHT,
        position: 'relative',
        display: 'flex',
        flexDirection: 'row',
        overflow: 'hidden',
        backgroundColor: '#09090b',
      },
    },
    React.createElement('img', {
      src: backgroundDataUrl(),
      width: WIDTH,
      height: HEIGHT,
      style: {
        position: 'absolute',
        top: 0,
        left: 0,
        objectFit: 'cover',
        width: WIDTH,
        height: HEIGHT,
      },
    }),
    React.createElement(
      'div',
      {
        style: {
          position: 'absolute',
          left: 80,
          bottom: 100,
          width: CONTENT_MAX_WIDTH,
          display: 'flex',
          flexDirection: 'column',
          alignItems: 'stretch',
        },
      },
      React.createElement(
        'div',
        {
          style: {
            width: '100%',
            fontSize: 82,
            fontWeight: 500,
            color: '#a1a1aa',
            lineHeight: 1,
            fontFamily: 'Inter',
            letterSpacing: -1,
            textShadow: '0 0 16px rgba(0,0,0,0.75), 0 1px 2px rgba(0,0,0,0.9)',
          },
        },
        'Docs/'
      ),
      React.createElement(
        'div',
        {
          style: {
            marginTop: 24,
            width: '100%',
            fontSize: title.length > 72 ? 40 : 64,
            fontWeight: 500,
            color: '#fafafa',
            lineHeight: 1.15,
            fontFamily: 'Inter',
            letterSpacing: -0.5,
            textShadow: '0 0 20px rgba(0,0,0,0.75), 0 1px 3px rgba(0,0,0,0.9)',
          },
        },
        title
      )
    )
  )

  const svg = await satori(markup, {
    width: WIDTH,
    height: HEIGHT,
    fonts,
  })

  const resvg = new Resvg(svg, {
    fitTo: { mode: 'width', value: WIDTH },
  })
  return resvg.render().asPng()
}
