import * as migration_20260507_182220 from './20260507_182220';

export const migrations = [
  {
    up: migration_20260507_182220.up,
    down: migration_20260507_182220.down,
    name: '20260507_182220',
  },
];
