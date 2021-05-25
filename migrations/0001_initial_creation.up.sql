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
	-- room id
	start timestamptz,
	end timestamptz
)

CREATE TABLE alerts (
	id integer PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
	issue_id integer REFERENCES issues (id) ON DELETE CASCADE,
	-- room id
	-- device id
	type text,
	start timestamptz,
	end timestamptz
)

CREATE TABLE incident_mappings (
	id integer PRIMARY KEY GENERATED ALWAYS AS IDENTITY, -- TODO don't really need an ID for this.. issue_id/sn_sys_id should be unique
	issue_id integer REFERENCES issues (id) ON DELETE CASCADE,
	sn_sys_id text, -- service now ticket ID
	sn_ticket_number text, -- ticket number (INCXXXXXX)
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
