--
--
-- SMEE
--
--

--
-- Create Types
--

--
-- Create Tables
--
CREATE TABLE issues (
	id integer PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
	couch_room_id text IS NOT NULL,
	start_time timestamptz IS NOT NULL,
	end_time timestamptz
	acknowledged_at timestamptz
	acknowledged_by text
);

CREATE TABLE alerts (
	id integer PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
	issue_id integer REFERENCES issues (id) ON DELETE CASCADE IS NOT NULL,
	couch_room_id text IS NOT NULL,
	couch_device_id text IS NOT NULL,
	alert_type text IS NOT NULL,
	start_time timestamptz IS NOT NULL,
	end_time timestamptz
	acknowledged_at timestamptz
	acknowledged_by text
);

CREATE TABLE sn_incident_mappings (
	issue_id integer REFERENCES issues (id) ON DELETE CASCADE IS NOT NULl,
	sn_sys_id text IS NOT NULL, -- ticket ID
	sn_ticket_number text IS NOT NULL, -- ticket number (INCXXXXXX)
	PRIMARY KEY (issue_id, sn_sys_id)
);

CREATE TABLE issue_events (
	id integer PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
	issue_id integer REFERENCES issues (id) ON DELETE CASCADE IS NOT NULL,
	time timestamptz IS NOT NULL,
	event_type text IS NOT NULL,
	data jsonb
);

CREATE TABLE room_maintenance_couch (
	couch_room_id text PRIMARY KEY,
	start_time timestamptz IS NOT NULL,
	end_time timestamptz IS NOT NULL
);
