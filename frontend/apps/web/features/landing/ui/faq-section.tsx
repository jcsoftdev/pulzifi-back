import { AnimatedSection } from './components/animated-section'
import { FaqItem } from './components/faq-item'
import { SectionHeader } from './components/section-header'

type FaqSectionProps = {
  items?: { question: string; answer: string }[]
}

export function FaqSection({ items }: Readonly<FaqSectionProps> = {}) {
  const faqs = items ?? []

  return (
    <section
      id="faq"
      className="mx-auto max-w-[1256px] overflow-hidden rounded-3xl bg-white px-6 py-12 md:px-[58px] md:py-16"
    >
      <div className="flex flex-col items-center gap-[59px]">
        <AnimatedSection>
          <SectionHeader
            title={<em className="font-heading">Questions Answered</em>}
          />
        </AnimatedSection>

        <div className="flex w-full flex-col gap-4">
          {faqs.map((item, i) => (
            <AnimatedSection key={item.question} animation="fade-up" delay={i * 100}>
              <FaqItem {...item} />
            </AnimatedSection>
          ))}
        </div>
      </div>
    </section>
  )
}
