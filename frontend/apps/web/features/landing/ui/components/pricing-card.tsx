import { Check } from 'lucide-react'
import { LandingButton } from './landing-button'

interface PricingCardProps {
  name: string
  price: string
  period?: string
  description: string
  cta: string
  features: readonly string[]
  popular?: boolean
}

export function PricingCard({
  name,
  price,
  period,
  description,
  cta,
  features,
  popular,
}: Readonly<PricingCardProps>) {
  return (
    <div className="flex flex-1 flex-col gap-6 rounded-3xl bg-white p-5 sm:p-[30px]">
      <div className="flex flex-col gap-3.5">
        <div className="flex items-center justify-between">
          <h3 className="text-xl font-medium leading-8 tracking-[-0.6px] text-[#111] capitalize">
            {name}
          </h3>
          {popular && (
            <span className="rounded-full bg-[#f2ebfd] px-2.5 py-1.5 text-sm font-medium leading-5 tracking-tight text-[#29144c]">
              Most Popular
            </span>
          )}
        </div>
        <div className="flex items-end gap-2.5">
          <span className="text-5xl font-medium leading-[56px] tracking-[-1.44px] text-[#111] capitalize">
            {price}
          </span>
          {period && (
            <span className="text-xl font-medium leading-8 tracking-[-0.6px] text-[#777]">
              {period}
            </span>
          )}
        </div>
        <p className="text-sm leading-5 text-[#444141]">{description}</p>
      </div>

      <LandingButton href="/register" variant="dark" className="w-full rounded-[10px]">
        {cta}
      </LandingButton>

      <div className="flex flex-col gap-4 p-3.5">
        <h4 className="text-xl font-medium leading-8 tracking-[-0.6px] text-[#111]">Features:</h4>
        <ul className="flex flex-col gap-2.5">
          {features.map((feature) => (
            <li key={feature} className="flex items-center gap-3.5">
              <span className="flex size-5 shrink-0 items-center justify-center rounded-full bg-[#29144c]">
                <Check className="size-3 text-white" strokeWidth={3} />
              </span>
              <span className="text-base leading-6 text-[#444141]">{feature}</span>
            </li>
          ))}
        </ul>
      </div>
    </div>
  )
}
