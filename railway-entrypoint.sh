#!/bin/sh
set -eu

# When a Railway Volume is mounted at /app/uploads it starts empty.
# Seed it once from the files bundled in the image so existing project files
# are available on first deployment, while later deploys keep uploaded data.
if [ -d /app/seed_uploads ]; then
  if [ -z "$(find /app/uploads -mindepth 1 -maxdepth 1 -print -quit 2>/dev/null)" ]; then
    cp -a /app/seed_uploads/. /app/uploads/ || true
  fi
fi

exec /app/server
