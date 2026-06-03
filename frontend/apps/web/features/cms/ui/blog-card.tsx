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
      className="group flex flex-col rounded-2xl bg-[var(--pz-card-bg)] shadow-[var(--pz-card-shadow-rest)] ring-1 ring-[var(--pz-card-border)] transition-all duration-200 hover:shadow-[var(--pz-card-shadow-hover)] hover:ring-[var(--pz-accent-line)]"
    >
      {heroImageUrl ? (
        // biome-ignore lint/performance/noImgElement: external CMS image, dimensions unknown at build time
        <img
          src={heroImageUrl}
          alt={title}
          className="h-48 w-full rounded-t-2xl object-cover"
        />
      ) : (
        <div className="h-2 w-full rounded-t-2xl bg-[var(--pz-accent-tint)]" />
      )}
      <div className="flex flex-1 flex-col p-6">
        {category && (
          <span className="mb-3 inline-block self-start rounded-full bg-[var(--pz-accent-tint)] px-3 py-1 text-xs font-medium text-[var(--pz-accent)]">
            {category}
          </span>
        )}
        <h2 className="mb-2 text-xl font-bold leading-snug text-[var(--pz-ink)] transition-colors group-hover:text-[var(--pz-accent)]">
          {title}
        </h2>
        {excerpt && (
          <p className="line-clamp-3 flex-1 text-sm leading-relaxed text-[var(--pz-ink-2)]">
            {excerpt}
          </p>
        )}
        {publishedAt && (
          <time
            dateTime={publishedAt}
            className="mt-4 text-xs text-[var(--pz-ink-2)] opacity-60"
          >
            {new Date(publishedAt).toLocaleDateString('en-US', {
              year: 'numeric',
              month: 'short',
              day: 'numeric',
            })}
          </time>
        )}
      </div>
    </Link>
  )
}
