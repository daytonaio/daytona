// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

// @ts-check
/* eslint-disable no-useless-escape */

import { MarkdownPageEvent } from 'typedoc-plugin-markdown'

/**
 * @param {import('typedoc-plugin-markdown').MarkdownApplication} app
 */
export function load(app) {
  // --- TITLE HACK ---
  app.renderer.markdownHooks.on('page.begin', () => {
    // We'll add the title later in the END event
    return '---\ntitle: ""\nhideTitleOnPage: true\n---\n'
  })

  // --- CONTENT HACKS ---
  app.renderer.on(MarkdownPageEvent.END, (page) => {
    if (!page.contents) return

    // Extract title from filename and capitalize first letter of each word
    let title = ''
    if (page.filename) {
      const filename = page.filename
      // Get the last part of the filename (after the last dot)
      const baseFilename = filename.split('/').pop()?.replace(/\.md$/, '').split('.').pop() || ''
      // Split into words and capitalize each word
      const words = baseFilename.split(/[-_]/)
      title = words
        .map((word) => {
          if (word.length === 0) return ''
          return word.charAt(0).toUpperCase() + word.slice(1)
        })
        .join('')
    }

    // Replace the empty title with the actual title
    page.contents = page.contents.replace(/title: ""/, `title: "${title}"`)

    page.contents = transformContent(page.contents)
    page.filename = transformFilename(page.filename)
  })
}

function transformContent(contents) {
  return [
    removeInternalLinks,
    escapePromiseSpecialCharacters,
    transformExtendsSection,
    transformParametersSection,
    transformReturnsSection,
    transformPropertiesSection,
    transformExamplesSection,
    transformEnumSection,
    transformThrowsSection,
    transformTypeDeclarationSection,
    fixFormattingArtifacts,
  ].reduce((acc, fn) => fn(acc), contents)
}

function transformFilename(filename) {
  return filename.replace(/\/([^/]+)\.md$/, (_, name) => {
    const formatted = name
      .split('.')
      .pop()
      .replace(/([a-z])([A-Z])/g, '$1-$2') // Add hyphen between lowercase & uppercase
      .replace(/([A-Z])([A-Z][a-z])/g, '$1-$2') // Add hyphen between uppercase followed by lowercase
      .replace(/([0-9])([A-Za-z])/g, '$1-$2') // Add hyphen between number & letter
      .toLowerCase() // Convert to lowercase
    return `/${formatted}.mdx`
  })
}

function removeInternalLinks(contents) {
  return contents.replace(/\[([^\]]+)]\([^)]+\)/g, '$1')
}

function escapePromiseSpecialCharacters(contents) {
  return contents.replace(/`Promise`\s*\\<((?:`?[^`<>]+`?|<[^<>]+>)*?)>/g, (_match, typeContent) => {
    return '`Promise<' + typeContent.replace(/[`\\]/g, '') + '>`'
  })
}

function transformParametersSection(contents) {
  for (let i = 6; i > 1; i--) {
    let paramsRegex = new RegExp(`\#{${i}} Parameters\\s*\\n\\n([\\s\\S]*?)(?=\\n\#{1,${i}} |$)`, 'g')
    if (i == 6) {
      paramsRegex = new RegExp(
        `\#{${i}} Parameters\\s*\\n\\n([\\s\\S]*?)(?=\\n\#{1,${
          i - 1
        }} |\#{1,${i}} Returns|\#{1,${i}} Example|\#{1,${i}} Examples|$)`,
        'g',
      )
    }
    contents = contents.replace(paramsRegex, (match, paramsContent) => {
      const paramHeadingLevel = i == 6 ? 6 : i + 1
      const headingHashes = '#'.repeat(paramHeadingLevel)
      const headingHashesShorter = Array.from({ length: paramHeadingLevel }, (_, k) => '#'.repeat(k + 1)).join('|')

      const paramBlockRegex = new RegExp(
        `${headingHashes} ([^\\n]+)\\n\\n` + // parameter name
          '([^\\n]+)' + // type line
          `(?:\\n\\n((?:(?!${headingHashes} |${headingHashesShorter} |\\*\\*\\*|#{1,6} ).+[\\r\\n]*)*))?`, // safe multiline description
        'g',
      )

      const parameters = []
      let paramMatch

      while ((paramMatch = paramBlockRegex.exec(paramsContent)) !== null) {
        const [, name, typeLine, rawDescription = ''] = paramMatch

        const lines = rawDescription
          .split('\n')
          .map((line) => line.trim())
          .filter((line) => line.length > 0)

        parameters.push({
          name,
          typeLine,
          mainDescription: lines[0] || '',
          otherLines: lines.slice(1),
        })
      }

      if (parameters.length === 0) return match

      let result = '**Parameters**:\n\n'

      for (const { name, typeLine, mainDescription, otherLines } of parameters) {
        let type = typeLine.replace(/`/g, '').trim()
        type = type.replace(/readonly\s+/, '').trim()
        type = type.replace(/(?<!\\)([*_`[\]()<>|])/g, '\\$1')

        result += `- \`${name}\` _${type}_`
        if (mainDescription) result += ` - ${mainDescription}`
        result += '\n'

        for (const line of otherLines) {
          result += `    ${line}\n`
        }
      }

      return result + '\n'
    })
  }

  return contents
}

function transformReturnsSection(contents) {
  return contents.replace(
    /^#{1,6} Returns\s*\n+`([^`]+)`\n+((?:(?!^#{1,6} |\*\*\*).*\n?)*)/gm,
    (_, type, rawDescription) => {
      const lines = rawDescription
        .split('\n')
        .map((l) => l.trim())
        .filter((l) => l && !/^#{1,6} /.test(l) && l !== '***') // ignore headings and separators

      let result = '**Returns**:\n\n- `' + type + '`'
      if (lines.length > 0) {
        result += ' - ' + lines[0] + '\n'
        for (const line of lines.slice(1)) {
          result += `    ${line}\n`
        }
      } else {
        result += '\n'
      }
      result += '\n'
      return result
    },
  )
}

function transformPropertiesSection(contents) {
  contents = transformPropsOrTypeDeclaration(contents, 'Properties')

  // Move Properties section right after each class/interface description
  const sections = contents.split(/^## /gm)
  const updatedSections = sections.map((section, i) => {
    if (i === 0) return section // Skip content before first ##

    const sectionLines = section.split('\n')
    const classHeader = sectionLines[0].trim()
    const body = sectionLines.slice(1).join('\n')

    const propsMatch = body.match(/\*\*Properties\*\*:\s*\n\n([\s\S]*?)(?=\n###|\n\*\*|\n## |\n# |$)/)
    if (!propsMatch) return '## ' + section

    const fullPropsBlock = propsMatch[0]
    const bodyWithoutProps = body.replace(fullPropsBlock, '').trim()

    let descEnd = bodyWithoutProps.search(/\n{2,}(?=###|\*\*|$)|(?=^\s*$)/m)
    if (descEnd === -1) {
      const trimmed = bodyWithoutProps.trim()

      if (!trimmed.includes('\n') && !trimmed.startsWith('#') && !trimmed.startsWith('**')) {
        descEnd = bodyWithoutProps.length
      }
    }
    let newBody

    if (descEnd !== -1) {
      const desc = bodyWithoutProps.slice(0, descEnd).trim()
      const rest = bodyWithoutProps.slice(descEnd).trim()
      newBody = `${desc}\n\n${fullPropsBlock}\n\n${rest}`
    } else {
      newBody = `${fullPropsBlock}\n\n${bodyWithoutProps}`
    }

    return `## ${classHeader}\n\n${newBody.trim()}`
  })

  return updatedSections.join('\n')
}

function transformExamplesSection(contents) {
  return contents.replace(/^#{1,10}\s*(Example|Examples)$/gm, '**$1:**')
}

function transformExtendsSection(contents) {
  return contents.replace(/^#{1,10}\s*(Extends)$/gm, '**$1:**')
}

function transformEnumSection(contents) {
  // First, find all sections with "Enumeration Members" headings
  const sections = contents.split(/^## /gm)
  let newContent = ''

  for (let i = 0; i < sections.length; i++) {
    const section = sections[i]

    if (i === 0) {
      // This is content before the first ## heading
      newContent += section
      continue
    }

    // Add back the ## that was removed in the split
    const sectionWithHeader = '## ' + section

    // Check if this section contains an enum
    if (sectionWithHeader.includes('### Enumeration Members')) {
      // Split at the enum members heading
      const [headerPart, membersPart] = sectionWithHeader.split('### Enumeration Members')

      // Parse and extract all enum values
      const enumValues = []
      const regex = /#### ([A-Z0-9_]+)[\s\S]*?```ts[\s\S]*?\1:\s*"([^"]+)"/g
      let match

      const memberPartCopy = membersPart
      while ((match = regex.exec(memberPartCopy)) !== null) {
        enumValues.push({
          name: match[1],
          value: match[2],
        })
      }

      // Create the transformed section
      let transformedSection = headerPart + '**Enum Members**:\n\n'

      if (enumValues.length > 0) {
        enumValues.forEach((entry) => {
          transformedSection += `- \`${entry.name}\` ("${entry.value}")\n`
        })
        transformedSection += '\n'
      } else {
        // If we couldn't parse any values, just keep the original content
        transformedSection = sectionWithHeader
      }

      newContent += transformedSection
    } else {
      // Non-enum section, just add it back unchanged
      newContent += sectionWithHeader
    }
  }

  return newContent
}

function transformThrowsSection(contents) {
  // Process "Throws" headers from level 2 to level 7
  for (let level = 2; level <= 7; level++) {
    const throwsHeader = '#'.repeat(level) + ' Throws' // Generate header (e.g., ## Throws, ### Throws)
    const sectionHeaderRegex = new RegExp(`(?=^#{${level - 1}} .+)`, 'gm') // Regex for section start (parent level)
    const throwsRegex = new RegExp(`(\n${throwsHeader}\n)`, 'g') // Matches only the "Throws" header itself

    if (!contents) continue

    // Split document into sections at parent level
    const sections = contents.split(sectionHeaderRegex)

    contents = sections
      .map((section) => {
        if (!section.includes(`\n${throwsHeader}`)) return section // Skip if no "Throws" found at this level

        // Capture all occurrences of "Throws" headers at this specific level
        const throwsMatches = [...section.matchAll(throwsRegex)]

        if (throwsMatches.length <= 1) {
          // Transform single occurrence
          return section.replace(throwsRegex, '\n**Throws**:\n')
        }

        // Keep the first "Throws" header and remove only subsequent ones
        let headerRemovedCount = 0
        const cleanedSection = section.replace(throwsRegex, () => {
          return headerRemovedCount++ === 0 ? '\n**Throws**:\n' : '' // Transform first one to bold, remove others
        })

        return cleanedSection
      })
      .join('')
  }

  return contents
}

function fixFormattingArtifacts(content) {
  return content.replace(/`~~([^`]+?)\?~~`/g, '~~`$1?`~~')
}

function transformTypeDeclarationSection(contents) {
  return transformPropsOrTypeDeclaration(contents, 'Type declaration')
}

function transformPropsOrTypeDeclaration(contents, headerTitle) {
  for (let i = 5; i > 1; i--) {
    contents = contents.replace(
      new RegExp(`\#{${i}} ${headerTitle}\\s*\\n\\n([\\s\\S]*?)(?=\\n\#{1,${i}} |$)`, 'g'),
      (match, sectionContent) => {
        const itemHeadingLevel = i + 1
        const headingHashes = '#'.repeat(itemHeadingLevel)
        const headingHashesShorter = Array.from({ length: itemHeadingLevel }, (_, k) => '#'.repeat(k + 1)).join('|')
        const itemBlockRegex = new RegExp(
          `${headingHashes} ([^\\n]+)\\n\\n` + // #### propName
            '```ts\\n([^\\n]+);\\n```\\n' + // code block
            '([\\s\\S]*?)' + // description block
            `(?=(?:\\n\\*\\*\\*\\n)?(?=\\n${headingHashes} )|\\n(?:${headingHashesShorter}) |$)`,
          'g',
        )

        const items = []
        let itemMatch

        while ((itemMatch = itemBlockRegex.exec(sectionContent)) !== null) {
          const [, name, typeLine, rawDescription] = itemMatch

          const lines = rawDescription
            .trim()
            .split('\n')
            .map((line) => line.trim())
            .filter((line) => !line.includes('***'))

          let deprecation = ''
          const contentLines = []

          for (let i = 0; i < lines.length; i++) {
            if (new RegExp(`^\#{${itemHeadingLevel + 1}}? Overrides`, 'i').test(lines[i])) {
              let j = i + 1
              while (j < lines.length && !lines[j].trim().startsWith('#')) j++
              i = j - 1
            } else if (new RegExp(`^\#{${itemHeadingLevel + 1}}? Deprecated`, 'i').test(lines[i])) {
              let j = i + 1
              while (j < lines.length && lines[j].trim() === '') j++
              if (j < lines.length) {
                deprecation = lines[j].trim()
                i = j
              }
            } else {
              contentLines.push(lines[i])
            }
          }

          const mainDescription = contentLines[0] || ''
          const otherLines = contentLines.slice(1)

          items.push({
            name,
            typeLine,
            mainDescription,
            otherLines,
            deprecation,
          })
        }

        if (items.length === 0) return match

        let result = `**${headerTitle}**:\n\n`

        for (const { name, typeLine, mainDescription, otherLines, deprecation } of items) {
          const typeMatch = typeLine.match(/:\s*([^;]+)/)
          if (!typeMatch) continue

          let type = typeMatch[1].trim()
          type = type.replace(/readonly\s+/, '').trim()
          type = type.replace(/([*_`\[\]()<>|])/g, '\\$1')

          if (!mainDescription && deprecation) {
            result += `- \`${name}\` _${type}_ - **_Deprecated_** - ${deprecation}\n`
            continue
          }

          result += `- \`${name}\` _${type}_`
          if (mainDescription) result += ` - ${mainDescription}`
          result += '\n'

          for (const line of otherLines) {
            result += `    ${line}\n`
          }

          if (deprecation) {
            result += `    - **_Deprecated_** - ${deprecation}\n`
          }
        }

        result = result.replace(/^\s{4}\*\*\*/gm, '***')

        return result + '\n'
      },
    )
  }

  return contents
}
