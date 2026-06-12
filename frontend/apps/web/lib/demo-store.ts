import { existsSync, readFileSync, writeFileSync } from 'node:fs'
import { join } from 'node:path'

const FILE = join(process.cwd(), '.demo-state.json')

type DemoState = {
  lectureAiVersion: '1' | '2'
}

const DEFAULT: DemoState = { lectureAiVersion: '2' }

export function getDemoState(): DemoState {
  try {
    if (!existsSync(FILE)) return DEFAULT
    return JSON.parse(readFileSync(FILE, 'utf-8'))
  } catch {
    return DEFAULT
  }
}

export function setDemoState(patch: Partial<DemoState>): DemoState {
  const next = { ...getDemoState(), ...patch }
  writeFileSync(FILE, JSON.stringify(next))
  return next
}
