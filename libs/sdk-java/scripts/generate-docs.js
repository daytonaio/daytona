#!/usr/bin/env node

// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

const fs = require('fs')
const path = require('path')

const SDK_SOURCE_DIR = path.resolve(__dirname, '../src/main/java/io/daytona/sdk')
const DOCS_OUTPUT_DIR = path.resolve(__dirname, '../../../apps/docs/src/content/docs/en/java-sdk')

const DOC_TARGETS = [
  { outputFile: 'daytona.mdx', logName: 'Daytona', classes: [{ file: 'Daytona.java', className: 'Daytona' }] },
  {
    outputFile: 'config.mdx',
    logName: 'DaytonaConfig',
    classes: [{ file: 'DaytonaConfig.java', className: 'DaytonaConfig', includeInner: ['Builder'] }],
  },
  { outputFile: 'sandbox.mdx', logName: 'Sandbox', classes: [{ file: 'Sandbox.java', className: 'Sandbox' }] },
  {
    outputFile: 'process.mdx',
    logName: 'Process',
    classes: [{ file: 'Process.java', className: 'Process' }],
  },
  {
    outputFile: 'file-system.mdx',
    logName: 'FileSystem',
    classes: [{ file: 'FileSystem.java', className: 'FileSystem' }],
  },
  { outputFile: 'git.mdx', logName: 'Git', classes: [{ file: 'Git.java', className: 'Git' }] },
  {
    outputFile: 'snapshot.mdx',
    logName: 'SnapshotService',
    classes: [{ file: 'SnapshotService.java', className: 'SnapshotService' }],
  },
  {
    outputFile: 'volume-service.mdx',
    logName: 'VolumeService',
    classes: [{ file: 'VolumeService.java', className: 'VolumeService' }],
  },
  { outputFile: 'image.mdx', logName: 'Image', classes: [{ file: 'Image.java', className: 'Image' }] },
  {
    outputFile: 'code-interpreter.mdx',
    logName: 'CodeInterpreter',
    classes: [
      { file: 'CodeInterpreter.java', className: 'CodeInterpreter' },
      { file: 'RunCodeOptions.java', className: 'RunCodeOptions' },
    ],
  },
  {
    outputFile: 'lsp-server.mdx',
    logName: 'LspServer',
    classes: [{ file: 'LspServer.java', className: 'LspServer' }],
  },
  {
    outputFile: 'computer-use.mdx',
    logName: 'ComputerUse',
    classes: [{ file: 'ComputerUse.java', className: 'ComputerUse' }],
  },
  {
    outputFile: 'pty-handle.mdx',
    logName: 'PtyHandle',
    classes: [{ file: 'PtyHandle.java', className: 'PtyHandle' }],
  },
  {
    outputFile: 'pty.mdx',
    logName: 'PtyCreateOptions, PtyResult',
    title: 'Pty',
    classes: [
      { file: 'PtyCreateOptions.java', className: 'PtyCreateOptions' },
      { file: 'PtyResult.java', className: 'PtyResult' },
    ],
  },
  {
    outputFile: 'errors.mdx',
    logName: 'Exception Classes',
    title: 'Errors',
    classes: [
      { file: 'exception/DaytonaException.java', className: 'DaytonaException' },
      { file: 'exception/DaytonaAuthenticationException.java', className: 'DaytonaAuthenticationException' },
      { file: 'exception/DaytonaBadRequestException.java', className: 'DaytonaBadRequestException' },
      { file: 'exception/DaytonaConflictException.java', className: 'DaytonaConflictException' },
      { file: 'exception/DaytonaConnectionException.java', className: 'DaytonaConnectionException' },
      { file: 'exception/DaytonaForbiddenException.java', className: 'DaytonaForbiddenException' },
      { file: 'exception/DaytonaNotFoundException.java', className: 'DaytonaNotFoundException' },
      { file: 'exception/DaytonaRateLimitException.java', className: 'DaytonaRateLimitException' },
      { file: 'exception/DaytonaServerException.java', className: 'DaytonaServerException' },
      { file: 'exception/DaytonaTimeoutException.java', className: 'DaytonaTimeoutException' },
      { file: 'exception/DaytonaValidationException.java', className: 'DaytonaValidationException' },
    ],
  },
]

function ensureDir(dir) {
  fs.mkdirSync(dir, { recursive: true })
}

function readUtf8(filePath) {
  return fs.readFileSync(filePath, 'utf8')
}

function normalizeWhitespace(value) {
  return String(value || '')
    .replace(/\s+/g, ' ')
    .trim()
}

function escapeType(type) {
  if (!type) return 'Object'
  return String(type).replace(/</g, '\\<').replace(/>/g, '\\>')
}

function splitTopLevel(value, separator) {
  const parts = []
  const source = String(value || '')
  if (!source.trim()) return parts

  let buffer = ''
  let angle = 0
  let paren = 0
  let square = 0

  for (let i = 0; i < source.length; i += 1) {
    const c = source[i]

    if (c === '<') angle += 1
    if (c === '>') angle = Math.max(0, angle - 1)
    if (c === '(') paren += 1
    if (c === ')') paren = Math.max(0, paren - 1)
    if (c === '[') square += 1
    if (c === ']') square = Math.max(0, square - 1)

    if (c === separator && angle === 0 && paren === 0 && square === 0) {
      parts.push(buffer.trim())
      buffer = ''
      continue
    }

    buffer += c
  }

  if (buffer.trim()) parts.push(buffer.trim())
  return parts
}

function cleanJavadocInline(text) {
  return String(text || '')
    .replace(/\{@code\s+([^}]*)\}/g, '`$1`')
    .replace(/\{@link\s+([^}]*)\}/g, '`$1`')
    .replace(/<\/?code>/g, '`')
    .trim()
}

function parseJavadoc(raw) {
  if (!raw) return { description: '', params: {}, returns: '', throws: [] }

  const lines = raw
    .replace(/^\s*\/\*\*/, '')
    .replace(/\*\/\s*$/, '')
    .split('\n')
    .map((line) => line.replace(/^\s*\*\s?/, '').trimEnd())

  const descriptionLines = []
  const params = {}
  let returns = ''
  const throws = []
  let currentTag = null

  for (const rawLine of lines) {
    const line = rawLine.trim()

    if (!line) {
      if (!currentTag) descriptionLines.push('')
      continue
    }

    if (line.startsWith('@param ')) {
      const body = line.replace(/^@param\s+/, '')
      const idx = body.indexOf(' ')
      const name = idx === -1 ? body.trim() : body.slice(0, idx).trim()
      const text = idx === -1 ? '' : body.slice(idx + 1).trim()
      params[name] = text
      currentTag = { kind: 'param', name }
      continue
    }

    if (line.startsWith('@return')) {
      returns = line.replace(/^@return\s*/, '').trim()
      currentTag = { kind: 'return' }
      continue
    }

    if (line.startsWith('@throws ') || line.startsWith('@exception ')) {
      const body = line.replace(/^@(throws|exception)\s+/, '')
      const idx = body.indexOf(' ')
      const type = idx === -1 ? body.trim() : body.slice(0, idx).trim()
      const text = idx === -1 ? '' : body.slice(idx + 1).trim()
      throws.push({ type, description: text })
      currentTag = { kind: 'throws', index: throws.length - 1 }
      continue
    }

    if (line.startsWith('@')) {
      currentTag = null
      continue
    }

    if (!currentTag) {
      descriptionLines.push(line)
      continue
    }

    if (currentTag.kind === 'param') {
      const prev = params[currentTag.name] || ''
      params[currentTag.name] = prev ? `${prev} ${line}` : line
      continue
    }

    if (currentTag.kind === 'return') {
      returns = returns ? `${returns} ${line}` : line
      continue
    }

    if (currentTag.kind === 'throws') {
      const prev = throws[currentTag.index].description
      throws[currentTag.index].description = prev ? `${prev} ${line}` : line
    }
  }

  const description = descriptionLines
    .join('\n')
    .replace(/\r/g, '')
    .replace(/\n{3,}/g, '\n\n')
    .replace(/[ \t]+$/gm, '')
    .replace(/<pre>\s*\{@code\s*/g, '\n```java\n')
    .replace(/\s*\}\s*<\/pre>/g, '\n```\n')
    .replace(/<pre>/g, '\n```java\n')
    .replace(/<\/pre>/g, '\n```\n')
    .replace(/\{@code\s+([^}]*)\}/g, '`$1`')
    .replace(/\{@link\s+([^}]*)\}/g, '`$1`')
    .replace(/<\/?p>/g, '\n')
    .replace(/<\/?code>/g, '`')
    .trim()

  return {
    description,
    params: Object.fromEntries(Object.entries(params).map(([k, v]) => [k, cleanJavadocInline(v)])),
    returns: cleanJavadocInline(returns),
    throws: throws.map((item) => ({
      type: String(item.type || '').trim(),
      description: cleanJavadocInline(item.description || ''),
    })),
  }
}

function findMatchingBrace(source, openBraceIndex) {
  let depth = 0
  for (let i = openBraceIndex; i < source.length; i += 1) {
    if (source[i] === '{') depth += 1
    if (source[i] === '}') depth -= 1
    if (depth === 0) return i
  }
  return -1
}

function findClassBlock(source, className) {
  const rx = new RegExp(
    `((?:\\s*\/\\*\\*[\\s\\S]*?\\*\\/\\s*)?)\\s*public\\s+(?:static\\s+)?(?:final\\s+)?class\\s+${className}\\b[^\\{]*\\{`,
    'm',
  )
  const match = rx.exec(source)
  if (!match) return null

  const start = match.index
  const declaration = match[0]
  const openBrace = start + declaration.lastIndexOf('{')
  const closeBrace = findMatchingBrace(source, openBrace)
  if (closeBrace < 0) return null

  return {
    className,
    javadoc: parseJavadoc(match[1] || ''),
    bodyStart: openBrace + 1,
    bodyEnd: closeBrace,
    body: source.slice(openBrace + 1, closeBrace),
  }
}

function parseParameters(paramList) {
  const params = []
  for (const part of splitTopLevel(paramList || '', ',')) {
    let token = normalizeWhitespace(part)
    if (!token) continue

    token = token.replace(/@[A-Za-z_][A-Za-z0-9_]*(?:\([^)]*\))?\s*/g, '')
    token = token.replace(/\bfinal\b\s+/g, '').trim()
    if (!token) continue

    const items = token.split(/\s+/)
    if (items.length < 2) {
      params.push({ type: token, name: token })
      continue
    }

    const name = items[items.length - 1].replace(/,$/, '').trim()
    const type = items.slice(0, -1).join(' ').trim()
    params.push({ type, name })
  }
  return params
}

function parsePublicMembers(classBlock) {
  const fields = []
  const constructors = []
  const methods = []

  const lines = classBlock.body.split('\n')
  let depth = 1
  let collecting = false
  let commentBuffer = []
  let pendingJavadocRaw = ''

  for (const line of lines) {
    const trimmed = line.trim()
    const currentDepth = depth

    if (!collecting && currentDepth === 1 && trimmed.startsWith('/**')) {
      collecting = true
      commentBuffer = [line]
      if (trimmed.includes('*/')) {
        collecting = false
        pendingJavadocRaw = commentBuffer.join('\n')
      }
    } else if (collecting) {
      commentBuffer.push(line)
      if (trimmed.includes('*/')) {
        collecting = false
        pendingJavadocRaw = commentBuffer.join('\n')
      }
    } else if (currentDepth === 1) {
      // Strip single-line method bodies: "public Type foo() { return x; }" → "public Type foo() {"
      const normalized = normalizeWhitespace(trimmed).replace(/\{[^{}]*\}\s*$/, '{')

      if (normalized.startsWith('public static class ') || normalized.startsWith('public class ')) {
        pendingJavadocRaw = ''
      } else if (
        normalized.startsWith('public ') &&
        !normalized.startsWith('public enum ') &&
        !normalized.startsWith('public interface ')
      ) {
        if (normalized.includes('(')) {
          const constructorRx = new RegExp(
            `^public\\s+${classBlock.className}\\s*\\(([^)]*)\\)\\s*(?:throws\\s+([^\\{]+))?\\s*\\{?$`,
          )
          const cMatch = normalized.match(constructorRx)
          if (cMatch) {
            constructors.push({
              signature: `public ${classBlock.className}(${(cMatch[1] || '').trim()})${cMatch[2] ? ` throws ${normalizeWhitespace(cMatch[2])}` : ''}`,
              parameters: parseParameters(cMatch[1]),
              throwsTypes: splitTopLevel(cMatch[2] || '', ',').map((v) => normalizeWhitespace(v)),
              javadoc: parseJavadoc(pendingJavadocRaw),
            })
            pendingJavadocRaw = ''
          } else {
            const mMatch = normalized.match(
              /^public\s+(static\s+)?([\w<>, ?\[\].]+?)\s+(\w+)\s*\(([^)]*)\)\s*(?:throws\s+([^\{]+))?\s*\{?$/,
            )
            if (mMatch) {
              methods.push({
                name: mMatch[3],
                returnType: normalizeWhitespace(mMatch[2]),
                signature:
                  `public ${mMatch[1] || ''}${mMatch[2]} ${mMatch[3]}(${(mMatch[4] || '').trim()})${mMatch[5] ? ` throws ${normalizeWhitespace(mMatch[5])}` : ''}`
                    .replace(/\s+/g, ' ')
                    .trim(),
                parameters: parseParameters(mMatch[4]),
                throwsTypes: splitTopLevel(mMatch[5] || '', ',').map((v) => normalizeWhitespace(v)),
                javadoc: parseJavadoc(pendingJavadocRaw),
              })
              pendingJavadocRaw = ''
            }
          }
        } else if (normalized.endsWith(';')) {
          const fMatch = normalized.match(
            /^public\s+(?:static\s+)?(?:final\s+)?([\w<>, ?\[\].]+?)\s+(\w+)\s*(?:=.*)?;$/,
          )
          if (fMatch) {
            fields.push({
              name: fMatch[2],
              type: normalizeWhitespace(fMatch[1]),
              javadoc: parseJavadoc(pendingJavadocRaw),
            })
            pendingJavadocRaw = ''
          }
        }
      }
    }

    for (const c of line) {
      if (c === '{') depth += 1
      if (c === '}') depth -= 1
    }
  }

  return { fields, constructors, methods }
}

function collectThrows(declaredThrows, javadocThrows) {
  const merged = []

  for (const type of declaredThrows || []) {
    const match = (javadocThrows || []).find((t) => t.type.endsWith(type) || type.endsWith(t.type))
    merged.push({ type, description: match ? match.description : '' })
  }

  for (const tag of javadocThrows || []) {
    if (!merged.some((item) => item.type === tag.type)) merged.push(tag)
  }

  return merged
}

function pushParamsSection(lines, parameters, javadocParams) {
  if (!parameters.length) return
  lines.push('**Parameters**:')
  lines.push('')
  for (const param of parameters) {
    const desc = (javadocParams && javadocParams[param.name]) || ''
    lines.push(`- \`${param.name}\` _${escapeType(param.type)}_ - ${desc}`.trimEnd())
  }
  lines.push('')
}

function pushThrowsSection(lines, throwsList) {
  if (!throwsList.length) return
  lines.push('**Throws**:')
  lines.push('')
  for (const item of throwsList) {
    lines.push(`- \`${escapeType(item.type)}\` - ${item.description || ''}`.trimEnd())
  }
  lines.push('')
}

function buildClassDoc(classData) {
  const lines = []
  lines.push(`## ${classData.className}`)
  lines.push('')

  const classDescription = classData.javadoc.description || `${classData.className} class for Daytona SDK.`
  lines.push(classDescription)
  lines.push('')

  if (classData.fields.length) {
    lines.push('**Properties**:')
    lines.push('')
    for (const field of classData.fields) {
      lines.push(`- \`${field.name}\` _${escapeType(field.type)}_ - ${field.javadoc.description || ''}`.trimEnd())
    }
    lines.push('')
  }

  if (classData.constructors.length) {
    lines.push('### Constructors')
    lines.push('')

    for (const ctor of classData.constructors) {
      lines.push(`#### new ${classData.className}()`)
      lines.push('')
      lines.push('```java')
      lines.push(ctor.signature)
      lines.push('```')
      lines.push('')

      if (ctor.javadoc.description) {
        lines.push(ctor.javadoc.description)
        lines.push('')
      }

      pushParamsSection(lines, ctor.parameters, ctor.javadoc.params)
      pushThrowsSection(lines, collectThrows(ctor.throwsTypes, ctor.javadoc.throws))
    }
  }

  if (classData.methods.length) {
    lines.push('### Methods')
    lines.push('')

    for (const method of classData.methods) {
      lines.push(`#### ${method.name}()`)
      lines.push('')
      lines.push('```java')
      lines.push(method.signature)
      lines.push('```')
      lines.push('')

      if (method.javadoc.description) {
        lines.push(method.javadoc.description)
        lines.push('')
      }

      pushParamsSection(lines, method.parameters, method.javadoc.params)

      if (method.returnType && method.returnType !== 'void') {
        lines.push('**Returns**:')
        lines.push('')
        lines.push(`- \`${escapeType(method.returnType)}\` - ${method.javadoc.returns || ''}`.trimEnd())
        lines.push('')
      }

      pushThrowsSection(lines, collectThrows(method.throwsTypes, method.javadoc.throws))
    }
  }

  return lines.join('\n')
}

function renderFrontmatter(title) {
  return `---\ntitle: "${title}"\nhideTitleOnPage: true\n---\n\n`
}

function postProcess(content) {
  return (
    String(content)
      .replace(/\r/g, '')
      .replace(/\n{3,}/g, '\n\n')
      .replace(/```java\n\n/g, '```java\n')
      .replace(/\n\n```/g, '\n```')
      .replace(/[ \t]+$/gm, '')
      .trim() + '\n'
  )
}

function parseTopClass(fileRelativePath, className) {
  const sourcePath = path.join(SDK_SOURCE_DIR, fileRelativePath)
  if (!fs.existsSync(sourcePath)) throw new Error(`Source file not found: ${sourcePath}`)

  const source = readUtf8(sourcePath)
  const classBlock = findClassBlock(source, className)
  if (!classBlock) throw new Error(`Class ${className} not found in ${fileRelativePath}`)

  const members = parsePublicMembers(classBlock)
  return {
    className,
    javadoc: classBlock.javadoc,
    fields: members.fields,
    constructors: members.constructors,
    methods: members.methods,
    source,
  }
}

function parseInnerClass(parentSource, parentClassName, innerClassName) {
  const parent = findClassBlock(parentSource, parentClassName)
  if (!parent) return null

  const classRx = new RegExp(`public\\s+(?:static\\s+)?class\\s+${innerClassName}\\b[^\\{]*\\{`, 'm')
  const classMatch = classRx.exec(parent.body)
  if (!classMatch) return null

  const textBefore = parent.body.slice(0, classMatch.index)
  const javadocRx = /\/\*\*(?:(?!\/\*\*)[\s\S])*?\*\/\s*$/
  const javadocMatch = textBefore.match(javadocRx)

  const relativeStart = classMatch.index
  const declaration = classMatch[0]
  const openBrace = parent.bodyStart + relativeStart + declaration.lastIndexOf('{')
  const closeBrace = findMatchingBrace(parentSource, openBrace)
  if (closeBrace < 0) return null

  const body = parentSource.slice(openBrace + 1, closeBrace)
  const block = { className: innerClassName, body }
  const members = parsePublicMembers(block)

  return {
    className: `${parentClassName}.${innerClassName}`,
    javadoc: parseJavadoc(javadocMatch ? javadocMatch[0] : ''),
    fields: members.fields,
    constructors: members.constructors,
    methods: members.methods,
  }
}

function generateTarget(target) {
  console.log(`📝 Generating docs for ${target.logName}...`)

  const classDocs = []
  for (const classTarget of target.classes) {
    const parsed = parseTopClass(classTarget.file, classTarget.className)
    classDocs.push(parsed)

    if (classTarget.includeInner && classTarget.includeInner.length) {
      for (const innerName of classTarget.includeInner) {
        const innerParsed = parseInnerClass(parsed.source, classTarget.className, innerName)
        if (!innerParsed) throw new Error(`Inner class ${innerName} not found in ${classTarget.className}`)
        classDocs.push(innerParsed)
      }
    }
  }

  const title = target.title || classDocs[0].className
  const body = classDocs.map((item) => buildClassDoc(item)).join('\n\n')
  const content = renderFrontmatter(title) + postProcess(body)

  const outputPath = path.join(DOCS_OUTPUT_DIR, target.outputFile)
  fs.writeFileSync(outputPath, content)
  console.log(`✅ Generated: ${target.outputFile}`)
}

function main() {
  console.log('🚀 Starting documentation generation...')
  console.log(`📂 Output directory: ${DOCS_OUTPUT_DIR}`)
  console.log('')

  ensureDir(DOCS_OUTPUT_DIR)
  for (const target of DOC_TARGETS) generateTarget(target)

  console.log('')
  console.log('✨ Documentation generation complete!')
}

if (require.main === module) main()
