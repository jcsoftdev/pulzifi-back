import { lexicalEditor } from '@payloadcms/richtext-lexical'
import type { Block } from 'payload'

export const HeroBlock: Block = {
  slug: 'hero',
  labels: {
    singular: 'Hero',
    plural: 'Heroes',
  },
  fields: [
    {
      name: 'eyebrowBadge',
      type: 'text',
    },
    {
      name: 'eyebrowText',
      type: 'text',
    },
    {
      name: 'headline',
      type: 'text',
      required: true,
    },
    {
      name: 'headlineHighlight',
      type: 'text',
    },
    {
      name: 'subheadline',
      type: 'textarea',
    },
    {
      name: 'primaryCtaLabel',
      type: 'text',
    },
    {
      name: 'primaryCtaHref',
      type: 'text',
    },
    {
      name: 'secondaryCtaLabel',
      type: 'text',
    },
    {
      name: 'secondaryCtaHref',
      type: 'text',
    },
    {
      name: 'trustLine',
      type: 'text',
    },
    {
      name: 'image',
      type: 'upload',
      relationTo: 'media',
    },
    {
      name: 'dashboardAlerts',
      type: 'array',
      labels: {
        singular: 'Alert',
        plural: 'Alerts',
      },
      fields: [
        {
          name: 'tone',
          type: 'select',
          options: [
            'signal',
            'amber',
            'teal',
            'ink',
          ],
          defaultValue: 'signal',
        },
        {
          name: 'icon',
          type: 'text',
        },
        {
          name: 'site',
          type: 'text',
          required: true,
        },
        {
          name: 'title',
          type: 'text',
          required: true,
        },
        {
          name: 'detail',
          type: 'text',
        },
        {
          name: 'time',
          type: 'text',
        },
      ],
    },
    {
      name: 'aiInsightTitle',
      type: 'text',
    },
    {
      name: 'aiInsightBody',
      type: 'textarea',
    },
    {
      name: 'kpis',
      type: 'array',
      maxRows: 6,
      fields: [
        {
          name: 'label',
          type: 'text',
          required: true,
        },
        {
          name: 'value',
          type: 'text',
          required: true,
        },
        {
          name: 'delta',
          type: 'text',
        },
        {
          name: 'deltaDirection',
          type: 'select',
          options: [
            'up',
            'down',
          ],
          defaultValue: 'up',
        },
      ],
    },
  ],
}

export const LogosBlock: Block = {
  slug: 'logos',
  labels: {
    singular: 'Logos Bar',
    plural: 'Logos Bars',
  },
  fields: [
    {
      name: 'label',
      type: 'text',
    },
    {
      name: 'items',
      type: 'array',
      fields: [
        {
          name: 'text',
          type: 'text',
          required: true,
        },
      ],
    },
  ],
}

export const ProblemBlock: Block = {
  slug: 'problem',
  labels: {
    singular: 'Problem',
    plural: 'Problems',
  },
  fields: [
    {
      name: 'eyebrow',
      type: 'text',
    },
    {
      name: 'headline',
      type: 'text',
      required: true,
    },
    {
      name: 'headlineHighlight',
      type: 'text',
    },
    {
      name: 'cards',
      type: 'array',
      fields: [
        {
          name: 'metric',
          type: 'text',
          required: true,
        },
        {
          name: 'label',
          type: 'text',
          required: true,
        },
        {
          name: 'description',
          type: 'textarea',
          required: true,
        },
      ],
    },
  ],
}

export const StatsBlock: Block = {
  slug: 'stats',
  labels: {
    singular: 'Stats',
    plural: 'Stats',
  },
  fields: [
    {
      name: 'items',
      type: 'array',
      fields: [
        {
          name: 'value',
          type: 'text',
          required: true,
        },
        {
          name: 'label',
          type: 'text',
          required: true,
        },
      ],
    },
  ],
}

export const HowItWorksBlock: Block = {
  slug: 'how-it-works',
  labels: {
    singular: 'How It Works',
    plural: 'How It Works',
  },
  fields: [
    {
      name: 'eyebrow',
      type: 'text',
    },
    {
      name: 'headline',
      type: 'text',
    },
    {
      name: 'headlineHighlight',
      type: 'text',
    },
    {
      name: 'subheadline',
      type: 'textarea',
    },
    {
      name: 'steps',
      type: 'array',
      fields: [
        {
          name: 'step',
          type: 'number',
          required: true,
        },
        {
          name: 'icon',
          type: 'text',
        },
        {
          name: 'title',
          type: 'text',
          required: true,
        },
        {
          name: 'description',
          type: 'textarea',
          required: true,
        },
        {
          name: 'mockType',
          type: 'select',
          options: [
            'url',
            'insight',
            'alerts',
          ],
          defaultValue: 'url',
          required: true,
        },
        {
          name: 'mockText',
          type: 'text',
        },
      ],
    },
  ],
}

export const FeaturesBlock: Block = {
  slug: 'features',
  labels: {
    singular: 'Features',
    plural: 'Features',
  },
  fields: [
    {
      name: 'eyebrow',
      type: 'text',
    },
    {
      name: 'headline',
      type: 'text',
    },
    {
      name: 'headlineHighlight',
      type: 'text',
    },
    {
      name: 'intro',
      type: 'textarea',
    },
    {
      name: 'bullets',
      type: 'array',
      fields: [
        {
          name: 'title',
          type: 'text',
          required: true,
        },
        {
          name: 'description',
          type: 'text',
          required: true,
        },
      ],
    },
    {
      name: 'demoTitle',
      type: 'text',
    },
    {
      name: 'demoBadge',
      type: 'text',
    },
    {
      name: 'demoSite',
      type: 'text',
    },
    {
      name: 'demoChange',
      type: 'text',
    },
    {
      name: 'demoAnalysis',
      type: 'textarea',
    },
    {
      name: 'demoActions',
      type: 'array',
      fields: [
        {
          name: 'label',
          type: 'text',
          required: true,
        },
      ],
    },
    {
      name: 'cards',
      type: 'array',
      admin: {
        description: 'Legacy card layout — leave empty when using bullets/demo',
      },
      fields: [
        {
          name: 'title',
          type: 'text',
          required: true,
        },
        {
          name: 'description',
          type: 'text',
          required: true,
        },
        {
          name: 'image',
          type: 'upload',
          relationTo: 'media',
        },
      ],
    },
  ],
}

export const InsightsBlock: Block = {
  slug: 'insights',
  labels: {
    singular: 'Insights',
    plural: 'Insights',
  },
  fields: [],
}

export const AiIntelligenceBlock: Block = {
  slug: 'ai-intelligence',
  labels: {
    singular: 'AI Intelligence',
    plural: 'AI Intelligence',
  },
  fields: [
    {
      name: 'eyebrow',
      type: 'text',
    },
    {
      name: 'headline',
      type: 'text',
      required: true,
    },
    {
      name: 'headlineHighlight',
      type: 'text',
    },
    {
      name: 'subheadline',
      type: 'textarea',
    },
    {
      name: 'tabs',
      type: 'array',
      minRows: 2,
      maxRows: 4,
      fields: [
        {
          name: 'label',
          type: 'text',
          required: true,
        },
        {
          name: 'items',
          type: 'array',
          minRows: 1,
          maxRows: 4,
          fields: [
            {
              name: 'title',
              type: 'text',
              required: true,
            },
            {
              name: 'body',
              type: 'textarea',
            },
            {
              name: 'image',
              type: 'text',
              admin: {
                description: 'Image URL or /images/landing/... path',
              },
            },
          ],
        },
      ],
    },
  ],
}

export const IndustriesBlock: Block = {
  slug: 'industries',
  labels: {
    singular: 'Industries / Use Cases',
    plural: 'Industries / Use Cases',
  },
  fields: [
    {
      name: 'compactMode',
      type: 'checkbox',
      defaultValue: true,
    },
    {
      name: 'eyebrow',
      type: 'text',
    },
    {
      name: 'headline',
      type: 'text',
    },
    {
      name: 'headlineHighlight',
      type: 'text',
    },
    {
      name: 'subheadline',
      type: 'textarea',
    },
    {
      name: 'items',
      type: 'array',
      fields: [
        {
          name: 'icon',
          type: 'text',
        },
        {
          name: 'title',
          type: 'text',
          required: true,
        },
        {
          name: 'description',
          type: 'textarea',
          required: true,
        },
        {
          name: 'realWin',
          type: 'textarea',
        },
      ],
    },
  ],
}

export const ComparisonBlock: Block = {
  slug: 'comparison',
  labels: {
    singular: 'Comparison Table',
    plural: 'Comparison Tables',
  },
  fields: [
    {
      name: 'eyebrow',
      type: 'text',
    },
    {
      name: 'headline',
      type: 'text',
    },
    {
      name: 'headlineHighlight',
      type: 'text',
    },
    {
      name: 'columns',
      type: 'array',
      minRows: 2,
      maxRows: 5,
      fields: [
        {
          name: 'name',
          type: 'text',
          required: true,
        },
        {
          name: 'isUs',
          type: 'checkbox',
        },
      ],
    },
    {
      name: 'rows',
      type: 'array',
      fields: [
        {
          name: 'feature',
          type: 'text',
          required: true,
        },
        {
          name: 'cells',
          type: 'array',
          fields: [
            {
              name: 'state',
              type: 'select',
              options: [
                'yes',
                'no',
                'partial',
              ],
              defaultValue: 'yes',
            },
            {
              name: 'note',
              type: 'text',
            },
          ],
        },
      ],
    },
  ],
}

export const PricingBlock: Block = {
  slug: 'pricing',
  labels: {
    singular: 'Pricing',
    plural: 'Pricing',
  },
  fields: [
    {
      name: 'eyebrow',
      type: 'text',
    },
    {
      name: 'headline',
      type: 'text',
    },
    {
      name: 'headlineHighlight',
      type: 'text',
    },
    {
      name: 'subheadline',
      type: 'textarea',
    },
    {
      name: 'plans',
      type: 'relationship',
      relationTo: 'plans',
      hasMany: true,
      admin: {
        description:
          'Select plans to display. Drag to reorder. Edit a plan record to update it everywhere it appears.',
      },
    },
    {
      name: 'guaranteeNote',
      type: 'text',
    },
    {
      type: 'group',
      name: 'billing',
      label: 'Billing Toggle',
      fields: [
        {
          name: 'monthlyLabel',
          type: 'text',
          admin: {
            description: 'Default: "Monthly"',
          },
        },
        {
          name: 'annualLabel',
          type: 'text',
          admin: {
            description: 'Default: "Annual"',
          },
        },
        {
          name: 'annualBadge',
          type: 'text',
          admin: {
            description: 'Default: "2 months free"',
          },
        },
        {
          name: 'annualNote',
          type: 'text',
          admin: {
            description: 'Default: "Billed annually · 2 months free"',
          },
        },
      ],
    },
    {
      name: 'comparePlansHeadline',
      type: 'text',
      admin: {
        description: 'Default: "Compare plans"',
      },
    },
    {
      name: 'featuresLabel',
      type: 'text',
      admin: {
        description: 'Default: "Features:"',
      },
    },
  ],
}

export const TestimonialsBlock: Block = {
  slug: 'testimonials',
  labels: {
    singular: 'Testimonials',
    plural: 'Testimonials',
  },
  fields: [
    {
      name: 'eyebrow',
      type: 'text',
    },
    {
      name: 'headline',
      type: 'text',
    },
    {
      name: 'items',
      type: 'array',
      fields: [
        {
          name: 'quote',
          type: 'textarea',
          required: true,
        },
        {
          name: 'author',
          type: 'text',
          required: true,
        },
        {
          name: 'role',
          type: 'text',
        },
        {
          name: 'avatar',
          type: 'upload',
          relationTo: 'media',
        },
      ],
    },
  ],
}

export const FaqBlock: Block = {
  slug: 'faq',
  labels: {
    singular: 'FAQ',
    plural: 'FAQs',
  },
  fields: [
    {
      name: 'eyebrow',
      type: 'text',
    },
    {
      name: 'headline',
      type: 'text',
    },
    {
      name: 'subheadline',
      type: 'text',
    },
    {
      name: 'items',
      type: 'array',
      fields: [
        {
          name: 'question',
          type: 'text',
          required: true,
        },
        {
          name: 'answer',
          type: 'textarea',
          required: true,
        },
      ],
    },
  ],
}

export const RichTextBlock: Block = {
  slug: 'rich-text',
  labels: {
    singular: 'Rich Text',
    plural: 'Rich Text',
  },
  fields: [
    {
      name: 'content',
      type: 'richText',
      editor: lexicalEditor({}),
      required: true,
    },
  ],
}

export const CtaBlock: Block = {
  slug: 'cta',
  labels: {
    singular: 'CTA',
    plural: 'CTAs',
  },
  fields: [
    {
      name: 'eyebrow',
      type: 'text',
    },
    {
      name: 'headline',
      type: 'text',
      required: true,
    },
    {
      name: 'headlineHighlight',
      type: 'text',
    },
    {
      name: 'subtext',
      type: 'textarea',
    },
    {
      name: 'primaryLabel',
      type: 'text',
    },
    {
      name: 'primaryHref',
      type: 'text',
    },
    {
      name: 'secondaryLabel',
      type: 'text',
    },
    {
      name: 'secondaryHref',
      type: 'text',
    },
    {
      name: 'riskNote',
      type: 'text',
    },
    {
      name: 'variant',
      type: 'select',
      options: [
        'primary',
        'secondary',
      ],
      defaultValue: 'primary',
    },
  ],
}

export const ImageBlock: Block = {
  slug: 'image',
  labels: {
    singular: 'Image',
    plural: 'Images',
  },
  fields: [
    {
      name: 'image',
      type: 'upload',
      relationTo: 'media',
      required: true,
    },
    {
      name: 'caption',
      type: 'text',
    },
    {
      name: 'size',
      type: 'select',
      options: [
        'full',
        'contained',
      ],
      defaultValue: 'contained',
    },
  ],
}

export const LoginFormBlock: Block = {
  slug: 'login-form',
  labels: {
    singular: 'Login Form',
    plural: 'Login Forms',
  },
  fields: [
    {
      name: 'headline',
      type: 'text',
      admin: {
        description: 'Default: "Welcome back"',
      },
    },
    {
      name: 'subheadline',
      type: 'text',
      admin: {
        description: 'Default: "Enter your credentials to continue"',
      },
    },
  ],
}

export const RegisterFormBlock: Block = {
  slug: 'register-form',
  labels: {
    singular: 'Register Form',
    plural: 'Register Forms',
  },
  fields: [
    {
      name: 'headline',
      type: 'text',
      admin: {
        description: 'Default: "Create your account"',
      },
    },
    {
      name: 'subheadline',
      type: 'text',
      admin: {
        description: 'Default: "No credit card required"',
      },
    },
    {
      name: 'trialBadge',
      type: 'text',
      admin: {
        description: 'Default: "Free 14-day trial"',
      },
    },
  ],
}

export const BlockRefBlock: Block = {
  slug: 'block-ref',
  labels: {
    singular: 'Block Reference',
    plural: 'Block References',
  },
  fields: [
    {
      name: 'ref',
      type: 'relationship',
      relationTo: 'block-library',
      required: true,
      admin: {
        description:
          'Pick a block from the library. Edit the library entry to update it everywhere.',
      },
    },
  ],
}

export const ALL_BLOCKS = [
  HeroBlock,
  LogosBlock,
  ProblemBlock,
  StatsBlock,
  HowItWorksBlock,
  FeaturesBlock,
  AiIntelligenceBlock,
  InsightsBlock,
  IndustriesBlock,
  ComparisonBlock,
  PricingBlock,
  TestimonialsBlock,
  FaqBlock,
  RichTextBlock,
  CtaBlock,
  ImageBlock,
  LoginFormBlock,
  RegisterFormBlock,
  BlockRefBlock,
]
