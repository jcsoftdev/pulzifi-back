export const SECTION_COLORS = [
  'rgb(59 130 246)',  // blue
  'rgb(16 185 129)',  // emerald
  'rgb(245 158 11)',  // amber
  'rgb(239 68 68)',   // red
  'rgb(168 85 247)',  // purple
  'rgb(236 72 153)',  // pink
  'rgb(6 182 212)',   // cyan
  'rgb(249 115 22)',  // orange
]

export function getSectionColor(index: number): string {
  return SECTION_COLORS[index % SECTION_COLORS.length] ?? SECTION_COLORS[0]!
}
