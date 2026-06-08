#!/usr/bin/env node
// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

const fs = require('fs')
const path = require('path')

const [distDir, workspaceRoot, sourceDir] = process.argv.slice(2)
const esmDir = path.join(distDir, 'esm')
const cjsDir = path.join(distDir, 'cjs')

const readJson = (p) => JSON.parse(fs.readFileSync(p, 'utf8'))
const writeJson = (p, data) => fs.writeFileSync(p, JSON.stringify(data, null, 2))

const generatedDeps = readJson(path.join(cjsDir, 'package.json')).dependencies ?? {}

const pkg = readJson(path.join(sourceDir, 'package.json'))
pkg.dependencies = { ...generatedDeps }
for (const name of ['api-client', 'toolbox-api-client']) {
  const distPkg = readJson(path.join(workspaceRoot, 'dist', 'libs', name, 'package.json'))
  pkg.dependencies[`@daytona/${name}`] = distPkg.version
}

for (const buildDir of [esmDir, cjsDir]) {
  const srcDir = path.join(buildDir, 'src')
  if (fs.existsSync(srcDir)) {
    for (const entry of fs.readdirSync(srcDir)) {
      fs.cpSync(path.join(srcDir, entry), path.join(buildDir, entry), { recursive: true, force: true })
    }
    fs.rmSync(srcDir, { recursive: true, force: true })
  }
}

writeJson(path.join(esmDir, 'package.json'), { type: 'module' })
writeJson(path.join(cjsDir, 'package.json'), { type: 'commonjs' })

const esmImportJs = path.join(esmDir, 'utils', 'Import.js')
if (fs.existsSync(esmImportJs)) {
  // Named `__esmRequire` (not `require`) to avoid shadowing the host CJS
  // `require` when a bundler re-compiles this ESM output to CommonJS.
  const shim =
    `import * as _m from 'module';\n` +
    `const __esmRequire = (() => {\n` +
    `  try { return _m.createRequire(import.meta.url); } catch {}\n` +
    `  try { if (typeof require !== 'undefined') return require; } catch {}\n` +
    `  return (id) => { throw new Error(\n` +
    `    'cannot require("' + id + '"): no CommonJS require available. ' +\n` +
    `    'If re-bundling @daytona/sdk to CJS, ensure createRequire or the host require is accessible.'\n` +
    `  ); };\n` +
    `})();\n`
  const original = fs.readFileSync(esmImportJs, 'utf8')
  const rewritten = original
    .replace(/require\s*\(\s*['"]\.\.\/\.\.\/package\.json['"]\s*\)/g, JSON.stringify({ name: pkg.name, version: pkg.version }))
    .replace(/\brequire\s*\(/g, '__esmRequire(')
  fs.writeFileSync(esmImportJs, shim + rewritten)
}

writeJson(path.join(distDir, 'package.json'), pkg)
for (const file of ['README.md', 'LICENSE']) {
  const src = path.join(sourceDir, file)
  if (fs.existsSync(src)) fs.copyFileSync(src, path.join(distDir, file))
}
