-- Row-Level Security (RLS) on the users table.
--
-- Model: a request can only touch its OWN row, admins see everything, and
-- pre-auth flows (login lookup, registration, password reset, seeding) run
-- in a privileged "service" role. The application carries the caller's
-- identity into each transaction via two session GUCs that the app sets
-- with SET LOCAL (see internal/datasources/repositories/postgres/users):
--
--     app.user_id    -- the authenticated user's UUID (text), or '' for service/admin
--     app.user_role  -- 'service' | 'admin' | 'user' (empty => deny all)
--
-- current_setting(name, true) returns NULL when the GUC is unset; the
-- IN (...) / equality checks then evaluate to NULL (not true), so a request
-- that forgot to set its identity is denied by default — defense in depth
-- on top of the WHERE clauses the repository already writes.

ALTER TABLE users ENABLE ROW LEVEL SECURITY;

-- FORCE is required here: the application connects as the table OWNER, and
-- owners BYPASS RLS unless it is forced. Without FORCE these policies would
-- have no effect for this app.
--
-- IMPORTANT — superusers ALWAYS bypass RLS, and FORCE cannot change that. The
-- official postgres Docker image creates POSTGRES_USER as a SUPERUSER, so for
-- these policies to actually take effect the app must connect as a
-- NON-superuser role (see docs/SECURITY.md "DB role separation"). Verify with:
--     SELECT rolsuper, rolbypassrls FROM pg_roles WHERE rolname = current_user;
-- both columns must be false for RLS to be enforced against the app.
ALTER TABLE users FORCE ROW LEVEL SECURITY;

-- SELECT: service/admin see all rows; a plain user sees only their own.
CREATE POLICY users_select ON users
    FOR SELECT
    USING (
        current_setting('app.user_role', true) IN ('service', 'admin')
        OR id::text = current_setting('app.user_id', true)
    );

-- INSERT: only service/admin may create user rows (registration runs as
-- service). A plain user can never insert arbitrary user rows.
CREATE POLICY users_insert ON users
    FOR INSERT
    WITH CHECK (
        current_setting('app.user_role', true) IN ('service', 'admin')
    );

-- UPDATE: service/admin may update any row; a plain user may update only
-- their own, and may not move a row out from under their own id (WITH CHECK
-- mirrors USING).
CREATE POLICY users_update ON users
    FOR UPDATE
    USING (
        current_setting('app.user_role', true) IN ('service', 'admin')
        OR id::text = current_setting('app.user_id', true)
    )
    WITH CHECK (
        current_setting('app.user_role', true) IN ('service', 'admin')
        OR id::text = current_setting('app.user_id', true)
    );

-- DELETE: service/admin may delete any row; a plain user may delete only
-- their own. (Soft-delete goes through UPDATE; this covers any hard DELETE.)
CREATE POLICY users_delete ON users
    FOR DELETE
    USING (
        current_setting('app.user_role', true) IN ('service', 'admin')
        OR id::text = current_setting('app.user_id', true)
    );
