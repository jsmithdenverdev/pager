-- Deploy pager:agency_devices to pg

BEGIN;

CREATE TABLE IF NOT EXISTS agency_devices (
  agency_id UUID NOT NULL REFERENCES agencies(id),
  device_id UUID NOT NULL REFERENCES devices(id),
  CONSTRAINT agency_devices_unique_agency_id_device_id UNIQUE (agency_id, device_id)
);

COMMIT;
