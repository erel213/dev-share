-- Note: is_admin column still exists, just sync it back from role
UPDATE users SET is_admin = 1 WHERE role = 'admin';
UPDATE users SET is_admin = 0 WHERE role != 'admin';

ALTER TABLE users DROP COLUMN role;
