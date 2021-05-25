--
--
-- SMEE
--
--

--
-- Create smee schema
--
CREATE SCHEMA smee AUTHORIZATION smee;

-- Switch to smee schema
SET search_path TO smee;
SET role smee;

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
	start timestamptz,
	end timestamptz
)

CREATE TABLE alerts (
	id integer PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
	issue_id integer REFERENCES issues (id) ON DELETE CASCADE,
	couch_room_id text,
	couch_device_id text,
	type text,
	start timestamptz,
	end timestamptz
)

CREATE TABLE sn_incident_mappings (
	issue_id integer REFERENCES issues (id) ON DELETE CASCADE,
	sn_sys_id text, -- service now ticket ID
	sn_ticket_number text, -- ticket number (INCXXXXXX)
	PRIMARY KEY (issue_id, sn_sys_id)
)

CREATE TABLE issue_events (
	id integer PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
	issue_id integer REFERENCES issues (id) ON DELETE CASCADE,
	timestamp timestamptz,
	type text,
	data jsonb
)

-- Exit schema
SET search_path TO public;
RESET ROLE;
