type Props = {
  block: {
    blockType: 'cta'
    headline: string
    subtext?: string | null
    buttonLabel: string
    buttonHref: string
    variant?: 'primary' | 'secondary' | null
  }
}

export function CtaBlock({ block }: Props) {
  return (
    <section className="rounded-3xl bg-[#0b1e3d] px-8 py-16 text-center text-white">
      <h2 className="mb-3 text-3xl font-bold">{block.headline}</h2>
      {block.subtext && <p className="mb-8 text-white/70">{block.subtext}</p>}
      <a
        href={block.buttonHref}
        className={`inline-block rounded-full px-8 py-3 font-semibold ${
          block.variant === 'secondary'
            ? 'border border-white text-white'
            : 'bg-white text-[#0b1e3d]'
        }`}
      >
        {block.buttonLabel}
      </a>
    </section>
  )
}
