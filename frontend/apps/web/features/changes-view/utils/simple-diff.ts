export type DiffResult = {
  value: string
  added?: boolean
  removed?: boolean
}[]

export function diffLines(oldText: string, newText: string): DiffResult {
  const oldLines = oldText.split('\n')
  const newLines = newText.split('\n')
  const result: DiffResult = []

  let i = 0
  let j = 0

  while (i < oldLines.length || j < newLines.length) {
    if (i < oldLines.length && j < newLines.length && oldLines[i] === newLines[j]) {
      result.push({ value: oldLines[i] || '' })
      i++
      j++
    } else if (j < newLines.length && (i >= oldLines.length || !oldLines.includes(newLines[j] || '', i))) {
      result.push({ value: newLines[j] || '', added: true })
      j++
    } else if (i < oldLines.length) {
      result.push({ value: oldLines[i] || '', removed: true })
      i++
    } else {
      // Should not happen if logic is correct for simple cases
      j++
    }
  }

  return result
}
