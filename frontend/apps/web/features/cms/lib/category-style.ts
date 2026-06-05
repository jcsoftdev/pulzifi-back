// Single source of truth for blog category visuals — drives the colored pill
// and the gradient placeholder shown when a post has no hero image.
// Guide → violet (brand accent), Product → teal, Company → gold.

export type BlogCategory = 'Guide' | 'Product' | 'Company'

export const BLOG_CATEGORIES: BlogCategory[] = [
  'Guide',
  'Product',
  'Company',
]

type CategoryStyle = {
  label: string
  /** Pill background + text classes. */
  pill: string
  /** Tailwind gradient stop classes for the image-less placeholder. */
  gradient: string
}

const STYLES: Record<BlogCategory, CategoryStyle> = {
  Guide: {
    label: 'Guide',
    pill: 'bg-[color-mix(in_srgb,var(--pz-accent)_12%,white)] text-[var(--pz-accent)]',
    gradient:
      'from-[color-mix(in_srgb,var(--pz-accent)_26%,white)] to-[color-mix(in_srgb,var(--pz-accent)_7%,white)]',
  },
  Product: {
    label: 'Product',
    pill: 'bg-[color-mix(in_srgb,var(--pz-accent-teal)_14%,white)] text-[color-mix(in_srgb,var(--pz-accent-teal)_72%,black)]',
    gradient:
      'from-[color-mix(in_srgb,var(--pz-accent-teal)_28%,white)] to-[color-mix(in_srgb,var(--pz-accent-teal)_7%,white)]',
  },
  Company: {
    label: 'Company',
    pill: 'bg-[color-mix(in_srgb,var(--pz-accent-gold)_16%,white)] text-[color-mix(in_srgb,var(--pz-accent-gold)_68%,black)]',
    gradient:
      'from-[color-mix(in_srgb,var(--pz-accent-gold)_30%,white)] to-[color-mix(in_srgb,var(--pz-accent-gold)_8%,white)]',
  },
}

/** Resolve a category to its visual style, falling back to Guide. */
export function categoryStyle(category?: string): CategoryStyle {
  if (category && category in STYLES) return STYLES[category as BlogCategory]
  return STYLES.Guide
}
