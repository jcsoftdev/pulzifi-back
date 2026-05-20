import { postgresAdapter } from '@payloadcms/db-postgres'
import { lexicalEditor } from '@payloadcms/richtext-lexical'
import { s3Storage } from '@payloadcms/storage-s3'
import { buildConfig } from 'payload'
import { ALL_BLOCKS } from './features/cms/blocks/schemas'
import { seedCMSIfEmpty } from './features/cms/seed'



function requireEnv(name: string): string {
  const value = process.env[name]
  if (!value || value.trim() === '') {
    // During `next build` the Docker ARGs may not propagate into Next.js process.env.
    // Return a non-empty placeholder so the module loads; runtime pages validate via try/catch.
    if (process.env.NEXT_PHASE === 'phase-production-build') {
      return `__missing_${name}__`
    }
    throw new Error(`Missing required environment variable: ${name}`)
  }
  return value
}

const csrfOrigins = requireEnv('PAYLOAD_CSRF_ORIGINS')
  .split(',')
  .map((o) => o.trim())
  .filter(Boolean)

export default buildConfig({
  serverURL: requireEnv('NEXT_PUBLIC_SERVER_URL'),
  csrf: csrfOrigins,
  cors: csrfOrigins,
  secret: requireEnv('PAYLOAD_SECRET'),
  db: postgresAdapter({
    pool: {
      connectionString:
        process.env.NEXT_PHASE === 'phase-production-build'
          ? 'postgresql://localhost:5432/dummy'
          : `postgresql://${requireEnv('DB_USER')}:${requireEnv('DB_PASSWORD')}@${requireEnv('DB_HOST')}:${requireEnv('DB_PORT')}/${requireEnv('DB_NAME')}`,
    },
    schemaName: 'cms',
  }),
  plugins: [
    s3Storage({
      collections: {
        media: true,
      },
      bucket: requireEnv('MINIO_BUCKET'),
      config: {
        endpoint: requireEnv('MINIO_ENDPOINT'),
        credentials: {
          accessKeyId: requireEnv('MINIO_ACCESS_KEY'),
          secretAccessKey: requireEnv('MINIO_SECRET_KEY'),
        },
        region: requireEnv('MINIO_REGION'),
        forcePathStyle: true,
      },
    }),
  ],
  routes: {
    admin: '/cms-admin',
    api: '/cms-api',
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
      slug: 'block-library',
      labels: {
        singular: 'Block',
        plural: 'Block Library',
      },
      access: {
        read: () => true,
      },
      hooks: {
        beforeValidate: [
          ({ operation, data }) => {
            if (operation === 'update' && data) delete data.id
            return data
          },
        ],
      },
      admin: {
        useAsTitle: 'name',
        description:
          'Reusable blocks. Reference them on any page via the "Block Reference" block type.',
        defaultColumns: [
          'name',
          'updatedAt',
        ],
        livePreview: {
          url: ({
            data,
          }: {
            data: {
              id?: number
            }
          }) => {
            const base = process.env.NEXT_PUBLIC_SERVER_URL ?? 'http://localhost:3001'
            return `${base}/preview/block/${data.id ?? 'new'}`
          },
          breakpoints: [
            {
              label: 'Mobile',
              name: 'mobile',
              width: 390,
              height: 844,
            },
            {
              label: 'Tablet',
              name: 'tablet',
              width: 820,
              height: 1180,
            },
            {
              label: 'Desktop',
              name: 'desktop',
              width: 1440,
              height: 900,
            },
          ],
        },
      },
      fields: [
        {
          name: 'name',
          type: 'text',
          required: true,
          admin: {
            description: 'Internal label — e.g. "Hero — Main", "Pricing — EN"',
          },
        },
        {
          name: 'block',
          type: 'blocks',
          minRows: 1,
          maxRows: 1,
          blocks: ALL_BLOCKS,
          admin: {
            description: 'Add exactly one block.',
          },
        },
      ],
    },
    {
      slug: 'pages',
      versions: {
        drafts: true,
      },
      access: {
        read: () => true,
      },
      admin: {
        useAsTitle: 'title',
        livePreview: {
          url: ({
            data,
          }: {
            data: {
              slug?: string
            }
          }) => {
            const base = process.env.NEXT_PUBLIC_SERVER_URL ?? 'http://localhost:3001'
            return `${base}/preview/page/${data.slug ?? 'home'}`
          },
          breakpoints: [
            {
              label: 'Mobile',
              name: 'mobile',
              width: 390,
              height: 844,
            },
            {
              label: 'Tablet',
              name: 'tablet',
              width: 820,
              height: 1180,
            },
            {
              label: 'Desktop',
              name: 'desktop',
              width: 1440,
              height: 900,
            },
          ],
        },
      },
      fields: [
        {
          name: 'title',
          type: 'text',
          required: true,
        },
        {
          name: 'slug',
          type: 'text',
          required: true,
          unique: true,
          admin: {
            position: 'sidebar',
          },
        },
        {
          name: 'blocks',
          type: 'blocks',
          blocks: ALL_BLOCKS,
          admin: {
            description:
              'Use "Block Reference" to reuse a shared library block (no editing). Use any other block type to create page-specific content.',
          },
        },
        {
          name: 'meta',
          type: 'group',
          fields: [
            {
              name: 'title',
              type: 'text',
            },
            {
              name: 'description',
              type: 'textarea',
            },
            {
              name: 'image',
              type: 'upload',
              relationTo: 'media',
            },
          ],
        },
      ],
    },
    {
      slug: 'posts',
      admin: {
        useAsTitle: 'title',
      },
      versions: {
        drafts: true,
      },
      access: {
        read: () => true,
      },
      fields: [
        {
          name: 'title',
          type: 'text',
          required: true,
        },
        {
          name: 'slug',
          type: 'text',
          required: true,
          unique: true,
          admin: {
            position: 'sidebar',
          },
        },
        {
          name: 'heroImage',
          type: 'upload',
          relationTo: 'media',
        },
        {
          name: 'excerpt',
          type: 'textarea',
        },
        {
          name: 'content',
          type: 'richText',
          editor: lexicalEditor({}),
        },
        {
          name: 'author',
          type: 'text',
        },
        {
          name: 'category',
          type: 'select',
          options: [
            'Product',
            'Company',
            'Guide',
          ],
        },
        {
          name: 'publishedAt',
          type: 'date',
          admin: {
            position: 'sidebar',
          },
        },
        {
          name: 'meta',
          type: 'group',
          fields: [
            {
              name: 'title',
              type: 'text',
            },
            {
              name: 'description',
              type: 'textarea',
            },
            {
              name: 'image',
              type: 'upload',
              relationTo: 'media',
            },
          ],
        },
      ],
    },
    {
      slug: 'plans',
      labels: {
        singular: 'Plan',
        plural: 'Plans',
      },
      admin: {
        useAsTitle: 'name',
        defaultColumns: [
          'name',
          'price',
          'highlighted',
        ],
        description:
          'Pricing plans. Referenced from the landing pricing block and the pricing page. Edit once, both pages update.',
      },
      access: {
        read: () => true,
      },
      fields: [
        {
          name: 'name',
          type: 'text',
          required: true,
        },
        {
          name: 'price',
          type: 'text',
          required: true,
        },
        {
          name: 'period',
          type: 'text',
        },
        {
          name: 'tagline',
          type: 'textarea',
        },
        {
          name: 'features',
          type: 'array',
          fields: [
            {
              name: 'text',
              type: 'text',
            },
            {
              name: 'included',
              type: 'checkbox',
              defaultValue: true,
            },
          ],
        },
        {
          name: 'priceAnnual',
          type: 'text',
          admin: {
            description: 'Annual price (e.g. "$16"). Leave blank to hide toggle.',
          },
        },
        {
          name: 'ctaLabel',
          type: 'text',
        },
        {
          name: 'ctaHref',
          type: 'text',
        },
        {
          name: 'highlighted',
          type: 'checkbox',
        },
        {
          name: 'popularBadge',
          type: 'text',
        },
      ],
    },
  ],
  globals: [
    {
      slug: 'navbar',
      access: {
        read: () => true,
      },
      admin: {
        livePreview: {
          url: () => {
            const base = process.env.NEXT_PUBLIC_SERVER_URL ?? 'http://localhost:3001'
            return `${base}/preview/navbar`
          },
          breakpoints: [
            {
              label: 'Mobile',
              name: 'mobile',
              width: 390,
              height: 844,
            },
            {
              label: 'Tablet',
              name: 'tablet',
              width: 820,
              height: 1180,
            },
            {
              label: 'Desktop',
              name: 'desktop',
              width: 1440,
              height: 900,
            },
          ],
        },
      },
      fields: [
        {
          name: 'logo',
          type: 'upload',
          relationTo: 'media',
        },
        {
          name: 'links',
          type: 'array',
          fields: [
            {
              name: 'label',
              type: 'text',
              required: true,
            },
            {
              name: 'href',
              type: 'text',
              required: true,
            },
          ],
        },
        {
          name: 'signinLabel',
          type: 'text',
          defaultValue: 'Sign in',
        },
        {
          name: 'signinHref',
          type: 'text',
          defaultValue: '/login',
        },
        {
          name: 'primaryCtaLabel',
          type: 'text',
          defaultValue: 'Start Monitoring Free',
        },
        {
          name: 'primaryCtaHref',
          type: 'text',
          defaultValue: '/register',
        },
      ],
    },
    {
      slug: 'footer',
      access: {
        read: () => true,
      },
      admin: {
        livePreview: {
          url: () => {
            const base = process.env.NEXT_PUBLIC_SERVER_URL ?? 'http://localhost:3001'
            return `${base}/preview/footer`
          },
          breakpoints: [
            {
              label: 'Mobile',
              name: 'mobile',
              width: 390,
              height: 844,
            },
            {
              label: 'Tablet',
              name: 'tablet',
              width: 820,
              height: 1180,
            },
            {
              label: 'Desktop',
              name: 'desktop',
              width: 1440,
              height: 900,
            },
          ],
        },
      },
      fields: [
        {
          name: 'logo',
          type: 'upload',
          relationTo: 'media',
        },
        {
          name: 'tagline',
          type: 'text',
        },
        {
          name: 'groups',
          type: 'array',
          fields: [
            {
              name: 'heading',
              type: 'text',
              required: true,
            },
            {
              name: 'links',
              type: 'array',
              fields: [
                {
                  name: 'label',
                  type: 'text',
                  required: true,
                },
                {
                  name: 'href',
                  type: 'text',
                  required: true,
                },
              ],
            },
          ],
        },
        {
          name: 'socialLinks',
          type: 'array',
          fields: [
            {
              name: 'platform',
              type: 'select',
              options: [
                'twitter',
                'linkedin',
                'github',
                'youtube',
                'other',
              ],
              required: true,
            },
            {
              name: 'href',
              type: 'text',
              required: true,
            },
          ],
        },
        {
          name: 'copyrightText',
          type: 'text',
        },
      ],
    },
    {
      slug: 'theme',
      label: 'Theme (Landing Colors)',
      access: {
        read: () => true,
      },
      admin: {
        description:
          'Edit any color value. Leave empty to keep built-in defaults. Affects the landing page only.',
        livePreview: {
          url: () => {
            const base = process.env.NEXT_PUBLIC_SERVER_URL ?? 'http://localhost:3001'
            return `${base}/preview/theme`
          },
          breakpoints: [
            {
              label: 'Mobile',
              name: 'mobile',
              width: 390,
              height: 844,
            },
            {
              label: 'Tablet',
              name: 'tablet',
              width: 820,
              height: 1180,
            },
            {
              label: 'Desktop',
              name: 'desktop',
              width: 1440,
              height: 900,
            },
          ],
        },
      },
      fields: [
        {
          type: 'collapsible',
          label: 'Surfaces',
          fields: [
            {
              name: 'pageBg',
              type: 'text',
              label: 'Page background',
              admin: {
                description: 'Default #f3f3f3',
                components: {
                  Field: '@/payload-fields/color-field/component#ColorField',
                },
              },
            },
            {
              name: 'pageBgAlt',
              type: 'text',
              label: 'Alt background (logos / testimonials / faq)',
              admin: {
                description: 'Default #e7e7eb',
                components: {
                  Field: '@/payload-fields/color-field/component#ColorField',
                },
              },
            },
            {
              name: 'cardBg',
              type: 'text',
              label: 'Card background',
              admin: {
                description: 'Default #ffffff',
                components: {
                  Field: '@/payload-fields/color-field/component#ColorField',
                },
              },
            },
            {
              name: 'darkSurface',
              type: 'text',
              label: 'Dark surface (problem section, CTA)',
              admin: {
                description: 'Default #29144c',
                components: {
                  Field: '@/payload-fields/color-field/component#ColorField',
                },
              },
            },
          ],
        },
        {
          type: 'collapsible',
          label: 'Text',
          fields: [
            {
              name: 'inkPrimary',
              type: 'text',
              label: 'Primary text',
              admin: {
                description: 'Default #131313',
                components: {
                  Field: '@/payload-fields/color-field/component#ColorField',
                },
              },
            },
            {
              name: 'inkSecondary',
              type: 'text',
              label: 'Secondary text',
              admin: {
                description: 'Default #444141',
                components: {
                  Field: '@/payload-fields/color-field/component#ColorField',
                },
              },
            },
          ],
        },
        {
          type: 'collapsible',
          label: 'Accents',
          fields: [
            {
              name: 'accentPrimary',
              type: 'text',
              label: 'Primary accent (buttons / highlights)',
              admin: {
                description: 'Default #6A35E0',
                components: {
                  Field: '@/payload-fields/color-field/component#ColorField',
                },
              },
            },
            {
              name: 'accentMuted',
              type: 'text',
              label: 'Primary muted',
              admin: {
                description: 'Default #8b5cf6',
                components: {
                  Field: '@/payload-fields/color-field/component#ColorField',
                },
              },
            },
            {
              name: 'accentGold',
              type: 'text',
              label: 'Gold accent (pricing highlight, stars)',
              admin: {
                description: 'Default #f59e0b',
                components: {
                  Field: '@/payload-fields/color-field/component#ColorField',
                },
              },
            },
            {
              name: 'accentTeal',
              type: 'text',
              label: 'Teal accent (AI insights)',
              admin: {
                description: 'Default #14b8a6',
                components: {
                  Field: '@/payload-fields/color-field/component#ColorField',
                },
              },
            },
          ],
        },
        {
          type: 'collapsible',
          label: 'Borders',
          fields: [
            {
              name: 'border',
              type: 'text',
              label: 'Subtle border',
              admin: {
                description: 'Default rgba(0,0,0,0.08)',
                components: {
                  Field: '@/payload-fields/color-field/component#ColorField',
                },
              },
            },
            {
              name: 'borderStrong',
              type: 'text',
              label: 'Strong border',
              admin: {
                description: 'Default rgba(0,0,0,0.16)',
                components: {
                  Field: '@/payload-fields/color-field/component#ColorField',
                },
              },
            },
          ],
        },
      ],
    },
    {
      slug: 'login-page',
      label: 'Login Page',
      access: {
        read: () => true,
      },
      fields: [
        {
          name: 'headline',
          type: 'text',
          defaultValue: 'Know before your competitors move.',
        },
        {
          name: 'subheadline',
          type: 'text',
          defaultValue: 'Join 2,500+ teams who react first.',
        },
        {
          name: 'bullets',
          type: 'array',
          fields: [
            {
              name: 'text',
              type: 'text',
              required: true,
            },
          ],
          defaultValue: [
            {
              text: 'Monitor any competitor page — no code required',
            },
            {
              text: 'AI explains what changed and what to do next',
            },
            {
              text: 'Get alerts in Slack, Email, Teams or WhatsApp',
            },
            {
              text: '14-day free trial · No credit card required',
            },
          ],
        },
        {
          name: 'welcomeHeadline',
          type: 'text',
          defaultValue: 'Welcome back',
        },
        {
          name: 'welcomeSubline',
          type: 'text',
          defaultValue: 'Sign in to your account',
        },
      ],
    },
    {
      slug: 'contact-page',
      label: 'Contact Page',
      access: {
        read: () => true,
      },
      fields: [
        {
          name: 'headline',
          type: 'text',
          defaultValue: 'Get in touch',
        },
        {
          name: 'subheadline',
          type: 'text',
          defaultValue: "We're here to help. Our team responds within one business day.",
        },
        {
          name: 'email',
          type: 'text',
          defaultValue: 'support@pulzifi.com',
        },
        {
          name: 'address',
          type: 'text',
          defaultValue: 'Boise, ID',
        },
      ],
    },
  ],
  onInit: async (payload) => {
    try {
      const result = await seedCMSIfEmpty(payload)
      if (result.seeded) payload.logger.info('[cms] auto-seeded landing/navbar/footer globals')
    } catch (err) {
      payload.logger.error(
        {
          err,
        },
        '[cms] auto-seed failed'
      )
    }
  },
})
