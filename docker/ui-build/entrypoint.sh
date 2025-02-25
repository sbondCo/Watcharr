#!/bin/sh

# This is loosely based on the main containers Dockerfile commands.
# Just everything needed to get ui built and ready and nothing else.

set -x

echo "wuib: Starting build"

# Install all deps
/usr/local/bin/npm install

# Run build
/usr/local/bin/npm run build

# Remove any existing build files from volume
rm -rf /ui/build

# Move build folder to /ui
mv /app/build /ui
# Move package.json and lock file to /ui/build for final npm ci call
mv /app/package.json /app/package-lock.json /ui/build

# cd to build files final destination and install dependencies for production
cd /ui/build && /usr/local/bin/npm ci --omit=dev --ignore-scripts=true
