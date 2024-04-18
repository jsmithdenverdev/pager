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
    'sar-writer@pager.com',
    'auth0|66199504eddfc16953adc92e',
    'ACTIVE',
    'SYSTEM',
    'SYSTEM'
  ),
  (
    'e3b2fc56-13e7-4336-b1ef-760a594dc50d',
    'sar-reader@pager.com',
    'auth0|6619959f91e0011a88bfbcd2',
    'ACTIVE',
    'SYSTEM',
    'SYSTEM'
  ),
  (
    'dbbef7f8-0011-4d53-9946-91980e47c9ec',
    'ems-writer@pager.com',
    'auth0|661995c5f2bd935c05de4298',
    'ACTIVE',
    'SYSTEM',
    'SYSTEM'
  ),
  (
    'd289262a-87aa-4d2e-bd59-149379f5c5f9',
    'ems-reader@pager.com',
    'auth0|6619960ae8fb8f9f19f937c6',
    'ACTIVE',
    'SYSTEM',
    'SYSTEM'
  )
  ON CONFLICT DO NOTHING;

-- Create agency record
INSERT INTO agencies (id, "name", "status", created_by, modified_by)
VALUES (
  '8d282649-708b-4523-ab2a-122bd8739bd1',
  'Sample SAR Agency',
  'ACTIVE',
  'auth0|6606e16332edd5a3975a6e9f',
  'auth0|6606e16332edd5a3975a6e9f'
),
(
  '78c95398-3b3d-4e82-a648-0787af58a945',
  'Sample EMS Agency',
  'ACTIVE',
  'auth0|6606e16332edd5a3975a6e9f',
  'auth0|6606e16332edd5a3975a6e9f'
)
ON CONFLICT DO NOTHING;

-- Create user -> agency relation
INSERT INTO user_agencies (user_id, agency_id)
VALUES (
  -- sar-writer@pager.com
  'ec3e634d-e966-44f4-9583-90e3c41fa932',
  -- Sample SAR Agency
  '8d282649-708b-4523-ab2a-122bd8739bd1'
),
(
  -- sar-reader@pager.com
  'e3b2fc56-13e7-4336-b1ef-760a594dc50d',
  -- Sample SAR Agency
  '8d282649-708b-4523-ab2a-122bd8739bd1'
),
(
  -- ems-writer@pager.com
  'dbbef7f8-0011-4d53-9946-91980e47c9ec',
  -- Sample EMS Agency
  '78c95398-3b3d-4e82-a648-0787af58a945'
),
(
  -- ems-reader@pager.com
  'd289262a-87aa-4d2e-bd59-149379f5c5f9',
  -- Sample EMS Agency
  '78c95398-3b3d-4e82-a648-0787af58a945'
)
ON CONFLICT DO NOTHING;

-- Create user roles
INSERT INTO user_roles (role, user_id, agency_id)
VALUES 
  (
    'PLATFORM_ADMIN',
    -- pager-admin@pager.com
    '006c50a1-653e-4aac-a133-89ae607b9a15',
    -- Sample SAR Agency
    NULL
  ),
  (
    'WRITER',
    -- sar-writer@pager.com
    'ec3e634d-e966-44f4-9583-90e3c41fa932',
    -- Sample SAR Agency
    '8d282649-708b-4523-ab2a-122bd8739bd1'
  ),
  (
    'READER',
    -- sar-reader@pager.com
    'e3b2fc56-13e7-4336-b1ef-760a594dc50d',
    -- Sample SAR Agency
    '8d282649-708b-4523-ab2a-122bd8739bd1'
  ),
  (
    'WRITER',
    -- ems-writer@pager.com
    'dbbef7f8-0011-4d53-9946-91980e47c9ec',
    -- Sample EMS Agency
    '78c95398-3b3d-4e82-a648-0787af58a945'
  ),
  (
    'READER',
    -- ems-reader@pager.com
    'd289262a-87aa-4d2e-bd59-149379f5c5f9',
    -- Sample EMS Agency
    '78c95398-3b3d-4e82-a648-0787af58a945'
  )
ON CONFLICT DO NOTHING;