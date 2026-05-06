import { postgresAdapter } from '@payloadcms/db-postgres'
import { buildConfig } from 'payload'

export default buildConfig({
  secret: process.env.PAYLOAD_SECRET ?? 'dev-secret-change-in-prod',
  db: postgresAdapter({
    pool: {
      connectionString: `postgresql://${process.env.DB_USER}:${process.env.DB_PASSWORD}@${process.env.DB_HOST}:${process.env.DB_PORT}/${process.env.DB_NAME}`,
    },
  }),
  routes: {
    admin: '/cms-admin',
  },
  collections: [],
  globals: [],
})
