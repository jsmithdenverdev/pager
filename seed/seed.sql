-- Create user records
INSERT INTO users (id, email, idp_id, "status", created_by, modified_by)
VALUES (
    '006c50a1-653e-4aac-a133-89ae607b9a15',
    'pager-admin@pager.com',
    'auth0|6606e16332edd5a3975a6e9f',
    'ACTIVE',
    'SYSTEM',
    'SYSTEM'
  ),
  (
    'ec3e634d-e966-44f4-9583-90e3c41fa932',
    'agency-admin@pager.com',
    'auth0|6606e29442d9ec8442da383c',
    'ACTIVE',
    'SYSTEM',
    'SYSTEM'
  )
  ON CONFLICT DO NOTHING;

-- Create agency record
INSERT INTO agencies (id, "name", "status", created_by, modified_by)
VALUES (
  '8d282649-708b-4523-ab2a-122bd8739bd1',
  'Sample Agency',
  'ACTIVE',
  'auth0|6606e16332edd5a3975a6e9f',
  'auth0|6606e16332edd5a3975a6e9f'
)
ON CONFLICT DO NOTHING;

-- Create user -> agency relation
INSERT INTO user_agencies (user_id, agency_id)
VALUES (
  'ec3e634d-e966-44f4-9583-90e3c41fa932',
  '8d282649-708b-4523-ab2a-122bd8739bd1'
)
ON CONFLICT DO NOTHING;