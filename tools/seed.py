#!/usr/bin/env python3
"""Database seeder for local development.

Creates sample organizations, users, SSH keys, and instances
for testing the dashboard and API.

Usage:
    python tools/seed.py              # seed with defaults
    python tools/seed.py --clean      # drop and re-seed
"""

# TODO: Implement database seeder:
#
# - Connect to DATABASE_URL from env
# - Create sample data:
#   - 2 organizations (GPU.ai Labs, Acme Corp)
#   - 3 users across orgs (admin + members)
#   - SSH keys per user
#   - A few instances in various states (running, terminated)
#   - Usage records for billing display
# - Support --clean flag to truncate tables first
