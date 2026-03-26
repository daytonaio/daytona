import { config } from 'dotenv'
import { cleanEnv, num } from 'envalid'

config()

export const env = cleanEnv(process.env, {
  FUNCTIONS_PORT: num(),
})
