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
  const shim = `import * as _m from 'module';\nconst require = (() => { try { return _m.createRequire(import.meta.url); } catch { return (id) => { throw new Error('require("' + id + '") unavailable'); }; } })();\n`
  fs.writeFileSync(esmImportJs, shim + fs.readFileSync(esmImportJs, 'utf8'))
}

writeJson(path.join(distDir, 'package.json'), pkg)
for (const file of ['README.md', 'LICENSE']) {
  const src = path.join(sourceDir, file)
  if (fs.existsSync(src)) fs.copyFileSync(src, path.join(distDir, file))
}
