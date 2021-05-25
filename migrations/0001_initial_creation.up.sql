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

-- Rooms
CREATE TABLE issues (
	id integer PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
	couch_room_id text,
	start_time timestamptz,
	end_time timestamptz
);

CREATE TABLE alerts (
	id integer PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
	issue_id integer REFERENCES issues (id) ON DELETE CASCADE,
	couch_room_id text,
	couch_device_id text,
	alert_type text,
	start_time timestamptz,
	end_time timestamptz
);

CREATE TABLE sn_incident_mappings (
	issue_id integer REFERENCES issues (id) ON DELETE CASCADE,
	sn_sys_id text, -- service now ticket ID
	sn_ticket_number text, -- ticket number (INCXXXXXX)
	PRIMARY KEY (issue_id, sn_sys_id)
);

CREATE TABLE issue_events (
	id integer PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
	issue_id integer REFERENCES issues (id) ON DELETE CASCADE,
	ts timestamptz,
	event_type text,
	data jsonb
);

