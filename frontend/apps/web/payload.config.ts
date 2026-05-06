import { postgresAdapter } from '@payloadcms/db-postgres'
import { s3Storage } from '@payloadcms/storage-s3'
import { buildConfig } from 'payload'

export default buildConfig({
  secret: process.env.PAYLOAD_SECRET ?? 'dev-secret-change-in-prod',
  db: postgresAdapter({
    pool: {
      connectionString: `postgresql://${process.env.DB_USER}:${process.env.DB_PASSWORD}@${process.env.DB_HOST}:${process.env.DB_PORT}/${process.env.DB_NAME}`,
    },
  }),
  plugins: [
    s3Storage({
      collections: { media: true },
      bucket: process.env.MINIO_BUCKET ?? 'payload-media',
      config: {
        endpoint: process.env.MINIO_ENDPOINT,
        credentials: {
          accessKeyId: process.env.MINIO_ACCESS_KEY ?? '',
          secretAccessKey: process.env.MINIO_SECRET_KEY ?? '',
        },
        region: 'us-east-1',
        forcePathStyle: true,
      },
    }),
  ],
  routes: {
    admin: '/cms-admin',
  },
  admin: {
    user: 'users',
  },
  collections: [
    {
      slug: 'media',
      upload: true,
      access: {
        read: () => true,
      },
      fields: [
        {
          name: 'alt',
          type: 'text',
        },
      ],
    },
    {
      slug: 'users',
      auth: true,
      fields: [],
    },
  ],
  globals: [],
})
