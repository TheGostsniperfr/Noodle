#!/usr/bin/env bash
# Kept because immutable verify steps in docs/features.json cite it (C12). Use make init.
exec make -C "$(dirname "$0")/.." init
