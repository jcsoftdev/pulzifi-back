-- Normalize verbose frequency values to short format in monitoring_configs

UPDATE monitoring_configs
SET check_frequency = CASE check_frequency
    WHEN 'Every 5 minutes'  THEN '5m'
    WHEN 'Every 10 minutes' THEN '10m'
    WHEN 'Every 15 minutes' THEN '15m'
    WHEN 'Every 30 minutes' THEN '30m'
    WHEN 'Every hour'       THEN '1h'
    WHEN 'Every 2 hours'    THEN '2h'
    WHEN 'Every 4 hours'    THEN '4h'
    WHEN 'Every 6 hours'    THEN '6h'
    WHEN 'Every 12 hours'   THEN '12h'
    WHEN 'Every day'        THEN '24h'
    WHEN 'Every week'       THEN '168h'
    ELSE check_frequency
END
WHERE check_frequency IN (
    'Every 5 minutes',
    'Every 10 minutes',
    'Every 15 minutes',
    'Every 30 minutes',
    'Every hour',
    'Every 2 hours',
    'Every 4 hours',
    'Every 6 hours',
    'Every 12 hours',
    'Every day',
    'Every week'
);
