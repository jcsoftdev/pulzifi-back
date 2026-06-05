// Plain, serializable shape the blog list/cards consume. The server page maps
// Payload's post docs into this so the client filter component never touches
// the raw CMS shape.

export type BlogPostItem = {
  id: string
  title: string
  slug: string
  excerpt?: string
  heroImageUrl?: string
  publishedAt?: string
  category?: string
  author?: string
  readingTime?: number
}

/** Human date for blog surfaces, e.g. "May 25, 2026". */
export function formatBlogDate(value?: string): string | undefined {
  if (!value) return undefined
  return new Date(value).toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
  })
}
