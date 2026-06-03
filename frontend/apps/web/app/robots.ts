import type { MetadataRoute } from 'next'

const privateRoutes = ['/api/', '/dashboard', '/workspaces', '/settings', '/team', '/admin']

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
    ],
    sitemap: `${baseUrl}/sitemap.xml`,
  }
}
