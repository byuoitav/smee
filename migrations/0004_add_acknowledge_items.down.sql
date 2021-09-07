ALTER TABLE issues
DROP COLUMN acknowledge_by,
DROP COLUMN acknowledge_time;

ALTER TABLE alerts
DROP COLUMN acknowledge_by;
DROP COLUMN acknowledge_time;
