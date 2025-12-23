#!/usr/bin/env node
const fs = require('fs')
const path = require('path')

const args = process.argv.slice(2)
const distDir = args[0]
const workspaceRoot = args[1]
const sourceDir = args[2]

const esmDir = path.join(distDir, 'esm')
const cjsDir = path.join(distDir, 'cjs')

const readJson = (p) => JSON.parse(fs.readFileSync(p, 'utf8'))
const writeJson = (p, data) => fs.writeFileSync(p, JSON.stringify(data, null, 2))
const exists = (p) => fs.existsSync(p)
const rm = (p) => exists(p) && fs.rmSync(p, { recursive: true, force: true })

/* -----------------------------
 * 1. Detect project dependencies
 * ----------------------------- */
function getDependencyProjects() {
  try {
    const projectJson = readJson(path.join(sourceDir, 'project.json'))
    const deps = new Set()

    ;['build', 'build-esm', 'build-cjs'].forEach((t) => {
      projectJson.targets?.[t]?.dependsOn?.forEach((d) => {
        if (d.projects) d.projects.forEach((p) => deps.add(p))
      })
    })

    return [...deps]
  } catch {
    return []
  }
}

/* Fallback: scan esm/libs for compiled dependencies */
function getBuiltDependencyProjects() {
  const libsDir = path.join(esmDir, 'libs')
  if (!exists(libsDir)) return []
  return fs.readdirSync(libsDir).filter((d) => d !== path.basename(sourceDir))
}

/* -----------------------------
 * 2. Merge dependency versions
 * ----------------------------- */
function collectDependencies() {
  const depProjects = getDependencyProjects()
  const deps = {}

  // If CJS build already contains dependencies, use them
  const cjsPkgPath = path.join(cjsDir, 'package.json')
  if (exists(cjsPkgPath)) {
    Object.assign(deps, readJson(cjsPkgPath).dependencies || {})
  }

  const finalProjects = depProjects.length ? depProjects : getBuiltDependencyProjects()

  console.log(`📦 Dependencies from projects:`, finalProjects)

  finalProjects.forEach((project) => {
    try {
      const pkg = readJson(path.join(workspaceRoot, 'dist', 'libs', project, 'package.json'))
      deps[pkg.name] = pkg.version

      // Add transitive dependencies ONLY if none existed yet
      if (Object.keys(deps).length === 1 && pkg.dependencies) {
        Object.assign(deps, pkg.dependencies)
      }
    } catch {
      console.warn(`⚠ Missing dependency package for ${project}`)
    }
  })

  return deps
}

/* -----------------------------
 * 3. Flatten build output
 * ----------------------------- */
function flatten(buildDir) {
  const srcDir = path.join(buildDir, 'src')
  if (!exists(srcDir)) return

  console.log(`📦 Flattening ${buildDir}`)
  fs.readdirSync(srcDir, { withFileTypes: true }).forEach((entry) => {
    const from = path.join(srcDir, entry.name)
    const to = path.join(buildDir, entry.name)
    rm(to)
    fs.cpSync(from, to, { recursive: true })
  })
  rm(srcDir)
}

/* -----------------------------
 * 4. Write final package metadata
 * ----------------------------- */
const pkg = readJson(path.join(sourceDir, 'package.json'))
pkg.dependencies = collectDependencies()

flatten(esmDir)
flatten(cjsDir)

writeJson(path.join(esmDir, 'package.json'), { type: 'module' })
writeJson(path.join(cjsDir, 'package.json'), { type: 'commonjs' })

writeJson(path.join(distDir, 'package.json'), pkg)
;['README.md', 'LICENSE'].forEach((file) => {
  const src = path.join(sourceDir, file)
  if (exists(src)) fs.copyFileSync(src, path.join(distDir, file))
})

console.log('✓ Post-build script completed')
