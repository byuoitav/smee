ALTER TABLE issues
ADD acknowledge_by text,
ADD acknowledge_time timestamptz;

ALTER TABLE alerts
ADD acknowledge_by text,
ADD acknowledge_time timestamptz;
