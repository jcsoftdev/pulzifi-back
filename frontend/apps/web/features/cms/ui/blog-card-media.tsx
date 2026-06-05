import Image from 'next/image'
import { categoryStyle } from '../lib/category-style'

type Props = {
  title: string
  heroImageUrl?: string
  category?: string
  sizes: string
  className?: string
}

/**
 * Card/hero imagery: renders the real heroImage when present, otherwise a
 * category-colored gradient block with the category label so image-less posts
 * still read as intentional.
 */
export function BlogCardMedia({ title, heroImageUrl, category, sizes, className }: Props) {
  if (heroImageUrl) {
    return (
      <div className={`relative overflow-hidden ${className ?? ''}`}>
        <Image src={heroImageUrl} alt={title} fill className="object-cover" sizes={sizes} />
      </div>
    )
  }

  const style = categoryStyle(category)
  return (
    <div
      className={`relative flex items-center justify-center overflow-hidden bg-gradient-to-br ${style.gradient} ${className ?? ''}`}
    >
      <span className="pz-display text-2xl text-[var(--pz-ink)] opacity-25">{style.label}</span>
    </div>
  )
}
