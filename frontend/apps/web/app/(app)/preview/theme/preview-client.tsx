'use client'

import { useLivePreview } from '@payloadcms/live-preview-react'
import { useMemo } from 'react'

import { PreviewModeProvider } from '@/features/landing/lib/preview-mode'
import { buildThemeStyle } from '@/lib/theme-style'

type ThemeData = Record<string, string | null | undefined>

type Props = {
  initialTheme: ThemeData
  serverURL: string
}

const SWATCH_GROUPS: Array<{
  heading: string
  swatches: Array<{
    key: string
    cssVar: string
    label: string
    defaultValue: string
  }>
}> = [
  {
    heading: 'Surfaces',
    swatches: [
      {
        key: 'pageBg',
        cssVar: '--pz-page-bg',
        label: 'Page background',
        defaultValue: '#f3f3f3',
      },
      {
        key: 'pageBgAlt',
        cssVar: '--pz-page-bg-alt',
        label: 'Alt background',
        defaultValue: '#fafafa',
      },
      {
        key: 'cardBg',
        cssVar: '--pz-card-bg',
        label: 'Card background',
        defaultValue: '#ffffff',
      },
      {
        key: 'darkSurface',
        cssVar: '--pz-dark-surface',
        label: 'Dark surface',
        defaultValue: '#29144c',
      },
    ],
  },
  {
    heading: 'Text',
    swatches: [
      {
        key: 'inkPrimary',
        cssVar: '--pz-ink',
        label: 'Primary text',
        defaultValue: '#131313',
      },
      {
        key: 'inkSecondary',
        cssVar: '--pz-ink-2',
        label: 'Secondary text',
        defaultValue: '#444141',
      },
    ],
  },
  {
    heading: 'Accents',
    swatches: [
      {
        key: 'accentPrimary',
        cssVar: '--pz-accent',
        label: 'Primary accent',
        defaultValue: '#7c3aed',
      },
      {
        key: 'accentMuted',
        cssVar: '--pz-accent-muted',
        label: 'Accent muted',
        defaultValue: '#a78bfa',
      },
      {
        key: 'accentGold',
        cssVar: '--pz-accent-gold',
        label: 'Gold',
        defaultValue: '#f59e0b',
      },
      {
        key: 'accentTeal',
        cssVar: '--pz-accent-teal',
        label: 'Teal',
        defaultValue: '#14b8a6',
      },
    ],
  },
  {
    heading: 'Borders',
    swatches: [
      {
        key: 'border',
        cssVar: '--pz-card-border',
        label: 'Subtle border',
        defaultValue: 'rgba(0,0,0,0.08)',
      },
      {
        key: 'borderStrong',
        cssVar: '--pz-border-strong',
        label: 'Strong border',
        defaultValue: 'rgba(0,0,0,0.16)',
      },
    ],
  },
]

const DERIVED = [
  {
    cssVar: '--pz-accent-tint',
    label: 'Accent tint (10%)',
  },
  {
    cssVar: '--pz-accent-soft',
    label: 'Accent soft (5%)',
  },
  {
    cssVar: '--pz-accent-line',
    label: 'Accent line (18%)',
  },
  {
    cssVar: '--pz-accent-pale',
    label: 'Accent pale',
  },
]

export function ThemePreviewClient({ initialTheme, serverURL }: Props) {
  const { data: theme } = useLivePreview<ThemeData>({
    initialData: initialTheme,
    serverURL,
    depth: 0,
  })

  const themeStyle = useMemo(
    () => buildThemeStyle(theme),
    [
      theme,
    ]
  )

  return (
    <PreviewModeProvider value={true}>
      <div
        className="min-h-screen p-6 md:p-10"
        style={{
          background: 'var(--pz-page-bg, #f3f3f3)',
          color: 'var(--pz-ink, #131313)',
          fontFamily: 'system-ui, -apple-system, sans-serif',
        }}
      >
        {themeStyle && (
          // biome-ignore lint/security/noDangerouslySetInnerHtml: intentional theme CSS injection — controlled server-side input
          <style dangerouslySetInnerHTML={{ __html: themeStyle }} />
        )}

        <header className="mx-auto max-w-[1100px] mb-8">
          <p
            className="text-xs font-semibold uppercase tracking-[0.18em]"
            style={{
              color: 'var(--pz-accent, #7c3aed)',
            }}
          >
            Theme Preview
          </p>
          <h1 className="mt-2 text-3xl font-bold tracking-tight">Design tokens</h1>
          <p
            className="mt-2 text-sm"
            style={{
              color: 'var(--pz-ink-2, #444141)',
            }}
          >
            Edit any color on the right — every swatch and component sample below updates live.
          </p>
        </header>

        {/* Swatches */}
        <section className="mx-auto max-w-[1100px] space-y-8">
          {SWATCH_GROUPS.map((group) => (
            <div key={group.heading}>
              <h2 className="mb-3 text-xs font-semibold uppercase tracking-wider opacity-70">
                {group.heading}
              </h2>
              <div className="grid grid-cols-2 gap-3 md:grid-cols-4">
                {group.swatches.map((s) => {
                  const value = (theme?.[s.key] as string) || s.defaultValue
                  return (
                    <div
                      key={s.key}
                      className="overflow-hidden rounded-lg"
                      style={{
                        background: 'var(--pz-card-bg, #ffffff)',
                        border: '1px solid var(--pz-card-border, rgba(0,0,0,0.08))',
                      }}
                    >
                      <div
                        className="h-20 w-full"
                        style={{
                          background: `var(${s.cssVar}, ${s.defaultValue})`,
                        }}
                      />
                      <div className="px-3 py-2">
                        <p className="text-sm font-medium">{s.label}</p>
                        <p
                          className="font-mono text-xs"
                          style={{
                            color: 'var(--pz-ink-2, #444141)',
                          }}
                        >
                          {value}
                        </p>
                        <p
                          className="font-mono text-[10px] opacity-50"
                          style={{
                            color: 'var(--pz-ink-2, #444141)',
                          }}
                        >
                          {s.cssVar}
                        </p>
                      </div>
                    </div>
                  )
                })}
              </div>
            </div>
          ))}

          {/* Derived tints */}
          <div>
            <h2 className="mb-3 text-xs font-semibold uppercase tracking-wider opacity-70">
              Derived (auto-computed from accent)
            </h2>
            <div className="grid grid-cols-2 gap-3 md:grid-cols-4">
              {DERIVED.map((d) => (
                <div
                  key={d.cssVar}
                  className="overflow-hidden rounded-lg"
                  style={{
                    background: 'var(--pz-card-bg, #ffffff)',
                    border: '1px solid var(--pz-card-border, rgba(0,0,0,0.08))',
                  }}
                >
                  <div
                    className="h-16 w-full"
                    style={{
                      background: `var(${d.cssVar})`,
                    }}
                  />
                  <div className="px-3 py-2">
                    <p className="text-sm font-medium">{d.label}</p>
                    <p
                      className="font-mono text-[10px] opacity-50"
                      style={{
                        color: 'var(--pz-ink-2, #444141)',
                      }}
                    >
                      {d.cssVar}
                    </p>
                  </div>
                </div>
              ))}
            </div>
          </div>
        </section>

        {/* Component samples */}
        <section className="mx-auto mt-12 max-w-[1100px] space-y-8">
          <h2 className="mb-3 text-xs font-semibold uppercase tracking-wider opacity-70">
            Component samples
          </h2>

          {/* Buttons */}
          <div
            className="rounded-xl p-6"
            style={{
              background: 'var(--pz-card-bg, #ffffff)',
              border: '1px solid var(--pz-card-border, rgba(0,0,0,0.08))',
            }}
          >
            <p className="mb-3 text-xs font-semibold uppercase tracking-wider opacity-70">
              Buttons
            </p>
            <div className="flex flex-wrap items-center gap-3">
              <button
                type="button"
                className="rounded-full px-5 py-2.5 text-sm font-medium text-white"
                style={{
                  background: 'var(--pz-accent, #7c3aed)',
                  boxShadow: 'var(--pz-shadow-accent, 0 4px 20px rgba(124,58,237,0.3))',
                }}
              >
                Primary CTA
              </button>
              <button
                type="button"
                className="rounded-full px-5 py-2.5 text-sm font-medium"
                style={{
                  background: 'var(--pz-card-bg, #ffffff)',
                  color: 'var(--pz-ink, #131313)',
                  border: '1px solid var(--pz-card-border, rgba(0,0,0,0.08))',
                }}
              >
                Secondary
              </button>
              <button
                type="button"
                className="rounded-full px-3 py-2 text-sm font-medium"
                style={{
                  color: 'var(--pz-ink-2, #444141)',
                }}
              >
                Ghost link →
              </button>
            </div>
          </div>

          {/* Eyebrow + Highlight + Heading */}
          <div
            className="rounded-xl p-6"
            style={{
              background: 'var(--pz-card-bg, #ffffff)',
              border: '1px solid var(--pz-card-border, rgba(0,0,0,0.08))',
            }}
          >
            <p className="mb-3 text-xs font-semibold uppercase tracking-wider opacity-70">
              Typography
            </p>
            <p
              className="text-xs font-semibold uppercase tracking-[0.18em]"
              style={{
                color: 'var(--pz-accent, #7c3aed)',
              }}
            >
              Eyebrow tag
            </p>
            <h3 className="mt-3 text-3xl font-bold tracking-tight">
              Headline with{' '}
              <em
                className="not-italic"
                style={{
                  color: 'var(--pz-accent, #7c3aed)',
                  fontStyle: 'italic',
                }}
              >
                accent highlight
              </em>
            </h3>
            <p
              className="mt-2 text-base leading-relaxed"
              style={{
                color: 'var(--pz-ink-2, #444141)',
              }}
            >
              Body copy uses the secondary text token for readable contrast.
            </p>
          </div>

          {/* Pricing card simulation */}
          <div
            className="rounded-xl p-6"
            style={{
              background: 'var(--pz-card-bg, #ffffff)',
              border: '2px solid var(--pz-accent, #7c3aed)',
              boxShadow: 'var(--pz-shadow-accent-lg, 0 24px 60px rgba(124,58,237,0.18))',
            }}
          >
            <div className="flex items-center justify-between">
              <p className="font-semibold">Popular plan</p>
              <span
                className="rounded-full px-3 py-1 text-[11px] font-semibold uppercase tracking-wider text-white"
                style={{
                  background: 'var(--pz-accent, #7c3aed)',
                }}
              >
                Pro
              </span>
            </div>
            <p className="mt-3 text-3xl font-bold">$62/mo</p>
            <p
              className="mt-1 text-sm"
              style={{
                color: 'var(--pz-ink-2, #444141)',
              }}
            >
              Featured plan styling: ring + soft accent shadow.
            </p>
          </div>

          {/* Dark surface band */}
          <div
            className="rounded-xl p-8 text-white"
            style={{
              background: 'var(--pz-dark-surface, #29144c)',
            }}
          >
            <p
              className="text-xs font-semibold uppercase tracking-[0.18em]"
              style={{
                color: 'var(--pz-accent-muted, #a78bfa)',
              }}
            >
              Dark surface
            </p>
            <h3 className="mt-2 text-2xl font-bold">Get started CTA band</h3>
            <p className="mt-1 text-sm text-white/70">
              Used for the final CTA section. Accent muted color provides high contrast on dark.
            </p>
            <button
              type="button"
              className="mt-4 rounded-full px-5 py-2.5 text-sm font-medium text-white"
              style={{
                background: 'var(--pz-accent, #7c3aed)',
              }}
            >
              Primary on dark
            </button>
          </div>

          {/* Gold + Teal accent samples */}
          <div className="grid grid-cols-1 gap-3 md:grid-cols-2">
            <div
              className="rounded-xl p-5"
              style={{
                background: 'var(--pz-card-bg, #ffffff)',
                border: '1px solid var(--pz-accent-gold, #f59e0b)',
              }}
            >
              <p className="text-xs font-semibold uppercase tracking-wider opacity-70">
                Gold accent
              </p>
              <p
                className="mt-2 text-2xl font-bold"
                style={{
                  color: 'var(--pz-accent-gold, #f59e0b)',
                }}
              >
                ★★★★★
              </p>
              <p
                className="mt-1 text-xs"
                style={{
                  color: 'var(--pz-ink-2, #444141)',
                }}
              >
                Used in pricing highlight + star ratings.
              </p>
            </div>
            <div
              className="rounded-xl p-5"
              style={{
                background: 'var(--pz-card-bg, #ffffff)',
                border: '1px solid var(--pz-accent-teal, #14b8a6)',
              }}
            >
              <p className="text-xs font-semibold uppercase tracking-wider opacity-70">
                Teal accent
              </p>
              <p
                className="mt-2 text-2xl font-bold"
                style={{
                  color: 'var(--pz-accent-teal, #14b8a6)',
                }}
              >
                ✦ AI Insight
              </p>
              <p
                className="mt-1 text-xs"
                style={{
                  color: 'var(--pz-ink-2, #444141)',
                }}
              >
                Used in AI insight callout in hero dashboard.
              </p>
            </div>
          </div>
        </section>
      </div>
    </PreviewModeProvider>
  )
}
