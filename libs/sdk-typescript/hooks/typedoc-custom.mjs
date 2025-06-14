// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

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
    transformInheritedSections,
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

function transformInheritedSections(contents) {
  // Transform "##### Inherited from" sections into inline inheritance notes
  // Handle both simple and complex patterns (with Memberof sections and code blocks)

  const hasInheritedFrom = contents.includes('##### Inherited from')

  if (!hasInheritedFrom) {
    return contents
  }

  // Use line-by-line processing for maximum reliability
  const lines = contents.split('\n')
  let modified = false

  for (let i = 0; i < lines.length; i++) {
    // Look for "##### Inherited from" lines
    if (lines[i].trim() === '##### Inherited from') {
      // Look backwards for property definition (list item or heading)
      let propertyLineIndex = -1
      let propertyLine = ''

      // Search backwards up to 20 lines
      for (let j = i - 1; j >= Math.max(0, i - 20); j--) {
        const line = lines[j].trim()
        // Property heading format: #### propertyName
        if (line.match(/^#### [^#]+$/)) {
          propertyLineIndex = j
          propertyLine = lines[j]
          break
        }
        // Property list item format: - `property` _type_ - description
        if (line.match(/^- `[^`]+` _[^_]+_/)) {
          propertyLineIndex = j
          propertyLine = lines[j]
          break
        }
      }

      // Look forwards for inheritance info
      let inheritanceInfo = ''
      let memberofInfo = ''
      let endIndex = i

      // Check for code block format
      for (let k = i + 1; k < Math.min(lines.length, i + 5); k++) {
        if (lines[k].trim() === '```ts' && k + 2 < lines.length) {
          const nextLine = lines[k + 1].trim()
          if (nextLine && lines[k + 2].trim() === '```') {
            inheritanceInfo = nextLine
            endIndex = k + 2
            break
          }
        }
        // Check for simple format: `Class`.`property`
        if (lines[k].trim().match(/^`[^`]+`\.`[^`]+`$/)) {
          inheritanceInfo = lines[k].trim().replace(/`/g, '')
          endIndex = k
          break
        }
      }

      // Look backwards for Memberof info
      for (let m = i - 1; m >= Math.max(0, i - 5); m--) {
        if (lines[m].trim() === '##### Memberof') {
          // Look for the memberof value in the next few lines
          for (let n = m + 1; n < Math.min(lines.length, m + 5); n++) {
            if (lines[n].trim() === '```ts' && n + 2 < lines.length) {
              const memberofLine = lines[n + 1].trim()
              if (memberofLine && lines[n + 2].trim() === '```') {
                memberofInfo = memberofLine
                break
              }
            }
            // Check for simple format: `ClassName`
            if (lines[n].trim().match(/^`[^`]+`$/)) {
              memberofInfo = lines[n].trim().replace(/`/g, '')
              break
            }
            // Check for plain text format: ClassName
            if (lines[n].trim().length > 0 && !lines[n].trim().startsWith('#') && !lines[n].trim().startsWith('```')) {
              memberofInfo = lines[n].trim()
              break
            }
          }
          break
        }
      }

      // If we found both property and inheritance info, transform it
      if (propertyLineIndex >= 0 && inheritanceInfo) {
        // Use memberof info if available, otherwise use inheritance info
        let finalInheritanceInfo
        if (memberofInfo && inheritanceInfo.includes('.')) {
          // Extract property name from inheritance info and combine with memberof class
          const propertyName = inheritanceInfo.split('.').pop()
          finalInheritanceInfo = `${memberofInfo}.${propertyName}`
        } else {
          finalInheritanceInfo = inheritanceInfo
        }

        // Add inheritance info to property line
        if (propertyLine.startsWith('#### ')) {
          // For headings, add after the heading
          lines[propertyLineIndex] = propertyLine + `\n\n_Inherited from_: \`${finalInheritanceInfo}\``
        } else if (propertyLine.startsWith('- `')) {
          // For list items, add as a sub-item
          lines[propertyLineIndex] = propertyLine + `\n    - _Inherited from_: \`${finalInheritanceInfo}\``
        }

        // Remove the inheritance section
        // Find the start of the section (might include Memberof)
        let startIndex = i
        for (let m = i - 1; m >= Math.max(0, i - 5); m--) {
          if (lines[m].trim() === '##### Memberof') {
            startIndex = m
            break
          }
          // Also check for empty lines to find the start of the inheritance block
          if (lines[m].trim() === '' && m > 0 && lines[m - 1].trim() !== '') {
            startIndex = m
            break
          }
        }

        // Remove all lines from startIndex to endIndex (inclusive)
        for (let r = endIndex; r >= startIndex; r--) {
          lines.splice(r, 1)
        }

        // Adjust our loop index since we removed lines
        i = propertyLineIndex
        modified = true
      }
    }
  }

  // Remove any remaining standalone "##### Memberof" sections
  if (modified) {
    const finalLines = lines.join('\n').split('\n')
    for (let i = finalLines.length - 1; i >= 0; i--) {
      if (finalLines[i].trim() === '##### Memberof') {
        // Remove the Memberof line and any following content until next heading or empty line
        let endMemberof = i
        for (let j = i + 1; j < finalLines.length; j++) {
          if (finalLines[j].trim() === '' || finalLines[j].match(/^#{1,6} /)) {
            break
          }
          endMemberof = j
        }
        finalLines.splice(i, endMemberof - i + 1)
      }
    }

    // Also remove any standalone class name lines that might be leftover
    const cleanedContent = finalLines
      .join('\n')
      .replace(/\n\s*\n\s*([A-Z][a-zA-Z]*)\s*\n\s*\n/g, '\n\n') // Remove standalone class names between empty lines
      .replace(/\n\s*([A-Z][a-zA-Z]*)\s*\n(?=\s*-|\s*\*\*)/g, '\n') // Remove class names before property lists or sections

    return cleanedContent
  }

  return modified ? lines.join('\n') : contents
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
            '(?:_Inherited from_: `([^`]+)`\\n\\n)?' + // optional inheritance info
            '(?:([A-Za-z]+)\\n)?' + // optional leftover memberof text (like "Workspace")
            '(?:\\n)?' + // optional empty line after leftover text
            '```ts\\n([^\\n]+);\\n```\\n' + // code block
            '([\\s\\S]*?)' + // description block (may include Index Signature)
            `(?=(?:\\n\\*\\*\\*\\n)?(?=\\n${headingHashes} )|\\n(?:${headingHashesShorter}) |$)`,
          'g',
        )

        const items = []
        let itemMatch

        while ((itemMatch = itemBlockRegex.exec(sectionContent)) !== null) {
          const [, name, inheritanceInfo, leftoverText, typeLine, rawDescription] = itemMatch

          const lines = rawDescription
            .trim()
            .split('\n')
            .map((line) => line.trim())
            .filter((line) => !line.includes('***'))

          let deprecation = ''
          const contentLines = []
          const indexSignatureLines = []
          let inIndexSignature = false

          for (let i = 0; i < lines.length; i++) {
            if (new RegExp(`^\#{${itemHeadingLevel + 1}}? Index Signature`, 'i').test(lines[i])) {
              inIndexSignature = true
              indexSignatureLines.push(lines[i])
              continue
            }

            if (inIndexSignature) {
              indexSignatureLines.push(lines[i])
              continue
            }

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

          // Extract index signature type if present
          let indexSignatureType = null
          for (const line of indexSignatureLines) {
            if (line.includes('[') && line.includes(']:')) {
              indexSignatureType = line.trim()
              break
            }
          }

          items.push({
            name,
            typeLine,
            mainDescription,
            otherLines,
            deprecation,
            inheritanceInfo,
            indexSignatureType,
          })
        }

        if (items.length === 0) return match

        let result = `**${headerTitle}**:\n\n`

        for (const {
          name,
          typeLine,
          mainDescription,
          otherLines,
          deprecation,
          inheritanceInfo,
          indexSignatureType,
        } of items) {
          const typeMatch = typeLine.match(/:\s*([^;]+)/)
          if (!typeMatch) continue

          let type = typeMatch[1].trim()
          type = type.replace(/readonly\s+/, '').trim()

          // Use index signature type if available, otherwise use the original type
          if (indexSignatureType) {
            type = indexSignatureType
          }

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

          if (inheritanceInfo) {
            result += `    - _Inherited from_: \`${inheritanceInfo}\`\n`
          }
        }

        result = result.replace(/^\s{4}\*\*\*/gm, '***')

        return result + '\n'
      },
    )
  }

  return contents
}
