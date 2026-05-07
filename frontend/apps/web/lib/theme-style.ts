type ThemeData = Record<string, string | null | undefined> | null | undefined

const THEME_VAR_MAP: Array<[string, string]> = [
  ['--pz-page-bg', 'pageBg'],
  ['--pz-page-bg-alt', 'pageBgAlt'],
  ['--pz-card-bg', 'cardBg'],
  ['--pz-dark-surface', 'darkSurface'],
  ['--pz-ink', 'inkPrimary'],
  ['--pz-ink-2', 'inkSecondary'],
  ['--pz-accent', 'accentPrimary'],
  ['--pz-accent-muted', 'accentMuted'],
  ['--pz-accent-gold', 'accentGold'],
  ['--pz-accent-teal', 'accentTeal'],
  ['--pz-border', 'border'],
  ['--pz-border-strong', 'borderStrong'],
]

export function buildThemeStyle(theme: ThemeData): string {
  if (!theme) return ''
  const decls = THEME_VAR_MAP.filter(([, field]) => typeof theme[field] === 'string' && (theme[field] as string).trim().length > 0)
    .map(([cssVar, field]) => `${cssVar}: ${theme[field]};`)
    .join(' ')
  return decls ? `:root { ${decls} }` : ''
}
