#!/bin/bash
# Wrapper to bootstrap the Go installer

export NEJEN_SOURCE_PATH="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$NEJEN_SOURCE_PATH" || exit 1
go run ./cmd/nejen install "$@"
