import * as migration_20260506_000000 from './20260506_000000'
import * as migration_20260506_000002 from './20260506_000002'
import * as migration_20260507_182220 from './20260507_182220'
import * as migration_20260604_000000 from './20260604_000000'

export const migrations = [
  {
    up: migration_20260506_000000.up,
    down: migration_20260506_000000.down,
    name: '20260506_000000',
  },
  {
    up: migration_20260506_000002.up,
    down: migration_20260506_000002.down,
    name: '20260506_000002',
  },
  {
    up: migration_20260507_182220.up,
    down: migration_20260507_182220.down,
    name: '20260507_182220',
  },
  {
    up: migration_20260604_000000.up,
    down: migration_20260604_000000.down,
    name: '20260604_000000',
  },
]
