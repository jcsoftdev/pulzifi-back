import Link from 'next/link'

type Props = {
  title: string
  slug: string
  excerpt?: string
  heroImageUrl?: string
  publishedAt?: string
  category?: string
}

export function BlogCard({ title, slug, excerpt, heroImageUrl, publishedAt, category }: Props) {
  return (
    <Link
      href={`/blog/${slug}`}
      className="group block rounded-2xl bg-white p-6 shadow-sm transition hover:shadow-md"
    >
      {heroImageUrl && (
        // biome-ignore lint/performance/noImgElement: external CMS image, dimensions unknown at build time
        <img src={heroImageUrl} alt={title} className="mb-4 h-48 w-full rounded-xl object-cover" />
      )}
      {category && (
        <span className="mb-2 inline-block rounded-full bg-purple-100 px-3 py-1 text-xs font-medium text-purple-700">
          {category}
        </span>
      )}
      <h2 className="mb-2 text-xl font-bold text-gray-900 group-hover:text-purple-700">{title}</h2>
      {excerpt && <p className="text-sm text-gray-500">{excerpt}</p>}
      {publishedAt && (
        <p className="mt-3 text-xs text-gray-400">{new Date(publishedAt).toLocaleDateString()}</p>
      )}
    </Link>
  )
}
