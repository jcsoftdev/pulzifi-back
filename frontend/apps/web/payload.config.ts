import { postgresAdapter } from '@payloadcms/db-postgres'
import { lexicalEditor } from '@payloadcms/richtext-lexical'
import { s3Storage } from '@payloadcms/storage-s3'
import { buildConfig } from 'payload'
import { ALL_BLOCKS } from '@/features/cms/blocks/schemas'

export default buildConfig({
  secret: process.env.PAYLOAD_SECRET ?? 'dev-secret-change-in-prod',
  db: postgresAdapter({
    pool: {
      connectionString: `postgresql://${process.env.DB_USER}:${process.env.DB_PASSWORD}@${process.env.DB_HOST}:${process.env.DB_PORT}/${process.env.DB_NAME}`,
    },
    schemaName: 'cms',
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
    {
      slug: 'pages',
      admin: { useAsTitle: 'title' },
      versions: { drafts: true },
      endpoints: false,
      access: { read: () => true },
      fields: [
        { name: 'title', type: 'text', required: true },
        {
          name: 'slug',
          type: 'text',
          required: true,
          unique: true,
          admin: { position: 'sidebar' },
        },
        {
          name: 'blocks',
          type: 'blocks',
          blocks: ALL_BLOCKS,
        },
        {
          name: 'meta',
          type: 'group',
          fields: [
            { name: 'title', type: 'text' },
            { name: 'description', type: 'textarea' },
            { name: 'image', type: 'upload', relationTo: 'media' },
          ],
        },
      ],
    },
    {
      slug: 'posts',
      admin: { useAsTitle: 'title' },
      versions: { drafts: true },
      endpoints: false,
      access: { read: () => true },
      fields: [
        { name: 'title', type: 'text', required: true },
        {
          name: 'slug',
          type: 'text',
          required: true,
          unique: true,
          admin: { position: 'sidebar' },
        },
        { name: 'heroImage', type: 'upload', relationTo: 'media' },
        { name: 'excerpt', type: 'textarea' },
        { name: 'content', type: 'richText', editor: lexicalEditor({}) },
        { name: 'author', type: 'text' },
        {
          name: 'category',
          type: 'select',
          options: ['Product', 'Company', 'Guide'],
        },
        { name: 'publishedAt', type: 'date', admin: { position: 'sidebar' } },
        {
          name: 'meta',
          type: 'group',
          fields: [
            { name: 'title', type: 'text' },
            { name: 'description', type: 'textarea' },
            { name: 'image', type: 'upload', relationTo: 'media' },
          ],
        },
      ],
    },
  ],
  globals: [
    {
      slug: 'landing',
      access: { read: () => true },
      fields: [
        {
          name: 'blocks',
          type: 'blocks',
          blocks: ALL_BLOCKS,
        },
      ],
    },
    {
      slug: 'navbar',
      access: { read: () => true },
      fields: [
        {
          name: 'links',
          type: 'array',
          fields: [
            { name: 'label', type: 'text', required: true },
            { name: 'href', type: 'text', required: true },
          ],
        },
      ],
    },
    {
      slug: 'footer',
      access: { read: () => true },
      fields: [
        {
          name: 'groups',
          type: 'array',
          fields: [
            { name: 'heading', type: 'text', required: true },
            {
              name: 'links',
              type: 'array',
              fields: [
                { name: 'label', type: 'text', required: true },
                { name: 'href', type: 'text', required: true },
              ],
            },
          ],
        },
        { name: 'copyrightText', type: 'text' },
      ],
    },
  ],
})
