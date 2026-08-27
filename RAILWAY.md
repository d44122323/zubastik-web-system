# Railway deployment — PostgreSQL

This project is prepared to run on Railway with PostgreSQL.

## What changed

- PostgreSQL uses `DATABASE_URL` when present (Railway PostgreSQL provides it automatically).
- Local development can use `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`, `DB_SSLMODE`.
- The server listens on Railway's `PORT` variable.
- Public forms use the relative `/submit` endpoint.
- Uploaded doctor photos and medical files use `UPLOADS_DIR`; Railway should mount a persistent Volume at `/app/uploads`.
- `/health` is available for Railway healthchecks and returns HTTP 200 only when PostgreSQL is reachable.
- Production cookies are marked Secure via `APP_ENV=production`.
- Secrets are not included in the production archive.

## Railway setup

1. Create a new Railway project.
2. Add a PostgreSQL service.
3. Add this project as a service from GitHub or deploy it with the Railway CLI. Railway detects the root `Dockerfile` automatically.
4. In the application service Variables, add the secret variables from `.env.example` (without the local DB_* variables if `DATABASE_URL` is supplied by PostgreSQL).
5. Set `DATABASE_URL` to the PostgreSQL service's `DATABASE_URL` reference if Railway does not automatically provide it to the app service. In the dashboard this can be a reference such as `${{Postgres.DATABASE_URL}}` where `Postgres` is the exact name of the database service.
6. Set `APP_ENV=production`.
7. Set `SITE_URL` to the final public URL, e.g. `https://your-app.up.railway.app`, and later replace it with the custom domain when you add one.
8. Create a Railway Volume for the application service and mount it at `/app/uploads`.
9. In service settings, set Healthcheck Path to `/health`.
10. Generate a public Railway domain under Networking.
11. After the first deployment, check `/health` and the deployment logs.

## Database data

The application creates/updates its required tables on startup, including the base patients/requests tables before feature-specific migrations. A fresh Railway PostgreSQL service will therefore get the schema, but it will not automatically contain your current local patients, doctors, requests, appointments, medical records, services, etc.

To preserve current data, export the local PostgreSQL database with `pg_dump` and restore it into the Railway PostgreSQL database before/around the first production launch. Do not overwrite a production database without a backup.

## Telegram

After the public URL is available, configure any Telegram webhook/callback URLs to use the public HTTPS address. The patient and staff bots remain separate and continue to use their own environment variables.

## Security

Do not commit `.env` or paste real production secrets into Git. The archive intentionally excludes `.env`. If a token/password has previously been exposed, rotate it before production deployment.
