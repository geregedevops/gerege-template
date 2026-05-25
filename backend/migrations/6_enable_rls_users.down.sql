-- Reverse migrations/6_enable_rls_users.up.sql: drop the policies and turn
-- RLS back off so the users table behaves as it did before.
DROP POLICY IF EXISTS users_delete ON users;
DROP POLICY IF EXISTS users_update ON users;
DROP POLICY IF EXISTS users_insert ON users;
DROP POLICY IF EXISTS users_select ON users;

ALTER TABLE users NO FORCE ROW LEVEL SECURITY;
ALTER TABLE users DISABLE ROW LEVEL SECURITY;
