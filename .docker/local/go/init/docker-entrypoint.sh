#!/bin/bash

set -e

sleep 10

go mod vendor

exec "$@"