import config from '@payload-config'
import { getPayload } from 'payload'

export async function getPayloadClient() {
  if (process.env.NEXT_PHASE === 'phase-production-build') {
    throw new Error('Payload unavailable at build time')
  }
  return getPayload({
    config,
  })
}
