import type { MetadataRoute } from 'next'
import { getPayloadClient } from '@/lib/payload'

export default async function sitemap(): Promise<MetadataRoute.Sitemap> {
  const baseUrl = process.env.NEXT_PUBLIC_APP_BASE_URL || 'https://pulzifi.com'

  const staticRoutes: MetadataRoute.Sitemap = [
    {
      url: baseUrl,
      lastModified: new Date(),
      changeFrequency: 'weekly',
      priority: 1,
    },
    {
      url: `${baseUrl}/pricing`,
      lastModified: new Date(),
      changeFrequency: 'monthly',
      priority: 0.8,
    },
    {
      url: `${baseUrl}/blog`,
      lastModified: new Date(),
      changeFrequency: 'weekly',
      priority: 0.7,
    },
    {
      url: `${baseUrl}/login`,
      lastModified: new Date(),
      changeFrequency: 'monthly',
      priority: 0.5,
    },
    {
      url: `${baseUrl}/register`,
      lastModified: new Date(),
      changeFrequency: 'monthly',
      priority: 0.8,
    },
  ]

  try {
    const payload = await getPayloadClient()
    const [pages, posts] = await Promise.all([
      payload.find({
        collection: 'pages',
        where: {
          _status: {
            equals: 'published',
          },
        },
        select: {
          slug: true,
        },
        limit: 1000,
      }),
      payload.find({
        collection: 'posts',
        where: {
          _status: {
            equals: 'published',
          },
        },
        select: {
          slug: true,
          publishedAt: true,
        },
        limit: 1000,
      }),
    ])

    const pageRoutes: MetadataRoute.Sitemap = pages.docs.map((page: Record<string, unknown>) => ({
      url: `${baseUrl}/${page.slug}`,
      lastModified: new Date(),
      changeFrequency: 'monthly' as const,
      priority: 0.7,
    }))

    const postRoutes: MetadataRoute.Sitemap = posts.docs.map((post: Record<string, unknown>) => ({
      url: `${baseUrl}/blog/${post.slug}`,
      lastModified: post.publishedAt ? new Date(post.publishedAt as string) : new Date(),
      changeFrequency: 'monthly' as const,
      priority: 0.6,
    }))

    return [
      ...staticRoutes,
      ...pageRoutes,
      ...postRoutes,
    ]
  } catch {
    return staticRoutes
  }
}
