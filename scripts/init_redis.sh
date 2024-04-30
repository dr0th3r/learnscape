#!/usr/bin/bash
set -x
set -eo pipefail

docker run \
	-p "6379:6379" \
	-d \
	--name "redis_$(date '+%s')" \
	redis:6

echo >&2 "Redis is up and running"
