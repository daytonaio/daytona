import express from 'express'

import { handler as ssrHandler } from '../dist/server/entry.mjs'
import { env } from './util/environment.mjs'

const app = express()
app.use(express.json())
app.use((req, res, next) => {
  res.setHeader('X-Frame-Options', 'SAMEORIGIN')
  next()
})
app.use('/docs', express.static('dist/client/'))
app.use(ssrHandler)
app.use((req, res) => {
  res.sendFile('404.html', { root: 'dist/client/' })
})

app.listen(env.FUNCTIONS_PORT, () => {
  console.log(`Functions available on port ${env.FUNCTIONS_PORT}`)
})
