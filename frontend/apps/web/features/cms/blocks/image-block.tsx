type Props = {
  block: {
    blockType: 'image'
    image: {
      url?: string | null
      alt?: string | null
    }
    caption?: string | null
    size?: 'full' | 'contained' | null
  }
}

export function ImageBlock({ block }: Props) {
  const wrapClass = block.size === 'full' ? 'w-full' : 'mx-auto max-w-4xl px-4'
  return (
    <figure className={wrapClass}>
      {/* eslint-disable-next-line @next/next/no-img-element */}
      {/* biome-ignore lint/performance/noImgElement: CMS image block — dimensions unknown at build time */}
      <img src={block.image.url ?? ''} alt={block.image.alt ?? ''} className="w-full rounded-2xl" />
      {block.caption && (
        <figcaption className="mt-2 text-center text-sm text-gray-500">{block.caption}</figcaption>
      )}
    </figure>
  )
}
