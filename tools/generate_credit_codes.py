#!/usr/bin/env python3
"""Generate GPU.ai credit codes and insert hashed versions into the database.

Usage:
    python tools/generate_credit_codes.py --amount 25.00 --count 10 [--expires 2026-12-31]

Requires DATABASE_URL environment variable.
Plaintext codes are printed to stdout but NEVER stored in the database.
"""

import argparse
import hashlib
import os
import secrets
import sys

import psycopg2

# Unambiguous character set (no 0/O/1/I/L)
CHARSET = "ABCDEFGHJKMNPQRSTUVWXYZ23456789"


def generate_code() -> str:
    """Generate a GPU-XXXX-XXXX credit code."""
    part1 = "".join(secrets.choice(CHARSET) for _ in range(4))
    part2 = "".join(secrets.choice(CHARSET) for _ in range(4))
    return f"GPU-{part1}-{part2}"


def hash_code(code: str) -> str:
    """SHA-256 hash of the normalized (uppercased, trimmed) code."""
    return hashlib.sha256(code.strip().upper().encode()).hexdigest()


def main():
    parser = argparse.ArgumentParser(description="Generate GPU.ai credit codes")
    parser.add_argument("--amount", type=float, required=True, help="Credit amount in dollars (e.g. 25.00)")
    parser.add_argument("--count", type=int, default=1, help="Number of codes to generate")
    parser.add_argument("--expires", type=str, default=None, help="Expiry date (YYYY-MM-DD)")
    args = parser.parse_args()

    if args.amount < 0.01:
        print("Error: amount must be at least $0.01", file=sys.stderr)
        sys.exit(1)

    database_url = os.environ.get("DATABASE_URL")
    if not database_url:
        print("Error: DATABASE_URL environment variable is required", file=sys.stderr)
        sys.exit(1)

    amount_cents = int(round(args.amount * 100))

    expires_at = None
    if args.expires:
        expires_at = args.expires + "T23:59:59Z"

    conn = psycopg2.connect(database_url)
    cur = conn.cursor()

    codes = []
    for _ in range(args.count):
        code = generate_code()
        cur.execute(
            "INSERT INTO credit_codes (code, amount_cents, expires_at) VALUES (%s, %s, %s)",
            (hash_code(code), amount_cents, expires_at),
        )
        codes.append(code)

    conn.commit()
    cur.close()
    conn.close()

    print(f"Generated {len(codes)} credit code(s) worth ${args.amount:.2f} each:")
    print("(Save these now — they are NOT stored in the database)")
    print()
    for code in codes:
        print(f"  {code}")


if __name__ == "__main__":
    main()
