interface FeatureCardProps {
  title: string
  description: string
  image: string
}

export function FeatureCard({ title, description, image }: Readonly<FeatureCardProps>) {
  return (
    <div className="group relative flex flex-1 flex-col overflow-hidden rounded-3xl bg-white">
      <div className="relative h-[280px] overflow-hidden">
        {/* eslint-disable-next-line @next/next/no-img-element */}
        {/* biome-ignore lint/performance/noImgElement: image dimensions unknown at build time */}
        <img
          src={image}
          alt={title}
          className="absolute inset-0 h-full w-full object-contain px-4 pt-4 transition-transform duration-500 group-hover:scale-105"
        />
        <div className="absolute inset-x-0 bottom-0 h-[120px] bg-gradient-to-b from-transparent to-white" />
      </div>
      <div className="flex flex-col gap-3 items-center px-6 py-6 text-center">
        <h3 className="text-[28px] font-normal leading-9 tracking-[-1.8px] text-[var(--pz-ink)]">
          {title}
        </h3>
        <p className="text-base leading-6 text-[var(--pz-ink-2)]">{description}</p>
      </div>
    </div>
  )
}
