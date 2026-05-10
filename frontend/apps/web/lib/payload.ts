import { getPayload } from 'payload'
import config from '@payload-config'

export async function getPayloadClient() {
  if (process.env.NEXT_PHASE === 'phase-production-build') {
    throw new Error('Payload unavailable at build time')
  }
  return getPayload({ config })
}
