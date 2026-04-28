#!/bin/sh
set -e

mc alias set local "http://minio:${MINIO_PORT}" "$MINIO_ROOT_USER" "$MINIO_ROOT_PASSWORD"
mc mb local/"$MINIO_DEFAULT_BUCKETS" --ignore-existing
