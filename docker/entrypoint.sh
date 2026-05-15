#!/bin/sh
set -e

IMGPROXY_BIND=:8080 IMGPROXY_LOG_LEVEL=${IMGPROXY_LOG_LEVEL:-warn} /usr/local/bin/imgproxy &

exec /usr/local/bin/img-fwd
