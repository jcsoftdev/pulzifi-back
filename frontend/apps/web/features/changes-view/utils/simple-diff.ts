export type DiffSegmentType = 'added' | 'removed' | 'unchanged'

export interface DiffSegment {
  type: DiffSegmentType
  text: string
}

export interface DiffRow {
  segments: DiffSegment[]
}

export type DiffResult = DiffRow[]

export function diffLines(oldText: string, newText: string): DiffResult {
  const oldLines = oldText.split('\n')
  const newLines = newText.split('\n')
  const result: DiffResult = []

  let i = 0
  let j = 0

  while (i < oldLines.length || j < newLines.length) {
    const oldLine = oldLines[i] ?? ''
    const newLine = newLines[j] ?? ''

    if (i < oldLines.length && j < newLines.length && oldLine === newLine) {
      result.push({
        segments: [
          {
            type: 'unchanged',
            text: oldLine,
          },
        ],
      })
      i++
      j++
      continue
    }

    if (i < oldLines.length && j < newLines.length) {
      const segments: DiffSegment[] = []

      if (oldLine.trim()) {
        segments.push({
          type: 'removed',
          text: oldLine,
        })
      }

      if (newLine.trim()) {
        segments.push({
          type: 'added',
          text: newLine,
        })
      }

      if (segments.length > 0) {
        result.push({ segments })
      }

      i++
      j++
      continue
    }

    if (j < newLines.length) {
      if (newLine.trim()) {
        result.push({
          segments: [
            {
              type: 'added',
              text: newLine,
            },
          ],
        })
      }
      j++
      continue
    }

    if (i < oldLines.length) {
      if (oldLine.trim()) {
        result.push({
          segments: [
            {
              type: 'removed',
              text: oldLine,
            },
          ],
        })
      }
      i++
    }
  }

  return result
}
