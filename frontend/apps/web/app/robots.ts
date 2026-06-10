import type { MetadataRoute } from 'next'

const privateRoutes = [
  '/api/',
  '/cms-api/',
  '/cms-admin',
  '/preview/',
  '/dashboard',
  '/workspaces',
  '/settings',
  '/team',
  '/admin',
  '/onboarding',
  '/lecture-ai',
]

export default function robots(): MetadataRoute.Robots {
  const baseUrl = process.env.NEXT_PUBLIC_APP_BASE_URL || 'https://pulzifi.com'

  return {
    rules: [
      // Default: allow public pages, block private app routes
      {
        userAgent: '*',
        allow: '/',
        disallow: privateRoutes,
      },
      // Explicit allow for AI crawlers (GEO/AEO indexing)
      {
        userAgent: 'GPTBot',
        allow: '/',
        disallow: privateRoutes,
      },
      {
        userAgent: 'ClaudeBot',
        allow: '/',
        disallow: privateRoutes,
      },
      {
        userAgent: 'anthropic-ai',
        allow: '/',
        disallow: privateRoutes,
      },
      {
        userAgent: 'PerplexityBot',
        allow: '/',
        disallow: privateRoutes,
      },
      {
        userAgent: 'Google-Extended',
        allow: '/',
        disallow: privateRoutes,
      },
      {
        userAgent: 'OAI-SearchBot',
        allow: '/',
        disallow: privateRoutes,
      },
      {
        userAgent: 'CCBot',
        allow: '/',
        disallow: privateRoutes,
      },
      {
        userAgent: 'meta-externalagent',
        allow: '/',
        disallow: privateRoutes,
      },
      {
        userAgent: 'Applebot-Extended',
        allow: '/',
        disallow: privateRoutes,
      },
      {
        userAgent: 'Bytespider',
        allow: '/',
        disallow: privateRoutes,
      },
    ],
    sitemap: `${baseUrl}/sitemap.xml`,
  }
}
