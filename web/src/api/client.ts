import createClient from 'openapi-fetch'
import type { paths } from './schema.gen'

export const client = createClient<paths>({
  baseUrl: import.meta.env.VITE_API_BASE_URL || '/api',
})
