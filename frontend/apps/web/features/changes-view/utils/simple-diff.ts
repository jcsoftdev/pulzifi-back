export type DiffSegmentType = 'added' | 'removed' | 'unchanged'

export interface DiffSegment {
  type: DiffSegmentType
  text: string
}

export interface DiffRow {
  segments: DiffSegment[]
}

export type DiffResult = DiffRow[]

function buildLcsTable(a: string[], b: string[]): number[][] {
  const m = a.length
  const n = b.length
  const dp: number[][] = Array.from(
    {
      length: m + 1,
    },
    () => new Array(n + 1).fill(0)
  )
  for (let i = 1; i <= m; i++) {
    for (let j = 1; j <= n; j++) {
      if (a[i - 1] === b[j - 1]) {
        dp[i][j] = dp[i - 1][j - 1] + 1
      } else {
        dp[i][j] = Math.max(dp[i - 1][j], dp[i][j - 1])
      }
    }
  }
  return dp
}

function diffWords(oldLine: string, newLine: string): DiffSegment[] {
  const oldWords = oldLine.split(/\s+/).filter(Boolean)
  const newWords = newLine.split(/\s+/).filter(Boolean)

  if (oldWords.length === 0) {
    return newLine.trim()
      ? [
          {
            type: 'added',
            text: newLine.trim(),
          },
        ]
      : []
  }
  if (newWords.length === 0) {
    return oldLine.trim()
      ? [
          {
            type: 'removed',
            text: oldLine.trim(),
          },
        ]
      : []
  }

  const dp = buildLcsTable(oldWords, newWords)
  const rawOps: Array<{
    type: DiffSegmentType
    word: string
  }> = []
  let i = oldWords.length
  let j = newWords.length

  while (i > 0 || j > 0) {
    if (i > 0 && j > 0 && oldWords[i - 1] === newWords[j - 1]) {
      rawOps.unshift({
        type: 'unchanged',
        word: oldWords[i - 1],
      })
      i--
      j--
    } else if (i > 0 && (j === 0 || dp[i - 1][j] >= dp[i][j - 1])) {
      rawOps.unshift({
        type: 'removed',
        word: oldWords[i - 1],
      })
      i--
    } else {
      rawOps.unshift({
        type: 'added',
        word: newWords[j - 1],
      })
      j--
    }
  }

  // Merge consecutive same-type ops into single segments
  const segments: DiffSegment[] = []
  for (const op of rawOps) {
    const last = segments[segments.length - 1]
    if (last && last.type === op.type) {
      last.text += ` ${op.word}`
    } else {
      segments.push({
        type: op.type,
        text: op.word,
      })
    }
  }

  return segments
}

export function diffLines(oldText: string, newText: string): DiffResult {
  const oldLines = oldText
    .split('\n')
    .map((l) => l.trim())
    .filter(Boolean)
  const newLines = newText
    .split('\n')
    .map((l) => l.trim())
    .filter(Boolean)

  const dp = buildLcsTable(oldLines, newLines)
  const lineOps: Array<{
    type: 'added' | 'removed' | 'unchanged'
    line: string
  }> = []
  let i = oldLines.length
  let j = newLines.length

  while (i > 0 || j > 0) {
    if (i > 0 && j > 0 && oldLines[i - 1] === newLines[j - 1]) {
      lineOps.unshift({
        type: 'unchanged',
        line: oldLines[i - 1],
      })
      i--
      j--
    } else if (i > 0 && (j === 0 || dp[i - 1][j] >= dp[i][j - 1])) {
      lineOps.unshift({
        type: 'removed',
        line: oldLines[i - 1],
      })
      i--
    } else {
      lineOps.unshift({
        type: 'added',
        line: newLines[j - 1],
      })
      j--
    }
  }

  const result: DiffResult = []
  let k = 0

  while (k < lineOps.length) {
    const op = lineOps[k]

    // Skip unchanged lines — only show what changed
    if (op.type === 'unchanged') {
      k++
      continue
    }

    // Collect a contiguous group of removed/added ops
    const removedLines: string[] = []
    const addedLines: string[] = []
    while (k < lineOps.length && lineOps[k].type !== 'unchanged') {
      if (lineOps[k].type === 'removed') removedLines.push(lineOps[k].line)
      else addedLines.push(lineOps[k].line)
      k++
    }

    // Pair removed↔added lines for inline word-level diff
    const pairs = Math.min(removedLines.length, addedLines.length)
    for (let p = 0; p < pairs; p++) {
      const segments = diffWords(removedLines[p], addedLines[p])
      if (segments.length > 0)
        result.push({
          segments,
        })
    }

    // Unpaired deletions
    for (let p = pairs; p < removedLines.length; p++) {
      result.push({
        segments: [
          {
            type: 'removed',
            text: removedLines[p],
          },
        ],
      })
    }

    // Unpaired insertions
    for (let p = pairs; p < addedLines.length; p++) {
      result.push({
        segments: [
          {
            type: 'added',
            text: addedLines[p],
          },
        ],
      })
    }
  }

  return result
}
