import express from 'express'
import { env } from './util/environment.mjs'
import { handler as ssrHandler } from '../dist/server/entry.mjs'

const app = express()
app.use(express.json())
app.use((req, res, next) => {
  res.setHeader('X-Frame-Options', 'SAMEORIGIN')
  next()
})
app.use('/', express.static('dist/client/'))
app.use((req, res, next) => {
  ssrHandler(req, res, next)
})
app.get('*', (req, res) => {
  res.sendFile('404.html', { root: 'dist/client/' })
})

app.listen(env.FUNCTIONS_PORT, () => {
  console.log(`Functions available on port ${env.FUNCTIONS_PORT}`)
})
