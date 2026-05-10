import * as migration_20260506_000000 from './20260506_000000';
import * as migration_20260506_000001 from './20260506_000001';
import * as migration_20260507_182220 from './20260507_182220';

export const migrations = [
  {
        up: migration_20260506_000000.up,
        down: migration_20260506_000000.down,
        name: '20260506_000000',
  },
  {
        up: migration_20260506_000001.up,
        down: migration_20260506_000001.down,
        name: '20260506_000001',
  },
  {
        up: migration_20260507_182220.up,
        down: migration_20260507_182220.down,
        name: '20260507_182220',
  },
  ];
