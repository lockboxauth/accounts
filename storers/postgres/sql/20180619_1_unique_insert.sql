-- +migrate Up
ALTER TABLE accounts ADD COLUMN is_registration BOOLEAN,
		     ADD CONSTRAINT unique_registration unique (profile_id, is_registration);

-- +migrate Down
ALTER TABLE accounts DROP COLUMN registration,
		     DROP CONSTRAINT unique_registration;
