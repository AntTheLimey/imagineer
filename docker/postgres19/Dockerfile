#--------------------------------------------------------------------------
#
# Imagineer - TTRPG Campaign Intelligence Platform
#
# Copyright (c) 2025 - 2026
# This software is released under The MIT License
#
#--------------------------------------------------------------------------

ARG PG19_IMAGE=postgres:19beta2
FROM ${PG19_IMAGE}
# btree_gist/pg_trgm ship in contrib; nothing to build.
# PG19_IMAGE is a build arg so the pgEdge PG19 image (not publicly
# published as of 2026-08-11; imagineer runs pgedge-postgres:18 today)
# can be swapped in via `make db2-up PG19_IMAGE=<ref>` when it lands.
