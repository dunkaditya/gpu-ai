#!/usr/bin/env python3
"""Database migration runner.

Applies SQL migration files from database/migrations/ in order.

Usage:
    python tools/migrate.py                  # apply pending migrations
    python tools/migrate.py --status         # show migration status
    python tools/migrate.py --target V0      # migrate to specific version
"""

import glob
import os
import sys

import click
import psycopg2
from tabulate import tabulate

# Resolve migrations directory relative to this script's location
SCRIPT_DIR = os.path.dirname(os.path.abspath(__file__))
PROJECT_ROOT = os.path.dirname(SCRIPT_DIR)
MIGRATIONS_DIR = os.path.join(PROJECT_ROOT, "database", "migrations")

SCHEMA_MIGRATIONS_DDL = """
CREATE TABLE IF NOT EXISTS schema_migrations (
    version VARCHAR(255) PRIMARY KEY,
    applied_at TIMESTAMPTZ DEFAULT NOW()
)
"""


def get_connection():
    """Connect to PostgreSQL using DATABASE_URL environment variable."""
    database_url = os.environ.get("DATABASE_URL")
    if not database_url:
        click.echo("Error: DATABASE_URL environment variable is not set.", err=True)
        sys.exit(1)

    try:
        conn = psycopg2.connect(database_url)
        return conn
    except psycopg2.Error as e:
        click.echo(f"Error: Could not connect to database: {e}", err=True)
        sys.exit(1)


def ensure_schema_migrations(conn):
    """Create the schema_migrations tracking table if it does not exist."""
    with conn.cursor() as cur:
        cur.execute(SCHEMA_MIGRATIONS_DDL)
    conn.commit()


def get_applied_migrations(conn):
    """Return a dict of {version: applied_at} for all applied migrations."""
    with conn.cursor() as cur:
        cur.execute("SELECT version, applied_at FROM schema_migrations ORDER BY version")
        return {row[0]: row[1] for row in cur.fetchall()}


def get_migration_files():
    """Scan the migrations directory for .sql files, sorted by filename."""
    pattern = os.path.join(MIGRATIONS_DIR, "*.sql")
    files = sorted(glob.glob(pattern))
    return files


@click.command()
@click.option("--status", is_flag=True, default=False, help="Show migration status (applied vs pending)")
@click.option("--target", default=None, help="Migrate up to and including this version filename")
def main(status, target):
    """Apply pending database migrations or show migration status."""
    conn = get_connection()

    try:
        ensure_schema_migrations(conn)
        applied = get_applied_migrations(conn)
        migration_files = get_migration_files()

        if not migration_files:
            click.echo("No migration files found in: {}".format(MIGRATIONS_DIR))
            return

        if status:
            show_status(migration_files, applied)
            return

        # Validate --target if provided
        if target is not None:
            target_found = False
            for f in migration_files:
                if os.path.basename(f) == target:
                    target_found = True
                    break
            if not target_found:
                click.echo(
                    "Error: Target version '{}' not found in migration files.".format(target),
                    err=True,
                )
                click.echo("Available migrations:", err=True)
                for f in migration_files:
                    click.echo("  - {}".format(os.path.basename(f)), err=True)
                sys.exit(1)

        apply_migrations(conn, migration_files, applied, target)

    finally:
        conn.close()


def show_status(migration_files, applied):
    """Display a table showing the status of each migration."""
    rows = []
    for f in migration_files:
        version = os.path.basename(f)
        if version in applied:
            applied_at = applied[version].strftime("%Y-%m-%d %H:%M:%S %Z")
            rows.append([version, "applied", applied_at])
        else:
            rows.append([version, "pending", "-"])

    headers = ["Version", "Status", "Applied At"]
    click.echo(tabulate(rows, headers=headers, tablefmt="simple"))


def apply_migrations(conn, migration_files, applied, target):
    """Apply pending migrations in order, each in its own transaction."""
    pending_count = 0

    for f in migration_files:
        version = os.path.basename(f)

        # Skip already-applied migrations
        if version in applied:
            continue

        # Read and execute the migration SQL
        try:
            with open(f, "r") as sql_file:
                sql_content = sql_file.read()

            with conn.cursor() as cur:
                cur.execute(sql_content)
                cur.execute(
                    "INSERT INTO schema_migrations (version) VALUES (%s)",
                    (version,),
                )
            conn.commit()
            click.echo("Applied: {}".format(version))
            pending_count += 1

        except psycopg2.Error as e:
            conn.rollback()
            click.echo(
                "Error applying migration '{}': {}".format(version, e),
                err=True,
            )
            sys.exit(1)

        # Stop if we reached the target version
        if target is not None and version == target:
            break

    if pending_count == 0:
        click.echo("No pending migrations.")
    else:
        click.echo("\n{} migration(s) applied successfully.".format(pending_count))


if __name__ == "__main__":
    main()
