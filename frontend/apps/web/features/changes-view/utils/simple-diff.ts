export type DiffSegmentType = 'added' | 'removed' | 'unchanged'

export interface DiffSegment {
  type: DiffSegmentType
  text: string
}

/**
 * kind:
 *  'inline'  – content-matched paragraph; segments contain mixed removed/added/unchanged words
 *  'removed' – paragraph disappeared with no counterpart (pure deletion)
 *  'added'   – paragraph appeared with no counterpart (pure addition)
 */
export interface DiffRow {
  kind: 'inline' | 'removed' | 'added'
  segments: DiffSegment[]
}

export type DiffResult = DiffRow[]

// ---------------------------------------------------------------------------
// Word-level diff (LCS on individual words within a paragraph)
// ---------------------------------------------------------------------------

function buildLcsTable(a: string[], b: string[]): number[][] {
  const m = a.length
  const n = b.length
  const dp: number[][] = Array.from({ length: m + 1 }, () => new Array(n + 1).fill(0))
  for (let i = 1; i <= m; i++) {
    for (let j = 1; j <= n; j++) {
      dp[i]![j] =
        a[i - 1] === b[j - 1]
          ? dp[i - 1]![j - 1]! + 1
          : Math.max(dp[i - 1]![j]!, dp[i]![j - 1]!)
    }
  }
  return dp
}

function backtrackLcs(
  oldWords: string[],
  newWords: string[],
  dp: number[][],
): Array<{ type: DiffSegmentType; word: string }> {
  const rawOps: Array<{ type: DiffSegmentType; word: string }> = []
  let i = oldWords.length
  let j = newWords.length

  while (i > 0 || j > 0) {
    if (i > 0 && j > 0 && oldWords[i - 1] === newWords[j - 1]) {
      rawOps.unshift({ type: 'unchanged', word: oldWords[i - 1]! })
      i--
      j--
    } else if (i > 0 && (j === 0 || dp[i - 1]![j]! >= dp[i]![j - 1]!)) {
      rawOps.unshift({ type: 'removed', word: oldWords[i - 1]! })
      i--
    } else {
      rawOps.unshift({ type: 'added', word: newWords[j - 1]! })
      j--
    }
  }

  return rawOps
}

function diffWords(oldLine: string, newLine: string): DiffSegment[] {
  const oldWords = oldLine.split(/\s+/).filter(Boolean)
  const newWords = newLine.split(/\s+/).filter(Boolean)

  if (oldWords.length === 0) return newLine.trim() ? [{ type: 'added', text: newLine.trim() }] : []
  if (newWords.length === 0) return oldLine.trim() ? [{ type: 'removed', text: oldLine.trim() }] : []

  const dp = buildLcsTable(oldWords, newWords)
  const rawOps = backtrackLcs(oldWords, newWords, dp)

  const segments: DiffSegment[] = []
  for (const op of rawOps) {
    const last = segments.at(-1)
    if (last?.type === op.type) last.text += ` ${op.word}`
    else segments.push({ type: op.type, text: op.word })
  }

  return segments
}

// ---------------------------------------------------------------------------
// Content-based paragraph matching (Jaccard word similarity)
// ---------------------------------------------------------------------------

const MATCH_THRESHOLD = 0.25

function jaccardSimilarity(a: string, b: string): number {
  const wa = new Set(a.toLowerCase().split(/\s+/).filter(Boolean))
  const wb = new Set(b.toLowerCase().split(/\s+/).filter(Boolean))
  if (wa.size === 0 && wb.size === 0) return 1
  if (wa.size === 0 || wb.size === 0) return 0
  let intersect = 0
  for (const w of wa) if (wb.has(w)) intersect++
  return intersect / (wa.size + wb.size - intersect)
}

function matchParagraphs(oldParas: string[], newParas: string[]): Map<number, number> {
  const candidates: Array<{ oldIdx: number; newIdx: number; score: number }> = []
  for (let i = 0; i < oldParas.length; i++) {
    for (let j = 0; j < newParas.length; j++) {
      const score = jaccardSimilarity(oldParas[i]!, newParas[j]!)
      if (score >= MATCH_THRESHOLD) candidates.push({ oldIdx: i, newIdx: j, score })
    }
  }
  candidates.sort((a, b) => b.score - a.score)

  const usedOld = new Set<number>()
  const usedNew = new Set<number>()
  const newToOld = new Map<number, number>()
  for (const { oldIdx, newIdx } of candidates) {
    if (!usedOld.has(oldIdx) && !usedNew.has(newIdx)) {
      newToOld.set(newIdx, oldIdx)
      usedOld.add(oldIdx)
      usedNew.add(newIdx)
    }
  }
  return newToOld
}

// ---------------------------------------------------------------------------
// Public API
// ---------------------------------------------------------------------------

export function diffLines(oldText: string, newText: string): DiffResult {
  const oldParas = oldText.split('\n').map((l) => l.trim()).filter(Boolean)
  const newParas = newText.split('\n').map((l) => l.trim()).filter(Boolean)

  const newToOld = matchParagraphs(oldParas, newParas)
  const matchedOldIndices = new Set(newToOld.values())

  const unmatchedOld = oldParas
    .map((text, i) => ({ text, i }))
    .filter(({ i }) => !matchedOldIndices.has(i))
  let unmatchedOldCursor = 0

  const result: DiffResult = []

  for (let j = 0; j < newParas.length; j++) {
    const matchedOldIdx = newToOld.get(j)

    if (matchedOldIdx === undefined) {
      // Emit the positionally corresponding removed paragraph first,
      // then the new one — the UI groups them into one before/after card
      if (unmatchedOldCursor < unmatchedOld.length) {
        const entry = unmatchedOld[unmatchedOldCursor++]!
        result.push({ kind: 'removed', segments: [{ type: 'removed', text: entry.text }] })
      }
      result.push({ kind: 'added', segments: [{ type: 'added', text: newParas[j]! }] })
    } else {
      const oldPara = oldParas[matchedOldIdx]!
      const newPara = newParas[j]!
      if (oldPara === newPara) continue
      const segments = diffWords(oldPara, newPara)
      if (segments.length > 0) result.push({ kind: 'inline', segments })
    }
  }

  // Net deletions (more removed than added)
  while (unmatchedOldCursor < unmatchedOld.length) {
    const entry = unmatchedOld[unmatchedOldCursor++]!
    result.push({ kind: 'removed', segments: [{ type: 'removed', text: entry.text }] })
  }

  return result
}
