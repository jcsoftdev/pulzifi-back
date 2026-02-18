'use client'

import { cn } from '@workspace/ui/lib/utils'
import type { DiffRow, DiffSegment } from '../utils/simple-diff'

export interface TextChangesProps {
  changes?: DiffRow[]
}

// ---------------------------------------------------------------------------
// Group consecutive removed→added rows into unified before/after blocks,
// matching how git's unified diff groups deletions with their replacements.
// ---------------------------------------------------------------------------

type InlineGroup = { kind: 'inline'; segments: DiffSegment[] }
type BlockGroup = { kind: 'block'; removed: string | null; added: string | null }
type DisplayGroup = InlineGroup | BlockGroup

function buildGroups(rows: DiffRow[]): DisplayGroup[] {
  const groups: DisplayGroup[] = []
  let i = 0
  while (i < rows.length) {
    const row = rows[i]!
    if (row.kind === 'removed') {
      const next = rows[i + 1]
      if (next?.kind === 'added') {
        groups.push({
          kind: 'block',
          removed: row.segments[0]?.text ?? null,
          added: next.segments[0]?.text ?? null,
        })
        i += 2
      } else {
        groups.push({ kind: 'block', removed: row.segments[0]?.text ?? null, added: null })
        i++
      }
    } else if (row.kind === 'added') {
      groups.push({ kind: 'block', removed: null, added: row.segments[0]?.text ?? null })
      i++
    } else {
      // 'inline' — word-level diff within a content-matched paragraph
      groups.push({ kind: 'inline', segments: row.segments })
      i++
    }
  }
  return groups
}

// ---------------------------------------------------------------------------
// Segment renderer shared between inline and block rows
// ---------------------------------------------------------------------------

function Segments({ segments }: Readonly<{ segments: DiffSegment[] }>) {
  return (
    <>
      {segments.map((seg, si) => (
        <span
          key={si}
          className={cn(
            si > 0 && 'ml-[0.25em]',
            seg.type === 'removed' && 'line-through text-foreground/40',
            seg.type === 'added' && 'text-emerald-400',
            seg.type === 'unchanged' && 'text-foreground',
          )}
        >
          {seg.text}
        </span>
      ))}
    </>
  )
}

// ---------------------------------------------------------------------------
// Main component
// ---------------------------------------------------------------------------

export function TextChanges({ changes = [] }: Readonly<TextChangesProps>) {
  if (!changes || changes.length === 0) {
    return (
      <div className="flex items-center justify-center h-64 rounded-xl border border-border bg-muted/10">
        <p className="text-sm text-muted-foreground">No text changes detected</p>
      </div>
    )
  }

  const groups = buildGroups(changes)

  return (
    <div className="rounded-xl border border-border bg-card">
      <div className="px-5 py-3.5 border-b border-border">
        <h3 className="text-sm font-medium text-muted-foreground tracking-wide">Text Changes</h3>
      </div>

      <div className="p-4 space-y-2 max-h-[520px] overflow-y-auto">
        {groups.map((group, idx) => {
          // ── Inline word-diff ─────────────────────────────────────────────
          if (group.kind === 'inline') {
            return (
              <div
                key={idx}
                className="px-4 py-3 rounded-lg border border-border/40 bg-muted/10 text-sm leading-relaxed"
              >
                <Segments segments={group.segments} />
              </div>
            )
          }

          // ── Before / after block ─────────────────────────────────────────
          // Mirrors git's unified diff: old line then new line in one hunk
          return (
            <div
              key={idx}
              className="rounded-lg border border-border/40 overflow-hidden text-sm"
            >
              {group.removed !== null && (
                <div className="flex gap-3 px-4 py-2.5 border-b border-border/30 bg-muted/10">
                  <span className="select-none shrink-0 text-foreground/25 font-mono text-xs pt-px">
                    −
                  </span>
                  <p className="leading-relaxed line-through text-foreground/40">
                    {group.removed}
                  </p>
                </div>
              )}
              {group.added !== null && (
                <div className="flex gap-3 px-4 py-2.5 bg-emerald-950/20">
                  <span className="select-none shrink-0 text-emerald-500/60 font-mono text-xs pt-px">
                    +
                  </span>
                  <p className="leading-relaxed text-emerald-400">{group.added}</p>
                </div>
              )}
            </div>
          )
        })}
      </div>
    </div>
  )
}
