#!/bin/sh
# Runs ONCE, at first database initialisation, by the postgres image
# entrypoint — executed as the bootstrap superuser POSTGRES_USER against
# POSTGRES_DB (before the migrate container runs).
#
# Why: the users table enforces Row-Level Security (migration 6), but a
# SUPERUSER (which the postgres image makes POSTGRES_USER) BYPASSES RLS even
# when it is FORCED. So we create a NON-superuser role for the API to connect
# as; then the RLS policies actually take effect. The migrate container keeps
# using POSTGRES_USER (it needs superuser for CREATE EXTENSION "uuid-ossp").
#
# If APP_DB_USER / APP_DB_PASSWORD are not provided, this is a no-op and the
# app falls back to connecting as the superuser (RLS present but bypassed).
set -e

if [ -z "${APP_DB_USER}" ] || [ -z "${APP_DB_PASSWORD}" ]; then
  echo "initdb: APP_DB_USER/APP_DB_PASSWORD not set — skipping non-superuser role (RLS will be bypassed by the superuser app role)."
  exit 0
fi

# Use a hex/alphanumeric password (no single quotes) — values are inlined into
# SQL below. The compose .env generates these with `openssl rand -hex`.
psql -v ON_ERROR_STOP=1 --username "${POSTGRES_USER}" --dbname "${POSTGRES_DB}" <<SQL
DO \$\$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = '${APP_DB_USER}') THEN
    CREATE ROLE "${APP_DB_USER}" LOGIN PASSWORD '${APP_DB_PASSWORD}' NOSUPERUSER NOBYPASSRLS;
  END IF;
END
\$\$;

GRANT USAGE ON SCHEMA public TO "${APP_DB_USER}";

-- Tables are created later by the migrate container (running as
-- ${POSTGRES_USER}); grant DML on those future tables automatically so the
-- API role never needs superuser. RLS still constrains which ROWS it sees.
ALTER DEFAULT PRIVILEGES FOR ROLE "${POSTGRES_USER}" IN SCHEMA public
  GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO "${APP_DB_USER}";
ALTER DEFAULT PRIVILEGES FOR ROLE "${POSTGRES_USER}" IN SCHEMA public
  GRANT USAGE, SELECT ON SEQUENCES TO "${APP_DB_USER}";
SQL

echo "initdb: created non-superuser role '${APP_DB_USER}' for the API (RLS enforced)."
