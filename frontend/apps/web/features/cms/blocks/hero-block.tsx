import { HeroSection } from '@/features/landing'

type Alert = {
  tone?: 'signal' | 'amber' | 'teal' | 'ink' | null
  icon?: string | null
  site?: string | null
  title?: string | null
  detail?: string | null
  time?: string | null
}

type Kpi = {
  label: string
  value: string
  delta?: string | null
  deltaDirection?: 'up' | 'down' | null
}

type Props = {
  block: {
    blockType: 'hero'
    eyebrowBadge?: string | null
    eyebrowText?: string | null
    headline?: string | null
    headlineHighlight?: string | null
    subheadline?: string | null
    primaryCtaLabel?: string | null
    primaryCtaHref?: string | null
    secondaryCtaLabel?: string | null
    secondaryCtaHref?: string | null
    trustLine?: string | null
    dashboardAlerts?: Alert[] | null
    aiInsightTitle?: string | null
    aiInsightBody?: string | null
    kpis?: Kpi[] | null
  }
}

export function HeroBlock({ block }: Props) {
  return (
    <HeroSection
      eyebrowBadge={block.eyebrowBadge ?? undefined}
      eyebrowText={block.eyebrowText ?? undefined}
      headline={block.headline ?? undefined}
      headlineHighlight={block.headlineHighlight ?? undefined}
      subheadline={block.subheadline ?? undefined}
      primaryCtaLabel={block.primaryCtaLabel ?? undefined}
      primaryCtaHref={block.primaryCtaHref ?? undefined}
      secondaryCtaLabel={block.secondaryCtaLabel ?? undefined}
      secondaryCtaHref={block.secondaryCtaHref ?? undefined}
      trustLine={block.trustLine ?? undefined}
      aiInsightTitle={block.aiInsightTitle ?? undefined}
      aiInsightBody={block.aiInsightBody ?? undefined}
      kpis={
        block.kpis?.map((k) => ({
          label: k.label,
          value: k.value,
          delta: k.delta ?? undefined,
          deltaDirection: (k.deltaDirection ?? 'up') as 'up' | 'down',
        })) ?? undefined
      }
      dashboardAlerts={
        block.dashboardAlerts?.map((a) => ({
          tone: (a.tone ?? 'ink') as 'signal' | 'amber' | 'teal' | 'ink',
          icon: a.icon ?? undefined,
          site: a.site ?? '',
          title: a.title ?? '',
          detail: a.detail ?? undefined,
          time: a.time ?? undefined,
        })) ?? undefined
      }
    />
  )
}
