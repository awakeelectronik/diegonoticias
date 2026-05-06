#!/usr/bin/env bash
set -euo pipefail

export DN_ENV="${DN_ENV:-development}"

go run ./cmd/server

