interface TestimonialCardProps {
  quote: string
  author: string
  role: string
}

export function TestimonialCard({ quote, author, role }: Readonly<TestimonialCardProps>) {
  return (
    <div className="flex w-[280px] shrink-0 flex-col justify-between gap-10 rounded-3xl border border-black/10 bg-white p-6 shadow-[0_0_30px_rgba(0,0,0,0.02)] sm:w-[329px]">
      <p className="text-lg leading-7 text-[#444141] sm:text-xl">&ldquo;{quote}&rdquo;</p>
      <div className="flex flex-col gap-1.5">
        <span className="text-xl font-semibold leading-7 tracking-[-0.6px] capitalize text-[#1f1f1f]">
          {author}
        </span>
        <span className="text-base leading-6 text-[#888]">{role}</span>
      </div>
    </div>
  )
}
